package api

import (
	v2 "sale_ranking/api/v2"

	"github.com/labstack/echo/v4"
)

// InitApiRouter for Export API
func InitApiRouter(g *echo.Group) error {
	// g.Use(auth.AuthMiddlewareWithConfig(auth.Config{Skipper: func(c echo.Context) bool {
	// 	skipper := server.NewSkipperPath("")
	// 	skipper.Add("/api/v1/export", http.MethodGet)
	// 	return skipper.Test(c)
	// }}))

	g.GET("/pending", v2.GetReportExcelSOPendingEndPoint)
	g.GET("/so", v2.GetReportExcelSOEndPoint)

	return nil
}
