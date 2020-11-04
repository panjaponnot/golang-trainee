package api

import (
	"encoding/base64"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"strings"
	"sync"

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
	log.Infoln(StaffId)

	var StaffInfo []m.StaffImg
	if err := dbSale.Ctx().Raw(`SELECT staff_info.staff_id, staff_info.prefix, staff_info.fname,
	staff_info.lname, staff_info.nname, staff_picture.staff_image
	FROM staff_info
	INNER JOIN staff_picture ON staff_info.staff_id = staff_picture.staff_id
	WHERE staff_info.staff_id = ?
	;`, StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("GetStaffInfo error :-", err)
	}
	if len(StaffInfo) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot find staff id"})
	}
	StaffInfo[0].Img = base64.StdEncoding.EncodeToString([]byte(StaffInfo[0].StaffImage))
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

}

func GetStaffProfileEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer dbSale.Close()

	StaffId := c.QueryParam(("staff_id"))
	var StaffMail m.StaffMail
	var StaffTel m.StaffTel
	var StaffOneMail m.StaffOneMail
	var StaffGoalMonth m.StaffGoalMonth
	var StaffGoalQuarter m.StaffGoalQuarter
	var StaffInfo []m.StaffProfile
	if err := dbSale.Ctx().Raw(`SELECT staff_id, prefix, fname, lname, nname, position
	from staff_info
	WHERE staff_id = ?`, StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("GetStaffInfo error :-", err)
	}
	if len(StaffInfo) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot find staff id"})
	}

	wg := sync.WaitGroup{}
	wg.Add(5)
	go func() {
		if err := dbSale.Ctx().Raw(`SELECT id, ref_staff, mail FROM staff_mail WHERE ref_staff = ?;`, StaffId).Scan(&StaffMail).Error; err != nil {
			log.Errorln("GetStaffMail error :-", err)
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(`SELECT id, ref_staff, onemail FROM staff_onemail WHERE ref_staff = ?`, StaffId).Scan(&StaffOneMail).Error; err != nil {
			log.Errorln("GetStaffOneMail error :-", err)
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(`SELECT id, ref_staff, tel, tel_sub FROM staff_tel WHERE ref_staff = ?`, StaffId).Scan(&StaffTel).Error; err != nil {
			log.Errorln("GetStaffTel error :-", err)
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(`SELECT id, ref_staff, year, month, goal_total, real_total, create_date, create_by
		FROM goal_month
		WHERE ref_staff = ?`, StaffId).Scan(&StaffGoalMonth).Error; err != nil {
			log.Errorln("GetStaffGoalMonth error :-", err)
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(`SELECT id, ref_staff, year, quarter, goal_total, real_total, create_date, create_by
		FROM goal_quarter
		WHERE ref_staff = ?`, StaffId).Scan(&StaffGoalQuarter).Error; err != nil {
			log.Errorln("GetStaffGoalQuarter error :-", err)
		}
		wg.Done()
	}()
	wg.Wait()
	StaffInfo[0].Mail = StaffMail
	StaffInfo[0].OneMail = StaffOneMail
	StaffInfo[0].Tel = StaffTel
	StaffInfo[0].Month = StaffGoalMonth
	StaffInfo[0].Quarter = StaffGoalQuarter
	return c.JSON(http.StatusOK, StaffInfo[0])
}
