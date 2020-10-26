package model

// User struct
type User struct {
	Model
	AccountId   string        `json:"account_id"`
	Status      string        `json:"status"`
	Permissions []*Permission `json:"permission" gorm:"foreignkey:account_id;association_foreignkey:account_id"`
}

// Role struct
type Role struct {
	Model
	RoleName string `json:"role_name"`
	Level    string `json:"level"`
	Detail   string `json:"detail"`
}

// Permission struct
type Permission struct {
	Model
	AccountId string   `json:"account_id"`
	TaxId     string   `json:"tax_id"`
	RoleUId   string   `json:"-"`
	User      *User    `json:"user" gorm:"foreignkey:account_id;association_foreignkey:account_id"`
	Company   *Company `json:"company" gorm:"foreignkey:tax_id;association_foreignkey:tax_id"`
	Role      *Role    `json:"role" gorm:"foreignkey:uid;association_foreignkey:role_uid"`
}

// UserActivityLog struct
type UserActivityLog struct {
	Model
	AccountId string `json:"account_id"`
	Status    string `json:"status"`
	Detail    string `json:"detail"`
}
