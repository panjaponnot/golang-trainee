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

type StaffAll struct {
	StaffId    string `json:"staff_id"`
	Fname      string `json:"fname"`
	Lname      string `json:"lname"`
	Nname      string `json:"nname"`
	StaffInfo  string `json:"staff_info"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Resign     string `json:"resign"`
	Available  string `json:"available"`
	Comment    string `json:"comment"`
	Position   string `json:"position"`
	Division   string `json:"division"`
	Department string `json:"department"`
}

type Staff struct {
	StaffId string `json:"staff_id"`
	Fname   string `json:"fname"`
	Lname   string `json:"lname"`
	Nname   string `json:"nname"`
	// StaffInfo  string `json:"staff_info"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Resign    string `json:"resign"`
	Available string `json:"available"`
	Comment   string `json:"comment"`
	// 	Position   string `json:"position"`
	// 	Division   string `json:"division"`
	// 	Department string `json:"department"`
	Inuse    string        `json:"inuse"`
	Email    StaffMail     `json:"email"`
	Tel      StaffTel      `json:"tel"`
	Position StaffPosition `json:"position"`
	Ability  StaffAbility  `json:"ability"`
}

type StaffMail struct {
	Id      string `json:"id"`
	Email   string `json:"email"`
	Comment string `json:"comment"`
	Status  string `json:"status"`
}

type StaffTel struct {
	Id      string `json:"id"`
	Tel     string `json:"tel"`
	TelSup  string `json:"tel_sup"`
	Comment string `json:"comment"`
	Status  string `json:"status"`
}

type StaffPosition struct {
	Id         string `json:"id"`
	Position   string `json:"position"`
	Division   string `json:"division"`
	Department string `json:"department"`
	StartDate  string `json:"start_date"`
	Comment    string `json:"comment"`
	Status     string `json:"status"`
}

type StaffAbility struct {
	Id      string `json:"id"`
	Skill   string `json:"skill"`
	Mark    string `json:"mark"`
	Comment string `json:"comment"`
	Status  string `json:"status"`
}
