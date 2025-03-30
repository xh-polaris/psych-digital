package cmd

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Paging struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
