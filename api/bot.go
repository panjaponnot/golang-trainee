package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sale_ranking/core"
	m "sale_ranking/model"
	"sale_ranking/pkg/attendant"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/requests"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func GetUserOneThEndPoint(c echo.Context) error {

	token := c.QueryParam("token")

	if strings.TrimSpace(c.QueryParam("token")) == "" {
		return echo.ErrBadRequest
	}

	url := "https://chat-manage.one.th:8997/api/v1/getprofile"
	headers := map[string]string{
		echo.HeaderContentType:   "application/json",
		echo.HeaderAuthorization: fmt.Sprintf("Bearer %s", util.GetEnv("WEB_HOOK_CHAT_TOKEN", "")),
	}
	body, _ := json.Marshal(&struct {
		BotId  string `json:"bot_id"`
		Source string `json:"source"`
		Phone  string `json:"phone"`
	}{
		BotId:  util.GetEnv("WEB_HOOK_CHAT_BOT_ID", ""),
		Source: token,
		Phone:  "true",
	})

	r, err := requests.Request(http.MethodPost, url, headers, bytes.NewBuffer(body), 0)
	// r, err := requests.Post(url, headers, bytes.NewBuffer(body), 5)
	if err != nil {
		log.Errorln(pkgName, err, "service chat unavailable")
		return echo.ErrServiceUnavailable
	}

	type Raw struct {
		OneId           string `json:"one_id"`
		EmployeeeDetail string `josn:"employee_detail"`
	}
	data := struct {
		Data   Raw    `json:"data"`
		Status string `json:"status"`
	}{}
	dataByte, err := json.Marshal(r)
	if err := json.Unmarshal(dataByte, &data); err != nil {
		log.Errorln(pkgName, err, "Json unmarshall chat error")
		return echo.ErrInternalServerError
	}
	// return c.JSON(http.StatusOK, data)
	var emp []attendant.EmployeeDetail
	if data.Status == "success" {
		// return c.JSON(http.StatusOK, "emp33333")
		a := core.AttendantClient()
		acc, err := a.GetAccountByID(data.Data.OneId)
		if err != nil {
			log.Errorln(pkgName, err, "service attendant unavailable")
			return echo.ErrServiceUnavailable
		}
		if len(acc.EmployeeDetail) != 0 {
			emp = append(emp, acc.EmployeeDetail[0])
		} else {
			return echo.ErrNotFound
		}
	} else {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid token"})
	}
	// return c.JSON(http.StatusOK, emp)
	return c.JSON(http.StatusOK, emp[0])
}

func AlertExpireEndPoint(c echo.Context) error {
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Mess: "not have param one_id"})
	}
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	var staff []StaffInfo
	if err := dbSale.Ctx().Raw(`SELECT distinct one_id as one_id,staff_id,fname,lname,nname,position,department,staff_child,mail from staff_info
	left join staff_mail on staff_info.staff_id = staff_mail.ref_staff;`).Scan(&staff).Error; err != nil {
		log.Errorln("GetStaff Select Staff error :-", err)
	}
	StrOneId := strings.TrimSpace(c.QueryParam("one_id"))
	log.Infoln("-->", staff)
	// return c.JSON(http.StatusOK, staff)
	if len(staff) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "sending fail"})
	}
	for _, s := range staff {

		if s.OneId == StrOneId {
			if s.StaffChild == "" {
				// return c.JSON(http.StatusOK, "staff")
				// log.Infoln("-----------normal------------")
				GetSoExpire(c, s)
			} else {
				// return c.JSON(http.StatusOK, "staff")
				res := ApiGetTeammember(s, staff)
				if res.Type == "teamlead" {
					// log.Infoln(res.Type)
					GetSoExpireLead(c, s)
				} else if res.Type == "error" {
					GetSoExpire(c, s)
				} else {
					if s.Department != "Based Sales" {
						// s.StaffChild = res.DepartmentTeam
						s.DepartmentChild = res.Dept
						// log.Infoln("group lead", s.Department)
						GetsoExpireGroupLead(c, s)
					} else {
						// s.StaffChild = res.DepartmentTeam
						s.DepartmentChild = res.Dept
						// log.Infoln(s.StaffChild)
						GetsoExpireGroupLead(c, s)
					}
				}
			}
			CreateSoExpireApprove(c, s)
			return c.JSON(http.StatusOK, m.Result{Message: "sending success"})
		}

	}
	return c.JSON(http.StatusBadRequest, m.Result{Message: "sending fail"})
	// return c.JSON(http.StatusOK, nil)
}

type Respons struct {
	Type string `json:"type"`
	Dept []Dept `json:"departments"`
}

type Dept struct {
	Name  string    `json:"name"`
	Staff []StaffId `json:"staff"`
}

type StaffId struct {
	StaffId string `json:"staff_id"`
}

type StaffOrg struct {
	StaffId string `json:"staff"`
}

type StaffInfo struct {
	OneId   string `json:"one_id"`
	StaffId string `json:"staff_id"`
	// Prefix          string `json:"prefix"`
	Fname           string `json:"fname"`
	Lname           string `json:"lname"`
	Nname           string `json:"nname"`
	Position        string `json:"position"`
	Department      string `json:"department"`
	StaffChild      string `json:"staff_child"`
	DepartmentChild []Dept `json:"department_child"`
	Mail            string `json:"mail"`
}

type SaleOrder struct {
	Id           string `json:"id" gorm:"column:id"`
	SoNumber     string `json:"sonumber" gorm:"column:sonumber"`
	CustomerID   string `json:"Customer_ID" gorm:"column:Customer_ID"`
	CustomerName string `json:"Customer_Name" gorm:"column:Customer_Name"`
	SaleCode     string `json:"sale_code" gorm:"column:sale_code"`
	SaleLead     string `json:"sale_lead" gorm:"column:sale_lead"`
	SoRefer      string `json:"so_refer" gorm:"column:so_refer"`
	SoTeam       string `json:"so_team" gorm:"column:so_team"`
	FirstName    string `json:"fname" gorm:"column:fname"`
	LastName     string `json:"lname" gorm:"column:lname"`
	NickName     string `json:"nname" gorm:"column:nname"`
	Department   string `json:"department" gorm:"column:department"`
	Days         string `json:"days" gorm:"column:days"`
}

type SoExpireApprove struct {
	ApproveTime time.Time `json:"approve_time"`
	Approve     int       `json:"approve"`
	OneId       string    `json:"one_id"`
}

type SoExpireApproveOneId struct {
	Approve   string `json:"approve"`
	Id        string `json:"id"`
	FirstName string `json:"fname"`
	LastName  string `json:"lname"`
	NickName  string `json:"nname"`
	StaffId   string `json:"staff_id"`
	OneId     string `json:"one_id"`
}

type SoMssql struct {
	SoNumber          string `json:"sonumber"`
	CustomerID        string `json:"Customer_ID"`
	CustomerName      string `json:"Customer_Name"`
	ContractStartDate string `json:"ContractStartDate"`
	ContractEndDate   string `json:"ContractEndDate"`
	SoRefer           string `json:"so_refer"`
	SaleCode          string `json:"sale_code"`
	SaleLead          string `json:"sale_lead"`
	Days              string `json:"days"`
	FirstName         string `json:"fname"`
	LastName          string `json:"lname"`
	NickName          string `json:"nname"`
	Department        string `json:"department"`
}

type SalesApprove struct {
	DocNumberEform string `json:"doc_number_eform"`
	Reason         string `json:"	reason"`
	Status         string `json:"	status"`
}

type QuatationTh struct {
	EformId          string `json:"eform_id"`
	PplDocumentId    string `json:"ppl_document_id"`
	DocNumberEform   string `json:"doc_number_eform"`
	DatetimeApprove  string `json:"datetime_approve"`
	CsNumber         string `json:"cs_number"`
	QuotationNo      string `json:"quotation_no"`
	Service          string `json:"service"`
	Type             string `json:"type"`
	EmployeeCode     string `json:"employee_code"`
	Title            string `json:"title"`
	Salename         string `json:"salename"`
	Datetime         string `json:"datetime"`
	CompanyName      string `json:"company_name"`
	Team             string `json:"team"`
	Total            string `json:"total"`
	TotalStr         string `json:"total_str"`
	TotalDiscount    string `json:"total_discount"`
	TotalDiscountStr string `json:"total_discount_str"`
	Head             string `json:"head"`
	DurationTime     string `json:"duration_time"`
	Unit             string `json:"unit"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	TypePayment      string `json:"type_payment"`
	RefQuotation     string `json:"ref_quotation"`
	Status           string `json:"status"`
	DateCancel       string `json:"date_cancel"`
	CreatedAt        string `json:"created_at"`
	RefSO            string `json:"refSO"`
	StatusQt         string `json:"status_qt"`
	RefSORevenue     string `json:"refSO_revenue"`
	StatusQuotation  string `json:"status_quotation"`
	ServicQlatform   string `json:"service_platform"`
}

func GetSoExpire(c echo.Context, s StaffInfo) error {
	var staff []SaleOrder
	if err := dbSale.Ctx().Raw(`select sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,
        DATEDIFF(ContractEndDate, NOW()) as days
        from so_mssql
        WHERE
        DATEDIFF(ContractEndDate, NOW()) <= 90 and has_refer = 0 and Active_Inactive = 'Active' and
        sale_code = ?
        group by sonumber
        order by days;`, s.StaffId).Scan(&staff).Error; err != nil {
		log.Errorln("GetSoExpire Select Staff error :-", err)
		SetFormat(staff, s)
	}

	return nil
}

func SetFormat(ListSo []SaleOrder, s StaffInfo) error {

	var CheckDays = 0
	var NDays string
	var TextMessage = ""
	DictDays := make(map[string][]SaleOrder)

	for _, so := range ListSo {
		for key, _ := range DictDays { //เช็คarray
			if so.Days == key { //ถ้ามีอยู่ในArray
				CheckDays += 1
				NDays = key
			} else {
				CheckDays += 0
			}
		}
		if CheckDays > 0 {
			if _, ok := DictDays[NDays]; !ok {
				DictDays[NDays] = append(DictDays[NDays], so)
				NDays = ""
			}
		} else {
			NDays = so.Days
			DictDays[NDays] = append(DictDays[NDays], so)
			NDays = ""
		}
		CheckDays = 0
	}

	for key, v := range DictDays {
		var ListText []string
		var CheckText = 0
		KeyId, err := strconv.ParseInt(strings.TrimSpace(key), 10, 32)
		if err != nil {
			log.Errorln("error :-", err)
		}
		if KeyId >= 60 {
			TextMessage += fmt.Sprintf("💚อีก %d วัน จะหมดสัญญา ( %d SO)\n", KeyId, len(v))
		} else if KeyId >= 30 {
			TextMessage += fmt.Sprintf("💛อีก %d วัน จะหมดสัญญา ( %d SO)\n", KeyId, len(v))
		} else {
			TextMessage += fmt.Sprintf("❤️อีก %d วัน จะหมดสัญญา ( %d SO)\n", KeyId, len(v))
		}
		for _, so := range v {
			for _, l := range ListText {
				//if so['nname'] not in list_text
				if so.NickName == l {
					CheckText += 1
				} else {
					CheckText += 0
				}
			}
			if CheckText == 0 {
				ListText = append(ListText, so.NickName)
				// str := fmt.Sprintf("%s%s", so.NickName, so.Days)
				TextMessage += fmt.Sprintf("- %s (%s)\n", so.CustomerName, so.SoNumber)
			}
		}
	}

	AlertOnechat(TextMessage, s, len(ListSo))
	return nil
}

func ApiGetTeammember(s StaffInfo, staff []StaffInfo) Respons {
	// url := fmt.Sprintf("https://attendants.sdi.inet.co.th/3rd/accounts/%s/teammember", s.OneId)

	// payload, _ := json.Marshal(&struct {
	// 	// TaxId string `json:"tax_id"`
	// }{
	// 	// TaxId: "0107544000094",
	// })

	// headers := map[string]string{
	// 	"Authorization": "Bearer 7U2F2narFeGQAYN4Toze/GiJ8FupjH72glxi6rhzaqQt+whObLJ3fR2ztPQljc/0sY0=",
	// }

	// rawResponse, err := Get(url, headers, bytes.NewBuffer(payload), 5)
	// if err != nil {
	// 	log.Errorln("Error GetStaffTeammember 1", err)
	// }

	a := core.AttendantClient()
	rawResponse, err := a.GetTeamMemberByID(s.OneId)
	if err != nil {
		log.Errorln(pkgName, err, "service attendant unavailable")
		return Respons{}
	}
	// var Respons Respons
	var Departments []Dept
	var HasAcc []StaffId
	var dataResult []attendant.CompanyDetail
	// var accIdTeam []string
	if len(rawResponse) != 0 {
		for _, c := range dataResult {
			if len(c.Departments) == 1 {
				res := Respons{
					Type: "teamlead",
				}
				return res
			}
			for _, dept := range c.Departments {
				var DeptData Dept
				DeptData.Name = dept.DeptName
				var ListData []string
				var ListDataChild []string
				for _, acc := range dept.HasAccount {
					for _, staffs := range staff {
						if staffs.StaffId == acc.EmployeeID {
							ListDataChild = strings.Split(staffs.StaffChild, ",")
							for _, LtDatChd := range ListDataChild {
								ListData = append(ListData, LtDatChd)
							}
							ListData = append(ListData, staffs.StaffId)
						}

					}
				}

				for _, LstDta := range ListData {

					if LstDta != "" {
						h := StaffId{
							StaffId: LstDta,
						}
						HasAcc = append(HasAcc, h)
					}
				}
				DeptData.Staff = HasAcc
				Departments = append(Departments, DeptData)
			}
			// departments.append({'name': '', 'staff': ['']})
			d := Dept{
				Name:  "",
				Staff: nil,
			}
			Departments = append(Departments, d)
			res := Respons{
				Type: "grouplead",
				Dept: Departments,
			}
			return res
		}
		res := Respons{
			Type: "no",
		}
		return res
	} else {
		res := Respons{
			Type: "error",
		}
		return res
	}

}

func GetSoExpireLead(c echo.Context, s StaffInfo) error {
	var ListSo []m.SaleOrder
	StaffChild := strings.Split(s.StaffChild, ",")
	var StaffChildStr string
	for n, s := range StaffChild {
		if n == 0 {
			StaffChildStr += fmt.Sprintf("'%s',", s)
		} else if n+1 == len(StaffChild) {
			StaffChildStr += fmt.Sprintf("'%s'", s)
		} else {
			StaffChildStr += fmt.Sprintf("'%s',", s)
		}

	}
	str := fmt.Sprintf(`select sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,
	DATEDIFF(ContractEndDate, NOW()) as days, fname, lname, nname, department
	from so_mssql
	left join staff_info on so_mssql.sale_code = staff_info.staff_id
	WHERE
	DATEDIFF(ContractEndDate, NOW()) <= 90 and has_refer = 0 and Active_Inactive = 'Active' and
	sale_code in (%s)
	group by sonumber
	order by days;`, StaffChildStr)
	if err := dbSale.Ctx().Raw(str).Scan(&ListSo).Error; err != nil {
		log.Errorln("GetSoExpireLead Select Staff error :-", err)
	}
	SetFormatLead(ListSo, s)
	return nil
}
func SetFormatLead(ListSo []m.SaleOrder, s StaffInfo) error {

	var CheckDays = 0
	var CheckName = 0
	var NDays string
	var NName string
	var TextMessage = ""
	// var DictDays map[string]SaleOrder
	DictDays := make(map[string][]m.SaleOrder)
	DictName := make(map[string][]m.SaleOrder)
	for _, so := range ListSo {
		//เช็คarray
		for key, _ := range DictDays {
			if so.Days == key { //ถ้ามีอยู่ในArray
				CheckDays += 1
				NDays = key
			} else {
				CheckDays += 0
			}
		}
		str := fmt.Sprintf("%s%s", so.NickName, so.Days)
		for key, _ := range DictName {
			if key == str { //ถ้ามีอยู่ในArray
				CheckName += 1
				NName = str
			} else {
				CheckName += 0
			}
		}
		if CheckDays > 0 {
			if _, ok := DictDays[NDays]; !ok {
				// DictDays[NDays] = so
				DictDays[NDays] = append(DictDays[NDays], so)
				NDays = ""
			}
		} else {
			// DictDays[so.Days] = ""
			NDays = so.Days
			// DictDays[NDays] = so
			DictDays[NDays] = append(DictDays[NDays], so)
			NDays = ""
		}
		if CheckName > 0 {
			if _, ok := DictName[NName]; !ok {
				// d.key = so
				// DictName[NName] = so
				DictName[NName] = append(DictName[NName], so)
				NName = ""
			}
		} else {
			// DictName[str] = SaleOrder{}
			// NName = str
			// DictName[str] = so
			DictName[NName] = append(DictName[NName], so)
			// NName = ""
		}
		CheckDays = 0
		CheckName = 0
	}

	for key, v := range DictDays {
		var ListText []string
		var CheckText = 0
		KeyId, err := strconv.ParseInt(strings.TrimSpace(key), 10, 32)
		if err != nil {
			log.Errorln(pkgName, err)
		}
		if KeyId >= 60 {
			TextMessage += fmt.Sprintf("💚อีก %d วัน จะหมดสัญญา ( %d SO)\n", KeyId, len(v))
		} else if KeyId >= 30 {
			TextMessage += fmt.Sprintf("💛อีก %d วัน จะหมดสัญญา ( %d SO)\n", KeyId, len(v))
		} else {
			TextMessage += fmt.Sprintf("❤️อีก %d วัน จะหมดสัญญา ( %d SO)\n", KeyId, len(v))
		}
		for _, so := range v {
			for _, l := range ListText {
				//if so['nname'] not in list_text
				if so.NickName == l {
					CheckText += 1
				} else {
					CheckText += 0
				}
			}
			if CheckText == 0 {
				ListText = append(ListText, so.NickName)
				str := fmt.Sprintf("%s%s", so.NickName, so.Days)
				TextMessage += fmt.Sprintf("- %s %s (SO %d ใบ)\n", so.SaleCode, so.NickName, len(DictName[str]))
			}
		}
	}

	AlertOnechat(TextMessage, s, len(ListSo))
	return nil
}

func GetsoExpireGroupLead(c echo.Context, s StaffInfo) error {
	// if err := Connect(host, port, dbUser, dbPass, dbName, false); err != nil {
	// 	log.Errorln("connect error :- ", err)
	// 	defer Close()
	// }
	// defer Close()
	StaffChild := strings.Split(s.StaffChild, ",")
	var StaffChildStr string
	for n, s := range StaffChild {
		if n == 0 {
			StaffChildStr += fmt.Sprintf("'%s',", s)
		} else if n+1 == len(StaffChild) {
			StaffChildStr += fmt.Sprintf("'%s'", s)
		} else {
			StaffChildStr += fmt.Sprintf("'%s',", s)
		}

	}
	var ListSo []SaleOrder
	// var ListSo []interface{}
	str := fmt.Sprintf(`select sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,
	DATEDIFF(ContractEndDate, NOW()) as days, fname, lname, nname, department
	from so_mssql
	left join staff_info on so_mssql.sale_code = staff_info.staff_id
	WHERE
	DATEDIFF(ContractEndDate, NOW()) <= 90 and has_refer = 0  and Active_Inactive = 'Active' and
	sale_code in (%s)
	group by sonumber
	order by days;`, StaffChildStr)
	if err := dbSale.Ctx().Raw(str).Scan(&ListSo).Error; err != nil {
		log.Errorln("GetsoExpireGroupLead Select Staff error :-", err)
	}
	// return c.JSON(http.StatusOK, StaffChildStr)
	SetFormatGroup(ListSo, s)
	return nil
}

func SetFormatGroup(ListSo []SaleOrder, s StaffInfo) error {
	var CheckDays = 0
	var CheckName = 0
	var CheckStaff = 0
	var NDays string
	var NName string
	var TextMessage = ""
	// var DictDays map[string]SaleOrder
	DictDays := make(map[string][]SaleOrder)
	DictName := make(map[string][]SaleOrder)
	for _, so := range ListSo {
		//เช็คarray
		so.SoTeam = s.Department
		for _, StaffDept := range s.DepartmentChild {

			for _, ds := range StaffDept.Staff {
				if so.SaleCode == ds.StaffId { //ถ้ามีอยู่ในArray
					CheckStaff += 1
				} else {
					CheckStaff += 0
				}
			}
			if CheckStaff > 0 {
				so.SoTeam = StaffDept.Name
			}
		}
		for key, _ := range DictDays {
			if so.Days == key { //ถ้ามีอยู่ในArray
				CheckDays += 1
				NDays = key
			} else {
				CheckDays += 0
			}
		}
		str := fmt.Sprintf("%s%s", so.SoTeam, so.Days)
		for key, _ := range DictName {
			if key == str { //ถ้ามีอยู่ในArray
				CheckName += 1
				NName = str
			} else {
				CheckName += 0
			}
		}
		if CheckDays > 0 {
			if _, ok := DictDays[NDays]; !ok {
				DictDays[NDays] = append(DictDays[NDays], so)
				NDays = ""
			}
		} else {
			NDays = so.Days
			DictDays[NDays] = append(DictDays[NDays], so)
			NDays = ""
		}
		if CheckName > 0 {
			if _, ok := DictName[NName]; !ok {
				DictName[NName] = append(DictName[NName], so)
				NName = ""
			}
		} else {
			DictName[NName] = append(DictName[NName], so)
		}
		CheckStaff = 0
		CheckDays = 0
		CheckName = 0
	}

	for key, v := range DictDays {
		var ListText []string
		var CheckText = 0
		KeyId, err := strconv.ParseInt(strings.TrimSpace(key), 10, 32)
		if err != nil {
			log.Errorln("error -:", err)
		}
		if KeyId >= 60 {
			TextMessage += fmt.Sprintf("💚อีก %d วัน\n", KeyId)
		} else if KeyId >= 30 {
			TextMessage += fmt.Sprintf("💛อีก %d วัน\n", KeyId)
		} else {
			TextMessage += fmt.Sprintf("❤️อีก %d วัน\n", KeyId)
		}
		for _, so := range v {
			for _, l := range ListText {
				//if so['nname'] not in list_text
				if so.NickName == l {
					CheckText += 1
				} else {
					CheckText += 0
				}
			}
			if CheckText == 0 {
				ListText = append(ListText, so.SoTeam)
				str := fmt.Sprintf("%s%s", so.SoTeam, so.Days)
				TextMessage += fmt.Sprintf("- %s (%d SO)\n", so.SoTeam, len(DictName[str]))
			}
			CheckText = 0
		}
	}

	AlertOnechat(TextMessage, s, len(ListSo))
	return nil
}

func AlertOnechat(TextMessage string, staff StaffInfo, l int) (interface{}, error) {

	if TextMessage == "" {
		TextMessage = "ไม่มี SO ใกล้หมดอายุภายใน 90 วันครับ"
	} else {
		TextMessage = fmt.Sprintf("แจ้งเตือนต่อสัญญา (ทั้งหมด %d  ลูกค้า) \n %s", l, TextMessage)
	}

	url := "https://chat-api.one.th/message/api/v1/push_message"

	payload, _ := json.Marshal(&struct {
		To                 string `json:"to"`
		BotId              string `json:"bot_id"`
		Type               string `json:"type"`
		Message            string `json:"message"`
		CustomNotification string `json:"custom_notification"`
	}{
		// To: "198008320896",
		To: "25078584384",
		// To:                 staff.OneId,
		// BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
		BotId:              "Becf3d73c867f508ab7a8f5d62ceceb64",
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
	// test
	// ApiPushQuickReply("198008320896")
	ApiPushQuickReply("25078584384")
	// ApiPushQuickReply(staff.OneId)
	return dataResult, nil
}

func ApiPushQuickReply(OneId string) (interface{}, error) {
	url := "https://chat-api.one.th/message/api/v1/push_quickreply"

	type Quick struct {
		Label   string `json:"label"`
		Type    string `json:"type"`
		Message string `json:"message"`
		Payload bool   `json:"payload"`
	}

	body, _ := json.Marshal(&struct {
		To         string  `json:"to"`
		BotId      string  `json:"bot_id"`
		Message    string  `json:"message"`
		QuickReply []Quick `json:"quick_reply"`
	}{
		// To: "198008320896",
		// To:      OneId,
		To: "25078584384",
		// BotId:   "B4f7385bc7ee356c89f3560795eeb8067",
		BotId:   "Becf3d73c867f508ab7a8f5d62ceceb64",
		Message: "",
		QuickReply: []Quick{{
			Label:   "รับทราบ",
			Type:    "text",
			Message: "รับทราบ",
			Payload: true,
		}},
	})

	headers := map[string]string{
		"Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",
		"Content-Type":  "application/json",
	}
	rawResponse, err := requests.Post(url, headers, bytes.NewBuffer(body), 5)
	if err != nil {
		log.Errorln("error :-", err)
	}
	var dataResult interface{}
	if err := json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
		log.Errorln("unmarshal staff one ra error :- ", err)
	}

	return dataResult, nil

}

//เก็บ status ของ sale เพื่อให้ทราบว่ามีใครรับรู้ หรือไม่รับรู้กี่คน
func CreateSoExpireApprove(c echo.Context, s StaffInfo) error {

	dt := time.Now()
	data := SoExpireApprove{
		ApproveTime: dt,
		Approve:     0,
		OneId:       s.OneId,
	}
	if err := dbSale.Ctx().Table("so_expire_approve").Create(&data).Error; err != nil {
		log.Errorln("Func Insert error :-", err)
		return err
	}

	return nil
}

func AlertApproveAllEndPoint(c echo.Context) error {
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Mess: "not have param one_id"})
	}
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	var SoExpireApproveOneId []SoExpireApproveOneId
	OneId := strings.TrimSpace(c.QueryParam("one_id"))
	var staff []StaffInfo
	if err := dbSale.Ctx().Raw(`select * from staff_info where one_id = ? and staff_child <> '';`, OneId).Scan(&staff).Error; err != nil {
		log.Errorln("GetStaff Select Staff error :-", err)
	}
	// return c.JSON(http.StatusOK, staff)
	if len(staff) == 0 {
		PushMessageApprove(SoExpireApproveOneId, OneId, "normal")
		return c.JSON(http.StatusBadRequest, m.Result{Message: "sending fail"})
	}
	StaffChild := strings.Split(staff[0].StaffChild, ",")
	var StaffChildStr string
	for n, s := range StaffChild {
		if n == 0 {
			StaffChildStr += fmt.Sprintf("%s',", s)
		} else if n+1 == len(StaffChild) {
			StaffChildStr += fmt.Sprintf("'%s", s)
		} else {
			StaffChildStr += fmt.Sprintf("'%s',", s)
		}
	}
	str := fmt.Sprintf(`SELECT approve,so_expire_approve.id,fname,lname, nname,staff_id,so_expire_approve.one_id
	FROM so_expire_approve
	LEFT JOIN (
		select * from staff_info where staff_id in (%s)
	) staff_info ON so_expire_approve.one_id = staff_info.one_id
	WHERE so_expire_approve.id in (select max(id) as id from so_expire_approve group by one_id) and staff_id is not null`, StaffChildStr)
	if err := dbSale.Ctx().Raw(str).Scan(&SoExpireApproveOneId).Error; err != nil {
		log.Errorln("GetStaff Select Staff error :-", err)
	}
	PushMessageApprove(SoExpireApproveOneId, OneId, "normal")
	return c.JSON(http.StatusOK, m.Result{Message: "sending success"})
}

func PushMessageApprove(so []SoExpireApproveOneId, OneId string, role string) error {

	var StrData string
	var NewStrData string
	Count := 0
	for _, s := range so {
		Approve := s.Approve
		Nname := s.NickName
		if Approve == "1" {
			Approve = "รับทราบแล้ว✅"
			Count++
			StrData += fmt.Sprintf("✅ %s %s\n", s.StaffId, Nname)
		} else {
			Approve = "ยังไม่รับทราบ"
			StrData += fmt.Sprintf("❌ %s %s\n", s.StaffId, Nname)
		}
	}
	NewStrData = fmt.Sprintf("แจ้งเตือนการ Approve \nApprove แล้ว(%d คน)\n\n", Count)
	NewStrData += StrData

	if role == "lead" {
		url := "https://chat-api.one.th/message/api/v1/push_message"

		payload, _ := json.Marshal(&struct {
			To                 string `json:"to"`
			BotId              string `json:"bot_id"`
			Type               string `json:"type"`
			Message            string `json:"message"`
			CustomNotification string `json:"custom_notification"`
		}{
			// To: "198008320896",
			To: "25078584384",
			// To:                 OneId,
			// BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
			BotId:              "Becf3d73c867f508ab7a8f5d62ceceb64",
			Type:               "text",
			Message:            NewStrData,
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
	} else {
		url := "https://chat-api.one.th/message/api/v1/push_message"

		payload, _ := json.Marshal(&struct {
			To                 string `json:"to"`
			BotId              string `json:"bot_id"`
			Type               string `json:"type"`
			Message            string `json:"message"`
			CustomNotification string `json:"custom_notification"`
		}{
			// To: "198008320896",25078584384
			To: "25078584384",
			// To:                 OneId,
			// BotId:              "B4f7385bc7ee356c89f3560795eeb8067",
			BotId:              "Becf3d73c867f508ab7a8f5d62ceceb64",
			Type:               "text",
			Message:            "ไม่มีสิทธิ์ในการเข้าถึง",
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
	}
	return nil
}

func AlertApproveEndPoint(c echo.Context) error {
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Mess: "not have param one_id"})
	}
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	tx := dbSale.Ctx().Begin()
	OneId := strings.TrimSpace(c.QueryParam("one_id"))
	if err := dbSale.Ctx().Exec(`
		UPDATE so_expire_approve
			SET approve = 1
		WHERE approve_time = ? AND one_id = ?
		;`, time.Now().Format("2006-01-02"), OneId).Error; err != nil {
		// log.Errorln("UPDATE so_expire_approve error :-", err)
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, m.Result{Message: err})
	}
	return c.JSON(http.StatusOK, m.Result{Message: "sending success"})
}

func CheckQuotationEndPoint(c echo.Context) error {
	data := struct {
		OneId          string `json:"one_id"`
		DocNumberEform string `json:"doc_number_eform"`
		Reason         string `json:"reason"`
		Status         string `json:"status"`
	}{}
	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}

	if data.OneId == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Mess: "not have param one_id"})
	}
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	// tx := dbSale.Ctx().Begin()
	OneId := data.OneId

	// BaseTime := time.Now()
	// FromDate := BaseTime.Add(-20 * 24 * time.Hour)

	var staff []StaffInfo
	if err := dbSale.Ctx().Raw(`select * from staff_info where one_id = ? and staff_child <> '';`, OneId).Scan(&staff).Error; err != nil {
		log.Errorln("GetStaff Select Staff error :-", err)
	}
	if len(staff) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot find one id in system"})
	}
	EmployeeId := staff[0].StaffId
	DocNumberEform := data.DocNumberEform
	Status := data.Status
	Reason := data.Reason
	var Text string
	tx := dbSale.Ctx().Begin()
	var SalesApprove []SalesApprove
	if err := dbSale.Ctx().Raw(`select * from sales_approve where doc_number_eform = ?`, DocNumberEform).Scan(&SalesApprove).Error; err != nil {
		log.Errorln("GetStaff Select Staff error :-", err)
	}
	if len(SalesApprove) == 0 {
		if err := dbSale.Ctx().Exec(`
		insert into sales_approve (doc_number_eform, reason, status)
            values ('?', '?', '?')
		;`, DocNumberEform, Reason, Status).Error; err != nil {
			// log.Errorln("UPDATE so_expire_approve error :-", err)
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, m.Result{Message: err})
		}
	}
	var DatabaseData []QuatationTh
	if err := dbSale.Ctx().Raw(`select * from quatation_th as qt left outer join sales_approve as sa on
	qt.doc_number_eform = sa.doc_number_eform
	where qt.employee_code = '?' and sa.doc_number_eform is null and qt.status_qt = 'Actual';`, EmployeeId).Scan(&DatabaseData).Error; err != nil {
		log.Errorln("GetStaff Select Staff error :-", err)
	}

	if len(DatabaseData) != 0 {
		TotalCost := CheckTotalCost(DatabaseData[0])
		if TotalCost == "" {
			TotalCost = "-"
		}
		Text += fmt.Sprintf("แจ้งเตือน Quotation จากทาง ? ราคาทั้งสิ้น ? บาท", DatabaseData[0].CompanyName, TotalCost)
		AlertOnechatQuo(DocNumberEform, Text, OneId)
	}
	return c.JSON(http.StatusOK, m.Result{Message: "sending success"})
}

func CheckTotalCost(quot QuatationTh) string {
	if quot.Total == "" {
		return quot.TotalDiscountStr
	} else {
		return quot.TotalStr
	}
}

func AlertOnechatQuo(QtNumber string, Message string, AccountId string) error {
	// To := "3148848982"
	To := AccountId
	url := "https://chat-api.one.th/message/api/v1/push_message"
	type Choices struct {
		Label   string `json:"label"`
		Type    string `json:"type"`
		Payload string `json:"payload"`
	}

	type Element struct {
		Image  string    `json:"image"`
		Title  string    `json:"title"`
		Detail string    `json:"detail"`
		Choice []Choices `json:"choice"`
	}

	payload, _ := json.Marshal(&struct {
		To       string    `json:"to"`
		BotId    string    `json:"bot_id"`
		Type     string    `json:"type"`
		CusNoti  string    `json:"custom_notification"`
		Elements []Element `json:"elements"`
	}{
		To: To,
		// BotId:   "B4f7385bc7ee356c89f3560795eeb8067",
		BotId:   "Becf3d73c867f508ab7a8f5d62ceceb64",
		Type:    "template",
		CusNoti: "เปิดอ่านข้อความใหม่จากทางเรา",
		Elements: []Element{{
			Image:  "",
			Title:  QtNumber,
			Detail: Message,
			Choice: []Choices{{
				Label:   "Win",
				Type:    "text",
				Payload: fmt.Sprintf("Win#%s", QtNumber),
			}, {
				Label:   "Lost",
				Type:    "text",
				Payload: fmt.Sprintf("Lost#%s", QtNumber),
			}, {
				Label:   "Resend",
				Type:    "text",
				Payload: fmt.Sprintf("Resend#%s", QtNumber),
			}},
		}},
	})

	headers := map[string]string{
		"Authorization": "Bearer A548a4dd47e3c5108affe99b48b5c0218db9bcaaca6b34470b389bd04a19c3e30e1b99dad38844be387e939f755d194be",
		// "Authorization": "Bearer A6ef7265bc6b057fabb531b9b0e4eeff6edb6086b1fe143ebb02523d72d7f2623421ead53c8e7497c89bd0694a7c469ef",

		"Content-Type": "application/json",
	}
	// log.Infoln("scsc1")
	_, err := requests.Post(url, headers, bytes.NewBuffer(payload), 5)
	if err != nil {
		log.Errorln("Error QuickReply", err)
		return err

	}
	// log.Infoln("scsc2", r.Code)
	return nil
	// var dataResult interface{}
	// if err := json.Unmarshal(rawResponse.Body, &dataResult); err != nil {
	// 	log.Panicln(err)
	// }
	// log.Infoln("scsc3")
	// return nil
}
