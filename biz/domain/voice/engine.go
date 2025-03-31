package voice

import (
	"github.com/hertz-contrib/websocket"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/application/dto"
	"github.com/xh-polaris/psych-digital/biz/domain"
	"github.com/xh-polaris/psych-digital/biz/domain/model"
	"github.com/xh-polaris/psych-digital/biz/domain/model/volc"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"golang.org/x/net/context"
	"time"
)

type Engine struct {
	// ctx 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// ws 管理ws连接
	ws *domain.WsHelper

	// asrApp 语音识别app
	asrApp model.AsrApp

	// finish 结束
	finish chan struct{}
}

// NewEngine 初始化
func NewEngine(ctx context.Context, conn *websocket.Conn) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	c := config.GetConfig()
	e := &Engine{
		ctx:    ctx,
		cancel: cancel,
		ws:     domain.NewWsHelper(conn),
		asrApp: volc.NewVcAsrApp(c.VolcAsr.AppKey, c.VolcAsr.AccessKey, c.VolcAsr.ResourceId, c.VolcAsr.Url),
		finish: make(chan struct{}),
	}
	return e
}

// Start 初始化
func (e *Engine) Start() error {
	if err := e.asrApp.Dial(); err != nil {
		return err
	}
	if err := e.asrApp.Start(); err != nil {
		return err
	}
	return nil
}

// Listen 主事件循环, 获取前端的音频流输入, 返回文字
func (e *Engine) Listen() {
	go e.listen()
	go e.recognise()
	<-e.finish
}

// recognise 识别音频并写入输入
func (e *Engine) recognise() {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			// 获取响应并写入ws
			text, err := e.asrApp.Receive()
			if err != nil {
				e.finish <- struct{}{}
				return
			}
			resp := &dto.AsrResp{
				Text:      text,
				Timestamp: time.Now().Unix(),
			}
			if err = e.ws.WriteJSON(resp); err != nil {
				e.finish <- struct{}{}
				return
			}
		}
	}
}

// listen 获取音频输入并发送给asr #生产者
func (e *Engine) listen() {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
			data, err := e.ws.ReadBytes()
			if err != nil {
				log.Error("listen audio err: ", err)
			} else if data == nil || len(data) == 0 {
				continue
			}
			if err = e.asrApp.Send(data); err != nil {
				e.finish <- struct{}{}
				return
			}
		}
	}
}

// Close 释放资源
func (e *Engine) Close() error {
	e.cancel()
	return e.ws.Close()
}
