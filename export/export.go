package export

import (
	"fmt"
	"net/http"
	"sale_ranking/core"
	"sale_ranking/model"
	"sale_ranking/pkg/cache"
	"sale_ranking/pkg/database"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/util"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/labstack/echo/v4"
)

const pkgName = "EXPORT"

var (
	db    database.Database
	redis cache.Redis
)

func initDataStore() error {
	// Database
	db = core.NewDatabase(pkgName)
	if err := db.Connect(); err != nil {
		log.Errorln(pkgName, err, "Connect to database error")
		return err
	}
	// Redis cache
	redis = core.NewRedis()
	if err := redis.Ping(); err != nil {
		log.Errorln(pkgName, err, "Connect to redis error ->")
		return err
	}
	return nil
}

// GetUserEndpoint for Get user
func GetReportExcelSOPendingEndPoint(c echo.Context) error {

	//////////////  getListStaffID  //////////////
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "init db error")
	}
	staff := []struct {
		StaffId    string `json:"staff_id"`
		Role       string `json:"role"`
		StaffChild string `json:"staff_child"`
	}{}
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, model.Result{Error: "Invalid one id"})
	}
	yearDefault := time.Now()
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("year")), 10); err == nil {
		yearDefault = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	year, _, _ := yearDefault.Date()
	log.Infoln(pkgName, year)
	oneId := strings.TrimSpace(c.QueryParam("one_id"))
	log.Infoln(" query staff ")
	if err := db.Ctx().Raw(` SELECT staff_id, role, "" as staff_child from user_info where role = "admin" and one_id = ? 
	union
	SELECT staff_id, "normal" as role, staff_child from staff_info where one_id = ? `, oneId, oneId).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select staff error")
		return echo.ErrInternalServerError
	}
	// var staffs []model.StaffInfo
	staffs := []struct {
		StaffId    string `json:"staff_id"`
		StaffChild string `json:"staff_child"`
	}{}
	var listStaffId []string
	if len(staff) != 0 {
		for _, v := range staff {
			log.Infoln(pkgName, v.Role)
			if strings.TrimSpace(v.Role) == "admin" {
				if err := db.Ctx().Raw(`select staff_id from staff_info;`).Scan(&staffs).Error; err != nil {
					log.Errorln(pkgName, err, "Select data error")
				}
				if len(staffs) != 0 {
					for _, id := range staffs {
						listStaffId = append(listStaffId, id.StaffId)
					}
				}
			} else {
				listStaffId = strings.Split(v.StaffChild, ",")
			}
		}
	}
	//////////////  getListStaffID  //////////////
	rawData := []struct {
		SOnumber          string  `json:"so_number" gorm:"column:sonumber"`
		CustomerId        string  `json:"customer_id" gorm:"column:Customer_ID"`
		CustomerName      string  `json:"customer_name" gorm:"column:Customer_Name"`
		ContractStartDate string  `json:"contract_start_date" gorm:"column:ContractStartDate"`
		ContractEndDate   string  `json:"contract_end_date" gorm:"column:ContractEndDate"`
		SORefer           string  `json:"so_refer" gorm:"column:so_refer"`
		SaleCode          string  `json:"sale_code" gorm:"column:sale_code"`
		SaleLead          string  `json:"sale_lead" gorm:"column:sale_lead"`
		Day               string  `json:"day" gorm:"column:days"`
		SoMonth           string  `json:"so_month" gorm:"column:so_month"`
		SOWebStatus       string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		PriceSale         float64 `json:"price_sale" gorm:"column:pricesale"`
		PeriodAmount      float64 `json:"period_amount" gorm:"column:PeriodAmount"`
		TotalAmount       float64 `json:"total_amount" gorm:"column:TotalAmount"`
		StaffId           string  `json:"staff_id" gorm:"column:staff_id"`
		PayType           string  `json:"pay_type" gorm:"column:pay_type"`
		SoType            string  `json:"so_type" gorm:"column:so_type"`
		Prefix            string  `json:"prefix"`
		Fname             string  `json:"fname"`
		Lname             string  `json:"lname"`
		Nname             string  `json:"nname"`
		Position          string  `json:"position"`
		Department        string  `json:"department"`
		Status            string  `json:"status"`
		Remark            string  `json:"remark"`
	}{}
	if err := db.Ctx().Raw(`SELECT so_mssql.sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,
                    	DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,
                    	PeriodAmount, SUM(PeriodAmount) as TotalAmount,
                    	staff_id,prefix,fname,lname,nname,position,department,
                    	(case
                    		when pay_type is null then ''
                    		else pay_type end
                    	) as pay_type,
                    	(case
                    		when so_type is null then ''
                    		else so_type end
                    	) as so_type,
                    	(case
                    		when status is null then 0
                    		else status end
                    	) as status,
                    	(case
                    		when tb_expire.remark is null then ''
                    		else tb_expire.remark end
                    	) as remark
                    from so_mssql
                    left join
                    (
                        select staff_id, prefix, fname, lname, nname, position, department from staff_info
                    ) tb_sale on so_mssql.sale_code = tb_sale.staff_id
                    left join
                    (
                    	select pay_type,sonumber,so_type from check_so
                    ) tb_check on so_mssql.sonumber = tb_check.sonumber
                    left join
                    (
                    	select id,sonumber,status,remark from check_expire
                    ) tb_expire on so_mssql.sonumber = tb_expire.sonumber
                    WHERE Active_Inactive = 'Active' and has_refer = 0 and year(ContractEndDate) = ? and staff_id in (?)
                    group by sonumber; `, year, listStaffId).Scan(&rawData).Error; err != nil {

		log.Errorln(pkgName, err, "Select data error")
	}

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "pending"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Employee ID")
	f.SetCellValue(mode, "B1", "SO Number")
	f.SetCellValue(mode, "C1", "Customer ID")
	f.SetCellValue(mode, "D1", "Contract Start Date")
	f.SetCellValue(mode, "E1", "Contract End Date")
	f.SetCellValue(mode, "F1", "SO Refer")
	f.SetCellValue(mode, "G1", "Employee ID")
	f.SetCellValue(mode, "H1", "Title")
	f.SetCellValue(mode, "I1", "First Name")
	f.SetCellValue(mode, "J1", "Last Name")
	f.SetCellValue(mode, "K1", "Nick Name")
	f.SetCellValue(mode, "L1", "Lead ID")
	f.SetCellValue(mode, "M1", "Position")
	f.SetCellValue(mode, "N1", "Department")
	f.SetCellValue(mode, "O1", "Price Sale")
	f.SetCellValue(mode, "P1", "Period Amount")
	f.SetCellValue(mode, "Q1", "Total Amount")
	f.SetCellValue(mode, "R1", "Day Remain")
	f.SetCellValue(mode, "S1", "SO Month")
	f.SetCellValue(mode, "T1", "SO Web Status")
	f.SetCellValue(mode, "U1", "Pay Type")
	f.SetCellValue(mode, "V1", "SO Type")
	f.SetCellValue(mode, "W1", "Status")
	f.SetCellValue(mode, "X1", "Remark")

	colSoNumber := "A"
	colCustomerId := "B"
	colCustomerName := "C"
	colContractStartDate := "D"
	colContractEndDate := "E"
	colSoRefer := "F"
	colStaffId := "G"
	colPrefix := "H"
	colFisrtName := "I"
	colLastName := "J"
	colNickName := "K"
	colStaffIdLead := "L"
	colPosition := "M"
	colDepartment := "N"
	colPriceSale := "O"
	colPeriodAmount := "P"
	colTotalAmount := "Q"
	colDays := "R"
	colSoMonth := "S"
	colSoWebStatus := "T"
	colPayType := "U"
	colSoType := "V"
	colStatus := "W"
	colRemark := "X"

	for k, v := range rawData {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colSoNumber, k+2), v.SOnumber)
		f.SetCellValue(mode, fmt.Sprint(colCustomerId, k+2), v.CustomerId)
		f.SetCellValue(mode, fmt.Sprint(colCustomerName, k+2), v.CustomerName)
		f.SetCellValue(mode, fmt.Sprint(colContractStartDate, k+2), v.ContractStartDate)
		f.SetCellValue(mode, fmt.Sprint(colContractEndDate, k+2), v.ContractEndDate)
		f.SetCellValue(mode, fmt.Sprint(colSoRefer, k+2), v.SORefer)
		f.SetCellValue(mode, fmt.Sprint(colStaffId, k+2), v.StaffId)
		f.SetCellValue(mode, fmt.Sprint(colPrefix, k+2), v.Prefix)
		f.SetCellValue(mode, fmt.Sprint(colFisrtName, k+2), v.Fname)
		f.SetCellValue(mode, fmt.Sprint(colLastName, k+2), v.Lname)
		f.SetCellValue(mode, fmt.Sprint(colNickName, k+2), v.Nname)
		f.SetCellValue(mode, fmt.Sprint(colStaffIdLead, k+2), v.SaleLead)

		f.SetCellValue(mode, fmt.Sprint(colPosition, k+2), v.Position)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colPriceSale, k+2), v.PriceSale)
		f.SetCellValue(mode, fmt.Sprint(colPeriodAmount, k+2), v.PeriodAmount)
		f.SetCellValue(mode, fmt.Sprint(colTotalAmount, k+2), v.TotalAmount)
		f.SetCellValue(mode, fmt.Sprint(colDays, k+2), v.Day)
		f.SetCellValue(mode, fmt.Sprint(colSoMonth, k+2), v.SoMonth)
		f.SetCellValue(mode, fmt.Sprint(colSoWebStatus, k+2), v.SOWebStatus)
		f.SetCellValue(mode, fmt.Sprint(colPayType, k+2), v.PayType)
		f.SetCellValue(mode, fmt.Sprint(colSoType, k+2), v.SoType)
		f.SetCellValue(mode, fmt.Sprint(colStatus, k+2), v.Status)
		f.SetCellValue(mode, fmt.Sprint(colRemark, k+2), v.Remark)
	}

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	buff, err := f.WriteToBuffer()
	if err != nil {
		log.Errorln("XLSX export error ->", err)
		return c.JSON(http.StatusInternalServerError, model.Result{Error: "export error"})
	}
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buff.Bytes())

	// return c.JSON(http.StatusOK, model.Result{Data: rawData, Total: len(rawData)})
}

func TestbotEndPoint(c echo.Context) error {
	if err := initDataStore(); err != nil {
		log.Errorln(pkgName, err, "init db error")
	}
	d := struct {
		StaffId string `json:"staff_id"`
		// Role       string `json:"role"`
		// StaffChild string `json:"staff_child"`
	}{}
	if err := db.Ctx().Raw(`SELECT * FROM user_info`).Scan(&d).Error; err != nil {
		log.Errorln(pkgName, err, "Select data error")
	}

	return c.JSON(http.StatusOK, model.Result{Data: d})
}
