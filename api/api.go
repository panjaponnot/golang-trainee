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
	// g.Use(auth.AuthMiddlewareWithConfig(auth.Config{Skipper: func(c echo.Context) bool {
	// 	skipper := server.NewSkipperPath("")
	// 	skipper.Add("/api/v1/export", http.MethodGet)
	// 	return skipper.Test(c)
	// }}))
	export := g.Group("/export")
	export.GET("/pending", GetReportExcelSOPendingEndPoint)
	export.GET("/so", GetReportExcelSOEndPoint)

	quotation := g.Group("/quotation")
	quotation.GET("/summary", GetSummaryQuotationEndPoint)
	return nil
}