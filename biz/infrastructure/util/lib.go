package util

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
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
