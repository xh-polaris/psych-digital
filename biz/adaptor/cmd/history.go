package cmd

type ListHistoryReq struct {
	Paging Paging `json:"paging"`
}

type ListHistoryResp struct {
	Code    int64      `json:"code"`
	Msg     string     `json:"msg"`
	History []*History `json:"history"`
	Total   int64      `json:"total"`
}

// History 聊天记录与报表
type History struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name"`
	Class     string    `json:"class"`
	StudentId string    `json:"student_id"`
	Dialogs   []*Dialog `json:"dialogs"`
	Report    *Report   `json:"report"`
	StartTime int64     `json:"start_time"`
	EndTime   int64     `json:"end_time"`
}

type Dialog struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Report struct {
	Keywords   []string `json:"keywords"`
	Type       []string `json:"type"`
	Content    string   `json:"content"`
	Grade      string   `json:"grade"`
	Suggestion []string `json:"suggestion"`
}
