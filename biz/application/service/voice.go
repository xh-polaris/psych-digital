package service

import (
	"github.com/hertz-contrib/websocket"
	"golang.org/x/net/context"
)

// AsrHandler 通用音频识别 TODO: 应该需要加上超时处理，避免连接空置太长时间
func AsrHandler(ctx context.Context, conn *websocket.Conn) {
	panic("implement me")
}
