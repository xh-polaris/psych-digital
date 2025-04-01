package chat

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/xh-polaris/psych-digital/biz/adaptor"
	"github.com/xh-polaris/psych-digital/biz/adaptor/cmd"
	"github.com/xh-polaris/psych-digital/provider"
)

// ListHistory .
// @router /chat/history/list [GET]
func ListHistory(ctx context.Context, c *app.RequestContext) {
	var err error
	var req cmd.ListHistoryReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	p := provider.Get()
	resp, err := p.HistoryService.ListHistory(ctx, &req)
	adaptor.PostProcess(ctx, c, &req, resp, err)
}
