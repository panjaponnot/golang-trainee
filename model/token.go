package model

type ApiClient struct {
	Model
	Name string `json:"name" gorm:"unique;not null"`
}
