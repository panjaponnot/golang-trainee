package model

type OrgChart struct {
	StaffId      string  `json:"staff_id"`
	Fname        string  `json:"fname"`
	Lname        string  `json:"lname"`
	Nname        string  `json:"nname"`
	Position     string  `json:"position"`
	Department   string  `json:"department"`
	StaffChild   string  `json:"staff_child"`
	InvAmount    float64 `json:"inv_amount"`
	InvAmountOld float64 `json:"inv_amount_old"`
	GoalTotal    float64 `json:"goal_total"`
	ScoreTarget  float64 `json:"score_target"`
	ScoreSf      float64 `json:"score_sf"`
	SaleFactor   float64 `json:"sale_factor"`
	TotalSo      float64 `json:"total_so"`
	IfFactor     float64 `json:"if_factor"`
	EngCost      float64 `json:"engcost" gorm:"column:engcost"`
	Revenue      float64 `json:"revanue"`
	ScoreIf      float64 `json:"score_if"`
	InFactor     float64 `json:"in_factor"`
	OneId        string  `json:"one_id"`
	Image        string  `json:"image"`
	FileName     string  `json:"filename" gorm:"column:filename"`
	GrowthRate   float64 `json:"growth_rate"`
	ScoreGrowth  string  `json:"score_growth"`
	ScoreAll     float64 `json:"score_all"`
	Quarter      string  `json:"quarter"`
	Year         float64 `json:"year"`
	JobMonths    int     `json:"job_months"`
}

type InvBefore struct {
	StaffID   string  `json:"staff_id"`
	InvAmount float64 `json:"inv_amount"`
	CheckData int     `json:"check_data" gorm:"column:checkdata"`
}
