package bailian

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/xh-polaris/psych-digital/biz/application/dto"
	"github.com/xh-polaris/psych-digital/biz/domain/model"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/util"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var _ model.ChatApp = (*BLChatApp)(nil)

// BLChatApp 是阿里云对话大模型应用
// 使用云端上下文管理，本地不管理聊天记录
type BLChatApp struct {
	appId  string
	apiKey string
	url    string
	header http.Header
	body   map[string]any
}

// NewBLChatApp 创建一个百炼模型应用实例
func NewBLChatApp(appId string, apiKey string) model.ChatApp {
	app := &BLChatApp{
		appId:  appId,
		apiKey: apiKey,
		url:    fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/apps/%s/completion", appId),
		header: http.Header{},
		body:   make(map[string]any),
	}

	// 初始化请求模板
	app.body["input"] = make(map[string]string)
	// 设置增量流式响应
	app.body["parameters"] = map[string]any{
		"incremental_output": true,
	}

	// 设置请求头,其中X-DashScope-SSE设置为enable，表示开启流式响应
	app.header.Set("Authorization", "Bearer "+apiKey)
	app.header.Set("Content-Type", "application/json")
	app.header.Set("X-DashScope-SSE", "enable")

	return app
}

// Call 非流式调用，暂时没用上
func (app *BLChatApp) Call(msg string) error {
	panic("implement me")
}

// StreamCall 流式调用
func (app *BLChatApp) StreamCall(msg string, sessionId string) (model.ChatAppScanner, error) {
	client := util.GetHttpClient()

	// 设置调用提示词
	app.body["input"].(map[string]string)["prompt"] = msg
	app.body["input"].(map[string]string)["session_id"] = sessionId

	// 获取流式响应reader
	reader, err := client.StreamReq(consts.Post, app.url, app.header, app.body)
	if err != nil {
		return nil, err
	}
	return newBLChatAppScanner(reader), nil
}

// Close 释放相关资源
// BLChat暂时没有需要释放的资源
func (app *BLChatApp) Close() error {
	return nil
}

// BLChatAppScanner 是百炼对话调用的响应
type BLChatAppScanner struct {
	closer  io.ReadCloser
	scanner *bufio.Scanner
}

// bLRawChatData 是百炼模型的原始响应
type bLRawChatData struct {
	Output struct {
		SessionId    string `json:"session_id"`
		FinishReason string `json:"finish_reason"`
		Text         string `json:"text"`
	} `json:"output"`

	Usage struct {
	} `json:"usage"`
}

// newBLChatAppScanner 创建一个新的大模型对话结果对象
// 非流式的可以模拟成一次返回然后io.EOF
func newBLChatAppScanner(r io.ReadCloser) *BLChatAppScanner {
	return &BLChatAppScanner{
		closer:  r,
		scanner: bufio.NewScanner(r),
	}
}

// Next 返回下一个读取到的对象或错误
func (s *BLChatAppScanner) Next() (*dto.ChatData, error) {
	var data dto.ChatData
	var err error

	for s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())
		// 跳过空行
		if line == "" {
			continue
		}

		switch {
		// 解析id
		case strings.HasPrefix(line, "id:"):
			if data.Id, err = strconv.ParseUint(strings.TrimPrefix(line, "id:"), 10, 64); err != nil {
				return nil, err
			}
		// 解析消息主体
		case strings.HasPrefix(line, "data:"):
			var raw bLRawChatData
			if err = json.Unmarshal([]byte(strings.TrimPrefix(line, "data:")), &raw); err != nil {
				return nil, err
			}
			data.SessionId = raw.Output.SessionId
			data.Content = raw.Output.Text
			data.Finish = raw.Output.FinishReason
			data.Timestamp = time.Now().Unix()
			return &data, nil
		}
	}

	if err = s.scanner.Err(); err != nil {
		return nil, err
	}

	// 没有更多内容
	return nil, io.EOF
}

// Close 释放资源
func (s *BLChatAppScanner) Close() error {
	return s.closer.Close()
}
