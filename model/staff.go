package model

import (
	"time"
)

type StaffInfo struct {
	OneId      string `json:"one_id"`
	StaffId    string `json:"staff_id"`
	Prefix     string `json:"prefix"`
	Fname      string `json:"fname"`
	Lname      string `json:"lname"`
	Nname      string `json:"nname"`
	Position   string `json:"position"`
	Department string `json:"department"`
	StaffChild string `json:"staff_child"`
}

type UserInfo struct {
	OneId      string    `json:"one_id"`
	StaffId    string    `json:"staff_id"`
	Username   string    `json:"username"`
	OneMail    string    `json:"onemail"`
	Role       string    `json:"role"`
	Comment    string    `json:"comment"`
	CreateDate time.Time `json:"create_date"`
}
