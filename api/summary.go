package api

import (
	"fmt"
	"net/http"
	"sale_ranking/model"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
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

func CheckPermissionSummary(acc string) (bool, string) {
	var user m.UserInfo
	if err := dbSale.Ctx().Model(&m.UserInfo{}).First(&user).Error; err == nil {
		return true, user.SubRole
	}

	var staff m.StaffInfo
	if err := dbSale.Ctx().Model(&m.StaffInfo{}).First(&staff).Error; err == nil {
		return true, "sale"
	}
	return false, ""
}

func GetSummarySOPending(c echo.Context) error {

	if strings.TrimSpace(c.Param("id")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}
	accountId := strings.TrimSpace(c.Param("id"))
	log.Infoln("ACC", accountId)
	check, role := CheckPermissionSummary(accountId)
	if !check {
		return echo.ErrUnauthorized
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
	// var rawData []PendingData
	var so []m.SOMssql
	if role == "admin" {
		log.Infoln("ROLE", "admin")
		// 	if err := dbSale.Ctx().Raw(`
		// SELECT Active_Inactive,has_refer,tb_ch_so.sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,
		// so_refer,sale_code,sale_lead,DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,PeriodAmount,
		//  SUM(PeriodAmount) as TotalAmount,staff_id,prefix,fname,lname,nname,position,department,SOType as so_type,
		//     (case
		//             when status is null then 0
		//             else status end
		//     ) as status,
		//       (case
		//             when tb_expire.remark is null then ''
		//             else tb_expire.remark end
		//     ) as remark  from (
		//             SELECT *  from (
		//             SELECT  Active_Inactive,has_refer,sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,
		//                             DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,
		//                                                             PeriodAmount, SUM(PeriodAmount) as TotalAmount,SOType,
		//                                                             staff_id,prefix,fname,lname,nname,position,department
		//                                                             FROM ( SELECT * FROM so_mssql WHERE SOType NOT IN ('onetime' , 'project base') ) as s
		//                                                     left join
		//                                                     (
		//                                                             select staff_id, prefix, fname, lname, nname, position, department from staff_info

		//                                                     ) tb_sale on s.sale_code = tb_sale.staff_id
		//                                                     WHERE Active_Inactive = 'Active' and has_refer = 0 and year(ContractEndDate) = ? and month(ContractEndDate) = ?
		//                                                     group by sonumber
		//                     ) as tb_so_number

		//             ) as tb_ch_so
		//             left join
		//             (
		//               select id,sonumber,
		//                     (case
		//                             when status is null then 0
		//                             else status end
		//                     ) as status,
		//                     (case
		//                             when remark is null then ''
		//                             else remark end
		//                     ) as remark
		//                     from check_expire
		// 			  ) tb_expire on tb_ch_so.sonumber = tb_expire.sonumber
		// 			  WHERE INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,Customer_ID,Customer_Name), ?)
		//               group by tb_ch_so.sonumber
		// 	  `, year, month, search).Scan(&rawData).Error; err != nil {
		// 		log.Errorln(pkgName, err, "Select data error")
		// 	}

		sqlActive := `SELECT Customer_ID from so_mssql WHERE SOType NOT IN ('onetime' , 'project base')  and Active_Inactive = 'Active' and has_refer = 0 and year(ContractEndDate) = ? and month(ContractEndDate) = ?  group by sonumber`
		if err := dbSale.Ctx().Raw(sqlActive, year, month).Scan(&so).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
		}
	} else {

		log.Infoln("ROLE", "sale")
	}

	return c.JSON(http.StatusOK, m.Result{Data: len(so)})
}
