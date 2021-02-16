package api

import (
	"net/http"
	"sale_ranking/model"
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

func GetTrackingInvoiceEndPoint(c echo.Context) error {

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	var so []m.TrackInvoice
	var nonSo []m.TrackInvoice
	var totalSo []m.TrackInvoice
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
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
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
					
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom).Scan(&so).Error; err != nil {
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
						WHERE Active_Inactive = 'Active' and BLSCDocNo = ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom).Scan(&nonSo).Error; err != nil {
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
						WHERE Active_Inactive = 'Active'
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom).Scan(&totalSo).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	dataInv := map[string]interface{}{
		"total_amount": so[0].SoAmount,
		"count_so":     so[0].SoNumberAll,
		"detail":       so[0],
	}
	dataNonInv := map[string]interface{}{
		"total_amount": nonSo[0].SoAmount,
		"count_so":     nonSo[0].SoNumberAll,
		"detail":       nonSo[0],
	}
	dataTotal := map[string]interface{}{
		"total_amount": nonSo[0].SoAmount + so[0].SoAmount,
		"count_so":     nonSo[0].SoNumberAll + so[0].SoNumberAll,
		"detail":       totalSo[0],
	}

	dataRaw := map[string]interface{}{
		"inv":     dataInv,
		"non_inv": dataNonInv,
		"total":   dataTotal,
	}

	return c.JSON(http.StatusOK, dataRaw)
}

func GetDocTrackingInvoiceEndPoint(c echo.Context) error {

	type SOCus struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SDPropertyCS28      string  `json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID          string  `json:"customer_id" gorm:"column:Customer_ID"`
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

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	var so []SOCus
	var soTotal []m.TrackInvoice
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sql := `		SELECT *,
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
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by sonumber
						;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId).Scan(&so).Error; err != nil {
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

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	dataInv := map[string]interface{}{
		"count_so": len(so),
		"total_so": soTotal,
		"detail":   so,
	}

	return c.JSON(http.StatusOK, dataInv)
}

func GetDocTrackingInvoiceSOEndPoint(c echo.Context) error {

	type SOCus struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SDPropertyCS28      string  `json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID          string  `json:"customer_id" gorm:"column:Customer_ID"`
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

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	SoNumber := strings.TrimSpace(c.QueryParam("so_number"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	var so []SOCus
	var soTotal []m.TrackInvoice
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sql := `		SELECT *,
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
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sonumber), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by sonumber
						;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SoNumber, StaffId).Scan(&so).Error; err != nil {
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
						and INSTR(CONCAT_WS('|', sonumber), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SoNumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	dataInv := map[string]interface{}{
		"count_so": len(so),
		"total_so": soTotal,
		"detail":   so,
	}

	return c.JSON(http.StatusOK, dataInv)
}

func GetTrackingBillingEndPoint(c echo.Context) error {
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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	so := []struct {
		InvNo string `json:"invoice_no" gorm:"column:invoice_no"`
	}{}

	sql := `
			 SELECT
			 invoice_no
				FROM (
					SELECT
					invoice_no,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
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
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom).Scan(&so).Error; err != nil {
		log.Errorln(pkgName, err, "select data error -:")
		return echo.ErrInternalServerError
	}
	listInv := []string{}
	for _, v := range so {
		listInv = append(listInv, v.InvNo)
	}
	var inv []m.Invoice
	if err := dbSale.Ctx().Model(&m.Invoice{}).Where(`invoice_no IN (?)`, listInv).Preload("InvStatus").Find(&inv).Error; err != nil {
		log.Errorln(pkgName, err, "select data invoice error -:")

	}
	hasBill := 0
	hadBill := 0
	notBill := 0

	listBill := []string{}
	listNonBill := []string{}
	listNotBill := []string{}
	for _, v := range inv {
		if len(v.InvStatus) != 0 {
			check := false
			for _, i := range v.InvStatus {
				if strings.TrimSpace(i.InvoiceStatusName) == "วางบิลแล้ว" {
					check = true
				} else {
					if !check {
						check = false
					}
				}
			}
			if check {
				hadBill += 1
				listBill = append(listBill, v.InvoiceNo)
			} else {
				hasBill += 1
				listNonBill = append(listNonBill, v.InvoiceNo)
			}
		} else {
			notBill += 1
			listNotBill = append(listNotBill, v.InvoiceNo)
		}
	}

	var billing []m.TrackInvoice
	var nonBilling []m.TrackInvoice
	var notBilling []m.TrackInvoice

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		sqlSum := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
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
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listBill).Scan(&billing).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}
		wg.Done()
	}()
	go func() {
		sqlSum := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
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
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listNonBill).Scan(&nonBilling).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}
		wg.Done()
	}()
	go func() {
		sqlSum := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
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
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listNotBill).Scan(&notBilling).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}
		wg.Done()
	}()
	wg.Wait()

	dataBill := map[string]interface{}{
		"total":  hadBill,
		"detail": billing[0],
	}
	dataNonBill := map[string]interface{}{
		"total":  hasBill,
		"detail": nonBilling[0],
	}
	dataNotBill := map[string]interface{}{
		"total":  notBill,
		"detail": notBilling[0],
	}
	dataInv := map[string]interface{}{
		"billing":     dataBill,
		"non_billing": dataNonBill,
	}
	dataRaw := map[string]interface{}{
		"bill":       dataInv,
		"not_bill":   dataNotBill,
		"date_start": dateFrom.Format("02-Jan-2006"),
		"date_end":   dateTo.Format("02-Jan-2006"),
	}
	return c.JSON(http.StatusOK, dataRaw)
}

func GetTrackingBillingStatusEndPoint(c echo.Context) error {

	// SELECT * FROM `so_mssql` JOIN invoice_status ON so_mssql.BLSCDocNo = invoice_status.inv_no;

	type SOCusBill struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SDPropertyCS28      string  `json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
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
		InvStatusName       string  `json:"invoice_status_name" gorm:"column:invoice_status_name"`
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

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

	// var sum []SOCus
	var soTotal []TrackInvoice
	var so []SOCusBill
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sql := `		SELECT *,
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
						sum(PeriodAmount) as amount,
						invoice_status_name
					 FROM invoice
					 LEFT JOIN invoice_status ON invoice.invoice_no = invoice_status.inv_no
					 LEFT JOIN so_mssql ON so_mssql.BLSCDocNo = invoice_status.inv_no
					 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					 WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate
							and sale_code in (?)
							and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							group by BLSCDocNo
							;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId).Scan(&so).Error; err != nil {
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
						SELECT SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
						FROM invoice
						LEFT JOIN invoice_status ON invoice.invoice_no = invoice_status.inv_no
						LEFT JOIN so_mssql ON so_mssql.BLSCDocNo = invoice_status.inv_no
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	dataInv := map[string]interface{}{
		"count_so": len(so),
		"total_so": soTotal,
		"detail":   so,
	}

	return c.JSON(http.StatusOK, dataInv)
}

func GetTrackingBillingStatusInvEndPoint(c echo.Context) error {

	// SELECT * FROM `so_mssql` JOIN invoice_status ON so_mssql.BLSCDocNo = invoice_status.inv_no;

	type SOCusBill struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SDPropertyCS28      string  `json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
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
		InvStatusName       string  `json:"invoice_status_name" gorm:"column:invoice_status_name"`
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

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))
	InvNumber := strings.TrimSpace(c.QueryParam("inv_number"))

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)

	// var sum []SOCus
	var soTotal []TrackInvoice
	var so []SOCusBill
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sql := `		SELECT *,
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
						sum(PeriodAmount) as amount,
						invoice_status_name
					 FROM invoice
					 LEFT JOIN invoice_status ON invoice.invoice_no = invoice_status.inv_no
					 LEFT JOIN so_mssql ON so_mssql.BLSCDocNo = invoice_status.inv_no
					 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					 WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate
							and sale_code in (?)
							and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							and INSTR(CONCAT_WS('|', invoice_no), ?)
							group by BLSCDocNo
							;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId, InvNumber).Scan(&so).Error; err != nil {
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
						SELECT SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
						FROM invoice
						LEFT JOIN invoice_status ON invoice.invoice_no = invoice_status.inv_no
						LEFT JOIN so_mssql ON so_mssql.BLSCDocNo = invoice_status.inv_no
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						and INSTR(CONCAT_WS('|', invoice_no), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId, InvNumber).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	dataInv := map[string]interface{}{
		"count_so": len(so),
		"total_so": soTotal,
		"detail":   so,
	}

	return c.JSON(http.StatusOK, dataInv)
}

func GetTrackingReceiptEndPoint(c echo.Context) error {
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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	so := []struct {
		InvNo string `json:"invoice_no" gorm:"column:invoice_no"`
	}{}

	sql := `
			 SELECT
			 invoice_no
				FROM (
					SELECT
					invoice_no,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
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
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom).Scan(&so).Error; err != nil {
		log.Errorln(pkgName, err, "select data error -:")
		return echo.ErrInternalServerError
	}

	listInv := []string{}
	for _, v := range so {
		listInv = append(listInv, v.InvNo)
	}
	var inv []m.Invoice
	if err := dbSale.Ctx().Model(&m.Invoice{}).Where(`invoice_no IN (?)`, listInv).Preload("InvStatus").Find(&inv).Error; err != nil {
		log.Errorln(pkgName, err, "select data invoice error -:")
	}
	hasBill := 0
	hadBill := 0
	nonBill := 0
	var listInvBilling []string
	for _, v := range inv {
		if len(v.InvStatus) != 0 {
			check := false
			for _, i := range v.InvStatus {
				if strings.TrimSpace(i.InvoiceStatusName) == "วางบิลแล้ว" {
					check = true
				} else {
					if !check {
						check = false
					}
				}
			}
			if check {
				listInvBilling = append(listInvBilling, v.InvoiceNo)
				hadBill += 1
			} else {
				hasBill += 1
			}
		} else {
			nonBill += 1
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	var sumSo []m.TrackInvoice
	var sumNonSo []m.TrackInvoice
	var receipt []m.SOMssql
	var nonReceipt []m.SOMssql
	go func() {
		if err := dbSale.Ctx().Raw(`select BLSCDocNo from so_mssql where BLSCDocNo IN (?) and INCSCDocNo <> '' group by BLSCDocNo`, listInvBilling).Scan(&receipt).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}
		sqlSum := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
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
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) and INCSCDocNo <> ''
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv).Scan(&sumSo).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(`select BLSCDocNo from so_mssql where BLSCDocNo IN (?) and INCSCDocNo = '' group by BLSCDocNo`, listInvBilling).Scan(&nonReceipt).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}
		sqlSum := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
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
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) and INCSCDocNo = ''
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`
		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv).Scan(&sumNonSo).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}
		wg.Done()
	}()
	wg.Wait()
	dataReceipt := map[string]interface{}{
		"has_receipt": len(receipt),
		"detail":      sumSo[0],
	}
	dataNonReceipt := map[string]interface{}{
		"non_receipt": len(nonReceipt),
		"detail":      sumNonSo[0],
	}

	dataRaw := map[string]interface{}{
		"receipt":       dataReceipt,
		"non_receipt":   dataNonReceipt,
		"total_billing": hadBill,
		"date_start":    dateFrom.Format("02-Jan-2006"),
		"date_end":      dateTo.Format("02-Jan-2006"),
	}

	return c.JSON(http.StatusOK, dataRaw)
}

func GetSOTrackingReceiptEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	so := []struct {
		InvNo string `json:"invoice_no" gorm:"column:invoice_no"`
	}{}

	sql := `
			 SELECT
			 invoice_no
				FROM (
					SELECT
					invoice_no,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
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
						) as so_amount
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom).Scan(&so).Error; err != nil {
		log.Errorln(pkgName, err, "select data error -:")
		return echo.ErrInternalServerError
	}

	listInv := []string{}
	for _, v := range so {
		listInv = append(listInv, v.InvNo)
	}
	var inv []m.Invoice
	if err := dbSale.Ctx().Model(&m.Invoice{}).Where(`invoice_no IN (?)`, listInv).Preload("InvStatus").Find(&inv).Error; err != nil {
		log.Errorln(pkgName, err, "select data invoice error -:")
	}
	hasBill := 0
	hadBill := 0
	nonBill := 0
	var listInvBilling []string
	for _, v := range inv {
		if len(v.InvStatus) != 0 {
			check := false
			for _, i := range v.InvStatus {
				if strings.TrimSpace(i.InvoiceStatusName) == "วางบิลแล้ว" {
					check = true
				} else {
					if !check {
						check = false
					}
				}
			}
			if check {
				listInvBilling = append(listInvBilling, v.InvoiceNo)
				hadBill += 1
			} else {
				hasBill += 1
			}
		} else {
			nonBill += 1
		}
	}
	// var sumSo []m.TrackInvoice
	// var receipt []m.SOMssql
	// if err := dbSale.Ctx().Raw(`select BLSCDocNo from so_mssql where BLSCDocNo IN (?) and INCSCDocNo <> '' group by BLSCDocNo`, listInvBilling).Scan(&receipt).Error; err != nil {
	// 	log.Errorln(pkgName, err, "select data error -:")
	// }

	hasErr := 0
	var sum []SOCus
	var soTotal []TrackInvoice
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {

		sqlSum := `		SELECT *,
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
					) as so_amount,
					sum(PeriodAmount) as amount
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) and INCSCDocNo <> ''
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
																	;`
		var sum []SOCus
		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
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
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) and INCSCDocNo <> ''
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	dataReceipt := map[string]interface{}{
		"has_receipt": len(sum),
		"total_so":    soTotal,
		"detail":      sum,
	}
	return c.JSON(http.StatusOK, dataReceipt)
}

func GetSOTrackingReceiptCsEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	InvNumber := strings.TrimSpace(c.QueryParam("inv_number"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	so := []struct {
		InvNo string `json:"invoice_no" gorm:"column:invoice_no"`
	}{}

	sql := `
			 SELECT
			 invoice_no
				FROM (
					SELECT
					invoice_no,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
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
						) as so_amount
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom).Scan(&so).Error; err != nil {
		log.Errorln(pkgName, err, "select data error -:")
		return echo.ErrInternalServerError
	}

	listInv := []string{}
	for _, v := range so {
		listInv = append(listInv, v.InvNo)
	}
	var inv []m.Invoice
	if err := dbSale.Ctx().Model(&m.Invoice{}).Where(`invoice_no IN (?)`, listInv).Preload("InvStatus").Find(&inv).Error; err != nil {
		log.Errorln(pkgName, err, "select data invoice error -:")
	}
	hasBill := 0
	hadBill := 0
	nonBill := 0
	var listInvBilling []string
	for _, v := range inv {
		if len(v.InvStatus) != 0 {
			check := false
			for _, i := range v.InvStatus {
				if strings.TrimSpace(i.InvoiceStatusName) == "วางบิลแล้ว" {
					check = true
				} else {
					if !check {
						check = false
					}
				}
			}
			if check {
				listInvBilling = append(listInvBilling, v.InvoiceNo)
				hadBill += 1
			} else {
				hasBill += 1
			}
		} else {
			nonBill += 1
		}
	}

	hasErr := 0
	var sum []SOCus
	var soTotal []TrackInvoice
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {

		sqlSum := `		SELECT *,
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
					) as so_amount,
					sum(PeriodAmount) as amount
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) and INCSCDocNo <> ''
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code), ?)
						and INSTR(CONCAT_WS('|', invoice_no), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
					;`
		var sum []SOCus
		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, InvNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
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
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) and INCSCDocNo <> ''

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						and INSTR(CONCAT_WS('|', invoice_no), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId, InvNumber).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	dataReceipt := map[string]interface{}{
		"has_receipt": len(sum),
		"total_so":    soTotal,
		"detail":      sum,
	}
	return c.JSON(http.StatusOK, dataReceipt)
}

func GetSOTrackingCsEndPoint(c echo.Context) error {

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
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"sale_factor" gorm:"column:sale_factor"`
	}

	hasErr := 0
	var soTotal []TrackInvoice
	var sum []SOCus
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sql := `SELECT *,
		SUM(CASE
			WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
			THEN 0
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
			THEN Revenue_Month
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?, StartDate_P1)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(EndDate_P1, ?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 = ?
			THEN 1*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?,?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
			ELSE 0 END
		) as so_amount,
		sum(Revenue_Month) as amount
		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', Customer_ID, Sales_Name, EmployeeID), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount, AVG(in_factor) as in_factor,AVG(ex_factor) as ex_factor,
		(sum(amount)/sum(amount_engcost)) as SaleFactors
		from (
			SELECT
				Customer_ID as Customer_ID,
				Cusname_thai as Cusname_thai,
				sum(Revenue_Month) as amount,
				sum(eng_cost) as amount_engcost,
				SaleFactors,
				in_factor,EmployeeID,Sales_Name,ex_factor
				FROM (
					SELECT
					doc_number_eform,StartDate_P1,EndDate_P1,Customer_ID,Cusname_thai,
					EmployeeID,	Sales_Name,Sale_Team,Revenue_Month, SaleFactors, Int_INET as in_factor, (Ext_JV + Ext) as ex_factor,
						(case
							when Revenue_Month is not null and SaleFactors is not null then Revenue_Month/SaleFactors
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
							THEN 0
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
							THEN Revenue_Month
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?, StartDate_P1)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(EndDate_P1, ?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 = ?
							THEN 1*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?,?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
							ELSE 0 END
						) as so_amount
					FROM (
						SELECT * FROM costsheet_info
						LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
						WHERE doc_number_eform <> ''
						and StartDate_P1 <= ? and EndDate_P1 >= ?
						and StartDate_P1 <= EndDate_P1

						and EmployeeID in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Sales_Name, EmployeeID,doc_number_eform), ?)
						and INSTR(CONCAT_WS('|', doc_number_eform), ?)
						and INSTR(CONCAT_WS('|', EmployeeID), ?)
					) sub_data
				) so_group
				group by doc_number_eform
			) cust_group`
		// and StartDate_P1 <= ? and EndDate_P1 >= ?
		// and StartDate_P1 <= EndDate_P1
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

func GetDetailCostsheetEndPoint(c echo.Context) error {

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
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"sale_factor" gorm:"column:sale_factor"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
	}

	hasErr := 0
	var soTotal []TrackInvoice
	var sum []SOCus
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sql := `SELECT *,
		SUM(CASE
			WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
			THEN 0
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
			THEN Revenue_Month
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?, StartDate_P1)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(EndDate_P1, ?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 = ?
			THEN 1*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?,?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
			ELSE 0 END
		) as so_amount,
		sum(Revenue_Month) as amount
		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', Customer_ID, Sales_Name, EmployeeID), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount, AVG(in_factor) as in_factor,AVG(ex_factor) as ex_factor,
		(sum(amount)/sum(amount_engcost)) as SaleFactors , SUM(Total_Revenue_Month/(DATEDIFF(?,?)+1)) as pro_rate
		from (
			SELECT
				Customer_ID as Customer_ID,
				Cusname_thai as Cusname_thai,
				sum(Revenue_Month) as amount,
				sum(eng_cost) as amount_engcost,
				SaleFactors,
				in_factor,EmployeeID,Sales_Name,ex_factor,Total_Revenue_Month
				FROM (
					SELECT
					doc_number_eform,StartDate_P1,EndDate_P1,Customer_ID,Cusname_thai,
					EmployeeID,	Sales_Name,Sale_Team,Revenue_Month, SaleFactors, Int_INET as in_factor, (Ext_JV + Ext) as ex_factor,Total_Revenue_Month,
						(case
							when Revenue_Month is not null and SaleFactors is not null then Revenue_Month/SaleFactors
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
							THEN 0
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
							THEN Revenue_Month
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?, StartDate_P1)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(EndDate_P1, ?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 = ?
							THEN 1*(Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?,?)+1)*(Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
							ELSE 0 END
						) as so_amount
					FROM (
						SELECT * FROM costsheet_info
						LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
						WHERE doc_number_eform <> ''
						and StartDate_P1 <= ? and EndDate_P1 >= ?
						and StartDate_P1 <= EndDate_P1

						and EmployeeID in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Sales_Name, EmployeeID,doc_number_eform), ?)
						and INSTR(CONCAT_WS('|', doc_number_eform), ?)
						and INSTR(CONCAT_WS('|', EmployeeID), ?)
					) sub_data
				) so_group
				group by doc_number_eform
			) cust_group`
		// and StartDate_P1 <= ? and EndDate_P1 >= ?
		// and StartDate_P1 <= EndDate_P1
		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&soTotal).Error; err != nil {
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

func GetDetailSoEndPoint(c echo.Context) error {

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
		ProRate           float64 `json:"pro_rate" gorm:"column:pro_rate"`
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
		sum(in_factor) as sum_if, sum(inv_amount) - sum(rc_amount) as outstainding_amount,AVG(ex_factor) as ex_factor,sum(ex_factor) as sum_ef, SUM(PeriodAmount/(DATEDIFF(?,?)+1)) as pro_rate,
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
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount
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

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&soTotal).Error; err != nil {
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

func GetDetailInvoiceEndPoint(c echo.Context) error {

	type SOCus struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SDPropertyCS28      string  `json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID          string  `json:"customer_id" gorm:"column:Customer_ID"`
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
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"sale_factor" gorm:"column:sale_factor"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
	}

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	SoNumber := strings.TrimSpace(c.QueryParam("so_number"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	var so []SOCus
	var soTotal []TrackInvoice
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		sql := `		SELECT *,
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
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code,BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sonumber), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by sonumber
						;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SoNumber, StaffId).Scan(&so).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(sonumber_all) as sonumber_all, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
		sum(so_amount) as so_amount, sum(inv_amount) as inv_amount, sum(cs_amount) as cs_amount, sum(rc_amount) as rc_amount, sum(cn_amount) as cn_amount, sum(amount) as amount, AVG(in_factor) as in_factor,
		sum(in_factor) as sum_if, sum(inv_amount) - sum(rc_amount) as outstainding_amount,AVG(ex_factor) as ex_factor,sum(ex_factor) as sum_ef, SUM(PeriodAmount/(DATEDIFF(?,?)+1)) as pro_rate,
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
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount
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
						and INSTR(CONCAT_WS('|', sonumber), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SoNumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	dataInv := map[string]interface{}{
		"count_so": len(so),
		"total_so": soTotal,
		"detail":   so,
	}

	return c.JSON(http.StatusOK, dataInv)
}
