package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	// "sale_ranking/core"
	m "sale_ranking/model"
	// "sale_ranking/pkg/attendant"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/requests"
	// "sale_ranking/pkg/server"
	// "sale_ranking/pkg/util"
	// "strconv"
	"strings"
	// "time"
	// "github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func AlertSoToBotEndPoint(c echo.Context) error {
	// data request
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		OrderBotId        	[]string		`json:"bot_id"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || len(bodyData.OrderBotId) == 0 || bodyData.EmployeeCode == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if len(bodyData.OrderBotId) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list bot_id"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}

	//fix text to sending alert
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	bot_id := strings.Join(bodyData.OrderBotId, ", ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nอยู่ในคิว BOT ลำดับที่ %s", paperless_no, bot_id)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertTerminateRunSuccessEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		EntryDate         	string			`json:"entry_date"`
		SoNumber        		string			`json:"so_number"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.EntryDate == "" || bodyData.SoNumber == "" || bodyData.EmployeeCode == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else if bodyData.EntryDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "entry date is empty"})
		} else if bodyData.SoNumber == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "so number is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}
	//
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nดำเนินการสำเร็จแล้ว เมื่อเวลา %s \nTerminate เลขที่ %s", paperless_no, bodyData.CustomerName, bodyData.EntryDate, bodyData.SoNumber)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertTerminateRunFailEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		EntryDate         	string			`json:"entry_date"`
		Comment	        		string			`json:"comment"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.EntryDate == "" || bodyData.EmployeeCode == "" || bodyData.Comment == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else if bodyData.EntryDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "entry date is empty"})
		} else if bodyData.Comment == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "comment is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}

	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nที่ส่งมามีข้อผิดพลาด คือ %s เมื่อเวลา %s \nรบกวนติดต่อสอบถามฝ่าย Business Support เพื่อดำเนินการต่อ Email: businesssupport@inet.co.th", paperless_no, bodyData.CustomerName, bodyData.Comment, bodyData.EntryDate)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertTerminateCreditNoteEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		TerminateDate       string			`json:"terminate_date"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.TerminateDate == "" || bodyData.EmployeeCode == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else if bodyData.TerminateDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "Terminate date is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}

	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nต้องการที่จะ Terminate และให้มีผลในวันที่ %s ตามเดิม \nรบกวนทำการลดหนี้ Invoice และนำส่งเอกสารไปที่ Paperless เพื่อทำการ Approve ใหม่อีกครั้ง", paperless_no, bodyData.CustomerName, bodyData.TerminateDate)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertSoRunSuccessEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		EntryDate         	string			`json:"entry_date"`
		SoNumber        		string			`json:"so_number"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.EntryDate == "" || bodyData.SoNumber == "" || bodyData.EmployeeCode == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else if bodyData.EntryDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "entry date is empty"})
		} else if bodyData.SoNumber == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "so number is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}
	//
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nดำเนินการสำเร็จแล้ว เมื่อเวลา %s \nSO เลขที่ %s", paperless_no, bodyData.CustomerName, bodyData.EntryDate, bodyData.SoNumber)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertSoRunFailEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		EntryDate         	string			`json:"entry_date"`
		Comment	        		string			`json:"comment"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.EntryDate == "" || bodyData.EmployeeCode == "" || bodyData.Comment == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else if bodyData.EntryDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "entry date is empty"})
		} else if bodyData.Comment == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "comment is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}

	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nที่ส่งมามีข้อผิดพลาด คือ %s \nรบกวนแก้ไขตามรายละเอียด เมื่อเวลา %s \nและนำส่งเอกสารใหม่ไปที่ Paperless เพื่อทำการ Approve อีกครั้ง", paperless_no, bodyData.CustomerName, bodyData.Comment, bodyData.EntryDate)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertSoRunMainEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.EmployeeCode == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}
	//
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nเป็นเอกสารหลัก จึงทำให้เอกสารนั้นไม่มีการเปิดเลข SO", paperless_no, bodyData.CustomerName)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertSoRunChangeEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.EmployeeCode == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}
	//
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nเป็น SO Status Change รบกวนติดต่อสอบถามฝ่ายสนับสนุนงานขาย (SO) เพื่อดำเนินการต่อ  Email : inet-so@inet.co.th ", paperless_no, bodyData.CustomerName)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

func AlertSoRunInvoiceEndPoint(c echo.Context) error {
	bodyData := struct {
		SoNumber         		string		`json:"so_number"`
		CustomerName        string		`json:"customer_name"`
		InvoiceNumber       string		`json:"invoice_number"`
		ContractStart       string		`json:"contract_start"`
		ContractEnd 	      string		`json:"contract_end"`
		Amount			 	      string		`json:"amount"`
		EmployeeCode        string 		`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if bodyData.SoNumber == "" || bodyData.CustomerName == "" || bodyData.InvoiceNumber == "" || bodyData.ContractStart == "" || bodyData.ContractEnd == "" || bodyData.Amount == "" || bodyData.EmployeeCode == "" {
		if bodyData.SoNumber == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "so number is empty"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no is empty"})
		} else if bodyData.InvoiceNumber == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "invoice number is empty"})
		} else if bodyData.ContractStart == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "contract start is empty"})
		} else if bodyData.ContractEnd == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "contract end is empty"})
		} else if bodyData.Amount == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "amount is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}

	//get staff onfo from attendant
	account_id, error := GetAccountID(bodyData.EmployeeCode)
	if error != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "get staff information error"})
	}
	if account_id == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "noy found staff information"})
	}
	//
	var TextMessage = fmt.Sprintf("SO เลขที่ %s \nลูกค้า %s \nมีการออก invoice แล้วเลขที่ %s \nรอบบริการ: %s ถึง %s ยอด %s บาท", bodyData.SoNumber, bodyData.CustomerName, bodyData.InvoiceNumber, bodyData.ContractStart, bodyData.ContractEnd, bodyData.Amount)

	//alert one chat
	result, err := SendAlertToOneChat(account_id, TextMessage)
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "error sending message to one chat"})
	}

	return c.JSON(http.StatusOK, m.Result{Message: result})
}

//sub function
func GetAccountID(staff_id string) (string, error) {
	//get one id from attendant
	var url = fmt.Sprintf("https://attendants.sdi.inet.co.th/3rd/employees/%s", staff_id)

	headers := map[string]string{
		"Authorization": "Bearer 7U2F2narFeGQAYN4Toze/GiJ8FupjH72glxi6rhzaqQt+whObLJ3fR2ztPQljc/0sY0=",
		"Content-Type":  "application/json",
	}

	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
	}{
		To: "198008320896",
	})
	rawResponse, err := requests.Get(url, headers, bytes.NewBuffer(payload), 5)
	if err != nil {
		return "", err
	}
	var dataResult interface{}
	if err := json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
		return "", err
	}
	result_total := dataResult.(map[string]interface{})
	if result_total["count"].(float64) < 1 {
		return "", nil
	}
	result_info := result_total["data"].(map[string]interface{})
	var account_id = result_info["account_id"].(string)
	return account_id, nil
}

func SendAlertToOneChat(account_id string, text_message string) (string, error) {
	// alert to one chat
	url := "https://chat-api.one.th/message/api/v1/push_message"

	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
		BotId              string `json:"bot_id"`
		Type               string `json:"type"`
		Message            string `json:"message"`
		CustomNotification string `json:"custom_notification"`
	}{
		To:                 account_id,
		BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		Type:               "text",
		Message:            text_message,
		CustomNotification: "เปิดอ่านข้อความใหม่จากทางเรา",
	})

	headers := map[string]string{
		"Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",
		"Content-Type":  "application/json",
	}

	rawResponse, err := requests.Post(url, headers, bytes.NewBuffer(payload), 5)
	if err != nil {
		log.Errorln("Error AlertOnechat", err)
		return "", err
	}
	var dataResult interface{}
	if err = json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
		log.Errorln("error -:", err)
		return "", err
	}
	return "send success", nil
}
