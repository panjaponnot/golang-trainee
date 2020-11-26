package api

import (
	"fmt"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func GetSummaryQuotationEndPoint(c echo.Context) error {

	type QuotationJoin struct {
		DocNumberEform  string    `json:"doc_number_eform"`
		Service         string    `json:"service"`
		EmployeeCode    string    `json:"employee_code"`
		SaleName        string    `json:"sale_name" gorm:"column:salename"`
		CompanyName     string    `json:"company_name"`
		Team            string    `json:"team"`
		Total           float64   `json:"total" `
		TotalDiscount   float64   `json:"total_discount"`
		TotalPrice      float64   `json:"total_price"`
		StartDate       time.Time `json:"start_date"`
		EndDate         time.Time `json:"end_date"`
		RefQuotation    string    `json:"ref_quotation"`
		RefSO           string    `json:"ref_so" gorm:"column:refSO"`
		DateTime        string    `json:"datetime" gorm:"column:datetime"`
		ServicePlatform string    `json:"service_platform"`
		Reason          string    `json:"reason"`
		Status          string    `json:"status" gorm:"column:status_sale"`
	}

	year := strings.TrimSpace(c.QueryParam("year"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("year")), 10); err == nil {
			yearDefault = time.Unix(util.ConvertTimeStamp(f), 0)
		}
		years, _, _ := yearDefault.Date()
		year = strconv.Itoa(years)
	}

	if strings.TrimSpace(c.QueryParam("id")) == "" {
		return echo.ErrBadRequest
	}

	id := strings.TrimSpace(c.QueryParam("id"))
	var quarter string
	var month string
	var search string
	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
	if strings.TrimSpace(c.QueryParam("quarter")) != "" {
		quarter = fmt.Sprintf("AND quarter(start_date) = %s", strings.TrimSpace(c.QueryParam("quarter")))
	}
	if strings.TrimSpace(c.QueryParam("month")) != "" {
		month = fmt.Sprintf("AND MONTH(start_date) = %s", strings.TrimSpace(c.QueryParam("month")))
	}
	if strings.TrimSpace(c.QueryParam("search")) != "" {
		search = fmt.Sprintf("AND INSTR(CONCAT_WS('|', company_name, service, employee_code, salename, team), %s)", strings.TrimSpace(c.QueryParam("search")))
	}

	dataResult := struct {
		Total  interface{} `json:"total"`
		Detail interface{} `json:"detail"`
	}{}
	dataCount := struct {
		Count        int
		Total        int
		Work         int
		NotWork      int
		Win          int
		Lost         int
		Resend       int
		ReasonWin    interface{}
		ReasonLost   interface{}
		CountService interface{}
		CountCompany interface{}
		CountType    interface{}
		CountTeam    interface{}
	}{}

	var user []m.UserInfo
	if err := dbSale.Ctx().Raw(` SELECT * FROM user_info WHERE role = 'admin' AND one_id = ? `, id).Scan(&user).Error; err != nil {
		log.Errorln(pkgName, err, "User Not Found")
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorln(pkgName, err, "Select user Error")
			return echo.ErrInternalServerError
		}
	}
	textStaffId := ""
	var listStaffId []string

	if len(user) == 0 {
		staff := struct {
			StaffId    string `json:"staff_id"`
			StaffChild string `json:"staff_child"`
		}{}
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where one_id = ?`, id).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
			return c.JSON(http.StatusNotFound, m.Result{Message: "Staff Not Found"})
		}

		if strings.TrimSpace(staff.StaffChild) != "" {
			raw := strings.Split(staff.StaffChild, ",")
			for _, id := range raw {
				listStaffId = append(listStaffId, id)
			}
			listStaffId = append(listStaffId, staff.StaffId)
		} else {
			listStaffId = append(listStaffId, staff.StaffId)
		}
		textStaffId = fmt.Sprintf("AND employee_code IN (%s)", strings.Join(listStaffId, ","))
	}
	log.Infoln(textStaffId)

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(12)
	go func() {
		// work
		var dataRaw []QuotationJoin
		sql := fmt.Sprintf(`SELECT *,(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE sales_approve.reason IS NOT NULL AND sales_approve.status_sale IS NOT NULL
		AND quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL 
		AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.Work = len(dataRaw)
		wg.Done()
	}()
	go func() {
		// total all
		var dataRaw []QuotationJoin
		sql := fmt.Sprintf(`SELECT *,(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve WHERE status IN ('Win','Lost','Resend/Revised')) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE  quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		if len(dataRaw) > (page * 20) {
			start := (page - 1) * 20
			end := (page * 20)
			dataResult.Detail = map[string]interface{}{
				"data":  dataRaw[start:end],
				"count": len(dataRaw[start:end]),
			}
		} else {
			start := (page * 20) - (20)
			dataResult.Detail = map[string]interface{}{
				"data":  dataRaw[start:],
				"count": len(dataRaw[start:]),
			}
		}
		dataCount.Total = len(dataRaw)
		wg.Done()
	}()
	go func() {
		// not work
		var dataRaw []QuotationJoin
		sql := fmt.Sprintf(`SELECT *,(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
WHERE quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL 
AND (total IS NOT NULL OR total_discount IS NOT NULL) AND sales_approve.status_sale IS NULL 
AND YEAR(start_date) = ? %s %s %s %s`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.NotWork = len(dataRaw)
		wg.Done()
	}()
	go func() {
		// win
		var dataRaw []QuotationJoin
		sql := fmt.Sprintf(`SELECT *,(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE sales_approve.reason IS NOT NULL AND sales_approve.status_sale = 'win'
		AND quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL 
		AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}

		dataCount.Win = len(dataRaw)
		wg.Done()
	}()
	go func() {
		// lost
		var dataRaw []QuotationJoin
		sql := fmt.Sprintf(`SELECT *,(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE sales_approve.reason IS NOT NULL AND sales_approve.status_sale = 'lost'
		AND quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL 
		AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.Lost = len(dataRaw)
		wg.Done()
	}()
	go func() {
		// reason win
		var dataRaw []struct {
			TotalReason int    `json:"total_reason_win" gorm:"column:total_reason_win"`
			Reason      string `json:"reason"`
		}
		sql := fmt.Sprintf(`SELECT sales_approve.reason,COUNT(sales_approve.reason) as total_reason_win FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE sales_approve.reason IS NOT NULL AND sales_approve.status_sale = 'win'
		AND quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL 
		AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s  GROUP BY sales_approve.reason`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.ReasonWin = dataRaw
		wg.Done()
	}()
	go func() {
		// reason lost
		var dataRaw []struct {
			TotalReason int    `json:"total_reason_lost" gorm:"column:total_reason_lost"`
			Reason      string `json:"reason"`
		}
		sql := fmt.Sprintf(`SELECT sales_approve.reason,COUNT(sales_approve.reason) as total_reason_lost FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE sales_approve.reason IS NOT NULL AND sales_approve.status_sale = 'lost'
		AND quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL 
		AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s  GROUP BY sales_approve.reason`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.ReasonLost = dataRaw
		wg.Done()
	}()
	go func() {
		// count service
		var dataRaw []struct {
			TotalPrice   float64 `json:"total_price"`
			TotalService int     `json:"total_service"`
			Service      string  `json:"service"`
		}
		sql := fmt.Sprintf(`SELECT SUM(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price,COUNT(service) as total_service,service FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE  quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL AND (total IS NOT NULL OR total_discount IS NOT NULL) 
		AND YEAR(start_date) = ? %s %s %s %s  GROUP BY service`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.CountService = dataRaw
		wg.Done()
	}()
	go func() {
		// count service
		var dataRaw []struct {
			TotalPrice   float64 `json:"total_price"`
			TotalCompany int     `json:"total_company"`
			CompanyName  string  `json:"company_name"`
		}
		sql := fmt.Sprintf(`SELECT SUM(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price,COUNT(company_name) as total_company,company_name FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE  quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL AND (total IS NOT NULL OR total_discount IS NOT NULL) 
		AND YEAR(start_date) = ? %s %s %s %s  GROUP BY company_name`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.CountCompany = dataRaw
		wg.Done()
	}()
	go func() {
		// count service
		var dataRaw []struct {
			TotalPrice float64 `json:"total_price"`
			TotalType  int     `json:"total_type"`
			Type       string  `json:"type"`
		}
		sql := fmt.Sprintf(`SELECT SUM(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price,COUNT(type) as total_type,type FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE  quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL AND (total IS NOT NULL OR total_discount IS NOT NULL) 
		AND YEAR(start_date) = ? %s %s %s %s  GROUP BY type`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.CountType = dataRaw
		wg.Done()
	}()
	go func() {
		// count service
		var dataRaw []struct {
			TotalPrice float64 `json:"total_price"`
			TotalTeam  int     `json:"total_team"`
			Teams      string  `json:"teams"`
		}
		sql := fmt.Sprintf(`SELECT SUM(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price,COUNT(team) as total_team,team ,(CASE
			WHEN team = '' THEN 'no name'
			ELSE team END
			) as teams FROM quatation_th 
			LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
			ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
			WHERE  quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL AND (total IS NOT NULL OR total_discount IS NOT NULL) 
			AND YEAR(start_date) = ? %s %s %s %s  GROUP BY team`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		dataCount.CountTeam = dataRaw
		wg.Done()
	}()
	go func() {
		// re send
		var dataRaw []QuotationJoin
		sql := fmt.Sprintf(`SELECT *,(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price FROM quatation_th 
		LEFT JOIN (SELECT doc_number_eform,reason,status as status_sale FROM sales_approve) as sales_approve 
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform 
		WHERE sales_approve.reason IS NOT NULL AND sales_approve.status_sale = 'Resend/Revised'
		AND quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL 
		AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s`, textStaffId, quarter, month, search)
		if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}

		dataCount.Resend = len(dataRaw)
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	dataResult.Total = map[string]interface{}{
		"total_all": dataCount.Total,
		"total_work": map[string]int{
			"all":       dataCount.Work,
			"win":       dataCount.Win,
			"lost":      dataCount.Lost,
			"not_check": dataCount.NotWork,
			"resend":    dataCount.Resend,
		},
		"reason_win":  dataCount.ReasonWin,
		"reason_lost": dataCount.ReasonLost,
		"service":     dataCount.CountService,
		"company":     dataCount.CountCompany,
		"type":        dataCount.CountType,
		"team":        dataCount.CountTeam,
	}
	return c.JSON(http.StatusOK, dataResult)
}

func CreateLogQuotation(c echo.Context) error {
	bodyData := []struct {
		OneId          string `json:"one_id"`
		StaffId        string `json:"staff_id"`
		Status         string `json:"status"`
		DocNumberEfrom string `json:"doc_number_eform"`
		Remark         string `json:"remark"`
		UserName       string `json:"user_name"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return echo.ErrBadRequest
	}

	for _, body := range bodyData {

		var sale m.SaleApprove
		if err := dbQuataion.Ctx().Model(&m.SaleApprove{}).Where(m.SaleApprove{DocNumberEfrom: body.DocNumberEfrom}).Attrs(m.SaleApprove{
			Reason:         strings.TrimSpace(body.Remark),
			DocNumberEfrom: strings.TrimSpace(body.DocNumberEfrom),
			Status:         strings.TrimSpace(body.Status),
			CreateAt:       time.Now(),
		}).FirstOrCreate(&sale).Error; err != nil {
			log.Errorln(pkgName, err, "Create sale approve error :-")
		}

		if sale.Status != "" && sale.DocNumberEfrom != "" {
			sale.Status = strings.TrimSpace(body.Status)
			sale.Reason = strings.TrimSpace(body.Remark)
			sale.CreateAt = time.Now()
			if err := dbQuataion.Ctx().Save(&sale).Error; err != nil {
				log.Errorln(pkgName, err, "save sale approve error :-")
				return echo.ErrInternalServerError
			}
		}

		d := time.Now()
		var quoLog m.QuotationLog
		quoLog.Date = d.Format("2006-Jan-02")
		quoLog.DocNumberEfrom = body.DocNumberEfrom
		quoLog.UserName = body.UserName
		quoLog.OneId = body.OneId
		quoLog.StaffId = body.StaffId
		quoLog.Status = body.Status
		quoLog.Remark = body.Remark

		if err := dbQuataion.Ctx().Model(&m.QuotationLog{}).Create(&quoLog).Error; err != nil {
			log.Errorln(pkgName, err, "create quotation log error :-")
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "create quotation log error"})
		}
	}
	return c.JSON(http.StatusNoContent, nil)
}
