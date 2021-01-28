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
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code in empty"})
		}
	}
	staff_id := strings.TrimSpace(bodyData.EmployeeCode)
	// var user []m.UserInfo
	// if err := dbSale.Ctx().Raw(` SELECT * FROM user_info WHERE role = 'admin' AND staff_id = ? `, staff_id).Scan(&user).Error; err != nil {
	// 	log.Errorln(pkgName, err, "User Not Found")
	// 	if !gorm.IsRecordNotFoundError(err) {
	// 		log.Errorln(pkgName, err, "Select user Error")
	// 		return echo.ErrInternalServerError
	// 	}
	// }
	staff := struct {
		StaffId    string `json:"staff_id"`
		OneId			 string `json:"one_id"`
	}{}
	if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where staff_id = ?`, staff_id).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
		return c.JSON(http.StatusNotFound, m.Result{Message: "Staff Not Found"})
	}
	log.Infoln("staff found: ", staff)
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	bot_id := strings.Join(bodyData.OrderBotId, ", ")
	var TextMessage = fmt.Sprintf("เอกสาร %s อยู่ในคิว BOT ลำดับที่ %s", paperless_no, bot_id)

	url := "https://chat-api.one.th/message/api/v1/push_message"

	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
		BotId              string `json:"bot_id"`
		Type               string `json:"type"`
		Message            string `json:"message"`
		CustomNotification string `json:"custom_notification"`
	}{
		// To: "198008320896",
		To:                 staff.OneId,
		BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		Type:               "text",
		Message:            TextMessage,
		CustomNotification: "เปิดอ่านข้อความใหม่จากทางเรา",
	})

	headers := map[string]string{
		"Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",
		"Content-Type":  "application/json",
	}

	rawResponse, err := requests.Post(url, headers, bytes.NewBuffer(payload), 5)
	if err != nil {
		log.Errorln("Error AlertOnechat", err)
	}
	var dataResult interface{}
	if err := json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
		log.Errorln("error -:", err)
	}
	return c.JSON(http.StatusOK, m.Result{Message: "send success"})
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
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no in empty"})
		} else if bodyData.EntryDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "entry date is empty"})
		} else if bodyData.SoNumber == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "so number is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}
	staff_id := strings.TrimSpace(bodyData.EmployeeCode)

	staff := struct {
		StaffId    string `json:"staff_id"`
		OneId			 string `json:"one_id"`
	}{}
	if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where staff_id = ?`, staff_id).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
		return c.JSON(http.StatusNotFound, m.Result{Message: "Staff Not Found"})
	}
	log.Infoln("staff found: ", staff)
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nดำเนินการสำเร็จแล้ว เมื่อเวลา %s \nTerminate เลขที่ %s", paperless_no, bodyData.CustomerName, bodyData.EntryDate, bodyData.SoNumber)

	url := "https://chat-api.one.th/message/api/v1/push_message"

	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
		BotId              string `json:"bot_id"`
		Type               string `json:"type"`
		Message            string `json:"message"`
		CustomNotification string `json:"custom_notification"`
	}{
		// To: "198008320896",
		To:                 staff.OneId,
		BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		Type:               "text",
		Message:            TextMessage,
		CustomNotification: "เปิดอ่านข้อความใหม่จากทางเรา",
	})

	headers := map[string]string{
		"Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",
		"Content-Type":  "application/json",
	}

	rawResponse, err := requests.Post(url, headers, bytes.NewBuffer(payload), 5)
	if err != nil {
		log.Errorln("Error AlertOnechat", err)
	}
	var dataResult interface{}
	if err := json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
		log.Errorln("error -:", err)
	}
	return c.JSON(http.StatusOK, m.Result{Message: "send success"})
}

func AlertTerminateRunFailEndPoint(c echo.Context) error {
	bodyData := struct {
		PaperLessNo         []string		`json:"paperless_no"`
		CustomerName        string			`json:"customer_name"`
		EntryDate         	string			`json:"entry_date"`
		Emails        			[]string			`json:"emails"`
		Comment	        		string			`json:"comment"`
		EmployeeCode        string 			`json:"employee_code"`
	}{}
	if err := c.Bind(&bodyData); err != nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "data type not correct"})
	}
	if len(bodyData.PaperLessNo) == 0 || bodyData.CustomerName == "" || bodyData.EntryDate == "" || len(bodyData.Emails) == 0 || bodyData.EmployeeCode == "" || bodyData.Comment == "" {
		if len(bodyData.PaperLessNo) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "0 data in list paperless_no"})
		} else if bodyData.CustomerName == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no in empty"})
		} else if bodyData.EntryDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "entry date is empty"})
		} else if len(bodyData.Emails) == 0 {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "Email Business support is empty"})
		} else if bodyData.Comment == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "comment is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}
	staff_id := strings.TrimSpace(bodyData.EmployeeCode)

	staff := struct {
		StaffId    string `json:"staff_id"`
		OneId			 string `json:"one_id"`
	}{}
	if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where staff_id = ?`, staff_id).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
		return c.JSON(http.StatusNotFound, m.Result{Message: "Staff Not Found"})
	}
	log.Infoln("staff found: ", staff)
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	emails := strings.Join(bodyData.Emails, ", ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nที่ส่งมามีข้อผิดพลาด คือ %s เมื่อเวลา %s \nรบกวนติดต่อสอบถามฝ่าย Business Support เพื่อดำเนินการต่อ Email: %s", paperless_no, bodyData.CustomerName, bodyData.Comment, bodyData.EntryDate, emails)

	url := "https://chat-api.one.th/message/api/v1/push_message"

	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
		BotId              string `json:"bot_id"`
		Type               string `json:"type"`
		Message            string `json:"message"`
		CustomNotification string `json:"custom_notification"`
	}{
		// To: "198008320896",
		To:                 staff.OneId,
		BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		Type:               "text",
		Message:            TextMessage,
		CustomNotification: "เปิดอ่านข้อความใหม่จากทางเรา",
	})

	headers := map[string]string{
		"Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",
		"Content-Type":  "application/json",
	}

	rawResponse, err := requests.Post(url, headers, bytes.NewBuffer(payload), 5)
	if err != nil {
		log.Errorln("Error AlertOnechat", err)
	}
	var dataResult interface{}
	if err := json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
		log.Errorln("error -:", err)
	}
	return c.JSON(http.StatusOK, m.Result{Message: "send success"})
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
			return c.JSON(http.StatusBadRequest, m.Result{Message: "customer no in empty"})
		} else if bodyData.TerminateDate == "" {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "Terminate date is empty"})
		} else {
			return c.JSON(http.StatusBadRequest, m.Result{Message: "employee_code is empty"})
		}
	}
	staff_id := strings.TrimSpace(bodyData.EmployeeCode)

	staff := struct {
		StaffId    string `json:"staff_id"`
		OneId			 string `json:"one_id"`
	}{}
	if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where staff_id = ?`, staff_id).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
		return c.JSON(http.StatusNotFound, m.Result{Message: "Staff Not Found"})
	}
	log.Infoln("staff found: ", staff)
	paperless_no := strings.Join(bodyData.PaperLessNo, " หรือ ")
	var TextMessage = fmt.Sprintf("เอกสาร %s \nลูกค้า %s \nต้องการที่จะ Terminate และให้มีผลในวันที่ %s ตามเดิม \nรบกวนทำการลดหนี้ Invoice และนำส่งเอกสารไปที่ Paperless เพื่อทำการ Approve ใหม่อีกครั้ง", paperless_no, bodyData.CustomerName, bodyData.TerminateDate)

	url := "https://chat-api.one.th/message/api/v1/push_message"

	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
		BotId              string `json:"bot_id"`
		Type               string `json:"type"`
		Message            string `json:"message"`
		CustomNotification string `json:"custom_notification"`
	}{
		// To: "198008320896",
		To:                 staff.OneId,
		BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		Type:               "text",
		Message:            TextMessage,
		CustomNotification: "เปิดอ่านข้อความใหม่จากทางเรา",
	})

	headers := map[string]string{
		"Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",
		"Content-Type":  "application/json",
	}

	rawResponse, err := requests.Post(url, headers, bytes.NewBuffer(payload), 5)
	if err != nil {
		log.Errorln("Error AlertOnechat", err)
	}
	var dataResult interface{}
	if err := json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
		log.Errorln("error -:", err)
	}
	return c.JSON(http.StatusOK, m.Result{Message: "send success"})
}
