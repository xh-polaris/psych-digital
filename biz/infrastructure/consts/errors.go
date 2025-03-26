package consts

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Errno struct {
	err  error
	code codes.Code
}

// GRPCStatus 实现 GRPCStatus 方法
func (en *Errno) GRPCStatus() *status.Status {
	return status.New(en.code, en.err.Error())
}

// 实现 Error 方法
func (en *Errno) Error() string {
	return en.err.Error()
}

// Code 返回错误码
func (en *Errno) Code() int { return int(en.code) }

// NewErrno 创建自定义错误
func NewErrno(code codes.Code, err error) *Errno {
	return &Errno{
		err:  err,
		code: code,
	}
}

// 定义常量错误
var (
	ErrForbidden   = NewErrno(codes.PermissionDenied, errors.New("forbidden"))
	ErrWsUpgrade   = NewErrno(codes.Code(1000), errors.New("websocket协议升级失败"))
	ErrInvalidUser = NewErrno(codes.Code(1001), errors.New("非授权用户，请重试或切换账号"))
)
