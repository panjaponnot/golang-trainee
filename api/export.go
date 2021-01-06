package api

import (
	"fmt"
	"net/http"
	"sale_ranking/model"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

// GetUserEndpoint for Get user
func GetReportExcelSOPendingEndPoint(c echo.Context) error {

	//////////////  getListStaffID  //////////////
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "init db error")
	// }
	staff := []struct {
		StaffId    string `json:"staff_id"`
		Role       string `json:"role"`
		StaffChild string `json:"staff_child"`
	}{}
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}
	oneId := strings.TrimSpace(c.QueryParam("one_id"))

	year := strings.TrimSpace(c.QueryParam("year"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("year")), 10); err == nil {
			yearDefault = time.Unix(util.ConvertTimeStamp(f), 0)
		}
		years, _, _ := yearDefault.Date()
		year = strconv.Itoa(years)
	}

	log.Infoln(pkgName, year)
	log.Infoln(" query staff ")
	if err := dbSale.Ctx().Raw(` SELECT staff_id, role, "" as staff_child from user_info where role = "admin" and one_id = ? 
	union
	SELECT staff_id, "normal" as role, staff_child from staff_info where one_id = ? `, oneId, oneId).Scan(&staff).Error; err != nil {
		log.Errorln(pkgName, err, "Select staff error")
		return echo.ErrInternalServerError
	}
	staffs := []struct {
		StaffId    string `json:"staff_id"`
		StaffChild string `json:"staff_child"`
	}{}
	var listStaffId []string
	if len(staff) != 0 {
		for _, v := range staff {
			log.Infoln(pkgName, v.Role)
			if strings.TrimSpace(v.Role) == "admin" {
				if err := dbSale.Ctx().Raw(`select staff_id from staff_info;`).Scan(&staffs).Error; err != nil {
					log.Errorln(pkgName, err, "Select data error")
				}
				if len(staffs) != 0 {
					for _, id := range staffs {
						listStaffId = append(listStaffId, id.StaffId)
					}
				}
			} else {
				if strings.TrimSpace(v.StaffChild) != "" {
					listStaffId = strings.Split(v.StaffChild, ",")
				}
				listStaffId = append(listStaffId, staff[0].StaffId)
			}
		}
		// if strings.TrimSpace(staff[0].Role) != "admin" {
		// 	listStaffId = append(listStaffId, staff[0].StaffId)
		// }
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

	if err := dbSale.Ctx().Raw(`
	SELECT Active_Inactive,has_refer,tb_ch_so.sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,PeriodAmount, SUM(PeriodAmount) as TotalAmount,staff_id,prefix,fname,lname,nname,position,department,
	(case
		when status is null then 0
		else status end
	) as status,
	  (case
		when tb_expire.remark is null then ''
		else tb_expire.remark end
	) as remark  from (
		SELECT *  from (
		SELECT 	Active_Inactive,has_refer,sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,
				DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,
								PeriodAmount, SUM(PeriodAmount) as TotalAmount,
								staff_id,prefix,fname,lname,nname,position,department,SOType
								FROM ( SELECT * FROM so_mssql WHERE SOType NOT IN ('onetime' , 'project base') ) as s
							left join
							(
								select staff_id, prefix, fname, lname, nname, position, department from staff_info
							
							) tb_sale on s.sale_code = tb_sale.staff_id
							WHERE Active_Inactive = 'Active' and has_refer = 0 and staff_id IN (?) and year(ContractEndDate) = ? 
							group by sonumber
			) as tb_so_number
			left join
			(
			 select 
			 	(case
					when pay_type is null then ''
					else pay_type end
				) as pay_type,
				sonumber as so_check,
				(case
					when so_type is null then ''
					else so_type end
				) as so_type 
			from check_so
			) tb_check on tb_so_number.sonumber = tb_check.so_check
		
		) as tb_ch_so
		left join
		(
		  select id,sonumber,
		  	(case
				when status is null then 0
				else status end
			) as status,
		  	(case
				when remark is null then ''
				else remark end
			) as remark 
			from check_expire
		  ) tb_expire on tb_ch_so.sonumber = tb_expire.sonumber
          group by tb_ch_so.sonumber
		  `, listStaffId, year).Scan(&rawData).Error; err != nil {

		log.Errorln(pkgName, err, "Select data error")
	}

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "pending"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "SO Number")
	f.SetCellValue(mode, "B1", "Customer ID")
	f.SetCellValue(mode, "C1", "Customer Name")
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
		return c.JSON(http.StatusInternalServerError, m.Result{Error: "export error"})
	}
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buff.Bytes())

	// return c.JSON(http.StatusOK, model.Result{Data: rawData, Total: len(rawData)})
}

func GetReportExcelSOEndPoint(c echo.Context) error {
	// if err := initDataStore(); err != nil {
	// 	log.Errorln(pkgName, err, "init db error")
	// }
	if strings.TrimSpace(c.QueryParam("one_id")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Error: "Invalid one id"})
	}
	oneId := strings.TrimSpace(c.QueryParam("one_id"))
	var user []m.UserInfo
	if err := dbSale.Ctx().Raw(` SELECT * FROM user_info WHERE role = 'admin' AND one_id = ? `, oneId).Scan(&user).Error; err != nil {
		log.Errorln(pkgName, err, "User Not Found")
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorln(pkgName, err, "Select user Error")
			return echo.ErrInternalServerError
		}
	}
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
	if len(user) != 0 {

		if err := dbSale.Ctx().Raw(`SELECT * FROM (SELECT check_so.remark_sale as remark,check_so.status_sale,check_so.status_so as status_so,check_so.sonumber,
			Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days , 'so' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_so,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,pricesale
			from so_mssql
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE check_so.status_sale = 0 and check_so.remark_sale <> ''
		group by sonumber
		union
		SELECT check_so.remark_sale as remark,check_so.status_sale,check_so.status_so as status_so,check_so.sonumber,Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days , 'so' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_so,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,PeriodAmount as pricesale
			from so_mssql_navision
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber not like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE check_so.status_sale = 0 and check_so.remark_sale <> ''
		group by sonumber order by status_sale) as data`).Scan(&rawData).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
		}

	} else {
		var listStaffId []string
		staff := struct {
			StaffId    string `json:"staff_id"`
			StaffChild string `json:"staff_child"`
		}{}
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where one_id = ?`, oneId).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
		}

		if strings.TrimSpace(staff.StaffChild) != "" {
			raw := strings.Split(staff.StaffChild, ",")
			for _, id := range raw {
				listStaffId = append(listStaffId, id)
			}
			listStaffId = append(listStaffId, staff.StaffId)
			log.Infoln(pkgName, "has team", staff)
		} else {
			listStaffId = append(listStaffId, staff.StaffId)
			log.Infoln(pkgName, "not found team", staff)
		}

		if err := dbSale.Ctx().Raw(`SELECT * FROM (SELECT check_so.status_so as status_so,check_so.status_sale as status_sale,so_mssql.sonumber,Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days ,check_so.remark_sale as remark,'sale' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_sale,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,pricesale
			from so_mssql
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE sale_code in (?)
			group by sonumber
		union
		SELECT check_so.status_so as status_so,check_so.status_sale as status_sale,so_mssql.sonumber,Customer_ID,Customer_Name,one_id, ContractStartDate,ContractEndDate,
			so_refer,sale_code,sale_lead,PeriodAmount,so_type,pay_type,
			in_factor,sale_factor,
			SUM(PeriodAmount) as TotalAmount_old,
			IFNULL(fname, '') as fname,
			IFNULL(lname, '') as lname,
			IFNULL(nname, '') as nname, department,TotalContractAmount as TotalAmount,SOWebStatus,pricesale,
			datediff(ContractEndDate,ContractStartDate) as days ,check_so.remark_sale as remark,'sale' as role,
			TIMESTAMPDIFF(month,ContractStartDate,DATE_ADD(ContractEndDate, INTERVAL 3 DAY)) as months
		FROM (
			select status_sale,sonumber,Customer_ID,Customer_Name,ContractStartDate,ContractEndDate,so_refer,sale_code,sale_lead,PeriodAmount,
			in_factor,sale_factor,(TotalContractAmount/1.07) as TotalContractAmount,
			SOWebStatus,PeriodAmount as pricesale
			from so_mssql_navision
			where has_refer = 0 and Active_Inactive = 'Active' and sonumber not like '%SO%' and SOType <> 'Onetime' and SOType <> 'Project Base'
		) so_mssql
		left join check_so on check_so.sonumber = so_mssql.sonumber
		LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
		WHERE sale_code in (?)
			group by sonumber order by status_sale ) as data
			`, listStaffId, listStaffId).Scan(&rawData).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
		}

		// return c.JSON(http.StatusOK, listStaffId)

	}

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "check_so"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "SO Number")
	f.SetCellValue(mode, "B1", "Customer ID")
	f.SetCellValue(mode, "C1", "Customer Name")
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
		return c.JSON(http.StatusInternalServerError, m.Result{Error: "export error"})
	}
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buff.Bytes())

}

func GetReportExcelTrackingEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}
	// b, e := strconv.ParseBool(strings.TrimSpace(c.QueryParam("check_amount")))
	// if e != nil {
	// 	return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid check amount"})
	// }

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	status := strings.TrimSpace(c.QueryParam("status"))
	// fmt.Println("====> filter", search)
	ds := time.Now()
	de := time.Now()
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("start_date")), 10); err == nil {
		ds = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("end_date")), 10); err == nil {
		de = time.Unix(util.ConvertTimeStamp(f), 0)
	}
	yearStart, monthStart, dayStart := ds.Date()
	yearEnd, monthEnd, dayEnd := de.Date()
	startRange := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	endRange := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	dateFrom := startRange.Format("2006-01-02")
	dateTo := endRange.Format("2006-01-02")
	m := endRange.Sub(startRange)
	if m < 0 {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid date"})
	}

	//// get staff id ////
	var user []model.UserInfo
	if err := dbSale.Ctx().Raw(`SELECT * FROM user_info WHERE staff_id = ? and role = 'admin';`, saleId).Scan(&user).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
	}
	var listId []string
	if len(user) != 0 {
		var staffAll []model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info ;`).Scan(&staffAll).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
			}
		}
		for _, i := range staffAll {
			listId = append(listId, i.StaffId)
		}
	} else {
		var staffAll model.StaffInfo
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info WHERE staff_id = ?;`, saleId).Scan(&staffAll).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
			}
			return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
		}
		if staffAll.StaffChild != "" {
			data := strings.Split(staffAll.StaffChild, ",")
			listId = data
		}
		listId = append(listId, staffAll.StaffId)
	}

	sql := `Select * From  ( SELECT DISTINCT Customer_ID as Customer_ID, Customer_Name, sum(sonumber) as total_so, sum(csnumber) as total_cs,sum(invnumber) as total_inv, sum(rcnumber) as total_rc, sum(cnnumber) as total_cn,
	sum(so_amount) as so_amount, sum(inv_amount) as inv_amount, sum(cs_amount) as cs_amount, sum(rc_amount) as rc_amount, sum(cn_amount) as cn_amount, sum(amount) as amount, AVG(in_factor) as in_factor,
	sum(in_factor) as sum_if, sum(inv_amount) - sum(rc_amount) as outstainding_amount,sale_code,sale_name,AVG(ex_factor) as ex_factor,sum(ex_factor) as sum_ef, department, nname,
	(CASE
		WHEN sum(inv_amount) = 0 THEN 'ยังไม่ออกใบแจ้งหนี้'
		WHEN sum(inv_amount) = sum(cn_amount) THEN 'ลดหนี้'
		WHEN sum(inv_amount) - sum(cn_amount) <= sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระแล้ว'
		WHEN sum(inv_amount) - sum(cn_amount) > sum(rc_amount) AND sum(rc_amount) <> 0 THEN 'ชำระไม่ครบ'
		ELSE 'ค้างชำระ' END
	) as status,
	sum((CASE
		WHEN inv_amount = rc_amount THEN inv_amount
		ELSE inv_amount - cn_amount END
	)) as inv_amount_cal,
	(sum(amount)/sum(amount_engcost)) as sale_factor,
	sonumber_all
	from (
		SELECT
			count(DISTINCT sonumber) as sonumber,
			count(sonumber) as sonumber_all,
			Customer_ID as Customer_ID,
			Customer_Name as Customer_Name,
			count(DISTINCT(CASE WHEN SDPropertyCS28 !='' THEN SDPropertyCS28 END)) as csnumber,
			count(DISTINCT(CASE WHEN BLSCDocNo !='' THEN BLSCDocNo END)) as invnumber,
			count(DISTINCT(CASE WHEN INCSCDocNo !='' THEN INCSCDocNo END)) as rcnumber,
			count(DISTINCT(CASE WHEN GetCN !='' THEN GetCN END)) as cnnumber,
			sum(so_amount) as so_amount,
			sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) as inv_amount,
			sum(CASE WHEN SDPropertyCS28 !='' THEN so_amount ELSE 0 END) as cs_amount,
			sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as rc_amount,
			sum(CASE WHEN GetCN !='' THEN so_amount ELSE 0 END) as cn_amount,
			sum(PeriodAmount) as amount,
			sum(eng_cost) as amount_engcost,
			sale_factor,
			in_factor,sale_code,sale_name,ex_factor
			FROM (
				SELECT
					SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
					sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,
					(case
						when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
						else 0 end
					) as eng_cost,
					(CASE
						WHEN DATEDIFF(PeriodEndDate, PeriodStartDate)+1 = 0
						THEN 0
						WHEN PeriodStartDate >= ? AND PeriodStartDate <= ? AND PeriodEndDate <= ?
						THEN PeriodAmount
						WHEN PeriodStartDate >= ? AND PeriodStartDate <= ? AND PeriodEndDate > ?
						THEN (DATEDIFF(?, PeriodStartDate)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
						WHEN PeriodStartDate < ? AND PeriodEndDate <= ? AND PeriodEndDate > ?
						THEN (DATEDIFF(PeriodEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
						WHEN PeriodStartDate < ? AND PeriodEndDate = ?
						THEN 1*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
						WHEN PeriodStartDate < ? AND PeriodEndDate > ?
						THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate,PeriodStartDate)+1))
						ELSE 0 END
					) as so_amount
				FROM (
					SELECT * FROM so_mssql
					WHERE Active_Inactive = 'Active'
					and PeriodStartDate <= ? and PeriodEndDate >= ?
					and PeriodStartDate <= PeriodEndDate
					and sale_code in (?)
				) sub_data
			) so_group
			WHERE so_amount <> 0 group by sonumber
		) cust_group
		LEFT JOIN staff_info ON cust_group.sale_code = staff_info.staff_id
		group by Customer_ID ) as a
		where INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code, nname, department), ?) AND INSTR(CONCAT_WS('|', status), ?);`
	var sum []model.SummaryCustomer
	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search, status).Scan(&sum).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
		}
		return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
	}

	// // *******check amount*********** // //
	// if b {
	// 	sqlAmount := `SELECT
	// 	count(DISTINCT sonumber) as total_so,
	// 	count(DISTINCT(CASE WHEN SDPropertyCS28 !='' THEN SDPropertyCS28 END)) as total_cs,
	// 	count(DISTINCT(CASE WHEN BLSCDocNo !='' THEN BLSCDocNo END)) as total_inv,
	// 	count(DISTINCT(CASE WHEN INCSCDocNo !='' THEN INCSCDocNo END)) as total_rc,
	// 	count(DISTINCT(CASE WHEN GetCN !='' THEN GetCN END)) as total_cn,
	// 	sum(so_amount) as so_amount,
	// 	sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) as inv_amount,
	// 	sum(CASE WHEN SDPropertyCS28 !='' THEN so_amount ELSE 0 END) as cs_amount,
	// 	sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as rc_amount,
	// 	sum(CASE WHEN GetCN !='' THEN so_amount ELSE 0 END) as cn_amount,
	// 	sum(PeriodAmount) as amount,
	// 	sum(eng_cost) as amount_engcost,
	// 	(sum(PeriodAmount)/sum(eng_cost)) as sale_factor,
	// 	sum(CASE WHEN BLSCDocNo !='' THEN so_amount ELSE 0 END) - sum(CASE WHEN INCSCDocNo !='' THEN so_amount ELSE 0 END) as outstanding_total,
	// 	count(sonumber) as total_all_so,
	// 	sum(CASE WHEN status_so = 'ยังไม่ออกใบแจ้งหนี้' THEN 1 ELSE 0 END) as total_status_noinv,
	// 	sum(CASE WHEN status_so = 'ออกใบแจ้งหนี้' THEN 1 ELSE 0 END) as total_status_haveinv,
	// 	sum(CASE WHEN status_so = 'ลดหนี้' THEN 1 ELSE 0 END) as total_status_havecn,
	// 	sum(CASE WHEN status_so = 'ยังไม่ออกใบแจ้งหนี้' THEN so_amount ELSE 0 END) as amount_status_noinv,
	// 	sum(CASE WHEN status_so = 'ออกใบแจ้งหนี้' THEN so_amount ELSE 0 END) as amount_status_haveinv,
	// 	sum(CASE WHEN status_so = 'ลดหนี้' THEN so_amount ELSE 0 END) as amount_status_havecn,
	// 	sum(CASE WHEN status_incoome = 'ค้างชำระ' THEN 1 ELSE 0 END) as total_status_norc,
	// 	sum(CASE WHEN status_incoome = 'ชำระแล้ว' THEN 1 ELSE 0 END) as total_status_haverc,
	// 	sum(CASE WHEN status_incoome = 'ค้างชำระ' THEN so_amount ELSE 0 END) as amount_status_norc,
	// 	sum(CASE WHEN status_incoome = 'ชำระแล้ว' THEN so_amount ELSE 0 END) as amount_status_haverc
	// 	,nname, department
	// 	FROM (
	// 		SELECT
	// 			SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
	// 			sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor,
	// 			(case
	// 				when PeriodAmount is not null and sale_factor is not null then PeriodAmount/sale_factor
	// 				else 0 end
	// 			) as eng_cost,
	// 			staff_info.nname, staff_info.department,
	// 			(CASE
	// 				WHEN sonumber <> '' AND BLSCDocNo = '' THEN 'ยังไม่ออกใบแจ้งหนี้'
	// 				WHEN sonumber <> '' AND BLSCDocNo <> '' AND GetCN = '' THEN 'ออกใบแจ้งหนี้'
	// 				WHEN sonumber <> '' AND BLSCDocNo <> '' AND GetCN <> '' AND INCSCDocNo = '' THEN 'ลดหนี้'
	// 				ELSE 'ออกใบแจ้งหนี้' END
	// 			) as status_so,
	// 			(CASE
	// 				WHEN sonumber <> '' AND BLSCDocNo <> '' AND GetCN = '' AND INCSCDocNo = '' THEN 'ค้างชำระ'
	// 				WHEN sonumber <> '' AND BLSCDocNo <> '' AND INCSCDocNo <> '' THEN 'ชำระแล้ว'
	// 				ELSE '' END
	// 			) as status_incoome,
	// 			(CASE
	// 				WHEN DATEDIFF(PeriodEndDate, PeriodStartDate)+1 = 0
	// 				THEN 0
	// 				WHEN PeriodStartDate >= ? AND PeriodStartDate <= ? AND PeriodEndDate <= ?
	// 				THEN PeriodAmount
	// 				WHEN PeriodStartDate >= ? AND PeriodStartDate <= ? AND PeriodEndDate > ?
	// 				THEN (DATEDIFF(?, PeriodStartDate)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
	// 				WHEN PeriodStartDate < ? AND PeriodEndDate <= ? AND PeriodEndDate > ?
	// 				THEN (DATEDIFF(PeriodEndDate, ?)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
	// 				WHEN PeriodStartDate < ? AND PeriodEndDate = ?
	// 				THEN 1*(PeriodAmount/(DATEDIFF(PeriodEndDate, PeriodStartDate)+1))
	// 				WHEN PeriodStartDate < ? AND PeriodEndDate > ?
	// 				THEN (DATEDIFF(?,?)+1)*(PeriodAmount/(DATEDIFF(PeriodEndDate,PeriodStartDate)+1))
	// 				ELSE 0 END
	// 			) as so_amount
	// 		FROM (
	// 			SELECT * FROM so_mssql
	// 			WHERE Active_Inactive = 'Active'
	// 			and PeriodStartDate <= ? and PeriodEndDate >= ?
	// 			and PeriodStartDate <= PeriodEndDate
	// 			and sale_code in (?)
	// 		) sub_data LEFT JOIN staff_info ON sub_data.sale_code = staff_info.staff_id
	// 	) so_group
	// 	WHERE so_amount <> 0 AND INSTR(CONCAT_WS('|', Customer_ID, Customer_Name, sale_code, nname, department), ?) `
	// 	var sumAmount model.SummaryCustomer
	// 	if err := dbSale.Ctx().Raw(sqlAmount, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId, search).Scan(&sumAmount).Error; err != nil {
	// 		if gorm.IsRecordNotFoundError(err) {
	// 			return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
	// 		}
	// 		return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
	// 	}
	// 	sumInFac := 0.0
	// 	sumExFac := 0.0

	// 	if sumAmount.TotalSo != 0 {
	// 		sumInFac = sumIF(sum) / sumAmount.TotalSo
	// 		sumExFac = sumEF(sum) / sumAmount.TotalSo
	// 	}

	// 	dataMap := map[string]interface{}{
	// 		"data":           sumAmount,
	// 		"customer_total": len(sum),
	// 		"in_factor":      sumInFac,
	// 		"ex_factor":      sumExFac,
	// 		"detail":         sum,
	// 	}

	// return c.JSON(http.StatusOK, dataMap)
	// }

	// // *******check amount*********** // //

	// return c.JSON(http.StatusOK, sum)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "tracking"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Customer ID")
	f.SetCellValue(mode, "B1", "Customer Name")
	f.SetCellValue(mode, "C1", "Total So")
	f.SetCellValue(mode, "D1", "Total Cs")
	f.SetCellValue(mode, "E1", "Total Inv")
	f.SetCellValue(mode, "F1", "Total Rc")
	f.SetCellValue(mode, "G1", "Total Cn")
	f.SetCellValue(mode, "H1", "So Amount")
	f.SetCellValue(mode, "I1", "Inv Amount")
	f.SetCellValue(mode, "J1", "Cs Amount")
	f.SetCellValue(mode, "K1", "Rc Amount")
	f.SetCellValue(mode, "L1", "Cn Amount")
	f.SetCellValue(mode, "M1", "Amount")
	f.SetCellValue(mode, "N1", "In Factor")
	f.SetCellValue(mode, "O1", "Sum If")
	f.SetCellValue(mode, "P1", "Outstanding Amount")
	f.SetCellValue(mode, "Q1", "Sale Code")
	f.SetCellValue(mode, "R1", "Sale Name")
	f.SetCellValue(mode, "S1", "Ex Factor")
	f.SetCellValue(mode, "T1", "Sum Ef")
	f.SetCellValue(mode, "U1", "Department")
	f.SetCellValue(mode, "V1", "Inv Amount Cal")
	f.SetCellValue(mode, "W1", "Nname")
	f.SetCellValue(mode, "X1", "Status")
	f.SetCellValue(mode, "Y1", "Sale Factor")
	f.SetCellValue(mode, "Z1", "Sonumber All")

	colCustomerID := "A"
	colCustomerName := "B"
	colTotalSo := "C"
	colTotalCs := "D"
	colSoAmountTotalInv := "E"
	colTotalRc := "F"
	colTotalCn := "G"
	colSoAmount := "H"
	colInvAmount := "I"
	colCsAmount := "J"
	colRcAmount := "K"
	colCnAmount := "L"
	colAmount := "M"
	colInFactor := "N"
	colSumIf := "O"
	colSaleCode := "Q"
	colSaleName := "R"
	colExFactor := "S"
	colSumEf := "T"
	colDepartment := "U"
	colInvAmountCal := "V"
	colNname := "W"
	colStatus := "X"
	colSaleFactor := "Y"
	colSoNumberAll := "Z"

	for k, v := range sum {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colCustomerID, k+2), v.CustomerID)
		f.SetCellValue(mode, fmt.Sprint(colCustomerName, k+2), v.CustomerName)
		f.SetCellValue(mode, fmt.Sprint(colTotalSo, k+2), v.TotalSo)
		f.SetCellValue(mode, fmt.Sprint(colTotalCs, k+2), v.TotalCs)
		f.SetCellValue(mode, fmt.Sprint(colSoAmountTotalInv, k+2), v.SoAmountTotalInv)
		f.SetCellValue(mode, fmt.Sprint(colTotalRc, k+2), v.TotalRc)
		f.SetCellValue(mode, fmt.Sprint(colTotalCn, k+2), v.TotalCn)
		f.SetCellValue(mode, fmt.Sprint(colSoAmount, k+2), v.SoAmount)
		f.SetCellValue(mode, fmt.Sprint(colInvAmount, k+2), v.InvAmount)
		f.SetCellValue(mode, fmt.Sprint(colCsAmount, k+2), v.CsAmount)
		f.SetCellValue(mode, fmt.Sprint(colRcAmount, k+2), v.RcAmount)
		f.SetCellValue(mode, fmt.Sprint(colCnAmount, k+2), v.CnAmount)

		f.SetCellValue(mode, fmt.Sprint(colAmount, k+2), v.Amount)
		f.SetCellValue(mode, fmt.Sprint(colInFactor, k+2), v.InFactor)
		f.SetCellValue(mode, fmt.Sprint(colSumIf, k+2), v.SumIf)
		f.SetCellValue(mode, fmt.Sprint(colSaleCode, k+2), v.SaleCode)
		f.SetCellValue(mode, fmt.Sprint(colSaleName, k+2), v.SaleName)
		f.SetCellValue(mode, fmt.Sprint(colExFactor, k+2), v.ExFactor)
		f.SetCellValue(mode, fmt.Sprint(colSumEf, k+2), v.SumEf)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colInvAmountCal, k+2), v.InvAmountCal)
		f.SetCellValue(mode, fmt.Sprint(colNname, k+2), v.Nname)
		f.SetCellValue(mode, fmt.Sprint(colStatus, k+2), v.Status)
		f.SetCellValue(mode, fmt.Sprint(colSaleFactor, k+2), v.SaleFactor)
		f.SetCellValue(mode, fmt.Sprint(colSoNumberAll, k+2), v.SoNumberAll)
	}

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	buff, err := f.WriteToBuffer()
	if err != nil {
		log.Errorln("XLSX export error ->", err)
		return c.JSON(http.StatusInternalServerError, model.Result{Error: "export error"})
	}
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buff.Bytes())

}
