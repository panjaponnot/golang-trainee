package api

import (
	"bytes"
	"fmt"
	"net/http"
	"sale_ranking/model"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/requests"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"

	"encoding/json"
)

func GetSummaryCustomerEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}
	b, e := strconv.ParseBool(strings.TrimSpace(c.QueryParam("check_amount")))
	if e != nil {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid check amount"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	status := strings.TrimSpace(c.QueryParam("status"))
	// fmt.Println("====> filter", search)
	ds := time.Now()
	de := time.Now()
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("start_date")), 10); err == nil {
		ds = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("end_date")), 10); err == nil {
		de = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	yearStart, monthStart, dayStart := ds.Date()
	yearEnd, monthEnd, dayEnd := de.Date()
	startRange := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	endRange := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	dateFrom := startRange.Format("2006-01-02")
	dateTo := endRange.Format("2006-01-02")
	m := endRange.Sub(startRange)
	if m < 0 {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid date"})
	}

	//// get staff id ////
	var user []model.UserInfo
	if err := dbSale.Ctx().Raw(`SELECT * FROM user_info WHERE staff_id = ? and role = 'admin';`, saleId).Scan(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
	}
	var listId []string
	if len(user) != 0 {
		var staffAll []model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info ;`).Scan(&staffAll).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
			}
		}
		for _, i := range staffAll {
			listId = append(listId, i.StaffId)
		}
	} else {
		var staffAll model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info WHERE staff_id = ?;`, saleId).Scan(&staffAll).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
			}
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
		if staffAll.StaffChild != "" {
			data := strings.Split(staffAll.StaffChild, ",")
			listId = data
		}
		listId = append(listId, staffAll.StaffId)
	}

	sql := `Select * From  ( SELECT DISTINCT Customer_ID as Customer_ID, Customer_Name, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
	sum(so_amount) as so_amount, sum(inv_amount) as inv_amount, sum(cs_amount) as cs_amount, sum(rc_amount) as rc_amount, sum(cn_amount) as cn_amount, sum(amount) as amount, AVG(in_factor) as in_factor,
	sum(in_factor) as sum_if, sum(inv_amount) - sum(rc_amount) as outstainding_amount,sale_code,sale_name,AVG(ex_factor) as ex_factor,sum(ex_factor) as sum_ef, department, nname,
	(CASE
		WHEN sum(inv_amount) = 0 THEN 'ยังไม่ออกใบแจ้งหนี้'
		WHEN sum(inv_amount) = sum(cn_amount) THEN 'ลดหนี้'
		WHEN sum(inv_amount) - sum(cn_amount) <= sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระแล้ว'
		WHEN sum(inv_amount) - sum(cn_amount) > sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระไม่ครบ'
		ELSE 'ค้างชำระ' END
	) as status,
	sum((CASE
		WHEN inv_amount = rc_amount THEN inv_amount
		ELSE inv_amount - cn_amount END
	)) as inv_amount_cal,
	(sum(amount)/sum(amount_engcost)) as sale_factor,
	sonumber_all
	from (
		SELECT
			count(DISTINCT sonumber) as sonumber,
			count(sonumber) as sonumber_all,
			Customer_ID as Customer_ID,
			Customer_Name as Customer_Name,
			count(DISTINCT(CASE WHEN SDPropertyCS28 !='' THEN SDPropertyCS28 END)) as csnumber,
			count(DISTINCT(CASE WHEN BLSCDocNo !='' THEN BLSCDocNo END)) as invnumber,
			count(DISTINCT(CASE WHEN INCSCDocNo !='' THEN INCSCDocNo END)) as rcnumber,
			count(DISTINCT(CASE WHEN GetCN !='' THEN GetCN END)) as cnnumber,
			sum(so_amount) as so_amount,
			sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) as inv_amount,
			sum(CASE WHEN SDPropertyCS28 !='' THEN so_amount ELSE 0 END) as cs_amount,
			sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as rc_amount,
			sum(CASE WHEN GetCN !='' THEN so_amount ELSE 0 END) as cn_amount,
			sum(PeriodAmount) as amount,
			sum(eng_cost) as amount_engcost,
			sale_factor,
			in_factor,sale_code,sale_name,ex_factor
			FROM (
				SELECT
					SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
					sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,
					(case
						when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
						else 0 end
					) as eng_cost,
					(CASE
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
					) as so_amount
				FROM (
					SELECT * FROM so_mssql
					WHERE Active_Inactive = 'Active'
					and PeriodStartDate <= ? and PeriodEndDate >= ?
					and PeriodStartDate <= PeriodEndDate
					and sale_code in (?)
				) sub_data
			) so_group
			WHERE so_amount <> 0 group by sonumber
		) cust_group
		LEFT JOIN staff_info ON cust_group.sale_code = staff_info.staff_id
		group by Customer_ID ) as a
		where INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code, nname, department), ?) AND INSTR(CONCAT_WS('|', status), ?);`
	var sum []model.SummaryCustomer
	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, status).Scan(&sum).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
		}
		return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
	}

	if b {
		sqlAmount := `SELECT
		count(DISTINCT sonumber) as total_so,
		count(DISTINCT(CASE WHEN SDPropertyCS28 !='' THEN SDPropertyCS28 END)) as total_cs,
		count(DISTINCT(CASE WHEN BLSCDocNo !='' THEN BLSCDocNo END)) as total_inv,
		count(DISTINCT(CASE WHEN INCSCDocNo !='' THEN INCSCDocNo END)) as total_rc,
		count(DISTINCT(CASE WHEN GetCN !='' THEN GetCN END)) as total_cn,
		sum(so_amount) as so_amount,
		sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) as inv_amount,
		sum(CASE WHEN SDPropertyCS28 !='' THEN so_amount ELSE 0 END) as cs_amount,
		sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as rc_amount,
		sum(CASE WHEN GetCN !='' THEN so_amount ELSE 0 END) as cn_amount,
		sum(PeriodAmount) as amount,
		sum(eng_cost) as amount_engcost,
		(sum(PeriodAmount)/sum(eng_cost)) as sale_factor,
		sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) - sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as outstanding_total,
		count(sonumber) as total_all_so,
		sum(CASE WHEN status_so = 'ยังไม่ออกใบแจ้งหนี้' THEN 1 ELSE 0 END) as total_status_noinv,
		sum(CASE WHEN status_so = 'ออกใบแจ้งหนี้' THEN 1 ELSE 0 END) as total_status_haveinv,
		sum(CASE WHEN status_so = 'ลดหนี้' THEN 1 ELSE 0 END) as total_status_havecn,
		sum(CASE WHEN status_so = 'ยังไม่ออกใบแจ้งหนี้' THEN so_amount ELSE 0 END) as amount_status_noinv,
		sum(CASE WHEN status_so = 'ออกใบแจ้งหนี้' THEN so_amount ELSE 0 END) as amount_status_haveinv,
		sum(CASE WHEN status_so = 'ลดหนี้' THEN so_amount ELSE 0 END) as amount_status_havecn,
		sum(CASE WHEN status_incoome = 'ค้างชำระ' THEN 1 ELSE 0 END) as total_status_norc,
		sum(CASE WHEN status_incoome = 'ชำระแล้ว' THEN 1 ELSE 0 END) as total_status_haverc,
		sum(CASE WHEN status_incoome = 'ค้างชำระ' THEN so_amount ELSE 0 END) as amount_status_norc,
		sum(CASE WHEN status_incoome = 'ชำระแล้ว' THEN so_amount ELSE 0 END) as amount_status_haverc
		,nname, department
		FROM (
			SELECT
				SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
				sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor,
				(case
					when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
					else 0 end
				) as eng_cost,
				staff_info.nname, staff_info.department,
				(CASE
					WHEN sonumber <> '' AND BLSCDocNo = '' THEN 'ยังไม่ออกใบแจ้งหนี้'
					WHEN sonumber <> '' AND BLSCDocNo <> '' AND GetCN = '' THEN 'ออกใบแจ้งหนี้'
					WHEN sonumber <> '' AND BLSCDocNo <> '' AND GetCN <> '' AND INCSCDocNo = '' THEN 'ลดหนี้'
					ELSE 'ออกใบแจ้งหนี้' END
				) as status_so,
				(CASE
					WHEN sonumber <> '' AND BLSCDocNo <> '' AND GetCN = '' AND INCSCDocNo = '' THEN 'ค้างชำระ'
					WHEN sonumber <> '' AND BLSCDocNo <> '' AND INCSCDocNo <> '' THEN 'ชำระแล้ว'
					ELSE '' END
				) as status_incoome,
				(CASE
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
				) as so_amount
			FROM (
				SELECT * FROM so_mssql
				WHERE Active_Inactive = 'Active'
				and PeriodStartDate <= ? and PeriodEndDate >= ?
				and PeriodStartDate <= PeriodEndDate
				and sale_code in (?)
			) sub_data LEFT JOIN staff_info ON sub_data.sale_code = staff_info.staff_id
		) so_group
		WHERE so_amount <> 0 AND INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code, nname, department), ?) `
		var sumAmount model.SummaryCustomer
		if err := dbSale.Ctx().Raw(sqlAmount, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search).Scan(&sumAmount).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
			}
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
		sumInFac := 0.0
		sumExFac := 0.0

		if sumAmount.TotalSo != 0 {
			sumInFac = sumIF(sum) / sumAmount.TotalSo
			sumExFac = sumEF(sum) / sumAmount.TotalSo
		}

		dataMap := map[string]interface{}{
			"data":           sumAmount,
			"customer_total": len(sum),
			"in_factor":      sumInFac,
			"ex_factor":      sumExFac,
			"detail":         sum,
		}

		return c.JSON(http.StatusOK, dataMap)
	}
	return c.JSON(http.StatusOK, sum)
}

func sumIF(input []model.SummaryCustomer) float64 {
	sum := 0.0

	for _, i := range input {
		sum += i.SumIf
	}

	fmt.Println("sum was if", sum)
	return sum
}
func sumEF(input []model.SummaryCustomer) float64 {
	sum := 0.0

	for _, i := range input {
		sum += i.SumEf
	}

	fmt.Println("sum was ef", sum)
	return sum
}

func GetSummaryPendingSOEndPoint(c echo.Context) error {
	// if strings.TrimSpace(c.Param("id")) == "" {
	// 	return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	// }
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))
	year := strings.TrimSpace(c.QueryParam("year"))
	month := strings.TrimSpace(c.QueryParam("month"))
	// search := strings.TrimSpace(c.QueryParam("search"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		year = strconv.Itoa(yearDefault.Year())
	}

	if strings.TrimSpace(c.QueryParam("month")) == "" {
		monthDefault := time.Now()
		month = strconv.Itoa(int(monthDefault.Month()))
	}
	// oneId := strings.TrimSpace(c.Param("id"))
	// staff := []struct {
	// 	ContractEndDate string `json:"ContractEndDate" gorm:"column:ContractEndDate"`
	// 	Status          string `json:"status" gorm:"column:status"`
	// 	Remark          string `json:"remark" gorm:"column:remark"`
	// 	Days            int    `json:"days" gorm:"column:days"`
	// }{}
	// if err := dbSale.Ctx().Raw(` SELECT so_mssql.ContractEndDate,check_expire.status,check_expire.remark,DATEDIFF(so_mssql.ContractEndDate, NOW()) as days from staff_info
	// join so_mssql on so_mssql.sale_code = staff_info.staff_id
	// left join check_expire on check_expire.sonumber = so_mssql.sonumber
	//  WHERE staff_info.one_id = ? `, oneId).Scan(&staff).Error; err != nil {
	// 	log.Errorln(pkgName, err, "Select staff error")
	// 	return echo.ErrInternalServerError
	// }
	type PendingDataSum struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
		SaleLead            string  `json:"sale_lead" gorm:"column:sale_lead"`
		Day                 string  `json:"day" gorm:"column:days"`
		SoMonth             string  `json:"so_month" gorm:"column:so_month"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		PeriodAmount        float64 `json:"period_amount" gorm:"column:PeriodAmount"`
		TotalAmount         float64 `json:"total_amount" gorm:"column:TotalAmount"`
		StaffId             string  `json:"staff_id" gorm:"column:staff_id"`
		PayType             string  `json:"pay_type" gorm:"column:pay_type"`
		SoType              string  `json:"so_type" gorm:"column:so_type"`
		Prefix              string  `json:"prefix"`
		Fname               string  `json:"fname"`
		Lname               string  `json:"lname"`
		Nname               string  `json:"nname"`
		Position            string  `json:"position"`
		Department          string  `json:"department"`
		Status              string  `json:"status"`
		Remark              string  `json:"remark"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
	}

	type PendingDataSumV2 struct {
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
	}

	Active := 0.0
	// var Active int64
	// var Update int64
	// var NotUpdate int64
	Update := 0.0
	NotUpdate := 0.0
	var DataActive []PendingDataSumV2
	var DataUpdate []PendingDataSumV2
	var DataNotUpdate []PendingDataSumV2
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		// if err := dbSale.Ctx().Raw(` SELECT * from staff_info
		// 	join so_mssql on so_mssql.sale_code = staff_info.staff_id
		// 	left join check_expire on check_expire.sonumber = so_mssql.sonumber
		// 	 WHERE staff_info.one_id = ?
		// 	 AND YEAR(ContractEndDate) >= ?
		// 	 AND MONTH(ContractEndDate) > ?
		// 	 GROUP BY so_mssql.sonumber`, oneId, year, month).Scan(&DataActive).Error; err != nil {
		DateStr := fmt.Sprintf(`%s-%s-01`, year, month)
		if err := dbSale.Ctx().Raw(` SELECT TotalContractAmount from so_mssql
			left join check_expire on check_expire.sonumber = so_mssql.sonumber
			WHERE ContractEndDate > ?
			AND INSTR(CONCAT_WS('|', sale_code), ?)
			 GROUP BY so_mssql.sonumber`, DateStr, StaffId).Scan(&DataActive).Error; err != nil {
			log.Errorln(pkgName, err, "Select DataActive error")
			// return echo.ErrInternalServerError
		}
		if len(DataActive) > 0 {
			for _, d := range DataActive {
				Active += d.TotalContractAmount
			}
		}
		wg.Done()
	}()
	go func() {
		// if err := dbSale.Ctx().Raw(` SELECT * from staff_info
		// 	join so_mssql on so_mssql.sale_code = staff_info.staff_id
		// 	left join check_expire on check_expire.sonumber = so_mssql.sonumber
		// 	 WHERE staff_info.one_id = ?
		// 	 AND YEAR(ContractEndDate) = ?
		// 	 AND MONTH(ContractEndDate) = ?
		// 	 AND check_expire.remark IS NOT NULL
		// 	 AND check_expire.status IS NOT NULL
		// 	 GROUP BY so_mssql.sonumber`, oneId, year, month).Scan(&DataUpdate).Error; err != nil {
		if err := dbSale.Ctx().Raw(` SELECT TotalContractAmount from so_mssql
			left join check_expire on check_expire.sonumber = so_mssql.sonumber
			WHERE YEAR(ContractEndDate) = ?
			 AND MONTH(ContractEndDate) = ?
			 AND check_expire.remark IS NOT NULL
			 AND check_expire.status IS NOT NULL
			 AND INSTR(CONCAT_WS('|', sale_code), ?)
			 GROUP BY so_mssql.sonumber`, year, month, StaffId).Scan(&DataUpdate).Error; err != nil {
			log.Errorln(pkgName, err, "Select DataUpdate error")
			// return echo.ErrInternalServerError
		}
		if len(DataUpdate) > 0 {
			for _, d := range DataUpdate {
				Update += d.TotalContractAmount
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(` SELECT TotalContractAmount from so_mssql
			left join check_expire on check_expire.sonumber = so_mssql.sonumber
			WHERE YEAR(ContractEndDate) = ?
			 AND MONTH(ContractEndDate) = ?
			 AND check_expire.remark IS NULL
			 AND check_expire.status IS NULL
			 AND INSTR(CONCAT_WS('|', sale_code), ?)
			 GROUP BY so_mssql.sonumber`, year, month, StaffId).Scan(&DataNotUpdate).Error; err != nil {
			log.Errorln(pkgName, err, "Select DataNotUpdate error")
			// return echo.ErrInternalServerError
		}
		if len(DataNotUpdate) > 0 {
			for _, d := range DataNotUpdate {
				NotUpdate += d.TotalContractAmount
			}
		}
		wg.Done()
	}()
	wg.Wait()

	dataActive := map[string]interface{}{
		"total_amount": Active,
		"count_so":     len(DataActive),
	}
	dataUpdate := map[string]interface{}{
		"total_amount": Update,
		"count_so":     len(DataUpdate),
	}
	dataNotUpdate := map[string]interface{}{
		"total_amount": NotUpdate,
		"count_so":     len(DataNotUpdate),
	}
	dataRaw := map[string]interface{}{
		"Active":    dataActive,
		"Update":    dataUpdate,
		"NotUpdate": dataNotUpdate,
	}
	return c.JSON(http.StatusOK, dataRaw)
}

func GetContractEndPoint(c echo.Context) error {
	// if strings.TrimSpace(c.Param("id")) == "" {
	// 	return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	// }
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))
	year := strings.TrimSpace(c.QueryParam("year"))
	month := strings.TrimSpace(c.QueryParam("month"))
	// search := strings.TrimSpace(c.QueryParam("search"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		year = strconv.Itoa(yearDefault.Year())
	}

	if strings.TrimSpace(c.QueryParam("month")) == "" {
		monthDefault := time.Now()
		month = strconv.Itoa(int(monthDefault.Month()))
	}
	// oneId := strings.TrimSpace(c.Param("id"))
	type PendingDataSum struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
		SaleLead            string  `json:"sale_lead" gorm:"column:sale_lead"`
		Day                 string  `json:"day" gorm:"column:days"`
		SoMonth             string  `json:"so_month" gorm:"column:so_month"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		PeriodAmount        float64 `json:"period_amount" gorm:"column:PeriodAmount"`
		TotalAmount         float64 `json:"total_amount" gorm:"column:TotalAmount"`
		StaffId             string  `json:"staff_id" gorm:"column:staff_id"`
		PayType             string  `json:"pay_type" gorm:"column:pay_type"`
		SoType              string  `json:"so_type" gorm:"column:so_type"`
		Prefix              string  `json:"prefix"`
		Fname               string  `json:"fname"`
		Lname               string  `json:"lname"`
		Nname               string  `json:"nname"`
		Position            string  `json:"position"`
		Department          string  `json:"department"`
		Status              string  `json:"status"`
		Remark              string  `json:"remark"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
	}

	type PendingDataSumV2 struct {
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
	}

	CheckTrue := 0.0
	CheckFalse := 0.0
	var DataCheckTrue []PendingDataSumV2
	var DataCheckFalse []PendingDataSumV2
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := dbSale.Ctx().Raw(` SELECT TotalContractAmount from so_mssql
			left join check_expire on check_expire.sonumber = so_mssql.sonumber
			 WHERE YEAR(ContractEndDate) = ?
			 AND MONTH(ContractEndDate) = ?
			 AND check_expire.status = '1'
			 AND INSTR(CONCAT_WS('|', sale_code), ?)
			 GROUP BY so_mssql.sonumber`, year, month, StaffId).Scan(&DataCheckTrue).Error; err != nil {
			log.Errorln(pkgName, err, "Select CheckTrue error")
			// AND check_expire.remark IS NOT NULL
			// return echo.ErrInternalServerError
		}
		if len(DataCheckTrue) > 0 {
			for _, d := range DataCheckTrue {
				CheckTrue += d.TotalContractAmount
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(` SELECT TotalContractAmount from so_mssql
			left join check_expire on check_expire.sonumber = so_mssql.sonumber
			 WHERE YEAR(ContractEndDate) = ?
			 AND MONTH(ContractEndDate) = ?
			 AND check_expire.status = '0'
			 AND INSTR(CONCAT_WS('|', sale_code), ?)
			 GROUP BY so_mssql.sonumber`, year, month, StaffId).Scan(&DataCheckFalse).Error; err != nil {
			log.Errorln(pkgName, err, "Select CheckFalse error")
			// AND check_expire.remark IS NULL
			// return echo.ErrInternalServerError
		}
		if len(DataCheckFalse) > 0 {
			for _, d := range DataCheckFalse {
				CheckFalse += d.TotalContractAmount
			}
		}
		wg.Done()
	}()
	wg.Wait()

	dataCheckTrue := map[string]interface{}{
		"total_amount": CheckTrue,
		"count_so":     len(DataCheckTrue),
	}
	dataCheckFalse := map[string]interface{}{
		"total_amount": CheckFalse,
		"count_so":     len(DataCheckFalse),
	}
	dataRaw := map[string]interface{}{
		"CheckTrue":  dataCheckTrue,
		"CheckFalse": dataCheckFalse,
	}
	return c.JSON(http.StatusOK, dataRaw)
}

func GetTeamsEndPoint(c echo.Context) error {
	// if strings.TrimSpace(c.Param("id")) == "" {
	// 	return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	// }
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))
	year := strings.TrimSpace(c.QueryParam("year"))
	month := strings.TrimSpace(c.QueryParam("month"))
	// search := strings.TrimSpace(c.QueryParam("search"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		year = strconv.Itoa(yearDefault.Year())
	}
	if strings.TrimSpace(c.QueryParam("month")) == "" {
		monthDefault := time.Now()
		month = strconv.Itoa(int(monthDefault.Month()))
	}
	// oneId := strings.TrimSpace(c.Param("id"))

	type PendingDataSum struct {
		SaleTeamnumber string  `json:"sale_team" gorm:"column:sale_team"`
		Sum            float64 `json:"sum" gorm:"column:sum"`
		CountSO        string  `json:"CountSO" gorm:"column:CountSO"`
	}

	// CheckTrue := 0
	// CheckFalse := 0
	var DataTeam []PendingDataSum
	// var DataCheckFalse []PendingDataSum
	// if err := dbSale.Ctx().Raw(` SELECT T1.sale_team,sum(T1.sum) as sum ,COUNT(T1.sonumber) as CountSO
	// 	FROM (SELECT so_mssql.sonumber,so_mssql.sale_team,SUM(so_mssql.TotalContractAmount) as sum from so_mssql
	// 				join check_expire on check_expire.sonumber = so_mssql.sonumber
	// 				 AND YEAR(ContractEndDate) = ?
	// 				 AND MONTH(ContractEndDate) = ?
	// 				 AND check_expire.status != '1'
	// 				 AND check_expire.remark IS NULL
	// 				 GROUP BY so_mssql.sale_team,so_mssql.sonumber) AS T1
	// 	GROUP BY T1.sale_team`, year, month).Scan(&DataTeam).Error; err != nil {
	if err := dbSale.Ctx().Raw(` SELECT T1.department as sale_team,sum(T1.sum) as sum ,COUNT(T1.sonumber) as CountSO
		FROM
		(
		 SELECT so_mssql.sonumber,department,SUM(so_mssql.TotalContractAmount) as sum
		 from so_mssql
		 left join check_expire on check_expire.sonumber = so_mssql.sonumber
		 left join staff_info on so_mssql.sale_code = staff_info.staff_id
		 WHERE YEAR(ContractEndDate) = ? AND MONTH(ContractEndDate) = ?
		 AND INSTR(CONCAT_WS('|', sale_code), ?)
		 GROUP BY so_mssql.sale_team,so_mssql.sonumber
		) AS T1
		GROUP BY T1.department`, year, month, StaffId).Scan(&DataTeam).Error; err != nil {
		log.Errorln(pkgName, err, "Select DataTeam error")
		// return echo.ErrInternalServerError
	}
	// if len(DataTeam) > 0 {
	// 	for _, d := range DataTeam {
	// 		CheckTrue += d.TotalContractAmount
	// 	}
	// }
	return c.JSON(http.StatusOK, DataTeam)
}

func GetTeamsDepartmentEndPoint(c echo.Context) error {
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))
	if strings.TrimSpace(c.QueryParam("department")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}
	year := strings.TrimSpace(c.QueryParam("year"))
	month := strings.TrimSpace(c.QueryParam("month"))
	// search := strings.TrimSpace(c.QueryParam("search"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		year = strconv.Itoa(yearDefault.Year())
	}
	if strings.TrimSpace(c.QueryParam("month")) == "" {
		monthDefault := time.Now()
		month = strconv.Itoa(int(monthDefault.Month()))
	}
	department := strings.TrimSpace(c.QueryParam("department"))

	type PendingDataSum struct {
		Fname     string  `json:"fname" gorm:"column:fname"`
		Lname     string  `json:"lname" gorm:"column:lname"`
		Sum       float64 `json:"sum" gorm:"column:sum"`
		CountSO   string  `json:"CountSO" gorm:"column:CountSO"`
		NotUpdate string  `json:"notupdate" gorm:"column:notupdate"`
	}

	// CheckTrue := 0
	// CheckFalse := 0
	var DataTeam []PendingDataSum
	// var DataCheckFalse []PendingDataSum
	// if err := dbSale.Ctx().Raw(` SELECT T1.sale_team,sum(T1.sum) as sum ,COUNT(T1.sonumber) as CountSO
	// 	FROM (SELECT so_mssql.sonumber,so_mssql.sale_team,SUM(so_mssql.TotalContractAmount) as sum from so_mssql
	// 				join check_expire on check_expire.sonumber = so_mssql.sonumber
	// 				 AND YEAR(ContractEndDate) = ?
	// 				 AND MONTH(ContractEndDate) = ?
	// 				 AND check_expire.status != '1'
	// 				 AND check_expire.remark IS NULL
	// 				 GROUP BY so_mssql.sale_team,so_mssql.sonumber) AS T1
	// 	GROUP BY T1.sale_team`, year, month).Scan(&DataTeam).Error; err != nil {
	if err := dbSale.Ctx().Raw(` SELECT T1.fname,T1.lname,sum(T1.sum) as sum ,COUNT(T1.sonumber) as CountSO,SUM(T1.remark IS NULL AND T1.status IS NULL) as notupdate
	FROM
	(SELECT so_mssql.sonumber,department,SUM(so_mssql.TotalContractAmount) as sum ,staff_info.one_id,staff_info.staff_id,staff_info.prefix,staff_info.fname,staff_info.lname,check_expire.remark,check_expire.status
	 from so_mssql
	 left join check_expire on check_expire.sonumber = so_mssql.sonumber
	 left join staff_info on so_mssql.sale_code = staff_info.staff_id
	 WHERE YEAR(ContractEndDate) = ? AND MONTH(ContractEndDate) = ?
	 AND department = ?
	 AND INSTR(CONCAT_WS('|', sale_code), ?)
	 GROUP BY staff_info.staff_id,so_mssql.sonumber) AS T1
	GROUP BY T1.staff_id`, year, month, department, StaffId).Scan(&DataTeam).Error; err != nil {
		log.Errorln(pkgName, err, "Select DataTeam error")
		// return echo.ErrInternalServerError
	}
	// if len(DataTeam) > 0 {
	// 	for _, d := range DataTeam {
	// 		CheckTrue += d.TotalContractAmount
	// 	}
	// }
	return c.JSON(http.StatusOK, DataTeam)
}

func GetVmSummaryEndPoint(c echo.Context) error {
	if strings.TrimSpace(c.QueryParam("so")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}
	so := strings.TrimSpace(c.QueryParam("so"))
	vm := []struct {
		EquipmentCode string `json:"EquipmentCode" gorm:"column:EquipmentCode"`
		Cpu           string `json:"cpu" gorm:"column:cpu"`
		Ram           string `json:"ram" gorm:"column:ram"`
		DiskStorage   string `json:"disk_storage" gorm:"column:disk_storage"`
		SOnumber      string `json:"sonumber" gorm:"column:sonumber"`
	}{}
	if err := dbEquip.Ctx().Raw(`  select
				EquipmentCode,
				sum(cast(cpu as integer)) as cpu,
				sum(cast(ram as integer)) as ram,
				sum(cast(disk_storage as integer)) as disk_storage,
				? as sonumber
			from
			(
				SELECT [EquipmentCode],
				(case
					when PropertiesCode = '001' then xValue
					else '0' end
				) as cpu,
				(case
					when PropertiesCode = '002' then xValue
					else '0' end
				) as ram,
				(case
					when PropertiesCode = '003' then xValue
					else '0' end
				) as disk_storage,
				(case
					when PropertiesCode = '004' then xValue
					else '0' end
				) as sonumber
				FROM [ITSM_UK].[dbo].[master_equipment]
				JOIN [ITSM_UK].[dbo].[master_equipment_properties] ON master_equipment_properties.ObjectID = master_equipment.ObjectID
				where master_equipment_properties.ObjectID in
				(
					select ObjectID from master_equipment_properties where xValue =?
				)
			) tb_real GROUP BY EquipmentCode
			 `, so, so).Scan(&vm).Error; err != nil {
		log.Errorln(pkgName, err, "Select vm error")
		return echo.ErrInternalServerError
	}

	dataRaw := map[string]interface{}{
		"vm":      vm,
		"summary": len(vm),
	}
	return c.JSON(http.StatusOK, dataRaw)
}

func GetVmSummaryV2EndPoint(c echo.Context) error {
	if strings.TrimSpace(c.QueryParam("customer")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid customer"})
	}
	Customer := strings.TrimSpace(c.QueryParam("customer"))
	vm := []struct {
		EquipmentCode string `json:"EquipmentCode" gorm:"column:EquipmentCode"`
		Description   string `json:"Description" gorm:"column:Description"`
		EquipmentType string `json:"EquipmentType" gorm:"column:EquipmentType"`
		ObjectID      string `json:"ObjectID" gorm:"column:ObjectID"`
		xValue        string `json:"xValue" gorm:"column:xValue"`
	}{}
	if err := dbEquip.Ctx().Raw(` SELECT master_equipment.EquipmentCode
				,[Description]
				,[EquipmentType]
				,master_equipment.ObjectID
				,master_equipment_properties.xValue
			FROM [ITSM_UK].[dbo].[master_equipment]
			JOIN [ITSM_UK].[dbo].[master_equipment_properties] ON master_equipment_properties.ObjectID = master_equipment.ObjectID
			JOIN [ITSM_UK].[dbo].[master_equipment_owner_assignment] ON master_equipment_owner_assignment.EquipmentCode = master_equipment.EquipmentCode
			WHERE master_equipment.EquipmentType = 'LVMG'
			AND master_equipment_properties.PropertiesCode = '004'
			AND master_equipment_owner_assignment.OwnerCode = ? `, Customer).Scan(&vm).Error; err != nil {
		log.Errorln(pkgName, err, "Select vm error")
		return echo.ErrInternalServerError
	}
	dataRaw := map[string]interface{}{
		"vm":      vm,
		"summary": len(vm),
	}
	return c.JSON(http.StatusOK, dataRaw)
}

func GetSOCustomerEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))
	// status := strings.TrimSpace(c.QueryParam("status"))
	// fmt.Println("====> filter", search)
	ds := time.Now()
	de := time.Now()
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("start_date")), 10); err == nil {
		ds = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("end_date")), 10); err == nil {
		de = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	yearStart, monthStart, dayStart := ds.Date()
	yearEnd, monthEnd, dayEnd := de.Date()
	startRange := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	endRange := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	dateFrom := startRange.Format("2006-01-02")
	dateTo := endRange.Format("2006-01-02")
	m := endRange.Sub(startRange)
	if m < 0 {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid date"})
	}

	//// get staff id ////
	var user []model.UserInfo
	if err := dbSale.Ctx().Raw(`SELECT * FROM user_info WHERE staff_id = ? and role = 'admin';`, saleId).Scan(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
	}
	var listId []string
	if len(user) != 0 {
		var staffAll []model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info ;`).Scan(&staffAll).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
			}
		}
		for _, i := range staffAll {
			listId = append(listId, i.StaffId)
		}
	} else {
		var staffAll model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info WHERE staff_id = ?;`, saleId).Scan(&staffAll).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
			}
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
		if staffAll.StaffChild != "" {
			data := strings.Split(staffAll.StaffChild, ",")
			listId = data
		}
		listId = append(listId, staffAll.StaffId)
	}

	type SOCus struct {
		SOnumber          string `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate string `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate   string `json:"contract_end_date" gorm:"column:ContractEndDate"`
		// PeriodStartDate     string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		// PeriodEndDate       string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
		SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
		SaleName            string  `json:"sale_name" gorm:"column:sale_name"`
		SaleTeam            string  `json:"sale_team" gorm:"column:sale_team"`
		SaleFactor          string  `json:"sale_factor" gorm:"column:sale_factor"`
		InFactor            string  `json:"in_factor" gorm:"column:in_factor"`
		ExFactor            string  `json:"ex_factor" gorm:"column:ex_factor"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SoType              string  `json:"SoType" gorm:"column:SoType"`
		Detail              string  `json:"detail" gorm:"column:detail"`
		Status              string  `json:"status" gorm:"column:status"`
		SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount              float64 `json:"amount" gorm:"column:amount"`
		StaffID             string  `json:"staff_id" gorm:"column:staff_id"`
		Prefix              string  `json:"prefix" gorm:"column:prefix"`
		Fname               string  `json:"fname" gorm:"column:fname"`
		Lname               string  `json:"lname" gorm:"column:lname"`
		Position            string  `json:"position" gorm:"column:position"`
		Department          string  `json:"department" gorm:"column:department"`
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

	hasErr := 0
	var sum []SOCus
	var soTotal []TrackInvoice
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {

		sql := `SELECT sonumber,ContractStartDate,ContractEndDate,pricesale,TotalContractAmount,SOWebStatus,Customer_ID,Customer_Name,sale_code,
		sale_name,sale_name,sale_team,sale_factor,in_factor,ex_factor,so_refer,SoType,detail,staff_id,prefix,fname,lname,position,department,
		(CASE
			WHEN GetCN !='' THEN 'ลดหนี้'
			ELSE 'Success' END
		) as status,
		SUM(CASE
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
		) as so_amount,
		sum(PeriodAmount) as amount
		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					WHERE Active_Inactive = 'Active'
					and PeriodStartDate <= ? and PeriodEndDate >= ?
					and PeriodStartDate <= PeriodEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
		sum(so_amount) as so_amount, sum(inv_amount) as inv_amount, sum(cs_amount) as cs_amount, sum(rc_amount) as rc_amount, sum(cn_amount) as cn_amount, sum(amount) as amount, AVG(in_factor) as in_factor,
		sum(in_factor) as sum_if, sum(inv_amount) - sum(rc_amount) as outstainding_amount,AVG(ex_factor) as ex_factor,sum(ex_factor) as sum_ef,
		(CASE
			WHEN sum(inv_amount) = 0 THEN 'ยังไม่ออกใบแจ้งหนี้'
			WHEN sum(inv_amount) = sum(cn_amount) THEN 'ลดหนี้'
			WHEN sum(inv_amount) - sum(cn_amount) <= sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระแล้ว'
			WHEN sum(inv_amount) - sum(cn_amount) > sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระไม่ครบ'
			ELSE 'ค้างชำระ' END
		) as status,
		sum((CASE
			WHEN inv_amount = rc_amount THEN inv_amount
			ELSE inv_amount - cn_amount END
		)) as inv_amount_cal,
		(sum(amount)/sum(amount_engcost)) as sale_factor,
		sonumber_all
		from (
			SELECT
				count(DISTINCT sonumber) as sonumber,
				count(sonumber) as sonumber_all,
				Customer_ID as Customer_ID,
				Customer_Name as Customer_Name,
				count(DISTINCT(CASE WHEN SDPropertyCS28 !='' THEN SDPropertyCS28 END)) as csnumber,
				count(DISTINCT(CASE WHEN BLSCDocNo !='' THEN BLSCDocNo END)) as invnumber,
				count(DISTINCT(CASE WHEN INCSCDocNo !='' THEN INCSCDocNo END)) as rcnumber,
				count(DISTINCT(CASE WHEN GetCN !='' THEN GetCN END)) as cnnumber,
				sum(so_amount) as so_amount,
				sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) as inv_amount,
				sum(CASE WHEN SDPropertyCS28 !='' THEN so_amount ELSE 0 END) as cs_amount,
				sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as rc_amount,
				sum(CASE WHEN GetCN !='' THEN so_amount ELSE 0 END) as cn_amount,
				sum(PeriodAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
						(CASE
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
						) as so_amount
					FROM (
						SELECT * FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"customer_total": len(sum),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataMap)
}

func GetSOCustomerCsNumberEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	CsNumber := strings.TrimSpace(c.QueryParam("cs_number"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

	ds := time.Now()
	de := time.Now()
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("start_date")), 10); err == nil {
		ds = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("end_date")), 10); err == nil {
		de = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	yearStart, monthStart, dayStart := ds.Date()
	yearEnd, monthEnd, dayEnd := de.Date()
	startRange := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	endRange := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	dateFrom := startRange.Format("2006-01-02")
	dateTo := endRange.Format("2006-01-02")
	m := endRange.Sub(startRange)
	if m < 0 {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid date"})
	}

	//// get staff id ////
	var user []model.UserInfo
	if err := dbSale.Ctx().Raw(`SELECT * FROM user_info WHERE staff_id = ? and role = 'admin';`, saleId).Scan(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
	}
	var listId []string
	if len(user) != 0 {
		var staffAll []model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info ;`).Scan(&staffAll).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
			}
		}
		for _, i := range staffAll {
			listId = append(listId, i.StaffId)
		}
	} else {
		var staffAll model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info WHERE staff_id = ?;`, saleId).Scan(&staffAll).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
			}
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
		if staffAll.StaffChild != "" {
			data := strings.Split(staffAll.StaffChild, ",")
			listId = data
		}
		listId = append(listId, staffAll.StaffId)
	}

	type SOCus struct {
		SOnumber          string `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate string `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate   string `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SDPropertyCS28    string `json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
		// PeriodEndDate       string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
		SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
		SaleName            string  `json:"sale_name" gorm:"column:sale_name"`
		SaleTeam            string  `json:"sale_team" gorm:"column:sale_team"`
		SaleFactor          string  `json:"sale_factor" gorm:"column:sale_factor"`
		InFactor            string  `json:"in_factor" gorm:"column:in_factor"`
		ExFactor            string  `json:"ex_factor" gorm:"column:ex_factor"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SoType              string  `json:"SoType" gorm:"column:SoType"`
		Detail              string  `json:"detail" gorm:"column:detail"`
		SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount              float64 `json:"amount" gorm:"column:amount"`
		StaffID             string  `json:"staff_id" gorm:"column:staff_id"`
		Prefix              string  `json:"prefix" gorm:"column:prefix"`
		Fname               string  `json:"fname" gorm:"column:fname"`
		Lname               string  `json:"lname" gorm:"column:lname"`
		Position            string  `json:"position" gorm:"column:position"`
		Department          string  `json:"department" gorm:"column:department"`
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

	hasErr := 0
	var soTotal []TrackInvoice
	var sum []SOCus
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {

		sql := `SELECT sonumber,SDPropertyCS28,ContractStartDate,ContractEndDate,pricesale,TotalContractAmount,SOWebStatus,Customer_ID,Customer_Name,sale_code,
		sale_name,sale_name,sale_team,sale_factor,in_factor,ex_factor,so_refer,SoType,detail,staff_id,prefix,fname,lname,position,department,
		SUM(CASE
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
		) as so_amount,
		sum(PeriodAmount) as amount
		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					WHERE Active_Inactive = 'Active'
					and PeriodStartDate <= ? and PeriodEndDate >= ?
					and PeriodStartDate <= PeriodEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code), ?)
					and INSTR(CONCAT_WS('|', SDPropertyCS28), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
		sum(so_amount) as so_amount, sum(inv_amount) as inv_amount, sum(cs_amount) as cs_amount, sum(rc_amount) as rc_amount, sum(cn_amount) as cn_amount, sum(amount) as amount, AVG(in_factor) as in_factor,
		sum(in_factor) as sum_if, sum(inv_amount) - sum(rc_amount) as outstainding_amount,AVG(ex_factor) as ex_factor,sum(ex_factor) as sum_ef,
		(CASE
			WHEN sum(inv_amount) = 0 THEN 'ยังไม่ออกใบแจ้งหนี้'
			WHEN sum(inv_amount) = sum(cn_amount) THEN 'ลดหนี้'
			WHEN sum(inv_amount) - sum(cn_amount) <= sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระแล้ว'
			WHEN sum(inv_amount) - sum(cn_amount) > sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระไม่ครบ'
			ELSE 'ค้างชำระ' END
		) as status,
		sum((CASE
			WHEN inv_amount = rc_amount THEN inv_amount
			ELSE inv_amount - cn_amount END
		)) as inv_amount_cal,
		(sum(amount)/sum(amount_engcost)) as sale_factor,
		sonumber_all
		from (
			SELECT
				count(DISTINCT sonumber) as sonumber,
				count(sonumber) as sonumber_all,
				Customer_ID as Customer_ID,
				Customer_Name as Customer_Name,
				count(DISTINCT(CASE WHEN SDPropertyCS28 !='' THEN SDPropertyCS28 END)) as csnumber,
				count(DISTINCT(CASE WHEN BLSCDocNo !='' THEN BLSCDocNo END)) as invnumber,
				count(DISTINCT(CASE WHEN INCSCDocNo !='' THEN INCSCDocNo END)) as rcnumber,
				count(DISTINCT(CASE WHEN GetCN !='' THEN GetCN END)) as cnnumber,
				sum(so_amount) as so_amount,
				sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) as inv_amount,
				sum(CASE WHEN SDPropertyCS28 !='' THEN so_amount ELSE 0 END) as cs_amount,
				sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as rc_amount,
				sum(CASE WHEN GetCN !='' THEN so_amount ELSE 0 END) as cn_amount,
				sum(PeriodAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
						(CASE
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
						) as so_amount
					FROM (
						SELECT * FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', SDPropertyCS28), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"customer_total": len(sum),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataMap)
}

func GetSOCustomerDetailTemplateEndPoint(c echo.Context) error {

	search := strings.TrimSpace(c.QueryParam("search"))

	type SOCus struct {
		StaffID     string `json:"sale_code" gorm:"column:sale_code"`
		CustomerID  string `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string `json:"Customer_Name" gorm:"column:Customer_Name"`
		Nname       string `json:"nname" gorm:"column:nname"`
		Status      string `json:"statusso" gorm:"column:statusso"`
		Department  string `json:"sale_team" gorm:"column:sale_team"`
	}

	// type SOCus struct {
	// 	SOnumber          string `json:"so_number" gorm:"column:sonumber"`
	// 	ContractStartDate string `json:"contract_start_date" gorm:"column:ContractStartDate"`
	// 	ContractEndDate   string `json:"contract_end_date" gorm:"column:ContractEndDate"`
	// 	SDPropertyCS28    string `json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
	// 	// PeriodEndDate       string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
	// 	PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
	// 	TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
	// 	SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
	// 	CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
	// 	CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
	// 	SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
	// 	SaleName            string  `json:"sale_name" gorm:"column:sale_name"`
	// 	SaleTeam            string  `json:"sale_team" gorm:"column:sale_team"`
	// 	SaleFactor          string  `json:"sale_factor" gorm:"column:sale_factor"`
	// 	InFactor            string  `json:"in_factor" gorm:"column:in_factor"`
	// 	ExFactor            string  `json:"ex_factor" gorm:"column:ex_factor"`
	// 	SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
	// 	SoType              string  `json:"SoType" gorm:"column:SoType"`
	// 	Detail              string  `json:"detail" gorm:"column:detail"`
	// 	SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
	// 	Amount              float64 `json:"amount" gorm:"column:amount"`
	// 	StaffID             string  `json:"staff_id" gorm:"column:staff_id"`
	// 	Prefix              string  `json:"prefix" gorm:"column:prefix"`
	// 	Fname               string  `json:"fname" gorm:"column:fname"`
	// 	Lname               string  `json:"lname" gorm:"column:lname"`
	// 	Position            string  `json:"position" gorm:"column:position"`
	// 	Department          string  `json:"department" gorm:"column:department"`
	// }

	var sum []SOCus

	sql := `SELECT Customer_ID,Customer_Name,sale_code,nname,sale_team,SUM(CASE
		WHEN ContractEndDate < CURDATE()
		THEN 1
		ELSE 0 END
		) as statusso FROM 
		(SELECT id,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,pricesale,TotalContractAmount
						,SOWebStatus,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN
						,INCSCDocNo,Customer_ID,Customer_Name,sale_code,sale_name
						,sale_team,sale_lead,PeriodAmount,sale_factor,in_factor
						,ex_factor,so_refer,Active_Inactive,so_renew,create_date
						,status_so,status_sale,has_refer,SOType,detail,remark
		FROM so_mssql 
		 GROUP by sonumber
		) aa
		LEFT JOIN (SELECT staff_id,nname,lname,fname,prefix,one_id,position FROM staff_info) as staff_info ON aa.sale_code = staff_info.staff_id
		 WHERE INSTR(CONCAT_WS('|', id,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,pricesale,TotalContractAmount
						,SOWebStatus,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN
						,INCSCDocNo,Customer_ID,Customer_Name,sale_code,sale_name
						,sale_team,sale_lead,PeriodAmount,sale_factor,in_factor
						,ex_factor,so_refer,Active_Inactive,so_renew,create_date
						,status_so,status_sale,has_refer,SOType,detail,remark
					   ), ?)
		GROUP by Customer_ID LIMIT 5;`

	if err := dbSale.Ctx().Raw(sql, search).Scan(&sum).Error; err != nil {
		log.Errorln(pkgName, err, "select data error -:")
		// hasErr += 1
	}

	// 	sql := `
	// 	SELECT id,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,pricesale,TotalContractAmount
	// 									,SOWebStatus,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN
	// 									,INCSCDocNo,Customer_ID,Customer_Name,sale_code,sale_name
	// 									,sale_team,sale_lead,PeriodAmount,sale_factor,in_factor
	// 									,ex_factor,so_refer,Active_Inactive,so_renew,create_date
	// 									,status_so,status_sale,has_refer,SOType,detail,remark
	// 									,nname,lname,fname,prefix,one_id ,position
	// 					FROM so_mssql
	// 					LEFT JOIN (SELECT staff_id,nname,lname,fname,prefix,one_id,position FROM staff_info) as staff_info ON so_mssql.sale_code = staff_info.staff_id
	// 					 GROUP by sonumber;`

	// if err := dbSale.Ctx().Raw(sql, search).Scan(&sum).Error; err != nil {
	// 	log.Errorln(pkgName, err, "select data error -:")
	// 	// hasErr += 1
	// }

	for _, v := range sum {
		if v.Status != "0" {
			v.Status = "ลูกค้าที่ซื้อขายปัจจุบัน"
		} else {
			v.Status = "ลูกค้าที่หมดสัญญาไปแล้ว"
		}
		NewStrData := fmt.Sprintf("ลูกค้า: %s \nชื่อลูกค้า: %s \nทีมที่ดูแล: %s \nพนักงานที่ดูแล: %s - %s\nสเตตัส: %s \n ", v.CustomerID, v.CusnameThai, v.Department, v.StaffID, v.Nname, v.Status)

		// To := 25078584384
		url := "https://chat-api.one.th/message/api/v1/push_message"

		payload, _ := json.Marshal(&struct {
			To                 string `json:"to"`
			BotId              string `json:"bot_id"`
			Type               string `json:"type"`
			Message            string `json:"message"`
			CustomNotification string `json:"custom_notification"`
		}{
			// To: "198008320896",
			To: "25078584384",
			// To:                 OneId,
			// BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
			BotId:              "Becf3d73c867f508ab7a8f5d62ceceb64",
			Type:               "text",
			Message:            NewStrData,
			CustomNotification: "เปิดอ่านข้อความใหม่จากทางเรา",
		})

		headers := map[string]string{
			"Authorization": "Bearer A548a4dd47e3c5108affe99b48b5c0218db9bcaaca6b34470b389bd04a19c3e30e1b99dad38844be387e939f755d194be",
			// "Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",

			"Content-Type": "application/json",
		}
		_, err := requests.Post(url, headers, bytes.NewBuffer(payload), 50)
		if err != nil {
			log.Errorln("Error QuickReply", err)
			return err

		}

	}

	return nil
}
