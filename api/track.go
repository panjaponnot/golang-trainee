package api

import (
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/util"
	"strconv"
	"strings"
	"sync"
	"time"

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
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 24, 0, 0, 0, time.Local)
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
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 24, 0, 0, 0, time.Local)
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
	// var listInvBilling []string
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
				// listInvBilling = append(listInvBilling, v.InvoiceNo)
				hadBill += 1
			} else {
				hasBill += 1
			}
		} else {
			nonBill += 1
		}
	}
	// var reciept []m.SOMssql
	// if err := dbSale.Ctx().Raw(`select * from so_mssql where BLSCDocNo IN (?) and INCSCDocNo <> ''`, listInvBilling).Scan(&reciept).Error; err != nil {
	// 	log.Errorln(pkgName, err, "select data error -:")
	// }

	log.Infoln("INV", "=========>", len(inv))
	dataInv := map[string]int{
		"billing":      hadBill,
		"ิnon_billing": hasBill,
	}
	dataRaw := map[string]interface{}{
		"has_bill":   dataInv,
		"non_bill":   nonBill,
		"date_start": dateFrom.Format("02-Jan-2006"),
		"date_end":   dateTo.Format("02-Jan-2006"),
	}

	return c.JSON(http.StatusOK, dataRaw)
}
