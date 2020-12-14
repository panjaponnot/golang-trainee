package model

import "time"

type QuotationLog struct {
	Model
	UserName       string `json:"user_name"`
	Date           string `json:"date"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
	Reason         string `json:"reason"`
	OneId          string `json:"one_id"`
	StaffId        string `json:"staff_id"`
	DocNumberEfrom string `json:"doc_number_eform" gorm:"column:doc_number_eform"`
}

type SaleApprove struct {
	Id             int       `gorm:"column:id_increment; primary_key;"`
	Reason         string    `gorm:"column:reason"`
	Remark         string    `gorm:"column:remark"`
	DocNumberEfrom string    `gorm:"column:doc_number_eform"`
	Status         string    `gorm:"column:status"`
	CreateAt       time.Time `gorm:"column:create_at"`
}

func (SaleApprove) TableName() string {
	return "sales_approve"
}
