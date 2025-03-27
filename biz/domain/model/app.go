package model

import (
	"github.com/xh-polaris/psych-digital/biz/application/dto"
)

// ChatApp 是第三方对话大模型应用的抽象
type ChatApp interface {
	// Call 整体调用
	Call(msg string)

	// StreamCall 流式调用, 默认应该采用增量输出, 即后续的输出不包括之前的输出
	StreamCall(msg string) (ChatAppScanner, error)

	// Close 关闭资源
	Close() error
}

// ChatAppScanner 是第三方对话调用的响应
type ChatAppScanner interface {
	Next() (*dto.ChatData, error)
	Close() error
}

// TtsApp 是第三方语音合成大模型的抽象
type TtsApp interface {
	// Dial 建立ws连接
	Dial() error

	// Start 建立application级连接
	Start() error

	// Send 发送文字请求
	Send(texts string) error

	// Receive 接受音频流响应
	Receive() []byte

	// Close 断开连接, 释放资源
	Close() error
}

// AsrApp 是第三方通用语音识别的抽象
type AsrApp interface {
	// Dial 建立ws连接
	Dial() error

	// Start 建立application级连接
	Start() error

	// Send 发送音频流
	Send(bytes []byte) error

	// Receive 接受文字
	Receive() string

	// Close  关闭连接, 释放资源
	Close() error
}
