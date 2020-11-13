package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/imagik"
	"sale_ranking/pkg/log"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
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

	var StaffIdList []m.StaffId
	if err := dbSale.Ctx().Raw(`SELECT staff_id FROM staff_info;`).Scan(&StaffIdList).Error; err != nil {
		log.Errorln("GetStaffIdList error :-", err)
	}

	for _, s := range StaffIdList {
		NamePic := strings.TrimSpace(s.StaffId)
		StaffId := strings.TrimSpace(s.StaffId)
		var StaffIds = fmt.Sprintf("%ss", StaffId)
		if CheckPictureUrl(NamePic) {
			log.Infoln("======>", NamePic)
			var b []byte
			var mimeType string
			var avatarUrl = fmt.Sprintf("https://intranet.inet.co.th/assets/upload/staff/%s.jpg", NamePic)
			if err := imagik.UrlGrabber(avatarUrl, nil, &b, &mimeType, 5); err != nil {
				log.Warnln("Get staff avatar error -:", err)
				return echo.ErrInternalServerError
			}
			src, err := imaging.Decode(bytes.NewReader(b))
			if err != nil {
				log.Warnln("Avatar decode error 1 -:", err)
				return echo.ErrInternalServerError
			}
			resized := imaging.Resize(src, 200, 250, imaging.Lanczos)
			var buf bytes.Buffer
			if err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(90)); err != nil {
				log.Warnln("Avatar encode error 2 -:", err)
				return echo.ErrInternalServerError
			}
			vars := m.StaffPicture{
				StaffId: StaffId,
				Img:     buf,
			}
			tx := dbSale.Ctx().Begin()
			if err := dbSale.Ctx().Table("staff_picture").Create(&vars).Error; err != nil {
				log.Errorln("Func Insert StaffPicture error :-", err)
				tx.Rollback()
				return err
			}
		} else if CheckPictureUrl(StaffIds) {
			log.Infoln("====2==>", StaffIds)
			var b []byte
			var mimeType string
			var avatarUrl = fmt.Sprintf("https://intranet.inet.co.th/assets/upload/staff/%s.jpg", NamePic)
			if err := imagik.UrlGrabber(avatarUrl, nil, &b, &mimeType, 5); err != nil {
				log.Warnln("Get staff avatar error -:", err)
				return echo.ErrInternalServerError
			}
			src, err := imaging.Decode(bytes.NewReader(b))
			if err != nil {
				log.Warnln("Avatar decode error -:", err)
				return echo.ErrInternalServerError
			}
			resized := imaging.Resize(src, 200, 250, imaging.Lanczos)
			var buf bytes.Buffer
			if err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(90)); err != nil {
				log.Warnln("Avatar encode error -:", err)
				return echo.ErrInternalServerError
			}
			vars := m.StaffPicture{
				StaffId: StaffId,
				Img:     buf,
			}
			tx := dbSale.Ctx().Begin()
			if err := dbSale.Ctx().Table("staff_picture").Create(&vars).Error; err != nil {
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
	// Url := fmt.Sprintf("https://intranet.inet.co.th/assets/upload/staff/%s.jpg", namepic)
	// resp, err := http.Get(Url)
	// if err != nil {
	// 	return false
	// }
	// if resp.StatusCode == 200 {
	// 	return true
	// } else {
	// 	return false
	// }

	// var b []byte
	// var mimeType string
	var avatarUrl = fmt.Sprintf("https://intranet.inet.co.th/assets/upload/staff/%s.jpg", namepic)
	// if err := imagik.UrlGrabber(avatarUrl, nil, &b, &mimeType, 5); err != nil {
	// 	log.Warnln("Get staff avatar error -:", err)
	// 	return false
	// }
	// return true

	// client := &http.Client{
	// 	CheckRedirect: redirectPolicyFunc,
	// }
	response, err := http.Get(avatarUrl)
	if err != nil {
		fmt.Println("HTTP call failed:", err)
		return false
	}
	// Don't forget, you're expected to close response body even if you don't want to read it.
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", response.StatusCode)
		// You may read / inspect response body
		return false
	}

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
		// 	if err := dbSale.Ctx().Raw(`SELECT id, ref_staff, year, quarter, goal_total, real_total, create_date, create_by
		// 	FROM goal_quarter
		// 	WHERE ref_staff = ?`, StaffId).Scan(&StaffGoalQuarter).Error; err != nil {
		// 		log.Errorln("GetStaffGoalQuarter error :-", err)
		// }

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
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "connect database error")
	// 	// return c.JSON(http.StatusInternalServerError, err)
	// }

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
		if err := dbSale.Ctx().Raw(`
		SELECT total_amount, goal_total, year, quarter, MONTH(?) as month
		FROM
			(
				WITH gq_data AS
				(
					SELECT goal_total, year, quarter, ref_staff
					FROM goal_quarter
					WHERE
						goal_quarter.year = YEAR(?) AND
						goal_quarter.quarter = CONCAT("Q", QUARTER(?))
				)
				SELECT so_amount.*, gq_data.*
				FROM
					(
						SELECT sonumber, sale_code, SUM(PeriodAmount) as total_amount
						FROM
							(
								SELECT sonumber, PeriodAmount, sale_code
								FROM so_mssql
								WHERE QUARTER(ContractStartDate) = QUARTER(?) AND
									MONTH(DATE(ContractStartDate)) <= MONTH(?) AND
									YEAR(DATE(ContractStartDate)) = YEAR(?) AND
									so_mssql.sale_code IN (?)
								GROUP BY sonumber
							) as so_data
					) as so_amount
				INNER JOIN gq_data ON gq_data.ref_staff = so_amount.sale_code
			) as staff_so_data
		;`, d, d, d, d, d, d, StaffChild).Scan(&StaffIdGoalQuarter).Error; err != nil {
			log.Errorln("GetStaffIdGoalQuarter error 1:-", err)
		}
		if len(StaffIdGoalQuarter) == 0 {
			if err := dbSale.Ctx().Raw(`
			SELECT goal_total, year, quarter, MONTH(?) as month
			FROM goal_quarter
			WHERE ref_staff IN (?)
			LIMIT 1;`, d, StaffChild).Scan(&StaffIdGoalQuarter).Error; err != nil {
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
