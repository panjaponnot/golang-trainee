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
	
	if Startdate == "" || Enddate == ""{
		dayStart = 1
	}

	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`
		

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
		AND ci.StartDate_P1 <= ? and ci.EndDate_P1 >= ? and ci.StartDate_P1 <= ci.EndDate_P1`

		if err := dbSale.Ctx().Raw(sql,listId,StaffID,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
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
	
	if Startdate == "" || Enddate == ""{
		dayStart = 1
	}

	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

	type Invoice_Data struct {
		PeriodAmount float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
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

	go func() {
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		var dataRaw []Invoice_Data
		sql := `select smt.PeriodAmount from so_mssql smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is not null AND smt.GetCN not like '' AND smt.sale_code in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND
		INSTR(CONCAT_WS('|', smt.sonumber,smt.SDPropertyCS28,smt.SoWebStatus,smt.BLSCDocNo,smt.GetCN,
		smt.INCSCDocNo,smt.Customer_ID,smt.Customer_Name,smt.sale_code,smt.sale_name,smt.sale_team,
		smt.sale_lead,smt.Active_Inactive,smt.so_refer), ?) and smt.PeriodStartDate <= ? and smt.PeriodEndDate >= ? 
		and smt.PeriodStartDate <= smt.PeriodEndDate`
		sql = sql + ` GROUP BY smt.sonumber` 

		if err := dbSale.Ctx().Raw(sql,listId,Staffid,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.Reduce = map[string]interface{}{
			"Count":  Count_Invoice,
			"total":  Total_PA,
			"status": "ลดหนี้",
		}
		CountReduce = Count_Invoice
		totalReduce = Total_PA
		wg.Done()
	}()
	go func() {
		var dataRaw []Invoice_Data
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		sql := `select smt.PeriodAmount from so_mssql smt 
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id 
		where smt.GetCN is null AND smt.BLSCDocNo is not null AND smt.BLSCDocNo not like '' AND 
		smt.sale_code in (?) AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', smt.sonumber,smt.SDPropertyCS28,smt.SoWebStatus,smt.BLSCDocNo,smt.GetCN,
		smt.INCSCDocNo,smt.Customer_ID,smt.Customer_Name,smt.sale_code,smt.sale_name,smt.sale_team,
		smt.sale_lead,smt.Active_Inactive,smt.so_refer), ?) and smt.PeriodStartDate <= ? 
		and smt.PeriodEndDate >= ? and smt.PeriodStartDate <= smt.PeriodEndDate
		OR smt.GetCN like '' AND smt.BLSCDocNo is not null AND smt.BLSCDocNo not like '' AND 
		smt.sale_code in (?) AND INSTR(CONCAT_WS('|', si.staff_id), ?) AND 
		INSTR(CONCAT_WS('|', smt.sonumber,smt.SDPropertyCS28,smt.SoWebStatus,smt.BLSCDocNo,smt.GetCN,
		smt.INCSCDocNo,smt.Customer_ID,smt.Customer_Name,smt.sale_code,smt.sale_name,smt.sale_team,
		smt.sale_lead,smt.Active_Inactive,smt.so_refer), ?) and smt.PeriodStartDate <= ? 
		and smt.PeriodEndDate >= ? and smt.PeriodStartDate <= smt.PeriodEndDate`

		if err := dbSale.Ctx().Raw(sql,listId,Staffid,Search,dateTo,dateFrom,
			listId,Staffid,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.Hasinvoice = map[string]interface{}{
			"Count":  Count_Invoice,
			"total":  Total_PA,
			"status": "ออก invoice เสร้จสิ้น",
		}
		Counthasinvoice = Count_Invoice
		totalhasinvoice = Total_PA
		wg.Done()
	}()
	go func() {
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		var dataRaw []Invoice_Data
		sql := `select smt.PeriodAmount from so_mssql smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is null AND smt.BLSCDocNo is null AND smt.sale_code in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND
		INSTR(CONCAT_WS('|', smt.sonumber,smt.SDPropertyCS28,smt.SoWebStatus,smt.BLSCDocNo,smt.GetCN,
		smt.INCSCDocNo,smt.Customer_ID,smt.Customer_Name,smt.sale_code,smt.sale_name,smt.sale_team,
		smt.sale_lead,smt.Active_Inactive,smt.so_refer), ?) and smt.PeriodStartDate <= ? and smt.PeriodEndDate >= ? 
		and smt.PeriodStartDate <= smt.PeriodEndDate`
		sql = sql + ` OR smt.GetCN like '' AND smt.BLSCDocNo is null AND smt.sale_code in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND
		INSTR(CONCAT_WS('|', smt.sonumber,smt.SDPropertyCS28,smt.SoWebStatus,smt.BLSCDocNo,smt.GetCN,
		smt.INCSCDocNo,smt.Customer_ID,smt.Customer_Name,smt.sale_code,smt.sale_name,smt.sale_team,
		smt.sale_lead,smt.Active_Inactive,smt.so_refer), ?) and smt.PeriodStartDate <= ? and smt.PeriodEndDate >= ? 
		and smt.PeriodStartDate <= smt.PeriodEndDate`
		sql = sql + ` OR smt.GetCN like '' AND smt.BLSCDocNo like '' AND smt.sale_code in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND
		INSTR(CONCAT_WS('|', smt.sonumber,smt.SDPropertyCS28,smt.SoWebStatus,smt.BLSCDocNo,smt.GetCN,
		smt.INCSCDocNo,smt.Customer_ID,smt.Customer_Name,smt.sale_code,smt.sale_name,smt.sale_team,
		smt.sale_lead,smt.Active_Inactive,smt.so_refer), ?) and smt.PeriodStartDate <= ? and smt.PeriodEndDate >= ? 
		and smt.PeriodStartDate <= smt.PeriodEndDate`
		sql = sql + ` OR smt.GetCN is null AND smt.BLSCDocNo like '' AND smt.sale_code in (?) AND 
		INSTR(CONCAT_WS('|', si.staff_id), ?) AND
		INSTR(CONCAT_WS('|', smt.sonumber,smt.SDPropertyCS28,smt.SoWebStatus,smt.BLSCDocNo,smt.GetCN,
		smt.INCSCDocNo,smt.Customer_ID,smt.Customer_Name,smt.sale_code,smt.sale_name,smt.sale_team,
		smt.sale_lead,smt.Active_Inactive,smt.so_refer), ?) and smt.PeriodStartDate <= ? and smt.PeriodEndDate >= ? 
		and smt.PeriodStartDate <= smt.PeriodEndDate`
		sql = sql + ` GROUP BY smt.sonumber `

		if err := dbSale.Ctx().Raw(sql,listId,Staffid,Search,dateTo,dateFrom,
			listId,Staffid,Search,dateTo,dateFrom,
			listId,Staffid,Search,dateTo,dateFrom,
			listId,Staffid,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.Noinvoice = map[string]interface{}{
			"Count":  Count_Invoice,
			"total":  Total_PA,
			"status": "ยังไม่ออก invoice",
		}
		Countnoinvoice = Count_Invoice
		totalnoinvoice = Total_PA
		wg.Done()
	}()
	wg.Wait()

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

	if Startdate == "" || Enddate == ""{
		dayStart = 1
	}

	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

	type Staffs_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
		Staff_child string `json:"staff_child" gorm:"column:staff_child"`
	}

	type Users_Data struct {
		Staff_id string `json:"staff_id" gorm:"column:staff_id"`
	}

	type Billing_Data struct {
		PeriodAmount float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
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
	wg := sync.WaitGroup{}
	wg.Add(2)

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
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql := `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางบิลแล้ว%' AND smt.sale_code in (?) AND 
			INSTR(CONCAT_WS('|', si.staff_id), ?) AND
			INSTR(CONCAT_WS('|', bi.invoice_no,bi.so_number,bi.status,bi.reason), ?) 
			and smt.PeriodStartDate <= ? and smt.PeriodEndDate >= ? and smt.PeriodStartDate <= smt.PeriodEndDate`

		sql = sql + ` group by smt.sonumber) BL`
		if err := dbSale.Ctx().Raw(sql,listId,Staffid,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.Hasbilling = map[string]interface{}{
			"Count":  CountBilling,
			"total":  TotalPeriodAmount,
			"status": "วางบิลแล้ว",
		}

		Counthasbilling = CountBilling
		totalhasbilling = TotalPeriodAmount
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql := `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' AND smt.sale_code in (?) AND 
			INSTR(CONCAT_WS('|', si.staff_id), ?) AND
			INSTR(CONCAT_WS('|', bi.invoice_no,bi.so_number,bi.status,bi.reason), ?) 
			and smt.PeriodStartDate <= ? and smt.PeriodEndDate >= ? and smt.PeriodStartDate <= smt.PeriodEndDate`
		
		sql = sql + ` group by smt.sonumber) BL`
		if err := dbSale.Ctx().Raw(sql,listId,Staffid,Search,dateTo,dateFrom).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.Nobilling = map[string]interface{}{
			"Count":  CountBilling,
			"total":  TotalPeriodAmount,
			"status": "วางไม่ได้",
		}
		Countnobilling = CountBilling
		totalnobilling = TotalPeriodAmount
		wg.Done()
	}()
	wg.Wait()

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

	if Startdate == "" || Enddate == ""{
		dayStart = 1
	}

	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

	Reciept_Data := []struct {
		PeriodAmount   float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Count_Reciept  int     `json:"Count_Reciept" gorm:"column:Count_Reciept"`
		PeriodAmountF  float64 `json:"PeriodAmountF" gorm:"column:PeriodAmountF"`
		Count_RecieptF int     `json:"Count_RecieptF" gorm:"column:Count_RecieptF"`
		// Invoice_status_name	string	`json:"invoice_status_name" gorm:"column:invoice_status_name"`
		// INCSCDocNo			string	`json:"INCSCDocNo" gorm:"column:INCSCDocNo"`
	}{}

	type reciept_Result_Data struct {
		CountReciept      int     `json:"CountReciept"`
		TotalPeriodAmount float64 `json:"TotalPeriodAmount"`
		Reciept_status    string  `json:"Reciept_status"`
	}

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

	var Reciept_Result_Data []reciept_Result_Data

	sql := `select
	SUM(CASE
		WHEN RE.INCSCDocNo is not null AND RE.INCSCDocNo NOT LIKE ''
		THEN RE.PeriodAmount
		ELSE NULL
	END) as PeriodAmount, 
	COUNT(CASE
		WHEN RE.INCSCDocNo is not null AND RE.INCSCDocNo NOT LIKE ''
		THEN 1
		ELSE NULL
	END) as Count_Reciept,
	SUM(CASE
		WHEN RE.INCSCDocNo is null OR RE.INCSCDocNo LIKE ''
		THEN RE.PeriodAmount
		ELSE NULL
	END) as PeriodAmountF, 
	COUNT(CASE
		WHEN RE.INCSCDocNo is null OR RE.INCSCDocNo LIKE ''
		THEN 1
		ELSE NULL
	END) as Count_RecieptF
	from
	(select smt.PeriodAmount,smt.INCSCDocNo
	from so_mssql smt
	LEFT JOIN staff_info si on smt.sale_code = si.staff_id
	LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
	where bi.status like '%วางบิลแล้ว%' AND smt.sale_code in (?) AND
	INSTR(CONCAT_WS('|', si.staff_id), ?) AND
	INSTR(CONCAT_WS('|', bi.invoice_no,bi.so_number,bi.status,bi.reason), ?) 
	and PeriodStartDate <= ? and PeriodEndDate >= ? and PeriodStartDate <= PeriodEndDate`
	sql = sql + ` GROUP BY smt.sonumber)RE`

	if err := dbSale.Ctx().Raw(sql,listId,Staffid,Search,dateTo,dateFrom).Scan(&Reciept_Data).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	DataA := reciept_Result_Data{
		CountReciept:      Reciept_Data[0].Count_Reciept,
		TotalPeriodAmount: Reciept_Data[0].PeriodAmount,
		Reciept_status:    "วาง Reciept เสร็จสิ้น",
	}
	Reciept_Result_Data = append(Reciept_Result_Data, DataA)

	DataB := reciept_Result_Data{
		CountReciept:      Reciept_Data[0].Count_RecieptF,
		TotalPeriodAmount: Reciept_Data[0].PeriodAmountF,
		Reciept_status:    "ยังไม่วาง Reciept",
	}
	Reciept_Result_Data = append(Reciept_Result_Data, DataB)

	status := map[string]interface{}{
		"total": Reciept_Data[0].PeriodAmount + Reciept_Data[0].PeriodAmountF,
		"count": Reciept_Data[0].Count_RecieptF + Reciept_Data[0].Count_Reciept,
	}
	Result := map[string]interface{}{
		"detail": Reciept_Result_Data,
		"total":  status,
	}
	return c.JSON(http.StatusOK, Result)

	// return c.JSON(http.StatusOK, Reciept_Result_Data)
}
