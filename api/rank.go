package api

import (
	"fmt"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/util"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func GetRankingBaseSale(c echo.Context) error {
	filterDepart := strings.Split(util.GetEnv("CONDITION_BASE_SALE", ""), ",")
	var dFilter []string
	for _, v := range filterDepart {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}
	listStaffId, err := CheckPermissionBaseSale(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		log.Errorln(pkgName, err, "func check permission error :-")
		return c.JSON(http.StatusInternalServerError, m.Result{Error: "check permission error"})
	}
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}
	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
	q := strings.TrimSpace(c.QueryParam("quarter"))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
		WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
		WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
		WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
		WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
		WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
		WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
		ELSE -5 END
	) as score_target,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 999
		ELSE (sum(inv_amount)/goal_total)*100 END
	) as percent_target,
	(CASE
		WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
		WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
		WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
		WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
		WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
		WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
		WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
		ELSE 0 END
	) as score_sf,
	(case
		when sum(revenue)/sum(engcost) is null then 0
		else sum(revenue)/sum(engcost) end
	) as sale_factor,
	(CASE
		WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
		WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
		WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
		WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
		WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
		WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
		WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
		ELSE 0 END
	) as score_if,
	(case
		when sum(sum_if)/sum(total_so) is null then 0
		else sum(sum_if)/sum(total_so) end
	) as in_factor,
	sum(revenue) as total_revenue,
	all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
	from (
		select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
		(case
			when inv_amount is null then 0
			ELSE inv_amount END) as inv_amount
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,
			'normal' as typestaff,
			sum((CASE
				WHEN TotalContractAmount is null THEN 0
				ELSE TotalContractAmount END
			)) as revenue,
			sum((CASE
				WHEN eng_cost is null THEN 0
				ELSE eng_cost END
			)) as engcost,
			sum((CASE
				WHEN in_factor is null THEN 0
				ELSE in_factor END
			)) as sum_if,
			sum((CASE
				WHEN sonumber is null THEN 0
				ELSE 1 END
			)) as total_so,
			one_id,
			quarter,year,job_months,staff_child
			from (
					select staff_id,fname,lname,nname,department,position,
					(CASE
						WHEN goal_total is null THEN 0
						ELSE goal_total END
					) as goal_total,
					staff_info.one_id,
					(CASE
						WHEN quarter is null THEN ?
						ELSE quarter END
					) as quarter,
					(CASE
						WHEN year is null THEN year(now())
						ELSE year END
					) as year,
					12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
					staff_child
					from staff_info
					left join
					(
						select * from goal_quarter where year = year(now()) and quarter =  ?
					) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
					left join staff_start on staff_info.one_id = staff_start.one_id
					WHERE staff_id in (
						select staff_id from staff_info WHERE staff_child = ''
					)
					group by staff_id
			) staff_detail
			LEFT JOIN (
				select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
				group by sonumber
			) total_so on total_so.sale_code = staff_detail.staff_id
			group by staff_id
		) tb_main
		LEFT join (
			select sum(PeriodAmount) as inv_amount, sale_code from (
				select PeriodAmount,sale_code
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now())   and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv group by sale_code
		) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
		where staff_id is not null and staff_id <> ''
	) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
	WHERE staff_id in (?)
	group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child = ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`
	sqlFilter := `select * from staff_info where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?) `

	var staffInfo []m.StaffInfo
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	wg.Wait()
	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for _, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
				if r.InvAmount > r.InvAmountOld {
					var so []m.SOMssql
					if err := dbSale.Ctx().Model(&m.SOMssql{}).Where(`sale_code = ? AND getCN <> ''`, r.StaffId).Group("getCN").Find(&so).Error; err != nil {
					}
					log.Infoln("SO", "=====>", len(so))

					// wait cal aging & blacklist

					baseCal := r.InvAmountOld * 0.003
					growthCal := (r.InvAmount - r.InvAmountOld) * 0.03
					saleFactor := (baseCal + growthCal) * (r.SaleFactor * r.SaleFactor)
					r.Commission = saleFactor * (r.InFactor / 0.7)

				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth

		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, r)
				}
			}
		}
	}

	if len(dataResult) > 1 {
		sort.SliceStable(dataResult, func(i, j int) bool { return dataResult[i].ScoreAll > dataResult[j].ScoreAll })
	}

	var result m.Result
	if len(dataResult) > (page * 10) {
		start := (page - 1) * 10
		end := (page * 10)
		result.Data = dataResult[start:end]
		result.Count = len(dataResult[start:end])
	} else {
		start := (page * 10) - (10)
		result.Data = dataResult[start:]
		result.Count = len(dataResult[start:])
	}
	result.Total = len(dataResult)
	return c.JSON(http.StatusOK, result)
}

func GetRankingKeyAccountEndPoint(c echo.Context) error {
	conKey := strings.Split(util.GetEnv("CONDITION_GOV_KEY_SALE", ""), ",")
	var dFilter []string
	for _, v := range conKey {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}

	listStaffId, err := CheckPermissionKeyAccount(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		return echo.ErrInternalServerError
	}
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}

	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
	q := strings.TrimSpace(c.QueryParam("quarter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
		WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
		WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
		WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
		WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
		WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
		WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
		ELSE -5 END
	) as score_target,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 999
		ELSE (sum(inv_amount)/goal_total)*100 END
	) as percent_target,
	(CASE
		WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
		WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
		WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
		WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
		WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
		WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
		WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
		ELSE 0 END
	) as score_sf,
	(case
		when sum(revenue)/sum(engcost) is null then 0
		else sum(revenue)/sum(engcost) end
	) as sale_factor,
	(CASE
		WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
		WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
		WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
		WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
		WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
		WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
		WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
		ELSE 0 END
	) as score_if,
	(case
		when sum(sum_if)/sum(total_so) is null then 0
		else sum(sum_if)/sum(total_so) end
	) as in_factor,
	sum(revenue) as total_revenue,
	all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
	from (
		select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
		(case
			when inv_amount is null then 0
			ELSE inv_amount END) as inv_amount
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,
			'normal' as typestaff,
			sum((CASE
				WHEN TotalContractAmount is null THEN 0
				ELSE TotalContractAmount END
			)) as revenue,
			sum((CASE
				WHEN eng_cost is null THEN 0
				ELSE eng_cost END
			)) as engcost,
			sum((CASE
				WHEN in_factor is null THEN 0
				ELSE in_factor END
			)) as sum_if,
			sum((CASE
				WHEN sonumber is null THEN 0
				ELSE 1 END
			)) as total_so,
			one_id,
			quarter,year,job_months,staff_child
			from (
					select staff_id,fname,lname,nname,department,position,
					(CASE
						WHEN goal_total is null THEN 0
						ELSE goal_total END
					) as goal_total,
					staff_info.one_id,
					(CASE
						WHEN quarter is null THEN ?
						ELSE quarter END
					) as quarter,
					(CASE
						WHEN year is null THEN year(now())
						ELSE year END
					) as year,
					12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
					staff_child
					from staff_info
					left join
					(
						select * from goal_quarter where year = year(now()) and quarter =  ?
					) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
					left join staff_start on staff_info.one_id = staff_start.one_id
					WHERE staff_id in (
						select staff_id from staff_info WHERE staff_child = ''
					)
					group by staff_id
			) staff_detail
			LEFT JOIN (
				select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
				group by sonumber
			) total_so on total_so.sale_code = staff_detail.staff_id
			group by staff_id
		) tb_main
		LEFT join (
			select sum(PeriodAmount) as inv_amount, sale_code from (
				select PeriodAmount,sale_code
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now())   and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv group by sale_code
		) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
		where staff_id is not null and staff_id <> ''
	) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
	WHERE staff_id in (?)
	group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child = ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`

	sqlFilter := `select * from staff_info where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?) `

	var staffInfo []m.StaffInfo
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	wg.Wait()
	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for _, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
				if r.InvAmount > r.InvAmountOld {
					// wait cal aging & blacklist
					baseCal := r.InvAmountOld * 0.003
					growthCal := (r.InvAmount - r.InvAmountOld) * 0.03
					saleFactor := (baseCal + growthCal) * (r.SaleFactor * r.SaleFactor)
					r.Commission = saleFactor * (r.InFactor / 0.7)
				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, r)
				}
			}
		}
	}
	if len(dataResult) > 1 {
		sort.SliceStable(dataResult, func(i, j int) bool { return dataResult[i].ScoreAll > dataResult[j].ScoreAll })
	}
	var result m.Result
	if len(dataResult) > (page * 10) {
		start := (page - 1) * 10
		end := (page * 10)
		result.Data = dataResult[start:end]
		result.Count = len(dataResult[start:end])
	} else {
		start := (page * 10) - (10)
		result.Data = dataResult[start:]
		result.Count = len(dataResult[start:])
	}
	result.Total = len(dataResult)
	return c.JSON(http.StatusOK, result)
}

func GetRankingRecoveryEndPoint(c echo.Context) error {

	conKey := strings.Split(util.GetEnv("CONDITION_GOV_RECOVER_SALE", ""), ",")
	var dFilter []string
	for _, v := range conKey {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}

	listStaffId, err := CheckPermissionRecovery(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		return echo.ErrInternalServerError
	}
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}
	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
	q := strings.TrimSpace(c.QueryParam("quarter"))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
		WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
		WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
		WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
		WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
		WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
		WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
		ELSE -5 END
	) as score_target,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 999
		ELSE (sum(inv_amount)/goal_total)*100 END
	) as percent_target,
	(CASE
		WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
		WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
		WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
		WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
		WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
		WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
		WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
		ELSE 0 END
	) as score_sf,
	(case
		when sum(revenue)/sum(engcost) is null then 0
		else sum(revenue)/sum(engcost) end
	) as sale_factor,
	(CASE
		WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
		WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
		WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
		WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
		WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
		WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
		WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
		ELSE 0 END
	) as score_if,
	(case
		when sum(sum_if)/sum(total_so) is null then 0
		else sum(sum_if)/sum(total_so) end
	) as in_factor,
	sum(revenue) as total_revenue,
	all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
	from (
		select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
		(case
			when inv_amount is null then 0
			ELSE inv_amount END) as inv_amount
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,
			'normal' as typestaff,
			sum((CASE
				WHEN TotalContractAmount is null THEN 0
				ELSE TotalContractAmount END
			)) as revenue,
			sum((CASE
				WHEN eng_cost is null THEN 0
				ELSE eng_cost END
			)) as engcost,
			sum((CASE
				WHEN in_factor is null THEN 0
				ELSE in_factor END
			)) as sum_if,
			sum((CASE
				WHEN sonumber is null THEN 0
				ELSE 1 END
			)) as total_so,
			one_id,
			quarter,year,job_months,staff_child
			from (
					select staff_id,fname,lname,nname,department,position,
					(CASE
						WHEN goal_total is null THEN 0
						ELSE goal_total END
					) as goal_total,
					staff_info.one_id,
					(CASE
						WHEN quarter is null THEN ?
						ELSE quarter END
					) as quarter,
					(CASE
						WHEN year is null THEN year(now())
						ELSE year END
					) as year,
					12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
					staff_child
					from staff_info
					left join
					(
						select * from goal_quarter where year = year(now()) and quarter =  ?
					) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
					left join staff_start on staff_info.one_id = staff_start.one_id
					WHERE staff_id in (
						select staff_id from staff_info WHERE staff_child = ''
					)
					group by staff_id
			) staff_detail
			LEFT JOIN (
				select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
				group by sonumber
			) total_so on total_so.sale_code = staff_detail.staff_id
			group by staff_id
		) tb_main
		LEFT join (
			select sum(PeriodAmount) as inv_amount, sale_code from (
				select PeriodAmount,sale_code
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now())   and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv group by sale_code
		) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
		where staff_id is not null and staff_id <> ''
	) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
	WHERE staff_id in (?)
	group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child = ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`
	sqlFilter := `select * from staff_info where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?) `

	var staffInfo []m.StaffInfo
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for _, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, r)
				}
			}
		}
	}
	if len(dataResult) != 0 {
		sort.SliceStable(dataResult, func(i, j int) bool { return dataResult[i].ScoreAll > dataResult[j].ScoreAll })
	}
	var result m.Result
	if len(dataResult) > (page * 10) {
		start := (page - 1) * 10
		end := (page * 10)
		result.Data = dataResult[start:end]
		result.Count = len(dataResult[start:end])
	} else {
		start := (page * 10) - (10)
		result.Data = dataResult[start:]
		result.Count = len(dataResult[start:])
	}
	result.Total = len(dataResult)
	return c.JSON(http.StatusOK, result)
}

func GetRankingTeamLeadEndPoint(c echo.Context) error {

	conKey := strings.Split(util.GetEnv("CONDITION_GOV_RECOVER_SALE", ""), ",")
	var dFilter []string
	for _, v := range conKey {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}

	listStaffId, err := CheckPermissionTeamLead(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		return echo.ErrInternalServerError
	}

	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
	q := strings.TrimSpace(c.QueryParam("quarter"))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
		(CASE
			WHEN goal_total is null or goal_total = 0 THEN 30
			WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
			WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
			WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
			WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
			WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
			WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
			WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
			WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
			ELSE -5 END
		) as score_target,
		(CASE
			WHEN goal_total is null or goal_total = 0 THEN 999
			ELSE (sum(inv_amount)/goal_total)*100 END
		) as percent_target,
		(CASE
			WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
			WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
			WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
			WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
			WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
			WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
			WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
			ELSE 0 END
		) as score_sf,
		(case
			when sum(revenue)/sum(engcost) is null then 0
			else sum(revenue)/sum(engcost) end
		) as sale_factor,
		(CASE
			WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
			WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
			WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
			WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
			WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
			WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
			WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
			ELSE 0 END
		) as score_if,
		(case
			when sum(sum_if)/sum(total_so) is null then 0
			else sum(sum_if)/sum(total_so) end
		) as in_factor,
		sum(revenue) as total_revenue,
		all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
			(case
				when inv_amount is null then 0
				ELSE inv_amount END) as inv_amount
			from (
				select staff_id,fname,lname,nname,department,position,goal_total,
				'normal' as typestaff,
				sum((CASE
					WHEN TotalContractAmount is null THEN 0
					ELSE TotalContractAmount END
				)) as revenue,
				sum((CASE
					WHEN eng_cost is null THEN 0
					ELSE eng_cost END
				)) as engcost,
				sum((CASE
					WHEN in_factor is null THEN 0
					ELSE in_factor END
				)) as sum_if,
				sum((CASE
					WHEN sonumber is null THEN 0
					ELSE 1 END
				)) as total_so,
				one_id,
				quarter,year,job_months,staff_child
				from (
						select staff_id,fname,lname,nname,department,position,
						(CASE
							WHEN goal_total is null THEN 0
							ELSE goal_total END
						) as goal_total,
						staff_info.one_id,
						(CASE
							WHEN quarter is null THEN ?
							ELSE quarter END
						) as quarter,
						(CASE
							WHEN year is null THEN year(now())
							ELSE year END
						) as year,
						12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
						staff_child
						from staff_info
						left join
						(
							select * from goal_quarter where year = year(now()) and quarter = ?
						) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
						left join staff_start on staff_info.one_id = staff_start.one_id
						WHERE staff_id in (
							select staff_id from staff_info WHERE staff_child <> ''
						)
						group by staff_id
				) staff_detail
				LEFT JOIN (
					select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
					group by sonumber
				) total_so on total_so.sale_code = staff_detail.staff_id
				group by staff_id
			) tb_main
			LEFT join (
				select sum(PeriodAmount) as inv_amount, sale_code from (
					select PeriodAmount,sale_code
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
					group by sonumber
				) tb_inv group by sale_code
			) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
			where staff_id is not null and staff_id <> ''
			union
			select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
			(case
				when inv_amount is null then 0
				ELSE inv_amount END) as inv_amount
			from (
				select staff_id,fname,lname,nname,department,position,goal_total,
				'normal' as typestaff,
				sum((CASE
					WHEN TotalContractAmount is null THEN 0
					ELSE TotalContractAmount END
				)) as revenue,
				sum((CASE
					WHEN eng_cost is null THEN 0
					ELSE eng_cost END
				)) as engcost,
				sum((CASE
					WHEN in_factor is null THEN 0
					ELSE in_factor END
				)) as sum_if,
				sum((CASE
					WHEN sonumber is null THEN 0
					ELSE 1 END
				)) as total_so,
				one_id,
				quarter,year,job_months,staff_child
				from (
						select staff_id,fname,lname,nname,department,position,
						(CASE
							WHEN goal_total is null THEN 0
							ELSE goal_total END
						) as goal_total,
						staff_info.one_id,
						(CASE
							WHEN quarter is null THEN ?
							ELSE quarter END
						) as quarter,
						(CASE
							WHEN year is null THEN year(now())
							ELSE year END
						) as year,
						12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
						staff_child
						from staff_info
						left join
						(
							select * from goal_quarter where year = year(now()) and quarter = ?
						) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
						left join staff_start on staff_info.one_id = staff_start.one_id
						WHERE staff_id in (
							select staff_id from staff_info WHERE staff_child <> ''
						)
						group by staff_id
				) staff_detail
				LEFT JOIN (
					select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
					group by sonumber
				) total_so on total_so.sale_lead = staff_detail.staff_id
				group by staff_id
			) tb_main
			LEFT join (
				select sum(PeriodAmount) as inv_amount,sale_lead from (
					select PeriodAmount,sale_code,sale_lead
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
					group by sonumber
				) tb_inv group by sale_lead
			) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_lead
			where staff_id is not null and staff_id <> ''
		) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
		WHERE staff_id in (?) 
		group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child <> ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
				union
				select PeriodAmount,sale_lead as sale_code,sonumber, 'lead'
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`

	sqlFilter := `select * from staff_info where staff_child <> '' AND  INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?)  `

	var staffInfo []m.StaffInfo
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for _, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth
		// dataResult = append(dataResult, r)
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				child := strings.Split(st.StaffChild, ",")
				if st.StaffId == r.StaffId && len(child) < 10 {
					dataResult = append(dataResult, r)
				}
			}
		}
	}
	if len(dataResult) != 0 {
		sort.SliceStable(dataResult, func(i, j int) bool { return dataResult[i].ScoreAll > dataResult[j].ScoreAll })
	}
	var result m.Result
	if len(dataResult) > (page * 10) {
		start := (page - 1) * 10
		end := (page * 10)
		result.Data = dataResult[start:end]
		result.Count = len(dataResult[start:end])
	} else {
		start := (page * 10) - (10)
		result.Data = dataResult[start:]
		result.Count = len(dataResult[start:])
	}
	result.Total = len(dataResult)
	return c.JSON(http.StatusOK, result)
}
