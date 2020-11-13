package model

type QuotationLog struct {
	Model
	UserName       string `json:"user_name"`
	Date           string `json:"date"`
	SoNumber       string `json:"so_number"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
	OneId          string `json:"one_id"`
	StaffId        string `json:"staff_id"`
	DocNumberEfrom string `json:"doc_number_eform" gorm:"column:doc_number_eform"`
}
