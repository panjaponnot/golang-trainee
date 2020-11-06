package model

import (
	"bytes"
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

type StaffImg struct {
	StaffId string `json:"staff_id"`
	Prefix  string `json:"prefix"`
	Fname   string `json:"fname"`
	Lname   string `json:"lname"`
	Nname   string `json:"nname"`
	// StaffImage string `json:"staff_image"`
	StaffImage []byte `json:"staff_image"`
	Img        string `json:"img"`
}

type StaffId struct {
	StaffId string `json:"staff_id"`
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

type StaffProfile struct {
	StaffId  string           `json:"staff_id"`
	Prefix   string           `json:"prefix"`
	Fname    string           `json:"fname"`
	Lname    string           `json:"lname"`
	Nname    string           `json:"nname"`
	Position string           `json:"position"`
	Mail     StaffMail        `json:"mail"`
	OneMail  StaffOneMail     `json:"onemail"`
	Tel      StaffTel         `json:"tel"`
	Month    StaffGoalMonth   `json:"goalmonth"`
	Quarter  StaffGoalQuarter `json:"goalquarter"`
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
	Id       string `json:"id"`
	RefStaff string `json:"ref_staff"`
	Mail     string `json:"mail"`
}

type StaffOneMail struct {
	Id       string `json:"id"`
	RefStaff string `json:"ref_staff"`
	Onemail  string `json:"onemail"`
}

type StaffTel struct {
	Id       string `json:"id"`
	RefStaff string `json:"ref_staff"`
	Tel      string `json:"tel"`
	TelSup   string `json:"tel_sup"`
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

type StaffPicture struct {
	StaffId string       `json:"staff_id"`
	Img     bytes.Buffer `json:"staff_img" gorm:"column:staff_img"`
}

type StaffGoalMonth struct {
	Id         string `json:"id"`
	RefStaff   string `json:"ref_staff"`
	Year       string `json:"year"`
	Month      string `json:"month"`
	GoalTotal  string `json:"goal_total"`
	RealTotal  string `json:"real_total"`
	CreateDate string `json:"create_date"`
	CreateBy   string `json:"create_by"`
}

type StaffGoalQuarter struct {
	Id         string `json:"id"`
	RefStaff   string `json:"ref_staff"`
	Year       string `json:"year"`
	Quarter    string `json:"quarter"`
	GoalTotal  string `json:"goal_total"`
	RealTotal  string `json:"real_total"`
	CreateDate string `json:"create_date"`
	CreateBy   string `json:"create_by"`
}

type GroupRelation struct {
	IdGroup      string `json:"id_group"`
	IdGroupChild string `json:"id_group_child"`
}

type StaffGroupRelation struct {
	IdStaff string `json:"id_staff"`
}
