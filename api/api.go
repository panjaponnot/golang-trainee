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
	// middleware and skipper path
	// g.Use(auth.UserAuthMiddleware(auth.Config{Skipper: func(c echo.Context) bool {
	// 	skipper := server.NewSkipperPath("")
	// 	skipper.Add("/api/v2/export", http.MethodGet)
	// 	return skipper.Test(c)
	// }}))

	track := g.Group("/track")
	track.GET("/invoice", GetTrackingInvoiceEndPoint)
	track.GET("/bill", GetTrackingBillingEndPoint)
	track.GET("/customer", GetSummaryCustomerEndPoint)

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
	report.PUT("/pending", UpdateSOEndPoint)

	sale := g.Group("/factor")
	sale.GET("/summary/:id", GetSummarySaleFactorEndPoint)
	sale.GET("/type/:id", GetSummaryInternalFactorAndExternalFactorEndPoint)
	sale.GET("/sale/:id", GetSaleFactorEndPoint)

	quotation := g.Group("/quotation")
	quotation.GET("/summary", GetSummaryQuotationEndPoint)
	quotation.PUT("/log", CreateLogQuotationEndPoint)
	quotation.GET("/log/:id", GetLogQuotationEndPoint)

	permission := g.Group("/permission")
	permission.GET("/lead/:id", CheckTeamLeadEndPoint)

	// summary := g.Group("/summary")
	// summary.GET("/customer", GetSummaryCustomerEndPoint)

	webhook := g.Group("/webhook")
	webhook.GET("/user", GetUserOneThEndPoint)
	webhook.GET("/expire", AlertExpireEndPoint)
	webhook.GET("/approve/all", AlertApproveAllEndPoint)
	webhook.GET("/approve", AlertApproveEndPoint)
	webhook.GET("/quotation", CheckQuotationEndPoint)

	// staff := g.Group("/staff")
	// report.GET("/org", GetDataOrgChartEndPoint)
	staff := g.Group("/staff")
	staff.GET("/all", GetAllStaffEndPoint)                  //success
	staff.GET("", GetStaffEndPoint)                         //success
	staff.GET("/profile", GetStaffProfileEndPoint)          //success
	staff.POST("", CreateStaffEndPoint)                     //success
	staff.PUT("", EditStaffEndPoint)                        //success
	staff.DELETE("", DeleteStaffEndPoint)                   //success
	staff.GET("/staffpicture", GetStaffPictureEndPoint)     //success
	staff.GET("/staffid", GetAllStaffIdEndPoint)            //success
	staff.GET("/substaff", GetSubordinateStaffEndPoint)     //success
	staff.GET("/profile/v2", GetStaffProfileV2EndPoint)     //success
	staff.POST("/staffpicture", CreateStaffPictureEndPoint) //success
	staff.GET("/summary", HeaderSummaryEndPoint)

	staff.GET("/dept", DepartmentStaffEndPoint)

	// report.GET("/ranking/base", GetRankingBaseSale)

	bot := g.Group("/bot")
	bot.GET("/userone", GetUserOneThEndPoint)

	bill := g.Group("/bill")
	bill.GET("", GetBillingEndPoint)

	return nil
}
