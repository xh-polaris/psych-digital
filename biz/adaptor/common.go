package adaptor

import (
	"context"
	"errors"
	"github.com/hertz-contrib/websocket"
	"github.com/xh-polaris/gopkg/util"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	hertz "github.com/cloudwego/hertz/pkg/protocol/consts"
	bizerrors "github.com/xh-polaris/gopkg/errors"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/status"
)

type wsHandler func(ctx context.Context, conn *websocket.Conn)

// upgrader 默认配置的协议升级器, 用于将HTTP请求升级为WebSocket请求
var upgrader = websocket.HertzUpgrader{
	CheckOrigin: func(ctx *app.RequestContext) bool {
		return true
	},
}

// UpgradeWs 将Http协议升级为WebSocket协议
func UpgradeWs(ctx context.Context, c *app.RequestContext, handler wsHandler) error {
	// 尝试升级协议, 处理请求
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		handler(ctx, conn)
	})
	if err != nil {
		log.Error(err.Error())
		return consts.ErrWsUpgrade
	}
	return nil
}

var _ propagation.TextMapCarrier = &headerProvider{}

type headerProvider struct {
	headers *protocol.ResponseHeader
}

// Get a value from metadata by key
func (m *headerProvider) Get(key string) string {
	return m.headers.Get(key)
}

// Set a value to metadata by k/v
func (m *headerProvider) Set(key, value string) {
	m.headers.Set(key, value)
}

// Keys Iteratively get all keys of metadata
func (m *headerProvider) Keys() []string {
	out := make([]string, 0)

	m.headers.VisitAll(func(key, value []byte) {
		out = append(out, string(key))
	})

	return out
}

func PostProcess(ctx context.Context, c *app.RequestContext, req, resp any, err error) {
	log.CtxInfo(ctx, "[%s] req=%s, resp=%s, err=%v", c.Path(), util.JSONF(req), util.JSONF(resp), err)
	b3.New().Inject(ctx, &headerProvider{headers: &c.Response.Header})

	switch {
	case err == nil:
		c.JSON(hertz.StatusOK, resp)
	case errors.Is(err, consts.ErrForbidden):
		c.JSON(hertz.StatusForbidden, err.Error())
	default:
		if s, ok := status.FromError(err); ok {
			c.JSON(http.StatusOK, &bizerrors.BizError{
				Code: uint32(s.Code()),
				Msg:  s.Message(),
			})
		} else {
			log.CtxError(ctx, "internal error, err=%s", err.Error())
			code := hertz.StatusInternalServerError
			c.String(code, hertz.StatusMessage(code))
		}
	}
}
