package api

import (
	"fmt"
	"net/http"
	"strings"
	// "sale_ranking/pkg/server"
	"sale_ranking/pkg/log"

	"github.com/labstack/echo/v4"
)

func TrackingEndPoint(c echo.Context) error {
	
	Date_st := strings.TrimSpace(c.QueryParam("date_start"))
	Date_en := strings.TrimSpace(c.QueryParam("date_end"))
	customer_id := strings.TrimSpace(c.QueryParam("customer_id"))
	customer_name := strings.TrimSpace(c.QueryParam("customer_name"))
	sale_code := strings.TrimSpace(c.QueryParam("sale_code"))
	sale_name := strings.TrimSpace(c.QueryParam("sale_name"))
	department := strings.TrimSpace(c.QueryParam("department"))
	sonumber := strings.TrimSpace(c.QueryParam("sonumber"))
	status := strings.TrimSpace(c.QueryParam("status"))
	
	if strings.TrimSpace(c.QueryParam("date_start")) == ""||strings.TrimSpace(c.QueryParam("date_end")) == ""{
		return echo.ErrBadRequest
	}
	
	rawData := []struct {
		Total_so 		  string  `json:"total_so" gorm:"column:total_so"`
		Total_cs		  string  `json:"total_cs" gorm:"column:total_cs"`
		Total_inv		  string  `json:"total_inv" gorm:"column:total_inv"`
		Total_rc		  string	`json:"total_rc" gorm:"column:total_rc"`
		Total_cn		  string	`json:"total_cn" gorm:"column:total_cn"`
		So_amount		  string	`json:"so_amount" gorm:"column:so_amount"`
		Inv_amount		  string	`json:"inv_amount" gorm:"column:inv_amount"`
		Cs_amount		  string	`json:"cs_amount" gorm:"column:cs_amount"`
		Rc_amount		  string	`json:"rc_amount" gorm:"column:rc_amount"`
		Cn_amount		  string	`json:"cn_amount" gorm:"column:cn_amount"`
		Amount			  string	`json:"amount" gorm:"column:amount"`
		In_factor		  string	`json:"in_factor" gorm:"column:in_factor"`
		Sum_if			  string	`json:"sum_if" gorm:"column:sum_if"`
		Outstainding_amount	string	`json:"outstainding_amount" gorm:"column:outstainding_amount"`
		Ex_factor		  string	`json:"ex_factor" gorm:"column:ex_factor"`
		Sum_ef			  string	`json:"sum_ef" gorm:"column:sum_ef"`
		Status			  string	`json:"status" gorm:"column:status"`
		Inv_amount_cal	  string	`json:"inv_amount_cal" gorm:"column:inv_amount_cal"`
		Sale_factor		  string	`json:"sale_factor" gorm:"column:sale_factor"`
	}{}

	sql := ` SELECT sum(tr.total_so) as total_so,sum(tr.total_cs) as total_cs,sum(tr.total_inv) as total_inv,sum(tr.total_rc) as total_rc,
	sum(tr.total_cn) as total_cn,sum(tr.so_amount) as so_amount,sum(tr.inv_amount) as inv_amount,sum(tr.cs_amount) as cs_amount,sum(tr.cs_amount) as cs_amount,sum(tr.rc_amount) as rc_amount,
	sum(tr.cn_amount) as cn_amount,sum(tr.amount) as amount,AVG(tr.sum_if) as in_factor,sum(tr.sum_if) as sum_if,sum(tr.inv_amount) - sum(tr.rc_amount) as outstainding_amount,AVG(sum_ef) as ex_factor,sum(tr.sum_ef) as sum_ef,tr.status as status,
	sum((CASE
			WHEN tr.inv_amount = tr.rc_amount THEN tr.inv_amount
			ELSE tr.inv_amount - tr.cn_amount END
		)) as inv_amount_cal,
	(sum(tr.amount)/sum(tr.amount_engcost)) as sale_factor,
	sum(tr.sonumber_all) as sonumber_all
	FROM(
		SELECT DISTINCT Customer_ID as Customer_ID, Customer_Name, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
		sum(so_amount) as so_amount, sum(inv_amount) as inv_amount, sum(cs_amount) as cs_amount, sum(rc_amount) as rc_amount, sum(cn_amount) as cn_amount, sum(amount) as amount,
		sum(in_factor) as sum_if,sale_code,sale_name,sum(ex_factor) as sum_ef,department, nname,amount_engcost,sonumber_name,
		(CASE
			WHEN sum(inv_amount) = 0 THEN 'ยังไม่ออกใบแจ้งหนี้'
			WHEN sum(inv_amount) = sum(cn_amount) THEN 'ลดหนี้'
			WHEN sum(inv_amount) - sum(cn_amount) <= sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระแล้ว'
			WHEN sum(inv_amount) - sum(cn_amount) > sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระไม่ครบ'
			ELSE
				CASE
					WHEN invoice_status_name is not null AND invoice_status_name not like '' THEN invoice_status_name
					ELSE 'ค้างชำระ'
				END
		END
		) as status,
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
				sonumber as sonumber_name,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,invoice_status_name
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,invoice_status_name,
						(case
							when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
							else 0 end
						) as eng_cost,
						(CASE
							WHEN DATEDIFF(PeriodEndDate, PeriodStartDate)+1 = 0
							THEN 0
							WHEN PeriodStartDate >= '`+Date_en+`' AND PeriodStartDate <= '`+Date_st+`' AND PeriodEndDate <= '`+Date_st+`'
							THEN PeriodAmount
							WHEN PeriodStartDate >= '`+Date_en+`' AND PeriodStartDate <= '`+Date_st+`' AND PeriodEndDate > '`+Date_st+`'
							THEN (DATEDIFF('`+Date_st+`' , PeriodStartDate)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
							WHEN PeriodStartDate <  '`+Date_en+`'  AND PeriodEndDate <= '`+Date_st+`'  AND PeriodEndDate >  '`+Date_en+`' 
							THEN (DATEDIFF(PeriodEndDate,  '`+Date_en+`' )+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
							WHEN PeriodStartDate <  '`+Date_en+`'  AND PeriodEndDate =  '`+Date_en+`'
							THEN 1*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
							WHEN PeriodStartDate <'`+Date_en+`'  AND PeriodEndDate > '`+Date_st+`'
							THEN (DATEDIFF('`+Date_st+`', '`+Date_en+`')+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate,PeriodStartDate)+1))
							ELSE 0 END
						) as so_amount
					FROM (
						SELECT so.*,inv_st.invoice_status_name 
						FROM so_mssql so
						LEFT JOIN invoice_status inv_st ON so.BLSCDocNo = inv_st.inv_no
						WHERE so.Active_Inactive = 'Active'
						and so.PeriodStartDate <= '`+Date_st+`' and so.PeriodEndDate >= '`+Date_en+`'
						and so.PeriodStartDate <= so.PeriodEndDate
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by sonumber
		) cust_group
			LEFT JOIN staff_info ON cust_group.sale_code = staff_info.staff_id
			group by Customer_ID
			) tr `
	
	if customer_id != "" || customer_name != "" || sale_code != "" || sale_name != "" || department != "" ||
	sonumber != "" || status != ""{
		sql = sql+` WHERE `
		if customer_id != ""{
			sql = sql+` tr.Customer_ID = `+customer_id+` `
			if customer_name != "" || sale_code != "" || sale_name != "" || department != "" ||
			sonumber != "" || status != ""{
				sql = sql+` AND `
			}
		}
		if customer_name != ""{
			sql = sql+` tr.Customer_Name like '%`+customer_name+`%' `
			if sale_code != "" || sale_name != "" || department != "" ||
			sonumber != "" || status != ""{
				sql = sql+` AND `
			}
		}
		if sale_code != ""{
			sql = sql+` tr.sale_code = `+sale_code+` `
			if sale_name != "" || department != "" ||
			sonumber != "" || status != ""{
				sql = sql+` AND `
			}
		}
		if sale_name != ""{
			sql = sql+` tr.sale_name like '%`+sale_name+`%' `
			if department != "" || sonumber != "" || status != ""{
				sql = sql+` AND `
			}
		}
		if department != ""{
			sql = sql+` tr.department like '%`+department+`%' `
			if sonumber != "" || status != ""{
				sql = sql+` AND `
			}
		}
		if sonumber != ""{
			sql = sql+` tr.sonumber_name like '%`+sonumber+`%' `
			if status != ""{
				sql = sql+` AND `
			}
		}
		if status != ""{
			sql = sql+` tr.status like '%`+status+`%' `
		}
	}
	sql = sql+` GROUP BY status`
	
	if err := dbSale.Ctx().Raw(sql).Scan(&rawData).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}


	fmt.Println(Date_st)
	fmt.Println(Date_en)
	return c.JSON(http.StatusOK, rawData)
}
