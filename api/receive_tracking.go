package api

import (
	"net/http"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/util"
	"sale_ranking/pkg/server"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func CostSheet_Status(c echo.Context) error {
	Startdate := strings.TrimSpace(c.QueryParam("startdate"))
	Enddate := strings.TrimSpace(c.QueryParam("enddate"))
	StaffID := strings.TrimSpace(c.QueryParam("staffid"))
	SaleID := strings.TrimSpace(c.QueryParam("saleid"))
	Search := strings.TrimSpace(c.QueryParam("search"))

	if strings.TrimSpace(c.QueryParam("saleid")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

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

	if Startdate == "" || Enddate == ""{
		dayStart = 1
		dateFromA := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
		dateToA := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
		dateFrom =  dateFromA.String()
		dateTo = dateToA.String()
	}else{
		dateFrom = Startdate
		dateTo = Enddate
	}

	type Costsheet_Data struct {
		Total_Revenue_Month string `json:"Total_Revenue_Month" gorm:"column:Total_Revenue_Month"`
	}

	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
	}

	type Staffs_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
	}

	dataResult := struct {
		SOCompelte            			interface{}
		Compeltefrompaperless 			interface{}
		Completefromeform     			interface{}
		Onprocessfrompaperless          interface{}
		Onprocessfromeform              interface{}
		Cancelfrompaperles    			interface{}
		Cancelfromeform                 interface{}
		Rejectfrompaperless     		interface{}
		Rejectfromeform                 interface{}
		Total			                interface{}
	}{}
	CountRejectfromeform := 0
	CountRejectfrompaperless := 0
	CountCancelfromeform := 0
	CountCancelfrompaperless := 0
	CountOnprocessfromeform := 0
	CountOnprocessfrompaperless := 0
	CountCompletefromeform := 0
	CountCompeltefrompaperless := 0
	CountSOCompelte := 0

	totalRejectfromeform := float64(0)
	totalRejectfrompaperless := float64(0)
	totalCancelfromeform := float64(0)
	totalCancelfrompaperless := float64(0)
	totalOnprocessfromeform := float64(0)
	totalOnprocessfrompaperless := float64(0)
	totalCompletefromeform := float64(0)
	totalCompeltefrompaperless := float64(0)
	totalSOCompelte := float64(0)

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(9)
	
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

	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN (
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from paperless%' AND 
		smt.SDPropertyCS28 is not null AND ci.EmployeeID in (?) AND
		smt.SDPropertyCS28 not like '' AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.SOCompelte = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "SO Compelte",
		}
		CountSOCompelte = Count_Costsheet
		totalSOCompelte = TRM_All
		wg.Done()

	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from paperless%' AND smt.SDPropertyCS28 is null AND 
		ci.EmployeeID in (?) AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`
		

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Compeltefrompaperless = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Compelte from paperless",
		}
		CountCompeltefrompaperless = Count_Costsheet
		totalCompeltefrompaperless = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN (
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from eform%' AND ci.EmployeeID in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Completefromeform = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Complete from eform",
		}
		CountCompletefromeform = Count_Costsheet
		totalCompletefromeform = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Onprocess from paperless%' AND ci.EmployeeID in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Onprocessfrompaperless = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Onprocess from paperless",
		}
		CountOnprocessfrompaperless = Count_Costsheet
		totalOnprocessfrompaperless = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Onprocess from eform%' AND ci.EmployeeID in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Onprocessfromeform = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Onprocess from eform",
		}
		CountOnprocessfromeform = Count_Costsheet
		totalOnprocessfromeform = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Cancel from paperless%' AND ci.EmployeeID in (?) 
		AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Cancelfrompaperles = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Cancel from paperless",
		}
		CountCancelfrompaperless = Count_Costsheet
		totalCancelfrompaperless = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Cancel from eform%' AND ci.EmployeeID in (?) 
		AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Cancelfromeform = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Cancel from eform",
		}
		CountCancelfromeform = Count_Costsheet
		totalCancelfromeform = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28
		where ci.status_eform like '%Reject from paperless%' AND ci.EmployeeID in (?) 
		AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Rejectfrompaperless = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Reject from paperless",
		}
		CountRejectfrompaperless = Count_Costsheet
		totalRejectfrompaperless = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28
		where ci.status_eform like '%Reject from eform%' AND ci.EmployeeID in (?) 
		AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', ci.status,ci.tracking_id,ci.doc_id,ci.documentJson,ci.doc_number_eform,
		ci.Customer_ID,ci.Cusname_thai,ci.Cusname_Eng,ci.ID_PreSale,ci.cvm_id,ci.Business_type,
		ci.Sale_Team,ci.Job_Status,ci.SO_Type,ci.Sales_Name,ci.Sales_Surname,ci.EmployeeID,ci.status_eform), ?) 
		AND ci.StartDate_P1 >= ?  AND ci.StartDate_P1 <= ? 
		and ci.EndDate_P1 >= ? AND ci.EndDate_P1 <= ?
		and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateFrom,dateTo,dateFrom,dateTo).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Rejectfromeform = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Reject from eform",
		}
		CountRejectfromeform = Count_Costsheet
		totalRejectfromeform = TRM_All
		wg.Done()
	}()
	wg.Wait()
		dataResult.Total = map[string]interface{}{
		"count": 	CountRejectfromeform + CountRejectfrompaperless + CountCancelfromeform + 
		CountCancelfrompaperless + CountOnprocessfromeform + CountOnprocessfrompaperless + 
		CountCompletefromeform + CountCompeltefrompaperless + CountSOCompelte,
		"status": "Total",
		"total": totalRejectfromeform + totalRejectfrompaperless + totalCancelfromeform + totalCancelfrompaperless + totalOnprocessfromeform +
		totalOnprocessfrompaperless + totalCompletefromeform + totalCompeltefrompaperless + totalSOCompelte,
	}

	// return c.JSON(http.StatusOK, Result)
	return c.JSON(http.StatusOK, dataResult)
}

func Invoice_Status(c echo.Context) error {
	Startdate := strings.TrimSpace(c.QueryParam("startdate"))
	Enddate := strings.TrimSpace(c.QueryParam("enddate"))
	Search := strings.TrimSpace(c.QueryParam("search"))
	Staffid := strings.TrimSpace(c.QueryParam("staffid"))
	SaleID := strings.TrimSpace(c.QueryParam("saleid"))
	Status := strings.TrimSpace(c.QueryParam("status"))

	if strings.TrimSpace(c.QueryParam("saleid")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

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

	if Startdate == "" || Enddate == ""{
		dayStart = 1
		dateFromA := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
		dateToA := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
		dateFrom =  dateFromA.String()
		dateTo = dateToA.String()
	}else{
		dateFrom = Startdate
		dateTo = Enddate
	}

	type Invoice_Data struct {
		PeriodAmount float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Inv_status string `json:"inv_status" gorm:"column:inv_status"`
	}

	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
	}

	type Staffs_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
	}

	dataCount := struct {
		Hasinvoice interface{}
		Reduce     interface{}
		Noinvoice  interface{}
	}{}

	Counthasinvoice := 0
	CountReduce := 0
	Countnoinvoice := 0

	totalhasinvoice := float64(0)
	totalReduce := float64(0)
	totalnoinvoice := float64(0)

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)

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
		var dataRaw []Invoice_Data
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		var Total_PA_B float64 = 0.0
		Count_Invoice_B := 0
		var Total_PA_C float64 = 0.0
		Count_Invoice_C := 0
		sql := `select SOO.PeriodAmount,SOO.inv_status 
		FROM (
			SELECT sonumber,BLSCDocNo,PeriodStartDate,PeriodEndDate,PeriodAmount,Customer_ID,
			Customer_Name,sale_code,sale_team,sale_name,in_factor,ex_factor,Active_Inactive,
			(CASE
				WHEN GetCN is null AND BLSCDocNo is not null AND BLSCDocNo not like '' 
				OR GetCN like '' AND BLSCDocNo is not null AND BLSCDocNo not like '' 
				THEN 'ออก invoice เสร็จสิ้น'
				WHEN GetCN is not null AND GetCN not like '' 
				THEN 'ลดหนี้'  
				ELSE 'ยังไม่ออก invoice' 
				END 
			) inv_status 
			FROM so_mssql ) SOO
		LEFT JOIN (select staff_id from staff_info) si on SOO.sale_code = si.staff_id 
		WHERE SOO.Active_Inactive = 'Active' and 
		SOO.PeriodStartDate >= ? and SOO.PeriodStartDate <= ?
		AND SOO.PeriodEndDate >= ? and SOO.PeriodEndDate <= ?
		and SOO.PeriodStartDate <= SOO.PeriodEndDate 
		AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND
		INSTR(CONCAT_WS('|', SOO.sonumber,SOO.BLSCDocNo,SOO.Customer_ID,SOO.Customer_Name,SOO.sale_code,
		SOO.sale_team,SOO.sale_name), ?) AND SOO.sale_code in (?) AND INSTR(CONCAT_WS('|', inv_status), ?) 
		group by SOO.sonumber`

		if err := dbSale.Ctx().Raw(sql,dateFrom,dateTo,dateFrom,dateTo,
			Staffid,Search,listId,Status).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for i, v := range dataRaw {
			if dataRaw[i].Inv_status == "ออก invoice เสร็จสิ้น"{
				Total_PA += v.PeriodAmount
				Count_Invoice += 1
			}else if dataRaw[i].Inv_status == "ลดหนี้"{
				Total_PA_B += v.PeriodAmount
				Count_Invoice_B += 1
			}else if dataRaw[i].Inv_status == "ยังไม่ออก invoice"{
				Total_PA_C += v.PeriodAmount
				Count_Invoice_C += 1
			}
		}
		dataCount.Hasinvoice = map[string]interface{}{
			"Count":  Count_Invoice,
			"total":  Total_PA,
			"status": "ออก invoice เสร็จสิ้น",
		}
		dataCount.Reduce = map[string]interface{}{
			"Count":  Count_Invoice_B,
			"total":  Total_PA_B,
			"status": "ลดหนี้",
		}
		dataCount.Noinvoice = map[string]interface{}{
			"Count":  Count_Invoice_C,
			"total":  Total_PA_C,
			"status": "ยังไม่ออก invoice",
		}
		Counthasinvoice = Count_Invoice
		totalhasinvoice = Total_PA
		CountReduce = Count_Invoice_B
		totalReduce = Total_PA_B
		Countnoinvoice = Count_Invoice_C
		totalnoinvoice = Total_PA_C


	status := map[string]interface{}{
		"total": totalhasinvoice + totalReduce + totalnoinvoice,
		"count": Counthasinvoice + CountReduce + Countnoinvoice,
	}
	Result := map[string]interface{}{
		"detail": dataCount,
		"total":  status,
	}
	return c.JSON(http.StatusOK, Result)
	// return c.JSON(http.StatusOK, dataCount)
}

func Billing_Status(c echo.Context) error {
	Startdate := strings.TrimSpace(c.QueryParam("startdate"))
	Enddate := strings.TrimSpace(c.QueryParam("enddate"))
	Staffid := strings.TrimSpace(c.QueryParam("staffid"))
	Search := strings.TrimSpace(c.QueryParam("search"))
	SaleID := strings.TrimSpace(c.QueryParam("saleid"))
	Status := strings.TrimSpace(c.QueryParam("status"))

	if strings.TrimSpace(c.QueryParam("saleid")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

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

	if Startdate == "" || Enddate == ""{
		dayStart = 1
		dateFromA := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
		dateToA := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
		dateFrom =  dateFromA.String()
		dateTo = dateToA.String()
	}else{
		dateFrom = Startdate
		dateTo = Enddate
	}

	type Staffs_Data struct {
		Staff_id 	string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
	}

	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
	}

	type Billing_Data struct {
		PeriodAmount float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Status 		 string `json:"status" gorm:"column:status"`
	}
	dataCount := struct {
		Hasbilling interface{}
		Nobilling  interface{}
	}{}

	Counthasbilling := 0
	Countnobilling := 0

	totalhasbilling := float64(0)
	totalnobilling := float64(0)

	hasErr := 0
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

	var TotalPeriodAmount float64 = 0.0
	CountBilling := 0
	var TotalPeriodAmountB float64 = 0.0
	CountBillingB := 0
	var dataRaw []Billing_Data
	sql := `select BL.PeriodAmount,
	(CASE 
		WHEN bi.status is not null AND bi.status not like ''
		THEN bi.status
		ELSE 'วางไม่ได้'
		END
	) status
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
		AND PeriodStartDate >= ? and PeriodStartDate <= ?
		AND PeriodEndDate >= ? and PeriodEndDate <= ?
		AND PeriodStartDate <= PeriodEndDate
		group by smt.sonumber
	) BL
	LEFT JOIN (select staff_id from staff_info) si on BL.sale_code = si.staff_id
	LEFT JOIN billing_info bi on BL.BLSCDocNo = bi.invoice_no
	WHERE  INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
	INSTR(CONCAT_WS('|',bi.invoice_no,BL.sonumber,bi.reason,BL.Customer_ID,BL.Customer_Name,
	BL.sale_team,BL.sale_name), ?) AND BL.sale_code in (?) AND INSTR(CONCAT_WS('|', bi.status), ?) `
	if err := dbSale.Ctx().Raw(sql,dateTo,dateFrom,dateFrom,dateTo,dateTo,dateFrom,dateTo,dateTo, 
		dateTo,dateFrom,dateTo,dateFrom,dateFrom,dateFrom,dateFrom,dateFrom,dateTo, 
		dateTo,dateFrom,dateFrom,dateTo,dateFrom,dateTo,Staffid,Search,listId,Status).Scan(&dataRaw).Error; err != nil {
		hasErr += 1
	}
	for i, v := range dataRaw {
		if dataRaw[i].Status == "วางบิลแล้ว"{
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}else if dataRaw[i].Status == "วางไม่ได้"{
			TotalPeriodAmountB += v.PeriodAmount
			CountBillingB += 1
		}
	}
	dataCount.Hasbilling = map[string]interface{}{
		"Count":  CountBilling,
		"total":  TotalPeriodAmount,
		"status": "วางบิลแล้ว",
	}
	dataCount.Nobilling = map[string]interface{}{
		"Count":  CountBillingB,
		"total":  TotalPeriodAmountB,
		"status": "วางไม่ได้",
	}
	Counthasbilling = CountBilling
	totalhasbilling = TotalPeriodAmount
	Countnobilling = CountBillingB
	totalnobilling = TotalPeriodAmountB

	status := map[string]interface{}{
		"total": totalhasbilling + totalnobilling,
		"count": Counthasbilling + Countnobilling,
	}
	Result := map[string]interface{}{
		"detail": dataCount,
		"total":  status,
	}
	return c.JSON(http.StatusOK, Result)
	// return c.JSON(http.StatusOK, dataCount)
}

func Reciept_Status(c echo.Context) error {
	Startdate := strings.TrimSpace(c.QueryParam("startdate"))
	Enddate := strings.TrimSpace(c.QueryParam("enddate"))
	Staffid := strings.TrimSpace(c.QueryParam("staffid"))
	SaleID := strings.TrimSpace(c.QueryParam("saleid"))
	Search := strings.TrimSpace(c.QueryParam("search"))
	Status := strings.TrimSpace(c.QueryParam("status"))

	if strings.TrimSpace(c.QueryParam("saleid")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

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

	if Startdate == "" || Enddate == ""{
		dayStart = 1
		dateFromA := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
		dateToA := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
		dateFrom =  dateFromA.String()
		dateTo = dateToA.String()
	}else{
		dateFrom = Startdate
		dateTo = Enddate
	}

	type reciept_Result_Data struct {
		PeriodAmount 	  float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Reciept_status    string  `json:"reciept_status" gorm:"column:reciept_status"`
	}

	dataCount := struct {
		HasReciept interface{}
		NoReciept  interface{}
	}{}

	type Staffs_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
	}

	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
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

	var TotalPeriodAmount float64 = 0.0
	CountReciept := 0
	var TotalPeriodAmountB float64 = 0.0
	CountRecieptB := 0
	var dataRaw []reciept_Result_Data

	sql := `select BL.PeriodAmount,BL.reciept_status
	from
	(select *,
		(CASE
			WHEN INCSCDocNo is not null AND INCSCDocNo not like ''
			THEN 'วาง Reciept เสร็จสิ้น'
			ELSE 'ยังไม่วาง Reciept'
			END
		) reciept_status 
		from so_mssql smt
		WHERE smt.Active_Inactive = 'Active'
		AND PeriodStartDate >= ? and PeriodStartDate <= ?
		AND PeriodEndDate >= ? and PeriodEndDate <= ?
		AND PeriodStartDate <= PeriodEndDate
		group by smt.sonumber
	) BL
	LEFT JOIN (select staff_id from staff_info) si on BL.sale_code = si.staff_id
	LEFT JOIN billing_info bi on BL.BLSCDocNo = bi.invoice_no
	WHERE bi.status like '%วางบิลแล้ว%' AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
	INSTR(CONCAT_WS('|',bi.invoice_no,BL.sonumber,BL.INCSCDocNo,bi.status,bi.reason,BL.Customer_ID,
	BL.Customer_Name,BL.sale_team,BL.sale_name), ?) AND BL.sale_code in (?) AND 
	INSTR(CONCAT_WS('|', BL.reciept_status), ?)`

	if err := dbSale.Ctx().Raw(sql,dateFrom,dateTo,dateFrom,dateTo,
		Staffid,Search,listId,Status).Scan(&dataRaw).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	for i, v := range dataRaw {
		if dataRaw[i].Reciept_status == "วาง Reciept เสร็จสิ้น"{
			TotalPeriodAmount += v.PeriodAmount
			CountReciept += 1
		}else if dataRaw[i].Reciept_status == "ยังไม่วาง Reciept"{
			TotalPeriodAmountB += v.PeriodAmount
			CountRecieptB += 1
		}
	}

	dataCount.HasReciept = map[string]interface{}{
		"Count":  CountReciept,
		"total":  TotalPeriodAmount,
		"status": "วาง Reciept เสร็จสิ้น",
	}
	dataCount.NoReciept = map[string]interface{}{
		"Count":  CountRecieptB,
		"total":  TotalPeriodAmountB,
		"status": "ยังไม่วาง Reciept",
	}
	status := map[string]interface{}{
		"total": TotalPeriodAmount+TotalPeriodAmountB,
		"count": CountReciept + CountRecieptB,
	}
	Result := map[string]interface{}{
		"detail": dataCount,
		"total":  status,
	}

	return c.JSON(http.StatusOK, Result)

	// return c.JSON(http.StatusOK, Reciept_Result_Data)
}
