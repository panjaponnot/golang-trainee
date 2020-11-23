package core

import (
	"encoding/json"
	"fmt"
	"os"
	m "sale_ranking/model"
	"sale_ranking/pkg/billing"
	"sale_ranking/pkg/log"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

type ClientClaims struct {
	Name string `json:"name"`
}

func GetApiClientByName(name string) (m.ApiClient, string, error) {
	var apiClient m.ApiClient
	if err := dbSale.Ctx().Model(&m.ApiClient{}).Where(m.ApiClient{Name: name}).First(&apiClient).Error; err != nil {
		return apiClient, "", err
	}
	token, err := GetApiClientTokenKey(apiClient.Name)
	return apiClient, token, err
}

func GetApiClientList() ([]m.ApiClient, error) {
	var apiClients []m.ApiClient
	if err := dbSale.Ctx().Model(&m.ApiClient{}).Find(&apiClients).Error; err != nil {
		return apiClients, err
	}
	return apiClients, nil
}

func AddNewApiClient(name string) (m.ApiClient, string, error) {
	var apiClient m.ApiClient
	if err := dbSale.Ctx().Model(&m.ApiClient{}).Where(m.ApiClient{Name: name}).Attrs(m.ApiClient{Name: name}).FirstOrCreate(&apiClient).Error; err != nil {
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
	if err := dbSale.Ctx().Model(m.ApiClient{}).Where(m.ApiClient{Name: name}).Unscoped().Delete(&m.ApiClient{}).Error; err != nil {
		log.Errorln("Error -:", err)
		return 1
	}
	return GetApiClientListCli()
}

func IsClientAccess(token string) (m.ApiClient, error) {
	claims, _, err := VerifyApiClientTokenKey(token)
	return claims, err
}

func SyncBillingToDB() int {
	_ = AlertErrorSyncToBot(">>>>> Start sync billing <<<<<")
	log.Infoln("SYNC", ">>>>> Start sync billing at", time.Now().Local().Format(time.ANSIC))
	/// get data billing
	var billClient billing.Billing
	bill := billing.NewBilling("")
	token, err := bill.GetToken()
	if err != nil {
		log.Errorln("SYNC", err, "get token billing error :-")
		return 0
	}
	billClient = billing.NewBilling(token.Token)
	b, err := billClient.GetInvoiceSO()

	/// connect db
	dbSale = NewDatabase(pkgName, "salerank")
	if err := dbSale.Connect(); err != nil {
		log.Errorln("SYNC", err, "Connect to database salerank error :-")
		return 0
	}
	defer dbSale.Close()

	/// truncate table
	if err := dbSale.Ctx().Exec(`TRUNCATE TABLE invoice_status`).Error; err != nil {
		return 0
	}
	if err := dbSale.Ctx().Exec(`TRUNCATE TABLE invoice`).Error; err != nil {
		return 0
	}

	if err := dbSale.MigrateDatabase(tablesSale); err != nil {
		log.Errorln(pkgName, err, "Migrate database sale ranking error")
		return 0
	}

	// goroutine
	wg := sync.WaitGroup{}
	hasErr := 0
	MaxConcurrentPool := 20
	workingJob := 0
	for _, val := range b.Data {
		wg.Add(1)
		workingJob += 1
		go func(v billing.DataInvoiceSO) {
			defer wg.Done()

			t, err := time.Parse("2006-01-02", v.DocDate)
			if err != nil {
				fmt.Println(err)
			}

			var i m.Invoice
			if err := dbSale.Ctx().Model(&m.Invoice{}).Where(m.Invoice{InvoiceNo: v.InvoiceNo}).Attrs(m.Invoice{
				SoRef:     strings.TrimSpace(v.SoRef),
				InvoiceNo: strings.TrimSpace(v.InvoiceNo),
				DocDate:   t,
			}).FirstOrCreate(&i).Error; err != nil {
				log.Errorln("SYNC", err, "First or create  ", v.InvoiceNo, " and ", v.SoRef, "  error :-")
				msg := fmt.Sprint("[SYNC] First or create", v.InvoiceNo, " and ", v.SoRef, "  error :-", err, "\n HasError :- ", hasErr)
				_ = AlertErrorSyncToBot(msg)
				hasErr += 1
				// break
				// return 0
			}
			if len(v.InvoiceStatus) != 0 {
				for _, s := range v.InvoiceStatus {
					iSta := m.InvoiceStatus{
						InvoiceUid:        i.Uid,
						SoRef:             strings.TrimSpace(v.SoRef),
						InvoiceStatusName: strings.TrimSpace(s.InvStatusname),
					}
					if err := dbSale.Ctx().Model(&m.InvoiceStatus{}).Create(&iSta).Error; err != nil {
						log.Errorln("SYNC", err, "create invoice status", v.InvoiceNo, "  error :-")
						msg := fmt.Sprint("[SYNC] create invoice status", v.InvoiceNo, "  error :-", err, "\n HasError :- ", hasErr)
						_ = AlertErrorSyncToBot(msg)
						hasErr += 1
						// return 0
					}
				}
			}
		}(val)
		if workingJob >= MaxConcurrentPool {
			wg.Wait()
			workingJob = 0
		}
	}
	wg.Wait()
	fmt.Println("len =>", len(b.Data))
	if hasErr != 0 {
		// log.Errorln("SYNC", e, "sync has error :-")
		return 0
	}
	return 1
}

func AlertErrorSyncToBot(msg string) error {
	if err := chatBotClient.PushTextMessage("3148848982", msg, nil); err != nil {

	}
	return nil
}
