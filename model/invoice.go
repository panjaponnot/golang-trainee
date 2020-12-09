package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Invoice struct {
	Model
	SoRef     string          `json:"so_ref" gorm:"not null;"`
	InvoiceNo string          `json:"invoice_no"`
	DocDate   time.Time       `json:"doc_date" gorm:"type:DATE"`
	InvStatus []InvoiceStatus `json:"invoice_status" gorm:"foreignkey:invoice_uid;association_foreignkey:uid"` // 1 ,
}

type InvoiceStatus struct {
	Model
	InvoiceUid        uuid.UUID `json:"invoice_uid" gorm:"not null;"`
	SoRef             string    `json:"so_ref"`
	InvNo             string    `json:"inv_no"`
	InvoiceStatusName string    `json:"inv_status_name"`
}

func (InvoiceStatus) TableName() string {
	return `invoice_status`
}
func (Invoice) TableName() string {
	return `invoice`
}
