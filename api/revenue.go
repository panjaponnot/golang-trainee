package api

import (
	"fmt"
	"sale_ranking/pkg/log"
	"net/http"
	"strings"
	"time"
	"strconv"

	"github.com/labstack/echo/v4"
)

func RevenueEndPoint(c echo.Context) error {
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	OneID := strings.TrimSpace(c.QueryParam("oneid"))
	Form_status := strings.TrimSpace(c.QueryParam("status"))

	rawData := []struct {
		Send_time 			string `json:"send_time" gorm:"column:send_time"`
		Status    			string `json:"status" gorm:"column:status"`
		Tracking_id 		string `json:"tracking_id" gorm:"column:tracking_id"`
		Doc_id 				string `json:"doc_id" gorm:"column:doc_id"`
		Update_time 		string `json:"update_time" gorm:"column:update_time"`
		Doc_number_eform	string `json:"doc_number_eform" gorm:"column:doc_number_eform"`
		Customer_ID			string `json:"Customer_ID" gorm:"column:Customer_ID"`
		Cusname_thai		string `json:"Cusname_thai" gorm:"column:Cusname_thai"`
		Cusname_Eng			string `json:"Cusname_Eng" gorm:"column:Cusname_Eng"`
		ID_PreSale			string `json:"ID_PreSale" gorm:"column:ID_PreSale"`
		Cvm_id				string `json:"cvm_id" gorm:"column:cvm_id"`
		Business_type		string `json:"Business_type" gorm:"column:Business_type"`
		Sale_Team			string `json:"Sale_Team" gorm:"column:Sale_Team"`
		Job_Status			string `json:"Job_Status" gorm:"column:Job_Status"`
		SO_Type				string `json:"SO_Type" gorm:"column:SO_Type"`
		Sales_Name			string `json:"Sales_Name" gorm:"column:Sales_Name"`
		Sales_Surname		string `json:"Sales_Surname" gorm:"column:Sales_Surname"`
		EmployeeID			string `json:"EmployeeID" gorm:"column:EmployeeID"`
		Total_Revenue_Month	string `json:"Total_Revenue_Month" gorm:"column:Total_Revenue_Month"`
		SaleFactors			string `json:"SaleFactors" gorm:"column:SaleFactors"`
		Revenue_Month		string `json:"Revenue_Month" gorm:"column:Revenue_Month"`
		Int_INET			string `json:"Int_INET" gorm:"column:Int_INET"`
		Ext_JV				string `json:"Ext_JV" gorm:"column:Ext_JV"`
		Ext					string `json:"Ext" gorm:"column:Ext"`
		StartDate_P1		string `json:"StartDate_P1" gorm:"column:StartDate_P1"`
		EndDate_P1			string `json:"EndDate_P1" gorm:"column:EndDate_P1"`
		Status_eform		string `json:"status_eform" gorm:"column:status_eform"`
		One_id				string `json:"one_id" gorm:"column:one_id"`
		Staff_id			string `json:"staff_id" gorm:"column:staff_id"`
		Prefix				string `json:"prefix" gorm:"column:prefix"`
		Fname				string `json:"fname" gorm:"column:fname"`
		Lname				string `json:"lname" gorm:"column:lname"`
		Nname				string `json:"nname" gorm:"column:nname"`
		Position			string `json:"position" gorm:"column:position"`
		Department			string `json:"department" gorm:"column:department"`
		Staff_child			string `json:"staff_child" gorm:"column:staff_child"`
		Revenue_Day			string
	}{}

	sql := `select * from costsheet_info ci left join staff_info si on ci.ID_PreSale = si.staff_id `
	if St_date != "" || En_date != "" || OneID != "" || Form_status != ""{
		sql = sql+` where `
		if St_date != ""{
			sql = sql+` ci.StartDate_P1 >= '`+St_date+`' `
			if En_date != "" || OneID != "" || Form_status != ""{
				sql = sql+` AND `
			}
		}
		if En_date != ""{
			sql = sql+` ci.EndDate_P1 >= '`+En_date+`' `
			if OneID != "" || Form_status != ""{
				sql = sql+` AND `
			}
		}
		if OneID != ""{
			sql = sql+` si.one_id like `+OneID+` `
			if Form_status != ""{
				sql = sql+` AND `
			}
		}
		if Form_status != ""{
			sql = sql+` ci.status_eform like '`+Form_status+`' `
		}

	}
	sql = sql+` limit 100 `
	if err := dbSale.Ctx().Raw(sql).Scan(&rawData).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	time1,_ := time.Parse("2006-01-02 00:00:00",St_date)
	time2,_ := time.Parse("2006-01-02 00:00:00",En_date)
	days := (time2.Sub(time1).Hours() / 24)+1

	for i , j := range rawData {
		if len(j.Total_Revenue_Month) > 0 && j.Total_Revenue_Month != ""{
			result,_ := strconv.ParseFloat(j.Total_Revenue_Month,64)
			result = result / days
			result_str := fmt.Sprintf("%f",result)
			rawData[i].Revenue_Day = result_str
		} else {
			rawData[i].Revenue_Day = "0.000000"
		}
	}

	// fmt.Println(days)

	return c.JSON(http.StatusOK, rawData)
}