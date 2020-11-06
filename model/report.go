package model

type OrgChart struct {
	StaffId      string  `json:"staff_id"`
	Fname        string  `json:"fname" gorm:"column:fname"`
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
	ScoreGrowth  float64 `json:"score_growth"`
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

type SummaryCustomer struct {
	CustomerID        string  `json:"customer_id" gorm:"column:Customer_ID"`
	CustomerName      string  `json:"customer_name" gorm:"column:Customer_Name"`
	TotalSo           float64 `json:"total_so" gorm:"column:total_so"`
	TotalCs           float64 `json:"total_cs" gorm:"column:total_cs"`
	TotalInv          float64 `json:"total_inv" gorm:"column:total_inv"`
	TotalRc           float64 `json:"total_rc" gorm:"column:total_rc"`
	TotalCn           float64 `json:"total_cn" gorm:"column:total_cn"`
	SoAmount          float64 `json:"so_amount" gorm:"column:so_amount"`
	InvAmount         float64 `json:"inv_amount" gorm:"column:inv_amount"`
	CsAmount          float64 `json:"cs_amount" gorm:"column:cs_amount"`
	RcAmount          float64 `json:"rc_amount" gorm:"column:rc_amount"`
	CnAmount          float64 `json:"cn_amount" gorm:"column:cn_amount"`
	Amount            float64 `json:"amount" gorm:"column:amount"`
	InFactor          float64 `json:"in_factor" gorm:"column:in_factor"`
	SumIf             float64 `json:"sum_if" gorm:"column:sum_if"`
	OutStandingAmount float64 `json:"outstanding_amount" gorm:"column:outstanding_amount"`
	SaleCode          string  `json:"sale_code" gorm:"column:sale_code"`
	SaleName          string  `json:"sale_name" gorm:"column:sale_name"`
	ExFactor          float64 `json:"ex_factor" gorm:"column:ex_factor"`
	SumEf             float64 `json:"sum_ef" gorm:"column:sum_ef"`
	Department        string  `json:"department" gorm:"column:department"`
	InvAmountCal      float64 `json:"inv_amount_cal" gorm:"column:inv_amount_cal"`
	Nname             string  `json:"nname" gorm:"column:nname"`
	Status            string  `json:"status" gorm:"column:status"`
	SaleFactor        float64 `json:"sale_factor" gorm:"column:sale_factor"`
	SoNumberAll       int     `json:"sonumber_all" gorm:"column:sonumber_all"`
}
