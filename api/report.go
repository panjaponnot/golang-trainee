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

func GetDataOrgChartEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id"})
	}
	staffId := strings.TrimSpace(c.QueryParam(("staff_id")))
	filter := strings.TrimSpace(c.QueryParam(("filter")))
	listStaffId, err := CheckPermissionOrg(staffId)
	if err != nil {
		log.Errorln(pkgName, err, "func check permission error :-")
		return c.JSON(http.StatusInternalServerError, m.Result{Error: "check permission error"})
	}
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

	var org []m.OrgChart
	var filterOrg []m.OrgChart
	var inv []m.InvBefore
	var result m.Result
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
            			WHEN total_contract is null THEN 0
            			ELSE total_contract END
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
            			WHEN so_number is null THEN 0
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
            			select total_contract,so_number,tb_cus.sale_cus_id as sale_id,sale_factor,in_factor,(total_contract/sale_factor) as eng_cost
            			from so_info
						left join
						(
							select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

						) tb_cus on so_info.customer_id = tb_cus.customer_id
            			WHERE quarter(contract_start_date) = quarter(now()) and year(contract_start_date) = year(now()) and active_inactive = 1
            			group by so_number
            		) total_so on total_so.sale_id = staff_detail.staff_id
            		group by staff_id
            	) tb_main
            	LEFT join (
            		select sum(total_contract_per_month) as inv_amount,sale_id from (
            			select total_contract_per_month,tb_cus.sale_cus_id as sale_id
            			from so_info
						left join
						(
							select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

						) tb_cus on so_info.customer_id = tb_cus.customer_id
            			WHERE quarter(contract_start_date) = quarter(now()) and year(contract_start_date) = year(now()) and so_refer = '' and active_inactive = 1
            			group by so_number
            		) tb_inv group by sale_id
            	) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_id
            	where staff_id is not null and staff_id <> ''
			) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
			group by staff_id`

	sqlDefault := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, 0 as inv_amount_old, -5 as score_target,
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
            			WHEN total_contract is null THEN 0
            			ELSE total_contract END
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
            			WHEN so_number is null THEN 0
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
            			select total_contract,so_number,tb_cus.sale_cus_id as sale_id,sale_factor,in_factor,(total_contract/sale_factor) as eng_cost
            			from so_info
						left join
						(
							select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

						) tb_cus on so_info.customer_id = tb_cus.customer_id
            			WHERE quarter(contract_start_date) = quarter(now()) and year(contract_start_date) = year(now()) and active_inactive = 1
            			group by so_number
            		) total_so on total_so.sale_id = staff_detail.staff_id
            		group by staff_id
            	) tb_main
            	LEFT join (
            		select sum(total_contract_per_month) as inv_amount,sale_id from (
            			select total_contract_per_month,tb_cus.sale_cus_id as sale_id
            			from so_info
						left join
						(
							select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

						) tb_cus on so_info.customer_id = tb_cus.customer_id
            			WHERE quarter(contract_start_date) = quarter(now()) and year(contract_start_date) = year(now()) and so_refer = '' and active_inactive = 1
            			group by so_number
            		) tb_inv group by sale_id
            	) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_id
            	where staff_id is not null and staff_id <> ''
			) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
            WHERE INSTR(CONCAT_WS('|', fname,staff_id,lname,nname,department,all_ranking.one_id,position), ?) AND staff_id IN (?)
			group by staff_id`
	sqlInv := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
                from (
                	select staff_id,sum(total_contract_per_month) as inv_amount,count(so_number) as total_so
                	from (
                		select staff_id from staff_info
                		left join
                		(
                			select * from goal_quarter where year = ? and quarter = ?
                		) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
                		group by staff_id
                	) staff_detail
                	LEFT JOIN (
                		select total_contract_per_month,sale_id,so_number, type_sale
                        from (
                        	select total_contract_per_month,tb_cus.sale_cus_id as sale_id,so_number , 'normal' as type_sale
                        	from so_info
							left join
						(
							select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

						) tb_cus on so_info.customer_id = tb_cus.customer_id
                        	WHERE quarter(contract_start_date) = ? and year(contract_start_date) = ? and so_refer = '' and active_inactive = 1
                        	group by so_number
                        ) tb_inv_old
                	) total_new_so on total_new_so.sale_id = staff_detail.staff_id
                	where staff_id is not null and staff_id <> '' and sale_id is not null
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
		if err := dbSale.Ctx().Raw(sqlDefault, filter, listStaffId[staffId]).Offset(p.Offset()).Limit(p.Size).Scan(&filterOrg).Error; err != nil {
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

	var dataResult []m.OrgChart
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
				for _, v := range s {
					if fil.StaffId == v && o.StaffId == v {
						dataResult = append(dataResult, o)
					}
				}
			}
		}

	}
	result.Data = dataResult
	result.Count = len(dataResult)
	result.Total = len(org)
	return c.JSON(http.StatusOK, result)
}

func CheckPermissionOrg(id string) (map[string][]string, error) {
	var user []m.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {

		mapStaff := map[string][]string{}
		var staff []m.StaffInfo
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
		var staffAll []m.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT staff_id,staff_child from staff_info where staff_id NOT IN (?);`, notSale).Scan(&staffAll).Error; err != nil {
			return nil, err
		}

		mapStaff := map[string][]string{}
		staff := struct {
			StaffId    string `json:"staff_id"`
			StaffChild string `json:"staff_child"`
		}{}
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where staff_id = ?`, id).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
			return nil, nil
		}
		var rawdata []string
		if strings.TrimSpace(staff.StaffChild) != "" {
			raw := strings.Split(staff.StaffChild, ",")
			for _, id := range raw {
				rawdata = append(rawdata, id)
			}
			rawdata = append(rawdata, staff.StaffId)
		} else {
			rawdata = append(rawdata, staff.StaffId)
		}

		for _, v := range staffAll {
			for _, c := range rawdata {
				if v.StaffId == c {
					var listStaffId []string
					if strings.TrimSpace(v.StaffChild) != "" {
						raw := strings.Split(v.StaffChild, ",")
						for _, id := range raw {
							listStaffId = append(listStaffId, id)
						}
						listStaffId = append(listStaffId, v.StaffId)
					} else {
						listStaffId = append(listStaffId, v.StaffId)
					}
					if _, ok := mapStaff[v.StaffId]; !ok {
						mapStaff[v.StaffId] = listStaffId
					}
				}
			}
		}

		return mapStaff, nil
	}
	// return nil, nil
}

func GetReportSOPendingEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}

	search := strings.TrimSpace(c.QueryParam("search"))
	// searchSO := strings.TrimSpace(c.QueryParam("searchSO"))
	oneId := strings.TrimSpace(c.QueryParam("one_id"))
	// StaffId := strings.TrimSpace(c.QueryParam("staff_id"))
	year := strings.TrimSpace(c.QueryParam("year"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("year")), 10); err == nil {
			yearDefault = time.Unix(util.ConvertTimeStamp(f), 0)
		}
		years, _, _ := yearDefault.Date()
		year = strconv.Itoa(years)
	}

	// log.Infoln(pkgName, year)
	// log.Infoln(" query staff ")
	staff := []struct {
		StaffId    string `json:"staff_id"`
		Role       string `json:"role"`
		StaffChild string `json:"staff_child"`
	}{}
	if err := dbSale.Ctx().Raw(` SELECT staff_id, role, "" as staff_child from user_info where role = "admin" and one_id = ?
	union
	SELECT staff_id, "normal" as role, staff_child from staff_info where one_id = ? `, oneId, oneId).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select staff error")
		return echo.ErrInternalServerError
	}
	staffs := []struct {
		StaffId    string `json:"staff_id"`
		StaffChild string `json:"staff_child"`
	}{}
	var listStaffId []string
	if len(staff) != 0 {
		for _, v := range staff {
			log.Infoln(pkgName, v.Role)
			if strings.TrimSpace(v.Role) == "admin" {
				if err := dbSale.Ctx().Raw(`select staff_id from staff_info;`).Scan(&staffs).Error; err != nil {
					log.Errorln(pkgName, err, "Select data error")
				}
				if len(staffs) != 0 {
					for _, id := range staffs {
						listStaffId = append(listStaffId, id.StaffId)
					}
					break
				}
			} else {
				if strings.TrimSpace(v.StaffChild) != "" {
					listStaffId = strings.Split(v.StaffChild, ",")
				}
				listStaffId = append(listStaffId, staff[0].StaffId)
			}
		}
	}
	//////////////  getListStaffID  //////////////

	var rawData []PendingData
	if err := dbSale.Ctx().Raw(`
				SELECT active_inactive,has_refer,s.so_number,customer_id,customer_nameTH,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,
						DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,
						so_refer,sale_id,DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, so_web_status,total_contract,total_contract_per_month,
						TotalAmount,staff_id,prefix,fname,lname,nname,position,department, so_type,so_type_change,pay_type_change,
								(case
								when status is null then 0
							else status end
								) as status,
						(case
							when reason is null then ''
							else reason end
						) as reason,
								(case
										when remark is null then ''
							else remark end
						) as remark
								FROM
					(
						SELECT *  from (
							SELECT  active_inactive,has_refer,so_number,s.customer_id,customer_nameTH,DATE_FORMAT(contract_start_date, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(contract_end_date, '%Y-%m-%d') as ContractEndDate,so_refer,sale_cus_id as sale_id,
											DATEDIFF(contract_end_date, NOW()) as days, month(contract_end_date) as so_month, so_web_status,total_contract,
													total_contract_per_month, SUM(total_contract_per_month) as TotalAmount,
													staff_id,prefix,fname,lname,nname,position,department, so_type, pay_type
													FROM ( SELECT * FROM so_info WHERE so_type NOT IN ('onetime' , 'project base') ) as s
											left join
											(
													select staff_id, prefix, fname, lname, nname, position, department from staff_info
			
											) tb_sale on s.sale_id = tb_sale.staff_id
											left join
											(
													select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info
			
											) tb_cus on s.customer_id = tb_cus.customer_id
											WHERE active_inactive = 1 
												and has_refer = 0 
												and staff_id IN (?) 
												and year(contract_end_date) = ?
											group by so_number
									) as tb_so_number
					) s
					left join
					(
						select
							id,sonumber as so_number,
							(case
								when status is null then 0
								else status end
									) as status,
							(case
								when reason is null then ''
											else reason end
									) as reason,
							(case
								when remark is null then ''
								else remark end
							) as remark,
							pay_type as pay_type_change,
							so_type as so_type_change
						from check_expire
					) tb_expire on s.so_number = tb_expire.so_number
			Where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,customer_id,customer_nameTH,s.so_number), ?)
		  `, listStaffId, year, search).Scan(&rawData).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
	}

	fmt.Println("   ==== ", len(rawData))
	mapData := map[string][]PendingData{}

	for _, v := range rawData {
		mapData[v.SoMonth] = append(mapData[v.SoMonth], v)
	}
	var result m.Result
	result.Data = mapData
	result.Total = len(rawData)
	return c.JSON(http.StatusOK, result)
}

func GetReportSOPendingTypeEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}

	search := strings.TrimSpace(c.QueryParam("search"))
	searchSOType := strings.TrimSpace(c.QueryParam("sotype"))
	oneId := strings.TrimSpace(c.QueryParam("one_id"))
	year := strings.TrimSpace(c.QueryParam("year"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("year")), 10); err == nil {
			yearDefault = time.Unix(util.ConvertTimeStamp(f), 0)
		}
		years, _, _ := yearDefault.Date()
		year = strconv.Itoa(years)
	}

	// log.Infoln(pkgName, year)
	// log.Infoln(" query staff ")
	staff := []struct {
		StaffId    string `json:"staff_id"`
		Role       string `json:"role"`
		StaffChild string `json:"staff_child"`
	}{}
	if err := dbSale.Ctx().Raw(` SELECT staff_id, role, "" as staff_child from user_info where role = "admin" and one_id = ?
	union
	SELECT staff_id, "normal" as role, staff_child from staff_info where one_id = ? `, oneId, oneId).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select staff error")
		return echo.ErrInternalServerError
	}
	staffs := []struct {
		StaffId    string `json:"staff_id"`
		StaffChild string `json:"staff_child"`
	}{}
	var listStaffId []string
	if len(staff) != 0 {
		for _, v := range staff {
			log.Infoln(pkgName, v.Role)
			if strings.TrimSpace(v.Role) == "admin" {
				if err := dbSale.Ctx().Raw(`select staff_id from staff_info;`).Scan(&staffs).Error; err != nil {
					log.Errorln(pkgName, err, "Select data error")
				}
				if len(staffs) != 0 {
					for _, id := range staffs {
						listStaffId = append(listStaffId, id.StaffId)
					}
				}
			} else {
				if strings.TrimSpace(v.StaffChild) != "" {
					listStaffId = strings.Split(v.StaffChild, ",")
				}
				listStaffId = append(listStaffId, staff[0].StaffId)
			}
		}
	}
	//////////////  getListStaffID  //////////////

	// FROM ( SELECT * FROM so_mssql WHERE SOType NOT IN ('onetime' , 'project base') ) as s
	var rawData []PendingData
	if err := dbSale.Ctx().Raw(`
	SELECT active_inactive,has_refer,s.so_number,customer_id,customer_nameTH,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,
	DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,
	so_refer,sale_id,DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, so_web_status,total_contract,total_contract_per_month,
	TotalAmount,staff_id,prefix,fname,lname,nname,position,department, so_type,so_type_change,pay_type_change,
			(case
			when status is null then 0
		else status end
			) as status,
	(case
		when reason is null then ''
		else reason end
	) as reason,
			(case
					when remark is null then ''
		else remark end
	) as remark
			FROM
(
	SELECT *  from (
		SELECT  active_inactive,has_refer,so_number,s.customer_id,customer_nameTH,DATE_FORMAT(contract_start_date, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(contract_end_date, '%Y-%m-%d') as ContractEndDate,so_refer,sale_cus_id as sale_id,
						DATEDIFF(contract_end_date, NOW()) as days, month(contract_end_date) as so_month, so_web_status,total_contract,
								total_contract_per_month, SUM(total_contract_per_month) as TotalAmount,
								staff_id,prefix,fname,lname,nname,position,department, so_type, pay_type
								FROM ( SELECT * FROM so_info WHERE so_type = ? ) as s
						left join
						(
								select staff_id, prefix, fname, lname, nname, position, department from staff_info

						) tb_sale on s.sale_id = tb_sale.staff_id
						left join
						(
								select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

						) tb_cus on s.customer_id = tb_cus.customer_id
						WHERE active_inactive = 1 
							and has_refer = 0 
							and staff_id IN (?) 
							and year(contract_end_date) = ?
						group by so_number
				) as tb_so_number
) s
left join
(
	select
		id,sonumber as so_number,
		(case
			when status is null then 0
			else status end
				) as status,
		(case
			when reason is null then ''
						else reason end
				) as reason,
		(case
			when remark is null then ''
			else remark end
		) as remark,
		pay_type as pay_type_change,
		so_type as so_type_change
	from check_expire
) tb_expire on s.so_number = tb_expire.so_number
Where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,customer_id,customer_nameTH,s.so_number), ?)
		  `, searchSOType, listStaffId, year, search).Scan(&rawData).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
	}

	fmt.Println("   ==== ", len(rawData))
	mapData := map[string][]PendingData{}

	for _, v := range rawData {
		mapData[v.SoMonth] = append(mapData[v.SoMonth], v)
	}
	var result m.Result
	result.Data = mapData
	result.Total = len(rawData)
	return c.JSON(http.StatusOK, result)
}

func GetReportSOEndPoint(c echo.Context) error {
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}
	oneId := strings.TrimSpace(c.QueryParam("one_id"))
	var user []m.UserInfo
	if err := dbSale.Ctx().Raw(` SELECT * FROM user_info WHERE role = 'admin' AND one_id = ? `, oneId).Scan(&user).Error; err != nil {
		log.Errorln(pkgName, err, "User Not Found")
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorln(pkgName, err, "Select user Error")
			return echo.ErrInternalServerError
		}
	}
	rawData := []struct {
		SOnumber          string  `json:"so_number" gorm:"column:so_number"`
		CustomerId        string  `json:"customer_id" gorm:"column:customer_id"`
		CustomerName      string  `json:"customer_name" gorm:"column:customer_nameTH"`
		ContractStartDate string  `json:"contract_start_date" gorm:"column:contract_start_date"`
		ContractEndDate   string  `json:"contract_end_date" gorm:"column:contract_end_date"`
		SORefer           string  `json:"so_refer" gorm:"column:so_refer"`
		SaleCode          string  `json:"sale_code" gorm:"column:sale_id"`
		SaleLead          string  `json:"sale_lead" gorm:"column:sale_lead"`
		Day               string  `json:"day" gorm:"column:days"`
		SoMonth           string  `json:"so_month" gorm:"column:so_month"`
		SOWebStatus       string  `json:"so_web_status" gorm:"column:so_web_status"`
		PriceSale         float64 `json:"price_sale" gorm:"column:price_sale"`
		PeriodAmount      float64 `json:"period_amount" gorm:"column:total_contract_per_month"`
		TotalAmount       float64 `json:"total_amount" gorm:"column:TotalAmount"`
		StaffId           string  `json:"staff_id" gorm:"column:staff_id"`
		PayType           string  `json:"pay_type" gorm:"column:pay_type"`
		SoType            string  `json:"so_type" gorm:"column:so_type"`
		Prefix            string  `json:"prefix"`
		Fname             string  `json:"fname"`
		Lname             string  `json:"lname"`
		Nname             string  `json:"nname"`
		Position          string  `json:"position"`
		Department        string  `json:"department"`
		Status            string  `json:"status"`
		Remark            string  `json:"remark"`
		StatusSO          bool    `json:"status_so" gorm:"column:status_so"`
		StatusSale        bool    `json:"status_sale" gorm:"column:status_sale"`
	}{}
	if len(user) != 0 {

		if err := dbSale.Ctx().Raw(`SELECT * FROM (SELECT check_so.status_so as status_so,check_so.status_sale as status_sale,so_info.so_number,tb_cus.customer_id,customer_nameTH,one_id, contract_start_date,contract_end_date,
			so_refer,sale_cus_id as sale_id,total_contract_per_month,so_type,pay_type,
			in_factor,sale_factor,
			SUM(total_contract_per_month) as TotalAmount_old,
			IFNULL(prefix, '') as prefix,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,	total_contract as TotalAmount,so_web_status,total_contract,
			datediff(contract_end_date,contract_start_date) as days ,check_so.remark_sale as remark,'sale' as role,
			TIMESTAMPDIFF(month,contract_start_date,DATE_ADD(contract_end_date, INTERVAL 3 DAY)) as months
	FROM (
			select so_number,customer_id,contract_start_date,contract_end_date,so_refer,sale_id,total_contract_per_month,
			in_factor,sale_factor,(	total_contract/1.07) as total_contract,
			so_web_status
			from so_info
			where has_refer = 0 and active_inactive = 1 and so_number like '%SO%' and so_type <> 'Onetime' and so_type <> 'Project Base'
	) so_info
	left join check_so on check_so.sonumber = so_info.so_number
	LEFT JOIN staff_info ON so_info.sale_id = staff_info.staff_id
	 left join
					(
							select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

					) tb_cus on so_info.customer_id = tb_cus.customer_id
					WHERE check_so.status_sale = 0 and check_so.remark_sale <> ''
			group by sonumber) as data`).Scan(&rawData).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
		}

	} else {
		var listStaffId []string
		staff := struct {
			StaffId    string `json:"staff_id"`
			StaffChild string `json:"staff_child"`
		}{}
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where one_id = ?`, oneId).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
		}

		if strings.TrimSpace(staff.StaffChild) != "" {
			raw := strings.Split(staff.StaffChild, ",")
			for _, id := range raw {
				listStaffId = append(listStaffId, id)
			}
			listStaffId = append(listStaffId, staff.StaffId)
			log.Infoln(pkgName, "has team", staff)
		} else {
			listStaffId = append(listStaffId, staff.StaffId)
			log.Infoln(pkgName, "not found team", staff)
		}

		if err := dbSale.Ctx().Raw(`SELECT * FROM (SELECT check_so.status_so as status_so,check_so.status_sale as status_sale,so_info.so_number,tb_cus.customer_id,customer_nameTH,one_id, contract_start_date,contract_end_date,
			so_refer,sale_cus_id as sale_id,total_contract_per_month,so_type,pay_type,
			in_factor,sale_factor,
			SUM(total_contract_per_month) as TotalAmount_old,
			IFNULL(prefix, '') as prefix,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,	total_contract as TotalAmount,so_web_status,total_contract,
			datediff(contract_end_date,contract_start_date) as days ,check_so.remark_sale as remark,'sale' as role,
			TIMESTAMPDIFF(month,contract_start_date,DATE_ADD(contract_end_date, INTERVAL 3 DAY)) as months
	FROM (
			select so_number,customer_id,contract_start_date,contract_end_date,so_refer,sale_id,total_contract_per_month,
			in_factor,sale_factor,(	total_contract/1.07) as total_contract,
			so_web_status
			from so_info
			where has_refer = 0 and active_inactive = 1 and so_number like '%SO%' and so_type <> 'Onetime' and so_type <> 'Project Base'
	) so_info
	left join check_so on check_so.sonumber = so_info.so_number
	LEFT JOIN staff_info ON so_info.sale_id = staff_info.staff_id
	 left join
					(
							select sale_id as sale_cus_id,customer_id,customer_nameTH from customer_info

					) tb_cus on so_info.customer_id = tb_cus.customer_id
	WHERE sale_id in (?)
			group by sonumber) as data
			`, listStaffId).Scan(&rawData).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
		}
	}

	return c.JSON(http.StatusOK, rawData)
}

func EditSOEndPoint(c echo.Context) error {
	type SoData struct {
		SOnumber          string  `json:"so_number" gorm:"column:sonumber"`
		CustomerId        string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName      string  `json:"customer_name" gorm:"column:Customer_Name"`
		ContractStartDate string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate   string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SORefer           string  `json:"so_refer" gorm:"column:so_refer"`
		SaleCode          string  `json:"sale_code" gorm:"column:sale_code"`
		SaleLead          string  `json:"sale_lead" gorm:"column:sale_lead"`
		Day               string  `json:"day" gorm:"column:days"`
		SoMonth           string  `json:"so_month" gorm:"column:so_month"`
		SOWebStatus       string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		PriceSale         float64 `json:"price_sale" gorm:"column:pricesale"`
		PeriodAmount      float64 `json:"period_amount" gorm:"column:PeriodAmount"`
		TotalAmount       float64 `json:"total_amount" gorm:"column:TotalAmount"`
		StaffId           string  `json:"staff_id" gorm:"column:staff_id"`
		PayType           string  `json:"pay_type" gorm:"column:pay_type"`
		SoType            string  `json:"so_type" gorm:"column:so_type"`
		Prefix            string  `json:"prefix"`
		Fname             string  `json:"fname"`
		Lname             string  `json:"lname"`
		Nname             string  `json:"nname"`
		Position          string  `json:"position"`
		Department        string  `json:"department"`
		Status            string  `json:"status"`
		Remark            string  `json:"remark"`
		StatusSO          bool    `json:"status_so" gorm:"column:status_so"`
		StatusSale        bool    `json:"status_sale" gorm:"column:status_sale"`
	}
	body := struct {
		Data  []SoData `json:"data"`
		OneId string   `json:"one_id"`
	}{}

	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid Body data"})
	}

	if strings.TrimSpace(body.OneId) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid one id"})
	}
	id := strings.TrimSpace(body.OneId)
	var user []m.UserInfo
	if err := dbSale.Ctx().Raw(`SELECT * FROM user_info WHERE role = 'admin' AND one_id = ?`, id).Scan(&user).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
	}

	if len(user) != 0 {
		for _, v := range body.Data {
			remark := CheckRemark(v.Remark, "so")
			statusMy := 0
			sqlMssql := `	UPDATE Check_SO
						SET  status_so = ?,
						Remark = ?,
						press_date = CURRENT_TIMESTAMP
						WHERE so_number = ?;`

			sqlSale := `Update check_so
					SET
					status_so = ?,
					update_by_so = ?,
					remark_so = ?
					WHERE sonumber = ?;`
			if v.StatusSO {
				statusMy = 1
				v.Status = "true"
			} else if !v.StatusSO && strings.TrimSpace(remark) != "" {
				statusMy = 0
				v.Status = "fail"
			} else {
				statusMy = 0
				v.Status = ""
			}
			if err := dbMssql.Ctx().Exec(sqlMssql, v.Status, strings.TrimSpace(remark), v.SOnumber).Error; err != nil {
				return echo.ErrInternalServerError
			}
			if err := dbSale.Ctx().Exec(sqlSale, statusMy, id, strings.TrimSpace(remark), v.SOnumber).Error; err != nil {
				return echo.ErrInternalServerError
			}
		}

	} else {
		sqlMssql := `	UPDATE Check_SO
						SET  status_sale = ?,
						Remark = ?,
						soType = ?,
						payType = ?,
						press_date = CURRENT_TIMESTAMP
						WHERE so_number = ?;`

		sqlSale := `Update check_so
					SET
					status_sale = ?,
					so_type = ?,
					pay_type = ?,
					update_by_sale  = ?,
					remark_sale = ?
					WHERE sonumber = ?;`

		for _, v := range body.Data {
			remark := CheckRemark(v.Remark, "sale")
			statusMy := 0
			if v.StatusSale {
				statusMy = 1
				v.Status = "true"
			} else if !v.StatusSale && strings.TrimSpace(remark) != "" {
				statusMy = 0
				v.Status = "fail"
			} else {
				statusMy = 0
				v.Status = ""
			}
			if err := dbMssql.Ctx().Exec(sqlMssql, v.Status, strings.TrimSpace(remark), v.SoType, v.PayType, v.SOnumber).Error; err != nil {
				return echo.ErrInternalServerError
			}
			if err := dbSale.Ctx().Exec(sqlSale, statusMy, v.SoType, v.PayType, id, strings.TrimSpace(remark), v.SOnumber).Error; err != nil {
				return echo.ErrInternalServerError
			}
		}
	}
	return c.JSON(http.StatusNoContent, nil)
}
func CheckRemark(remark string, types string) string {

	if types == "so" {
		check := strings.Split(remark, "SO comment: ")
		if len(check) == 1 {
			return check[0]
		}
		checkSa := strings.Split(check[1], "Sale comment: ")
		return checkSa[0]
	} else {
		checkSa := strings.Split(remark, "Sale comment: ")
		if len(checkSa) > 1 {
			return checkSa[1]
		}
		check := strings.Split(checkSa[0], "SO comment: ")
		return check[0]
	}

	return ""

}

func UpdateSOEndPoint(c echo.Context) error {
	type SoData struct {
		SOnumber string `json:"sonumber" gorm:"column:sonumber"`
		Status   string `json:"status"`
		Remark   string `json:"remark"`
	}
	body := struct {
		Data  []SoData `json:"data"`
		OneId string   `json:"one_id"`
	}{}

	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid Body data"})
	}

	if strings.TrimSpace(body.OneId) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid one id"})
	}
	id := strings.TrimSpace(body.OneId)
	var CheckExpire []m.CheckExpire
	if err := dbSale.Ctx().Raw(`select * from check_expire;`).Scan(&CheckExpire).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
	}

	type SoAll struct {
		SOnumber string `json:"sonumber" gorm:"column:sonumber"`
	}
	var SoAllData []SoAll
	var SoAllStr string
	if len(CheckExpire) > 0 {
		for _, c := range CheckExpire {
			data := SoAll{
				SOnumber: c.SOnumber,
			}
			SoAllStr += c.SOnumber
			SoAllData = append(SoAllData, data)
		}
	}
	var ListData []SoData
	for _, d := range body.Data {
		ListData = append(ListData, d)
	}
	var ValuesUpdate []m.CheckExpire
	var ValuesInsert []m.CheckExpire

	for _, d := range ListData {
		if strings.Contains(SoAllStr, "BOI-20190101-0026") {
			value := m.CheckExpire{
				SOnumber: d.SOnumber,
				Status:   d.Status,
				Remark:   d.Remark,
				CreateBy: id,
			}
			ValuesUpdate = append(ValuesUpdate, value)
		} else {
			value := m.CheckExpire{
				SOnumber: d.SOnumber,
				Status:   d.Status,
				Remark:   d.Remark,
				CreateBy: id,
			}
			ValuesInsert = append(ValuesInsert, value)
		}
	}

	sqlinsert := `insert into check_expire(sonumber,status,remark,create_by) values(?,?,?,?);`
	sqlupdate := `update check_expire set status = ?, remark = ?, create_by = ? where sonumber = ?;`

	if len(ValuesUpdate) > 0 {
		for _, v := range ValuesUpdate {
			if err := dbSale.Ctx().Exec(sqlupdate, v.Status, v.Remark, v.CreateBy, v.SOnumber).Error; err != nil {
				return echo.ErrInternalServerError
			}
			//Log
			if err := dbSale.Ctx().Model(&m.CheckExpire{}).Create(&v).Error; err != nil {
				log.Errorln(pkgName, err, "create CheckExpire log error :-")
				return c.JSON(http.StatusInternalServerError, server.Result{Message: "create CheckExpire log error"})
			}
		}
	}

	if len(ValuesInsert) > 0 {
		for _, v := range ValuesInsert {
			if err := dbSale.Ctx().Exec(sqlinsert, v.SOnumber, v.Status, v.Remark, v.CreateBy).Error; err != nil {
				return echo.ErrInternalServerError
			}
			//Log
			if err := dbSale.Ctx().Model(&m.CheckExpire{}).Create(&v).Error; err != nil {
				log.Errorln(pkgName, err, "create CheckExpire log error :-")
				return c.JSON(http.StatusInternalServerError, server.Result{Message: "create CheckExpire log error"})
			}
		}
	}

	return c.JSON(http.StatusNoContent, nil)

}

type PendingData struct {
	SOnumber          string  `json:"so_number" gorm:"column:so_number"`
	CustomerId        string  `json:"customer_id" gorm:"column:customer_id"`
	CustomerName      string  `json:"customer_name" gorm:"column:customer_nameTH"`
	ContractStartDate string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
	ContractEndDate   string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
	SORefer           string  `json:"so_refer" gorm:"column:so_refer"`
	SaleCode          string  `json:"sale_code" gorm:"column:sale_id"`
	Day               string  `json:"day" gorm:"column:days"`
	SoMonth           string  `json:"so_month" gorm:"column:so_month"`
	SOWebStatus       string  `json:"so_web_status" gorm:"column:so_web_status"`
	PriceSale         float64 `json:"price_sale" gorm:"column:total_contract"`
	PeriodAmount      float64 `json:"period_amount" gorm:"column:total_contract_per_month"`
	TotalAmount       float64 `json:"total_amount" gorm:"column:TotalAmount"`
	StaffId           string  `json:"staff_id" gorm:"column:staff_id"`
	PayType           string  `json:"pay_type" gorm:"column:pay_type"`
	SoType            string  `json:"so_type" gorm:"column:so_type"`
	Prefix            string  `json:"prefix"`
	Fname             string  `json:"fname"`
	Lname             string  `json:"lname"`
	Nname             string  `json:"nname"`
	Position          string  `json:"position"`
	Department        string  `json:"department"`
	Status            string  `json:"status"`
	PayTypeChange     string  `json:"pay_type_change"`
	SoTypeChange      string  `json:"so_type_change"`
	Reason            string  `json:"reason"`
	Remark            string  `json:"remark"`
}
