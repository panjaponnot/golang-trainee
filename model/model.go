package model

import (
	"time"

	"github.com/jinzhu/gorm"

	uuid "github.com/satori/go.uuid"
)

// Model struct
type Model struct {
	Seq       int64      `json:"seq" gorm:"primary_key;auto_increment:false;" sql:"index"`
	Uid       uuid.UUID  `json:"uid" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

type Result struct {
	Error   interface{} `json:"error,omitempty"`
	Message interface{} `json:"msg,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Total   int         `json:"total,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// BeforeCreate hook table
func (m *Model) BeforeCreate(scope *gorm.Scope) error {
	if f, ok := scope.FieldByName("uid"); ok {
		if f.IsBlank {
			if err := scope.SetColumn("uid", uuid.NewV4()); err != nil {
				return err
			}
		}
	}
	return scope.SetColumn("seq", time.Now().UnixNano())
}
