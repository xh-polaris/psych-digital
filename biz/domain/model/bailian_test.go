package model

import (
	"fmt"
	"io"
	"testing"
)

func TestBaiLianChatApp_StreamCall(t *testing.T) {
	app := NewBLChatApp("d37840a0f7d6490f87952dd3ca0bb441", "sk-02654c3231f54c90b3500a1b75003e5f")
	scanner, err := app.StreamCall("你好")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = scanner.Close() }()

	for {
		data, err := scanner.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		fmt.Println(data)
	}
}
