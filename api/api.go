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
	dbEquip    database.Database
	dbSale     database.Database
	dbQuataion database.Database
	dbMssql    database.Database
	redis      cache.Redis
)

func initDataStore() error {
	// Database
	// dbEquip = core.NewDatabaseMssql(pkgName, "equip")
	// if err := dbEquip.Connect(); err != nil {
	// 	log.Errorln(pkgName, err, "Connect to database equip error")
	// 	return err
	// }
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
	// dbMssql = core.NewDatabaseMssql(pkgName, "mssql")
	// if err := dbMssql.Connect(); err != nil {
	// 	log.Errorln(pkgName, err, "Connect to database sql server error")
	// 	return err
	// }
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
	track.GET("/so", GetSummaryCustomerEndPoint)
	track.GET("/doc/so", GetSOCustomerEndPoint)            //doc/so  /cus/so
	track.GET("/doc/so/cs", GetSOCustomerCsNumberEndPoint) //doc/so/cs  /cus/so/cs

	track.GET("/invoice", GetTrackingInvoiceEndPoint)
	track.GET("/doc/inv", GetDocTrackingInvoiceEndPoint)
	track.GET("/doc/inv/so", GetDocTrackingInvoiceSOEndPoint)

	track.GET("/bill", GetTrackingBillingEndPoint)
	track.GET("/doc/bill", GetTrackingBillingStatusEndPoint)        //doc/bill  /bill/status
	track.GET("/doc/bill/inv", GetTrackingBillingStatusInvEndPoint) //doc/bill/inv

	track.GET("/receipt", GetTrackingReceiptEndPoint)
	track.GET("/doc/rc", GetSOTrackingReceiptEndPoint)       //doc/rc /receipt/so
	track.GET("/doc/rc/inv", GetSOTrackingReceiptCsEndPoint) //doc/rc/inv /receipt/so/cs

	track.GET("/doc/cs", GetSOTrackingCsEndPoint) //doc/cs

	track.GET("/detail_reciept", GetDetailReceiptEndPoint)     //
	track.GET("/detail_billing", GetDetailBillingEndPoint)     //
	track.GET("/detail_invoice", GetDetailInvoiceEndPoint)     //
	track.GET("/detail_so", GetDetailSoEndPoint)               //**
	track.GET("/detail_costsheet", GetDetailCostsheetEndPoint) //**ต้องกรอกวันที่ เพราะในdbเขากรอกผิด

	track.GET("/detail_reciept/change", GetDetailReceiptChangeEndPoint)     //
	track.GET("/detail_billing/change", GetDetailBillingChangeEndPoint)     //
	track.GET("/detail_invoice/change", GetDetailInvoiceChangeEndPoint)     //
	track.GET("/detail_so/change", GetDetailSoChangeEndPoint)               //**
	track.GET("/detail_costsheet/change", GetDetailCostsheetChangeEndPoint) //**ต้องกรอกวันที่ เพราะในdbเขากรอกผิด

	track.GET("/detail_reciept/notchange", GetDetailReceiptNotChangeEndPoint)     //
	track.GET("/detail_billing/notchange", GetDetailBillingNotChangeEndPoint)     //
	track.GET("/detail_invoice/notchange", GetDetailInvoiceNotChangeEndPoint)     //
	track.GET("/detail_so/notchange", GetDetailSoNotChangeEndPoint)               //**
	track.GET("/detail_costsheet/notchange", GetDetailCostsheetNotChangeEndPoint) //**ต้องกรอกวันที่ เพราะในdbเขากรอกผิด

	export := g.Group("/export")
	export.GET("/pending", GetReportExcelSOPendingEndPoint)
	export.GET("/so", GetReportExcelSOEndPoint)
	export.GET("/tracking", GetReportExcelTrackingEndPoint)
	export.GET("/quotation", GetReportExcelQuotationEndPoint)
	export.GET("/ranking/base", GettReportExcelRankBaseSaleEndPoint)
	export.GET("/ranking/key", GettReportExcelRankKeyAccEndPoint)
	export.GET("/ranking/recovery", GettReportExcelRankRecoveEndPoint)
	export.GET("/ranking/lead", GetReportExcelRankTeamLeadEndPoint)
	export.GET("/factor/sale/:id", GetReportExcelSaleFactorEndPoint)

	export.GET("/detail_reciept", GetExcelDetailReceiptEndPoint) //
	// export.GET("/detail_billing", GetExcelDetailBillingEndPoint)     //
	// export.GET("/detail_invoice", GetExcelDetailInvoiceEndPoint)     //
	// export.GET("/detail_so", GetExcelDetailSoEndPoint)               //**
	// export.GET("/detail_costsheet", GetExcelDetailCostsheetEndPoint) //**ต้องกรอกวันที่ เพราะในdbเขากรอกผิด

	report := g.Group("/report")
	report.GET("/org", GetDataOrgChartEndPoint)
	report.GET("/pending", GetReportSOPendingEndPoint)
	report.GET("/pending/type", GetReportSOPendingTypeEndPoint)
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

	summary := g.Group("/summary")
	summary.GET("/all", GetSummaryPendingSOEndPoint)
	summary.GET("/contract", GetContractEndPoint)
	summary.GET("/teams", GetTeamsEndPoint)
	summary.GET("/teams/department", GetTeamsDepartmentEndPoint)

	summary.GET("/vm", GetVmSummaryEndPoint)
	summary.GET("/vm/v2", GetVmSummaryV2EndPoint)

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
	staff.GET("/dept/:id", DepartmentStaffV2EndPoint)
	staff.GET("/teams/dept", GetTeamsDeptStaffEndPoint)

	staff.GET("/dept/child/:id", DepartmentStaffAllEndPoint)

	staff.GET("/ranking/base", GetRankingBaseSale2)
	// report.GET("/ranking/base", GetRankingBaseSale)
	staff.GET("/child/:id", StaffChildAllEndPoint)

	bot := g.Group("/bot")
	bot.GET("/userone", GetUserOneThEndPoint)

	alert := g.Group("/alert")
	alert.POST("/bot", AlertSoToBotEndPoint)
	alert.POST("/terminate/success", AlertTerminateRunSuccessEndPoint)
	alert.POST("/terminate/fail", AlertTerminateRunFailEndPoint)
	alert.POST("/terminate/credit", AlertTerminateCreditNoteEndPoint)
	alert.POST("/so/success", AlertSoRunSuccessEndPoint)
	alert.POST("/so/fail", AlertSoRunFailEndPoint)
	alert.POST("/so/main", AlertSoRunMainEndPoint)
	alert.POST("/so/change", AlertSoRunChangeEndPoint)
	alert.POST("/so/invoice", AlertSoRunInvoiceEndPoint)

	bill := g.Group("/bill")
	bill.GET("", GetBillingEndPoint)

	tracking := g.Group("/tracking")
	tracking.GET("/all", TrackingEndPoint)

	revanue := g.Group("/revenue")
	revanue.GET("/all", RevenueEndPoint)

	so_receive := g.Group("/soreceive")
	so_receive.GET("/all", SOReceiveEndPoint)

	receive_tracking := g.Group("/receive_tracking")
	receive_tracking.GET("/costsheet", CostSheet_Status)
	receive_tracking.GET("/invoice", Invoice_Status)
	receive_tracking.GET("/billing", Billing_Status)
	receive_tracking.GET("/reciept", Reciept_Status)

	tracking_costsheet := g.Group("/tracking_costsheet")
	tracking_costsheet.GET("/Detail", Costsheet_Detail)

	tracking_invoice := g.Group("/tracking_invoice")
	tracking_invoice.GET("/Detail", SO_Detail)

	tracking_billing := g.Group("/tracking_billing")
	tracking_billing.GET("/Detail", Invoice_Detail)

	tracking_reciept := g.Group("/tracking_reciept")
	tracking_reciept.GET("/Detail", Reciept_Detail)

	return nil
}
