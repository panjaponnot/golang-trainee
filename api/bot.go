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
	var emp []attendant.EmployeeDetail
	if data.Status == "success" {
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
	}
	return c.JSON(http.StatusOK, emp[0])
}

func CheckAlertExpireEndPoint(c echo.Context) error {
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Mess: "not have param one_id"})
	}

	var staff []m.StaffInfo
	if err := dbSale.Ctx().Raw(`SELECT distinct one_id as one_id,staff_id,fname,lname,nname,position,department,staff_child,mail from staff_info
	left join staff_mail on staff_info.staff_id = staff_mail.ref_staff;`).Scan(&staff).Error; err != nil {
		log.Errorln("GetStaff Select Staff error :-", err)
	}

	// log.Infoln(staff)
	// for _, s := range staff {
	// 	if s.StaffChild == "" {
	// 		log.Infoln("-----------normal------------")
	// 		GetSoExpire(s)
	// 	} else {
	// 		res := ApiGetTeammember(s, staff)
	// 		if res.Type == "teamlead" {
	// 			log.Infoln(res.Type)
	// 			GetSoExpireLead(s)
	// 		} else if res.Type == "error" {
	// 			GetSoExpire(s)
	// 		} else {
	// 			if s.Department != "Based Sales" {
	// 				// s.StaffChild = res.DepartmentTeam
	// 				s.DepartmentChild = res.Dept
	// 				log.Infoln("group lead", s.Department)
	// 				GetsoExpireGroupLead(s)
	// 			} else {
	// 				// s.StaffChild = res.DepartmentTeam
	// 				s.DepartmentChild = res.Dept
	// 				log.Infoln(s.StaffChild)
	// 				GetsoExpireGroupLead(s)
	// 			}
	// 		}
	// 	}
	// 	CreateSoExpireApprove(s)
	// }

	return c.JSON(http.StatusOK, nil)
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

func GetSoExpire(s m.StaffInfo) error {
	var staff []m.StaffInfo
	if err := dbSale.Ctx().Raw(`select sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,
        DATEDIFF(ContractEndDate, NOW()) as days
        from so_mssql
        WHERE
        DATEDIFF(ContractEndDate, NOW()) <= 90 and has_refer = 0 and Active_Inactive = 'Active' and
        sale_code = ?
        group by sonumber
        order by days;`, s.StaffId).Scan(&staff).Error; err != nil {
		log.Errorln("GetSoExpire Select Staff error :-", err)

	}
	return nil
}

func ApiGetTeammember(s m.StaffInfo, staff []m.StaffInfo) Respons {
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

func GetSoExpireLead(s m.StaffInfo) error {
	var ListSo []m.SaleOrder
	if err := dbSale.Ctx().Raw(`select sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,
        DATEDIFF(ContractEndDate, NOW()) as days, fname, lname, nname, department
        from so_mssql
        left join staff_info on so_mssql.sale_code = staff_info.staff_id
        WHERE
        DATEDIFF(ContractEndDate, NOW()) <= 90 and has_refer = 0 and Active_Inactive = 'Active' and
        sale_code in (?)
        group by sonumber
        order by days;`, s.StaffChild).Scan(&ListSo).Error; err != nil {
		log.Errorln("GetSoExpireLead Select Staff error :-", err)
	}
	SetFormatLead(ListSo, s)
	return nil
}
func SetFormatLead(ListSo []m.SaleOrder, s m.StaffInfo) error {

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

	// AlertOnechat(TextMessage, s, len(ListSo))
	return nil
}
