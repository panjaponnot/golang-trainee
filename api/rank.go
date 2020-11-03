package api

import (
	"fmt"
	"net/http"
	"sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/util"
	"strings"

	"github.com/labstack/echo/v4"
)

func CheckPermissionBaseSale(id string, filter string) (map[string][]string, error) {
	var user []model.UserInfo
	notSale := util.GetEnv("ACCOUNT_NOT_SALE", "")
	sqlUsr := `SELECT * from user_info WHERE role = 'admin' and staff_id = ?`
	if err := dbSale.Ctx().Raw(sqlUsr, id).Scan(&user).Error; err != nil {
		return nil, err
	}
	if len(user) != 0 {
		var staff []model.StaffInfo
		mapStaff := map[string][]string{}
		sql := fmt.Sprintf(`SELECT staff_id,staff_child from staff_info WHERE staff_id NOT IN (?) and department NOT IN  ( select department from staff_info where %s)`, filter)
		if err := dbSale.Ctx().Raw(sql, notSale).Scan(&staff).Error; err != nil {
			return nil, err
		}

		for _, s := range staff {
			var listStaffId []string
			if strings.TrimSpace(s.StaffChild) != "" {
				raw := strings.Split(s.StaffChild, ",")
				for _, id := range raw {
					listStaffId = append(listStaffId, id)
				}
				listStaffId = append(listStaffId, s.StaffId)
			} else {
				listStaffId = append(listStaffId, s.StaffId)
			}
			if _, ok := mapStaff[s.StaffId]; !ok {
				mapStaff[s.StaffId] = listStaffId
			}
		}
		return mapStaff, nil
	} else {
		mapStaff := map[string][]string{}
		staff := struct {
			StaffId    string `json:"staff_id"`
			StaffChild string `json:"staff_child"`
		}{}
		var staffAll []model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT staff_id,staff_child from staff_info where staff_id NOT IN (?);`, notSale).Scan(&staffAll).Error; err != nil {
			return nil, err
		}
		sql := fmt.Sprintf(`SELECT staff_id,staff_child,department from staff_info WHERE  department NOT IN  ( select department from staff_info where %s )`, filter)

		if err := dbSale.Ctx().Raw(sql).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
			return nil, nil
		}

		var rawdata []string
		if strings.TrimSpace(staff.StaffChild) != "" {
			raw := strings.Split(staff.StaffChild, ",")
			for _, id := range raw {
				rawdata = append(rawdata, id)
			}
			rawdata = append(rawdata, staff.StaffId)
		} else {
			rawdata = append(rawdata, staff.StaffId)
		}

		for _, v := range staffAll {
			for _, c := range rawdata {
				if v.StaffId == c {
					var listStaffId []string
					if strings.TrimSpace(v.StaffChild) != "" {
						raw := strings.Split(v.StaffChild, ",")
						for _, id := range raw {
							listStaffId = append(listStaffId, id)
						}
						listStaffId = append(listStaffId, v.StaffId)
					} else {
						listStaffId = append(listStaffId, v.StaffId)
					}
					if _, ok := mapStaff[v.StaffId]; !ok {
						mapStaff[v.StaffId] = listStaffId
					}
				}
			}
		}

		return mapStaff, nil
	}
	// return nil, nil
}

func GetRankingBaseSale(c echo.Context) error {
	filterDepart := strings.Split(util.GetEnv("CONDITION_BASE_SALE", ""), ",")

	fmt.Println("-=-=-=-=>", util.GetEnv("ACCOUNT_NOT_SALE", ""))
	var dFilter []string
	for _, v := range filterDepart {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	fmt.Println("=======>", finalFilter)

	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "connect database error")
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer dbSale.Close()

	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" {
		return c.JSON(http.StatusBadRequest, model.Result{Message: "invalid staff id"})
	}
	listStaffId, err := CheckPermissionBaseSale(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		log.Errorln(pkgName, err, "func check permission error :-")
		return c.JSON(http.StatusInternalServerError, model.Result{Error: "check permission error"})
	}
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}
	return c.JSON(http.StatusOK, listStaffId)
}
