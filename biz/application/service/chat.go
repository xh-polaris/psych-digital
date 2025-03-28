package service

import (
	"context"
	"github.com/hertz-contrib/websocket"
	"github.com/xh-polaris/psych-digital/biz/domain/chat"
)

// ChatHandler 处理长对话 TODO: 应该需要加上超时处理，避免连接空置太长时间
func ChatHandler(ctx context.Context, conn *websocket.Conn) {
	var err error

	// 初始化本轮对话的engine
	engine := chat.NewEngine(ctx, conn)
	defer func() { engine.Close() }()

	// 执行初始化操作
	err = engine.Start()
	if err != nil {
		return
	}

	engine.Chat()
}
