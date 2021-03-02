package api

import (
	"fmt"
	"net/http"
	"time"
	"strconv"
	"strings"
	"sale_ranking/pkg/util"

	"github.com/labstack/echo/v4"
)

func Costsheet_Detail(c echo.Context) error{
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	staffid := strings.TrimSpace(c.QueryParam("staff_id"))
	search := strings.TrimSpace(c.QueryParam("search"))

	type Costsheet_Val struct {
		Doc_number_eform			string `json:"doc_number_eform" gorm:"column:doc_number_eform"`
		Sonumber					string	`json:"sonumber" gorm:"column:sonumber"`
		StartDate_P1				string	`json:"StartDate_P1" gorm:"column:StartDate_P1"`
		EndDate_P1					string	`json:"EndDate_P1" gorm:"column:EndDate_P1"`
		Status_eform				string `json:"status_eform" gorm:"column:status_eform"`
		Customer_ID					string `json:"Customer_ID" gorm:"column:Customer_ID"`
		Cusname_thai				string `json:"Cusname_thai" gorm:"column:Cusname_thai"`
		Cusname_Eng					string `json:"Cusname_Eng" gorm:"column:Cusname_Eng"`
		ID_PreSale					string `json:"ID_PreSale" gorm:"column:ID_PreSale"`
		Sale_Team					string `json:"Sale_Team" gorm:"column:Sale_Team"`
		Sales_Name					string `json:"Sales_Name" gorm:"column:Sales_Name"`
		Sales_Surname				string `json:"Sales_Surname" gorm:"column:Sales_Surname"`
		Int_INET					string `json:"Int_INET" gorm:"column:Int_INET"`
		Ext							string `json:"External" gorm:"column:External"`
		So_amount					string `json:"so_amount" gorm:"column:so_amount"`
	}

	var errr int = 0
	var dataRaw []Costsheet_Val
	ds := time.Now()
	de := time.Now()
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("startdate")), 10); err == nil {
		ds = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("enddate")), 10); err == nil {
		de = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	yearStart, monthStart, dayStart := ds.Date()
	yearEnd, monthEnd, dayEnd := de.Date()
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

	sql := `select QW.*
	from 
	(
		select ci.doc_number_eform,smt.sonumber,ci.StartDate_P1,
		 ci.EndDate_P1,ci.status_eform,ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.Sale_Team,
		 ci.Sales_Name,ci.Sales_Surname,ci.Int_INET,(ci.Ext_JV+ci.Ext) as External,
		  (CASE
			 	WHEN DATEDIFF(ci.EndDate_P1,ci.StartDate_P1)+1 = 0
			 	THEN 0
			 	WHEN ci.StartDate_P1 >= ? AND ci.StartDate_P1 <= ? AND ci.EndDate_P1 <= ?
			 	THEN ci.Total_Revenue_Month
			 	WHEN ci.StartDate_P1 >= ? AND ci.StartDate_P1 <= ? AND ci.EndDate_P1 > ?
			 	THEN (DATEDIFF(?, ci.StartDate_P1)+1)*(ci.Total_Revenue_Month/(DATEDIFF(ci.EndDate_P1, ci.StartDate_P1)+1))
			 	WHEN ci.StartDate_P1 < ? AND ci.EndDate_P1 <= ? AND ci.EndDate_P1 > ?
			 	THEN (DATEDIFF(ci.EndDate_P1, ?)+1)*(ci.Total_Revenue_Month/(DATEDIFF(ci.EndDate_P1, ci.StartDate_P1)+1))
			 	WHEN ci.StartDate_P1 < ? AND ci.EndDate_P1 = ?
			 	THEN 1*(ci.Total_Revenue_Month/(DATEDIFF(ci.EndDate_P1, ci.StartDate_P1)+1))
			 	WHEN ci.StartDate_P1 < ? AND ci.EndDate_P1 > ?
			 	THEN (DATEDIFF(?,?)+1)*(ci.Total_Revenue_Month/(DATEDIFF(ci.EndDate_P1,ci.StartDate_P1)+1))
			 	ELSE 0 END
			) as so_amount
		from costsheet_info ci
		LEFT JOIN (
			select * 
			from so_mssql_test
			group by sonumber
			)smt on ci.doc_number_eform = smt.SDPropertyCS28
		LEFT JOIN staff_info si on ci.ID_Presale = si.staff_id 
		where INSTR(CONCAT_WS('|', ci.tracking_id,ci.doc_id,ci.doc_number_eform,ci.Customer_ID,
		ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,ci.Sale_Team,
		ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?)
		and INSTR(CONCAT_WS('|', si.staff_id), ?) `
	if St_date != "" || En_date != ""{
		sql = sql+` AND `
		if St_date != ""{
			sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
			if En_date != "" {
				sql = sql+` AND `
			}
		}
		if En_date != ""{
			sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
		}
		sql = sql+`) QW`

		if err := dbSale.Ctx().Raw(sql, St_date,En_date,En_date, St_date,En_date, 
			En_date,En_date, St_date, En_date,St_date,St_date,St_date,St_date, 
			St_date, En_date,En_date,St_date,search,staffid).Scan(&dataRaw).Error; err != nil {
			errr += 1
			return echo.ErrInternalServerError
		}
	}else{
		sql = sql+`) QW`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, 
			dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, 
			dateFrom,search,staffid).Scan(&dataRaw).Error; err != nil {
			errr += 1
			return echo.ErrInternalServerError
		}
	}

	fmt.Println(dateFrom)
	fmt.Println(dateTo)

	return c.JSON(http.StatusOK,dataRaw)
}