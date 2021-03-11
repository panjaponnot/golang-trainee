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
		So_amount			string	`json:"so_amount" gorm:"column:so_amount"`
		Inv_status			string	`json:"inv_status" gorm:"column:inv_status"`
	}
	
	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
	}

	type Staffs_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
	}

	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	staffid := strings.TrimSpace(c.QueryParam("staffid"))
	SaleID := strings.TrimSpace(c.QueryParam("saleid"))
	search := strings.TrimSpace(c.QueryParam("search"))
	Status := strings.TrimSpace(c.QueryParam("status"))

	if strings.TrimSpace(c.QueryParam("saleid")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

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

	sql := `select SOO.sonumber,SOO.BLSCDocNo,SOO.PeriodStartDate,SOO.PeriodEndDate,SOO.Customer_ID,
	SOO.Customer_Name,SOO.sale_code,SOO.sale_team,SOO.sale_name,SOO.in_factor,SOO.ex_factor,inv_status,
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
		Customer_Name,sale_code,sale_team,sale_name,in_factor,ex_factor,
		(CASE
			WHEN GetCN is not null AND GetCN not like '' 
			THEN 'ลดหนี้' 
			WHEN GetCN is null AND BLSCDocNo is not null AND BLSCDocNo not like '' 
			OR GetCN like '' AND BLSCDocNo is not null AND BLSCDocNo not like '' 
			THEN 'ออก invoice เสร็จสิ้น' 
			ELSE 'ยังไม่ออก invoice' 
			END 
		) inv_status 
		FROM so_mssql
		WHERE Active_Inactive = 'Active' and PeriodStartDate <= ? and PeriodEndDate >= ? 
		and PeriodStartDate <= PeriodEndDate`
		sql = sql +` group by sonumber
	) SOO
	LEFT JOIN (select staff_id from staff_info) si on SOO.sale_code = si.staff_id
	WHERE INSTR(CONCAT_WS('|', si.staff_id), ?) AND
	INSTR(CONCAT_WS('|', SOO.sonumber,SOO.BLSCDocNo,SOO.Customer_ID,SOO.Customer_Name,SOO.sale_code,
	SOO.sale_team,SOO.sale_name), ?) AND SOO.sale_code in (?) AND INSTR(CONCAT_WS('|', inv_status), ?)`

	if err := dbSale.Ctx().Raw(sql,dateTo,dateFrom,dateFrom,dateTo,dateTo,dateFrom,dateTo,dateTo, 
		dateTo,dateFrom,dateTo,dateFrom,dateFrom,dateFrom,dateFrom,dateFrom,dateTo, 
		dateTo,dateFrom,dateTo,dateFrom,staffid,search,listId,Status).Scan(&dataRaw).Error; err != nil {
		errr += 1
	}

	fmt.Println(`------------------------`)

	return c.JSON(http.StatusOK,dataRaw)
}