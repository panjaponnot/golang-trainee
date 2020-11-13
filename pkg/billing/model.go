package billing

type Billing struct {
	token       string
	tokenType   string
	apiEndpoint string
}

// AttendantAPIResult struct
type APIResult struct {
	Error   interface{} `json:"error"`
	Message interface{} `json:"msg"`
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Count   int         `json:"count"`
}

type AuthenticationResult struct {
	Code     string `json:"code"`
	Expire   int    `json:"expire"`
	Token    string `json:"token"`
	Username string `json:"username"`
}
