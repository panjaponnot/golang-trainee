package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetBillingEndPoint(c echo.Context) error {

	// if err := billing.NewBilling("", ""); err != nil {

	// }

	return c.JSON(http.StatusOK, nil)
}
