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
}

// ChatAppScanner 是第三方对话调用的响应
type ChatAppScanner interface {
	Next() (*dto.ChatData, error)
	Close() error
}
