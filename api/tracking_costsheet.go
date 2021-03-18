package api

import (
	"fmt"
	"net/http"
	"time"
	"strconv"
	"strings"
	"sale_ranking/pkg/util"
	"sale_ranking/pkg/server"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func Costsheet_Detail(c echo.Context) error{
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	staffid := strings.TrimSpace(c.QueryParam("staffid"))
	SaleID := strings.TrimSpace(c.QueryParam("saleid"))
	search := strings.TrimSpace(c.QueryParam("search"))
	Status := strings.TrimSpace(c.QueryParam("status"))

	if strings.TrimSpace(c.QueryParam("saleid")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

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
		So_status					string `json:"so_status" gorm:"column:so_status"`
	}

	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
	}

	type Staffs_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
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
	
	dateFrom :=  ""
	dateTo := ""

	if St_date == "" || En_date == ""{
		dayStart = 1
		dateFromA := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
		dateToA := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
		dateFrom =  dateFromA.String()
		dateTo = dateToA.String()
	}else{
		dateFrom = St_date
		dateTo = En_date
	}

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

	sql := `select QW.*
	from 
	(
		select ci.doc_number_eform,smt.sonumber,ci.StartDate_P1,
		 ci.EndDate_P1,ci.status_eform,ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.EmployeeID,ci.Sale_Team,
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
			) as so_amount,
			(CASE
				WHEN smt.SDPropertyCS28 is not null AND smt.SDPropertyCS28 not like '' 
				AND ci.status_eform like '%Complete from paperless%'
				THEN 'ออก so เสร็จสิ้น'
				ELSE 'ยังไม่ออก so'
			END) so_status
		from costsheet_info ci
		LEFT JOIN (
			select * 
			from so_mssql
			where SDPropertyCS28 <> ''
			group by sonumber
			)smt on ci.doc_number_eform = smt.SDPropertyCS28
		LEFT JOIN staff_info si on ci.EmployeeID = si.staff_id 
		where INSTR(CONCAT_WS('|', ci.tracking_id,ci.doc_id,ci.doc_number_eform,ci.Customer_ID,
		ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,ci.Sale_Team,
		ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID), ?)
		and INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1
		and ci.EmployeeID in (?) `
	if Status == "SO Compelte" || Status == "ออก so เสร็จสิ้น"{
		sql = sql+` smt.SDPropertyCS28 is not null AND smt.SDPropertyCS28 not like '' 
		AND ci.status_eform like '%Complete from paperless%'`
	}else{
		sql = sql+` AND INSTR(CONCAT_WS('|', ci.status_eform), ?) `
	}
	sql = sql+`) QW`
	
	if Status == "SO Compelte" || Status == "ออก so เสร็จสิ้น"{
		if err := dbSale.Ctx().Raw(sql,dateFrom,dateTo,dateTo,dateFrom,dateTo,dateTo, 
			dateTo,dateFrom,dateTo,dateFrom,dateFrom,dateFrom,dateFrom,dateFrom,dateTo, 
			dateTo,dateFrom,search,staffid,dateFrom,dateTo,dateFrom,
			dateTo,listId).Scan(&dataRaw).Error; err != nil {
			errr += 1
		}
	}else{
		if err := dbSale.Ctx().Raw(sql,dateFrom,dateTo,dateTo,dateFrom,dateTo,dateTo, 
			dateTo,dateFrom,dateTo,dateFrom,dateFrom,dateFrom,dateFrom,dateFrom,dateTo, 
			dateTo,dateFrom,search,staffid,dateFrom,dateTo,dateFrom,
			dateTo,listId,Status).Scan(&dataRaw).Error; err != nil {
			errr += 1
		}
	}
	
	fmt.Println(dateFrom)
	fmt.Println(dateTo)

	return c.JSON(http.StatusOK,dataRaw)
}