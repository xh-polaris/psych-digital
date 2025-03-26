package chat

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/application/service"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
)

// upgrader 默认配置的协议升级器, 用于将HTTP请求升级为WebSocket请求
var upgrader = websocket.HertzUpgrader{}

// LongChat 开启一轮长对话
// @router /chat/ [GET]
func LongChat(ctx context.Context, c *app.RequestContext) {
	// 尝试升级协议, 并处理
	err := upgradeWs(ctx, c, service.ChatHandler)
	if err != nil {
		log.Error(err.Error())
	}
}

// 将Http协议升级为WebSocket协议
func upgradeWs(ctx context.Context, c *app.RequestContext, handler service.WsHandler) error {
	// 尝试升级协议, 处理请求
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		handler(ctx, conn)
	})
	if err != nil {
		return consts.ErrWsUpgrade
	}
	return nil
}
