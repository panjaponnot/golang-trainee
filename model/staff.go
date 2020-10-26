package model

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
