package api

import (
	"fmt"
	"sale_ranking/pkg/log"
	"strconv"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)
type Resultdata struct{
	DocNumberEform			string	`json:"doc_number_eform"`
	TotalRevenueMonth		string	`json:"Total_Revenue_Month"`
	StatusEform				string	`json:"status_eform"`
}
func CostSheet_Status(c echo.Context) error{
	var Total_TRM float64 = 0.0
	var Count_DNE int = 0
	var Total_TRM_CS float64 = 0.0
	Count_DNE_CS := 0
	var Total_TRM_CP float64 = 0.0
	Count_DNE_CP := 0
	var Total_TRM_CE float64 = 0.0
	Count_DNE_CE := 0
	var Total_TRM_Onprocess float64 = 0.0
	Count_DNE_Onprocess := 0
	var Total_TRM_Reject float64 = 0.0
	Count_DNE_Reject := 0
	var Total_TRM_Cancel float64 = 0.0
	Count_DNE_Cancel := 0

	rawData := []struct {
		Doc_number_eform			string 	`json:"doc_number_eform" gorm:"column:doc_number_eform"`
		Total_Revenue_Month			string	`json:"Total_Revenue_Month" gorm:"column:Total_Revenue_Month"`
		Status_eform				string	`json:"status_eform" gorm:"column:status_eform"`
		SDPropertyCS28				string	`json:"SDPropertyCS28" gorm:"column:SDPropertyCS28"`
	}{}
	
	var resultData []Resultdata
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

	sql := `select ci.doc_number_eform,ci.Total_Revenue_Month,ci.status_eform,smt.SDPropertyCS28
	from costsheet_info ci
	left join staff_info si on ci.sale_code = si.staff_id
	LEFT JOIN so_mssql_test smt on ci.doc_number_eform = smt.SDPropertyCS28 `
	if St_date != "" || En_date != "" || StaffID != "" || Status != "" || Tracking_id != "" || Doc_id != "" ||
	DocumentJson != "" || Doc_number_eform != "" || Customer_ID != "" || Cusname_thai != "" ||
	Cusname_Eng != "" || ID_PreSale != "" || Cvm_id != "" || Bussiness_type != "" || 
	Sale_Team != "" || Job_Status != "" || SO_Type != "" || Sales_Name != "" || 
	Sales_Surname != "" || EmployeeID != "" || Status_eform != ""{
		sql = sql+` where `
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
	// sql := `select DISTINCT status_eform from costsheet_so cs
	// left join staff_info si on cs.sale_code = si.staff_id `

	if err := dbSale.Ctx().Raw(sql).Scan(&rawData).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}
	for _ , j := range rawData {
		if len(j.Doc_number_eform) > 0 && j.Doc_number_eform != ""{
			Count_DNE = Count_DNE+1
			if j.Status_eform == "Complete from paperless" && j.SDPropertyCS28 != "" || 
			j.Status_eform == "Complete from paperless" && len(j.SDPropertyCS28) > 0{
				Count_DNE_CS = Count_DNE_CS+1
			}else if j.Status_eform == "Complete from paperless" && j.SDPropertyCS28 == "" || 
			j.Status_eform == "Complete from paperless" && len(j.SDPropertyCS28) == 0{
				Count_DNE_CP = Count_DNE_CP+1
			}else if j.Status_eform == "Complete from eform"{
				Count_DNE_CE = Count_DNE_CE+1
			}else if j.Status_eform == "Onprocess"{
				Count_DNE_Onprocess = Count_DNE_Onprocess+1
			}else if j.Status_eform == "Cancel"{
				Count_DNE_Cancel = Count_DNE_Cancel+1
			}else if j.Status_eform == "Reject"{
				Count_DNE_Reject = Count_DNE_Reject+1
			}
		}
		if len(j.Total_Revenue_Month) > 0 && j.Total_Revenue_Month != ""{
			Float_valA,_ := strconv.ParseFloat(j.Total_Revenue_Month,64)
			Total_TRM = Total_TRM+Float_valA
			if j.Status_eform == "Complete from paperless" && j.SDPropertyCS28 != "" || 
			j.Status_eform == "Complete from paperless" && len(j.SDPropertyCS28) > 0{
				Total_TRM_CS = Total_TRM_CS+Float_valA
			}else if j.Status_eform == "Complete from paperless" && j.SDPropertyCS28 == "" || 
			j.Status_eform == "Complete from paperless" && len(j.SDPropertyCS28) == 0{
				Total_TRM_CP = Total_TRM_CP+Float_valA
			}else if j.Status_eform == "Complete from eform"{
				Total_TRM_CE = Total_TRM_CE+Float_valA
			}else if j.Status_eform == "Onprocess"{
				Total_TRM_Onprocess = Total_TRM_Onprocess+Float_valA
			}else if j.Status_eform == "Cancel"{
				Total_TRM_Cancel = Total_TRM_Cancel+Float_valA
			}else if j.Status_eform == "Reject"{
				Total_TRM_Reject = Total_TRM_Reject+Float_valA
			}
		}
	}
	var Sum_Total_TRM float64 = Total_TRM_CS+Total_TRM_CP+Total_TRM_CE+Total_TRM_Onprocess+Total_TRM_Cancel+Total_TRM_Reject
	var Sum_Count_DNE int = Count_DNE_CS+Count_DNE_CP+Count_DNE_CE+Count_DNE_Onprocess+Count_DNE_Cancel+Count_DNE_Reject
	fmt.Println(Sum_Total_TRM)
	fmt.Println(Total_TRM)
	fmt.Println(Sum_Count_DNE)
	fmt.Println(Count_DNE)
	if Sum_Count_DNE == Count_DNE{
		fmt.Println("----------------------------------------------")
		data := Resultdata{
			DocNumberEform: strconv.Itoa(Count_DNE_CS),
			TotalRevenueMonth: fmt.Sprintf("%f",Total_TRM_CS),
			StatusEform:"SO Compelte",
		}
		resultData = append(resultData,data)
		dataB := Resultdata{
			DocNumberEform: strconv.Itoa(Count_DNE_CP),
			TotalRevenueMonth: fmt.Sprintf("%f",Total_TRM_CP),
			StatusEform:"Compelte from paperless",
		}
		resultData = append(resultData,dataB)
		dataC := Resultdata{
			DocNumberEform: strconv.Itoa(Count_DNE_CE),
			TotalRevenueMonth: fmt.Sprintf("%f",Total_TRM_CE),
			StatusEform:"Compelte from eform",
		}
		resultData = append(resultData,dataC)
		dataD := Resultdata{
			DocNumberEform: strconv.Itoa(Count_DNE_Onprocess),
			TotalRevenueMonth: fmt.Sprintf("%f",Total_TRM_Onprocess),
			StatusEform:"Onprocess",
		}
		resultData = append(resultData,dataD)
		dataE := Resultdata{
			DocNumberEform: strconv.Itoa(Count_DNE_Cancel),
			TotalRevenueMonth: fmt.Sprintf("%f",Total_TRM_Cancel),
			StatusEform:"Cancel",
		}
		resultData = append(resultData,dataE)
		dataF := Resultdata{
			DocNumberEform: strconv.Itoa(Count_DNE_Reject),
			TotalRevenueMonth: fmt.Sprintf("%f",Total_TRM_Reject),
			StatusEform:"Reject",
		}
		resultData = append(resultData,dataF)
	}
	return c.JSON(http.StatusOK, resultData)
}

func Invoice_Status(c echo.Context) error{
	Total_BLSC := 0
	Total_CN := 0
	Total_W := 0
	var Total_PA float64 = 0.0
	var Total_PA_BLSC float64 = 0.0
	var Total_PA_CN float64 = 0.0
	var Total_PA_W float64 = 0.0

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

	INV_Set_Data := []struct {
		Sonumber		string	`json:"sonumber" gorm:"column:sonumber"`
		BLSCDocNo		string	`json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		GetCN			string	`json:"GetCN" gorm:"column:GetCN"`
		PeriodAmount	float64	`json:"PeriodAmount" gorm:"column:PeriodAmount"`
	}{}

	type inv_Result_Data struct{
		TotalInvoice			string	`json:"TotalInvoice"`
		TotalPeriodAmount		string	`json:"TotalPeriodAmount"`
		PeriodAmountPerDay		string	`json:"PeriodAmountPerDay"`
		InvoiceStatus			string	`json:"InvoiceStatus"`
	}
	
	var INV_Result_Data []inv_Result_Data

	sql:= `select * from so_mssql_test smt
	LEFT JOIN staff_info si on smt.sale_code = si.staff_id`
	if St_date != "" || En_date != "" || Sonumber != "" || Staff_id != "" || SDPropertyCS28 != "" ||
	So_Web_Status != "" || BLSCDocNo != "" || GetCN != "" || INCSCDocNo != "" || Customer_ID != "" ||
	Customer_Name != "" || Sale_code != "" || Sale_name != "" || Sale_team != "" || Sale_lead != "" ||
	Active_Inactive != "" || So_refer != ""{
		sql = sql+` where `
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

	if err := dbSale.Ctx().Raw(sql).Scan(&INV_Set_Data).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	for _ , j := range INV_Set_Data {
		Total_PA = Total_PA+j.PeriodAmount
		if len(j.GetCN) > 0 && j.GetCN != ""{
			Total_PA_CN = Total_PA_CN+j.PeriodAmount
			Total_CN = Total_CN+1
		}else if len(j.BLSCDocNo) > 0 && j.BLSCDocNo != ""{
			Total_PA_BLSC = Total_PA_BLSC+j.PeriodAmount
			Total_BLSC = Total_BLSC+1
		}else{
			Total_PA_W = Total_PA_W+j.PeriodAmount
			Total_W = Total_W+1
		}
	}
	// fmt.Println(fmt.Sprintf("%f",Total_PA))
	time1,_ := time.Parse("2006-01-02",St_date)
	time2,_ := time.Parse("2006-01-02",En_date)
	days := (time2.Sub(time1).Hours() / 24)+1
	Sum_PA :=  Total_PA_CN+Total_PA_BLSC+Total_PA_W
	PA_BLSC_Day := Total_PA_BLSC/days
	PA_CN_Day := Total_PA_CN/days
	PA_W_Day := Total_PA_W/days

	if fmt.Sprintf("%.3f",Total_PA) == fmt.Sprintf("%.3f",Sum_PA){
		DataA := inv_Result_Data{
			TotalInvoice: strconv.Itoa(Total_BLSC),
			TotalPeriodAmount: fmt.Sprintf("%f",Total_PA_BLSC ),
			PeriodAmountPerDay : fmt.Sprintf("%f",PA_BLSC_Day) ,
			InvoiceStatus: "ออก invoice เสร้จสิ้น",
		}
		INV_Result_Data = append(INV_Result_Data,DataA)
		
		DataB := inv_Result_Data{
			TotalInvoice: strconv.Itoa(Total_CN),
			TotalPeriodAmount: fmt.Sprintf("%f",Total_PA_CN ),
			PeriodAmountPerDay : fmt.Sprintf("%f",PA_CN_Day) ,
			InvoiceStatus: "ลดหนี้",
		}
		INV_Result_Data = append(INV_Result_Data,DataB)

		DataC := inv_Result_Data{
			TotalInvoice: strconv.Itoa(Total_W),
			TotalPeriodAmount: fmt.Sprintf("%f",Total_PA_W ),
			PeriodAmountPerDay : fmt.Sprintf("%f",PA_W_Day) ,
			InvoiceStatus: "ไม่มี invoice",
		}
		INV_Result_Data = append(INV_Result_Data,DataC)
	}

	return c.JSON(http.StatusOK,INV_Result_Data)
}

func Billing_Status(c echo.Context) error{
	// CountA := 0
	// CountB := 0
	// CountC := 0
	// CountD := 0
	// CountE := 0
	// CountF := 0
	// CountG := 0
	// CountH := 0
	// CountI := 0
	// var Total  float64 = 0.0
	// var TotalA float64 = 0.0
	// var TotalB float64 = 0.0
	// var TotalC float64 = 0.0
	// var TotalD float64 = 0.0
	// var TotalE float64 = 0.0
	// var TotalF float64 = 0.0
	// var TotalG float64 = 0.0
	// var TotalH float64 = 0.0
	// var TotalI float64 = 0.0

	// St_date := strings.TrimSpace(c.QueryParam("startdate"))
	// En_date := strings.TrimSpace(c.QueryParam("enddate"))
	// Staff_id := strings.TrimSpace(c.QueryParam("staffid"))
	// Seq := strings.TrimSpace(c.QueryParam("seq"))
	// Uid := strings.TrimSpace(c.QueryParam("uid"))
	// So_ref := strings.TrimSpace(c.QueryParam("so_ref"))
	// Invoice_no := strings.TrimSpace(c.QueryParam("invoice_no"))

	so_mssql_Data := []struct {
		PeriodAmount		float64	`json:"PeriodAmount" gorm:"column:PeriodAmount"`
		Sale_code			string	`json:"sale_code" gorm:"column:sale_code"`
		BLSCDocNo			string	`json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
	}{}

	Billing_Data := []struct {
		Invoice_no	string	`json:"Invoice_no" gorm:"column:Invoice_no"`
		Seq			string	`json:"seq" gorm:"column:seq"`
		Uid			string	`json:"uid" gorm:"column:uid"`
		So_ref		string	`json:"so_ref" gorm:"column:so_ref"`
		Updated_at	string	`json:"Updated_at" gorm:"column:Updated_at"`
	}{}

	Billing_Status_Data := []struct {
		Invoice_uid		string	`json:"invoice_uid	" gorm:"column:invoice_uid"`
		Invoice_status	string	`json:"invoice_status" gorm:"column:invoice_status"`
	}{}
	
	// type billing_Result_Data struct{
	// 	CountBilling			string	`json:"CountBilling"`
	// 	TotalPeriodAmount		string	`json:"TotalPeriodAmount"`
	// 	Invoice_status_name		string	`json:"invoice_status_name"`

	// }
	
	// var Billing_Result_Data []billing_Result_Data
	
	sql := `select * from so_mssql_test`

	if err := dbSale.Ctx().Raw(sql).Scan(&so_mssql_Data).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	sql = `select * from invoice`

	if err := dbSale.Ctx().Raw(sql).Scan(&Billing_Data).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	sql = `select * from invoice_status`

	if err := dbSale.Ctx().Raw(sql).Scan(&Billing_Status_Data).Error; err != nil {
		log.Errorln("GettrackingList error :-", err)
	}

	fmt.Println(Billing_Data)
	fmt.Println("------------------")
	fmt.Println(Billing_Status_Data)

	// for _ , j := range Billing_Data {
	// 	Total = Total+j.PeriodAmount
	// 	if j.Invoice_status_name == "วางบิลแล้ว"{
	// 		CountA = CountA+1
	// 		TotalA = TotalA+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == "วางไม่ได้"{
	// 		CountB = CountB+1
	// 		TotalB = TotalB+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == "ลดหนี้"{
	// 		CountC = CountC+1
	// 		TotalC = TotalC+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == "ติดส่งมอบงาน"{
	// 		CountD = CountD+1
	// 		TotalD = TotalD+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == "อื่นๆ"{
	// 		CountE = CountE+1
	// 		TotalE = TotalE+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == "แก้ไขใบแจ้งหนี้"{
	// 		CountF = CountF+1
	// 		TotalF = TotalF+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == "ติด PO"{
	// 		CountG = CountG+1
	// 		TotalG = TotalG+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == "ติดสัญญา"{
	// 		CountH = CountH+1
	// 		TotalH = TotalH+j.PeriodAmount
	// 	}
	// 	if j.Invoice_status_name == ""{
	// 		CountI = CountI+1
	// 		TotalI = TotalI+j.PeriodAmount
	// 	}
	// }
	// Sum_Total := TotalA+TotalB+TotalC+TotalD+TotalE+TotalF+TotalG+TotalH+TotalI
	// if fmt.Sprintf("%.3f",Total) == fmt.Sprintf("%.3f",Sum_Total){
	// 	DataA := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountA),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalA),
	// 		Invoice_status_name: "วางบิลแล้ว",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataA)
		
	// 	DataB := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountB),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalB),
	// 		Invoice_status_name: "วางไม่ได้",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataB)

	// 	DataC := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountC),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalC),
	// 		Invoice_status_name: "ลดหนี้",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataC)

	// 	DataD := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountD),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalD),
	// 		Invoice_status_name: "ติดต่อส่งมอบงาน",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataD)

	// 	DataE := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountE),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalE),
	// 		Invoice_status_name: "อื่นๆ",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataE)

	// 	DataF := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountF),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalF),
	// 		Invoice_status_name: "แก้ไขใบแจ้งหนี้",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataF)


	// 	DataG := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountG),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalG),
	// 		Invoice_status_name: "ติด PO",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataG)

	// 	DataH := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountH),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalH),
	// 		Invoice_status_name: "ติดสัญญา",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataH)

	// 	DataI := billing_Result_Data{
	// 		CountBilling: strconv.Itoa(CountI),
	// 		TotalPeriodAmount: fmt.Sprintf("%f",TotalI),
	// 		Invoice_status_name: "วางบิลไม่ได้",
	// 	}
	// 	Billing_Result_Data = append(Billing_Result_Data,DataI)
	// }
	
	

	return c.JSON(http.StatusOK,so_mssql_Data)
}

func Reciept_Status(c echo.Context) error{
	St_date := strings.TrimSpace(c.QueryParam("startdate"))
	En_date := strings.TrimSpace(c.QueryParam("enddate"))
	Staff_id := strings.TrimSpace(c.QueryParam("staffid"))
	Seq := strings.TrimSpace(c.QueryParam("seq"))
	Uid := strings.TrimSpace(c.QueryParam("uid"))
	So_ref := strings.TrimSpace(c.QueryParam("so_ref"))
	Invoice_no := strings.TrimSpace(c.QueryParam("invoice_no"))
	INCSCDocNo := strings.TrimSpace(c.QueryParam("INCSCDocNo"))

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
	LEFT JOIN invoice inv on smt.BLSCDocNo = inv.invoice_no
	LEFT JOIN invoice_status invs on inv.uid = invs.invoice_uid 
	where invs.invoice_status_name like '%วางบิลแล้ว%' `
	
	if St_date != "" || En_date != "" || Staff_id != "" ||  Seq != "" || Uid != "" || So_ref != "" || 
	Invoice_no != "" || INCSCDocNo != ""{
		sql = sql+` AND `
		if St_date != ""{
			sql = sql+` smt.PeriodStartDate >= '`+St_date+`' AND smt.PeriodStartDate <= '`+En_date+`' `
			if En_date != "" || Staff_id != "" ||  Seq != "" || Uid != "" || So_ref != "" || 
			Invoice_no != "" || INCSCDocNo != ""{
				sql = sql+` AND `
			}
		}
		if En_date != ""{
			sql = sql+` smt.PeriodEndDate <= '`+En_date+`' AND smt.PeriodEndDate >= '`+St_date+`' `
			if Staff_id != "" ||  Seq != "" || Uid != "" || So_ref != "" || 
			Invoice_no != "" || INCSCDocNo != ""{
				sql = sql+` AND `
			}
		}
		if Staff_id != ""{
			sql = sql+` si.staff_id like '`+Staff_id+`'`
			if Seq != "" || Uid != "" || So_ref != "" || Invoice_no != "" || INCSCDocNo != ""{
				sql = sql+` AND `
			}
		}
		if Seq != ""{
			sql = sql+` inv.seq like '`+Seq+`'`
			if Uid != "" || So_ref != "" || Invoice_no != "" || INCSCDocNo != ""{
				sql = sql+` AND `
			}
		}
		if Uid != ""{
			sql = sql+` inv.uid like '`+Uid+`'`
			if So_ref != "" || Invoice_no != "" || INCSCDocNo != ""{
				sql = sql+` AND `
			}
		}
		if So_ref != ""{
			sql = sql+` inv.so_ref like '`+So_ref+`'`
			if Invoice_no != "" || INCSCDocNo != ""{
				sql = sql+` AND `
			}
		}
		if Invoice_no != ""{
			sql = sql+` inv.invoice_no like '`+Invoice_no+`'`
			if INCSCDocNo != ""{
				sql = sql+` AND `
			}
		}
		if INCSCDocNo != ""{
			sql = sql+` inv.INCSCDocNo like '`+INCSCDocNo+`'`
		}
	}

	sql = sql+` GROUP BY inv.invoice_no )RE`
	
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