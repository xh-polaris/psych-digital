package volc

import (
	"context"
	"io"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	testAsrURL        = "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel"
	testAsrAppKey     = "8390064657"
	testAsrAccessKey  = "4Y-BEHltDMMGtnEFg85xdiifFsGlGlBS"
	testAsrResourceId = "volc.bigasr.sauc.duration"

	inputFile  = "output.pcm" // 16位单声道PCM文件
	outputFile = "output.txt" // 识别结果保存路径
)

// TestASRStreaming 流式语音识别测试
func TestASRStreaming(t *testing.T) {
	// 1. 初始化ASR客户端
	asrApp := NewVcAsrApp(testAsrAppKey, testAsrAccessKey, testAsrResourceId, testAsrURL)

	// 2. 建立连接
	if err := asrApp.Dial(); err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 3. 启动会话
	if err := asrApp.Start(); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}

	// 4. 打开音频文件
	file, err := os.Open(inputFile)
	if err != nil {
		t.Fatalf("无法打开音频文件: %v", err)
	}
	defer file.Close()

	// 5. 创建结果文件
	resultFile, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("无法创建结果文件: %v", err)
	}
	defer resultFile.Close()

	// 6. 使用WaitGroup协调goroutine
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// 发送协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		sendAudio(ctx, t, asrApp, file)
	}()

	// 接收协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveResults(ctx, t, asrApp, resultFile)
	}()

	// 7. 设置超时机制
	go func() {
		<-time.After(60 * time.Second)
		cancel()
		t.Error("测试超时")
	}()

	// 8. 等待任务完成
	wg.Wait()
	cancel()
}

// sendAudio 发送音频数据
func sendAudio(ctx context.Context, t *testing.T, app *VcAsrApp, file *os.File) {
	buf := make([]byte, 3200) // 每次发送3200字节（约200ms 16kHz音频）

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := file.Read(buf)
			if err == io.EOF {
				t.Log("音频发送完成")
				return
			}
			if err != nil {
				t.Errorf("读取音频失败: %v", err)
				return
			}

			if err := app.Send(buf[:n]); err != nil {
				t.Errorf("发送失败: %v", err)
				return
			}
			time.Sleep(200 * time.Millisecond) // 模拟实时流
		}
	}
}

// receiveResults 接收识别结果
func receiveResults(ctx context.Context, t *testing.T, app *VcAsrApp, output *os.File) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := app.Receive()
			if err != nil {
				if err == io.EOF {
					t.Log("连接正常关闭")
					return
				}
				t.Errorf("接收错误: %v", err)
				return
			}

			if len(res) > 0 {
				log.Printf("识别结果: %s", res)
				if _, err := output.WriteString(string(res) + "\n"); err != nil {
					t.Errorf("写入结果失败: %v", err)
				}
			}
		}
	}
}
