package dto

type (
	// ChatStartReq 开始对话请求
	ChatStartReq struct {
		// 开始的时间戳
		Timestamp int64 `json:"timestamp"`
		// 使用者标记
		From string `json:"from"`
		// 单位id
		UnitId string `json:"unitId"`
		// 学号
		StudentId string `json:"studentId"`
		// 密码
		Password string `json:"password"`
	}

	// ChatReq 对话请求
	ChatReq struct {
		// 命令, 0对话, -1结束
		Cmd int64  `json:"cmd"`
		Msg string `json:"msg"`
	}

	// ChatEndResp 对话结束响应
	ChatEndResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	// ChatData 一次流式响应
	ChatData struct {
		Id        uint64 `json:"id"`
		Content   string `json:"content"`
		SessionId string `json:"session_id"`
		Timestamp int64  `json:"timestamp"`
		Finish    string `json:"finish"`
	}

	// ChatHistory 对话记录
	ChatHistory struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	// ChatReport 对话分析报告
	ChatReport struct {
		Name   string `json:"name"`
		Class  string `json:"class"`
		Report struct {
			Keywords   []string `json:"keywords"`
			Type       []string `json:"type"`
			Content    string   `json:"content"`
			Grade      string   `json:"grade"`
			Suggestion []string `json:"suggestion"`
		} `json:"report"`
	}
)
