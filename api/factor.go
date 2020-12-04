package api

import (
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

type SaleFactor struct {
	TotalRevenue float64 `json:"total_revenue" gorm:"column:total_revenue"`
	CountSo      int     `json:"count_so" gorm:"column:count_so"`
	InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
	ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
	EngCost      float64 `json:"eng_cost" gorm:"column:engcost"`
	RealSF       float64 `json:"real_sf" gorm:"column:real_sf"`
	Department   string  `json:"department" gorm:"column:department"`
}

type CountSoPerson struct {
	CountSo    int    `json:"count_so" gorm:"column:count_so"`
	Department string `json:"department" gorm:"column:department"`
	Fname      string `json:"fname" gorm:"column:fname"`
	Lname      string `json:"lname" gorm:"column:lname"`
	StaffId    string `json:"staff_id" gorm:"column:staff_id"`
}

func GetSummarySaleFactorEndPoint(c echo.Context) error {
	accountId := strings.TrimSpace(c.Param("id"))
	check := checkPermissionUser(accountId)
	if !check {
		return echo.ErrNotFound
	}
	today := time.Now()
	year, month, _ := today.Date()
	sql := `select 
				sum(revenue) as total_revenue,
				sum(engcost) as engcost,
				sum(revenue)/sum(engcost) as real_sf,
				department
			from (
				Select 
					TotalContractAmount as revenue,
					(CASE
						WHEN TotalContractAmount is not null and sale_factor is not null and sale_factor != 0 THEN TotalContractAmount/sale_factor
						ELSE 0 END
					) as engcost,
					sale_factor,
					sale_code
				from so_mssql where month(PeriodStartDate) = ? and year(PeriodStartDate) = ?
				group by sonumber
			) tb_so
			LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
			where department in (
				SELECT department FROM staff_info WHERE staff_child = '' and department <> 'Sale JV' GROUP BY department
			)
			GROUP BY department ORDER BY real_sf desc  `
	saleFac := []struct {
		TotalRevenue float64 `json:"total_revenue" gorm:"column:total_revenue"`
		EngCost      float64 `json:"eng_cost" gorm:"column:engcost"`
		RealSF       float64 `json:"real_sf" gorm:"column:real_sf"`
		Department   string  `json:"department" gorm:"column:department"`
	}{}
	if err := dbSale.Ctx().Raw(sql, month, year).Scan(&saleFac).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}

	sumRevenue := 0.0
	sumEngCost := 0.0
	sumSF := 0.0
	for _, v := range saleFac {
		sumRevenue += v.TotalRevenue
		sumEngCost += v.EngCost
		sumSF += v.RealSF
	}

	var dataRaw interface{}
	if len(saleFac) < 5 {
		dataRaw = saleFac[0:]
	} else {
		dataRaw = saleFac[0:5]
	}

	dataResult := map[string]interface{}{
		"data":              dataRaw,
		"sale_factor_total": sumRevenue / sumEngCost,
	}
	return c.JSON(http.StatusOK, dataResult)
}

func GetSummaryInternalFactorAndExternalFactorEndPoint(c echo.Context) error {
	accountId := strings.TrimSpace(c.Param("id"))
	check := checkPermissionUser(accountId)
	if !check {
		return echo.ErrNotFound
	}
	today := time.Now()
	year, month, _ := today.Date()
	sql := `SELECT
			sum(in_factor) as in_factor,
			sum(ex_factor) as ex_factor,
			sum(revenue) as total_revenue,
			sum(engcost) as engcost,
			sum(revenue)/sum(engcost) as real_sf,
			department
		from (
			Select 
				TotalContractAmount as revenue,
				(CASE
					WHEN TotalContractAmount is not null and sale_factor is not null and sale_factor != 0 THEN TotalContractAmount/sale_factor
					ELSE 0 END
				) as engcost,
				sale_factor,
				sale_code,
				in_factor,
				sonumber,
				ex_factor
			from so_mssql where month(PeriodStartDate) = ? and year(PeriodStartDate) = ?
			group by sonumber
		) tb_so
		LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
		where department in (
			SELECT department FROM staff_info WHERE staff_child = '' and department <> 'Sale JV' GROUP BY department
		)
		GROUP BY department ORDER BY real_sf desc  `

	countCompany := `SELECT COUNT(Customer_ID) as count_so , department
					from (
						Select 
								Customer_ID, Customer_name, sale_code
						from so_mssql where month(PeriodStartDate) = ? and year(PeriodStartDate) = ?
						group by Customer_ID
					) tb_so
					LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
					where department in (
						SELECT department FROM staff_info WHERE staff_child = '' and department <> 'Sale JV' GROUP BY department
					)
					GROUP BY department`

	var saleFac []SaleFactor
	var countCom []SaleFactor
	var dataRaw []SaleFactor
	if err := dbSale.Ctx().Raw(sql, month, year).Scan(&saleFac).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}
	if err := dbSale.Ctx().Raw(countCompany, month, year).Scan(&countCom).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}

	sumRevenue := 0.0
	sumEngCost := 0.0

	for _, v := range saleFac {
		for _, c := range countCom {
			if v.Department == c.Department {
				v.CountSo = c.CountSo
				v.InFactor = v.InFactor / float64(c.CountSo)
				v.ExFactor = v.ExFactor / float64(c.CountSo)

				dataRaw = append(dataRaw, v)
			}
		}
		sumRevenue += v.TotalRevenue
		sumEngCost += v.EngCost
	}

	dataResult := map[string]interface{}{
		"data":              dataRaw,
		"sale_factor_total": sumRevenue / sumEngCost,
	}
	return c.JSON(http.StatusOK, dataResult)
}

func GetSaleFactorEndPoint(c echo.Context) error {
	accountId := strings.TrimSpace(c.Param("id"))
	check := checkPermissionUser(accountId)
	if !check {
		return echo.ErrNotFound
	}
	type SaleFactorPerson struct {
		TotalRevenue float64 `json:"total_revenue" gorm:"column:total_revenue"`
		CountSo      int     `json:"count_so" gorm:"column:count_so"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		EngCost      float64 `json:"eng_cost" gorm:"column:engcost"`
		RealSF       float64 `json:"real_sf" gorm:"column:real_sf"`
		Department   string  `json:"department" gorm:"column:department"`
		StaffId      string  `json:"staff_id" gorm:"column:staff_id"`
		StaffChild   string  `json:"staff_child" gorm:"column:staff_child"`
		Fname        string  `json:"fname" gorm:"column:fname"`
		Lname        string  `json:"lname" gorm:"column:lname"`
		Nname        string  `json:"nname" gorm:"column:nname"`
	}
	today := time.Now()
	year, month, _ := today.Date()
	var countSale []CountSoPerson
	var saleFac []SaleFactorPerson
	var dataRaw []SaleFactorPerson
	sql := `select 
				sum(in_factor) as in_factor,
				sum(ex_factor) as ex_factor,
				sum(revenue) as total_revenue,
				sum(engcost) as engcost,
				sum(revenue)/sum(engcost) as real_sf,
				department,staff_id,fname,lname,nname,staff_child
			from (
				Select 
						TotalContractAmount as revenue,
						(CASE
								WHEN TotalContractAmount is not null and sale_factor is not null and sale_factor != 0 THEN TotalContractAmount/sale_factor
								ELSE 0 END
						) as engcost,
						sale_factor,
						sale_code,
						in_factor,
						ex_factor
				from so_mssql where month(PeriodStartDate) = ? and year(PeriodStartDate) = ?
				group by sonumber
			) tb_so
			LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
			where department in (
				SELECT department FROM staff_info WHERE staff_child = '' and department <> 'Sale JV' GROUP BY department
			)
			GROUP BY staff_id ORDER BY real_sf desc`

	countCompany := `SELECT COUNT(Customer_ID) as count_so , department ,fname,lname,staff_id
			from (
					Select 
									Customer_ID, Customer_name, sale_code
					from so_mssql where month(PeriodStartDate) = ? and year(PeriodStartDate) = ?
					group by Customer_ID
			) tb_so
			LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
			where department in (
					SELECT department FROM staff_info WHERE  department <> 'Sale JV' GROUP BY department
			) GROUP BY staff_id`
	sqlFactor := `select 
					sum(in_factor)/count(sonumber) as in_fac,
					sum(ex_factor)/count(sonumber) as ex_fac

				from (
				Select 
					TotalContractAmount as revenue,
					(CASE
						WHEN TotalContractAmount is not null and sale_factor is not null and sale_factor != 0 THEN TotalContractAmount/sale_factor
						ELSE 0 END
					) as engcost,
					sale_factor,
					in_factor,
					ex_factor,
					sale_code,
				sonumber
				from so_mssql
				group by sonumber
				) tb_so
				LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
				where department in (
				SELECT department FROM staff_info WHERE department <> 'Sale JV' GROUP BY department)`
	if err := dbSale.Ctx().Raw(sql, month, year).Scan(&saleFac).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}
	if err := dbSale.Ctx().Raw(countCompany, month, year).Scan(&countSale).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}
	sumFac := struct {
		InFactor float64 `gorm:"column:in_fac"`
		ExFactor float64 `gorm:"column:ex_fac"`
	}{}
	if err := dbSale.Ctx().Raw(sqlFactor).Scan(&sumFac).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}

	sumInFac := 0.0
	sumExFac := 0.0
	sumSo := 0
	for _, v := range saleFac {
		for _, c := range countSale {
			if v.StaffId == c.StaffId {
				childCountSo := 0
				childInfac := 0.0
				childExfac := 0.0
				if v.StaffChild != "" {
					splitStaff := strings.Split(v.StaffChild, ",")
					for _, s := range splitStaff {
						for _, f := range saleFac {
							if s == f.StaffId {
								childInfac += v.InFactor
								childExfac += v.ExFactor
							}
						}
						for _, ch := range countSale {
							if v.StaffId == ch.StaffId {
								childCountSo += ch.CountSo
							}
						}
					}
				}
				sumSo += c.CountSo
				v.CountSo = c.CountSo + childCountSo
				v.InFactor = (v.InFactor + childInfac) / float64(c.CountSo+childCountSo)
				v.ExFactor = (v.ExFactor + childExfac) / float64(c.CountSo+childCountSo)

				sumInFac += (v.InFactor)
				sumExFac += (v.ExFactor)

				dataRaw = append(dataRaw, v)
			}
		}
	}

	dataResult := map[string]interface{}{
		"data":      dataRaw,
		"in_factor": sumFac.InFactor,
		"ex_factor": sumFac.ExFactor,
	}
	return c.JSON(http.StatusOK, dataResult)
}

func checkPermissionUser(oneId string) bool {
	var user m.UserInfo
	if err := dbSale.Ctx().Model(&m.UserInfo{}).Where(m.UserInfo{OneId: oneId}).First(&user).Error; err != nil {
		log.Errorln(pkgName, err, "check user error :-")
		return false
	}
	if user.OneId != "" && user.Username != "" && user.Role != "" && user.SubRole != "" {
		return true
	}
	return false
}
