package export

import (
	"github.com/labstack/echo/v4"
)

// InitApiRouter for Export API
func InitApiRouter(g *echo.Group) error {
	// g.Use(auth.AuthMiddlewareWithConfig(auth.Config{Skipper: func(c echo.Context) bool {
	// 	skipper := server.NewSkipperPath("")
	// 	skipper.Add("/api/v1/export", http.MethodGet)
	// 	return skipper.Test(c)
	// }}))

	g.GET("", GetReportExcelSOPendingEndPoint)
	g.GET("/test", TestbotEndPoint)

	return nil
}
