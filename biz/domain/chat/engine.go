package chat

import (
	"context"
	"errors"
	"github.com/hertz-contrib/websocket"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/application/dto"
	"github.com/xh-polaris/psych-digital/biz/domain/model"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
	"io"
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
	ws *WsHelper

	// rs 提供redis的读写功能 TODO: 此处用于测试，暂时不用redis
	// rs *RedisHelper
	rs *MemoryRedisHelper

	// sessionId 是本轮对话的唯一标记, 只有第一次调用时会写入, 应该不需要互斥锁
	sessionId string

	// chatApp 是调用的对话大模型
	chatApp model.ChatApp

	// aiHistory 记录AI输出历史
	aiHistory chan string

	// userHistory 记录用户输入历史
	userHistory chan string

	// outw ai的流式文本, 用于语音合成
	outw chan string

	// stop 用于打断AI输出
	stop chan bool
}

// NewChatEngine 初始化一个ChatEngine
func NewChatEngine(ctx context.Context, conn *websocket.Conn) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	// 暂时先固定为BaiLian之后类型多再换成工厂方法
	c := config.GetConfig()
	return &Engine{
		ctx:    ctx,
		cancel: cancel,
		ws:     NewWsHelper(conn),
		//rs:      NewRedisHelper(c),
		rs:          NewMemoryRedisHelper(),
		chatApp:     model.NewBLChatApp(c.BaiLian.AppId, c.BaiLian.ApiKey),
		aiHistory:   make(chan string, 10),
		userHistory: make(chan string, 10),
		outw:        make(chan string, 50),
		stop:        make(chan bool),
	}
}

// Start 开始一轮对话, 执行相关初始化
func (e *Engine) Start() error {
	var err error

	// 鉴权
	if !e.validate() {
		_ = e.ws.Error(consts.ErrInvalidUser)
		return consts.ErrInvalidUser
	}

	msg := "你好呀, 请问你是谁"
	// 音频生成
	go e.TTS()

	go e.StreamCall(msg)

	// 由于sessionId由第三方给出, 所以这里需要手动管理聊天记录的顺序
	// 等待获取sessionId, 初始化redis
	his := <-e.aiHistory
	if err = e.rs.Init(e.sessionId); err != nil {
		return err
	}
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

	err := e.ws.ReadJson(&startReq)
	if err != nil {
		log.Error("read json err:", err)
		return false
	}
	log.Info("调用方: %s, 调用时间: %s", startReq.From, time.Unix(startReq.Timestamp, 0).String())
	return true
}

// Chat 长对话的主体部分
func (e *Engine) Chat() {
	var req dto.ChatReq
	var err error

	defer func() {
		if err != nil {
			log.Error("chat err:", err)
		}
	}()

	// 处理聊天记录
	go e.History()

	for {
		// 获取前端对话内容
		err = e.ws.ReadJson(&req)
		if err != nil {
			return
		}

		// 判断是否结束
		if req.Cmd == consts.EndCmd {
			return
		}

		// 写入用户消息
		e.userHistory <- req.Msg

		// 读取用户输入
		go e.StreamCall(req.Msg)
	}
}

// StreamCall 调用chatApp并流式写入响应
func (e *Engine) StreamCall(msg string) {
	var record string
	var data *dto.ChatData

	// 流式响应的scanner
	scanner, err := e.chatApp.StreamCall(msg)
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
		// 用于打断, TODO: 效果不好的话就删掉
		case <-e.stop:
			err = e.ws.WriteJson(dto.StopData)
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

			// 写入文本, 用于音频合成
			e.outw <- data.Content

			// 写入响应 TODO: test待删除
			log.Info("data: ", data)
			err = e.ws.WriteJson(data)
			if err != nil {
				return
			}

			// 拼接聊天记录
			record += data.Content
		}
	}
}

// TTS 音频合成
func (e *Engine) TTS() {
	for {
		select {
		case <-e.ctx.Done():
			return
		// TODO 暂时用日志模拟
		case word := <-e.outw:
			log.Info("音频合成: ", word)
		}
	}
}

// History 处理聊天记录
func (e *Engine) History() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case his := <-e.aiHistory:
			err := e.rs.AddAi(e.sessionId, his)
			if err != nil {
				log.Error("ai history err:", err)
			}
		case his := <-e.userHistory:
			err := e.rs.AddUser(e.sessionId, his)
			if err != nil {
				log.Error("user history err:", err)
			}
		}
	}
}

// End 结束本轮对话
func (e *Engine) End() {
	defer func() { _ = e.close() }()

	err := e.ws.WriteJson(&dto.ChatEndResp{
		Code: 0,
		Msg:  "对话结束",
	})
	if err != nil {
		log.Error(err.Error())
		return
	}

	// 关闭所有协程
	e.cancel()
}

// close 释放相关资源
func (e *Engine) close() error {
	return e.ws.Close()
}
