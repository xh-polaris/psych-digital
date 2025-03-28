package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// GzipCompress 按照gzip的方式压缩
func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	var err error
	gz := gzip.NewWriter(&buf)
	defer func() { _ = gz.Close() }()

	_, err = gz.Write(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
