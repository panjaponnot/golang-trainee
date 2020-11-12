package billing

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sale_ranking/pkg/requests"
	"sale_ranking/pkg/util"

	"github.com/labstack/echo"
)

// AttendantEndpoint API Endpoint
const (
	pkgName            = "BILLING"
	DefaultApiEndpoint = "https:/203.150.199.36:9191"
)

var (
	Username = util.GetEnv("USERNAME_BILLING", "")
)

func NewBilling(token string, tokenType string) Billing {
	return Billing{
		token:       token,
		tokenType:   tokenType,
		apiEndpoint: DefaultApiEndpoint,
	}
}

// url Set URL Path
func (billing *Billing) url(path string) string {
	return fmt.Sprintf("%s%s", billing.apiEndpoint, path)
}

// get Get Data
func (billing *Billing) get(uri string) (interface{}, error) {
	var data []byte
	headers := map[string]string{
		echo.HeaderContentType:   "application/json",
		echo.HeaderAuthorization: fmt.Sprintf("%s %s", billing.tokenType, billing.token),
	}
	r, err := requests.Get(billing.url(uri), headers, bytes.NewBuffer(data), 30)
	if err != nil {
		return nil, err
	}
	if r.Code != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("server return code %d %s", r.Code, string(r.Body)))
	}
	var billingAPIResult APIResult
	if err := json.Unmarshal(r.Body, &billingAPIResult); err != nil {
		return nil, err
	}
	return reflect.ValueOf(billingAPIResult.Data).Interface(), nil
}

func (billing *Billing) GetToken() (AuthenticationResult, error) {
	var Authen AuthenticationResult
	headers := map[string]string{
		echo.HeaderContentType: "application/json",
	}

	body, _ := json.Marshal(&map[string]string{
		"username": Username,
	})
	r, err := requests.Post(billing.url("/checkuser"), headers, bytes.NewBuffer(body), 30)
	if err != nil {
		return Authen, err
	}
	if r.Code != http.StatusOK {
		return Authen, errors.New(fmt.Sprintf("server return code %d %s", r.Code, string(r.Body)))
	}
	var billingAPIResult AuthenticationResult
	if err := json.Unmarshal(r.Body, &billingAPIResult); err != nil {
		return Authen, err
	}
	return Authen, nil
}
