package api

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"time"
	"sale_ranking/pkg/util"
	"sale_ranking/pkg/server"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func Invoice_Detail(c echo.Context) error{
	type Invoice_Data struct {
		Invoice_no			string	`json:"invoice_no" gorm:"column:invoice_no"`
		So_number			string	`json:"sonumber" gorm:"column:sonumber"`
		Status				string	`json:"status" gorm:"column:status"`
		Reason				string	`json:"reason" gorm:"column:reason"`
		Customer_ID			string	`json:"Customer_ID" gorm:"column:Customer_ID"`
		Customer_Name		string	`json:"Customer_Name" gorm:"column:Customer_Name"`
		Sale_team			string	`json:"sale_team" gorm:"column:sale_team"`
		Sale_name			string	`json:"sale_name" gorm:"column:sale_name"`
		In_factor			string	`json:"in_factor" gorm:"column:in_factor"`
		Ex_factor			string	`json:"ex_factor" gorm:"column:ex_factor"`
		So_amount			string	`json:"so_amount" gorm:"column:so_amount"`
	}

	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
	}

	type Staffs_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
	}

	var dataRaw []Invoice_Data

	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	staffid := strings.TrimSpace(c.QueryParam("staffid"))
	SaleID := strings.TrimSpace(c.QueryParam("saleid"))
	search := strings.TrimSpace(c.QueryParam("search"))
	Status := strings.TrimSpace(c.QueryParam("status"))

	if strings.TrimSpace(c.QueryParam("saleid")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

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

	if St_date == "" || En_date == ""{
		dayStart = 1
	}

	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

	var user_data_raw []Users_Data

	if err := dbSale.Ctx().Raw(`SELECT * FROM user_info WHERE staff_id = ? and role = 'admin';`, SaleID).Scan(&user_data_raw).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
	}
	var listId []string
	if len(user_data_raw) != 0 {
		var staffAll []Staffs_Data
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info ;`).Scan(&staffAll).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
			}
		}
		for _, i := range staffAll {
			listId = append(listId, i.Staff_id)
		}
	} else {
		var staffAll Staffs_Data
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info WHERE staff_id = ?;`, SaleID).Scan(&staffAll).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
			}
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
		if staffAll.Staff_child != "" {
			data := strings.Split(staffAll.Staff_child, ",")
			listId = data
		}
		listId = append(listId, staffAll.Staff_id)
	}

	sql := `select bi.invoice_no,BL.sonumber,bi.status,bi.reason,BL.Customer_ID,BL.Customer_Name,
	BL.sale_team,BL.sale_name,BL.in_factor,BL.ex_factor,BL.so_amount
	from (select *,
		(CASE
			WHEN DATEDIFF(?, ?) = 0
			THEN 0
			WHEN DATEDIFF(smt.PeriodEndDate,smt.PeriodStartDate)+1 = 0
			THEN 0
			WHEN smt.PeriodStartDate >= ? AND smt.PeriodStartDate <= ? AND smt.PeriodEndDate <= ?
			THEN PeriodAmount
			WHEN smt.PeriodStartDate >= ? AND smt.PeriodStartDate <= ? AND smt.PeriodEndDate > ?
			THEN (DATEDIFF(?, smt.PeriodStartDate)+1)*(smt.PeriodAmount/(DATEDIFF(smt.PeriodEndDate, smt.PeriodStartDate)+1))
			WHEN smt.PeriodStartDate < ? AND smt.PeriodEndDate <= ? AND smt.PeriodEndDate > ?
			THEN (DATEDIFF(smt.PeriodEndDate, ?)+1)*(smt.PeriodAmount/(DATEDIFF(smt.PeriodEndDate, smt.PeriodStartDate)+1))
			WHEN smt.PeriodStartDate < ? AND smt.PeriodEndDate = ?
			THEN 1*(smt.PeriodAmount/(DATEDIFF(smt.PeriodEndDate, smt.PeriodStartDate)+1))
			WHEN smt.PeriodStartDate < ? AND smt.PeriodEndDate > ?
			THEN (DATEDIFF(?,?)+1)*(smt.PeriodAmount/(DATEDIFF(smt.PeriodEndDate,smt.PeriodStartDate)+1))
			ELSE 0 END
		) so_amount
		from so_mssql smt
		WHERE smt.Active_Inactive = 'Active' 
		AND smt.sonumber is not null 
		AND smt.sonumber not like '' 
		and PeriodStartDate <= ? 
		and PeriodEndDate >= ? 
		and PeriodStartDate <= PeriodEndDate`
		sql = sql+` group by smt.sonumber
	) BL
	LEFT JOIN (select staff_id from staff_info) si on BL.sale_code = si.staff_id
	LEFT JOIN billing_info bi on BL.BLSCDocNo = bi.invoice_no
	WHERE  INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
	INSTR(CONCAT_WS('|',bi.invoice_no,BL.sonumber,bi.reason,BL.Customer_ID,BL.Customer_Name,
	BL.sale_team,BL.sale_name), ?) AND BL.sale_code in (?) AND INSTR(CONCAT_WS('|', bi.status), ?) `

	if err := dbSale.Ctx().Raw(sql,dateTo,dateFrom,dateFrom,dateTo,dateTo,dateFrom,dateTo,dateTo, 
		dateTo,dateFrom,dateTo,dateFrom,dateFrom,dateFrom,dateFrom,dateFrom,dateTo, 
		dateTo,dateFrom,dateTo,dateFrom,staffid,search,listId,Status).Scan(&dataRaw).Error; err != nil {
		errr += 1
	}

	fmt.Println("--------------------------------")
	
	return c.JSON(http.StatusOK,dataRaw)
}