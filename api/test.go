package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/requests"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/labstack/echo/v4"
)

// package api

// import (
// 	"net/http"
// 	m "sale_ranking/model"
// 	"sale_ranking/pkg/log"
// 	"strings"

// 	"github.com/labstack/echo/v4"
// )

func GettestEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	search := strings.TrimSpace(c.QueryParam("search"))
	var CusAll []m.CustomerTest

	// sql := `SELECT custormer_info,

	sql := `SELECT customer_info.customer_id,
	customer_info.customer_nameTH,
	customer_info.customer_nameEN,
	customer_info.sale_id,
	customer_info.tax_id

	FROM customer_info
	where INSTR(CONCAT_WS('|', customer_id, customer_nameTH , customer_nameEN, sale_id, tax_id), ?)
   ; `

	if err := dbSale.Ctx().Raw(sql, search).Scan(&CusAll).Error; err != nil {
		log.Errorln("GetAllStaff error :-", err)
	}
	log.Infoln("-->", CusAll)
	return c.JSON(http.StatusOK, CusAll)
}

func CreateTestEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	//หน้าบ้านส่งมา
	hasErr := 0
	data := struct {
		CustomerId     string `json:"customer_id" gorm:"customer_id"`
		CustomerNameTH string `json:"customer_nameTH" gorm:"customer_nameTH"`
		CustomerNameEN string `json:"customer_nameEN" gorm:"customer_nameEN"`
		SaleId         string `json:"sale_id" gorm:"sale_id"`
		TaxId          string `json:"tax_id" gorm:"tax_id"`
		Head           int    `json:"head" gorm:"head"`
	}{}

	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}
	//โค้ดเช็คซ้ำ
	var CusInfo []m.CustomerTest
	Tim := time.Now()
	create_by := ""
	tel1 := ""
	tel2 := ""
	tel3 := ""
	fax := ""
	email := ""
	website := ""
	// if err := dbSale.Ctx().Raw(`SELECT customer_id from customer_info WHERE customer_id = ?;`, data.CustomerId).Scan(&CusInfo).Error; err != nil {
	//  log.Errorln("CheckStaffInfo error :-", err)
	// }
	// if len(CusInfo) > 0 {
	//  return c.JSON(http.StatusBadRequest, m.Result{Message: "duplicate customer id"})
	// }

	if err := dbSale.Ctx().Exec(`insert into customer_info (customer_id, customer_nameTH, customer_nameEN, sale_id, tax_id, head, 
	 customer_createdate, create_by,tel1,tel2,tel3,fax,email,website)
	 values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	;`, data.CustomerId, data.CustomerNameTH, data.CustomerNameEN, data.SaleId, data.TaxId, data.Head, Tim, create_by, tel1, tel2, tel3, fax, email, website).Error; err != nil {
		log.Errorln("create error :-", err)
	}
	if len(CusInfo) > 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "duplicate customer id"})
	}

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, m.Result{Message: "create success"})
}

// func EditTestEndPoint(c echo.Context) error {
// 	if err := initDataStore(); err != nil {
// 		log.Errorln(pkgName, err, "connect database error")
// 		return c.JSON(http.StatusInternalServerError, err)
// 	}

// 	data := struct {
// 		CustomerId     string `json:"customer_id" `
// 		CustomerNameTH string `json:"customer_nameTH" `
// 		CustomerNameEN string `json:"customer_nameEN" `
// 		SaleId         string `json:"sale_id" `
// 		TaxId          string `json:"tax_id" `
// 		Head           int    `json:"head"`
// 	}{}
// 	if err := c.Bind(&data); err != nil {
// 		return echo.ErrBadRequest
// 	}
// 	if err := dbSale.Ctx().Exec("UPDATE customer_info SET customer_id = ?,customer_nameTH = ?,customer_nameEN = ?,sale_id = ?,tax_id = ?,head = ? WHERE customer_info = ?;", data.CustomerId, data.CustomerNameTH, data.CustomerNameEN, data.SaleId, data.TaxId, data.Head).Error; err != nil {
// 		log.Errorln("UPDATE error :-", err)
// 		return c.JSON(http.StatusInternalServerError, err)
// 	}

// 	return c.JSON(http.StatusOK, m.Result{Message: "update success"})
// }

//Update data
func EditTestEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	//รับมาจากหน้าบ้าน
	data := struct {
		CustomerId     string `json:"customer_id" `
		CustomerNameTH string `json:"customer_nameTH" `
		CustomerNameEN string `json:"customer_nameEN" `
		SaleId         string `json:"sale_id" `
		TaxId          string `json:"tax_id" `
		Head           int    `json:"head"`
	}{}
	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}

	Tim := time.Now()

	//write logic by yourself
	if err := dbSale.Ctx().Exec("UPDATE customer_info SET customer_id = ?,customer_nameTH = ?,customer_nameEN = ?,sale_id = ?,tax_id = ?,head = ?,customer_updatedate = ? WHERE customer_id = ?;",
		data.CustomerId, data.CustomerNameTH, data.CustomerNameEN, data.SaleId, data.TaxId, data.Head, Tim, data.CustomerId).Error; err != nil {
		log.Errorln("UPDATEStaffInfo error :-", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, m.Result{Message: "update success"})
}

//ลบข้อมูลจาก ไอดี
func DeleteTestEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	data := struct {
		CustomerId string `json:"customer_id" gorm:"column:customer_id"`
	}{}
	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}
	if err := dbSale.Ctx().Exec("DELETE FROM customer_info WHERE customer_id = ?", data.CustomerId).Error; err != nil {
		log.Errorln("DeleteStaffInfo error :-", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, m.Result{Message: "DELETE success"})
}

//alert chatbox one chat
func AlertTestEndPoint(c echo.Context) error {

	str := strings.TrimSpace(c.QueryParam("str"))
	name := strings.TrimSpace(c.QueryParam("name"))
	lname := strings.TrimSpace(c.QueryParam("lname"))
	NewStrData := fmt.Sprintf("ข้อความว่า: %s \nชื่อลูกค้า: %s \nนามสกุล %s ", str, name, lname)

	// To := 25078584384
	url := "https://chat-api.one.th/message/api/v1/push_message"
	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
		BotId              string `json:"bot_id"`
		Type               string `json:"type"`
		Message            string `json:"message"`
		CustomNotification string `json:"custom_notification"`
	}{
		// To: "198008320896",
		To: "1038412861448086",
		// To: OneId,
		// To:                 OneId,
		// BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		BotId: "Becf3d73c867f508ab7a8f5d62ceceb64", //จุกกุ
		// BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		Type:               "text",
		Message:            NewStrData,
		CustomNotification: "เปิดอ่านข้อความใหม่จากทางเราดิวะสาส",
	})

	headers := map[string]string{
		// "Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef", //ของจริง
		"Authorization": "Bearer A548a4dd47e3c5108affe99b48b5c0218db9bcaaca6b34470b389bd04a19c3e30e1b99dad38844be387e939f755d194be", //จุกกุ
		"Content-Type":  "application/json",
	}
	_, err := requests.Post(url, headers, bytes.NewBuffer(payload), 50)
	if err != nil {
		log.Errorln("Error QuickReply", err)
		return err

	}

	return nil
}

//ออกReport Excel
func GetReportExcelTestEndPoint(c echo.Context) error {

	//////////////  getListStaffID  //////////////
	rawData := []struct {
		CustomerId     string `json:"customer_id" gorm:"customer_id"`
		CustomerNameTH string `json:"customer_nameTH" gorm:"customer_nameTH"`
		CustomerNameEN string `json:"customer_nameEN" gorm:"customer_nameEN"`
		SaleId         string `json:"sale_id" gorm:"sale_id"`
		TaxId          string `json:"tax_id" gorm:"tax_id"`
		Head           int    `json:"head" gorm:"head"`
	}{}

	if err := dbSale.Ctx().Raw(`
	SELECT customer_info.customer_id,
	customer_info.customer_nameTH,
	customer_info.customer_nameEN,
	customer_info.sale_id,
	customer_info.tax_id,
	customer_info.head
	FROM customer_info
		`).Scan(&rawData).Error; err != nil {

		log.Errorln(pkgName, err, "Select data error")
	}

	log.Infoln("-->", rawData)
	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "Test"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "customer_id")
	f.SetCellValue(mode, "B1", "customer_nameTH")
	f.SetCellValue(mode, "C1", "customer_nameEN")
	f.SetCellValue(mode, "D1", "sale_id")
	f.SetCellValue(mode, "E1", "tax_id")
	f.SetCellValue(mode, "F1", "head")

	colCustomerId := "A"
	colCustomerNameTH := "B"
	colCustomerNameEN := "C"
	colSaleId := "D"
	colTaxId := "E"
	colHead := "F"

	//ตัวแรก เป็นตำแหน่ง ตัว2 จะเป็นข้อมูล
	for k, v := range rawData {
		f.SetCellValue(mode, fmt.Sprint(colCustomerId, k+2), v.CustomerId)
		f.SetCellValue(mode, fmt.Sprint(colCustomerNameTH, k+2), v.CustomerNameTH)
		f.SetCellValue(mode, fmt.Sprint(colCustomerNameEN, k+2), v.CustomerNameEN)
		f.SetCellValue(mode, fmt.Sprint(colSaleId, k+2), v.SaleId)
		f.SetCellValue(mode, fmt.Sprint(colTaxId, k+2), v.TaxId)
		f.SetCellValue(mode, fmt.Sprint(colHead, k+2), v.Head)

	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	buff, err := f.WriteToBuffer()
	if err != nil {
		log.Errorln("XLSX export error ->", err)
		return c.JSON(http.StatusInternalServerError, m.Result{Error: "export error"})
	}
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buff.Bytes())

	// return c.JSON(http.StatusOK, model.Result{Data: rawData, Total: len(rawData)})
}
