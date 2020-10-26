package transaction

import "github.com/labstack/echo/v4"

// InitAPIRouter for Transection API
func InitAPIRouter(g *echo.Group) error {
	costSheet := g.Group("/costsheet")
	costSheet.GET("", CostSheetEndpoint)

	return nil
}
