package api

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func GetSummarySaleFactorEndPoint(c echo.Context) error {
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
		TotalRevenue float64 `gorm:"column:total_revenue"`
		EngCost      float64 `gorm:"column:engcost"`
		RealSF       float64 `gorm:"column:real_sf"`
		Department   string  `gorm:"column:department"`
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

func GetSummaryInteranlFactorAndExternalFactorEndPoint(c echo.Context) error {
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
	type SaleFactor struct {
		TotalRevenue float64 `gorm:"column:total_revenue"`
		CountSo      int     `gorm:"column:count_so"`
		InFactor     float64 `gorm:"column:in_factor"`
		ExFactor     float64 `gorm:"column:ex_factor"`
		EngCost      float64 `gorm:"column:engcost"`
		RealSF       float64 `gorm:"column:real_sf"`
		Department   string  `gorm:"column:department"`
	}
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
