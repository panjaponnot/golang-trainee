package api

import (
	"encoding/base64"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"strings"

	"github.com/labstack/echo/v4"
)

func GetAllStaffEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer dbSale.Close()
	// var StaffAll []m.StaffAll
	StaffAll := struct {
		StaffId    string `json:"staff_id"`
		Prefix     string `json:"prefix"`
		Fname      string `json:"fname"`
		Lname      string `json:"lname"`
		Nname      string `json:"nname"`
		StaffImage string `json:"staff_image"`
	}{}
	sql := `SELECT staff_info.staff_id,
	staff_info.prefix,
	staff_info.fname,
	staff_info.lname,
	staff_info.nname,
	staff_picture.staff_image
FROM staff_info
INNER JOIN staff_picture ON staff_info.staff_id = staff_picture.staff_id
; `

	if err := dbSale.Ctx().Raw(sql).Scan(&StaffAll).Error; err != nil {
		log.Errorln("GetAllStaff error :-", err)
	}
	return c.JSON(http.StatusOK, StaffAll)
}

func GetStaffEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer dbSale.Close()

	StaffId := c.QueryParam(("staff_id"))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id"})
	}

	var StaffInfo *m.StaffImg
	if err := dbSale.Ctx().Raw(`SELECT staff_info.staff_id, staff_info.prefix, staff_info.fname,
	staff_info.lname, staff_info.nname, staff_picture.staff_image
FROM staff_info
INNER JOIN staff_picture ON staff_info.staff_id = staff_picture.staff_id
WHERE staff_info.staff_id = ?
;`, StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("GetStaffInfo error :-", err)
	}
	if StaffInfo == nil {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot find staff id"})
	}
	StaffInfo.Img = base64.StdEncoding.EncodeToString([]byte(StaffInfo.StaffImage))
	return c.JSON(http.StatusOK, StaffInfo)
	// StaffInfo.StaffImage = str((result[0]['staff_image']).decode("utf-8"))
	// return json_response(StaffInfo, 200)
	// wg := sync.WaitGroup{}
	// wg.Add(4)
	// go func() {

	// 	for _, m := range data.Mail {
	// 		if err := dbSale.Ctx().Table("staff_mail").Create(&m).Error; err != nil {
	// 			return c.JSON(http.StatusInternalServerError, err)
	// 		}
	// 	}

	// 	if err := dbSale.Ctx().Raw(`SELECT id, email, remark as comment, '' as status FROM staff_mail WHERE ref_staff = ?;`, data.StaffId).Scan(&StaffMail).Error; err != nil {
	// 		log.Errorln("GetStaffMail error :-", err)
	// 	}
	// 	wg.Done()
	// }()
	// go func() {
	// 	if err := dbSale.Ctx().Raw(`SELECT id, tel, tel_sup, remark as comment, '' as status FROM staff_tel WHERE ref_staff = ?;`, data.StaffId).Scan(&StaffTel).Error; err != nil {
	// 		log.Errorln("GetStaffTel error :-", err)
	// 	}
	// 	wg.Done()
	// }()
	// go func() {
	// 	if err := dbSale.Ctx().Raw(`SELECT id, position, division, department, start_date, remark as comment, '' as status FROM staff_position WHERE ref_staff = ?;`, data.StaffId).Scan(&StaffPosition).Error; err != nil {
	// 		log.Errorln("GetStaffPosition error :-", err)
	// 	}
	// 	wg.Done()
	// }()
	// go func() {
	// 	if err := dbSale.Ctx().Raw(`SELECT id, skill, mark, comment, '' as status FROM staff_ability WHERE ref_staff = ?;`, data.StaffId).Scan(&StaffAbility).Error; err != nil {
	// 		log.Errorln("GetStaffAbility error :-", err)
	// 	}
	// 	wg.Done()
	// }()
	// wg.Wait()

	// for mail in data['mail']:
	// 	sql_str = """INSERT INTO staff_mail(email, ref_staff, remark) VALUES(%s,%s,%s);"""
	// 	values = (mail['email'],data['staff_id'],mail['comment'])
	// 	cur.execute(sql_str, values)
	// for tel in data['tel']:
	// 	sql_str = """INSERT INTO staff_tel(tel, tel_sup, ref_staff, remark) VALUES(%s,%s,%s,%s);"""
	// 	values = (tel['tel'],tel['tel_sup'],data['staff_id'],tel['comment'])
	// 	cur.execute(sql_str, values)
	// for position in data['position']:
	// 	sql_str = """INSERT INTO staff_position(position, division, department, ref_staff, start_date, remark) VALUES(%s,%s,%s,%s,%s,%s);"""
	// 	values = (position['position'],position['division'],position['department'],data['staff_id'],position['start_date'],position['comment'])
	// 	cur.execute(sql_str, values)
	// for ability in data['ability']:
	// 	sql_str = """INSERT INTO staff_ability(ref_staff, skill, mark, comment) VALUES(%s,%s,%s,%s);"""
	// 	values = (data['staff_id'],ability['skill'],ability['mark'],ability['comment'])
	// 	cur.execute(sql_str, values)
	// sql_str = """INSERT INTO staff_log(ref_staff, log_case, create_date, create_by, data_change, remark) VALUES(%s,%s,%s,%s,%s,%s);"""
	// values = (data['staff_id'], 'CREATE',datetime.datetime.now() , data['create_by'], '', '')
	// cur.execute(sql_str, values)
	// con.commit()
	// return json_response({"message": "create success"}, 200)

}
