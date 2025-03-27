package voice

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/adaptor"
	"github.com/xh-polaris/psych-digital/biz/application/service"
)

// Asr 通用语音识别
// @router /voice/asr [GET]
func Asr(ctx context.Context, c *app.RequestContext) {
	// 尝试升级协议
	err := adaptor.UpgradeWs(ctx, c, service.ChatHandler)
	if err != nil {
		log.Error(err.Error())
	}
}
