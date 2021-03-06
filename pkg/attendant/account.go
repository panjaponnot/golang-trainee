package attendant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/requests"

	"github.com/labstack/echo"
)

// GetAccountByID Get Account by ID API
func (client *Client) GetAccountByID(accountID string) (AccountProfile, error) {
	var account AccountProfile
	data, err := client.get(fmt.Sprintf("/accounts/%s", accountID))
	if err != nil {
		return account, err
	}
	dataByte, err := json.Marshal(data)
	if err := json.Unmarshal(dataByte, &account); err != nil {
		log.Errorln(pkgName, err, "Json unmarshall team member error")
		return account, nil
	}
	return account, nil
}

func (client *Client) GetAvatarByID(accountID string) (string, []byte, error) {
	var data []byte
	headers := map[string]string{
		echo.HeaderContentType:   "application/json",
		echo.HeaderAuthorization: fmt.Sprintf("%s %s", client.tokenType, client.token),
	}
	uri := fmt.Sprintf("/accounts/%s/avatar", accountID)
	r, err := requests.Get(client.url(uri), headers, bytes.NewBuffer(data), 30)
	if err != nil {
		return "", nil, err
	}
	if r.Code != http.StatusOK {
		return "", nil, errors.New(fmt.Sprintf("server return code %d %s", r.Code, string(r.Body)))
	}
	return r.Header["Content-Type"][0], r.Body, nil
}

// GetTeamMemberByID Get Team Member by ID API
func (client *Client) GetTeamMemberByID(accountID string) ([]CompanyDetail, error) {
	var company []CompanyDetail
	data, err := client.get(fmt.Sprintf("/accounts/%s/teammember", accountID))
	if err != nil {
		return company, err
	}
	dataByte, err := json.Marshal(data)
	if err := json.Unmarshal(dataByte, &company); err != nil {
		log.Errorln(pkgName, err, "Json unmarshall team member error")
		return company, nil
	}
	return company, nil
}

// GetHeadByID Get Head by ID API
func (client *Client) GetHeadByID(accountID string) ([]CompanyDetail, error) {
	var company []CompanyDetail
	data, err := client.get(fmt.Sprintf("/accounts/%s/head", accountID))
	if err != nil {
		return company, err
	}
	dataByte, err := json.Marshal(data)
	if err := json.Unmarshal(dataByte, &company); err != nil {
		log.Errorln(pkgName, err, "Json unmarshall team member error")
		return company, nil
	}
	return company, nil
}

// url Set URL Path
func (client *Client) url(path string) string {
	return fmt.Sprintf("%s%s", client.apiEndpoint, path)
}

// get Get Data
func (client *Client) get(uri string) (interface{}, error) {
	var data []byte
	headers := map[string]string{
		echo.HeaderContentType:   "application/json",
		echo.HeaderAuthorization: fmt.Sprintf("%s %s", client.tokenType, client.token),
	}
	r, err := requests.Get(client.url(uri), headers, bytes.NewBuffer(data), 30)
	if err != nil {
		return nil, err
	}
	if r.Code != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("server return code %d %s", r.Code, string(r.Body)))
	}
	var attendantAPIResult APIResult
	if err := json.Unmarshal(r.Body, &attendantAPIResult); err != nil {
		return nil, err
	}
	return reflect.ValueOf(attendantAPIResult.Data).Interface(), nil
}
