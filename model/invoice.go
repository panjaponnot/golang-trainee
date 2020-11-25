package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Invoice struct {
	Model
	SoRef         string           `json:"so_ref" gorm:"not null;"`
	InvoiceNo     string           `json:"invoice_no"`
	DocDate       time.Time        `json:"doc_date" gorm:"type:DATE"`
	InvoiceStatus *[]InvoiceStatus `json:"InvoiceStatus" gorm:"foreignkey:invoice_uid;association_foreignkey:uid` // 1 ,
}

type InvoiceStatus struct {
	Model
	InvoiceUid        uuid.UUID `json:"invoice_uid" gorm:"not null;"`
	SoRef             string    `json:"so_ref"`
	InvoiceStatusName string    `json:"inv_status_name"`
}

func (InvoiceStatus) TableName() string {
	return `invoice_status`
}
func (Invoice) TableName() string {
	return `invoice`
}
