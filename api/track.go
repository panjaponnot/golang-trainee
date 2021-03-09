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
		SaleFactor          float64 `json:"sale_factor" gorm:"column:sale_factor"`
		InFactor            float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor            float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SoType              string  `json:"SoType" gorm:"column:SoType"`
		Detail              string  `json:"detail" gorm:"column:detail"`
		SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount              float64 `json:"amount" gorm:"column:amount"`
		InvStatusName       string  `json:"status" gorm:"column:status"`
		Reason              string  `json:"reason" gorm:"column:reason"`
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
						status
						FROM billing_info
						JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
						JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
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
						FROM billing_info
					  JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
					  JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
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
		SaleFactor          float64 `json:"sale_factor" gorm:"column:sale_factor"`
		InFactor            float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor            float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SoType              string  `json:"SoType" gorm:"column:SoType"`
		Detail              string  `json:"detail" gorm:"column:detail"`
		SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount              float64 `json:"amount" gorm:"column:amount"`
		InvStatusName       string  `json:"status" gorm:"column:status"`
		Reason              string  `json:"reason" gorm:"column:reason"`
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
						status
						FROM billing_info
						JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
						JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
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
						FROM billing_info
						JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
						JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
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
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
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
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId).Scan(&so).Error; err != nil {
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
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId).Scan(&so).Error; err != nil {
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
		DocNumberEform string  `json:"doc_number_eform" gorm:"column:doc_number_eform"`
		StaffID        string  `json:"staff_id" gorm:"column:staff_id"`
		Fname          string  `json:"fname" gorm:"column:fname"`
		Lname          string  `json:"lname" gorm:"column:lname"`
		Nname          string  `json:"nname" gorm:"column:nname"`
		Department     string  `json:"department" gorm:"column:department"`
		SoAmount       float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount         float64 `json:"amount" gorm:"column:amount"`
		StatusEform    string  `json:"status_eform" gorm:"column:status_eform"`
		CustomerID     string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai    string  `json:"Cusname_thai" gorm:"column:Cusname_thai"`
		CusnameEng     string  `json:"Cusname_Eng" gorm:"column:Cusname_Eng"`
		BusinessType   string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus      string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType         string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor     float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor       float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor       float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate      string  `json:"StartDate_P1" gorm:"column:StartDate_P1"`
		EndDate        string  `json:"EndDate_P1" gorm:"column:EndDate_P1"`
	}
	type TrackInvoice struct {
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
	}

	hasErr := 0
	var soTotal []TrackInvoice
	var sum []SOCus
	cus := []struct {
		CusnameThai string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
	status := struct {
		CompleteFromPaperless  string `json:"Complete_from_paperless" gorm:"column:Complete_from_paperless"`
		CompleteFromEform      string `json:"Complete_from_eform" gorm:"column:Complete_from_eform"`
		OnprocessFromEform     string `json:"Onprocess_from_eform" gorm:"column:Onprocess_from_eform"`
		RejectFromPaperless    string `json:"Reject_from_paperless" gorm:"column:Reject_from_paperless"`
		RejectFromEform        string `json:"Reject_from_eform" gorm:"column:Reject_from_eform"`
		CancelFromEform        string `json:"Cancel_from_eform" gorm:"column:Cancel_from_eform"`
		OnprocessFromPaperless string `json:"Onprocess_from_paperless" gorm:"column:Onprocess_from_paperless"`
	}{}
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		sql := `SELECT *,
		SUM(CASE
			WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
			THEN 0
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
			THEN Total_Revenue_Month
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?, StartDate_P1)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(EndDate_P1, ?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 = ?
			THEN 1*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?,?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
			ELSE 0 END
		) as so_amount,
		sum(Total_Revenue_Month) as amount,
		sum(COALESCE(Int_INET, 0))/100 as in_factor,
		sum((COALESCE(Ext_JV, 0) + COALESCE(Ext, 0)))/100 as ex_factor
		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
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
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor)/100 as in_factor,
		AVG(ex_factor)/100 as ex_factor,
		(sum(amount)/sum(amount_engcost)) as SaleFactors ,
		SUM(so_amount) as pro_rate
		from (
			SELECT
				Customer_ID as Customer_ID,
				Cusname_thai as Cusname_thai,
				sum(Total_Revenue_Month) as amount,
				sum(eng_cost) as amount_engcost,
				SaleFactors,
				sum(in_factor) as in_factor,
				EmployeeID,
				Sales_Name,
				sum(ex_factor) as ex_factor,
				Total_Revenue_Month,
				so_amount
				FROM (
					SELECT
					doc_number_eform,StartDate_P1,EndDate_P1,Customer_ID,Cusname_thai,
					EmployeeID,	Sales_Name,Sale_Team,Total_Revenue_Month, SaleFactors,
					COALESCE(Int_INET, 0) as in_factor,
					(COALESCE(Ext_JV, 0) + COALESCE(Ext, 0)) as ex_factor,
						(case
							when Total_Revenue_Month is not null and SaleFactors is not null then Total_Revenue_Month/SaleFactors
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
							THEN 0
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
							THEN Total_Revenue_Month
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?, StartDate_P1)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(EndDate_P1, ?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 = ?
							THEN 1*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?,?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
							ELSE 0 END
						) as so_amount
					FROM (
						SELECT * FROM costsheet_info
						LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
						WHERE doc_number_eform <> ''
						and StartDate_P1 <= ? and EndDate_P1 >= ?
						and StartDate_P1 <= EndDate_P1

						and EmployeeID in (?)
						and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
						and INSTR(CONCAT_WS('|', doc_number_eform), ?)
						and INSTR(CONCAT_WS('|', EmployeeID), ?)
					) sub_data
				) so_group
				group by doc_number_eform
			) cust_group
			`
		// and StartDate_P1 <= ? and EndDate_P1 >= ?
		// and StartDate_P1 <= EndDate_P1
		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `SELECT distinct Customer_ID
		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `
		SELECT SUM(Complete_from_paperless) as Complete_from_paperless,
SUM(Complete_from_eform) as Complete_from_eform,
SUM(Onprocess_from_eform) as Onprocess_from_eform,
SUM(Reject_from_paperless) as Reject_from_paperless,
SUM(Reject_from_eform) as Reject_from_eform,
SUM(Cancel_from_eform) as Cancel_from_eform,
SUM(Onprocess_from_paperless) as Onprocess_from_paperless
FROM(
		SELECT SUM(CASE
			WHEN status_eform = 'Complete from paperless' THEN 1
		END) Complete_from_paperless,
		SUM(CASE
			WHEN status_eform = 'Complete from eform' THEN 1
		END) Complete_from_eform,
			SUM(CASE
			WHEN status_eform = 'Onprocess from eform' THEN 1
		END) Onprocess_from_eform,
		SUM(CASE
			WHEN status_eform = 'Reject from paperless' THEN 1
		END) Reject_from_paperless,
		 SUM(CASE
			WHEN status_eform = 'Reject from eform' THEN 1
		END) Reject_from_eform,
			 SUM(CASE
			WHEN status_eform = 'Cancel from eform' THEN 1
		END) Cancel_from_eform,
		SUM(CASE
			WHEN status_eform = 'Onprocess from paperless' THEN 1
		END) Onprocess_from_paperless
		FROM (
			SELECT status_eform
			FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					) as ss
					group by status_eform
					) as aa
					;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&status).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"total":          len(sum),
		"customer_total": len(cus),
		"total_so":       soTotal,
		"detail":         sum,
		// "CompleteFromPaperless":  status.CompleteFromPaperless,
		// "CompleteFromEform":      status.CompleteFromEform,
		// "OnprocessFromEform":     status.OnprocessFromEform,
		// "RejectFromPaperless":    status.RejectFromPaperless,
		// "RejectFromEform":        status.RejectFromEform,
		// "CancelFromEform":        status.CancelFromEform,
		// "OnprocessFromPaperless": status.OnprocessFromPaperless,
		"status_eform": status,
	}
	return c.JSON(http.StatusOK, dataMap)
}

func GetDetailSoEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	SONumber := strings.TrimSpace(c.QueryParam("cs_number"))
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
		SOnumber    string  `json:"sonumber" gorm:"column:sonumber"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"ContractStartDate" gorm:"column:ContractStartDate"`
		EndDate      string  `json:"ContractEndDate" gorm:"column:ContractEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
		Remark       string  `json:"remark" gorm:"column:remark"`
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
	cus := []struct {
		CustomerID string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {

		sql := `SELECT *,
		SUM(CASE
			WHEN DATEDIFF(ContractEndDate, ContractStartDate)+1 = 0
			THEN 0
			WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate <= ?
			THEN PeriodAmount
			WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate > ?
			THEN (DATEDIFF(?, ContractStartDate)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate <= ? AND ContractEndDate > ?
			THEN (DATEDIFF(ContractEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate = ?
			THEN 1*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate > ?
			THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate,ContractStartDate)+1))
			ELSE 0 END
		) as so_amount,
		in_factor as in_factor,
		ex_factor as ex_factor,
		sum(PeriodAmount) as amount
		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					WHERE Active_Inactive = 'Active'
					and ContractStartDate <= ? and ContractEndDate >= ?
					and ContractStartDate <= ContractEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
					and INSTR(CONCAT_WS('|', sonumber), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(ContractEndDate, ContractStartDate)+1 = 0
							THEN 0
							WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate <= ?
							THEN PeriodAmount
							WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate > ?
							THEN (DATEDIFF(?, ContractStartDate)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate <= ? AND ContractEndDate > ?
							THEN (DATEDIFF(ContractEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate = ?
							THEN 1*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate > ?
							THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate,ContractStartDate)+1))
							ELSE 0 END
						) as so_amount
					FROM (
						SELECT * FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active'
						and ContractStartDate <= ? and ContractEndDate >= ?
						and ContractStartDate <= ContractEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', sonumber), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {

		sql := `SELECT distinct Customer_ID
		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					WHERE Active_Inactive = 'Active'
					and ContractStartDate <= ? and ContractEndDate >= ?
					and ContractStartDate <= ContractEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
					and INSTR(CONCAT_WS('|', sonumber), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"customer_total": len(cus),
		"total":          len(sum),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataMap)
}

func GetDetailInvoiceEndPoint(c echo.Context) error {

	type SOCus struct {
		BLSCDocNo   string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		EndDate      string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
		Remark       string  `json:"remark" gorm:"column:remark"`
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

	cus := []struct {
		CusnameThai string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	InvNumber := strings.TrimSpace(c.QueryParam("so_number"))
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
	wg.Add(3)
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
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
						;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, InvNumber, StaffId).Scan(&so).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `SELECT
		sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, InvNumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `		SELECT distinct Customer_ID
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
						;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, InvNumber, StaffId).Scan(&cus).Error; err != nil {
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
		"count_so":       len(so),
		"customer_total": len(cus),
		"total_so":       soTotal,
		"detail":         so,
	}

	return c.JSON(http.StatusOK, dataInv)
}

func GetDetailBillingEndPoint(c echo.Context) error {

	// SELECT * FROM `so_mssql` JOIN invoice_status ON so_mssql.BLSCDocNo = invoice_status.inv_no;

	type SOCusBill struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
		SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
		SaleName            string  `json:"sale_name" gorm:"column:sale_name"`
		SaleTeam            string  `json:"sale_team" gorm:"column:sale_team"`
		SaleFactor          float64 `json:"sale_factor" gorm:"column:sale_factor"`
		InFactor            float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor            float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SoType              string  `json:"SoType" gorm:"column:SoType"`
		Detail              string  `json:"detail" gorm:"column:detail"`
		SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount              float64 `json:"amount" gorm:"column:amount"`
		InvStatusName       string  `json:"status" gorm:"column:status"`
		Reason              string  `json:"reason" gorm:"column:reason"`
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

	status := struct {
		HasBilling       float64 `json:"has_billing" gorm:"column:has_billing"`
		NoBilling        float64 `json:"no_billing" gorm:"column:no_billing"`
		HasBillingAmount float64 `json:"has_billing_amount" gorm:"column:has_billing_amount"`
		NoBillingAmount  float64 `json:"no_billing_amount" gorm:"column:no_billing_amount"`
	}{}

	cus := []struct {
		CusnameThai string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}

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
	wg.Add(4)
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
						status
					 FROM billing_info
					  JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
					  JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					 WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate
							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
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
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		sum(ex_factor) as sum_ef, SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount
						FROM billing_info
							JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
							JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
							WHERE Active_Inactive = 'Active'
							and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate

							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
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
	go func() {
		sql := `SELECT SUM(has_billing) as has_billing,
		SUM(no_billing) as no_billing,
		SUM(has_billing_amount) as has_billing_amount,
		SUM(no_billing_amount) as no_billing_amount
		FROM(
				SELECT
					SUM(CASE
					WHEN status = 'วางบิลแล้ว' THEN 1
				END) has_billing,
				SUM(CASE
					WHEN status = 'วางไม่ได้' THEN 1
				END) no_billing,
				SUM(CASE
						WHEN status = 'วางบิลแล้ว' THEN TotalContractAmount
				END) has_billing_amount,
				SUM(CASE
						WHEN status = 'วางไม่ได้' THEN TotalContractAmount
				END) no_billing_amount
				FROM (

			SELECT
						status,TotalContractAmount
					 FROM billing_info
					  JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
					  JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					 WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate
							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							and INSTR(CONCAT_WS('|', invoice_no), ?)
							group by BLSCDocNo
							) as ss
							group by status
							) as aa;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, StaffId, InvNumber).Scan(&status).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `
			SELECT distinct Customer_ID
					 FROM billing_info
					  JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
					  JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					 WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate
							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							and INSTR(CONCAT_WS('|', invoice_no), ?)
							group by BLSCDocNo;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, StaffId, InvNumber).Scan(&cus).Error; err != nil {
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
		"count_so":       len(so),
		"customer_total": len(cus),
		"total_so":       soTotal,
		"detail":         so,
		"status":         status,
	}

	return c.JSON(http.StatusOK, dataInv)
}

func GetDetailReceiptEndPoint(c echo.Context) error {

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
		BLSCDocNo   string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SO_Type" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		EndDate      string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
	}
	type TrackInvoice struct {
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"sale_factor" gorm:"column:sale_factor"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
	}
	cus := []struct {
		CusnameThai string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
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
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId).Scan(&so).Error; err != nil {
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
	wg.Add(3)
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
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
					;`

		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, InvNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT
		sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, StaffId, InvNumber).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {

		sqlSum := `	SELECT distinct Customer_ID
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active'
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
					;`

		if err := dbSale.Ctx().Raw(sqlSum, dateTo, dateFrom, listInv, listId, search, InvNumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}

		wg.Done()
	}()
	wg.Wait()

	dataReceipt := map[string]interface{}{
		"has_receipt":    len(sum),
		"customer_total": len(cus),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataReceipt)
}

func GetDetailReceiptChangeEndPoint(c echo.Context) error {

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
		BLSCDocNo   string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SO_Type" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		EndDate      string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
	}
	type TrackInvoice struct {
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"sale_factor" gorm:"column:sale_factor"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
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
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId).Scan(&so).Error; err != nil {
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

		sqlSum := `
		select *
		from (
			SELECT *,
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
					sum(PeriodAmount) as amount,
					(CASE
						WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
						WHEN sale_code <> owner Then 'not change'
						Else 'not change' END
					) as status_code
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
				 left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
						WHERE Active_Inactive = 'Active'
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
		) data
		where status_code = 'change'

					;`

		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, InvNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT
		sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						) as so_amount,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					FROM (
						SELECT * FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0
				and status_code = 'change'
				group by BLSCDocNo

			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, StaffId, InvNumber).Scan(&soTotal).Error; err != nil {
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

func GetDetailReceiptNotChangeEndPoint(c echo.Context) error {

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
		BLSCDocNo   string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SO_Type" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		EndDate      string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
	}
	type TrackInvoice struct {
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"sale_factor" gorm:"column:sale_factor"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
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
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId).Scan(&so).Error; err != nil {
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

		sqlSum := `
		select *
		from (
			SELECT *,
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
					sum(PeriodAmount) as amount,
					(CASE
						WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
						WHEN sale_code <> owner Then 'not change'
						Else 'not change' END
					) as status_code
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
				 left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
						WHERE Active_Inactive = 'Active'
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
		) data
		where status_code = 'not change'

					;`

		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, InvNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT
		sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						) as so_amount,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					FROM (
						SELECT *
						FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?)

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
				and status_code = 'not change'
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, StaffId, InvNumber).Scan(&soTotal).Error; err != nil {
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

func GetDetailBillingChangeEndPoint(c echo.Context) error {

	// SELECT * FROM `so_mssql` JOIN invoice_status ON so_mssql.BLSCDocNo = invoice_status.inv_no;

	type SOCusBill struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
		SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
		SaleName            string  `json:"sale_name" gorm:"column:sale_name"`
		SaleTeam            string  `json:"sale_team" gorm:"column:sale_team"`
		SaleFactor          float64 `json:"sale_factor" gorm:"column:sale_factor"`
		InFactor            float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor            float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SoType              string  `json:"SoType" gorm:"column:SoType"`
		Detail              string  `json:"detail" gorm:"column:detail"`
		SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount              float64 `json:"amount" gorm:"column:amount"`
		InvStatusName       string  `json:"status" gorm:"column:status"`
		Reason              string  `json:"reason" gorm:"column:reason"`
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
		sql := `
		select sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,pricesale,TotalContractAmount,SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,sale_factor,in_factor,ex_factor,so_refer,SoType,detail,so_amount,amount,status,reason,staff_id,prefix,fname,lname,position,department
		from (
		SELECT sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,pricesale,TotalContractAmount,SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,sale_factor,in_factor,ex_factor,so_refer,SoType,detail,reason,staff_id,prefix,fname,lname,position,department,
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
						status,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					 FROM billing_info
					  JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
					  JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					  left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
					 WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate
							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							and INSTR(CONCAT_WS('|', invoice_no), ?)
							group by BLSCDocNo
							) data
							where status_code = 'change'
							;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId, InvNumber).Scan(&so).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		sum(ex_factor) as sum_ef, SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount,status_code
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						) as so_amount,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					FROM (
						SELECT SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,owner,owner_old,update_date
						FROM billing_info
						JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
						JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
							left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
							WHERE Active_Inactive = 'Active'
							and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate

							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							and INSTR(CONCAT_WS('|', invoice_no), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0
				and status_code = 'change'
				group by BLSCDocNo
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

func GetDetailBillingNotChangeEndPoint(c echo.Context) error {

	// SELECT * FROM `so_mssql` JOIN invoice_status ON so_mssql.BLSCDocNo = invoice_status.inv_no;

	type SOCusBill struct {
		SOnumber            string  `json:"so_number" gorm:"column:sonumber"`
		ContractStartDate   string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate     string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		BLSCDocNo           string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		PriceSale           float64 `json:"price_sale" gorm:"column:pricesale"`
		TotalContractAmount float64 `json:"TotalContractAmount" gorm:"column:TotalContractAmount"`
		SOWebStatus         string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerId          string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName        string  `json:"customer_name" gorm:"column:Customer_Name"`
		SaleCode            string  `json:"sale_code" gorm:"column:sale_code"`
		SaleName            string  `json:"sale_name" gorm:"column:sale_name"`
		SaleTeam            string  `json:"sale_team" gorm:"column:sale_team"`
		SaleFactor          float64 `json:"sale_factor" gorm:"column:sale_factor"`
		InFactor            float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor            float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SORefer             string  `json:"so_refer" gorm:"column:so_refer"`
		SoType              string  `json:"SoType" gorm:"column:SoType"`
		Detail              string  `json:"detail" gorm:"column:detail"`
		SoAmount            float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount              float64 `json:"amount" gorm:"column:amount"`
		InvStatusName       string  `json:"status" gorm:"column:status"`
		Reason              string  `json:"reason" gorm:"column:reason"`
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
		sql := `
		select sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,pricesale,TotalContractAmount,SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,sale_factor,in_factor,ex_factor,so_refer,SoType,detail,so_amount,amount,status,reason,staff_id,prefix,fname,lname,position,department
		from (
		SELECT sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,pricesale,TotalContractAmount,SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,sale_factor,in_factor,ex_factor,so_refer,SoType,detail,reason,staff_id,prefix,fname,lname,position,department,
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
						status,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					 FROM billing_info
					  JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
					  JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
					  left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
					 WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate
							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							and INSTR(CONCAT_WS('|', invoice_no), ?)
							group by BLSCDocNo
							) data
							where status_code = 'not change'
							;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, StaffId, InvNumber).Scan(&so).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		sum(ex_factor) as sum_ef, SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount,status_code
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						) as so_amount,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					FROM (
						SELECT SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,owner,owner_old,update_date
						FROM billing_info
						JOIN so_mssql ON so_mssql.BLSCDocNo = billing_info.invoice_no
						JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
							left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
							WHERE Active_Inactive = 'Active'
							and BLSCDocNo <> ''
							and PeriodStartDate <= ? and PeriodEndDate >= ?
							and PeriodStartDate <= PeriodEndDate

							and sale_code in (?)
							and INSTR(CONCAT_WS('|', SOnumber,SDPropertyCS28,BLSCDocNo, SOWebStatus,Customer_ID,Customer_Name,sale_code,sale_name,sale_team,so_refer,SoType,detail,status,staff_id,prefix,fname,lname,position,department), ?)
							and INSTR(CONCAT_WS('|', sale_code), ?)
							and INSTR(CONCAT_WS('|', invoice_no), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0
				and status_code = 'not change'
				group by BLSCDocNo
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

func GetDetailInvoiceChangeEndPoint(c echo.Context) error {

	type SOCus struct {
		BLSCDocNo   string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		EndDate      string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
		Remark       string  `json:"remark" gorm:"column:remark"`
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
	InvNumber := strings.TrimSpace(c.QueryParam("so_number"))
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
		sql := `select *
		from (
					SELECT *,
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
								(CASE
									WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
									WHEN sale_code <> owner Then 'not change'
									Else 'not change' END
								) as status_code
							FROM so_mssql
							LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
							left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
									WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
									and PeriodStartDate <= ? and PeriodEndDate >= ?
									and PeriodStartDate <= PeriodEndDate
									and sale_code in (?)
									and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
									and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
									and INSTR(CONCAT_WS('|', sale_code), ?)
									group by BLSCDocNo
									) data
									where status_code = 'change'			;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, InvNumber, StaffId).Scan(&so).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `SELECT
		sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						) as so_amount,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					FROM (
						SELECT * FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0
				and status_code = 'change'
				group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, InvNumber, StaffId).Scan(&soTotal).Error; err != nil {
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

func GetDetailInvoiceNotChangeEndPoint(c echo.Context) error {

	type SOCus struct {
		BLSCDocNo   string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		EndDate      string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
		Remark       string  `json:"remark" gorm:"column:remark"`
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
	InvNumber := strings.TrimSpace(c.QueryParam("so_number"))
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
		sql := `select *
		from (
					SELECT *,
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
								(CASE
									WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
									WHEN sale_code <> owner Then 'not change'
									Else 'not change' END
								) as status_code
							FROM so_mssql
							LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
							left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
									WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
									and PeriodStartDate <= ? and PeriodEndDate >= ?
									and PeriodStartDate <= PeriodEndDate
									and sale_code in (?)
									and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
									and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
									and INSTR(CONCAT_WS('|', sale_code), ?)
									group by BLSCDocNo
									) data
									where status_code = 'not change'			;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, InvNumber, StaffId).Scan(&so).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `SELECT
		sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						) as so_amount,
						(CASE
							WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
							WHEN sale_code <> owner Then 'not change'
							Else 'not change' END
						) as status_code
					FROM (
						SELECT * FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0
				and status_code = 'not change'
				group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, InvNumber, StaffId).Scan(&soTotal).Error; err != nil {
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

func GetDetailSoChangeEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	SONumber := strings.TrimSpace(c.QueryParam("cs_number"))
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
		SOnumber    string  `json:"sonumber" gorm:"column:sonumber"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"ContractStartDate" gorm:"column:ContractStartDate"`
		EndDate      string  `json:"ContractEndDate" gorm:"column:ContractEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
		Remark       string  `json:"remark" gorm:"column:remark"`
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
	cus := []struct {
		CustomerID string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {

		sql := `select *
		from (


		SELECT *,
		SUM(CASE
			WHEN DATEDIFF(ContractEndDate, ContractStartDate)+1 = 0
			THEN 0
			WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate <= ?
			THEN PeriodAmount
			WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate > ?
			THEN (DATEDIFF(?, ContractStartDate)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate <= ? AND ContractEndDate > ?
			THEN (DATEDIFF(ContractEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate = ?
			THEN 1*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate > ?
			THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate,ContractStartDate)+1))
			ELSE 0 END
		) as so_amount,
		sum(PeriodAmount) as amount,
		(CASE
			WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
			WHEN sale_code <> owner Then 'not change'
			Else 'not change' END
		) as status_code

		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id

					WHERE Active_Inactive = 'Active'
					and ContractStartDate <= ? and ContractEndDate >= ?
					and ContractStartDate <= ContractEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
					and INSTR(CONCAT_WS('|', sonumber), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber

					) data
where status_code = 'change'
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount,status_code
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(ContractEndDate, ContractStartDate)+1 = 0
							THEN 0
							WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate <= ?
							THEN PeriodAmount
							WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate > ?
							THEN (DATEDIFF(?, ContractStartDate)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate <= ? AND ContractEndDate > ?
							THEN (DATEDIFF(ContractEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate = ?
							THEN 1*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate > ?
							THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate,ContractStartDate)+1))
							ELSE 0 END
						) as so_amount,
						(CASE
																WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
																WHEN sale_code <> owner Then 'not change'
																Else 'not change' END
															) as status_code

					FROM (
						SELECT *
						FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id

						WHERE Active_Inactive = 'Active'
						and ContractStartDate <= ? and ContractEndDate >= ?
						and ContractStartDate <= ContractEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', sonumber), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0
				and status_code = 'change'
				group by sonumber
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {

		sql := `

select *
from (
		SELECT distinct Customer_ID,(CASE
			WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
			WHEN sale_code <> owner Then 'not change'
			Else 'not change' END
		) as status_code
		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id

					WHERE Active_Inactive = 'Active'
					and ContractStartDate <= ? and ContractEndDate >= ?
					and ContractStartDate <= ContractEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
					and INSTR(CONCAT_WS('|', sonumber), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber

) data
where status_code = 'change'
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"customer_total": len(cus),
		"total":          len(sum),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataMap)
}

func GetDetailSoNotChangeEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	SONumber := strings.TrimSpace(c.QueryParam("cs_number"))
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
		SOnumber    string  `json:"sonumber" gorm:"column:sonumber"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"ContractStartDate" gorm:"column:ContractStartDate"`
		EndDate      string  `json:"ContractEndDate" gorm:"column:ContractEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
		Remark       string  `json:"remark" gorm:"column:remark"`
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
	cus := []struct {
		CustomerID string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {

		sql := `select *
		from (


		SELECT *,
		SUM(CASE
			WHEN DATEDIFF(ContractEndDate, ContractStartDate)+1 = 0
			THEN 0
			WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate <= ?
			THEN PeriodAmount
			WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate > ?
			THEN (DATEDIFF(?, ContractStartDate)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate <= ? AND ContractEndDate > ?
			THEN (DATEDIFF(ContractEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate = ?
			THEN 1*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
			WHEN ContractStartDate < ? AND ContractEndDate > ?
			THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate,ContractStartDate)+1))
			ELSE 0 END
		) as so_amount,
		sum(PeriodAmount) as amount,
		(CASE
			WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
			WHEN sale_code <> owner Then 'not change'
			Else 'not change' END
		) as status_code

		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id

					WHERE Active_Inactive = 'Active'
					and ContractStartDate <= ? and ContractEndDate >= ?
					and ContractStartDate <= ContractEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
					and INSTR(CONCAT_WS('|', sonumber), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber

					) data
where status_code = 'not change'
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount,status_code
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(ContractEndDate, ContractStartDate)+1 = 0
							THEN 0
							WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate <= ?
							THEN PeriodAmount
							WHEN ContractStartDate >= ? AND ContractStartDate <= ? AND ContractEndDate > ?
							THEN (DATEDIFF(?, ContractStartDate)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate <= ? AND ContractEndDate > ?
							THEN (DATEDIFF(ContractEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate = ?
							THEN 1*(PeriodAmount/(DATEDIFF(ContractEndDate, ContractStartDate)+1))
							WHEN ContractStartDate < ? AND ContractEndDate > ?
							THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(ContractEndDate,ContractStartDate)+1))
							ELSE 0 END
						) as so_amount,
						(CASE
																WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < sub_data.update_date and owner <> owner_old Then 'change'
																WHEN sale_code <> owner Then 'not change'
																Else 'not change' END
															) as status_code

					FROM (
						SELECT *
						FROM so_mssql
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id

						WHERE Active_Inactive = 'Active'
						and ContractStartDate <= ? and ContractEndDate >= ?
						and ContractStartDate <= ContractEndDate

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
						and INSTR(CONCAT_WS('|', sonumber), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0
				and status_code = 'not change'
				group by sonumber
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {

		sql := `

select *
from (
		SELECT distinct Customer_ID,(CASE
			WHEN sale_code = owner and owner_old <> '' and PeriodStartDate < cc.update_date and owner <> owner_old Then 'change'
			WHEN sale_code <> owner Then 'not change'
			Else 'not change' END
		) as status_code
		FROM so_mssql
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on so_mssql.Customer_ID = cc.cc_customer_id

					WHERE Active_Inactive = 'Active'
					and ContractStartDate <= ? and ContractEndDate >= ?
					and ContractStartDate <= ContractEndDate
					and sale_code in (?)
					and INSTR(CONCAT_WS('|', sonumber, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail,remark), ?)
					and INSTR(CONCAT_WS('|', sonumber), ?)
					and INSTR(CONCAT_WS('|', sale_code), ?)
					group by sonumber

) data
where status_code = 'not change'
					 ;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, SONumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"customer_total": len(cus),
		"total":          len(sum),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataMap)
}

func GetDetailCostsheetChangeEndPoint(c echo.Context) error {

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
		DocNumberEform string  `json:"doc_number_eform" gorm:"column:doc_number_eform"`
		StaffID        string  `json:"staff_id" gorm:"column:staff_id"`
		Fname          string  `json:"fname" gorm:"column:fname"`
		Lname          string  `json:"lname" gorm:"column:lname"`
		Nname          string  `json:"nname" gorm:"column:nname"`
		Department     string  `json:"department" gorm:"column:department"`
		SoAmount       float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount         float64 `json:"amount" gorm:"column:amount"`
		StatusEform    string  `json:"status" gorm:"column:status"`
		CustomerID     string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai    string  `json:"Cusname_thai" gorm:"column:Cusname_thai"`
		CusnameEng     string  `json:"Cusname_Eng" gorm:"column:Cusname_Eng"`
		BusinessType   string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus      string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType         string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor     float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor       float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor       float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate      string  `json:"StartDate_P1" gorm:"column:StartDate_P1"`
		EndDate        string  `json:"EndDate_P1" gorm:"column:EndDate_P1"`
	}
	type TrackInvoice struct {
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
	}

	hasErr := 0
	var soTotal []TrackInvoice
	var sum []SOCus
	cus := []struct {
		CusnameThai string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		sql := `

select *
from (
	SELECT *,
		SUM(CASE
			WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
			THEN 0
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
			THEN Total_Revenue_Month
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?, StartDate_P1)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(EndDate_P1, ?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 = ?
			THEN 1*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?,?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
			ELSE 0 END
		) as so_amount,
		sum(Total_Revenue_Month) as amount,
		sum(COALESCE(Int_INET, 0))/100 as in_factor,
		sum((COALESCE(Ext_JV, 0) + COALESCE(Ext, 0)))/100 as ex_factor,
		(CASE
			WHEN 	EmployeeID = owner and owner_old <> '' and StartDate_P1 < cc.update_date and owner <> owner_old Then 'change'
			WHEN 	EmployeeID <> owner Then 'not change'
			Else 'not change' END
		) as status_code
		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on costsheet_info.Customer_ID = cc.cc_customer_id

				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					) data
where status_code = 'change'
 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor)/100 as in_factor,
		AVG(ex_factor)/100 as ex_factor,
		(sum(amount)/sum(amount_engcost)) as SaleFactors ,
		SUM(so_amount) as pro_rate
		from (
			SELECT
				Customer_ID as Customer_ID,
				Cusname_thai as Cusname_thai,
				sum(Total_Revenue_Month) as amount,
				sum(eng_cost) as amount_engcost,
				SaleFactors,
				sum(in_factor) as in_factor,
				EmployeeID,
				Sales_Name,
				sum(ex_factor) as ex_factor,
				Total_Revenue_Month,
				so_amount
				FROM (
					SELECT
					doc_number_eform,StartDate_P1,EndDate_P1,Customer_ID,Cusname_thai,
					EmployeeID,	Sales_Name,Sale_Team,Total_Revenue_Month, SaleFactors,
					COALESCE(Int_INET, 0) as in_factor,
					(COALESCE(Ext_JV, 0) + COALESCE(Ext, 0)) as ex_factor,
						(case
							when Total_Revenue_Month is not null and SaleFactors is not null then Total_Revenue_Month/SaleFactors
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
							THEN 0
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
							THEN Total_Revenue_Month
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?, StartDate_P1)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(EndDate_P1, ?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 = ?
							THEN 1*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?,?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
							ELSE 0 END
						) as so_amount,

(CASE
	WHEN 	EmployeeID = owner and owner_old <> '' and StartDate_P1 < sub_data.update_date and owner <> owner_old Then 'change'
	WHEN 	EmployeeID <> owner Then 'not change'
	Else 'not change' END
) as status_code
					FROM (
						SELECT * FROM costsheet_info
						LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on costsheet_info.Customer_ID = cc.cc_customer_id

						WHERE doc_number_eform <> ''
						and StartDate_P1 <= ? and EndDate_P1 >= ?
						and StartDate_P1 <= EndDate_P1

						and EmployeeID in (?)
						and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
						and INSTR(CONCAT_WS('|', doc_number_eform), ?)
						and INSTR(CONCAT_WS('|', EmployeeID), ?)
					) sub_data
				) so_group
				where status_code = 'change'
				group by doc_number_eform
			) cust_group
			`
		// and StartDate_P1 <= ? and EndDate_P1 >= ?
		// and StartDate_P1 <= EndDate_P1
		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `
		select *
		from (
			SELECT distinct Customer_ID,(CASE
			WHEN 	EmployeeID = owner and owner_old <> '' and StartDate_P1 < cc.update_date and owner <> owner_old Then 'change'
			WHEN 	EmployeeID <> owner Then 'not change'
			Else 'not change' END
		) as status_code

		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on costsheet_info.Customer_ID = cc.cc_customer_id

				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					) data
where status_code = 'change' ;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"total":          len(sum),
		"customer_total": len(cus),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataMap)
}

func GetDetailCostsheetNotChangeEndPoint(c echo.Context) error {

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
		DocNumberEform string  `json:"doc_number_eform" gorm:"column:doc_number_eform"`
		StaffID        string  `json:"staff_id" gorm:"column:staff_id"`
		Fname          string  `json:"fname" gorm:"column:fname"`
		Lname          string  `json:"lname" gorm:"column:lname"`
		Nname          string  `json:"nname" gorm:"column:nname"`
		Department     string  `json:"department" gorm:"column:department"`
		SoAmount       float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount         float64 `json:"amount" gorm:"column:amount"`
		StatusEform    string  `json:"status" gorm:"column:status"`
		CustomerID     string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai    string  `json:"Cusname_thai" gorm:"column:Cusname_thai"`
		CusnameEng     string  `json:"Cusname_Eng" gorm:"column:Cusname_Eng"`
		BusinessType   string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus      string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType         string  `json:"SOType" gorm:"column:SOType"`
		SaleFactor     float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor       float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor       float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate      string  `json:"StartDate_P1" gorm:"column:StartDate_P1"`
		EndDate        string  `json:"EndDate_P1" gorm:"column:EndDate_P1"`
	}
	type TrackInvoice struct {
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
	}

	hasErr := 0
	var soTotal []TrackInvoice
	var sum []SOCus
	cus := []struct {
		CusnameThai string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		sql := `

select *
from (
	SELECT *,
		SUM(CASE
			WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
			THEN 0
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
			THEN Total_Revenue_Month
			WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?, StartDate_P1)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
			THEN (DATEDIFF(EndDate_P1, ?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 = ?
			THEN 1*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
			WHEN StartDate_P1 < ? AND EndDate_P1 > ?
			THEN (DATEDIFF(?,?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
			ELSE 0 END
		) as so_amount,
		sum(Total_Revenue_Month) as amount,
		sum(COALESCE(Int_INET, 0))/100 as in_factor,
		sum((COALESCE(Ext_JV, 0) + COALESCE(Ext, 0)))/100 as ex_factor,
		(CASE
			WHEN 	EmployeeID = owner and owner_old <> '' and StartDate_P1 < cc.update_date and owner <> owner_old Then 'change'
			WHEN 	EmployeeID <> owner Then 'not change'
			Else 'not change' END
		) as status_code
		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on costsheet_info.Customer_ID = cc.cc_customer_id

				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					) data
where status_code = 'not change'
 ;`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT sum(amount) as amount,
		AVG(in_factor)/100 as in_factor,
		AVG(ex_factor)/100 as ex_factor,
		(sum(amount)/sum(amount_engcost)) as SaleFactors ,
		SUM(so_amount) as pro_rate
		from (
			SELECT
				Customer_ID as Customer_ID,
				Cusname_thai as Cusname_thai,
				sum(Total_Revenue_Month) as amount,
				sum(eng_cost) as amount_engcost,
				SaleFactors,
				sum(in_factor) as in_factor,
				EmployeeID,
				Sales_Name,
				sum(ex_factor) as ex_factor,
				Total_Revenue_Month,
				so_amount
				FROM (
					SELECT
					doc_number_eform,StartDate_P1,EndDate_P1,Customer_ID,Cusname_thai,
					EmployeeID,	Sales_Name,Sale_Team,Total_Revenue_Month, SaleFactors,
					COALESCE(Int_INET, 0) as in_factor,
					(COALESCE(Ext_JV, 0) + COALESCE(Ext, 0)) as ex_factor,
						(case
							when Total_Revenue_Month is not null and SaleFactors is not null then Total_Revenue_Month/SaleFactors
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(EndDate_P1, StartDate_P1)+1 = 0
							THEN 0
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 <= ?
							THEN Total_Revenue_Month
							WHEN StartDate_P1 >= ? AND StartDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?, StartDate_P1)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 <= ? AND EndDate_P1 > ?
							THEN (DATEDIFF(EndDate_P1, ?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 = ?
							THEN 1*(Total_Revenue_Month/(DATEDIFF(EndDate_P1, StartDate_P1)+1))
							WHEN StartDate_P1 < ? AND EndDate_P1 > ?
							THEN (DATEDIFF(?,?)+1)*(Total_Revenue_Month/(DATEDIFF(EndDate_P1,StartDate_P1)+1))
							ELSE 0 END
						) as so_amount,

(CASE
	WHEN 	EmployeeID = owner and owner_old <> '' and StartDate_P1 < sub_data.update_date and owner <> owner_old Then 'change'
	WHEN 	EmployeeID <> owner Then 'not change'
	Else 'not change' END
) as status_code
					FROM (
						SELECT * FROM costsheet_info
						LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
						left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on costsheet_info.Customer_ID = cc.cc_customer_id

						WHERE doc_number_eform <> ''
						and StartDate_P1 <= ? and EndDate_P1 >= ?
						and StartDate_P1 <= EndDate_P1

						and EmployeeID in (?)
						and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
						and INSTR(CONCAT_WS('|', doc_number_eform), ?)
						and INSTR(CONCAT_WS('|', EmployeeID), ?)
					) sub_data
				) so_group
				where status_code = 'not change'
				group by doc_number_eform
			) cust_group
			`
		// and StartDate_P1 <= ? and EndDate_P1 >= ?
		// and StartDate_P1 <= EndDate_P1
		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {
		sql := `
		select *
		from (
			SELECT distinct Customer_ID,(CASE
			WHEN 	EmployeeID = owner and owner_old <> '' and StartDate_P1 < cc.update_date and owner <> owner_old Then 'change'
			WHEN 	EmployeeID <> owner Then 'not change'
			Else 'not change' END
		) as status_code

		FROM costsheet_info
		LEFT JOIN staff_info ON costsheet_info.EmployeeID = staff_info.staff_id
		left join (select customer_id as cc_customer_id, owner, owner_old,update_date from cust_change) cc on costsheet_info.Customer_ID = cc.cc_customer_id

				 WHERE doc_number_eform <> ''
				 	and StartDate_P1 <= ? and EndDate_P1 >= ?
					and StartDate_P1 <= EndDate_P1
					and EmployeeID in (?)
					and INSTR(CONCAT_WS('|', doc_number_eform, staff_id, fname,lname,nname,department,status,Customer_ID,Cusname_thai,Cusname_Eng), ?)
					and INSTR(CONCAT_WS('|', doc_number_eform), ?)
					and INSTR(CONCAT_WS('|', EmployeeID), ?)
					group by doc_number_eform
					) data
where status_code = 'not change' ;`

		if err := dbSale.Ctx().Raw(sql, dateTo, dateFrom, listId, search, CsNumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}

		wg.Done()
	}()
	wg.Wait()

	dataMap := map[string]interface{}{
		"total":          len(sum),
		"customer_total": len(cus),
		"total_so":       soTotal,
		"detail":         sum,
	}
	return c.JSON(http.StatusOK, dataMap)
}
