package api

import (
	"fmt"
	"net/http"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/util"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

func CheckPermissionBaseSale(id string, filter string) ([]string, error) {
	var user []m.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {
		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department NOT IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}
		return listStaffId, nil
	} else {
		var staffCheck []m.StaffInfo
		sqlCheck := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id  = ?`)
		if err := dbSale.Ctx().Raw(sqlCheck, id).Scan(&staffCheck).Error; err != nil {
			return nil, err
		}
		if len(staffCheck) == 0 {
			return nil, nil
		}

		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department NOT IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}

		return listStaffId, nil
	}
}

func CheckPermissionBaseSale2(id string, filter string) ([]string, error) {
	var user []m.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {
		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) `)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}
		return listStaffId, nil
	} else {
		var staffCheck []m.StaffInfo
		sqlCheck := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id  = ?`)
		if err := dbSale.Ctx().Raw(sqlCheck, id).Scan(&staffCheck).Error; err != nil {
			return nil, err
		}
		if len(staffCheck) == 0 {
			return nil, nil
		}

		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department NOT IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}

		return listStaffId, nil
	}
}

func CheckPermissionKeyAccount(id string, filter string) ([]string, error) {
	var user []m.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {
		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department  IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "user select error")
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}
		return listStaffId, nil
	} else {
		var staffCheck []m.StaffInfo
		sqlCheck := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id  = ?`)
		if err := dbSale.Ctx().Raw(sqlCheck, id).Scan(&staffCheck).Error; err != nil {
			return nil, err
		}
		if len(staffCheck) == 0 {
			return nil, nil
		}
		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department  IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "user select error")
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}

		return listStaffId, nil
	}

}

func CheckPermissionRecovery(id string, filter string) ([]string, error) {
	var user []m.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {
		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department  IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "user select error")
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}
		return listStaffId, nil
	} else {
		var staffCheck []m.StaffInfo
		sqlCheck := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id  = ?`)
		if err := dbSale.Ctx().Raw(sqlCheck, id).Scan(&staffCheck).Error; err != nil {
			return nil, err
		}
		if len(staffCheck) == 0 {
			return nil, nil
		}

		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department  IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "user select error")
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}

		return listStaffId, nil
	}
}

func CheckPermissionTeamLead(id string, filter string) ([]string, error) {
	var user []m.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {
		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id from staff_info where staff_id NOT IN (?) and  staff_child <> '' and department not in('Up&Cross 2', 'Up&Cross 1', 'Retention', 'Sale JV', '???????????? Up and Cross Sales');`)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "user select error")
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}
		return listStaffId, nil
	} else {
		var staffCheck m.StaffInfo
		sqlCheck := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE  staff_id  = ? and  staff_child <> ''`)
		if err := dbSale.Ctx().Raw(sqlCheck, id).Scan(&staffCheck).Error; err != nil {
			return nil, err
		}
		if strings.TrimSpace(staffCheck.StaffChild) == "" {
			return nil, nil
		}

		var staff []m.StaffInfo
		sql := fmt.Sprintf(`SELECT staff_id from staff_info where staff_id NOT IN (?) and  staff_child <> '' and department not in('Up&Cross 2', 'Up&Cross 1', 'Retention', 'Sale JV', '???????????? Up and Cross Sales');`)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "user select error")
			return nil, err
		}

		var listStaffId []string
		for _, s := range staff {
			listStaffId = append(listStaffId, s.StaffId)
		}

		return listStaffId, nil
	}
}

func CheckTeamLeadEndPoint(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	var user []m.UserInfo
	if err := dbSale.Ctx().Raw(`SELECT * from user_info WHERE role = 'admin' and staff_id = ?`, id).Scan(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorln(pkgName, err, "user select error")
			return echo.ErrInternalServerError
		}
	}

	if len(user) != 0 {
		return c.JSON(http.StatusOK, true)
	}

	var staff m.StaffInfo
	if err := dbSale.Ctx().Raw(`SELECT staff_id,staff_child from staff_info WHERE staff_id = ?`, id).Scan(&staff).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			log.Errorln(pkgName, err, "user is not teamlead")
			return c.JSON(http.StatusOK, false)
		}
		log.Errorln(pkgName, err, "user select error")
		return echo.ErrInternalServerError
	}
	if staff.StaffChild != "" {
		return c.JSON(http.StatusOK, true)
	}
	return c.JSON(http.StatusOK, false)
}
