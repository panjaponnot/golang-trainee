package api

import (
	"fmt"
	"sale_ranking/pkg/log"
	"strconv"
	"net/http"
	"strings"
	// "time"
	"sync"

	"github.com/labstack/echo/v4"
)

func CostSheet_Status(c echo.Context) error{
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	StaffID := strings.TrimSpace(c.QueryParam("staffid"))
	Status := strings.TrimSpace(c.QueryParam("status"))
	Tracking_id := strings.TrimSpace(c.QueryParam("tracking_id"))
	Doc_id := strings.TrimSpace(c.QueryParam("doc_id"))
	DocumentJson := strings.TrimSpace(c.QueryParam("documentJson"))
	Doc_number_eform := strings.TrimSpace(c.QueryParam("doc_number_eform"))
	Customer_ID := strings.TrimSpace(c.QueryParam("Customer_ID"))
	Cusname_thai := strings.TrimSpace(c.QueryParam("Cusname_thai"))
	Cusname_Eng := strings.TrimSpace(c.QueryParam("Cusname_Eng"))
	ID_PreSale := strings.TrimSpace(c.QueryParam("ID_PreSale"))
	Cvm_id := strings.TrimSpace(c.QueryParam("cvm_id"))
	Bussiness_type := strings.TrimSpace(c.QueryParam("bussiness_type"))
	Sale_Team := strings.TrimSpace(c.QueryParam("Sale_Team"))
	Job_Status := strings.TrimSpace(c.QueryParam("Job_Status"))
	SO_Type := strings.TrimSpace(c.QueryParam("SO_Type"))
	Sales_Name := strings.TrimSpace(c.QueryParam("Sales_Name"))
	Sales_Surname := strings.TrimSpace(c.QueryParam("Sales_Surname"))
	EmployeeID := strings.TrimSpace(c.QueryParam("EmployeeID"))
	Status_eform := strings.TrimSpace(c.QueryParam("status_eform"))

	type Costsheet_Data struct {
		Total_Revenue_Month			string	`json:"Total_Revenue_Month" gorm:"column:Total_Revenue_Month"`
	}

	dataResult := struct {
		AAA	 interface{}
		AAB  interface{}
		AAC  interface{}
		AAD	 interface{}
		AAE  interface{}
		AAF  interface{}
	}{}

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(6)

	go func(){
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.ID_Presale = si.staff_id
		LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from paperless%' AND smt.SDPropertyCS28 is not null 
		AND smt.SDPropertyCS28 not like '' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
		DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
		Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
		Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
		Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if StaffID != ""{
				sql = sql+` si.staff_id like '`+StaffID+`' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` ci.status like '`+Status+`' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Tracking_id != ""{
				sql = sql+` ci.tracking_id like '`+Tracking_id+`' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_id != ""{
				sql = sql+` ci.doc_id like '`+Doc_id+`' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if DocumentJson != ""{
				sql = sql+` ci.documentJson like '`+DocumentJson+`' `
				if Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_number_eform != ""{
				sql = sql+` ci.doc_number_eform like '`+Doc_number_eform+`' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` ci.Customer_ID like '`+Customer_ID+`' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_thai != ""{
				sql = sql+` ci.Cusname_thai like '%`+Cusname_thai+`%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_Eng != ""{
				sql = sql+` ci.Cusname_Eng like '%`+Cusname_Eng+`%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if ID_PreSale != ""{
				sql = sql+` ci.ID_PreSale like '%`+ID_PreSale+`%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cvm_id != ""{
				sql = sql+` ci.cvm_id like '`+Cvm_id+`' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Bussiness_type != ""{
				sql = sql+` ci.Business_type like '%`+Bussiness_type+`%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sale_Team != ""{
				sql = sql+` ci.sale_team like '%`+Sale_Team+`%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Job_Status != ""{
				sql = sql+` ci.Job_Status like '%`+Job_Status+`%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if SO_Type != ""{
				sql = sql+` ci.SO_Type like '`+SO_Type+`' `
				if Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Name != ""{
				sql = sql+` ci.Sales_Name like '`+Sales_Name+`' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Surname != ""{
				sql = sql+` ci.Sales_Surname like '`+Sales_Surname+`' `
				if EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if EmployeeID != ""{
				sql = sql+` ci.EmployeeID like '`+EmployeeID+`' `
				if Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status_eform != ""{
				sql = sql+` ci.status_eform like '`+Status_eform+`' `
			}
		}
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA,_ := strconv.ParseFloat(v.Total_Revenue_Month,64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.AAA = map[string]interface{}{
			"CountBilling": Count_Costsheet,
			"PeriodAmount": TRM_All,
			"BLSCDocNo": "SO Compelte",
		}
		wg.Done()

	}()
	go func(){
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.ID_Presale = si.staff_id
		LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from paperless%' AND smt.SDPropertyCS28 is null `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
		DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
		Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
		Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
		Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if StaffID != ""{
				sql = sql+` si.staff_id like '`+StaffID+`' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` ci.status like '`+Status+`' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Tracking_id != ""{
				sql = sql+` ci.tracking_id like '`+Tracking_id+`' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_id != ""{
				sql = sql+` ci.doc_id like '`+Doc_id+`' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if DocumentJson != ""{
				sql = sql+` ci.documentJson like '`+DocumentJson+`' `
				if Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_number_eform != ""{
				sql = sql+` ci.doc_number_eform like '`+Doc_number_eform+`' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` ci.Customer_ID like '`+Customer_ID+`' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_thai != ""{
				sql = sql+` ci.Cusname_thai like '%`+Cusname_thai+`%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_Eng != ""{
				sql = sql+` ci.Cusname_Eng like '%`+Cusname_Eng+`%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if ID_PreSale != ""{
				sql = sql+` ci.ID_PreSale like '%`+ID_PreSale+`%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cvm_id != ""{
				sql = sql+` ci.cvm_id like '`+Cvm_id+`' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Bussiness_type != ""{
				sql = sql+` ci.Business_type like '%`+Bussiness_type+`%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sale_Team != ""{
				sql = sql+` ci.sale_team like '%`+Sale_Team+`%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Job_Status != ""{
				sql = sql+` ci.Job_Status like '%`+Job_Status+`%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if SO_Type != ""{
				sql = sql+` ci.SO_Type like '`+SO_Type+`' `
				if Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Name != ""{
				sql = sql+` ci.Sales_Name like '`+Sales_Name+`' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Surname != ""{
				sql = sql+` ci.Sales_Surname like '`+Sales_Surname+`' `
				if EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if EmployeeID != ""{
				sql = sql+` ci.EmployeeID like '`+EmployeeID+`' `
				if Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status_eform != ""{
				sql = sql+` ci.status_eform like '`+Status_eform+`' `
			}
		}
		sql = sql+` or smt.SDPropertyCS28 like '' AND ci.status_eform like '%Complete from paperless%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
		DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
		Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
		Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
		Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if StaffID != ""{
				sql = sql+` si.staff_id like '`+StaffID+`' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` ci.status like '`+Status+`' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Tracking_id != ""{
				sql = sql+` ci.tracking_id like '`+Tracking_id+`' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_id != ""{
				sql = sql+` ci.doc_id like '`+Doc_id+`' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if DocumentJson != ""{
				sql = sql+` ci.documentJson like '`+DocumentJson+`' `
				if Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_number_eform != ""{
				sql = sql+` ci.doc_number_eform like '`+Doc_number_eform+`' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_thai != ""{
				sql = sql+` ci.Cusname_thai like '%`+Cusname_thai+`%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_Eng != ""{
				sql = sql+` ci.Cusname_Eng like '%`+Cusname_Eng+`%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if ID_PreSale != ""{
				sql = sql+` ci.ID_PreSale like '%`+ID_PreSale+`%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cvm_id != ""{
				sql = sql+` ci.cvm_id like '`+Cvm_id+`' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Bussiness_type != ""{
				sql = sql+` ci.Bussiness_type like '`+Bussiness_type+`' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sale_Team != ""{
				sql = sql+` ci.Sale_Team like '`+Sale_Team+`' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Job_Status != ""{
				sql = sql+` ci.Job_Status like '`+Job_Status+`' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if SO_Type != ""{
				sql = sql+` ci.SO_Type like '`+SO_Type+`' `
				if Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Name != ""{
				sql = sql+` ci.Sales_Name like '`+Sales_Name+`' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Surname != ""{
				sql = sql+` ci.Sales_Surname like '`+Sales_Surname+`' `
				if EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if EmployeeID != ""{
				sql = sql+` ci.EmployeeID like '`+EmployeeID+`' `
				if Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status_eform != ""{
				sql = sql+` ci.status_eform like '`+Status_eform+`' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA,_ := strconv.ParseFloat(v.Total_Revenue_Month,64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.AAB = map[string]interface{}{
			"CountBilling": Count_Costsheet,
			"PeriodAmount": TRM_All,
			"BLSCDocNo": "Compelte from paperless",
		}
		wg.Done()
	}()
	go func(){
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.ID_Presale = si.staff_id
		LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from eform%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
		DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
		Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
		Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
		Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if StaffID != ""{
				sql = sql+` si.staff_id like '`+StaffID+`' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` ci.status like '`+Status+`' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Tracking_id != ""{
				sql = sql+` ci.tracking_id like '`+Tracking_id+`' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_id != ""{
				sql = sql+` ci.doc_id like '`+Doc_id+`' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if DocumentJson != ""{
				sql = sql+` ci.documentJson like '`+DocumentJson+`' `
				if Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_number_eform != ""{
				sql = sql+` ci.doc_number_eform like '`+Doc_number_eform+`' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` ci.Customer_ID like '`+Customer_ID+`' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_thai != ""{
				sql = sql+` ci.Cusname_thai like '%`+Cusname_thai+`%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_Eng != ""{
				sql = sql+` ci.Cusname_Eng like '%`+Cusname_Eng+`%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if ID_PreSale != ""{
				sql = sql+` ci.ID_PreSale like '%`+ID_PreSale+`%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cvm_id != ""{
				sql = sql+` ci.cvm_id like '`+Cvm_id+`' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Bussiness_type != ""{
				sql = sql+` ci.Business_type like '%`+Bussiness_type+`%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sale_Team != ""{
				sql = sql+` ci.sale_team like '%`+Sale_Team+`%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Job_Status != ""{
				sql = sql+` ci.Job_Status like '%`+Job_Status+`%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if SO_Type != ""{
				sql = sql+` ci.SO_Type like '`+SO_Type+`' `
				if Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Name != ""{
				sql = sql+` ci.Sales_Name like '`+Sales_Name+`' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Surname != ""{
				sql = sql+` ci.Sales_Surname like '`+Sales_Surname+`' `
				if EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if EmployeeID != ""{
				sql = sql+` ci.EmployeeID like '`+EmployeeID+`' `
				if Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status_eform != ""{
				sql = sql+` ci.status_eform like '`+Status_eform+`' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA,_ := strconv.ParseFloat(v.Total_Revenue_Month,64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.AAC = map[string]interface{}{
			"CountBilling": Count_Costsheet,
			"PeriodAmount": TRM_All,
			"BLSCDocNo": "Complete from eform",
		}
		wg.Done()
	}()
	go func(){
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.ID_Presale = si.staff_id
		LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Onprocess%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
		DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
		Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
		Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
		Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if StaffID != ""{
				sql = sql+` si.staff_id like '`+StaffID+`' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` ci.status like '`+Status+`' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Tracking_id != ""{
				sql = sql+` ci.tracking_id like '`+Tracking_id+`' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_id != ""{
				sql = sql+` ci.doc_id like '`+Doc_id+`' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if DocumentJson != ""{
				sql = sql+` ci.documentJson like '`+DocumentJson+`' `
				if Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_number_eform != ""{
				sql = sql+` ci.doc_number_eform like '`+Doc_number_eform+`' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` ci.Customer_ID like '`+Customer_ID+`' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_thai != ""{
				sql = sql+` ci.Cusname_thai like '%`+Cusname_thai+`%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_Eng != ""{
				sql = sql+` ci.Cusname_Eng like '%`+Cusname_Eng+`%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if ID_PreSale != ""{
				sql = sql+` ci.ID_PreSale like '%`+ID_PreSale+`%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cvm_id != ""{
				sql = sql+` ci.cvm_id like '`+Cvm_id+`' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Bussiness_type != ""{
				sql = sql+` ci.Business_type like '%`+Bussiness_type+`%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sale_Team != ""{
				sql = sql+` ci.sale_team like '%`+Sale_Team+`%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Job_Status != ""{
				sql = sql+` ci.Job_Status like '%`+Job_Status+`%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if SO_Type != ""{
				sql = sql+` ci.SO_Type like '`+SO_Type+`' `
				if Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Name != ""{
				sql = sql+` ci.Sales_Name like '`+Sales_Name+`' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Surname != ""{
				sql = sql+` ci.Sales_Surname like '`+Sales_Surname+`' `
				if EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if EmployeeID != ""{
				sql = sql+` ci.EmployeeID like '`+EmployeeID+`' `
				if Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status_eform != ""{
				sql = sql+` ci.status_eform like '`+Status_eform+`' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA,_ := strconv.ParseFloat(v.Total_Revenue_Month,64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.AAD = map[string]interface{}{
			"CountBilling": Count_Costsheet,
			"PeriodAmount": TRM_All,
			"BLSCDocNo": "Onprocess",
		}
		wg.Done()
	}()
	go func(){
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.ID_Presale = si.staff_id
		LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Cancel%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
		DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
		Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
		Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
		Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if StaffID != ""{
				sql = sql+` si.staff_id like '`+StaffID+`' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` ci.status like '`+Status+`' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Tracking_id != ""{
				sql = sql+` ci.tracking_id like '`+Tracking_id+`' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_id != ""{
				sql = sql+` ci.doc_id like '`+Doc_id+`' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if DocumentJson != ""{
				sql = sql+` ci.documentJson like '`+DocumentJson+`' `
				if Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_number_eform != ""{
				sql = sql+` ci.doc_number_eform like '`+Doc_number_eform+`' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` ci.Customer_ID like '`+Customer_ID+`' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_thai != ""{
				sql = sql+` ci.Cusname_thai like '%`+Cusname_thai+`%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_Eng != ""{
				sql = sql+` ci.Cusname_Eng like '%`+Cusname_Eng+`%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if ID_PreSale != ""{
				sql = sql+` ci.ID_PreSale like '%`+ID_PreSale+`%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cvm_id != ""{
				sql = sql+` ci.cvm_id like '`+Cvm_id+`' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Bussiness_type != ""{
				sql = sql+` ci.Business_type like '%`+Bussiness_type+`%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sale_Team != ""{
				sql = sql+` ci.sale_team like '%`+Sale_Team+`%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Job_Status != ""{
				sql = sql+` ci.Job_Status like '%`+Job_Status+`%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if SO_Type != ""{
				sql = sql+` ci.SO_Type like '`+SO_Type+`' `
				if Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Name != ""{
				sql = sql+` ci.Sales_Name like '`+Sales_Name+`' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Surname != ""{
				sql = sql+` ci.Sales_Surname like '`+Sales_Surname+`' `
				if EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if EmployeeID != ""{
				sql = sql+` ci.EmployeeID like '`+EmployeeID+`' `
				if Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status_eform != ""{
				sql = sql+` ci.status_eform like '`+Status_eform+`' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA,_ := strconv.ParseFloat(v.Total_Revenue_Month,64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.AAE = map[string]interface{}{
			"CountBilling": Count_Costsheet,
			"PeriodAmount": TRM_All,
			"BLSCDocNo": "Cancel",
		}
		wg.Done()
	}()
	go func(){
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.ID_Presale = si.staff_id
		LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Reject%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
		DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
		Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
		Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
		Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` ci.StartDate_P1 >= '`+St_date+`' AND ci.StartDate_P1 <= '`+En_date+`' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` ci.EndDate_P1 <= '`+En_date+`' AND ci.EndDate_P1 >= '`+St_date+`' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
				DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if StaffID != ""{
				sql = sql+` si.staff_id like '`+StaffID+`' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` ci.status like '`+Status+`' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" || 
				Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
				Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
				Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Tracking_id != ""{
				sql = sql+` ci.tracking_id like '`+Tracking_id+`' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_id != ""{
				sql = sql+` ci.doc_id like '`+Doc_id+`' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if DocumentJson != ""{
				sql = sql+` ci.documentJson like '`+DocumentJson+`' `
				if Doc_number_eform != "" || Customer_ID != "" || 
				Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || 
				Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Doc_number_eform != ""{
				sql = sql+` ci.doc_number_eform like '`+Doc_number_eform+`' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` ci.Customer_ID like '`+Customer_ID+`' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || 
				Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_thai != ""{
				sql = sql+` ci.Cusname_thai like '%`+Cusname_thai+`%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cusname_Eng != ""{
				sql = sql+` ci.Cusname_Eng like '%`+Cusname_Eng+`%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
				Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if ID_PreSale != ""{
				sql = sql+` ci.ID_PreSale like '%`+ID_PreSale+`%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Cvm_id != ""{
				sql = sql+` ci.cvm_id like '`+Cvm_id+`' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Bussiness_type != ""{
				sql = sql+` ci.Business_type like '%`+Bussiness_type+`%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" || 
				Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sale_Team != ""{
				sql = sql+` ci.sale_team like '%`+Sale_Team+`%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Job_Status != ""{
				sql = sql+` ci.Job_Status like '%`+Job_Status+`%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if SO_Type != ""{
				sql = sql+` ci.SO_Type like '`+SO_Type+`' `
				if Sales_Name != "" || Sales_Surname != "" || 
				EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Name != ""{
				sql = sql+` ci.Sales_Name like '`+Sales_Name+`' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Sales_Surname != ""{
				sql = sql+` ci.Sales_Surname like '`+Sales_Surname+`' `
				if EmployeeID != "" || Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if EmployeeID != ""{
				sql = sql+` ci.EmployeeID like '`+EmployeeID+`' `
				if Status_eform != ""{
					sql = sql+` AND `
				}
			}
			if Status_eform != ""{
				sql = sql+` ci.status_eform like '`+Status_eform+`' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA,_ := strconv.ParseFloat(v.Total_Revenue_Month,64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.AAF = map[string]interface{}{
			"CountBilling": Count_Costsheet,
			"PeriodAmount": TRM_All,
			"BLSCDocNo": "Reject",
		}
		wg.Done()
	}()
	wg.Wait()

	return c.JSON(http.StatusOK,dataResult)
}

func Invoice_Status(c echo.Context) error{

	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	Sonumber := strings.TrimSpace(c.QueryParam("sonumber"))
	Staff_id := strings.TrimSpace(c.QueryParam("staff_id"))
	SDPropertyCS28 := strings.TrimSpace(c.QueryParam("SDPropertyCS28"))
	So_Web_Status := strings.TrimSpace(c.QueryParam("So_Web_Status"))
	BLSCDocNo := strings.TrimSpace(c.QueryParam("BLSCDocNo"))
	GetCN := strings.TrimSpace(c.QueryParam("GetCN"))
	INCSCDocNo := strings.TrimSpace(c.QueryParam("INCSCDocNo"))
	Customer_ID := strings.TrimSpace(c.QueryParam("Customer_ID"))
	Customer_Name := strings.TrimSpace(c.QueryParam("Customer_Name"))
	Sale_code := strings.TrimSpace(c.QueryParam("sale_code"))
	Sale_name := strings.TrimSpace(c.QueryParam("sale_name"))
	Sale_team := strings.TrimSpace(c.QueryParam("sale_team"))
	Sale_lead := strings.TrimSpace(c.QueryParam("sale_lead"))
	Active_Inactive := strings.TrimSpace(c.QueryParam("Active_Inactive"))
	So_refer := strings.TrimSpace(c.QueryParam("so_refer"))

	type Invoice_Data struct {
		PeriodAmount		float64	`json:"PeriodAmount" gorm:"column:PeriodAmount"`
	}

	dataCount := struct {
		AAA	 interface{}
		AAB  interface{}
		AAC  interface{}
	}{}
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func(){
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		var dataRaw []Invoice_Data
		sql:= `select smt.PeriodAmount from so_mssql_test smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is not null AND smt.GetCN not like ''`
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
		So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
		Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
		Active_Inactive != "" || So_refer != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sonumber != ""{
				sql = sql+` smt.sonumber like '`+Sonumber+`' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" || 
				BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id +`' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if SDPropertyCS28 != ""{
				sql = sql+` smt.SDPropertyCS28 like '`+SDPropertyCS28+`' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_Web_Status != ""{
				sql = sql+` smt.So_Web_Status like '`+So_Web_Status+`' `
				if BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if BLSCDocNo != ""{
				sql = sql+` smt.BLSCDocNo like '`+BLSCDocNo+`' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if GetCN != ""{
				sql = sql+` smt.GetCN like '`+GetCN+`' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if INCSCDocNo != ""{
				sql = sql+` smt.INCSCDocNo like '`+INCSCDocNo+`' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` smt.Customer_ID like '`+Customer_ID+`' `
				if Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_Name != ""{
				sql = sql+` smt.Customer_Name like '`+Customer_Name+`' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_code != ""{
				sql = sql+` smt.sale_code like '`+Sale_code+`' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_name != ""{
				sql = sql+` smt.sale_name like '`+Sale_name+`' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_team != ""{
				sql = sql+` smt.sale_team like '`+Sale_team+`' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_lead != ""{
				sql = sql+` smt.sale_lead like '`+Sale_lead+`' `
				if Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Active_Inactive != ""{
				sql = sql+` smt.Active_Inactive  like '`+Active_Inactive+`' `
				if So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_refer != ""{
				sql = sql+` smt.so_refer  like '`+So_refer+`' `
			}
		}
		sql = sql+` GROUP BY smt.sonumber`

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.AAB = map[string]interface{}{
			"CountBilling": Count_Invoice,
			"PeriodAmount": Total_PA,
			"BLSCDocNo": "ลดหนี้",
		}
		wg.Done()
	}()
	go func(){
		var dataRaw []Invoice_Data
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		sql:= `select smt.PeriodAmount from so_mssql_test smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is null AND smt.BLSCDocNo is not null AND smt.BLSCDocNo not like '' `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
		So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
		Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
		Active_Inactive != "" || So_refer != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sonumber != ""{
				sql = sql+` smt.sonumber like '`+Sonumber+`' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" || 
				BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id +`' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if SDPropertyCS28 != ""{
				sql = sql+` smt.SDPropertyCS28 like '`+SDPropertyCS28+`' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_Web_Status != ""{
				sql = sql+` smt.So_Web_Status like '`+So_Web_Status+`' `
				if BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if BLSCDocNo != ""{
				sql = sql+` smt.BLSCDocNo like '`+BLSCDocNo+`' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if GetCN != ""{
				sql = sql+` smt.GetCN like '`+GetCN+`' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if INCSCDocNo != ""{
				sql = sql+` smt.INCSCDocNo like '`+INCSCDocNo+`' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` smt.Customer_ID like '`+Customer_ID+`' `
				if Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_Name != ""{
				sql = sql+` smt.Customer_Name like '`+Customer_Name+`' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_code != ""{
				sql = sql+` smt.sale_code like '`+Sale_code+`' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_name != ""{
				sql = sql+` smt.sale_name like '`+Sale_name+`' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_team != ""{
				sql = sql+` smt.sale_team like '`+Sale_team+`' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_lead != ""{
				sql = sql+` smt.sale_lead like '`+Sale_lead+`' `
				if Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Active_Inactive != ""{
				sql = sql+` smt.Active_Inactive  like '`+Active_Inactive+`' `
				if So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_refer != ""{
				sql = sql+` smt.so_refer  like '`+So_refer+`' `
			}
		} 
		sql = sql+` OR smt.GetCN like '' AND smt.BLSCDocNo is not null AND smt.BLSCDocNo not like ''`
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
		So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
		Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
		Active_Inactive != "" || So_refer != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sonumber != ""{
				sql = sql+` smt.sonumber like '`+Sonumber+`' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" || 
				BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id +`' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if SDPropertyCS28 != ""{
				sql = sql+` smt.SDPropertyCS28 like '`+SDPropertyCS28+`' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_Web_Status != ""{
				sql = sql+` smt.So_Web_Status like '`+So_Web_Status+`' `
				if BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if BLSCDocNo != ""{
				sql = sql+` smt.BLSCDocNo like '`+BLSCDocNo+`' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if GetCN != ""{
				sql = sql+` smt.GetCN like '`+GetCN+`' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if INCSCDocNo != ""{
				sql = sql+` smt.INCSCDocNo like '`+INCSCDocNo+`' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` smt.Customer_ID like '`+Customer_ID+`' `
				if Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_Name != ""{
				sql = sql+` smt.Customer_Name like '`+Customer_Name+`' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_code != ""{
				sql = sql+` smt.sale_code like '`+Sale_code+`' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_name != ""{
				sql = sql+` smt.sale_name like '`+Sale_name+`' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_team != ""{
				sql = sql+` smt.sale_team like '`+Sale_team+`' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_lead != ""{
				sql = sql+` smt.sale_lead like '`+Sale_lead+`' `
				if Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Active_Inactive != ""{
				sql = sql+` smt.Active_Inactive  like '`+Active_Inactive+`' `
				if So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_refer != ""{
				sql = sql+` smt.so_refer  like '`+So_refer+`' `
			}
		}
		sql = sql+` GROUP BY smt.sonumber`

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.AAA = map[string]interface{}{
			"CountBilling": Count_Invoice,
			"PeriodAmount": Total_PA,
			"BLSCDocNo": "ออก invoice เสร้จสิ้น",
		}
		wg.Done()
	}()
	go func(){
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		var dataRaw []Invoice_Data
		sql:= `select smt.PeriodAmount from so_mssql_test smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is null AND smt.BLSCDocNo is null `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
		So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
		Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
		Active_Inactive != "" || So_refer != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sonumber != ""{
				sql = sql+` smt.sonumber like '`+Sonumber+`' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" || 
				BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id +`' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if SDPropertyCS28 != ""{
				sql = sql+` smt.SDPropertyCS28 like '`+SDPropertyCS28+`' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_Web_Status != ""{
				sql = sql+` smt.So_Web_Status like '`+So_Web_Status+`' `
				if BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if BLSCDocNo != ""{
				sql = sql+` smt.BLSCDocNo like '`+BLSCDocNo+`' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if GetCN != ""{
				sql = sql+` smt.GetCN like '`+GetCN+`' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if INCSCDocNo != ""{
				sql = sql+` smt.INCSCDocNo like '`+INCSCDocNo+`' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` smt.Customer_ID like '`+Customer_ID+`' `
				if Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_Name != ""{
				sql = sql+` smt.Customer_Name like '`+Customer_Name+`' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_code != ""{
				sql = sql+` smt.sale_code like '`+Sale_code+`' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_name != ""{
				sql = sql+` smt.sale_name like '`+Sale_name+`' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_team != ""{
				sql = sql+` smt.sale_team like '`+Sale_team+`' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_lead != ""{
				sql = sql+` smt.sale_lead like '`+Sale_lead+`' `
				if Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Active_Inactive != ""{
				sql = sql+` smt.Active_Inactive  like '`+Active_Inactive+`' `
				if So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_refer != ""{
				sql = sql+` smt.so_refer  like '`+So_refer+`' `
			}
		}
		sql = sql+` OR smt.GetCN like '' AND smt.BLSCDocNo is null `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
		So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
		Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
		Active_Inactive != "" || So_refer != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sonumber != ""{
				sql = sql+` smt.sonumber like '`+Sonumber+`' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" || 
				BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id +`' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if SDPropertyCS28 != ""{
				sql = sql+` smt.SDPropertyCS28 like '`+SDPropertyCS28+`' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_Web_Status != ""{
				sql = sql+` smt.So_Web_Status like '`+So_Web_Status+`' `
				if BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if BLSCDocNo != ""{
				sql = sql+` smt.BLSCDocNo like '`+BLSCDocNo+`' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if GetCN != ""{
				sql = sql+` smt.GetCN like '`+GetCN+`' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if INCSCDocNo != ""{
				sql = sql+` smt.INCSCDocNo like '`+INCSCDocNo+`' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` smt.Customer_ID like '`+Customer_ID+`' `
				if Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_Name != ""{
				sql = sql+` smt.Customer_Name like '`+Customer_Name+`' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_code != ""{
				sql = sql+` smt.sale_code like '`+Sale_code+`' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_name != ""{
				sql = sql+` smt.sale_name like '`+Sale_name+`' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_team != ""{
				sql = sql+` smt.sale_team like '`+Sale_team+`' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_lead != ""{
				sql = sql+` smt.sale_lead like '`+Sale_lead+`' `
				if Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Active_Inactive != ""{
				sql = sql+` smt.Active_Inactive  like '`+Active_Inactive+`' `
				if So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_refer != ""{
				sql = sql+` smt.so_refer  like '`+So_refer+`' `
			}
		}
		sql = sql+` OR smt.GetCN like '' AND smt.BLSCDocNo like '' `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
		So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
		Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
		Active_Inactive != "" || So_refer != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sonumber != ""{
				sql = sql+` smt.sonumber like '`+Sonumber+`' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" || 
				BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id +`' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if SDPropertyCS28 != ""{
				sql = sql+` smt.SDPropertyCS28 like '`+SDPropertyCS28+`' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_Web_Status != ""{
				sql = sql+` smt.So_Web_Status like '`+So_Web_Status+`' `
				if BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if BLSCDocNo != ""{
				sql = sql+` smt.BLSCDocNo like '`+BLSCDocNo+`' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if GetCN != ""{
				sql = sql+` smt.GetCN like '`+GetCN+`' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if INCSCDocNo != ""{
				sql = sql+` smt.INCSCDocNo like '`+INCSCDocNo+`' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` smt.Customer_ID like '`+Customer_ID+`' `
				if Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_Name != ""{
				sql = sql+` smt.Customer_Name like '`+Customer_Name+`' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_code != ""{
				sql = sql+` smt.sale_code like '`+Sale_code+`' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_name != ""{
				sql = sql+` smt.sale_name like '`+Sale_name+`' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_team != ""{
				sql = sql+` smt.sale_team like '`+Sale_team+`' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_lead != ""{
				sql = sql+` smt.sale_lead like '`+Sale_lead+`' `
				if Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Active_Inactive != ""{
				sql = sql+` smt.Active_Inactive  like '`+Active_Inactive+`' `
				if So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_refer != ""{
				sql = sql+` smt.so_refer  like '`+So_refer+`' `
			}
		}
		sql = sql+` OR smt.GetCN is null AND smt.BLSCDocNo like '' `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
		So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
		Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
		Active_Inactive != "" || So_refer != ""{
		sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
				So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sonumber != ""{
				sql = sql+` smt.sonumber like '`+Sonumber+`' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" || 
				BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
				Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id +`' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if SDPropertyCS28 != ""{
				sql = sql+` smt.SDPropertyCS28 like '`+SDPropertyCS28+`' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_Web_Status != ""{
				sql = sql+` smt.So_Web_Status like '`+So_Web_Status+`' `
				if BLSCDocNo != "" || GetCN != "" || 
				INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if BLSCDocNo != ""{
				sql = sql+` smt.BLSCDocNo like '`+BLSCDocNo+`' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if GetCN != ""{
				sql = sql+` smt.GetCN like '`+GetCN+`' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if INCSCDocNo != ""{
				sql = sql+` smt.INCSCDocNo like '`+INCSCDocNo+`' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_ID != ""{
				sql = sql+` smt.Customer_ID like '`+Customer_ID+`' `
				if Customer_Name != "" || Sale_code != "" || 
				Sale_name != "" || Sale_team != "" || Sale_lead != "" ||Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Customer_Name != ""{
				sql = sql+` smt.Customer_Name like '`+Customer_Name+`' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_code != ""{
				sql = sql+` smt.sale_code like '`+Sale_code+`' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
				Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_name != ""{
				sql = sql+` smt.sale_name like '`+Sale_name+`' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_team != ""{
				sql = sql+` smt.sale_team like '`+Sale_team+`' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Sale_lead != ""{
				sql = sql+` smt.sale_lead like '`+Sale_lead+`' `
				if Active_Inactive != "" || So_refer != ""{
					sql = sql+` AND `
				}
			}
			if Active_Inactive != ""{
				sql = sql+` smt.Active_Inactive  like '`+Active_Inactive+`' `
				if So_refer != ""{
					sql = sql+` AND `
				}
			}
			if So_refer != ""{
				sql = sql+` smt.so_refer  like '`+So_refer+`' `
			}
		}
		sql = sql+` GROUP BY smt.sonumber `

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.AAC = map[string]interface{}{
			"CountBilling": Count_Invoice,
			"PeriodAmount": Total_PA,
			"BLSCDocNo": "ยังไม่ออก invoice",
		}
		wg.Done()
	}()
	wg.Wait()
	return c.JSON(http.StatusOK,dataCount)
}

func Billing_Status(c echo.Context) error{
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	Staff_id := strings.TrimSpace(c.QueryParam("staffid"))
	Invoice_no := strings.TrimSpace(c.QueryParam("invoice_no"))
	So_number := strings.TrimSpace(c.QueryParam("so_number"))
	Status := strings.TrimSpace(c.QueryParam("status"))
	Reason := strings.TrimSpace(c.QueryParam("reason"))

	type Billing_Data struct {
		PeriodAmount		float64	`json:"PeriodAmount" gorm:"column:PeriodAmount"`
	}
	dataCount := struct {
		AAA	 interface{}
		AAB  interface{}
		AAC  interface{}
		AAD	 interface{}
		AAE  interface{}
		AAF  interface{}
		AAG	 interface{}
		AAH  interface{}
		AAI  interface{}
	}{}
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(9)
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางบิลแล้ว%'`
		if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
		sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAA = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "วางบิลแล้ว",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' `
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAB = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "วางไม่ได้",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' and bi.reason like '%ลดหนี้%'`
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAC = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "ลดหนี้",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' and bi.reason like '%ติดส่งมอบงาน%' `
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAD = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "ติดส่งมอบงาน",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no 
			where bi.status like '%วางไม่ได้%' and bi.reason like '%อื่นๆ%'`
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAE = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "อื่นๆ",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' and bi.reason like '%แก้ไขใบแจ้งหนี้%'`
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAF = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "แก้ไขใบแจ้งหนี้",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' and bi.reason like '%ติด PO%'`
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAG = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "ติด PO",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' and bi.reason like '%ติดสัญญา%' `
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAH = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "ติดสัญญา",
		}
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql:= `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' and bi.reason like '%ติด Report%' `
			if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
		}
			sql = sql+`) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.AAI = map[string]interface{}{
			"CountBilling": CountBilling,
			"PeriodAmount": TotalPeriodAmount,
			"BLSCDocNo": "ติด Report",
		}
		wg.Done()
	}()
	wg.Wait()

	return c.JSON(http.StatusOK,dataCount)
}

func Reciept_Status(c echo.Context) error{
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	Staff_id := strings.TrimSpace(c.QueryParam("staffid"))
	Invoice_no := strings.TrimSpace(c.QueryParam("invoice_no"))
	So_number := strings.TrimSpace(c.QueryParam("so_number"))
	Status := strings.TrimSpace(c.QueryParam("status"))
	Reason := strings.TrimSpace(c.QueryParam("reason"))

	Reciept_Data := []struct {
		PeriodAmount		float64	`json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Count_Reciept		int		`json:"Count_Reciept" gorm:"column:Count_Reciept"`
		PeriodAmountF		float64	`json:"PeriodAmountF" gorm:"column:PeriodAmountF"`
		Count_RecieptF		int		`json:"Count_RecieptF" gorm:"column:Count_RecieptF"`
		// Invoice_status_name	string	`json:"invoice_status_name" gorm:"column:invoice_status_name"`
		// INCSCDocNo			string	`json:"INCSCDocNo" gorm:"column:INCSCDocNo"`
	}{}
	
	type reciept_Result_Data struct{
		CountReciept			string	`json:"CountReciept"`
		TotalPeriodAmount		string	`json:"TotalPeriodAmount"`
		Reciept_status			string	`json:"Reciept_status"`

	}
	
	var Reciept_Result_Data []reciept_Result_Data

	sql:= `select
	SUM(CASE
		WHEN RE.INCSCDocNo is not null AND RE.INCSCDocNo NOT LIKE ''
		THEN RE.PeriodAmount
		ELSE NULL
	END) as PeriodAmount, 
	COUNT(CASE
		WHEN RE.INCSCDocNo is not null AND RE.INCSCDocNo NOT LIKE ''
		THEN 1
		ELSE NULL
	END) as Count_Reciept,
	SUM(CASE
		WHEN RE.INCSCDocNo is null OR RE.INCSCDocNo LIKE ''
		THEN RE.PeriodAmount
		ELSE NULL
	END) as PeriodAmountF, 
	COUNT(CASE
		WHEN RE.INCSCDocNo is null OR RE.INCSCDocNo LIKE ''
		THEN 1
		ELSE NULL
	END) as Count_RecieptF
	from
	(select smt.PeriodAmount,smt.INCSCDocNo
	from so_mssql_test smt
	LEFT JOIN staff_info si on smt.sale_code = si.staff_id
	LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
	where bi.status like '%วางบิลแล้ว%' `
	
	if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" || 
			So_number != "" || Status != "" || Reason != ""{
			sql = sql+` AND `
			if St_date != ""{
				sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if En_date != ""{
				sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
				if Staff_id != "" || Invoice_no != "" || 
				So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Staff_id != ""{
				sql = sql+` si.staff_id like '`+Staff_id+`'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Invoice_no != ""{
				sql = sql+` bi.invoice_no like '`+Invoice_no+`'`
				if So_number != "" || Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if So_number != ""{
				sql = sql+` bi.so_number like '`+So_number+`'`
				if Status != "" || Reason != ""{
					sql = sql+` AND `
				}
			}
			if Status != ""{
				sql = sql+` bi.status like '`+Status+`'`
				if Reason != ""{
					sql = sql+` AND `
				}
			}
			if Reason != ""{
				sql = sql+` bi.reason like '`+Reason+`'`
			}
	}

	sql = sql+` )RE`
	
	if err := dbSale.Ctx().Raw(sql).Scan(&Reciept_Data).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	DataA := reciept_Result_Data{
		CountReciept: strconv.Itoa(Reciept_Data[0].Count_Reciept),
		TotalPeriodAmount: fmt.Sprintf("%f",Reciept_Data[0].PeriodAmount),
		Reciept_status: "วาง Reciept เสร้จสิ้น",
	}
	Reciept_Result_Data = append(Reciept_Result_Data,DataA)

	DataB := reciept_Result_Data{
		CountReciept: strconv.Itoa(Reciept_Data[0].Count_RecieptF),
		TotalPeriodAmount: fmt.Sprintf("%f",Reciept_Data[0].PeriodAmountF),
		Reciept_status: "ยังไม่วาง Reciept",
	}
	Reciept_Result_Data = append(Reciept_Result_Data,DataB)

	return c.JSON(http.StatusOK,Reciept_Result_Data)
}