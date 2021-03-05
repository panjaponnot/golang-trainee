package api

import (
	"net/http"
	"sale_ranking/pkg/log"
	"strconv"
	"strings"

	// "time"
	"sync"

	"github.com/labstack/echo/v4"
)

func CostSheet_Status(c echo.Context) error {
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
		Total_Revenue_Month string `json:"Total_Revenue_Month" gorm:"column:Total_Revenue_Month"`
	}

	dataResult := struct {
		SOCompelte            interface{}
		Compeltefrompaperless interface{}
		Completefromeform     interface{}
		Onprocess             interface{}
		Cancel                interface{}
		Reject                interface{}
	}{}
	CountReject := 0
	CountCancel := 0
	CountOnprocess := 0
	CountCompletefromeform := 0
	CountCompeltefrompaperless := 0
	CountSOCompelte := 0

	totalReject := float64(0)
	totalCancel := float64(0)
	totalOnprocess := float64(0)
	totalCompletefromeform := float64(0)
	totalCompeltefrompaperless := float64(0)
	totalSOCompelte := float64(0)

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(6)

	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN (
			select * 
			from so_mssql_test
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from paperless%' AND smt.SDPropertyCS28 is not null 
		AND smt.SDPropertyCS28 not like '' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
			DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
			Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
			Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
			Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` ci.StartDate_P1 >= '` + St_date + `' AND ci.StartDate_P1 <= '` + En_date + `' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` ci.EndDate_P1 <= '` + En_date + `' AND ci.EndDate_P1 >= '` + St_date + `' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if StaffID != "" {
				sql = sql + ` si.staff_id like '` + StaffID + `' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` ci.status like '` + Status + `' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Tracking_id != "" {
				sql = sql + ` ci.tracking_id like '` + Tracking_id + `' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_id != "" {
				sql = sql + ` ci.doc_id like '` + Doc_id + `' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if DocumentJson != "" {
				sql = sql + ` ci.documentJson like '` + DocumentJson + `' `
				if Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_number_eform != "" {
				sql = sql + ` ci.doc_number_eform like '` + Doc_number_eform + `' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` ci.Customer_ID like '` + Customer_ID + `' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_thai != "" {
				sql = sql + ` ci.Cusname_thai like '%` + Cusname_thai + `%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_Eng != "" {
				sql = sql + ` ci.Cusname_Eng like '%` + Cusname_Eng + `%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if ID_PreSale != "" {
				sql = sql + ` ci.ID_PreSale like '%` + ID_PreSale + `%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cvm_id != "" {
				sql = sql + ` ci.cvm_id like '` + Cvm_id + `' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Bussiness_type != "" {
				sql = sql + ` ci.Business_type like '%` + Bussiness_type + `%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_Team != "" {
				sql = sql + ` ci.sale_team like '%` + Sale_Team + `%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Job_Status != "" {
				sql = sql + ` ci.Job_Status like '%` + Job_Status + `%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if SO_Type != "" {
				sql = sql + ` ci.SO_Type like '` + SO_Type + `' `
				if Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Name != "" {
				sql = sql + ` ci.Sales_Name like '` + Sales_Name + `' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Surname != "" {
				sql = sql + ` ci.Sales_Surname like '` + Sales_Surname + `' `
				if EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if EmployeeID != "" {
				sql = sql + ` ci.EmployeeID like '` + EmployeeID + `' `
				if Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status_eform != "" {
				sql = sql + ` ci.status_eform like '` + Status_eform + `' `
			}
		}
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.SOCompelte = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "SO Compelte",
		}
		CountSOCompelte = Count_Costsheet
		totalSOCompelte = TRM_All
		wg.Done()

	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql_test
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from paperless%' AND smt.SDPropertyCS28 is null `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
			DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
			Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
			Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
			Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` ci.StartDate_P1 >= '` + St_date + `' AND ci.StartDate_P1 <= '` + En_date + `' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` ci.EndDate_P1 <= '` + En_date + `' AND ci.EndDate_P1 >= '` + St_date + `' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if StaffID != "" {
				sql = sql + ` si.staff_id like '` + StaffID + `' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` ci.status like '` + Status + `' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Tracking_id != "" {
				sql = sql + ` ci.tracking_id like '` + Tracking_id + `' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_id != "" {
				sql = sql + ` ci.doc_id like '` + Doc_id + `' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if DocumentJson != "" {
				sql = sql + ` ci.documentJson like '` + DocumentJson + `' `
				if Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_number_eform != "" {
				sql = sql + ` ci.doc_number_eform like '` + Doc_number_eform + `' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` ci.Customer_ID like '` + Customer_ID + `' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_thai != "" {
				sql = sql + ` ci.Cusname_thai like '%` + Cusname_thai + `%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_Eng != "" {
				sql = sql + ` ci.Cusname_Eng like '%` + Cusname_Eng + `%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if ID_PreSale != "" {
				sql = sql + ` ci.ID_PreSale like '%` + ID_PreSale + `%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cvm_id != "" {
				sql = sql + ` ci.cvm_id like '` + Cvm_id + `' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Bussiness_type != "" {
				sql = sql + ` ci.Business_type like '%` + Bussiness_type + `%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_Team != "" {
				sql = sql + ` ci.sale_team like '%` + Sale_Team + `%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Job_Status != "" {
				sql = sql + ` ci.Job_Status like '%` + Job_Status + `%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if SO_Type != "" {
				sql = sql + ` ci.SO_Type like '` + SO_Type + `' `
				if Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Name != "" {
				sql = sql + ` ci.Sales_Name like '` + Sales_Name + `' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Surname != "" {
				sql = sql + ` ci.Sales_Surname like '` + Sales_Surname + `' `
				if EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if EmployeeID != "" {
				sql = sql + ` ci.EmployeeID like '` + EmployeeID + `' `
				if Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status_eform != "" {
				sql = sql + ` ci.status_eform like '` + Status_eform + `' `
			}
		}
		sql = sql + ` or smt.SDPropertyCS28 like '' AND ci.status_eform like '%Complete from paperless%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
			DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
			Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
			Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
			Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` ci.StartDate_P1 >= '` + St_date + `' AND ci.StartDate_P1 <= '` + En_date + `' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` ci.EndDate_P1 <= '` + En_date + `' AND ci.EndDate_P1 >= '` + St_date + `' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if StaffID != "" {
				sql = sql + ` si.staff_id like '` + StaffID + `' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` ci.status like '` + Status + `' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Tracking_id != "" {
				sql = sql + ` ci.tracking_id like '` + Tracking_id + `' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_id != "" {
				sql = sql + ` ci.doc_id like '` + Doc_id + `' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if DocumentJson != "" {
				sql = sql + ` ci.documentJson like '` + DocumentJson + `' `
				if Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_number_eform != "" {
				sql = sql + ` ci.doc_number_eform like '` + Doc_number_eform + `' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_thai != "" {
				sql = sql + ` ci.Cusname_thai like '%` + Cusname_thai + `%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_Eng != "" {
				sql = sql + ` ci.Cusname_Eng like '%` + Cusname_Eng + `%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if ID_PreSale != "" {
				sql = sql + ` ci.ID_PreSale like '%` + ID_PreSale + `%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cvm_id != "" {
				sql = sql + ` ci.cvm_id like '` + Cvm_id + `' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Bussiness_type != "" {
				sql = sql + ` ci.Bussiness_type like '` + Bussiness_type + `' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_Team != "" {
				sql = sql + ` ci.Sale_Team like '` + Sale_Team + `' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Job_Status != "" {
				sql = sql + ` ci.Job_Status like '` + Job_Status + `' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if SO_Type != "" {
				sql = sql + ` ci.SO_Type like '` + SO_Type + `' `
				if Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Name != "" {
				sql = sql + ` ci.Sales_Name like '` + Sales_Name + `' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Surname != "" {
				sql = sql + ` ci.Sales_Surname like '` + Sales_Surname + `' `
				if EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if EmployeeID != "" {
				sql = sql + ` ci.EmployeeID like '` + EmployeeID + `' `
				if Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status_eform != "" {
				sql = sql + ` ci.status_eform like '` + Status_eform + `' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Compeltefrompaperless = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Compelte from paperless",
		}
		CountCompeltefrompaperless = Count_Costsheet
		totalCompeltefrompaperless = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN (
			select * 
			from so_mssql_test
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Complete from eform%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
			DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
			Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
			Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
			Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` ci.StartDate_P1 >= '` + St_date + `' AND ci.StartDate_P1 <= '` + En_date + `' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` ci.EndDate_P1 <= '` + En_date + `' AND ci.EndDate_P1 >= '` + St_date + `' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if StaffID != "" {
				sql = sql + ` si.staff_id like '` + StaffID + `' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` ci.status like '` + Status + `' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Tracking_id != "" {
				sql = sql + ` ci.tracking_id like '` + Tracking_id + `' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_id != "" {
				sql = sql + ` ci.doc_id like '` + Doc_id + `' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if DocumentJson != "" {
				sql = sql + ` ci.documentJson like '` + DocumentJson + `' `
				if Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_number_eform != "" {
				sql = sql + ` ci.doc_number_eform like '` + Doc_number_eform + `' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` ci.Customer_ID like '` + Customer_ID + `' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_thai != "" {
				sql = sql + ` ci.Cusname_thai like '%` + Cusname_thai + `%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_Eng != "" {
				sql = sql + ` ci.Cusname_Eng like '%` + Cusname_Eng + `%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if ID_PreSale != "" {
				sql = sql + ` ci.ID_PreSale like '%` + ID_PreSale + `%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cvm_id != "" {
				sql = sql + ` ci.cvm_id like '` + Cvm_id + `' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Bussiness_type != "" {
				sql = sql + ` ci.Business_type like '%` + Bussiness_type + `%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_Team != "" {
				sql = sql + ` ci.sale_team like '%` + Sale_Team + `%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Job_Status != "" {
				sql = sql + ` ci.Job_Status like '%` + Job_Status + `%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if SO_Type != "" {
				sql = sql + ` ci.SO_Type like '` + SO_Type + `' `
				if Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Name != "" {
				sql = sql + ` ci.Sales_Name like '` + Sales_Name + `' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Surname != "" {
				sql = sql + ` ci.Sales_Surname like '` + Sales_Surname + `' `
				if EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if EmployeeID != "" {
				sql = sql + ` ci.EmployeeID like '` + EmployeeID + `' `
				if Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status_eform != "" {
				sql = sql + ` ci.status_eform like '` + Status_eform + `' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Completefromeform = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Complete from eform",
		}
		CountCompletefromeform = Count_Costsheet
		totalCompletefromeform = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql_test
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Onprocess%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
			DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
			Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
			Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
			Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` ci.StartDate_P1 >= '` + St_date + `' AND ci.StartDate_P1 <= '` + En_date + `' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` ci.EndDate_P1 <= '` + En_date + `' AND ci.EndDate_P1 >= '` + St_date + `' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if StaffID != "" {
				sql = sql + ` si.staff_id like '` + StaffID + `' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` ci.status like '` + Status + `' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Tracking_id != "" {
				sql = sql + ` ci.tracking_id like '` + Tracking_id + `' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_id != "" {
				sql = sql + ` ci.doc_id like '` + Doc_id + `' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if DocumentJson != "" {
				sql = sql + ` ci.documentJson like '` + DocumentJson + `' `
				if Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_number_eform != "" {
				sql = sql + ` ci.doc_number_eform like '` + Doc_number_eform + `' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` ci.Customer_ID like '` + Customer_ID + `' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_thai != "" {
				sql = sql + ` ci.Cusname_thai like '%` + Cusname_thai + `%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_Eng != "" {
				sql = sql + ` ci.Cusname_Eng like '%` + Cusname_Eng + `%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if ID_PreSale != "" {
				sql = sql + ` ci.ID_PreSale like '%` + ID_PreSale + `%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cvm_id != "" {
				sql = sql + ` ci.cvm_id like '` + Cvm_id + `' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Bussiness_type != "" {
				sql = sql + ` ci.Business_type like '%` + Bussiness_type + `%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_Team != "" {
				sql = sql + ` ci.sale_team like '%` + Sale_Team + `%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Job_Status != "" {
				sql = sql + ` ci.Job_Status like '%` + Job_Status + `%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if SO_Type != "" {
				sql = sql + ` ci.SO_Type like '` + SO_Type + `' `
				if Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Name != "" {
				sql = sql + ` ci.Sales_Name like '` + Sales_Name + `' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Surname != "" {
				sql = sql + ` ci.Sales_Surname like '` + Sales_Surname + `' `
				if EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if EmployeeID != "" {
				sql = sql + ` ci.EmployeeID like '` + EmployeeID + `' `
				if Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status_eform != "" {
				sql = sql + ` ci.status_eform like '` + Status_eform + `' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Onprocess = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Onprocess",
		}
		CountOnprocess = Count_Costsheet
		totalOnprocess = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql_test
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28 
		where ci.status_eform like '%Cancel%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
			DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
			Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
			Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
			Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` ci.StartDate_P1 >= '` + St_date + `' AND ci.StartDate_P1 <= '` + En_date + `' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` ci.EndDate_P1 <= '` + En_date + `' AND ci.EndDate_P1 >= '` + St_date + `' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if StaffID != "" {
				sql = sql + ` si.staff_id like '` + StaffID + `' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` ci.status like '` + Status + `' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Tracking_id != "" {
				sql = sql + ` ci.tracking_id like '` + Tracking_id + `' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_id != "" {
				sql = sql + ` ci.doc_id like '` + Doc_id + `' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if DocumentJson != "" {
				sql = sql + ` ci.documentJson like '` + DocumentJson + `' `
				if Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_number_eform != "" {
				sql = sql + ` ci.doc_number_eform like '` + Doc_number_eform + `' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` ci.Customer_ID like '` + Customer_ID + `' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_thai != "" {
				sql = sql + ` ci.Cusname_thai like '%` + Cusname_thai + `%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_Eng != "" {
				sql = sql + ` ci.Cusname_Eng like '%` + Cusname_Eng + `%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if ID_PreSale != "" {
				sql = sql + ` ci.ID_PreSale like '%` + ID_PreSale + `%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cvm_id != "" {
				sql = sql + ` ci.cvm_id like '` + Cvm_id + `' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Bussiness_type != "" {
				sql = sql + ` ci.Business_type like '%` + Bussiness_type + `%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_Team != "" {
				sql = sql + ` ci.sale_team like '%` + Sale_Team + `%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Job_Status != "" {
				sql = sql + ` ci.Job_Status like '%` + Job_Status + `%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if SO_Type != "" {
				sql = sql + ` ci.SO_Type like '` + SO_Type + `' `
				if Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Name != "" {
				sql = sql + ` ci.Sales_Name like '` + Sales_Name + `' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Surname != "" {
				sql = sql + ` ci.Sales_Surname like '` + Sales_Surname + `' `
				if EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if EmployeeID != "" {
				sql = sql + ` ci.EmployeeID like '` + EmployeeID + `' `
				if Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status_eform != "" {
				sql = sql + ` ci.status_eform like '` + Status_eform + `' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Cancel = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Cancel",
		}
		CountCancel = Count_Costsheet
		totalCancel = TRM_All
		wg.Done()
	}()
	go func() {
		var TRM_All float64 = 0.0
		Count_Costsheet := 0
		var dataRaw []Costsheet_Data
		sql := `select ci.Total_Revenue_Month
		from costsheet_info ci
		left join staff_info si on ci.EmployeeID = si.staff_id
		LEFT JOIN(
			select * 
			from so_mssql_test
			group by sonumber
			) smt on ci.doc_number_eform = smt.SDPropertyCS28
		where ci.status_eform like '%Reject%' `
		if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
			DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
			Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
			Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
			Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` ci.StartDate_P1 >= '` + St_date + `' AND ci.StartDate_P1 <= '` + En_date + `' `
				if En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` ci.EndDate_P1 <= '` + En_date + `' AND ci.EndDate_P1 >= '` + St_date + `' `
				if StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
					DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if StaffID != "" {
				sql = sql + ` si.staff_id like '` + StaffID + `' `
				if Status != "" || Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` ci.status like '` + Status + `' `
				if Tracking_id != "" || Doc_id != "" || DocumentJson != "" ||
					Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
					Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" ||
					Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Tracking_id != "" {
				sql = sql + ` ci.tracking_id like '` + Tracking_id + `' `
				if Doc_id != "" || DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_id != "" {
				sql = sql + ` ci.doc_id like '` + Doc_id + `' `
				if DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if DocumentJson != "" {
				sql = sql + ` ci.documentJson like '` + DocumentJson + `' `
				if Doc_number_eform != "" || Customer_ID != "" ||
					Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" ||
					Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Doc_number_eform != "" {
				sql = sql + ` ci.doc_number_eform like '` + Doc_number_eform + `' `
				if Customer_ID != "" || Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` ci.Customer_ID like '` + Customer_ID + `' `
				if Cusname_thai != "" || Cusname_Eng != "" || ID_PreSale != "" ||
					Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_thai != "" {
				sql = sql + ` ci.Cusname_thai like '%` + Cusname_thai + `%' `
				if Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cusname_Eng != "" {
				sql = sql + ` ci.Cusname_Eng like '%` + Cusname_Eng + `%' `
				if ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" ||
					Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if ID_PreSale != "" {
				sql = sql + ` ci.ID_PreSale like '%` + ID_PreSale + `%' `
				if Cvm_id != "" || Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Cvm_id != "" {
				sql = sql + ` ci.cvm_id like '` + Cvm_id + `' `
				if Bussiness_type != "" || Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Bussiness_type != "" {
				sql = sql + ` ci.Business_type like '%` + Bussiness_type + `%' `
				if Sale_Team != "" || Job_Status != "" || SO_Type != "" ||
					Sales_Name != "" || Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_Team != "" {
				sql = sql + ` ci.sale_team like '%` + Sale_Team + `%' `
				if Job_Status != "" || SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Job_Status != "" {
				sql = sql + ` ci.Job_Status like '%` + Job_Status + `%' `
				if SO_Type != "" || Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if SO_Type != "" {
				sql = sql + ` ci.SO_Type like '` + SO_Type + `' `
				if Sales_Name != "" || Sales_Surname != "" ||
					EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Name != "" {
				sql = sql + ` ci.Sales_Name like '` + Sales_Name + `' `
				if Sales_Surname != "" || EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Sales_Surname != "" {
				sql = sql + ` ci.Sales_Surname like '` + Sales_Surname + `' `
				if EmployeeID != "" || Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if EmployeeID != "" {
				sql = sql + ` ci.EmployeeID like '` + EmployeeID + `' `
				if Status_eform != "" {
					sql = sql + ` AND `
				}
			}
			if Status_eform != "" {
				sql = sql + ` ci.status_eform like '` + Status_eform + `' `
			}
		}

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Float_valA, _ := strconv.ParseFloat(v.Total_Revenue_Month, 64)
			TRM_All += Float_valA
			Count_Costsheet += 1
		}
		dataResult.Reject = map[string]interface{}{
			"Count":  Count_Costsheet,
			"total":  TRM_All,
			"status": "Reject",
		}
		CountReject = Count_Costsheet
		totalReject = TRM_All
		wg.Done()
	}()
	wg.Wait()
	status := map[string]interface{}{
		"total": totalCancel + totalOnprocess + totalCompletefromeform + totalCompeltefrompaperless + totalSOCompelte + totalReject,
		"count": CountCancel + CountOnprocess + CountCompletefromeform + CountCompeltefrompaperless + CountSOCompelte + CountReject,
	}

	Result := map[string]interface{}{
		"detail": dataResult,
		"total":  status,
	}
	return c.JSON(http.StatusOK, Result)
	// return c.JSON(http.StatusOK, dataResult)
}

func Invoice_Status(c echo.Context) error {

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
		PeriodAmount float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
	}

	dataCount := struct {
		Hasinvoice interface{}
		Reduce     interface{}
		Noinvoice  interface{}
	}{}

	Counthasinvoice := 0
	CountReduce := 0
	Countnoinvoice := 0

	totalhasinvoice := float64(0)
	totalReduce := float64(0)
	totalnoinvoice := float64(0)

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		var dataRaw []Invoice_Data
		sql := `select smt.PeriodAmount from so_mssql_test smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is not null AND smt.GetCN not like ''`
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
			So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
			Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
			Active_Inactive != "" || So_refer != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sonumber != "" {
				sql = sql + ` smt.sonumber like '` + Sonumber + `' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" ||
					BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if SDPropertyCS28 != "" {
				sql = sql + ` smt.SDPropertyCS28 like '` + SDPropertyCS28 + `' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_Web_Status != "" {
				sql = sql + ` smt.So_Web_Status like '` + So_Web_Status + `' `
				if BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if BLSCDocNo != "" {
				sql = sql + ` smt.BLSCDocNo like '` + BLSCDocNo + `' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if GetCN != "" {
				sql = sql + ` smt.GetCN like '` + GetCN + `' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if INCSCDocNo != "" {
				sql = sql + ` smt.INCSCDocNo like '` + INCSCDocNo + `' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` smt.Customer_ID like '` + Customer_ID + `' `
				if Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_Name != "" {
				sql = sql + ` smt.Customer_Name like '` + Customer_Name + `' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_code != "" {
				sql = sql + ` smt.sale_code like '` + Sale_code + `' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_name != "" {
				sql = sql + ` smt.sale_name like '` + Sale_name + `' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_team != "" {
				sql = sql + ` smt.sale_team like '` + Sale_team + `' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_lead != "" {
				sql = sql + ` smt.sale_lead like '` + Sale_lead + `' `
				if Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Active_Inactive != "" {
				sql = sql + ` smt.Active_Inactive  like '` + Active_Inactive + `' `
				if So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_refer != "" {
				sql = sql + ` smt.so_refer  like '` + So_refer + `' `
			}
		}
		sql = sql + ` GROUP BY smt.sonumber`

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.Reduce = map[string]interface{}{
			"Count":  Count_Invoice,
			"total":  Total_PA,
			"status": "ลดหนี้",
		}
		CountReduce = Count_Invoice
		totalReduce = Total_PA
		wg.Done()
	}()
	go func() {
		var dataRaw []Invoice_Data
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		sql := `select smt.PeriodAmount from so_mssql_test smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is null AND smt.BLSCDocNo is not null AND smt.BLSCDocNo not like '' `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
			So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
			Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
			Active_Inactive != "" || So_refer != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sonumber != "" {
				sql = sql + ` smt.sonumber like '` + Sonumber + `' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" ||
					BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if SDPropertyCS28 != "" {
				sql = sql + ` smt.SDPropertyCS28 like '` + SDPropertyCS28 + `' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_Web_Status != "" {
				sql = sql + ` smt.So_Web_Status like '` + So_Web_Status + `' `
				if BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if BLSCDocNo != "" {
				sql = sql + ` smt.BLSCDocNo like '` + BLSCDocNo + `' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if GetCN != "" {
				sql = sql + ` smt.GetCN like '` + GetCN + `' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if INCSCDocNo != "" {
				sql = sql + ` smt.INCSCDocNo like '` + INCSCDocNo + `' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` smt.Customer_ID like '` + Customer_ID + `' `
				if Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_Name != "" {
				sql = sql + ` smt.Customer_Name like '` + Customer_Name + `' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_code != "" {
				sql = sql + ` smt.sale_code like '` + Sale_code + `' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_name != "" {
				sql = sql + ` smt.sale_name like '` + Sale_name + `' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_team != "" {
				sql = sql + ` smt.sale_team like '` + Sale_team + `' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_lead != "" {
				sql = sql + ` smt.sale_lead like '` + Sale_lead + `' `
				if Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Active_Inactive != "" {
				sql = sql + ` smt.Active_Inactive  like '` + Active_Inactive + `' `
				if So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_refer != "" {
				sql = sql + ` smt.so_refer  like '` + So_refer + `' `
			}
		}
		sql = sql + ` OR smt.GetCN like '' AND smt.BLSCDocNo is not null AND smt.BLSCDocNo not like ''`
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
			So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
			Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
			Active_Inactive != "" || So_refer != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sonumber != "" {
				sql = sql + ` smt.sonumber like '` + Sonumber + `' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" ||
					BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if SDPropertyCS28 != "" {
				sql = sql + ` smt.SDPropertyCS28 like '` + SDPropertyCS28 + `' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_Web_Status != "" {
				sql = sql + ` smt.So_Web_Status like '` + So_Web_Status + `' `
				if BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if BLSCDocNo != "" {
				sql = sql + ` smt.BLSCDocNo like '` + BLSCDocNo + `' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if GetCN != "" {
				sql = sql + ` smt.GetCN like '` + GetCN + `' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if INCSCDocNo != "" {
				sql = sql + ` smt.INCSCDocNo like '` + INCSCDocNo + `' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` smt.Customer_ID like '` + Customer_ID + `' `
				if Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_Name != "" {
				sql = sql + ` smt.Customer_Name like '` + Customer_Name + `' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_code != "" {
				sql = sql + ` smt.sale_code like '` + Sale_code + `' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_name != "" {
				sql = sql + ` smt.sale_name like '` + Sale_name + `' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_team != "" {
				sql = sql + ` smt.sale_team like '` + Sale_team + `' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_lead != "" {
				sql = sql + ` smt.sale_lead like '` + Sale_lead + `' `
				if Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Active_Inactive != "" {
				sql = sql + ` smt.Active_Inactive  like '` + Active_Inactive + `' `
				if So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_refer != "" {
				sql = sql + ` smt.so_refer  like '` + So_refer + `' `
			}
		}
		sql = sql + ` GROUP BY smt.sonumber`

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.Hasinvoice = map[string]interface{}{
			"Count":  Count_Invoice,
			"total":  Total_PA,
			"status": "ออก invoice เสร้จสิ้น",
		}
		Counthasinvoice = Count_Invoice
		totalhasinvoice = Total_PA
		wg.Done()
	}()
	go func() {
		var Total_PA float64 = 0.0
		Count_Invoice := 0
		var dataRaw []Invoice_Data
		sql := `select smt.PeriodAmount from so_mssql_test smt
		LEFT JOIN staff_info si on smt.sale_code = si.staff_id
		where smt.GetCN is null AND smt.BLSCDocNo is null `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
			So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
			Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
			Active_Inactive != "" || So_refer != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sonumber != "" {
				sql = sql + ` smt.sonumber like '` + Sonumber + `' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" ||
					BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if SDPropertyCS28 != "" {
				sql = sql + ` smt.SDPropertyCS28 like '` + SDPropertyCS28 + `' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_Web_Status != "" {
				sql = sql + ` smt.So_Web_Status like '` + So_Web_Status + `' `
				if BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if BLSCDocNo != "" {
				sql = sql + ` smt.BLSCDocNo like '` + BLSCDocNo + `' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if GetCN != "" {
				sql = sql + ` smt.GetCN like '` + GetCN + `' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if INCSCDocNo != "" {
				sql = sql + ` smt.INCSCDocNo like '` + INCSCDocNo + `' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` smt.Customer_ID like '` + Customer_ID + `' `
				if Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_Name != "" {
				sql = sql + ` smt.Customer_Name like '` + Customer_Name + `' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_code != "" {
				sql = sql + ` smt.sale_code like '` + Sale_code + `' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_name != "" {
				sql = sql + ` smt.sale_name like '` + Sale_name + `' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_team != "" {
				sql = sql + ` smt.sale_team like '` + Sale_team + `' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_lead != "" {
				sql = sql + ` smt.sale_lead like '` + Sale_lead + `' `
				if Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Active_Inactive != "" {
				sql = sql + ` smt.Active_Inactive  like '` + Active_Inactive + `' `
				if So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_refer != "" {
				sql = sql + ` smt.so_refer  like '` + So_refer + `' `
			}
		}
		sql = sql + ` OR smt.GetCN like '' AND smt.BLSCDocNo is null `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
			So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
			Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
			Active_Inactive != "" || So_refer != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sonumber != "" {
				sql = sql + ` smt.sonumber like '` + Sonumber + `' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" ||
					BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if SDPropertyCS28 != "" {
				sql = sql + ` smt.SDPropertyCS28 like '` + SDPropertyCS28 + `' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_Web_Status != "" {
				sql = sql + ` smt.So_Web_Status like '` + So_Web_Status + `' `
				if BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if BLSCDocNo != "" {
				sql = sql + ` smt.BLSCDocNo like '` + BLSCDocNo + `' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if GetCN != "" {
				sql = sql + ` smt.GetCN like '` + GetCN + `' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if INCSCDocNo != "" {
				sql = sql + ` smt.INCSCDocNo like '` + INCSCDocNo + `' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` smt.Customer_ID like '` + Customer_ID + `' `
				if Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_Name != "" {
				sql = sql + ` smt.Customer_Name like '` + Customer_Name + `' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_code != "" {
				sql = sql + ` smt.sale_code like '` + Sale_code + `' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_name != "" {
				sql = sql + ` smt.sale_name like '` + Sale_name + `' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_team != "" {
				sql = sql + ` smt.sale_team like '` + Sale_team + `' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_lead != "" {
				sql = sql + ` smt.sale_lead like '` + Sale_lead + `' `
				if Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Active_Inactive != "" {
				sql = sql + ` smt.Active_Inactive  like '` + Active_Inactive + `' `
				if So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_refer != "" {
				sql = sql + ` smt.so_refer  like '` + So_refer + `' `
			}
		}
		sql = sql + ` OR smt.GetCN like '' AND smt.BLSCDocNo like '' `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
			So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
			Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
			Active_Inactive != "" || So_refer != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sonumber != "" {
				sql = sql + ` smt.sonumber like '` + Sonumber + `' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" ||
					BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if SDPropertyCS28 != "" {
				sql = sql + ` smt.SDPropertyCS28 like '` + SDPropertyCS28 + `' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_Web_Status != "" {
				sql = sql + ` smt.So_Web_Status like '` + So_Web_Status + `' `
				if BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if BLSCDocNo != "" {
				sql = sql + ` smt.BLSCDocNo like '` + BLSCDocNo + `' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if GetCN != "" {
				sql = sql + ` smt.GetCN like '` + GetCN + `' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if INCSCDocNo != "" {
				sql = sql + ` smt.INCSCDocNo like '` + INCSCDocNo + `' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` smt.Customer_ID like '` + Customer_ID + `' `
				if Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_Name != "" {
				sql = sql + ` smt.Customer_Name like '` + Customer_Name + `' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_code != "" {
				sql = sql + ` smt.sale_code like '` + Sale_code + `' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_name != "" {
				sql = sql + ` smt.sale_name like '` + Sale_name + `' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_team != "" {
				sql = sql + ` smt.sale_team like '` + Sale_team + `' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_lead != "" {
				sql = sql + ` smt.sale_lead like '` + Sale_lead + `' `
				if Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Active_Inactive != "" {
				sql = sql + ` smt.Active_Inactive  like '` + Active_Inactive + `' `
				if So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_refer != "" {
				sql = sql + ` smt.so_refer  like '` + So_refer + `' `
			}
		}
		sql = sql + ` OR smt.GetCN is null AND smt.BLSCDocNo like '' `
		if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
			So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
			Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
			Active_Inactive != "" || So_refer != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `'`
				if Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
					So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sonumber != "" {
				sql = sql + ` smt.sonumber like '` + Sonumber + `' `
				if Staff_id != "" || SDPropertyCS28 != "" || So_Web_Status != "" ||
					BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
					Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `' `
				if SDPropertyCS28 != "" || So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if SDPropertyCS28 != "" {
				sql = sql + ` smt.SDPropertyCS28 like '` + SDPropertyCS28 + `' `
				if So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_Web_Status != "" {
				sql = sql + ` smt.So_Web_Status like '` + So_Web_Status + `' `
				if BLSCDocNo != "" || GetCN != "" ||
					INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if BLSCDocNo != "" {
				sql = sql + ` smt.BLSCDocNo like '` + BLSCDocNo + `' `
				if GetCN != "" || INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if GetCN != "" {
				sql = sql + ` smt.GetCN like '` + GetCN + `' `
				if INCSCDocNo != "" || Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if INCSCDocNo != "" {
				sql = sql + ` smt.INCSCDocNo like '` + INCSCDocNo + `' `
				if Customer_ID != "" || Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_ID != "" {
				sql = sql + ` smt.Customer_ID like '` + Customer_ID + `' `
				if Customer_Name != "" || Sale_code != "" ||
					Sale_name != "" || Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Customer_Name != "" {
				sql = sql + ` smt.Customer_Name like '` + Customer_Name + `' `
				if Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_code != "" {
				sql = sql + ` smt.sale_code like '` + Sale_code + `' `
				if Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
					Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_name != "" {
				sql = sql + ` smt.sale_name like '` + Sale_name + `' `
				if Sale_team != "" || Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_team != "" {
				sql = sql + ` smt.sale_team like '` + Sale_team + `' `
				if Sale_lead != "" || Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Sale_lead != "" {
				sql = sql + ` smt.sale_lead like '` + Sale_lead + `' `
				if Active_Inactive != "" || So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if Active_Inactive != "" {
				sql = sql + ` smt.Active_Inactive  like '` + Active_Inactive + `' `
				if So_refer != "" {
					sql = sql + ` AND `
				}
			}
			if So_refer != "" {
				sql = sql + ` smt.so_refer  like '` + So_refer + `' `
			}
		}
		sql = sql + ` GROUP BY smt.sonumber `

		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			Total_PA += v.PeriodAmount
			Count_Invoice += 1
		}
		dataCount.Noinvoice = map[string]interface{}{
			"Count":  Count_Invoice,
			"total":  Total_PA,
			"status": "ยังไม่ออก invoice",
		}
		Countnoinvoice = Count_Invoice
		totalnoinvoice = Total_PA
		wg.Done()
	}()
	wg.Wait()

	status := map[string]interface{}{
		"total": totalhasinvoice + totalReduce + totalnoinvoice,
		"count": Counthasinvoice + CountReduce + Countnoinvoice,
	}
	Result := map[string]interface{}{
		"detail": dataCount,
		"total":  status,
	}
	return c.JSON(http.StatusOK, Result)
	// return c.JSON(http.StatusOK, dataCount)
}

func Billing_Status(c echo.Context) error {
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	Staff_id := strings.TrimSpace(c.QueryParam("staffid"))
	Invoice_no := strings.TrimSpace(c.QueryParam("invoice_no"))
	So_number := strings.TrimSpace(c.QueryParam("so_number"))
	Status := strings.TrimSpace(c.QueryParam("status"))
	Reason := strings.TrimSpace(c.QueryParam("reason"))

	type Billing_Data struct {
		PeriodAmount float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
	}
	dataCount := struct {
		Hasbilling interface{}
		Nobilling  interface{}
	}{}

	Counthasbilling := 0
	Countnobilling := 0

	totalhasbilling := float64(0)
	totalnobilling := float64(0)

	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql := `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางบิลแล้ว%'`
		if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" ||
			So_number != "" || Status != "" || Reason != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" ||
					So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `' `
				if Staff_id != "" || Invoice_no != "" ||
					So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Invoice_no != "" {
				sql = sql + ` bi.invoice_no like '` + Invoice_no + `'`
				if So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if So_number != "" {
				sql = sql + ` bi.so_number like '` + So_number + `'`
				if Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` bi.status like '` + Status + `'`
				if Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Reason != "" {
				sql = sql + ` bi.reason like '` + Reason + `'`
			}
		}
		sql = sql + ` group by smt.sonumber) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.Hasbilling = map[string]interface{}{
			"Count":  CountBilling,
			"total":  TotalPeriodAmount,
			"status": "วางบิลแล้ว",
		}

		Counthasbilling = CountBilling
		totalhasbilling = TotalPeriodAmount
		wg.Done()
	}()
	go func() {
		var TotalPeriodAmount float64 = 0.0
		CountBilling := 0
		var dataRaw []Billing_Data
		sql := `select BL.PeriodAmount
			from
			(select smt.PeriodAmount,smt.INCSCDocNo
			from so_mssql_test smt
			LEFT JOIN staff_info si on smt.sale_code = si.staff_id
			LEFT JOIN billing_info bi on smt.BLSCDocNo = bi.invoice_no
			where bi.status like '%วางไม่ได้%' `
		if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" ||
			So_number != "" || Status != "" || Reason != "" {
			sql = sql + ` AND `
			if St_date != "" {
				sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
				if En_date != "" || Staff_id != "" || Invoice_no != "" ||
					So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if En_date != "" {
				sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `' `
				if Staff_id != "" || Invoice_no != "" ||
					So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Staff_id != "" {
				sql = sql + ` si.staff_id like '` + Staff_id + `'`
				if Invoice_no != "" || So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Invoice_no != "" {
				sql = sql + ` bi.invoice_no like '` + Invoice_no + `'`
				if So_number != "" || Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if So_number != "" {
				sql = sql + ` bi.so_number like '` + So_number + `'`
				if Status != "" || Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Status != "" {
				sql = sql + ` bi.status like '` + Status + `'`
				if Reason != "" {
					sql = sql + ` AND `
				}
			}
			if Reason != "" {
				sql = sql + ` bi.reason like '` + Reason + `'`
			}
		}
		sql = sql + ` group by smt.sonumber) BL`
		if err := dbSale.Ctx().Raw(sql).Scan(&dataRaw).Error; err != nil {
			hasErr += 1
		}
		for _, v := range dataRaw {
			TotalPeriodAmount += v.PeriodAmount
			CountBilling += 1
		}
		dataCount.Nobilling = map[string]interface{}{
			"Count":  CountBilling,
			"total":  TotalPeriodAmount,
			"status": "วางไม่ได้",
		}
		Counthasbilling = CountBilling
		totalhasbilling = TotalPeriodAmount
		wg.Done()
	}()
	wg.Wait()

	status := map[string]interface{}{
		"total": totalhasbilling + totalnobilling,
		"count": Counthasbilling + Countnobilling,
	}
	Result := map[string]interface{}{
		"detail": dataCount,
		"total":  status,
	}
	return c.JSON(http.StatusOK, Result)
	// return c.JSON(http.StatusOK, dataCount)
}

func Reciept_Status(c echo.Context) error {
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	Staff_id := strings.TrimSpace(c.QueryParam("staffid"))
	Invoice_no := strings.TrimSpace(c.QueryParam("invoice_no"))
	So_number := strings.TrimSpace(c.QueryParam("so_number"))
	Status := strings.TrimSpace(c.QueryParam("status"))
	Reason := strings.TrimSpace(c.QueryParam("reason"))

	Reciept_Data := []struct {
		PeriodAmount   float64 `json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Count_Reciept  int     `json:"Count_Reciept" gorm:"column:Count_Reciept"`
		PeriodAmountF  float64 `json:"PeriodAmountF" gorm:"column:PeriodAmountF"`
		Count_RecieptF int     `json:"Count_RecieptF" gorm:"column:Count_RecieptF"`
		// Invoice_status_name	string	`json:"invoice_status_name" gorm:"column:invoice_status_name"`
		// INCSCDocNo			string	`json:"INCSCDocNo" gorm:"column:INCSCDocNo"`
	}{}

	type reciept_Result_Data struct {
		CountReciept      int     `json:"CountReciept"`
		TotalPeriodAmount float64 `json:"TotalPeriodAmount"`
		Reciept_status    string  `json:"Reciept_status"`
	}

	var Reciept_Result_Data []reciept_Result_Data

	sql := `select
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
	where bi.status like '%วางบิลแล้ว%' 
	GROUP BY smt.sonumber`

	if St_date != "" || En_date != "" || Staff_id != "" || Invoice_no != "" ||
		So_number != "" || Status != "" || Reason != "" {
		sql = sql + ` AND `
		if St_date != "" {
			sql = sql + ` smt.PeriodStartDate >= '` + St_date + `' AND smt.PeriodStartDate <= '` + En_date + `' `
			if En_date != "" || Staff_id != "" || Invoice_no != "" ||
				So_number != "" || Status != "" || Reason != "" {
				sql = sql + ` AND `
			}
		}
		if En_date != "" {
			sql = sql + ` smt.PeriodEndDate <= '` + En_date + `' AND smt.PeriodEndDate >= '` + St_date + `' `
			if Staff_id != "" || Invoice_no != "" ||
				So_number != "" || Status != "" || Reason != "" {
				sql = sql + ` AND `
			}
		}
		if Staff_id != "" {
			sql = sql + ` si.staff_id like '` + Staff_id + `'`
			if Invoice_no != "" || So_number != "" || Status != "" || Reason != "" {
				sql = sql + ` AND `
			}
		}
		if Invoice_no != "" {
			sql = sql + ` bi.invoice_no like '` + Invoice_no + `'`
			if So_number != "" || Status != "" || Reason != "" {
				sql = sql + ` AND `
			}
		}
		if So_number != "" {
			sql = sql + ` bi.so_number like '` + So_number + `'`
			if Status != "" || Reason != "" {
				sql = sql + ` AND `
			}
		}
		if Status != "" {
			sql = sql + ` bi.status like '` + Status + `'`
			if Reason != "" {
				sql = sql + ` AND `
			}
		}
		if Reason != "" {
			sql = sql + ` bi.reason like '` + Reason + `'`
		}
	}

	sql = sql + ` )RE`

	if err := dbSale.Ctx().Raw(sql).Scan(&Reciept_Data).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	DataA := reciept_Result_Data{
		CountReciept:      Reciept_Data[0].Count_Reciept,
		TotalPeriodAmount: Reciept_Data[0].PeriodAmount,
		Reciept_status:    "วาง Reciept เสร็จสิ้น",
	}
	Reciept_Result_Data = append(Reciept_Result_Data, DataA)

	DataB := reciept_Result_Data{
		CountReciept:      Reciept_Data[0].Count_RecieptF,
		TotalPeriodAmount: Reciept_Data[0].PeriodAmountF,
		Reciept_status:    "ยังไม่วาง Reciept",
	}
	Reciept_Result_Data = append(Reciept_Result_Data, DataB)

	status := map[string]interface{}{
		"total": Reciept_Data[0].PeriodAmount + Reciept_Data[0].PeriodAmountF,
		"count": Reciept_Data[0].Count_RecieptF + Reciept_Data[0].Count_Reciept,
	}
	Result := map[string]interface{}{
		"detail": Reciept_Result_Data,
		"total":  status,
	}
	return c.JSON(http.StatusOK, Result)

	// return c.JSON(http.StatusOK, Reciept_Result_Data)
}
