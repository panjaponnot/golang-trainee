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
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "connect database error")
	// 	return c.JSON(http.StatusInternalServerError, err)
	// }

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
	// fmt.Println("listStaffId ===>", listStaffId[strings.TrimSpace(c.QueryParam(("staff_id")))])

	// return c.JSON(http.StatusOK, listStaffId)
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
	// var defaultOrg []OrgChart
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
            WHERE INSTR(CONCAT_WS('|', fname,staff_id,lname,nname,department,all_ranking.one_id,position), ?) AND staff_id IN (?)
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
		if err := dbSale.Ctx().Raw(sqlDefault, filter, listStaffId[staffId]).Offset(p.Offset()).Limit(p.Size).Scan(&filterOrg).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
			}
		}
		wg.Done()
	}()
	// go func() {
	// 	if err := dbSale.Ctx().Raw(sqlDefault, "").Offset(p.Offset()).Limit(p.Size).Scan(&defaultOrg).Error; err != nil {
	// 		if !gorm.IsRecordNotFoundError(err) {
	// 			hasErr += 1
	// 		}
	// 	}
	// 	wg.Done()
	// }()
	wg.Wait()
	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	var dataResult []m.OrgChart
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
				for _, v := range s {
					if fil.StaffId == v && o.StaffId == v {
						dataResult = append(dataResult, o)
					}
				}
			}
		}

	}

	// fmt.Println(len(filterOrg))
	// fmt.Println(len(dataResult))
	// fmt.Println("listStaffId ===>", listStaffId)
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
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "init db error")
	// }

	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}

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

	log.Infoln(pkgName, year)
	log.Infoln(" query staff ")
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
		// if strings.TrimSpace(staff[0].Role) != "admin" {
		// 	listStaffId = append(listStaffId, staff[0].StaffId)
		// }
	}
	//////////////  getListStaffID  //////////////
	type PendingData struct {
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
	}
	var rawData []PendingData
	if err := dbSale.Ctx().Raw(`
	SELECT Active_Inactive,has_refer,tb_ch_so.sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,PeriodAmount, SUM(PeriodAmount) as TotalAmount,staff_id,prefix,fname,lname,nname,position,department,so_type,
        (case
                when status is null then 0
                else status end
        ) as status,
          (case
                when tb_expire.remark is null then ''
                else tb_expire.remark end
        ) as remark  from (
                SELECT *  from (
                SELECT  Active_Inactive,has_refer,sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,
                                DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,
                                                                PeriodAmount, SUM(PeriodAmount) as TotalAmount,
                                                                staff_id,prefix,fname,lname,nname,position,department
                                                                FROM ( SELECT * FROM so_mssql WHERE SOType NOT IN ('onetime' , 'project base') ) as s
                                                        left join
                                                        (
                                                                select staff_id, prefix, fname, lname, nname, position, department from staff_info

                                                        ) tb_sale on s.sale_code = tb_sale.staff_id
                                                        WHERE Active_Inactive = 'Active' and has_refer = 0 and staff_id IN (?) and year(ContractEndDate) = ?
                                                        group by sonumber
                        ) as tb_so_number
                        left join
                        (
                         select 
                                (case
                                        when pay_type is null then ''
                                        else pay_type end
                                ) as pay_type,
                                sonumber as so_check,
                                (case
                                        when so_type is null then ''
                                        else so_type end
                                ) as so_type 
                        from check_so
                        ) tb_check on tb_so_number.sonumber = tb_check.so_check

                ) as tb_ch_so
                left join
                (
                  select id,sonumber,
                        (case
                                when status is null then 0
                                else status end
                        ) as status,
                        (case
                                when remark is null then ''
                                else remark end
                        ) as remark 
                        from check_expire
                  ) tb_expire on tb_ch_so.sonumber = tb_expire.sonumber
                  group by tb_ch_so.sonumber
		  `, listStaffId, year).Scan(&rawData).Error; err != nil {
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
	}{}
	if len(user) != 0 {

		if err := dbSale.Ctx().Raw(`SELECT * FROM (SELECT check_so.remark_sale as remark,check_so.status_sale,check_so.status_so as status_so,check_so.sonumber,
			Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days , 'so' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_so,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,pricesale
			from so_mssql
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE check_so.status_sale = 0 and check_so.remark_sale <> ''
		group by sonumber
		union
		SELECT check_so.remark_sale as remark,check_so.status_sale,check_so.status_so as status_so,check_so.sonumber,Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days , 'so' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_so,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,PeriodAmount as pricesale
			from so_mssql_navision
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber not like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE check_so.status_sale = 0 and check_so.remark_sale <> ''
		group by sonumber order by status_sale) as data`).Scan(&rawData).Error; err != nil {
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

		if err := dbSale.Ctx().Raw(`SELECT * FROM (SELECT check_so.status_so as status_so,check_so.status_sale as status_sale,so_mssql.sonumber,Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days ,check_so.remark_sale as remark,'sale' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_sale,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,pricesale
			from so_mssql
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE sale_code in (?)
			group by sonumber
		union
		SELECT check_so.status_so as status_so,check_so.status_sale as status_sale,so_mssql.sonumber,Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days ,check_so.remark_sale as remark,'sale' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_sale,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,PeriodAmount as pricesale
			from so_mssql_navision
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber not like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE sale_code in (?)
			group by sonumber order by status_sale ) as data
			`, listStaffId, listStaffId).Scan(&rawData).Error; err != nil {
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
