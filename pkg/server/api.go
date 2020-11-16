package server

type Result struct {
	Error   interface{} `json:"error,omitempty"`
	Message interface{} `json:"msg,omitempty"`
	Mess    interface{} `json:"messages,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Total   int         `json:"total,omitempty"`
	Count   int         `json:"count,omitempty"`
}
