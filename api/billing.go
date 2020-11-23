package api

import (
	"fmt"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/billing"
	"sale_ranking/pkg/log"

	"github.com/labstack/echo/v4"
)

var billClient billing.Billing

func InitBill() error {
	bill := billing.NewBilling("")
	token, err := bill.GetToken()
	fmt.Println("=========token====", token.Token)
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

	// d, _ := billClient.GetInvoiceSO()
	// b := 0
	// for _, v := range d.Data {
	// 	b += hasDupes(d.Data, v)
	// }
	// b := hasDupes(d.Data)
	// fmt.Println("get data invoice >>", b)
	var i []m.Invoice
	if err := dbSale.Ctx().Model(&m.Invoice{}).Find(&i).Error; err != nil {

	}

	return c.JSON(http.StatusOK, i)
}

func hasDupes(m []billing.DataInvoiceSO, bill billing.DataInvoiceSO) int {
	// x := make(map[string]struct{})
	a := 0
	for _, val := range m {
		if bill.InvoiceNo == val.InvoiceNo {
			a += 1
		}
	}
	if a > 1 {
		return 1
	}
	return 0
}
