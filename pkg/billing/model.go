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
	Code     int    `json:"code"`
	Expire   string `json:"expire"`
	Token    string `json:"token"`
	Username string `json:"username"`
}

type InvoiceSO struct {
	Total string          `json:"total"`
	Data  []DataInvoiceSO `json:"data"`
}

type InvoiceStatus struct {
	InvStatusname string `json:"inv_status_name"`
}

type DataInvoiceSO struct {
	SoRef         string          `json:"SoRef"`
	InvoiceNo     string          `json:"invoice_no"`
	DocDate       string          `json:"doc_date"`
	InvoiceStatus []InvoiceStatus `json:"invoice_status"`
}
