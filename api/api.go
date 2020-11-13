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

	permission := g.Group("/permission")
	permission.GET("/lead/:id", CheckTeamLeadEndPoint)

	summary := g.Group("/summary")
	summary.GET("/customer", GetSummaryCustomerEndPoint)
	// staff := g.Group("/staff")
	// report.GET("/org", GetDataOrgChartEndPoint)
	staff := g.Group("/staff")
	staff.GET("/all", GetAllStaffEndPoint)              //success
	staff.GET("", GetStaffEndPoint)                     //success แต่ไม่ได้เทสpython
	staff.GET("/profile", GetStaffProfileEndPoint)      //success
	staff.POST("", CreateStaffEndPoint)                 //success
	staff.PUT("", EditStaffEndPoint)                    //success
	staff.DELETE("", DeleteStaffEndPoint)               //success
	staff.GET("/staffpicture", GetStaffPictureEndPoint) //success
	staff.GET("/staffid", GetAllStaffIdEndPoint)        //success
	staff.GET("/substaff", GetSubordinateStaffEndPoint) //success
	staff.GET("/profile/v2", GetStaffProfileV2EndPoint) //success
	staff.POST("/staffpicture", CreateStaffPictureEndPoint)

	// report.GET("/ranking/base", GetRankingBaseSale)

	return nil
}
