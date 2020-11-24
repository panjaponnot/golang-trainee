package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/imagik"
	"sale_ranking/pkg/log"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

func GetAllStaffEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	StaffAll := []struct {
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
}

func GetStaffProfileEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	StaffId := c.QueryParam(("staff_id"))
	var StaffMail m.StaffMail
	var StaffTel []m.StaffTel
	var StaffOneMail m.StaffOneMail
	var StaffGoalMonth []m.StaffGoalMonth
	var StaffGoalQuarter []m.StaffGoalQuarter
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

func CreateStaffEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	hasErr := 0
	data := struct {
		OneId      string           `json:"one_id" gorm:"column:one_id"`
		StaffId    string           `json:"staff_id" gorm:"column:staff_id"`
		Prefix     string           `json:"prefix" gorm:"column:prefix"`
		Fname      string           `json:"fname" gorm:"column:fname"`
		Lname      string           `json:"lname" gorm:"column:lname"`
		Nname      string           `json:"nname" gorm:"column:nname"`
		Position   string           `json:"position" gorm:"column:position"`
		Department string           `json:"department" gorm:"column:department"`
		StaffChild string           `json:"staff_child" gorm:"column:staff_child"`
		Mail       []m.StaffMail    `json:"mail" gorm:"column:mail"`
		OneMail    []m.StaffOneMail `json:"OneMail" gorm:"column:OneMail"`
		Tel        []m.StaffTel     `json:"Tel" gorm:"column:Tel"`
	}{}
	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}
	// log.Infoln(data)
	fmt.Println("====", data)
	var StaffInfo []m.StaffInfo
	if err := dbSale.Ctx().Raw(`SELECT staff_id from staff_info WHERE staff_id = ?;`, data.StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("CheckStaffInfo error :-", err)
	}
	if len(StaffInfo) > 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "duplicate staff id"})
	}
	if err := dbSale.Ctx().Table("staff_info").Create(&data).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	if err := dbSale.Ctx().Raw(`SELECT staff_id from staff_info WHERE staff_id = ?;`, data.StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("CheckStaffInfo error :-", err)
	}
	if len(StaffInfo) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "not find staff id"})
	}

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		for _, mail := range data.Mail {
			dataMail := struct {
				RefStaff string `json:"ref_staff" gorm:"column:ref_staff"`
				Mail     string `json:"mail" gorm:"column:mail"`
			}{
				RefStaff: data.StaffId,
				Mail:     mail.Mail,
			}
			if err := dbSale.Ctx().Table("staff_mail").Create(&dataMail).Error; err != nil {
				hasErr++
			}
		}
		wg.Done()
	}()
	go func() {
		for _, onemail := range data.OneMail {
			dataOneMail := struct {
				RefStaff string `json:"ref_staff" gorm:"column:ref_staff"`
				OneMail  string `json:"onemail" gorm:"column:onemail"`
			}{
				RefStaff: data.StaffId,
				OneMail:  onemail.Onemail,
			}
			if err := dbSale.Ctx().Table("staff_onemail").Create(&dataOneMail).Error; err != nil {
				hasErr++
			}
		}
		wg.Done()
	}()
	go func() {
		for _, tel := range data.Tel {
			dataTel := struct {
				RefStaff string `json:"ref_staff" gorm:"column:ref_staff"`
				Tel      string `json:"tel" gorm:"column:tel"`
				TelSub   string `json:"tel_sub" gorm:"column:tel_sub"`
			}{
				RefStaff: data.StaffId,
				Tel:      tel.Tel,
				TelSub:   tel.TelSub,
			}
			if err := dbSale.Ctx().Table("staff_tel").Create(&dataTel).Error; err != nil {
				hasErr++
			}
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, m.Result{Message: "create success"})
}

func EditStaffEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	data := struct {
		OneId       string `json:"one_id"`
		EditStaffId string `json:"editstaff_id"`
		Prefix      string `json:"prefix"`
		Fname       string `json:"fname"`
		Lname       string `json:"lname"`
		Nname       string `json:"nname"`
		StaffId     string `json:"staff_id"`
	}{}
	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}
	if err := dbSale.Ctx().Exec("UPDATE staff_info SET one_id = ?,staff_id = ?,prefix = ?,fname = ?,lname = ?,nname = ? WHERE staff_id = ?;", data.OneId, data.EditStaffId, data.Prefix, data.Fname, data.Lname, data.Nname, data.StaffId).Error; err != nil {
		log.Errorln("UPDATEStaffInfo error :-", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, m.Result{Message: "update success"})
}

func DeleteStaffEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	data := struct {
		StaffId string `json:"staff_id" gorm:"column:staff_id"`
	}{}
	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}
	if err := dbSale.Ctx().Exec("DELETE FROM staff_info WHERE staff_id = ?", data.StaffId).Error; err != nil {
		log.Errorln("DeleteStaffInfo error :-", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	if err := dbSale.Ctx().Exec("DELETE FROM staff_mail WHERE ref_staff = ?", data.StaffId).Error; err != nil {
		log.Errorln("DeleteStaffMail error :-", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	if err := dbSale.Ctx().Exec("DELETE FROM staff_onemail WHERE ref_staff = ?", data.StaffId).Error; err != nil {
		log.Errorln("DeleteStaffOneMail error :-", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	if err := dbSale.Ctx().Exec("DELETE FROM staff_tel WHERE ref_staff = ?", data.StaffId).Error; err != nil {
		log.Errorln("DeleteStaffTel error :-", err)
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, m.Result{Message: "DELETE success"})
}

func GetStaffPictureEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	StaffId := c.QueryParam(("one_id"))
	if strings.TrimSpace(c.QueryParam(("one_id"))) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid one id"})
	}
	// log.Infoln(StaffId)
	var StaffInfo []m.StaffImg
	if err := dbSale.Ctx().Raw(`SELECT *
	FROM staff_info AS sif
	LEFT JOIN staff_images  AS sp
	ON sif.one_id = sp.one_id
	WHERE sif.one_id = ?
	;`, StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("GetStaffInfoPicture error :-", err)
	}
	if len(StaffInfo) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot find staff id"})
	}
	StaffInfo[0].Img = base64.StdEncoding.EncodeToString([]byte(StaffInfo[0].StaffImage))
	return c.JSON(http.StatusOK, StaffInfo)
}

func GetAllStaffIdEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	var StaffInfo []m.StaffId
	if err := dbSale.Ctx().Raw(`SELECT staff_id FROM staff_info;`).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("GetStaffInfo error :-", err)
	}
	if len(StaffInfo) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot find staff id"})
	}
	return c.JSON(http.StatusOK, StaffInfo)
}

func GetGroupChild(c echo.Context, group []m.GroupRelation) []m.GroupRelation {
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "connect database error")
	// 	// return c.JSON(http.StatusInternalServerError, err)
	// 	defer dbSale.Close()
	// }
	// defer dbSale.Close()
	log.Infoln("==--->", group)
	var IdGroup string
	var InsResult []m.GroupRelation
	var Result []m.GroupRelation
	var Result2 []m.GroupRelation
	var em []m.GroupRelation
	// var check int
	// var InsResults []m.GroupRelation
	for _, g := range group {

		if g.IdGroup != "" {
			IdGroup = g.IdGroup
		} else {
			IdGroup = g.IdGroupChild
		}
		if err := dbSale.Ctx().Raw(`SELECT id_group_child
			FROM group_relation
			WHERE id_group = ?`, IdGroup).Scan(&InsResult).Error; err != nil {
			log.Errorln("GetInsResult error :-", err)
		}
		// log.Infoln("==---++>", InsResult)
		if len(InsResult) == 0 {
			InsResult = em
		}
		Result = GetGroupChild(c, InsResult)
		for _, r := range InsResult {
			Result = append(Result, r)
		}
		Result2 = Result
	}
	// log.Infoln("==--Result2--++>", Result2)
	return Result2
}

func GetStaffByIdGroup(c echo.Context, group []m.GroupRelation) []m.StaffGroupRelationId {
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "connect database error")
	// 	// return c.JSON(http.StatusInternalServerError, err)
	// 	// defer dbSale.Close()
	// }
	// defer dbSale.Close()
	var IdGroup string
	var InsResult []m.StaffGroupRelationId
	var IdStarfList []m.StaffGroupRelationId
	for _, g := range group {
		if g.IdGroup != "" {
			IdGroup = g.IdGroup
		} else {
			IdGroup = g.IdGroupChild
		}
		// log.Infoln("==--000000--++>")
		if err := dbSale.Ctx().Raw(`SELECT id_staff
		FROM staff_group_relation
		WHERE id_group = ?`, IdGroup).Scan(&InsResult).Error; err != nil {
			log.Errorln("GetInsResult error :-", err)
		}

		for _, i := range InsResult {
			IdStarfList = append(IdStarfList, i)
		}
	}
	// log.Infoln("111111111111111111111")
	// log.Infoln("==>", IdStarfList)
	return IdStarfList
}

func CheckDupStaffId(sgr []m.StaffGroupRelationId) []m.StaffGroupRelationId {
	check := 0
	var CheckedStaffIdList []m.StaffGroupRelationId
	for _, sgr := range sgr {
		for _, c := range CheckedStaffIdList {
			if sgr.IdGroup == c.IdGroup {
				check++
			}
		}
		if check == 0 {
			CheckedStaffIdList = append(CheckedStaffIdList, sgr)
		}
		check = 0
	}
	// log.Infoln("222222222222222222")
	// log.Infoln("==>", CheckedStaffIdList)
	return CheckedStaffIdList
}

func GetStaffInfoById(c echo.Context, CheckedStaffIdList []m.StaffGroupRelationId) ([]m.StaffInfo, error) {
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "connect database error")
	// 	// return c.JSON(http.StatusInternalServerError, err)
	// }
	// defer dbSale.Close()
	// log.Infoln("333333333333333333333")
	// log.Infoln("==>", CheckedStaffIdList)
	var StaffInfo []m.StaffInfo
	var Result []m.StaffInfo
	var StaffId string
	for _, c := range CheckedStaffIdList {
		StaffId = c.IdGroup
		if err := dbSale.Ctx().Raw(`SELECT prefix, fname, lname
		FROM staff_info
		WHERE staff_id = ?`, StaffId).Scan(&Result).Error; err != nil {
			log.Errorln("GetStaffInfo error :-", err)
		}
		for _, r := range Result {
			StaffInfo = append(StaffInfo, r)
		}
	}
	// log.Infoln("444444444444444444444444")
	// log.Infoln("==>", StaffInfo)
	return StaffInfo, nil
}

func GetSubordinateStaff(c echo.Context) ([]m.StaffInfo, error) {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		// return c.JSON(http.StatusInternalServerError, err)
	}
	defer dbSale.Close()
	StaffId := c.QueryParam(("staff_id"))
	log.Infoln(StaffId)
	var Result []m.StaffGroupRelation
	var ResultData []m.GroupRelation
	if err := dbSale.Ctx().Raw(`SELECT id_group
	FROM staff_group_relation
	WHERE id_staff = ?`, StaffId).Scan(&Result).Error; err != nil {
		log.Errorln("GetResult error :-", err)
	}
	// log.Infoln(Result)
	for _, r := range Result {
		data := m.GroupRelation{
			IdGroup:      r.IdGroup,
			IdGroupChild: "",
		}
		ResultData = append(ResultData, data)
	}
	log.Infoln("==>", ResultData)
	GroupChildList := GetGroupChild(c, ResultData)
	log.Infoln("==>GroupChildList==>", len(GroupChildList))
	StaffIdList := GetStaffByIdGroup(c, GroupChildList)
	CheckedStaffIdList := CheckDupStaffId(StaffIdList)
	StaffInfo, _ := GetStaffInfoById(c, CheckedStaffIdList)
	return StaffInfo, nil
}

func GetSubordinateStaffEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
	}
	defer dbSale.Close()
	StaffInfo, err := GetSubordinateStaff(c)
	if err != nil {
		log.Errorln("cannot get subordinate staff", err)
	}

	return c.JSON(http.StatusOK, StaffInfo)
}

func TruncateTable(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		// return c.JSON(http.StatusInternalServerError, err)
	}
	defer dbSale.Close()
	tx := dbSale.Ctx().Begin()
	if err := dbSale.Ctx().Exec(" TRUNCATE TABLE so_mssql; ").Error; err != nil {
		log.Errorln("Truncate table error :-", err)
		tx.Rollback()
		return err
	}
	return nil
}

func CreateStaffPictureEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
	}
	defer dbSale.Close()

	var StaffIdList []m.StaffOneId
	if err := dbSale.Ctx().Raw(`SELECT one_id,staff_id FROM staff_info;`).Scan(&StaffIdList).Error; err != nil {
		log.Errorln("GetOneIdList error :-", err)
	}

	for _, s := range StaffIdList {
		NamePic := strings.TrimSpace(s.StaffId)
		StaffId := strings.TrimSpace(s.StaffId)
		var StaffIds = fmt.Sprintf("%ss", StaffId)
		OneId := strings.TrimSpace(s.OneId)
		var FileName = fmt.Sprintf("%s.jpg", StaffId)
		if CheckPictureUrl(NamePic) {
			log.Infoln("======>", NamePic)
			var b []byte
			var mimeType string
			var avatarUrl = fmt.Sprintf("https://intranet.inet.co.th/assets/upload/staff/%s.jpg", NamePic)
			if err := imagik.UrlGrabber(avatarUrl, nil, &b, &mimeType, 5); err != nil {
				log.Warnln("Get staff avatar error -:", err)
				return echo.ErrInternalServerError
			}
			str := base64.StdEncoding.EncodeToString(b)
			vars := m.StaffPicture{
				OneId:    OneId,
				Image:    str,
				FileName: FileName,
			}
			tx := dbSale.Ctx().Begin()
			if err := dbSale.Ctx().Table("staff_images").Create(&vars).Error; err != nil {
				log.Errorln("Func Insert StaffPicture error :-", err)
				tx.Rollback()
				return err
			}
		} else if CheckPictureUrl(StaffIds) == true {
			log.Infoln("====2==>", StaffIds)
			log.Infoln("====222==>", CheckPictureUrl(StaffIds))
			var b []byte
			var mimeType string
			var avatarUrl = fmt.Sprintf("https://intranet.inet.co.th/assets/upload/staff/%s.jpg", StaffIds)
			if err := imagik.UrlGrabber(avatarUrl, nil, &b, &mimeType, 5); err != nil {
				log.Warnln("Get staff avatar error --:", err)
				return echo.ErrInternalServerError
			}
			str := base64.StdEncoding.EncodeToString(b)
			vars := m.StaffPicture{
				OneId:    OneId,
				Image:    str,
				FileName: FileName,
			}
			log.Infoln("===val==>", vars)
			tx := dbSale.Ctx().Begin()
			if err := dbSale.Ctx().Table("staff_images").Create(&vars).Error; err != nil {
				log.Errorln("Func Insert StaffPicture error :-", err)
				tx.Rollback()
				return err
			}
		}
	}
	Message := m.Message{
		Message: "create success",
	}
	return c.JSON(http.StatusOK, Message)
	// return c.JSON(http.StatusOK, StaffInfo)
}

func CheckPictureUrl(namepic string) bool {
	var avatarUrl = fmt.Sprintf("https://intranet.inet.co.th/assets/upload/staff/%s.jpg", namepic)
	response, err := http.Get(avatarUrl)
	if err != nil {
		fmt.Println("HTTP call failed:", err)
		return false
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", response.StatusCode)
		return false
	}
	defer response.Body.Close()
	return true

}

func GetStaffProfileV2EndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}

	StaffId := c.QueryParam(("staff_id"))
	var StaffMail m.StaffMail
	var StaffTel []m.StaffTel
	var StaffOneMail m.StaffOneMail
	// var StaffGoalMonth m.StaffGoalMonth
	// var StaffGoalQuarter m.StaffGoalQuarter
	var StaffInfo []m.StaffProfileV2
	if err := dbSale.Ctx().Raw(`SELECT staff_id, prefix, fname, lname, nname, position
	from staff_info
	WHERE staff_id = ?`, StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("GetStaffInfo error :-", err)
	}
	if len(StaffInfo) == 0 {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot find staff id"})
	}

	wg := sync.WaitGroup{}
	wg.Add(4)
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
		gq := GetGq(c, StaffId)
		StaffInfo[0].Quarter = gq
		wg.Done()
	}()
	wg.Wait()

	StaffInfo[0].Mail = StaffMail
	StaffInfo[0].OneMail = StaffOneMail
	StaffInfo[0].Tel = StaffTel
	// StaffInfo[0].Month = StaffGoalMonth
	// StaffInfo[0].Quarter = StaffGoalQuarter

	return c.JSON(http.StatusOK, StaffInfo[0])
}

func GetGq(c echo.Context, StaffId string) []m.GqDict {
	OwnStaffId := StaffId
	var StaffInfo m.StaffInfo
	var DateResult []m.DateResult
	var StaffIdGoalQuarter []m.GqDict
	var GqDict []m.GqDict
	var GqList []m.GqDict
	if err := dbSale.Ctx().Raw(`SELECT
	CURDATE() as cur_0,
	SUBDATE(CURDATE(), INTERVAL 1 MONTH) as pv_1,
	SUBDATE(CURDATE(), INTERVAL 2 MONTH) as pv_2,
	DATE_ADD(CURDATE(), INTERVAL 1 MONTH) as nt_1,
	DATE_ADD(CURDATE(), INTERVAL 2 MONTH) as nt_2
	LIMIT 1
;`).Scan(&DateResult).Error; err != nil {
		log.Errorln("GetStaffTel error :-", err)
	}

	if err := dbSale.Ctx().Raw(`SELECT *
	FROM staff_info
	WHERE staff_id = ?`, StaffId).Scan(&StaffInfo).Error; err != nil {
		log.Errorln("GetStaffInfo error :-", err)
	}

	CheckStaffChild := strings.Split(strings.TrimSpace(StaffInfo.StaffChild), ",")
	StaffChild := StaffInfo.StaffChild

	if len(CheckStaffChild) == 1 && len(CheckStaffChild[0]) == 0 {
		StaffChild = strings.TrimSpace(StaffId)
	} else {
		StaffChild = ""
		for n, s := range CheckStaffChild {
			if n == 0 {
				StaffChild += fmt.Sprintf("%s',", s)
			} else if n+1 == len(CheckStaffChild) {
				StaffChild += fmt.Sprintf("'%s", s)
			} else {
				StaffChild += fmt.Sprintf("'%s',", s)
			}

		}
	}

	var DateData []string
	DateData = append(DateData, DateResult[0].Cur0)
	DateData = append(DateData, DateResult[0].Pv1)
	DateData = append(DateData, DateResult[0].Pv2)
	DateData = append(DateData, DateResult[0].Nt1)
	DateData = append(DateData, DateResult[0].Nt2)
	// log.Infoln("1311313133131133131")
	// log.Infoln("---", DateResult)
	for i, d := range DateData {

		str := fmt.Sprintf(`
		SELECT total_amount, goal_total, year, quarter, MONTH(%s) as month
		FROM
			(
				WITH gq_data AS
				(
					SELECT goal_total, year, quarter, ref_staff
					FROM goal_quarter
					WHERE
						goal_quarter.year = YEAR(%s) AND
						goal_quarter.quarter = CONCAT("Q", QUARTER(%s))
				)
				SELECT so_amount.*, gq_data.*
				FROM
					(
						SELECT sonumber, sale_code, SUM(PeriodAmount) as total_amount
						FROM
							(
								SELECT sonumber, PeriodAmount, sale_code
								FROM so_mssql
								WHERE QUARTER(ContractStartDate) = QUARTER(%s) AND
									MONTH(DATE(ContractStartDate)) <= MONTH(%s) AND
									YEAR(DATE(ContractStartDate)) = YEAR(%s) AND
									so_mssql.sale_code IN (%s)
								GROUP BY sonumber
							) as so_data
					) as so_amount
				INNER JOIN gq_data ON gq_data.ref_staff = so_amount.sale_code
			) as staff_so_data
		;`, d, d, d, d, d, d, StaffChild)
		if err := dbSale.Ctx().Raw(str).Scan(&StaffIdGoalQuarter).Error; err != nil {
			log.Errorln("GetStaffIdGoalQuarter error 1:-", err)
		}
		if len(StaffIdGoalQuarter) == 0 {
			str := fmt.Sprintf(`
			SELECT goal_total, year, quarter, MONTH(%s) as month
			FROM goal_quarter
			WHERE ref_staff IN (%s)
			LIMIT 1;`, d, StaffChild)
			if err := dbSale.Ctx().Raw(str).Scan(&StaffIdGoalQuarter).Error; err != nil {
				log.Errorln("GetStaffIdGoalQuarter error 2:-", err)
			}
			GqDictData := m.GqDict{
				TotalAmount: 0,
				GoalTotal:   StaffIdGoalQuarter[0].GoalTotal,
				Year:        StaffIdGoalQuarter[0].Year,
				Quarter:     StaffIdGoalQuarter[0].Quarter,
				Month:       StaffIdGoalQuarter[0].Month,
			}
			GqDict = append(GqDict, GqDictData)
		} else {
			GqDict = append(GqDict, StaffIdGoalQuarter[0])
			if len(CheckStaffChild) > 1 {
				if err := dbSale.Ctx().Raw(`
			SELECT goal_total, year, quarter, MONTH(?) as month
			FROM goal_quarter
			WHERE ref_staff IN (?)
			LIMIT 1;`, d, OwnStaffId).Scan(&StaffIdGoalQuarter).Error; err != nil {
					log.Errorln("GetStaffIdGoalQuarter error 3:-", err)
				}

				GqDict[i].GoalTotal = StaffIdGoalQuarter[0].GoalTotal
			}
		}
		GqList = append(GqList, GqDict[i])
	}
	return GqList
}

func HeaderSummaryEndPoint(c echo.Context) error {
	data := struct {
		StaffId string `json:"staff_id"`
		Type    string `json:"type"`
		Year    string `json:"year"`
		Month   string `json:"month"`
		Quarter string `json:"quarter"`
	}{}
	if err := c.Bind(&data); err != nil {
		return echo.ErrBadRequest
	}

	// if data.OneId == "" {
	// 	return c.JSON(http.StatusBadRequest, server.Result{Mess: "not have param one_id"})
	// }
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	// tx := dbSale.Ctx().Begin()

	StaffId := data.StaffId
	Type := data.Type
	Year := data.Year
	Month := data.Month
	Quarter := data.Quarter

	StaffId = CheckPermission(StaffId)
	StaffProfile := PersonalInformation(StaffId)
	if StaffProfile.StaffId == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "cannot get staff information"})
	}
	var StaffChild []string
	StaffChildList := strings.Split(StaffProfile.StaffChild, ",")
	for _, s := range StaffChildList {
		if s != "" {
			StaffChild = append(StaffChild, s)
		}
	}

	StaffProfile.SummarySostatus = SOstatus(StaffChild, Type, Year, Quarter, Month)
	StaffProfile.SummaryHeader = HeaderSummary(StaffChild, Type, Year, Quarter, Month)
	StaffProfile.CustomerNewso = NewSOSummary(StaffChild, Type, Year, Quarter, Month)
	StaffProfile.ChartNewso = SOChart(StaffChild, Type, Year, Quarter, Month)
	StaffProfile.ChartSaleFactor = SFChart(StaffChild, Type, Year, Quarter, Month)

	return c.JSON(http.StatusOK, StaffProfile)
}

func CheckPermission(id string) string {
	var user []m.UserInfo
	// notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return "error"
	}
	if len(user) == 0 {

		return id
	}
	sqlStaff := `select * from staff_info where CHAR_LENGTH(staff_child) in (select Max(CHAR_LENGTH(staff_child)) as aaa from staff_info);`
	if err := dbSale.Ctx().Raw(sqlStaff).Scan(&user).Error; err != nil {
		return "error"
	}
	return user[0].StaffId
}

type StaffProfileV3 struct {
	OneId           string          `json:"one_id"`
	StaffId         string          `json:"staff_id"`
	Fname           string          `json:"fname"`
	Lname           string          `json:"lname"`
	Nname           string          `json:"nname"`
	Position        string          `json:"position"`
	Department      string          `json:"department"`
	StaffChild      string          `json:"staff_child"`
	Image           string          `json:"image"`
	StartJob        string          `json:"start_job"`
	Mail            string          `json:"mail"`
	OneMail         string          `json:"onemail"`
	SummarySostatus []Somssql       `json:"SummarySostatus"`
	SummaryHeader   []SomssqlHeader `json:"SummaryHeader"`
	CustomerNewso   []SomssqlNew    `json:"CustomerNewso"`
	ChartNewso      []SoSummary     `json:"ChartNewso"`
	ChartSaleFactor []SoSFChart     `json:"ChartSaleFactor"`
}

func PersonalInformation(StaffId string) StaffProfileV3 {

	var user []StaffProfileV3
	var userEmp []StaffProfileV3
	sqlUsr := `SELECT staff_info.one_id,staff_id,fname,lname,nname,position,department,staff_child,image,start_job,
		GROUP_CONCAT(staff_mail.mail SEPARATOR ', ') as mail,
		GROUP_CONCAT(staff_onemail.onemail SEPARATOR ', ') as onemail
	FROM staff_info
	left join staff_mail on staff_info.staff_id = staff_mail.ref_staff
	left join staff_onemail on staff_info.staff_id = staff_onemail.ref_staff
	left join staff_start on staff_info.one_id = staff_start.one_id
	left join staff_images on staff_info.one_id = staff_images.one_id
	where staff_id = ?
	group by staff_info.one_id;`
	if err := dbSale.Ctx().Raw(sqlUsr, StaffId).Scan(&user).Error; err != nil {

		log.Errorln("Error Get User", err)
		return userEmp[0]
	}
	if len(user) == 0 {
		return userEmp[0]
	}
	return user[0]
}

type Somssql struct {
	SOWebStatus string `json:"SOWebStatus"`
	Total       string `json:"total"`
}

func SOstatus(StaffChild []string, Type string, Year string, Quarter string, Month string) []Somssql {

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
	var Result []Somssql
	var SqlStr string
	SqlStr = fmt.Sprintf(`select
	SOWebStatus, count(SOWebStatus) as total
	from (
	select sonumber,SOWebStatus from so_mssql
	where
	sale_code in (%s) and `, StaffChildStr)

	if Type == "Quarter" {
		SqlStr += fmt.Sprintf(` quarter(ContractStartDate) = '%s' and year(ContractStartDate) = '%s' `, Quarter, Year)
	} else {
		SqlStr += fmt.Sprintf(` month(ContractStartDate) = '%s' and year(ContractStartDate) = '%s' `, Month, Year)
	}

	SqlStr += ` GROUP BY sonumber
	) tb_so
	group by SOWebStatus;`

	if err := dbSale.Ctx().Raw(SqlStr).Scan(&Result).Error; err != nil {
		log.Errorln("Error Get User", err)
		return nil
	}
	return Result
}

type SomssqlHeader struct {
	TotalCustomer       string `json:"total_customer"`
	TotalSo             string `json:"total_so"`
	AmountSo            string `json:"amount_so"`
	AmountRecuring      string `json:"amount_recuring"`
	AmountRecuringNewso string `json:"amount_recuring_newso"`
}

func HeaderSummary(StaffChild []string, Type string, Year string, Quarter string, Month string) []SomssqlHeader {

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
	var Result []SomssqlHeader
	var SqlStr string
	SqlStr = ` select
		count( distinct Customer_ID) as total_customer,
		count(sonumber) as total_so,
		sum(sum_period_amount) as amount_so,
		sum(PeriodAmount) as amount_recuring,
		sum(CASE
			when has_refer = 0 then PeriodAmount
			else 0 end
		) as amount_recuring_newso
	from (
		SELECT sonumber, Customer_ID,Customer_Name, ContractStartDate, ContractEndDate, sale_code,
			sum(PeriodAmount) as sum_period_amount ,
			PeriodAmount,has_refer
		FROM so_mssql
		WHERE
		sale_code in (?) and `

	if Type == "Quarter" {
		SqlStr += fmt.Sprintf(` quarter(ContractStartDate) = %s and year(ContractStartDate) = %s `, Quarter, Year)
	} else {
		SqlStr += fmt.Sprintf(` month(ContractStartDate) = %s and year(ContractStartDate) = %s `, Month, Year)
	}

	SqlStr += ` group by sonumber
	) tb_so;`

	if err := dbSale.Ctx().Raw(SqlStr, StaffChildStr).Scan(&Result).Error; err != nil {
		log.Errorln("Error Get User", err)
		return nil
	}
	return Result
}

type SomssqlNew struct {
	CustomerId   string `json:"Customer_ID"`
	CustomerName string `json:"Customer_Name"`
	AmountSo     string `json:"amount_so"`
	TotalSo      string `json:"total_so"`
}

func NewSOSummary(StaffChild []string, Type string, Year string, Quarter string, Month string) []SomssqlNew {

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
	var Result []SomssqlNew
	var SqlStr string
	SqlStr = fmt.Sprintf(` select
		Customer_ID,Customer_Name, sum(PeriodAmount) as amount_so, count(sonumber) as total_so
	from (
		SELECT sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,TotalContractAmount,PeriodAmount
		FROM so_mssql
		where sale_code in (%s) and has_refer = 0 and `, StaffChildStr)

	if Type == "Quarter" {
		SqlStr += fmt.Sprintf(` quarter(ContractStartDate) = %s and year(ContractStartDate) = %s `, Quarter, Year)
	} else {
		SqlStr += fmt.Sprintf(` month(ContractStartDate) = %s and year(ContractStartDate) = %s `, Month, Year)
	}

	SqlStr += ` group by sonumber
	) tb_so
	group by Customer_ID;`

	if err := dbSale.Ctx().Raw(SqlStr).Scan(&Result).Error; err != nil {
		log.Errorln("Error Get User", err)
		return nil
	}
	return Result
}

type ListQuarterData struct {
	Year    int `json:"year"`
	Quarter int `json:"quarter"`
}

type ListMonthData struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

func ListQuarter(Year int, Quarter int) []ListQuarterData {
	var ListQuarterList []ListQuarterData
	if Quarter == 1 {
		data := []ListQuarterData{{
			Year:    Year - 1,
			Quarter: 3,
		}, {
			Year:    Year - 1,
			Quarter: 4,
		}, {
			Year:    Year,
			Quarter: 1,
		}, {
			Year:    Year,
			Quarter: 2,
		}, {
			Year:    Year,
			Quarter: 3,
		}}
		for _, d := range data {
			ListQuarterList = append(ListQuarterList, d)
		}
	} else if Quarter == 2 {
		data := []ListQuarterData{{
			Year:    Year - 1,
			Quarter: 4,
		}, {
			Year:    Year,
			Quarter: 1,
		}, {
			Year:    Year,
			Quarter: 2,
		}, {
			Year:    Year,
			Quarter: 3,
		}, {
			Year:    Year,
			Quarter: 4,
		}}
		for _, d := range data {
			ListQuarterList = append(ListQuarterList, d)
		}
	} else if Quarter == 3 {
		data := []ListQuarterData{{
			Year:    Year,
			Quarter: 1,
		}, {
			Year:    Year,
			Quarter: 2,
		}, {
			Year:    Year,
			Quarter: 3,
		}, {
			Year:    Year,
			Quarter: 4,
		}, {
			Year:    Year + 1,
			Quarter: 1,
		}}
		for _, d := range data {
			ListQuarterList = append(ListQuarterList, d)
		}
	} else {
		data := []ListQuarterData{{
			Year:    Year,
			Quarter: 2,
		}, {
			Year:    Year,
			Quarter: 3,
		}, {
			Year:    Year,
			Quarter: 4,
		}, {
			Year:    Year + 1,
			Quarter: 1,
		}, {
			Year:    Year + 1,
			Quarter: 2,
		}}
		for _, d := range data {
			ListQuarterList = append(ListQuarterList, d)
		}
	}
	return ListQuarterList
}

func ListMonth(Year int, Month int) []ListMonthData {
	var ListMonthList []ListMonthData
	if Month == 1 {
		data := []ListMonthData{{
			Year:  Year - 1,
			Month: 11,
		}, {
			Year:  Year - 1,
			Month: 12,
		}, {
			Year:  Year,
			Month: 1,
		}, {
			Year:  Year,
			Month: 2,
		}, {
			Year:  Year,
			Month: 3,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 2 {
		data := []ListMonthData{{
			Year:  Year - 1,
			Month: 11,
		}, {
			Year:  Year - 1,
			Month: 12,
		}, {
			Year:  Year,
			Month: 1,
		}, {
			Year:  Year,
			Month: 2,
		}, {
			Year:  Year,
			Month: 3,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 3 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 1,
		}, {
			Year:  Year,
			Month: 2,
		}, {
			Year:  Year,
			Month: 3,
		}, {
			Year:  Year,
			Month: 4,
		}, {
			Year:  Year,
			Month: 5,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 4 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 2,
		}, {
			Year:  Year,
			Month: 3,
		}, {
			Year:  Year,
			Month: 4,
		}, {
			Year:  Year,
			Month: 5,
		}, {
			Year:  Year,
			Month: 6,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 5 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 3,
		}, {
			Year:  Year,
			Month: 4,
		}, {
			Year:  Year,
			Month: 5,
		}, {
			Year:  Year,
			Month: 6,
		}, {
			Year:  Year,
			Month: 7,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 6 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 4,
		}, {
			Year:  Year,
			Month: 5,
		}, {
			Year:  Year,
			Month: 6,
		}, {
			Year:  Year,
			Month: 7,
		}, {
			Year:  Year,
			Month: 8,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 7 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 5,
		}, {
			Year:  Year,
			Month: 6,
		}, {
			Year:  Year,
			Month: 7,
		}, {
			Year:  Year,
			Month: 8,
		}, {
			Year:  Year,
			Month: 9,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 8 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 6,
		}, {
			Year:  Year,
			Month: 7,
		}, {
			Year:  Year,
			Month: 8,
		}, {
			Year:  Year,
			Month: 9,
		}, {
			Year:  Year,
			Month: 10,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 9 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 7,
		}, {
			Year:  Year,
			Month: 8,
		}, {
			Year:  Year,
			Month: 9,
		}, {
			Year:  Year,
			Month: 10,
		}, {
			Year:  Year,
			Month: 11,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 10 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 8,
		}, {
			Year:  Year,
			Month: 9,
		}, {
			Year:  Year,
			Month: 10,
		}, {
			Year:  Year,
			Month: 11,
		}, {
			Year:  Year,
			Month: 12,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else if Month == 11 {
		data := []ListMonthData{{
			Year:  Year,
			Month: 9,
		}, {
			Year:  Year,
			Month: 10,
		}, {
			Year:  Year,
			Month: 11,
		}, {
			Year:  Year,
			Month: 12,
		}, {
			Year:  Year + 1,
			Month: 1,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	} else {
		data := []ListMonthData{{
			Year:  Year,
			Month: 2,
		}, {
			Year:  Year,
			Month: 3,
		}, {
			Year:  Year,
			Month: 4,
		}, {
			Year:  Year + 1,
			Month: 1,
		}, {
			Year:  Year + 1,
			Month: 2,
		}}
		for _, d := range data {
			ListMonthList = append(ListMonthList, d)
		}
	}
	return ListMonthList
}

type SoSummary struct {
	YearChart    string `json:"year_chart"`
	QuarterChart string `json:"quarter_chart"`
	Amount       string `json:"amount"`
}

func SOChart(StaffChild []string, Type string, Year string, Quarter string, Month string) []SoSummary {
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
	YearInt, err := strconv.Atoi(Year)
	if err != nil {
		log.Errorln(pkgName, err)
	}
	QuarterInt, err := strconv.Atoi(Quarter)
	if err != nil {
		log.Errorln(pkgName, err)
	}
	MonthInt, err := strconv.Atoi(Month)
	if err != nil {
		log.Errorln(pkgName, err)
	}
	var SqlStr string
	if Type == "Quarter" {
		ListQuarter := ListQuarter(YearInt, QuarterInt)
		for n, l := range ListQuarter {
			SqlStr += fmt.Sprintf(`
				select
					year(ContractStartDate) as year_chart,
					quarter(ContractStartDate) as quarter_chart,
					sum(PeriodAmount) as amount
				from (
					SELECT sonumber,PeriodAmount,ContractStartDate
					FROM so_mssql
					where has_refer = 0 and sale_code in (%s)
					and year(ContractStartDate) = %d
					and quarter(ContractStartDate) = %d
					group by sonumber
				) tb_so`, StaffChildStr, l.Year, l.Quarter)
			if n != 4 {
				SqlStr += ` union `
			}
		}
		var Result []SoSummary
		if err := dbSale.Ctx().Raw(SqlStr).Scan(&Result).Error; err != nil {
			log.Errorln("Error Get SoSummary", err)
			return nil
		}

		return Result
	} else {
		ListMonth := ListMonth(YearInt, MonthInt)
		for n, l := range ListMonth {
			SqlStr += fmt.Sprintf(`
			select
			year(ContractStartDate) as year_chart,
			month(ContractStartDate) as month_chart,
			sum(PeriodAmount) as amount
		from (
			SELECT sonumber,PeriodAmount,ContractStartDate
			FROM so_mssql
			where has_refer = 0 and sale_code in (%s)
			and year(ContractStartDate) = %d
			and month(ContractStartDate) = %d
			group by sonumber
		) tb_so`, StaffChildStr, l.Year, l.Month)
			if n != 4 {
				SqlStr += ` union `
			}
		}
		var Result []SoSummary
		if err := dbSale.Ctx().Raw(SqlStr).Scan(&Result).Error; err != nil {
			log.Errorln("Error Get SoSummary", err)
			return nil
		}

		return Result
	}
}

type SoSFChart struct {
	YearChart    string `json:"year_chart"`
	QuarterChart string `json:"quarter_chart"`
	SaleFactor   string `json:"sale_factor"`
}

func SFChart(StaffChild []string, Type string, Year string, Quarter string, Month string) []SoSFChart {
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
	YearInt, err := strconv.Atoi(Year)
	if err != nil {
		log.Errorln(pkgName, err)
	}
	QuarterInt, err := strconv.Atoi(Quarter)
	if err != nil {
		log.Errorln(pkgName, err)
	}
	MonthInt, err := strconv.Atoi(Month)
	if err != nil {
		log.Errorln(pkgName, err)
	}
	var SqlStr string
	if Type == "Quarter" {

		ListQuarter := ListQuarter(YearInt, QuarterInt)
		for n, l := range ListQuarter {
			// YearInt, err := strconv.Atoi(l.Year)
			// if err != nil {
			// 	log.Errorln(pkgName, err)
			// }
			// QuarterInt, err := strconv.Atoi(l.Quarter)
			// if err != nil {
			// 	log.Errorln(pkgName, err)
			// }
			SqlStr += fmt.Sprintf(`
			select
			year(ContractStartDate) as year_chart,
			quarter(ContractStartDate) as quarter_chart,
			sum(revenue)/sum(eng_cost) as sale_factor
		from (
			SELECT sonumber,sum(PeriodAmount) as revenue, ContractStartDate,
				sum(case
				   when sale_factor != 0 and sale_factor is not null then PeriodAmount/sale_factor
				   else 0 end
				) as eng_cost
			FROM so_mssql
			where sale_code in (%s) 
				and year(ContractStartDate) = %d
				and quarter(ContractStartDate) = %d
				group by sonumber
		) tb_so`, StaffChildStr, l.Year, l.Quarter)
			if n != 4 {
				SqlStr += ` union `
			}
		}
		var Result []SoSFChart
		if err := dbSale.Ctx().Raw(SqlStr).Scan(&Result).Error; err != nil {
			log.Errorln("Error Get SoSFChart", err)
			return nil
		}

		return Result
	} else {
		ListMonth := ListMonth(YearInt, MonthInt)
		for n, l := range ListMonth {
			SqlStr += fmt.Sprintf(`
			select
                	year(ContractStartDate) as year_chart,
                    month(ContractStartDate) as month_chart,
                    sum(revenue)/sum(eng_cost) as sale_factor
                from (
                    SELECT sonumber,sum(PeriodAmount) as revenue, ContractStartDate,
                    	sum(case
                           when sale_factor != 0 and sale_factor is not null then PeriodAmount/sale_factor
                           else 0 end
                        ) as eng_cost
                    FROM so_mssql
                    where sale_code in (%s)
                    	and year(ContractStartDate) = %d
                    	and month(ContractStartDate) = %d
                    	group by sonumber
                ) tb_so`, StaffChildStr, l.Year, l.Month)
			if n != 4 {
				SqlStr += ` union `
			}
		}
		var Result []SoSFChart
		if err := dbSale.Ctx().Raw(SqlStr).Scan(&Result).Error; err != nil {
			log.Errorln("Error Get SoSFChart", err)
			return nil
		}

		return Result
	}
}
