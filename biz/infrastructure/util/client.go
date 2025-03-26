package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"io"
	"log"
	"net/http"
)

var client *HttpClient

// HttpClient 是一个简单的 HTTP 客户端
type HttpClient struct {
	Client *http.Client
	Config *config.Config
}

// NewHttpClient 创建一个新的 HttpClient 实例
func NewHttpClient() *HttpClient {
	return &HttpClient{
		Client: &http.Client{},
	}
}

// GetHttpClient 获取客户端单例
func GetHttpClient() *HttpClient {
	if client == nil {
		client = NewHttpClient()
	}
	return client
}

// Req 发送 HTTP 请求
func (c *HttpClient) Req(method, url string, headers http.Header, body interface{}) (map[string]interface{}, error) {
	resp, err := c.do(method, url, headers, body)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("关闭请求失败: %v", closeErr)
		}
	}()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_resp, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("unexpected status code: %d, response body: %s", resp.StatusCode, _resp)
		return nil, fmt.Errorf(errMsg)
	}

	// 读取响应
	_resp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 反序列化响应体
	var respMap map[string]interface{}
	if err := json.Unmarshal(_resp, &respMap); err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %w", err)
	}

	return respMap, nil
}

// StreamReq 流式响应的请求
func (c *HttpClient) StreamReq(method, url string, headers http.Header, body interface{}) (*StreamReader, error) {

	resp, err := c.do(method, url, headers, body)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	reader := &StreamReader{
		resp:   resp,
		reader: resp.Body,
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = reader.Close() }()
		_resp, _ := reader.ReadAll()
		errMsg := fmt.Sprintf("unexpected status code: %d, response body: %s", resp.StatusCode, _resp)
		return nil, fmt.Errorf(errMsg)
	}

	return reader, nil
}

// do 实际执行请求
func (c *HttpClient) do(method, url string, headers http.Header, body interface{}) (*http.Response, error) {
	// 将 body 序列化为 JSON
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("请求体序列化失败: %w", err)
	}

	// 创建新的请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := c.Client.Do(req)

	return resp, err
}

// StreamReader 流式请求Reader, 封装是为了避免只返回reader时无法关闭resp.Body
// 调用方需要负责将流关闭
type StreamReader struct {
	resp   *http.Response
	reader io.ReadCloser
}

// Read 从Reader中读取
func (r *StreamReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

// ReadAll 读取所有的
func (r *StreamReader) ReadAll() ([]byte, error) {
	return io.ReadAll(r.reader)
}

// Close 关闭resp.Body
func (r *StreamReader) Close() error {
	return r.resp.Body.Close()
}
