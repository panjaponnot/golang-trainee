package core

import (
	"encoding/json"
	"fmt"
	"os"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"time"

	"github.com/olekukonko/tablewriter"
)

type ClientClaims struct {
	Name string `json:"name"`
}

func GetApiClientByName(name string) (m.ApiClient, string, error) {
	var apiClient m.ApiClient
	if err := db.Ctx().Model(&m.ApiClient{}).Where(m.ApiClient{Name: name}).First(&apiClient).Error; err != nil {
		return apiClient, "", err
	}
	token, err := GetApiClientTokenKey(apiClient.Name)
	return apiClient, token, err
}

func GetApiClientList() ([]m.ApiClient, error) {
	var apiClients []m.ApiClient
	if err := db.Ctx().Model(&m.ApiClient{}).Find(&apiClients).Error; err != nil {
		return apiClients, err
	}
	return apiClients, nil
}

func AddNewApiClient(name string) (m.ApiClient, string, error) {
	var apiClient m.ApiClient
	if err := db.Ctx().Model(&m.ApiClient{}).Where(m.ApiClient{Name: name}).Attrs(m.ApiClient{Name: name}).FirstOrCreate(&apiClient).Error; err != nil {
		return apiClient, "", err
	}
	token, err := GetApiClientTokenKey(apiClient.Name)
	return apiClient, token, err
}

func GetApiClientTokenKey(name string) (string, error) {
	claims := ClientClaims{Name: name}
	claimsByte, _ := json.Marshal(claims)
	return EncryptWithServerKey(claimsByte)
}

func VerifyApiClientTokenKey(token string) (m.ApiClient, string, error) {
	claimsByte, err := DecryptWithServerKey(token)
	if err != nil {
		return m.ApiClient{}, token, err
	}
	var apiClient m.ApiClient
	var claims ClientClaims
	if err := json.Unmarshal(claimsByte, &claims); err != nil {
		return apiClient, token, err
	}
	return GetApiClientByName(claims.Name)
}

func GetApiClientListCli() int {
	apiClients, err := GetApiClientList()
	if err != nil {
		log.Errorln("Get api client error -:", err)
		return 1
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{fmt.Sprintf("UID (%d)", len(apiClients)), "Created At", "Name", "Key"})
	table.SetAutoWrapText(false)
	for _, v := range apiClients {
		token, err := GetApiClientTokenKey(v.Name)
		if err != nil {
			log.Errorln("Encrypt client key error -:", err)
			return 1
		}
		table.Append([]string{v.Uid.String(), v.CreatedAt.Local().Format(time.ANSIC), v.Name, token})
	}
	table.Render()
	return 0
}

func AddApiClient(name string) int {
	_, _, err := AddNewApiClient(name)
	if err != nil {
		log.Errorln("Add new client error -:", err)
		return 1
	}
	return GetApiClientListCli()
}

func DeleteApiClient(name string) int {
	if err := db.Ctx().Model(m.ApiClient{}).Where(m.ApiClient{Name: name}).Unscoped().Delete(&m.ApiClient{}).Error; err != nil {
		log.Errorln("Error -:", err)
		return 1
	}
	return GetApiClientListCli()
}

func IsClientAccess(token string) (m.ApiClient, error) {
	claims, _, err := VerifyApiClientTokenKey(token)
	return claims, err
}
