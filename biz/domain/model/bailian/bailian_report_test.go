package bailian

import (
	"fmt"
	"testing"
)

func TestBaiLianReportApp_Call(t *testing.T) {
	const (
		appId  = "0e08c61661e143ae85b63ee61ec07b69"    // 替换实际值
		apiKey = "sk-02654c3231f54c90b3500a1b75003e5f" // 替换实际值
	)

	// 创建应用实例
	app := NewBLReportApp(appId, apiKey)
	defer func() { _ = app.Close() }()

	// 完整对话文本（注意保留换行符）
	msg := `学生:我是四(1)班的陈思雨...最近一写数学作业就心跳特别快
AI:思雨同学，这种身体反应是从什么时候开始的呢?(提取班级信息)
学生:上周模拟考之后..明明复习了还是错了好多题(揉眼睛)
AI:听起来你对自己的要求很高，可以具体说说错题的类型吗?(压力源分析)
学生:应用题全错了..现在看到林小美就难受，她考了满分还不理我
AI:好朋友的态度变化确实让人困扰，你们之前有发生过什么不愉快吗?(社会支持评估)
学生:她说我总问问题很烦….晚上翻来要去睡不着，妈妈又天天加班
AI:这些压力交织在一起确实不容易，我们试试用呼吸法缓解紧张好吗?`

	// 调用大模型进行综合分析
	resp, err := app.Call(msg)
	if err != nil {
		fmt.Printf("API调用失败: %v\n", err)
		return
	}
	// 打印结构化分析结果
	fmt.Println("【心理评估报告】")
	fmt.Println(resp)
}
