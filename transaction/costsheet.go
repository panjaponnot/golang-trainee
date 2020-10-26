package transaction

import (
	"golang_template/core"
	"golang_template/pkg/server"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CostSheetEndpoint ...
func CostSheetEndpoint(c echo.Context) error {
	var result server.Result
	// attend, _ := core.Newattendant()
	// a, _ := attend.GetAccountByID("6137211048")
	result.Data = "ok"
	return c.JSON(http.StatusOK, result)
}

// AvatarEndpoint ...
func AvatarEndpoint(c echo.Context) error {
	attendantClient := core.AttendantClient()
	mimeType, image, err := attendantClient.GetAvatarByID("6137211048")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error")
	}
	return c.Blob(http.StatusOK, mimeType, image)
}
