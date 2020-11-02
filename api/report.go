package api

import (
	"fmt"
	"net/http"
	"sale_ranking/model"
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

func GetDataOrgChartEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer dbSale.Close()

	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" {
		return c.JSON(http.StatusBadRequest, model.Result{Message: "invalid staff id"})
	}
	// staffId := strings.TrimSpace(c.QueryParam(("staff_id")))
	filter := strings.TrimSpace(c.QueryParam(("filter")))
	listStaffId, err := CheckPermissionOrg(strings.TrimSpace(c.QueryParam(("staff_id"))))
	if err != nil {
		log.Errorln(pkgName, err, "func check permission error :-")
		return c.JSON(http.StatusInternalServerError, model.Result{Error: "check permission error"})
	}
	// fmt.Println("listStaffId ===>", listStaffId[strings.TrimSpace(c.QueryParam(("staff_id")))])
	fmt.Println("listStaffId ===>", len(listStaffId))
	// return c.JSON(http.StatusNoContent, nil)
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	p := server.GetPagination(c)
	p.Page = uint(page)
	p.Size = server.DefaultQuerySize
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		p.Page = server.DefaultQueryPage
		p.Size = server.MaxQuerySize
	}
	type OrgChart struct {
		StaffId      string  `json:"staff_id"`
		Fname        string  `json:"fname"`
		Lname        string  `json:"lname"`
		Nname        string  `json:"nname"`
		Position     string  `json:"position"`
		Department   string  `json:"department"`
		StaffChild   string  `json:"staff_child"`
		InvAmount    float64 `json:"inv_amount"`
		InvAmountOld float64 `json:"inv_amount_old"`
		GoalTotal    float64 `json:"goal_total"`
		ScoreTarget  float64 `json:"score_target"`
		ScoreSf      float64 `json:"score_sf"`
		SaleFactor   float64 `json:"sale_factor"`
		TotalSo      float64 `json:"total_so"`
		IfFactor     float64 `json:"if_factor"`
		EngCost      float64 `json:"engcost" gorm:"column:engcost"`
		Revenue      float64 `json:"revanue"`
		ScoreIf      float64 `json:"score_if"`
		InFactor     float64 `json:"in_factor"`
		OneId        string  `json:"one_id"`
		// Image        string  `json:"image"`
		FileName    string  `json:"filename" gorm:"column:filename"`
		GrowthRate  float64 `json:"growth_rate"`
		ScoreGrowth string  `json:"score_growth"`
		ScoreAll    float64 `json:"score_all"`
		Quarter     string  `json:"quarter"`
		Year        float64 `json:"year"`
		JobMonths   int     `json:"job_months"`
	}
	type InvBefore struct {
		StaffID   string  `json:"staff_id"`
		InvAmount float64 `json:"inv_amount"`
		CheckData int     `json:"check_data" gorm:"column:checkdata"`
	}
	var org []OrgChart
	var filterOrg []OrgChart
	var inv []InvBefore
	var result model.Result
	today := time.Now()
	yearNow, mon, _ := today.Date()
	yearBefore := yearNow
	month := int(mon)
	var quarterBefore string
	var quarterBeforeNum int

	if month >= 1 && 3 >= month {
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if month >= 4 && 6 >= month {
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if month >= 7 && 9 >= month {
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}

	sqlStr := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, 0 as inv_amount_old, -5 as score_target,
            -10 as score_sf, 0.0 as sale_factor,sum(total_so) as total_so, sum(sum_if) as if_factor, sum(engcost) as engcost, sum(revenue) as revenue,
            -6 as score_if, 0.0 as in_factor,
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
            					WHEN quarter is null THEN concat('Q', CONVERT(quarter(now()), CHAR))
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
            					select * from goal_quarter where year = year(now()) and quarter = concat('Q', CONVERT(quarter(now()), CHAR))
            				) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
            				left join staff_start on staff_info.one_id = staff_start.one_id
            				group by staff_id
            		) staff_detail
            		LEFT JOIN (
            			select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
            			from so_mssql
            			WHERE quarter(ContractStartDate) = quarter(now()) and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
            			group by sonumber
            		) total_so on total_so.sale_code = staff_detail.staff_id
            		group by staff_id
            	) tb_main
            	LEFT join (
            		select sum(PeriodAmount) as inv_amount, sale_code from (
            			select PeriodAmount,sale_code
            			from so_mssql
            			WHERE quarter(ContractStartDate) = quarter(now()) and year(ContractStartDate) = year(now()) and so_refer = '' and Active_Inactive = 'Active'
            			group by sonumber
            		) tb_inv group by sale_code
            	) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
            	where staff_id is not null and staff_id <> ''
			) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
			group by staff_id`
	sqlFilter := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, 0 as inv_amount_old, -5 as score_target,
            -10 as score_sf, 0.0 as sale_factor,sum(total_so) as total_so, sum(sum_if) as if_factor, sum(engcost) as engcost, sum(revenue) as revenue,
            -6 as score_if, 0.0 as in_factor,
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
            					WHEN quarter is null THEN concat('Q', CONVERT(quarter(now()), CHAR))
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
            					select * from goal_quarter where year = year(now()) and quarter = concat('Q', CONVERT(quarter(now()), CHAR))
            				) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
            				left join staff_start on staff_info.one_id = staff_start.one_id
            				group by staff_id
            		) staff_detail
            		LEFT JOIN (
            			select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
            			from so_mssql
            			WHERE quarter(ContractStartDate) = quarter(now()) and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
            			group by sonumber
            		) total_so on total_so.sale_code = staff_detail.staff_id
            		group by staff_id
            	) tb_main
            	LEFT join (
            		select sum(PeriodAmount) as inv_amount, sale_code from (
            			select PeriodAmount,sale_code
            			from so_mssql
            			WHERE quarter(ContractStartDate) = quarter(now()) and year(ContractStartDate) = year(now()) and so_refer = '' and Active_Inactive = 'Active'
            			group by sonumber
            		) tb_inv group by sale_code
            	) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
            	where staff_id is not null and staff_id <> ''
			) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
            WHERE INSTR(CONCAT_WS('|', fname,staff_id,lname,nname,department,all_ranking.one_id,position), ?)
			group by staff_id`

	sqlInv := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
                from (
                	select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
                	from (
                		select staff_id from staff_info
                		left join
                		(
                			select * from goal_quarter where year = ? and quarter = ?
                		) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
                		group by staff_id
                	) staff_detail
                	LEFT JOIN (
                		select PeriodAmount,sale_code,sonumber, type_sale
                        from (
                        	select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
                        	from so_mssql
                        	WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active'
                        	group by sonumber
                        ) tb_inv_old
                	) total_new_so on total_new_so.sale_code = staff_detail.staff_id
                	where staff_id is not null and staff_id <> '' and sale_code is not null
                	group by staff_id
                ) all_ranking
                group by staff_id; `

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		if err := dbSale.Ctx().Raw(sqlStr).Scan(&org).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		// query before
		if err := dbSale.Ctx().Raw(sqlInv, yearBefore, quarterBefore, quarterBeforeNum, yearBefore).Scan(&inv).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Offset(p.Offset()).Limit(p.Size).Scan(&filterOrg).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
			}
		}
		wg.Done()
	}()

	wg.Wait()
	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	var dataResult []OrgChart
	// set data
	for _, o := range org {
		if s, ok := listStaffId[o.StaffId]; ok {
			for _, or := range org {
				for _, val := range s {
					if val == or.StaffId && o.StaffId != or.StaffId {
						o.InvAmount += or.InvAmount
					}
				}
			}
			for _, in := range inv {
				for _, v := range s {
					if v == in.StaffID {
						o.InvAmountOld += in.InvAmount
					}
				}
			}
			for _, or := range org {
				for _, v := range s {
					if v == or.StaffId && o.StaffId != or.StaffId {
						o.Revenue += or.Revenue
					}
				}
			}
			for _, or := range org {
				for _, v := range s {
					if v == or.StaffId && o.StaffId != or.StaffId {
						o.EngCost += or.EngCost
					}
				}
			}
			for _, or := range org {
				for _, v := range s {
					if v == or.StaffId && o.StaffId != or.StaffId {
						o.TotalSo += or.TotalSo
					}
				}
			}
			for _, or := range org {
				for _, v := range s {
					if v == or.StaffId && o.StaffId != or.StaffId {
						o.IfFactor += or.IfFactor
					}
				}
			}
			if o.TotalSo != 0 {
				o.IfFactor = o.IfFactor / o.TotalSo
			}
			if o.EngCost != 0 {
				o.SaleFactor = o.Revenue / o.EngCost
			}
			if o.InvAmountOld != 0 {
				o.GrowthRate = ((o.InvAmount - o.InvAmountOld) / o.InvAmountOld) * 100
			}

			// filter data
			for _, fil := range filterOrg {
				if fil.StaffId == o.StaffId {
					dataResult = append(dataResult, o)
				}
			}
		}
	}

	fmt.Println(len(dataResult))
	// fmt.Println(quarterBeforeNum)
	// fmt.Println(yearNow)
	result.Data = dataResult
	result.Count = len(dataResult)
	result.Total = len(org)
	return c.JSON(http.StatusOK, result)
}

func CheckPermissionOrg(id string) (map[string][]string, error) {
	var user []model.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {
		var staff []model.StaffInfo

		mapStaff := map[string][]string{}
		if err := dbSale.Ctx().Raw(`SELECT staff_id,staff_child from staff_info where staff_id NOT IN (?);`, notSale).Scan(&staff).Error; err != nil {
			return nil, err
		}

		for _, s := range staff {
			var listStaffId []string
			if strings.TrimSpace(s.StaffChild) != "" {
				raw := strings.Split(s.StaffChild, ",")
				for _, id := range raw {
					listStaffId = append(listStaffId, id)
				}
				listStaffId = append(listStaffId, s.StaffId)
			} else {
				listStaffId = append(listStaffId, s.StaffId)
			}
			if _, ok := mapStaff[s.StaffId]; !ok {
				mapStaff[s.StaffId] = listStaffId
			}
		}
		return mapStaff, nil
	} else {
		var listStaffId []string
		mapStaff := map[string][]string{}
		staff := struct {
			StaffId    string `json:"staff_id"`
			StaffChild string `json:"staff_child"`
		}{}
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where staff_id = ?`, id).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
			return nil, nil
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
		if _, ok := mapStaff[id]; !ok {
			mapStaff[id] = listStaffId
		}

		return mapStaff, nil
	}
	// return nil, nil
}
