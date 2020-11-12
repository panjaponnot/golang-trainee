package api

import (
	"sale_ranking/core"
	"sale_ranking/pkg/cache"
	"sale_ranking/pkg/database"
	"sale_ranking/pkg/log"

	"github.com/labstack/echo/v4"
)

const pkgName = "API"

var (
	dbSale     database.Database
	dbQuataion database.Database
	dbMssql    database.Database
	redis      cache.Redis
)

func initDataStore() error {
	// Database
	dbSale = core.NewDatabase(pkgName, "salerank")
	if err := dbSale.Connect(); err != nil {
		log.Errorln(pkgName, err, "Connect to database salerank error")
		return err
	}
	dbQuataion = core.NewDatabase(pkgName, "quotation")
	if err := dbQuataion.Connect(); err != nil {
		log.Errorln(pkgName, err, "Connect to database quotation error")
		return err
	}
	dbMssql = core.NewDatabaseMssql(pkgName, "mssql")
	if err := dbMssql.Connect(); err != nil {
		log.Errorln(pkgName, err, "Connect to database sql server error")
		return err
	}
	// Redis cache
	redis = core.NewRedis()
	if err := redis.Ping(); err != nil {
		log.Errorln(pkgName, err, "Connect to redis error ->")
		return err
	}
	return nil
}

// InitApiRouter for Export API
func InitApiRouter(g *echo.Group) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
	}
	// defer dbSale.Close()
	// g.Use(auth.AuthMiddlewareWithConfig(auth.Config{Skipper: func(c echo.Context) bool {
	// 	skipper := server.NewSkipperPath("")
	// 	skipper.Add("/api/v1/export", http.MethodGet)
	// 	return skipper.Test(c)
	// }}))
	export := g.Group("/export")
	export.GET("/pending", GetReportExcelSOPendingEndPoint)
	export.GET("/so", GetReportExcelSOEndPoint)

	report := g.Group("/report")
	report.GET("/org", GetDataOrgChartEndPoint)
	report.GET("/pending", GetReportSOPendingEndPoint)
	report.GET("/so", GetReportSOEndPoint)
	report.PUT("/so", EditSOEndPoint)
	report.GET("/ranking/base", GetRankingBaseSale)
	report.GET("/ranking/key", GetRankingKeyAccountEndPoint)
	report.GET("/ranking/recovery", GetRankingRecoveryEndPoint)
	report.GET("/ranking/lead", GetRankingTeamLeadEndPoint)

	quotation := g.Group("/quotation")
	quotation.GET("/summary", GetSummaryQuotationEndPoint)
	quotation.POST("/log", CreateLogQuotation)

	permission := g.Group("/permission")
	permission.GET("/lead/:id", CheckTeamLeadEndPoint)

	summary := g.Group("/summary")
	summary.GET("/customer", GetSummaryCustomerEndPoint)
	// staff := g.Group("/staff")
	// report.GET("/org", GetDataOrgChartEndPoint)
	staff := g.Group("/staff")
	staff.GET("", GetStaffEndPoint)
	staff.GET("/profile", GetStaffProfileEndPoint)
	staff.POST("", CreateStaffEndPoint)
	staff.PUT("", EditStaffEndPoint)
	staff.DELETE("", DeleteStaffEndPoint)
	staff.GET("/staffpicture", GetStaffPictureEndPoint)
	staff.GET("/staffid", GetAllStaffIdEndPoint)
	staff.GET("/SubStaff", GetSubordinateStaffEndPoint)
	staff.GET("/profilev2", GetStaffProfileV2EndPoint)
	staff.GET("/cratestaffpicture", CreateStaffPictureEndPoint)

	// report.GET("/ranking/base", GetRankingBaseSale)

	return nil
}
