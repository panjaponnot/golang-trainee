package api

import (
	"fmt"
	"net/http"
	"sale_ranking/pkg/log"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func SOReceiveEndPoint(c echo.Context) error {
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	OneID := strings.TrimSpace(c.QueryParam("oneid"))

	if len(St_date) > 0 && len(En_date) == 0 || len(St_date) == 0 && len(En_date) > 0{
		return echo.ErrBadRequest
	}

	rawData := []struct {
		SDPropertyCS28 				string	`json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
		Count_SDPropertyCS28 		string	`json:"count_SDPropertyCS28" gorm:"column:count_SDPropertyCS28"`
		Sonumber 					string	`json:"sonumber" gorm:"column:sonumber"`
		ContractStartDate			string	`json:"ContractStartDate" gorm:"column:ContractStartDate"`
		ContractEndDate				string	`json:"ContractEndDate" gorm:"column:ContractEndDate"`
		Pricesale					string	`json:"pricesale" gorm:"column:pricesale"`
		TotalContractAmount 		string	`json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		Sum_TotalContractAmount 	string	`json:"sum_TotalContractAmount" gorm:"column:sum_TotalContractAmount"`
		SOWebStatus					string	`json:"SOWebStatus" gorm:"column:SOWebStatus"`
		BLSCDocNo					string	`json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PeriodStartDate				string	`json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		PeriodEndDate				string	`json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		GetCN						string	`json:"GetCN" gorm:"column:GetCN"`
		INCSCDocNo					string	`json:"INCSCDocNo" gorm:"column:INCSCDocNo"`
		Soms_Customer_ID			string	`json:"soms_Customer_ID" gorm:"column:soms_Customer_ID"`
		Soms_Customer_Name			string	`json:"soms_Customer_Name" gorm:"column:soms_Customer_Name"`
		Sale_code					string	`json:"sale_code" gorm:"column:sale_code"`
		Sale_name					string	`json:"sale_name" gorm:"column:sale_name"`
		Sale_team					string	`json:"sale_team" gorm:"column:sale_team"`
		Sale_lead					string	`json:"sale_lead" gorm:"column:sale_lead"`
		PeriodAmount				string	`json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Soms_sale_factor			string	`json:"soms_sale_factor" gorm:"column:soms_sale_factor"`
		Soms_in_factor				string	`json:"soms_in_factor" gorm:"column:soms_in_factor"`
		Soms_ex_factor				string	`json:"soms_ex_factor" gorm:"column:soms_ex_factor"`
		Active_Inactive				string	`json:"Active_Inactive" gorm:"column:Active_Inactive"`
		So_refer					string	`json:"so_refer" gorm:"column:so_refer"`
		Create_date					string	`json:"create_date" gorm:"column:create_date"`
		Status_so					string	`json:"status_so" gorm:"column:status_so"`
		Status_sale					string	`json:"status_sale" gorm:"column:status_sale"`
		Has_refer					string	`json:"has_refer" gorm:"column:has_refer"`
		Doc_number_eform			string 	`json:"doc_number_eform" gorm:"column:doc_number_eform"`
		Ci_Customer_ID				string 	`json:"ci_Customer_ID" gorm:"column:ci_Customer_ID"`
		Ci_Cusname_thai				string 	`json:"ci_Cusname_thai" gorm:"column:ci_Cusname_thai"`
		SO_Type						string 	`json:"SO_Type" gorm:"column:SO_Type"`
		Job_Status					string 	`json:"Job_Status" gorm:"column:Job_Status"`
		Ci_Total_Revenue_Month		string 	`json:"ci_Total_Revenue_Month" gorm:"column:ci_Total_Revenue_Month"`
		Ci_SaleFactors				string 	`json:"ci_SaleFactors" gorm:"column:ci_SaleFactors"`
		Ci_Int_INET					string 	`json:"ci_Int_INET" gorm:"column:ci_Int_INET"`
		Ci_Ext_JV					string 	`json:"ci_Ext_JV" gorm:"column:ci_Ext_JV"`
		Ci_Ext						string 	`json:"ci_Ext" gorm:"column:ci_Ext"`
		Status_eform				string	`json:"status_eform" gorm:"column:status_eform"`
		Eng_cost					string	`json:"eng_cost" gorm:"column:eng_cost"`
		Total_sale_factor			string
		Ci_external_factor			string	`json:"ci_external_factor" gorm:"column:ci_external_factor"`
		ContractAndRevenue_status	string
	}{}

	sql := `select smt.SDPropertyCS28,count(smt.SDPropertyCS28) as count_SDPropertyCS28,
	smt.sonumber,smt.ContractStartDate,smt.ContractEndDate,smt.pricesale,
	smt.TotalContractAmount,sum(TotalContractAmount) as sum_TotalContractAmount,smt.SOWebStatus,
	smt.BLSCDocNo,smt.PeriodStartDate,smt.PeriodEndDate,smt.GetCN,smt.INCSCDocNo,
	smt.Customer_ID as soms_Customer_ID,smt.Customer_Name,
	smt.sale_code,smt.sale_name,smt.sale_team,smt.sale_lead,smt.PeriodAmount,
	smt.sale_factor as soms_sale_factor,smt.in_factor as soms_in_factor,
	smt.ex_factor as soms_ex_factor,smt.Active_Inactive,
	smt.so_refer,smt.create_date,smt.status_so,smt.status_sale,smt.has_refer,
	ci.doc_number_eform,ci.Customer_ID as ci_Customer_ID,ci.Cusname_thai as ci_Cusname_thai,
	ci.SO_Type,ci.Job_Status,ci.Total_Revenue_Month as ci_Total_Revenue_Month,
	ci.SaleFactors as ci_SaleFactors,ci.Int_INET as ci_Int_INET,
	ci.Ext_JV as ci_Ext_JV,ci.Ext as ci_Ext,ci.status_eform,(smt.TotalContractAmount/smt.sale_factor) as eng_cost,
	(ci.ext+ci.ext_jv) as ci_external_factor
	from costsheet_info ci 
	left join staff_info si on ci.ID_PreSale = si.staff_id
	LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28`
	
	if St_date != "" || En_date != "" || OneID != ""{
		sql = sql+` where `
		if St_date != ""{
			sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
			if En_date != "" || OneID != ""{
				sql = sql+` AND `
			}
		}
		if En_date != ""{
			sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
			if OneID != ""{
				sql = sql+` AND `
			}
		}
		if OneID != ""{
			sql = sql+` si.one_id like `+OneID+` `
		}

	}

	sql = sql+` GROUP BY smt.SDPropertyCS28 `
	
	if err := dbSale.Ctx().Raw(sql).Scan(&rawData).Error; err != nil {
		log.Errorln("GetReceive error :-", err)
	}

	for i , j := range rawData {
		tfc := "0"
		if j.Sum_TotalContractAmount != "" && j.Sum_TotalContractAmount != "0" && j.Eng_cost != "" && j.Eng_cost != "0"{
			tfa,_ := strconv.ParseFloat(j.Sum_TotalContractAmount,64)
			tfb,_ := strconv.ParseFloat(j.Eng_cost,64)
			tfc = fmt.Sprintf("%f",tfa/tfb)
		}
		rawData[i].Total_sale_factor = tfc

		csa,_ := strconv.Atoi(j.Count_SDPropertyCS28)
		if csa > 0{
			rawData[i].ContractAndRevenue_status = "True"
		}else{
			rawData[i].ContractAndRevenue_status = "False"
		}
	}
	
	fmt.Println("Success")

	return c.JSON(http.StatusOK, rawData)
}