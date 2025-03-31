package util

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/xh-polaris/psych-digital/biz/adaptor/cmd"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"io"
	"log"
	"net/smtp"
	"strconv"
)

// GzipCompress 按照gzip的方式压缩
func GzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write(data)
	_ = w.Close()
	return b.Bytes(), nil
}

// GzipDecompress 解压
func GzipDecompress(src []byte) ([]byte, error) {
	// 1. 空数据检查
	if len(src) == 0 {
		return nil, nil
	}

	// 2. 创建GZIP读取器
	r, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("创建解压器失败: %w", err)
	}
	defer func() { _ = r.Close() }()

	// 3. 读取解压数据
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("解压数据读取失败: %w", err)
	}

	// 4. 返回解压结果
	return buf.Bytes(), nil
}

// IntToBytes 将整数变成字节数组
func IntToBytes(n int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return b
}

// BytesToInt 将字节数组变成整数
func BytesToInt(data []byte) (int, error) {
	if len(data) != 4 || data == nil {
		return 0, fmt.Errorf("BytesToInt err")
	}
	return int(binary.BigEndian.Uint32(data)), nil
}

// FailOnError 出现异常时中止
func FailOnError(msg string, err error) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}

// ParsePaging 解析分页参数
func ParsePaging(p *cmd.Paging) (skip, limit int64) {
	// 设置分页参数
	skip = int64((p.Page - 1) * p.Limit)
	limit = int64(p.Limit)
	return skip, limit
}

// AlertEMail 发送邮件shallwii@126.com
func AlertEMail() (err error) {
	//c := config.GetConfig().SMTP
	c := &config.SMTP{
		Username: "bot@xhpolaris.com",
		Password: "XHpolaris123",
		Host:     "smtp.feishu.cn",
		Port:     587,
	}
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
	err = smtp.SendMail(c.Host+":"+strconv.Itoa(c.Port), auth, c.Username, []string{"1449610641@qq.com"}, []byte("检测到心理空间出现一位高风险学生，请立即前往处理"))
	return err
}
