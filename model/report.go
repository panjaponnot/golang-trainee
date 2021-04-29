package model

import "time"

type OrgChart struct {
	Order        int     `json:"order"`
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
	Commission   float64 `json:"commission"`
}

type InvBefore struct {
	StaffID   string  `json:"staff_id"`
	InvAmount float64 `json:"inv_amount"`
	CheckData int     `json:"check_data" gorm:"column:checkdata"`
}

type SoExport struct {
	CostsheetNumber     string  `json:"costsheet_number" gorm:"column:costsheetnumber"`
	SoNumber            string  `json:"so_number" gorm:"column:sonumber"`
	ContractStartDate   string  `json:"ContractStartDate" gorm:"column:ContractStartDate"`
	ContractEndDate     string  `json:"ContractEndDate" gorm:"column:ContractEndDate"`
	PriceSale           float64 `json:"pricesale" gorm:"column:pricesale"`
	TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
	SOWebStatus         string  `json:"SOWebStatus" gorm:"column:SOWebStatus"`
	CustomerId          string  `json:"Customer_ID" gorm:"column:Customer_ID"`
	CustomerName        string  `json:"Customer_Name" gorm:"column:Customer_Name"`
	SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
	SaleName            string  `json:"sale_name" gorm:"column:sale_name"`
	SaleTeam            string  `json:"sale_team" gorm:"column:sale_team"`
	PeriodAmount        float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
	SaleFactor          float64 `json:"sale_factor" gorm:"column:sale_factor"`
	Infactor            float64 `json:"in_factor" gorm:"column:in_factor"`
	ExFactor            float64 `json:"ex_factor" gorm:"column:ex_factor"`
	SoRefer             string  `json:"so_refer" gorm:"column:so_refer"`
	SoType              string  `json:"SOType" gorm:"column:SOType"`
	Detail              string  `json:"detail" gorm:"column:detail"`
}

type SummaryCustomer struct {
	BLSCDocNo         string  `json:"blsc_doc_no" gorm:"column:BLSCDocNo"`
	CustomerID        string  `json:"customer_id" gorm:"column:Customer_ID"`
	CustomerName      string  `json:"customer_name" gorm:"column:Customer_Name"`
	TotalSo           float64 `json:"total_so" gorm:"column:total_so"`
	TotalCs           float64 `json:"total_cs" gorm:"column:total_cs"`
	SoAmountTotalInv  float64 `json:"total_inv" gorm:"column:total_inv"`
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

type TrackInvoice struct {
	BLSCDocNo         string  `json:"blsc_doc_no" gorm:"column:BLSCDocNo"`
	TotalSo           float64 `json:"total_so" gorm:"column:total_so"`
	TotalCs           float64 `json:"total_cs" gorm:"column:total_cs"`
	SoAmountTotalInv  float64 `json:"total_inv" gorm:"column:total_inv"`
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
	ExFactor          float64 `json:"ex_factor" gorm:"column:ex_factor"`
	SumEf             float64 `json:"sum_ef" gorm:"column:sum_ef"`
	InvAmountCal      float64 `json:"inv_amount_cal" gorm:"column:inv_amount_cal"`
	SaleFactor        float64 `json:"sale_factor" gorm:"column:sale_factor"`
	SoNumberAll       int     `json:"sonumber_all" gorm:"column:sonumber_all"`
}

type SOMssql struct {
	BlscDocNo         string   `json:"blsc_doc_no" gorm:"column:BLSCDocNo"`
	SOnumber          string   `json:"so_number" gorm:"column:sonumber"`
	CustomerId        string   `json:"customer_id" gorm:"column:Customer_ID"`
	CustomerName      string   `json:"customer_name" gorm:"column:Customer_Name"`
	ContractStartDate string   `json:"contract_start_date" gorm:"column:ContractStartDate"`
	ContractEndDate   string   `json:"contract_end_date" gorm:"column:ContractEndDate"`
	SORefer           string   `json:"so_refer" gorm:"column:so_refer"`
	SaleCode          string   `json:"sale_code" gorm:"column:sale_code"`
	SaleLead          string   `json:"sale_lead" gorm:"column:sale_lead"`
	Day               string   `json:"day" gorm:"column:days"`
	SoMonth           string   `json:"so_month" gorm:"column:so_month"`
	SOWebStatus       string   `json:"so_web_status" gorm:"column:SOWebStatus"`
	PriceSale         float64  `json:"price_sale" gorm:"column:pricesale"`
	PeriodAmount      float64  `json:"period_amount" gorm:"column:PeriodAmount"`
	TotalAmount       float64  `json:"total_amount" gorm:"column:TotalAmount"`
	StaffId           string   `json:"staff_id" gorm:"column:staff_id"`
	PayType           string   `json:"pay_type" gorm:"column:pay_type"`
	SoType            string   `json:"so_type" gorm:"column:so_type"`
	INCSCDocNo        string   `json:"INCSCDocNo" gorm:"column:INCSCDocNo"`
	Prefix            string   `json:"prefix"`
	Fname             string   `json:"fname"`
	Lname             string   `json:"lname"`
	Nname             string   `json:"nname"`
	Position          string   `json:"position"`
	Department        string   `json:"department"`
	Status            string   `json:"status"`
	Remark            string   `json:"remark"`
	GetCN             string   `json:"get_cn" gorm:"column:getCN"`
	Inv               *Invoice `json:"invoice" gorm:"foreignkey:BLSCDocNo;association_foreignkey:invoice_no"`
}

type SOMssqlInfo struct {
	// BlscDocNo         string   `json:"blsc_doc_no" gorm:"column:BLSCDocNo"`
	SOnumber   string `json:"so_number" gorm:"column:so_number"`
	CSnumber   string `json:"cs_number" gorm:"column:cs_number"`
	INCSCDocNo string `json:"rc_number" gorm:"column:rc_number"`
	CustomerId string `json:"customer_id" gorm:"column:customer_id"`
	// CustomerName      string   `json:"customer_name" gorm:"column:Customer_Name"`
	ContractStartDate string  `json:"contract_start_date" gorm:"column:contract_start_date"`
	ContractEndDate   string  `json:"contract_end_date" gorm:"column:contract_end_date"`
	SORefer           string  `json:"so_refer" gorm:"column:so_refer"`
	SaleCode          string  `json:"sale_id" gorm:"column:sale_id"`
	SaleLead          string  `json:"sale_lead" gorm:"column:sale_lead"`
	Day               string  `json:"day" gorm:"column:days"`
	SoMonth           string  `json:"so_month" gorm:"column:so_month"`
	SOWebStatus       string  `json:"so_web_status" gorm:"column:so_web_status"`
	PriceSale         float64 `json:"total_contract" gorm:"column:total_contract"`
	PeriodAmount      float64 `json:"total_contract_per_month" gorm:"column:total_contract_per_month"`
	TotalAmount       float64 `json:"total_contract" gorm:"column:total_contract"`
	StaffId           string  `json:"staff_id" gorm:"column:staff_id"`
	PayType           string  `json:"pay_type" gorm:"column:pay_type"`
	SoType            string  `json:"so_type" gorm:"column:so_type"`
	// INCSCDocNo        string   `json:"INCSCDocNo" gorm:"column:INCSCDocNo"`
	Prefix     string `json:"prefix"`
	Fname      string `json:"fname"`
	Lname      string `json:"lname"`
	Nname      string `json:"nname"`
	Position   string `json:"position"`
	Department string `json:"department"`
	Status     string `json:"status"`
	Remark     string `json:"remark"`
	// GetCN             string   `json:"get_cn" gorm:"column:getCN"`
	// Inv               *Invoice `json:"invoice" gorm:"foreignkey:BLSCDocNo;association_foreignkey:invoice_no"`
}

func (SOMssql) TableName() string {
	return "so_mssql"
}

type CheckExpire struct {
	Model
	SOnumber   string    `json:"so_number" gorm:"column:sonumber"`
	Status     string    `json:"status" gorm:"column:status"`
	Remark     string    `json:"remark" gorm:"column:remark"`
	CreateDate time.Time `json:"create_date" gorm:"column:create_date"`
	CreateBy   string    `json:"create_by" gorm:"column:create_by"`
}

func (CheckExpire) TableName() string {
	return "check_expire_log"
}
