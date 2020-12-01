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
