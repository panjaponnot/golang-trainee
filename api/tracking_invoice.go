package api

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"time"
	"sale_ranking/pkg/util"

	"github.com/labstack/echo/v4"
)

func SO_Detail(c echo.Context) error{
	type SO_Data struct {
		Sonumber			string	`json:"sonumber" gorm:"column:sonumber"`
		BLSCDocNo			string	`json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PeriodStartDate 	string	`json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		PeriodEndDate		string	`json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Customer_ID			string	`json:"Customer_ID" gorm:"column:Customer_ID"`
		Customer_Name		string	`json:"Customer_Name" gorm:"column:Customer_Name"`
		Sale_code			string	`json:"sale_code" gorm:"column:sale_code"`
		Sale_team			string	`json:"sale_team" gorm:"column:sale_team"`
		Sale_name			string	`json:"sale_name" gorm:"column:sale_name"`
		In_factor			string	`json:"in_factor" gorm:"column:in_factor"`
		Ex_factor			string	`json:"ex_factor" gorm:"column:ex_factor"`
		Active_so_status	string	`json:"active_so_status" gorm:"column:active_so_status"`
		So_amount			string	`json:"so_amount" gorm:"column:so_amount"`
	}
	
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	staffid := strings.TrimSpace(c.QueryParam("staff_id"))
	search := strings.TrimSpace(c.QueryParam("search"))

	var dataRaw []SO_Data
	var errr int = 0
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

	sql := `select SOO.sonumber,SOO.BLSCDocNo,SOO.PeriodStartDate,SOO.PeriodEndDate,SOO.Customer_ID,
	SOO.Customer_Name,SOO.sale_code,SOO.sale_team,SOO.sale_name,SOO.in_factor,SOO.ex_factor,
	(CASE
		WHEN BLSCDocNo is not null or BLSCDocNo not like ''
		THEN 'ออก invoice เสร็จสิ้น'
		ELSE 'ยังไม่ออก invoice'
	END) active_so_status,
	(CASE
		WHEN DATEDIFF(?, ?) = 0
		THEN 0
		WHEN DATEDIFF(PeriodEndDate, PeriodStartDate)+1 = 0
		THEN 0
		WHEN PeriodStartDate >= ? AND PeriodStartDate <= ? AND PeriodEndDate <= ?
		THEN PeriodAmount
		WHEN PeriodStartDate >= ? AND PeriodStartDate <= ? AND PeriodEndDate > ?
		THEN (DATEDIFF(?, PeriodStartDate)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
		WHEN PeriodStartDate < ? AND PeriodEndDate <= ? AND PeriodEndDate > ?
		THEN (DATEDIFF(PeriodEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
		WHEN PeriodStartDate < ? AND PeriodEndDate = ?
		THEN 1*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
		WHEN PeriodStartDate < ? AND PeriodEndDate > ?
		THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate,PeriodStartDate)+1))
		ELSE 0 END
	) so_amount
	FROM (
		SELECT sonumber,BLSCDocNo,PeriodStartDate,PeriodEndDate,PeriodAmount,Customer_ID,
		Customer_Name,sale_code,sale_team,sale_name,in_factor,ex_factor 
		FROM so_mssql
		WHERE Active_Inactive = 'Active' `
		if St_date != "" || En_date != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` PeriodStartDate >= '`+St_date+`' AND PeriodStartDate <= '`+En_date+`' `
				if En_date != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` PeriodEndDate <= '`+En_date+`' AND PeriodEndDate >= '`+St_date+`' `
			}
		}else{
			St_date = dateFrom.String()
			En_date = dateTo.String()
		}
		sql = sql +` group by sonumber
	) SOO
	LEFT JOIN (select staff_id from staff_info) si on SOO.sale_code = si.staff_id
	WHERE INSTR(CONCAT_WS('|', si.staff_id), ?) AND
	INSTR(CONCAT_WS('|', SOO.sonumber,SOO.BLSCDocNo,SOO.Customer_ID,SOO.Customer_Name,SOO.sale_code,
	SOO.sale_team,SOO.sale_name), ?)`

	if err := dbSale.Ctx().Raw(sql,St_date,En_date,St_date,En_date,En_date,St_date,En_date,En_date, 
		En_date, St_date,En_date,St_date,St_date,St_date,St_date,St_date,En_date, 
		En_date, St_date,staffid,search).Scan(&dataRaw).Error; err != nil {
		errr += 1
	}

	fmt.Println(`------------------------`)

	return c.JSON(http.StatusOK,dataRaw)
}