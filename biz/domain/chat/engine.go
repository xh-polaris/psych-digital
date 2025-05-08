package chat

import (
	"context"
	"errors"
	"github.com/hertz-contrib/websocket"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/application/dto"
	"github.com/xh-polaris/psych-digital/biz/domain"
	"github.com/xh-polaris/psych-digital/biz/domain/model"
	"github.com/xh-polaris/psych-digital/biz/domain/model/bailian"
	"github.com/xh-polaris/psych-digital/biz/domain/model/volc"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/mq"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/util"
	"io"
	"strings"
	"time"
)

// Engine 是处理一轮对话的核心对象
// 只读取文字对话, 语言识别由另一个ws连接处理
type Engine struct {
	// ctx 上下文
	ctx context.Context

	// cancel 取消goroutine的广播函数
	cancel context.CancelFunc

	// ws 提供WebSocket的读写功能
	ws *domain.WsHelper

	// rs 提供redis的读写功能
	rs *domain.RedisHelper
	//rs *domain.MemoryRedisHelper

	// chatApp 是调用的对话大模型
	chatApp model.ChatApp

	// ttsApp 是调用的语音合成大模型
	ttsApp model.TtsApp

	// sessionId 是本轮对话的唯一标记, 只有第一次调用时会写入, 应该不需要互斥锁
	// 目前使用的是BaiLian提供的sessionId管理, 如果有更好的方式, 可以考虑自己实现
	sessionId string

	// aiHistory 记录AI输出历史
	aiHistory chan string

	// userHistory 记录用户输入历史
	userHistory chan string

	// outw ai的流式文本, 用于语音合成
	outw chan string

	// outv 合成的流式语音
	outv chan []byte

	// stop 用于打断AI输出
	stop chan bool

	// startTime 开始对话时间
	startTime time.Time

	// provider 消息生产者
	provider *mq.HistoryProducer
}

// NewEngine 初始化一个ChatEngine
// 暂时先固定为BaiLian之后类型多再换成工厂方法
func NewEngine(ctx context.Context, conn *websocket.Conn) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	c := config.GetConfig()
	e := &Engine{
		ctx:    ctx,
		cancel: cancel,
		ws:     domain.NewWsHelper(conn),
		rs:     domain.GetRedisHelper(),
		//rs:          domain.NewMemoryRedisHelper(),
		chatApp:     bailian.NewBLChatApp(c.BaiLianChat.AppId, c.BaiLianChat.ApiKey),
		ttsApp:      volc.NewVcTtsApp(c.VolcTts.AppKey, c.VolcTts.AccessKey, c.VolcTts.Speaker, c.VolcTts.ResourceId, c.VolcTts.Url),
		aiHistory:   make(chan string, 10),
		userHistory: make(chan string, 10),
		outw:        make(chan string, 50),
		outv:        make(chan []byte, 50),
		stop:        make(chan bool),
		startTime:   time.Now(),
		provider:    mq.GetHistoryProducer(),
	}
	return e
}

// Start 开始一轮对话, 执行相关初始化
func (e *Engine) Start() error {
	var err error

	// 鉴权
	if !e.validate() {
		_ = e.ws.Error(consts.ErrInvalidUser)
		return consts.ErrInvalidUser
	}

	msg := "你好呀"

	// 音频生成
	if err = e.tts(); err != nil {
		return err
	}

	// chat模型调用
	go e.streamCall(msg)

	// 由于sessionId由第三方给出, 所以这里需要手动管理聊天记录的顺序
	// 等待获取sessionId, 初始化redis
	his := <-e.aiHistory
	if err = e.rs.AddSystem(e.sessionId, msg); err != nil {
		return err
	}
	if err = e.rs.AddAi(e.sessionId, his); err != nil {
		return err
	}
	return err
}

// validate 校验使用者信息, 目前没有鉴权，只做一下日志
func (e *Engine) validate() bool {
	var startReq dto.ChatStartReq

	err := e.ws.ReadJSON(&startReq)
	if err != nil {
		log.Error("read json err:", err)
		return false
	}
	log.Info("调用方: %s, 调用时间: %s", startReq.From, time.Unix(startReq.Timestamp, 0).String())
	return true
}

// Chat 长对话的主体部分 #生产者
func (e *Engine) Chat() {
	var req dto.ChatReq
	var err error
	defer func() {
		if err != nil {
			log.Error("chat err:", err)
		}
	}()

	// 启动聊天记录处理
	go e.history(e.aiHistory, e.userHistory)

	for {
		// 获取前端对话内容
		err = e.ws.ReadJSON(&req)
		if err != nil {
			return
		}
		// 判断是否结束
		switch req.Cmd {
		case consts.EndCmd:
			return
		case consts.Ping:
			err := e.ws.WriteBytes([]byte{})
			if err != nil {
				return
			}
			continue

		}
		// 写入用户消息
		e.userHistory <- req.Msg
		// 调用ai, 流式响应
		go e.streamCall(req.Msg)
	}
}

// streamCall 调用chatApp并流式写入响应 #生产者
func (e *Engine) streamCall(msg string) {
	var record string
	var data *dto.ChatData

	// 流式响应的scanner
	scanner, err := e.chatApp.StreamCall(msg, e.sessionId)
	defer func() {
		_ = scanner.Close()
		switch {
		case errors.Is(err, io.EOF):
			e.aiHistory <- record
		default:
			// 错误时写入异常值, 避免主协程无限等待
			e.aiHistory <- "stop:" + err.Error()
		}
	}()

	// 将模型结果响应给前端
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			// 获取下一次响应
			data, err = scanner.Next()
			if err != nil {
				return
			}
			// 第一次调用, 写入sessionId
			if e.sessionId == "" {
				e.sessionId = data.SessionId
			}
			// 风险分析
			analyse(&data.Content)
			// 写入文本, 用于音频合成
			e.outw <- data.Content
			// 写入响应 TODO: test待删除
			log.Info("data: ", data)
			err = e.ws.WriteJSON(data)
			if err != nil {
				return
			}
			// 拼接聊天记录
			record += data.Content
		}
	}
}

// tts 初始化tts app 并启动发送和接受goroutine
func (e *Engine) tts() error {
	err := e.ttsInit()
	if err != nil {
		return err
	}
	go e.ttsUp(e.outw)
	go e.ttsDown()
	return nil
}

// ttsInit 初始话音频生成
func (e *Engine) ttsInit() (err error) {
	if err = e.ttsApp.Dial(); err != nil {
		return
	}
	if err = e.ttsApp.Start(); err != nil {
		return
	}
	return
}

// ttsUp 上传合成音频用文字 #消费者
func (e *Engine) ttsUp(texts chan string) {
	var err error

	for text := range texts {
		if err = e.ttsApp.Send(text); err != nil {
			log.Error("send tts err:", err)
			return
		}
	}
}

// ttsDown 获取生成的音频 #生产者
func (e *Engine) ttsDown() {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			audio := e.ttsApp.Receive()
			if audio != nil {
				err := e.ws.WriteBytes(audio)
				if err != nil {
					log.Error("ws write audio err:", err)
				}
			}
		}
	}
}

// history 处理聊天记录 #消费者
func (e *Engine) history(ai, user chan string) {
	for {
		select {
		case his, ok := <-ai:
			if !ok {
				ai = nil
			}
			if his != "" {
				if err := e.rs.AddAi(e.sessionId, his); err != nil {
					log.Error("ai history err:", err)
				}
			}
		case his, ok := <-user:
			if !ok {
				user = nil
			}
			if his != "" {
				if err := e.rs.AddUser(e.sessionId, his); err != nil {
					log.Error("user history err:", err)
				}
			}
		}
	}
}

// Close 结束本轮对话
func (e *Engine) Close() {
	// 发送结束标识
	err := e.ws.WriteJSON(&dto.ChatEndResp{
		Code: 0,
		Msg:  "对话结束",
	})
	if err != nil {
		log.Error(err.Error())
		return
	}
	e.cancel()
	_ = e.close()
	// 发送对话历史记录消息
	if err = e.provider.Produce(e.ctx, e.sessionId, e.startTime, time.Now()); err != nil {
		log.Error("消息发送失败, sessionId: ", e.sessionId)
	}

}

// close 释放相关资源
// 所有的通道由close统一关闭, 生产者不负责关闭, 生成者由ctx.Done()关闭
// 消费者需要因为所有的通道关闭结束
func (e *Engine) close() (err error) {
	close(e.aiHistory)
	close(e.userHistory)
	close(e.outw)
	close(e.outv)
	close(e.stop)

	if err = e.ws.Close(); err != nil {
		log.Error("close ws err:", err)
	}
	if err = e.chatApp.Close(); err != nil {
		log.Error("close chat err:", err)
	}
	if err = e.ttsApp.Close(); err != nil {
		log.Error("close tts err:", err)
	}
	return
}

// analyse 风险分析
func analyse(text *string) {
	if strings.Contains(*text, "&") {
		if err := util.AlertEMail(); err != nil {
			log.Error("邮件发送失败", err)
		}
		*text = strings.Replace(*text, "&", " ", -1)
	}
}
