package volc

import (
	"os"
	"testing"
	"time"
)

// 测试前请替换以下参数为有效值
const (
	testTtsUrl        = "wss://openspeech.bytedance.com/api/v3/tts/bidirection"
	testTtsAppKey     = "8390064657"
	testTtsAccessKey  = "4Y-BEHltDMMGtnEFg85xdiifFsGlGlBS"
	testTtsSpeaker    = "zh_female_roumeinvyou_emo_v2_mars_bigtts"
	testTtsResourceId = "volc.service_type.10029"
)

func TestTTSGeneration(t *testing.T) {
	// 初始化TTS应用
	app := NewVcTtsApp(
		testTtsAppKey,
		testTtsAccessKey,
		testTtsSpeaker,
		testTtsResourceId,
		testTtsUrl,
	)

	// 建立连接
	if err := app.Dial(); err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer app.Close()

	// 启动会话
	if err := app.Start(); err != nil {
		t.Fatalf("会话启动失败: %v", err)
	}

	// 创建输出文件
	outputFile, err := os.Create("./output.pcm")
	if err != nil {
		t.Fatalf("创建文件失败: %v", err)
	}
	defer outputFile.Close()

	var audioData []byte

	go func() {
		// 发送要合成的文本
		t.Log("start at: ", time.Now().String())
		testText := []string{"你好呀", "小朋友", "我是张老师", "很高兴你能来我聊天。", "我能知道你叫什么名字吗?"}
		for _, text := range testText {
			if err := app.Send(text); err != nil {
				t.Errorf("文本发送失败: %v", err)
				return
			}
		}
		time.Sleep(10 * time.Second)
		app.Close()
	}()

	for {
		data := app.Receive()
		if data == nil || len(data) == 0 {
			break
		}
		t.Logf("get a data with len: %d, at %s ", len(data), time.Now().String())
		audioData = append(audioData, data...)
	}

	// 写入文件
	if _, err := outputFile.Write(audioData); err != nil {
		t.Fatalf("文件写入失败: %v", err)
	}

	t.Logf("写入完成")
	// 验证文件内容
	fileInfo, err := outputFile.Stat()
	if err != nil {
		t.Fatalf("文件校验失败: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Fatal("生成的音频文件为空")
	}

	t.Logf("成功生成音频文件，大小: %d 字节", fileInfo.Size())
}
