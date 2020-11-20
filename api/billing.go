package api

import (
	"net/http"
	"sale_ranking/pkg/billing"
	"sale_ranking/pkg/log"

	"github.com/labstack/echo/v4"
)

var billClient billing.Billing

func InitBill() error {
	bill := billing.NewBilling("")
	token, err := bill.GetToken()
	// fmt.Println("=========token====", token.Token)
	if err != nil {
		return err
	}
	billClient = billing.NewBilling(token.Token)
	return nil
}

func GetBillingEndPoint(c echo.Context) error {
	if err := InitBill(); err != nil {
		log.Errorln(pkgName, err, "New bill err")
		return echo.ErrInternalServerError
	}

	d, _ := billClient.GetInvoiceSO()

	// fmt.Println("get data invoice >>", d)
	return c.JSON(http.StatusOK, d)
}
