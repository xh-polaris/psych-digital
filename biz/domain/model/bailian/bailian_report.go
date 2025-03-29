package bailian

import (
	"encoding/json"
	"fmt"
	"github.com/xh-polaris/psych-digital/biz/application/dto"
	"github.com/xh-polaris/psych-digital/biz/domain/model"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/util"
	"net/http"
	"strings"
	"sync"
)

var _ model.ReportApp = (*BLReportApp)(nil)

// BLReportApp 是阿里云报告分析大模型应用
// 单次对话, 无需管理上下文
type BLReportApp struct {
	appId  string
	apiKey string
	url    string
	header http.Header
	body   map[string]any
}

// NewBLReportApp 创建一个百炼报告分析模型应用实例
func NewBLReportApp(appId string, apiKey string) model.ReportApp {
	app := &BLReportApp{
		appId:  appId,
		apiKey: apiKey,
		url:    fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/apps/%s/completion", appId),
		header: http.Header{},
		body:   make(map[string]any),
	}

	// 初始化请求模板
	app.body["input"] = make(map[string]string)
	// 设置增量流式响应
	app.body["parameters"] = map[string]any{}

	// 设置请求头,其中X-DashScope-SSE设置为enable，表示开启流式响应
	app.header.Set("Authorization", "Bearer "+apiKey)
	app.header.Set("Content-Type", "application/json")

	return app
}

var instance model.ReportApp
var once sync.Once

// GetBLReportApp 获取百炼报告分析模型单例
func GetBLReportApp() model.ReportApp {
	once.Do(func() {
		c := config.GetConfig()
		instance = NewBLReportApp(c.BaiLianReport.AppId, c.BaiLianReport.ApiKey)
	})
	return instance
}

func (app *BLReportApp) Call(prompt string) (*dto.ChatReport, error) {
	var err error
	var report dto.ChatReport
	client := util.GetHttpClient()

	// 设置调用提示词
	app.body["input"].(map[string]string)["prompt"] = prompt
	res, err := client.Req(consts.Post, app.url, app.header, app.body)
	if err != nil {
		return nil, err
	}
	text, ok := res["output"].(map[string]any)["text"].(string)
	if !ok {
		return nil, nil
	}
	strings.Replace(text, "`", "", -1)
	err = json.Unmarshal([]byte(text), &report)
	return &report, err
}

// Close 释放相关资源
// BLChat暂时没有需要释放的资源
func (app *BLReportApp) Close() error {
	return nil
}
