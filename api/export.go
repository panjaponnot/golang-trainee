package api

import (
	"fmt"
	"net/http"
	"sale_ranking/model"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"
	"sort"
	"strconv"
	"strings"
	"sync"
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
					break
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
		PayTypeChange     string  `json:"pay_type_change"`
		SoTypeChange      string  `json:"so_type_change"`
		Reason            string  `json:"reason"`
		Remark            string  `json:"remark"`
	}{}

	if err := dbSale.Ctx().Raw(`
	SELECT Active_Inactive,has_refer,tb_ch_so.sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,
	DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,DATEDIFF(ContractEndDate, NOW()) as days,
	month(ContractEndDate) as so_month, SOWebStatus,pricesale,PeriodAmount, SUM(PeriodAmount) as TotalAmount,staff_id,prefix,fname,lname,nname,position,
	department,so_type_change,pay_type_change,so_type,pay_type,
	(case
		when status is null then 0
		else status end
	) as status,
	(case
					when tb_expire.reason is null then ''
					else tb_expire.reason end
	) as reason,
	  (case
		when tb_expire.remark is null then ''
		else tb_expire.remark end
	) as remark  from (
		SELECT *  from (
		SELECT 	Active_Inactive,has_refer,sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,
				DATEDIFF(ContractEndDate, NOW()) as days, month(ContractEndDate) as so_month, SOWebStatus,pricesale,
					PeriodAmount, SUM(PeriodAmount) as TotalAmount,
					staff_id,prefix,fname,lname,nname,position,department,SOType as so_type, '' as pay_type
					FROM ( SELECT * FROM so_mssql WHERE SOType NOT IN ('onetime' , 'project base') ) as s
				left join
				(
					select staff_id, prefix, fname, lname, nname, position, department from staff_info

				) tb_sale on s.sale_code = tb_sale.staff_id
				WHERE Active_Inactive = 'Active' and has_refer = 0 and staff_id IN (?) and year(ContractEndDate) = ?
				group by sonumber
			) as tb_so_number
		) as tb_ch_so
		left join
		(
		  select id,sonumber,
		  	(case
				when status is null then 0
				else status end
			) as status,
			(case
							when reason is null then ''
							else reason end
			) as reason,
		  	(case
				when remark is null then ''
				else remark end
			) as remark,
			pay_type as pay_type_change,
			so_type as so_type_change
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
	f.SetCellValue(mode, "W1", "Pay Type Sale Update")
	f.SetCellValue(mode, "X1", "SO Type Sale Update")
	f.SetCellValue(mode, "Y1", "Status")
	f.SetCellValue(mode, "Z1", "Reason")
	f.SetCellValue(mode, "AA1", "Remark")

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
	colPayTypeChange := "W"
	colSoTypeChange := "X"
	colStatus := "Y"
	colReason := "Z"
	colRemark := "AA"

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
		f.SetCellValue(mode, fmt.Sprint(colPayTypeChange, k+2), v.PayTypeChange)
		f.SetCellValue(mode, fmt.Sprint(colSoTypeChange, k+2), v.SoTypeChange)
		f.SetCellValue(mode, fmt.Sprint(colStatus, k+2), v.Status)
		f.SetCellValue(mode, fmt.Sprint(colReason, k+2), v.Reason)
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
	// search := strings.TrimSpace(c.QueryParam("search"))
	// status := strings.TrimSpace(c.QueryParam("status"))
	// fmt.Println("====> filter", search)
	// ds := time.Now()
	// de := time.Now()
	// if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("start_date")), 10); err == nil {
	// 	ds = time.Unix(util.ConvertTimeStamp(f), 0)
	// }
	// if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("end_date")), 10); err == nil {
	// 	de = time.Unix(util.ConvertTimeStamp(f), 0)
	// }
	// yearStart, monthStart, dayStart := ds.Date()
	// yearEnd, monthEnd, dayEnd := de.Date()
	// startRange := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	// endRange := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	// dateFrom := startRange.Format("2006-01-02")
	// dateTo := endRange.Format("2006-01-02")
	// m := endRange.Sub(startRange)
	// if m < 0 {
	// 	return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid date"})
	// }

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

	sql := `SELECT SDPropertyCS28 as costsheetnumber, sonumber, DATE_FORMAT(ContractStartDate, "%Y-%m-%d") as ContractStartDate,DATE_FORMAT(ContractEndDate, "%Y-%m-%d") as ContractEndDate, pricesale, TotalContractAmount, SOWebStatus,
        Customer_ID, Customer_Name, sale_code, 	sale_name, sale_team, 	PeriodAmount, sale_factor, in_factor, ex_factor, so_refer,
        SOType, detail
        FROM so_mssql
        WHERE Active_Inactive = 'Active'
				and sale_code in (?)
        group by sonumber order by ContractStartDate DESC;`
	var sum []model.SoExport
	if err := dbSale.Ctx().Raw(sql, listId).Scan(&sum).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return c.JSON(http.StatusNotFound, server.Result{Message: "not found staff"})
		}
		return c.JSON(http.StatusInternalServerError, server.Result{Message: "select user error"})
	}

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "tracking"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Costsheet Number")
	f.SetCellValue(mode, "B1", "SONumber")
	f.SetCellValue(mode, "C1", "Contract Start Date")
	f.SetCellValue(mode, "D1", "Contract End Date")
	f.SetCellValue(mode, "E1", "Price Sale")
	f.SetCellValue(mode, "F1", "Total Contract Amount")
	f.SetCellValue(mode, "G1", "SO Status")
	f.SetCellValue(mode, "H1", "Customer ID")
	f.SetCellValue(mode, "I1", "Customer Name")
	f.SetCellValue(mode, "J1", "Sale Code")
	f.SetCellValue(mode, "K1", "Sale Name")
	f.SetCellValue(mode, "L1", "Sale Team")
	f.SetCellValue(mode, "M1", "Period Amount")
	f.SetCellValue(mode, "N1", "Sale Factor")
	f.SetCellValue(mode, "O1", "Internal Factor")
	f.SetCellValue(mode, "P1", "External Factor")
	f.SetCellValue(mode, "Q1", "So Refer")
	f.SetCellValue(mode, "R1", "SO Type")
	f.SetCellValue(mode, "S1", "Detail")

	colCostsheetNumber := "A"
	colSoNumber := "B"
	colContractStartDate := "C"
	colContractEndDate := "D"
	colPriceSale := "E"
	colTotalContractAmount := "F"
	colSOWebStatus := "G"
	colCustomerId := "H"
	colCustomerName := "I"
	colSaleCode := "J"
	colSaleName := "K"
	colSaleTeam := "L"
	colPeriodAmount := "M"
	colSaleFactor := "N"
	colInfactor := "O"
	colExFactor := "P"
	colSoRefer := "Q"
	colSoType := "R"
	colDetail := "S"

	for k, v := range sum {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colCostsheetNumber, k+2), v.CostsheetNumber)
		f.SetCellValue(mode, fmt.Sprint(colSoNumber, k+2), v.SoNumber)
		f.SetCellValue(mode, fmt.Sprint(colContractStartDate, k+2), v.ContractStartDate)
		f.SetCellValue(mode, fmt.Sprint(colContractEndDate, k+2), v.ContractEndDate)
		f.SetCellValue(mode, fmt.Sprint(colPriceSale, k+2), v.PriceSale)
		f.SetCellValue(mode, fmt.Sprint(colTotalContractAmount, k+2), v.TotalContractAmount)
		f.SetCellValue(mode, fmt.Sprint(colSOWebStatus, k+2), v.SOWebStatus)
		f.SetCellValue(mode, fmt.Sprint(colCustomerId, k+2), v.CustomerId)
		f.SetCellValue(mode, fmt.Sprint(colCustomerName, k+2), v.CustomerName)
		f.SetCellValue(mode, fmt.Sprint(colSaleCode, k+2), v.SaleCode)
		f.SetCellValue(mode, fmt.Sprint(colSaleName, k+2), v.SaleName)
		f.SetCellValue(mode, fmt.Sprint(colSaleTeam, k+2), v.SaleTeam)
		f.SetCellValue(mode, fmt.Sprint(colPeriodAmount, k+2), v.PeriodAmount)
		f.SetCellValue(mode, fmt.Sprint(colSaleFactor, k+2), v.SaleFactor)
		f.SetCellValue(mode, fmt.Sprint(colInfactor, k+2), v.Infactor)
		f.SetCellValue(mode, fmt.Sprint(colExFactor, k+2), v.ExFactor)
		f.SetCellValue(mode, fmt.Sprint(colSoRefer, k+2), v.SoRefer)
		f.SetCellValue(mode, fmt.Sprint(colSoType, k+2), v.SoType)
		f.SetCellValue(mode, fmt.Sprint(colDetail, k+2), v.Detail)
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

func GetReportExcelTrackingOldEndPoint(c echo.Context) error {

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
	colOutStandingAmount := "P"
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
		f.SetCellValue(mode, fmt.Sprint(colOutStandingAmount, k+2), v.OutStandingAmount)
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

func GetReportExcelQuotationEndPoint(c echo.Context) error {

	type QuotationJoin struct {
		DocNumberEform  string    `json:"doc_number_eform"`
		Service         string    `json:"service"`
		EmployeeCode    string    `json:"employee_code"`
		SaleName        string    `json:"sale_name" gorm:"column:salename"`
		CompanyName     string    `json:"company_name"`
		Team            string    `json:"team"`
		Total           float64   `json:"total" `
		TotalDiscount   float64   `json:"total_discount"`
		TotalPrice      float64   `json:"total_price"`
		StartDate       time.Time `json:"start_date"`
		EndDate         time.Time `json:"end_date"`
		RefQuotation    string    `json:"ref_quotation"`
		RefSO           string    `json:"ref_so" gorm:"column:refSO"`
		DateTime        string    `json:"datetime" gorm:"column:datetime"`
		ServicePlatform string    `json:"service_platform"`
		Reason          string    `json:"reason"`
		Status          string    `json:"status" gorm:"column:status_sale"`
		Remark          string    `json:"remark" gorm:"column:remark"`
	}

	year := strings.TrimSpace(c.QueryParam("year"))
	if strings.TrimSpace(c.QueryParam("year")) == "" {
		yearDefault := time.Now()
		if f, err := strconv.ParseFloat(strings.TrimSpace(c.QueryParam("year")), 10); err == nil {
			yearDefault = time.Unix(util.ConvertTimeStamp(f), 0)
		}
		years, _, _ := yearDefault.Date()
		year = strconv.Itoa(years)
	}

	if strings.TrimSpace(c.QueryParam("id")) == "" {
		return echo.ErrBadRequest
	}

	id := strings.TrimSpace(c.QueryParam("id"))
	var quarter string
	var month string
	var search string
	// page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	// if strings.TrimSpace(c.QueryParam("page")) == "" {
	// 	page = 1
	// }
	if strings.TrimSpace(c.QueryParam("quarter")) != "" {
		quarter = fmt.Sprintf("AND quarter(start_date) = %s", strings.TrimSpace(c.QueryParam("quarter")))
	}
	if strings.TrimSpace(c.QueryParam("month")) != "" {
		month = fmt.Sprintf("AND MONTH(start_date) = %s", strings.TrimSpace(c.QueryParam("month")))
	}
	if strings.TrimSpace(c.QueryParam("search")) != "" {
		search = fmt.Sprintf("AND INSTR(CONCAT_WS('|', company_name, service, employee_code, salename, team,quatation_th.doc_number_eform), '%s')", strings.TrimSpace(c.QueryParam("search")))
	}

	dataResult := struct {
		Total  interface{} `json:"total"`
		Detail interface{} `json:"detail"`
	}{}
	dataCount := struct {
		Count        int
		Total        int
		Work         int
		NotWork      int
		Win          int
		Lost         int
		Resend       int
		ReasonWin    interface{}
		ReasonResend interface{}
		ReasonLost   interface{}
		CountService interface{}
		CountCompany interface{}
		CountType    interface{}
		CountTeam    interface{}
	}{}

	var user []m.UserInfo
	if err := dbSale.Ctx().Raw(` SELECT * FROM user_info WHERE role = 'admin' AND one_id = ? `, id).Scan(&user).Error; err != nil {
		log.Errorln(pkgName, err, "User Not Found")
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorln(pkgName, err, "Select user Error")
			return echo.ErrInternalServerError
		}
	}
	textStaffId := ""
	var listStaffId []string

	if len(user) == 0 {
		staff := struct {
			StaffId    string `json:"staff_id"`
			StaffChild string `json:"staff_child"`
		}{}
		if err := dbSale.Ctx().Raw(`SELECT * FROM staff_info where one_id = ?`, id).Scan(&staff).Error; err != nil {
			log.Errorln(pkgName, err, "Select data error")
			return c.JSON(http.StatusNotFound, m.Result{Message: "Staff Not Found"})
		}

		if strings.TrimSpace(staff.StaffChild) != "" {
			raw := strings.Split(staff.StaffChild, ",")
			for _, id := range raw {
				listStaffId = append(listStaffId, id)
			}
			listStaffId = append(listStaffId, staff.StaffId)
		} else {
			listStaffId = append(listStaffId, staff.StaffId)
		}
		textStaffId = fmt.Sprintf("AND employee_code IN (%s)", strings.Join(listStaffId, ","))
	}
	log.Infoln(textStaffId)

	hasErr := 0
	// total all
	var dataRaw []QuotationJoin
	var dataRawRes []QuotationJoin
	sql := fmt.Sprintf(`SELECT *,(CASE WHEN total IS NULL THEN total_discount ELSE total end) as total_price FROM quatation_th
		LEFT JOIN (SELECT doc_number_eform,reason,remark,status as status_sale FROM sales_approve WHERE status IN ('Win','Lost','Resend/Revised','Cancel')) as sales_approve
		ON quatation_th.doc_number_eform = sales_approve.doc_number_eform
		WHERE  quatation_th.doc_number_eform IS NOT NULL AND employee_code IS NOT NULL AND (total IS NOT NULL OR total_discount IS NOT NULL)
		AND YEAR(start_date) = ? %s %s %s %s`, textStaffId, quarter, month, search)
	if err := dbQuataion.Ctx().Raw(sql, year).Scan(&dataRaw).Error; err != nil {
		hasErr += 1
	}
	// if len(dataRaw) > (page * 20) {
	// 	start := (page - 1) * 20
	// 	end := (page * 20)
	// 	dataResult.Detail = map[string]interface{}{
	// 		"data":  dataRaw[start:end],
	// 		"count": len(dataRaw[start:end]),
	// 	}
	// 	dataRawRes = dataRaw[start:end]
	// } else {
	// 	start := (page * 20) - (20)
	// 	dataResult.Detail = map[string]interface{}{
	// 		"data":  dataRaw[start:],
	// 		"count": len(dataRaw[start:]),
	// 	}
	// 	dataRawRes = dataRaw[start:]
	// }

	dataResult.Detail = map[string]interface{}{
		"data":  dataRaw,
		"count": len(dataRaw),
	}
	dataRawRes = dataRaw
	dataCount.Total = len(dataRaw)

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}

	// return c.JSON(http.StatusOK, dataResult)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "quotation"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "DocNumber Eform")
	f.SetCellValue(mode, "B1", "Service")
	f.SetCellValue(mode, "C1", "Employee Code")
	f.SetCellValue(mode, "D1", "Sale Name")
	f.SetCellValue(mode, "E1", "Company Name")
	f.SetCellValue(mode, "F1", "Team")
	f.SetCellValue(mode, "G1", "Total")
	f.SetCellValue(mode, "H1", "Total Discount")
	f.SetCellValue(mode, "I1", "Total Price")
	f.SetCellValue(mode, "J1", "Start Date")
	f.SetCellValue(mode, "K1", "End Date")
	f.SetCellValue(mode, "L1", "Ref Quotation")
	f.SetCellValue(mode, "M1", "Ref SO")
	f.SetCellValue(mode, "N1", "Date Time")
	f.SetCellValue(mode, "O1", "Service Platform")
	f.SetCellValue(mode, "P1", "Reason")
	f.SetCellValue(mode, "Q1", "Status")
	f.SetCellValue(mode, "R1", "Remark")

	colDocNumberEform := "A"
	colService := "B"
	colEmployeeCode := "C"
	colSaleName := "D"
	colCompanyName := "E"
	colTeam := "F"
	colTotal := "G"
	colTotalDiscount := "H"
	colTotalPrice := "I"
	colStartDate := "J"
	colEndDate := "K"
	colRefQuotation := "L"
	colRefSO := "M"
	colDateTime := "N"
	colServicePlatform := "O"
	colReason := "P"
	colStatus := "Q"
	colRemark := "R"

	for k, v := range dataRawRes {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colDocNumberEform, k+2), v.DocNumberEform)
		f.SetCellValue(mode, fmt.Sprint(colService, k+2), v.Service)
		f.SetCellValue(mode, fmt.Sprint(colEmployeeCode, k+2), v.EmployeeCode)
		f.SetCellValue(mode, fmt.Sprint(colSaleName, k+2), v.SaleName)
		f.SetCellValue(mode, fmt.Sprint(colCompanyName, k+2), v.CompanyName)
		f.SetCellValue(mode, fmt.Sprint(colTeam, k+2), v.Team)
		f.SetCellValue(mode, fmt.Sprint(colTotal, k+2), v.Total)
		f.SetCellValue(mode, fmt.Sprint(colTotalDiscount, k+2), v.TotalDiscount)
		f.SetCellValue(mode, fmt.Sprint(colTotalPrice, k+2), v.TotalPrice)
		f.SetCellValue(mode, fmt.Sprint(colStartDate, k+2), v.StartDate)
		f.SetCellValue(mode, fmt.Sprint(colEndDate, k+2), v.EndDate)
		f.SetCellValue(mode, fmt.Sprint(colRefQuotation, k+2), v.RefQuotation)

		f.SetCellValue(mode, fmt.Sprint(colRefSO, k+2), v.RefSO)
		f.SetCellValue(mode, fmt.Sprint(colDateTime, k+2), v.DateTime)
		f.SetCellValue(mode, fmt.Sprint(colServicePlatform, k+2), v.ServicePlatform)
		f.SetCellValue(mode, fmt.Sprint(colReason, k+2), v.Reason)
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

}

func GettReportExcelRankBaseSaleEndPoint(c echo.Context) error {

	filterDepart := strings.Split(util.GetEnv("CONDITION_BASE_SALE", ""), ",")
	var dFilter []string
	for _, v := range filterDepart {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}
	listStaffId, err := CheckPermissionBaseSale(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		log.Errorln(pkgName, err, "func check permission error :-")
		return c.JSON(http.StatusInternalServerError, m.Result{Error: "check permission error"})
	}
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}
	// page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	// if strings.TrimSpace(c.QueryParam("page")) == "" {
	// 	page = 1
	// }
	q := strings.TrimSpace(c.QueryParam("quarter"))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
		WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
		WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
		WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
		WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
		WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
		WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
		ELSE -5 END
	) as score_target,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 999
		ELSE (sum(inv_amount)/goal_total)*100 END
	) as percent_target,
	(CASE
		WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
		WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
		WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
		WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
		WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
		WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
		WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
		ELSE 0 END
	) as score_sf,
	(case
		when sum(revenue)/sum(engcost) is null then 0
		else sum(revenue)/sum(engcost) end
	) as sale_factor,
	(CASE
		WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
		WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
		WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
		WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
		WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
		WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
		WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
		ELSE 0 END
	) as score_if,
	(case
		when sum(sum_if)/sum(total_so) is null then 0
		else sum(sum_if)/sum(total_so) end
	) as in_factor,
	sum(revenue) as total_revenue,
	all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
	from (
		select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
		(case
			when inv_amount is null then 0
			ELSE inv_amount END) as inv_amount
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,
			'normal' as typestaff,
			sum((CASE
				WHEN TotalContractAmount is null THEN 0
				ELSE TotalContractAmount END
			)) as revenue,
			sum((CASE
				WHEN eng_cost is null THEN 0
				ELSE eng_cost END
			)) as engcost,
			sum((CASE
				WHEN in_factor is null THEN 0
				ELSE in_factor END
			)) as sum_if,
			sum((CASE
				WHEN sonumber is null THEN 0
				ELSE 1 END
			)) as total_so,
			one_id,
			quarter,year,job_months,staff_child
			from (
					select staff_id,fname,lname,nname,department,position,
					(CASE
						WHEN goal_total is null THEN 0
						ELSE goal_total END
					) as goal_total,
					staff_info.one_id,
					(CASE
						WHEN quarter is null THEN ?
						ELSE quarter END
					) as quarter,
					(CASE
						WHEN year is null THEN year(now())
						ELSE year END
					) as year,
					12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
					staff_child
					from staff_info
					left join
					(
						select * from goal_quarter where year = year(now()) and quarter =  ?
					) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
					left join staff_start on staff_info.one_id = staff_start.one_id
					WHERE staff_id in (
						select staff_id from staff_info WHERE staff_child = ''
					)
					group by staff_id
			) staff_detail
			LEFT JOIN (
				select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
				group by sonumber
			) total_so on total_so.sale_code = staff_detail.staff_id
			group by staff_id
		) tb_main
		LEFT join (
			select sum(PeriodAmount) as inv_amount, sale_code from (
				select PeriodAmount,sale_code
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now())   and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv group by sale_code
		) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
		where staff_id is not null and staff_id <> ''
	) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
	WHERE staff_id in (?)
	group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child = ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`
	sqlFilter := `select * from staff_info where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?) `

	var staffInfo []m.StaffInfo
	mapCnStaff := map[string][]string{}
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		var so []m.SOMssql
		if err := dbSale.Ctx().Model(&m.SOMssql{}).Where(`sale_code IN (?) AND INCSCDocNo = '' AND quarter(ContractStartDate) = ? AND year(ContractStartDate) = year(now()) AND DATEDIFF(NOW(),PeriodEndDate) > 60`, listStaffId, quarterNum-1).Group("Customer_ID").Find(&so).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		for _, s := range so {
			mapCnStaff[s.SaleCode] = append(mapCnStaff[s.SaleCode], s.INCSCDocNo)
		}
		wg.Done()
	}()
	wg.Wait()
	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for index, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
				if r.InvAmount > r.InvAmountOld {
					// var so []m.SOMssql
					// if err := dbSale.Ctx().Model(&m.SOMssql{}).Where(`sale_code = ? AND INCSCDocNo <> ''`, r.StaffId).Group("INCSCDocNo").Find(&so).Error; err != nil {
					// }
					// 1 => 1000
					// 2 => 3000
					// 3 => 5000
					// 4 => 7000
					i := len(mapCnStaff[r.StaffId])
					x := 0
					if len(mapCnStaff[r.StaffId]) > 0 {
						x = (i * 1000) + ((i - 1) * 1000)
					}

					baseCal := r.InvAmountOld * 0.003
					growthCal := (r.InvAmount - r.InvAmountOld) * 0.03
					saleFactor := (baseCal + growthCal) * (r.SaleFactor * r.SaleFactor)
					r.Commission = (saleFactor * (r.InFactor / 0.7)) - float64(x)

					log.Infoln("SO", "=====>", len(mapCnStaff[r.StaffId]), " === comission not cal== ", int(r.Commission), "  == aging =", x)
				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth
		report[index] = r
	}
	if len(report) > 1 {
		sort.SliceStable(report, func(i, j int) bool { return report[i].ScoreAll > report[j].ScoreAll })
	}
	for i, r := range report {
		report[i].Order = i + 1
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, report[i])
				}
			}
		}
	}
	var dataResultEx []m.OrgChart
	var result m.Result
	// if len(dataResult) > (page * 10) {
	// 	start := (page - 1) * 10
	// 	end := (page * 10)
	// 	result.Data = dataResult[start:end]
	// 	result.Count = len(dataResult[start:end])
	// 	dataResultEx = dataResult[start:end]
	// } else {
	// 	start := (page * 10) - (10)
	// 	result.Data = dataResult[start:]
	// 	result.Count = len(dataResult[start:])
	// 	dataResultEx = dataResult[start:]
	// }

	result.Data = dataResult
	result.Count = len(dataResult)
	dataResultEx = dataResult

	result.Total = len(dataResult)
	// return c.JSON(http.StatusOK, result)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "rankingbase"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Order")
	f.SetCellValue(mode, "B1", "StaffId")
	f.SetCellValue(mode, "C1", "First Name")
	f.SetCellValue(mode, "D1", "Last Name")
	f.SetCellValue(mode, "E1", "Nick Name")
	f.SetCellValue(mode, "F1", "Position")
	f.SetCellValue(mode, "G1", "Department")
	f.SetCellValue(mode, "H1", "Staff Child")
	f.SetCellValue(mode, "I1", "Inv Amount")
	f.SetCellValue(mode, "J1", "Inv Amount Old")
	f.SetCellValue(mode, "K1", "Goal Total")
	f.SetCellValue(mode, "L1", "Score Target")
	f.SetCellValue(mode, "M1", "Score Sf")
	f.SetCellValue(mode, "N1", "Sale Factor")
	f.SetCellValue(mode, "O1", "Total So")
	f.SetCellValue(mode, "P1", "If Factor")
	f.SetCellValue(mode, "Q1", "Eng Cost")
	f.SetCellValue(mode, "R1", "Revenue")
	f.SetCellValue(mode, "S1", "Score If")
	f.SetCellValue(mode, "T1", "In Factor")
	f.SetCellValue(mode, "U1", "One Id")
	f.SetCellValue(mode, "V1", "Image")
	f.SetCellValue(mode, "W1", "File Name")
	f.SetCellValue(mode, "X1", "Growth Rate")
	f.SetCellValue(mode, "Y1", "Score Growth")
	f.SetCellValue(mode, "Z1", "Score All")
	f.SetCellValue(mode, "AA1", "Quarter")
	f.SetCellValue(mode, "AB1", "Year")
	f.SetCellValue(mode, "AC1", "Job Months")
	f.SetCellValue(mode, "AD1", "Commission")

	colOrder := "A"
	colStaffId := "B"
	colFname := "C"
	colLname := "D"
	colNname := "E"
	colPosition := "F"
	colDepartment := "G"
	colStaffChild := "H"
	colInvAmount := "I"
	colInvAmountOld := "J"
	colGoalTotal := "K"
	colScoreTarget := "L"
	colScoreSf := "M"
	colSaleFactor := "N"
	colTotalSo := "O"
	colIfFactor := "P"
	colEngCost := "Q"
	colRevenue := "R"
	colScoreIf := "S"
	colInFactor := "T"
	colOneId := "U"
	colImage := "V"
	colFileName := "W"
	colGrowthRate := "X"
	colScoreGrowth := "Y"
	colScoreAll := "Z"
	colQuarter := "AA"
	colYear := "AB"
	colJobMonths := "AC"
	colCommission := "AD"

	for k, v := range dataResultEx {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colOrder, k+2), v.Order)
		f.SetCellValue(mode, fmt.Sprint(colStaffId, k+2), v.StaffId)
		f.SetCellValue(mode, fmt.Sprint(colFname, k+2), v.Fname)
		f.SetCellValue(mode, fmt.Sprint(colLname, k+2), v.Lname)
		f.SetCellValue(mode, fmt.Sprint(colNname, k+2), v.Nname)
		f.SetCellValue(mode, fmt.Sprint(colPosition, k+2), v.Position)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colStaffChild, k+2), v.StaffChild)
		f.SetCellValue(mode, fmt.Sprint(colInvAmount, k+2), v.InvAmount)
		f.SetCellValue(mode, fmt.Sprint(colInvAmountOld, k+2), v.InvAmountOld)
		f.SetCellValue(mode, fmt.Sprint(colGoalTotal, k+2), v.GoalTotal)
		f.SetCellValue(mode, fmt.Sprint(colScoreTarget, k+2), v.ScoreTarget)
		f.SetCellValue(mode, fmt.Sprint(colScoreSf, k+2), v.ScoreSf)

		f.SetCellValue(mode, fmt.Sprint(colSaleFactor, k+2), v.SaleFactor)
		f.SetCellValue(mode, fmt.Sprint(colTotalSo, k+2), v.TotalSo)
		f.SetCellValue(mode, fmt.Sprint(colIfFactor, k+2), v.IfFactor)
		f.SetCellValue(mode, fmt.Sprint(colEngCost, k+2), v.EngCost)
		f.SetCellValue(mode, fmt.Sprint(colRevenue, k+2), v.Revenue)
		f.SetCellValue(mode, fmt.Sprint(colScoreIf, k+2), v.ScoreIf)

		f.SetCellValue(mode, fmt.Sprint(colInFactor, k+2), v.InFactor)
		f.SetCellValue(mode, fmt.Sprint(colOneId, k+2), v.OneId)
		f.SetCellValue(mode, fmt.Sprint(colImage, k+2), v.Image)
		f.SetCellValue(mode, fmt.Sprint(colFileName, k+2), v.FileName)
		f.SetCellValue(mode, fmt.Sprint(colGrowthRate, k+2), v.GrowthRate)
		f.SetCellValue(mode, fmt.Sprint(colScoreGrowth, k+2), v.ScoreGrowth)
		f.SetCellValue(mode, fmt.Sprint(colScoreAll, k+2), v.ScoreAll)
		f.SetCellValue(mode, fmt.Sprint(colQuarter, k+2), v.Quarter)
		f.SetCellValue(mode, fmt.Sprint(colYear, k+2), v.Year)
		f.SetCellValue(mode, fmt.Sprint(colJobMonths, k+2), v.JobMonths)
		f.SetCellValue(mode, fmt.Sprint(colCommission, k+2), v.Commission)

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

func GettReportExcelRankKeyAccEndPoint(c echo.Context) error {

	conKey := strings.Split(util.GetEnv("CONDITION_GOV_KEY_SALE", ""), ",")
	var dFilter []string
	for _, v := range conKey {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}

	listStaffId, err := CheckPermissionKeyAccount(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		return echo.ErrInternalServerError
	}
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}

	// page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	// if strings.TrimSpace(c.QueryParam("page")) == "" {
	// 	page = 1
	// }
	q := strings.TrimSpace(c.QueryParam("quarter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
		WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
		WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
		WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
		WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
		WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
		WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
		ELSE -5 END
	) as score_target,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 999
		ELSE (sum(inv_amount)/goal_total)*100 END
	) as percent_target,
	(CASE
		WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
		WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
		WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
		WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
		WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
		WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
		WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
		ELSE 0 END
	) as score_sf,
	(case
		when sum(revenue)/sum(engcost) is null then 0
		else sum(revenue)/sum(engcost) end
	) as sale_factor,
	(CASE
		WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
		WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
		WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
		WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
		WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
		WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
		WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
		ELSE 0 END
	) as score_if,
	(case
		when sum(sum_if)/sum(total_so) is null then 0
		else sum(sum_if)/sum(total_so) end
	) as in_factor,
	sum(revenue) as total_revenue,
	all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
	from (
		select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
		(case
			when inv_amount is null then 0
			ELSE inv_amount END) as inv_amount
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,
			'normal' as typestaff,
			sum((CASE
				WHEN TotalContractAmount is null THEN 0
				ELSE TotalContractAmount END
			)) as revenue,
			sum((CASE
				WHEN eng_cost is null THEN 0
				ELSE eng_cost END
			)) as engcost,
			sum((CASE
				WHEN in_factor is null THEN 0
				ELSE in_factor END
			)) as sum_if,
			sum((CASE
				WHEN sonumber is null THEN 0
				ELSE 1 END
			)) as total_so,
			one_id,
			quarter,year,job_months,staff_child
			from (
					select staff_id,fname,lname,nname,department,position,
					(CASE
						WHEN goal_total is null THEN 0
						ELSE goal_total END
					) as goal_total,
					staff_info.one_id,
					(CASE
						WHEN quarter is null THEN ?
						ELSE quarter END
					) as quarter,
					(CASE
						WHEN year is null THEN year(now())
						ELSE year END
					) as year,
					12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
					staff_child
					from staff_info
					left join
					(
						select * from goal_quarter where year = year(now()) and quarter =  ?
					) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
					left join staff_start on staff_info.one_id = staff_start.one_id
					WHERE staff_id in (
						select staff_id from staff_info WHERE staff_child = ''
					)
					group by staff_id
			) staff_detail
			LEFT JOIN (
				select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
				group by sonumber
			) total_so on total_so.sale_code = staff_detail.staff_id
			group by staff_id
		) tb_main
		LEFT join (
			select sum(PeriodAmount) as inv_amount, sale_code from (
				select PeriodAmount,sale_code
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now())   and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv group by sale_code
		) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
		where staff_id is not null and staff_id <> ''
	) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
	WHERE staff_id in (?)
	group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child = ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`

	sqlFilter := `select * from staff_info where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?) `

	var staffInfo []m.StaffInfo
	mapCnStaff := map[string][]string{}
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		wg.Done()
	}()
	go func() {
		var so []m.SOMssql
		if err := dbSale.Ctx().Model(&m.SOMssql{}).Where(`sale_code IN (?) AND INCSCDocNo = '' AND quarter(ContractStartDate) = ? AND year(ContractStartDate) = year(now()) AND DATEDIFF(NOW(),PeriodEndDate) > 60`, listStaffId, quarterNum-1).Group("Customer_ID").Find(&so).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		for _, s := range so {
			mapCnStaff[s.SaleCode] = append(mapCnStaff[s.SaleCode], s.INCSCDocNo)
		}
		wg.Done()
	}()
	wg.Wait()
	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for index, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
				if r.InvAmount > r.InvAmountOld {
					i := len(mapCnStaff[r.StaffId])
					x := 0
					if len(mapCnStaff[r.StaffId]) > 0 {
						x = (i * 1000) + ((i - 1) * 1000)
					}
					// wait cal aging & blacklist
					baseCal := r.InvAmountOld * 0.003
					growthCal := (r.InvAmount - r.InvAmountOld) * 0.03
					saleFactor := (baseCal + growthCal) * (r.SaleFactor * r.SaleFactor)
					r.Commission = saleFactor*(r.InFactor/0.7) - float64(x)
				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth
		report[index] = r
	}
	if len(report) > 1 {
		sort.SliceStable(report, func(i, j int) bool { return report[i].ScoreAll > report[j].ScoreAll })
	}
	for i, r := range report {
		report[i].Order = i + 1
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, report[i])
				}
			}
		}
	}
	var dataResultEx []m.OrgChart
	var result m.Result
	// if len(dataResult) > (page * 10) {
	// 	start := (page - 1) * 10
	// 	end := (page * 10)
	// 	result.Data = dataResult[start:end]
	// 	result.Count = len(dataResult[start:end])
	// 	dataResultEx = dataResult[start:end]
	// } else {
	// 	start := (page * 10) - (10)
	// 	result.Data = dataResult[start:]
	// 	result.Count = len(dataResult[start:])
	// 	dataResultEx = dataResult[start:]
	// }
	result.Data = dataResult
	result.Count = len(dataResult)
	dataResultEx = dataResult

	result.Total = len(dataResult)
	// return c.JSON(http.StatusOK, result)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "rankingkey"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Order")
	f.SetCellValue(mode, "B1", "StaffId")
	f.SetCellValue(mode, "C1", "First Name")
	f.SetCellValue(mode, "D1", "Last Name")
	f.SetCellValue(mode, "E1", "Nick Name")
	f.SetCellValue(mode, "F1", "Position")
	f.SetCellValue(mode, "G1", "Department")
	f.SetCellValue(mode, "H1", "Staff Child")
	f.SetCellValue(mode, "I1", "Inv Amount")
	f.SetCellValue(mode, "J1", "Inv Amount Old")
	f.SetCellValue(mode, "K1", "Goal Total")
	f.SetCellValue(mode, "L1", "Score Target")
	f.SetCellValue(mode, "M1", "Score Sf")
	f.SetCellValue(mode, "N1", "Sale Factor")
	f.SetCellValue(mode, "O1", "Total So")
	f.SetCellValue(mode, "P1", "If Factor")
	f.SetCellValue(mode, "Q1", "Eng Cost")
	f.SetCellValue(mode, "R1", "Revenue")
	f.SetCellValue(mode, "S1", "Score If")
	f.SetCellValue(mode, "T1", "In Factor")
	f.SetCellValue(mode, "U1", "One Id")
	f.SetCellValue(mode, "V1", "Image")
	f.SetCellValue(mode, "W1", "File Name")
	f.SetCellValue(mode, "X1", "Growth Rate")
	f.SetCellValue(mode, "Y1", "Score Growth")
	f.SetCellValue(mode, "Z1", "Score All")
	f.SetCellValue(mode, "AA1", "Quarter")
	f.SetCellValue(mode, "AB1", "Year")
	f.SetCellValue(mode, "AC1", "Job Months")
	f.SetCellValue(mode, "AD1", "Commission")

	colOrder := "A"
	colStaffId := "B"
	colFname := "C"
	colLname := "D"
	colNname := "E"
	colPosition := "F"
	colDepartment := "G"
	colStaffChild := "H"
	colInvAmount := "I"
	colInvAmountOld := "J"
	colGoalTotal := "K"
	colScoreTarget := "L"
	colScoreSf := "M"
	colSaleFactor := "N"
	colTotalSo := "O"
	colIfFactor := "P"
	colEngCost := "Q"
	colRevenue := "R"
	colScoreIf := "S"
	colInFactor := "T"
	colOneId := "U"
	colImage := "V"
	colFileName := "W"
	colGrowthRate := "X"
	colScoreGrowth := "Y"
	colScoreAll := "Z"
	colQuarter := "AA"
	colYear := "AB"
	colJobMonths := "AC"
	colCommission := "AD"

	for k, v := range dataResultEx {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colOrder, k+2), v.Order)
		f.SetCellValue(mode, fmt.Sprint(colStaffId, k+2), v.StaffId)
		f.SetCellValue(mode, fmt.Sprint(colFname, k+2), v.Fname)
		f.SetCellValue(mode, fmt.Sprint(colLname, k+2), v.Lname)
		f.SetCellValue(mode, fmt.Sprint(colNname, k+2), v.Nname)
		f.SetCellValue(mode, fmt.Sprint(colPosition, k+2), v.Position)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colStaffChild, k+2), v.StaffChild)
		f.SetCellValue(mode, fmt.Sprint(colInvAmount, k+2), v.InvAmount)
		f.SetCellValue(mode, fmt.Sprint(colInvAmountOld, k+2), v.InvAmountOld)
		f.SetCellValue(mode, fmt.Sprint(colGoalTotal, k+2), v.GoalTotal)
		f.SetCellValue(mode, fmt.Sprint(colScoreTarget, k+2), v.ScoreTarget)
		f.SetCellValue(mode, fmt.Sprint(colScoreSf, k+2), v.ScoreSf)

		f.SetCellValue(mode, fmt.Sprint(colSaleFactor, k+2), v.SaleFactor)
		f.SetCellValue(mode, fmt.Sprint(colTotalSo, k+2), v.TotalSo)
		f.SetCellValue(mode, fmt.Sprint(colIfFactor, k+2), v.IfFactor)
		f.SetCellValue(mode, fmt.Sprint(colEngCost, k+2), v.EngCost)
		f.SetCellValue(mode, fmt.Sprint(colRevenue, k+2), v.Revenue)
		f.SetCellValue(mode, fmt.Sprint(colScoreIf, k+2), v.ScoreIf)

		f.SetCellValue(mode, fmt.Sprint(colInFactor, k+2), v.InFactor)
		f.SetCellValue(mode, fmt.Sprint(colOneId, k+2), v.OneId)
		f.SetCellValue(mode, fmt.Sprint(colImage, k+2), v.Image)
		f.SetCellValue(mode, fmt.Sprint(colFileName, k+2), v.FileName)
		f.SetCellValue(mode, fmt.Sprint(colGrowthRate, k+2), v.GrowthRate)
		f.SetCellValue(mode, fmt.Sprint(colScoreGrowth, k+2), v.ScoreGrowth)
		f.SetCellValue(mode, fmt.Sprint(colScoreAll, k+2), v.ScoreAll)
		f.SetCellValue(mode, fmt.Sprint(colQuarter, k+2), v.Quarter)
		f.SetCellValue(mode, fmt.Sprint(colYear, k+2), v.Year)
		f.SetCellValue(mode, fmt.Sprint(colJobMonths, k+2), v.JobMonths)
		f.SetCellValue(mode, fmt.Sprint(colCommission, k+2), v.Commission)
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

func GettReportExcelRankRecoveEndPoint(c echo.Context) error {

	conKey := strings.Split(util.GetEnv("CONDITION_GOV_RECOVER_SALE", ""), ",")
	var dFilter []string
	for _, v := range conKey {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}

	listStaffId, err := CheckPermissionRecovery(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		return echo.ErrInternalServerError
	}
	if len(listStaffId) == 0 {
		return c.JSON(http.StatusNoContent, nil)
	}
	// page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	// if strings.TrimSpace(c.QueryParam("page")) == "" {
	// 	page = 1
	// }
	q := strings.TrimSpace(c.QueryParam("quarter"))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
		WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
		WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
		WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
		WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
		WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
		WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
		WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
		ELSE -5 END
	) as score_target,
	(CASE
		WHEN goal_total is null or goal_total = 0 THEN 999
		ELSE (sum(inv_amount)/goal_total)*100 END
	) as percent_target,
	(CASE
		WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
		WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
		WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
		WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
		WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
		WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
		WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
		ELSE 0 END
	) as score_sf,
	(case
		when sum(revenue)/sum(engcost) is null then 0
		else sum(revenue)/sum(engcost) end
	) as sale_factor,
	(CASE
		WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
		WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
		WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
		WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
		WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
		WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
		WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
		ELSE 0 END
	) as score_if,
	(case
		when sum(sum_if)/sum(total_so) is null then 0
		else sum(sum_if)/sum(total_so) end
	) as in_factor,
	sum(revenue) as total_revenue,
	all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
	from (
		select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
		(case
			when inv_amount is null then 0
			ELSE inv_amount END) as inv_amount
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,
			'normal' as typestaff,
			sum((CASE
				WHEN TotalContractAmount is null THEN 0
				ELSE TotalContractAmount END
			)) as revenue,
			sum((CASE
				WHEN eng_cost is null THEN 0
				ELSE eng_cost END
			)) as engcost,
			sum((CASE
				WHEN in_factor is null THEN 0
				ELSE in_factor END
			)) as sum_if,
			sum((CASE
				WHEN sonumber is null THEN 0
				ELSE 1 END
			)) as total_so,
			one_id,
			quarter,year,job_months,staff_child
			from (
					select staff_id,fname,lname,nname,department,position,
					(CASE
						WHEN goal_total is null THEN 0
						ELSE goal_total END
					) as goal_total,
					staff_info.one_id,
					(CASE
						WHEN quarter is null THEN ?
						ELSE quarter END
					) as quarter,
					(CASE
						WHEN year is null THEN year(now())
						ELSE year END
					) as year,
					12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
					staff_child
					from staff_info
					left join
					(
						select * from goal_quarter where year = year(now()) and quarter =  ?
					) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
					left join staff_start on staff_info.one_id = staff_start.one_id
					WHERE staff_id in (
						select staff_id from staff_info WHERE staff_child = ''
					)
					group by staff_id
			) staff_detail
			LEFT JOIN (
				select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
				group by sonumber
			) total_so on total_so.sale_code = staff_detail.staff_id
			group by staff_id
		) tb_main
		LEFT join (
			select sum(PeriodAmount) as inv_amount, sale_code from (
				select PeriodAmount,sale_code
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now())   and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv group by sale_code
		) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
		where staff_id is not null and staff_id <> ''
	) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
	WHERE staff_id in (?)
	group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child = ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`
	sqlFilter := `select * from staff_info where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?) `

	var staffInfo []m.StaffInfo
	mapCnStaff := map[string][]string{}
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		var so []m.SOMssql
		if err := dbSale.Ctx().Model(&m.SOMssql{}).Where(`sale_code IN (?) AND INCSCDocNo = '' AND quarter(ContractStartDate) = ? AND year(ContractStartDate) = year(now()) AND DATEDIFF(NOW(),PeriodEndDate) > 60`, listStaffId, quarterNum-1).Group("Customer_ID").Find(&so).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		for _, s := range so {
			mapCnStaff[s.SaleCode] = append(mapCnStaff[s.SaleCode], s.INCSCDocNo)
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for index, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
				if r.InvAmount > r.InvAmountOld {
					i := len(mapCnStaff[r.StaffId])
					x := 0
					if len(mapCnStaff[r.StaffId]) > 0 {
						x = (i * 1000) + ((i - 1) * 1000)
					}
					// wait cal aging & blacklist
					baseCal := r.InvAmountOld * 0.003
					growthCal := (r.InvAmount - r.InvAmountOld) * 0.03
					saleFactor := (baseCal + growthCal) * (r.SaleFactor * r.SaleFactor)
					r.Commission = saleFactor*(r.InFactor/0.7) - float64(x)
				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth
		report[index] = r
	}
	if len(report) > 1 {
		sort.SliceStable(report, func(i, j int) bool { return report[i].ScoreAll > report[j].ScoreAll })
	}
	for i, r := range report {
		report[i].Order = i + 1
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, report[i])
				}
			}
		}
	}
	var dataResultEx []m.OrgChart
	var result m.Result
	// if len(dataResult) > (page * 10) {
	// 	start := (page - 1) * 10
	// 	end := (page * 10)
	// 	result.Data = dataResult[start:end]
	// 	result.Count = len(dataResult[start:end])
	// 	dataResultEx = dataResult[start:end]
	// } else {
	// 	start := (page * 10) - (10)
	// 	result.Data = dataResult[start:]
	// 	result.Count = len(dataResult[start:])
	// 	dataResultEx = dataResult[start:]
	// }

	result.Data = dataResult
	result.Count = len(dataResult)
	dataResultEx = dataResult

	result.Total = len(dataResult)
	// return c.JSON(http.StatusOK, result)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "rankingrecovery"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Order")
	f.SetCellValue(mode, "B1", "StaffId")
	f.SetCellValue(mode, "C1", "First Name")
	f.SetCellValue(mode, "D1", "Last Name")
	f.SetCellValue(mode, "E1", "Nick Name")
	f.SetCellValue(mode, "F1", "Position")
	f.SetCellValue(mode, "G1", "Department")
	f.SetCellValue(mode, "H1", "Staff Child")
	f.SetCellValue(mode, "I1", "Inv Amount")
	f.SetCellValue(mode, "J1", "Inv Amount Old")
	f.SetCellValue(mode, "K1", "Goal Total")
	f.SetCellValue(mode, "L1", "Score Target")
	f.SetCellValue(mode, "M1", "Score Sf")
	f.SetCellValue(mode, "N1", "Sale Factor")
	f.SetCellValue(mode, "O1", "Total So")
	f.SetCellValue(mode, "P1", "If Factor")
	f.SetCellValue(mode, "Q1", "Eng Cost")
	f.SetCellValue(mode, "R1", "Revenue")
	f.SetCellValue(mode, "S1", "Score If")
	f.SetCellValue(mode, "T1", "In Factor")
	f.SetCellValue(mode, "U1", "One Id")
	f.SetCellValue(mode, "V1", "Image")
	f.SetCellValue(mode, "W1", "File Name")
	f.SetCellValue(mode, "X1", "Growth Rate")
	f.SetCellValue(mode, "Y1", "Score Growth")
	f.SetCellValue(mode, "Z1", "Score All")
	f.SetCellValue(mode, "AA1", "Quarter")
	f.SetCellValue(mode, "AB1", "Year")
	f.SetCellValue(mode, "AC1", "Job Months")
	f.SetCellValue(mode, "AD1", "Commission")

	colOrder := "A"
	colStaffId := "B"
	colFname := "C"
	colLname := "D"
	colNname := "E"
	colPosition := "F"
	colDepartment := "G"
	colStaffChild := "H"
	colInvAmount := "I"
	colInvAmountOld := "J"
	colGoalTotal := "K"
	colScoreTarget := "L"
	colScoreSf := "M"
	colSaleFactor := "N"
	colTotalSo := "O"
	colIfFactor := "P"
	colEngCost := "Q"
	colRevenue := "R"
	colScoreIf := "S"
	colInFactor := "T"
	colOneId := "U"
	colImage := "V"
	colFileName := "W"
	colGrowthRate := "X"
	colScoreGrowth := "Y"
	colScoreAll := "Z"
	colQuarter := "AA"
	colYear := "AB"
	colJobMonths := "AC"
	colCommission := "AD"

	for k, v := range dataResultEx {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colOrder, k+2), v.Order)
		f.SetCellValue(mode, fmt.Sprint(colStaffId, k+2), v.StaffId)
		f.SetCellValue(mode, fmt.Sprint(colFname, k+2), v.Fname)
		f.SetCellValue(mode, fmt.Sprint(colLname, k+2), v.Lname)
		f.SetCellValue(mode, fmt.Sprint(colNname, k+2), v.Nname)
		f.SetCellValue(mode, fmt.Sprint(colPosition, k+2), v.Position)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colStaffChild, k+2), v.StaffChild)
		f.SetCellValue(mode, fmt.Sprint(colInvAmount, k+2), v.InvAmount)
		f.SetCellValue(mode, fmt.Sprint(colInvAmountOld, k+2), v.InvAmountOld)
		f.SetCellValue(mode, fmt.Sprint(colGoalTotal, k+2), v.GoalTotal)
		f.SetCellValue(mode, fmt.Sprint(colScoreTarget, k+2), v.ScoreTarget)
		f.SetCellValue(mode, fmt.Sprint(colScoreSf, k+2), v.ScoreSf)

		f.SetCellValue(mode, fmt.Sprint(colSaleFactor, k+2), v.SaleFactor)
		f.SetCellValue(mode, fmt.Sprint(colTotalSo, k+2), v.TotalSo)
		f.SetCellValue(mode, fmt.Sprint(colIfFactor, k+2), v.IfFactor)
		f.SetCellValue(mode, fmt.Sprint(colEngCost, k+2), v.EngCost)
		f.SetCellValue(mode, fmt.Sprint(colRevenue, k+2), v.Revenue)
		f.SetCellValue(mode, fmt.Sprint(colScoreIf, k+2), v.ScoreIf)

		f.SetCellValue(mode, fmt.Sprint(colInFactor, k+2), v.InFactor)
		f.SetCellValue(mode, fmt.Sprint(colOneId, k+2), v.OneId)
		f.SetCellValue(mode, fmt.Sprint(colImage, k+2), v.Image)
		f.SetCellValue(mode, fmt.Sprint(colFileName, k+2), v.FileName)
		f.SetCellValue(mode, fmt.Sprint(colGrowthRate, k+2), v.GrowthRate)
		f.SetCellValue(mode, fmt.Sprint(colScoreGrowth, k+2), v.ScoreGrowth)
		f.SetCellValue(mode, fmt.Sprint(colScoreAll, k+2), v.ScoreAll)
		f.SetCellValue(mode, fmt.Sprint(colQuarter, k+2), v.Quarter)
		f.SetCellValue(mode, fmt.Sprint(colYear, k+2), v.Year)
		f.SetCellValue(mode, fmt.Sprint(colJobMonths, k+2), v.JobMonths)
		f.SetCellValue(mode, fmt.Sprint(colCommission, k+2), v.Commission)
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

func GetReportExcelRankTeamLeadEndPoint(c echo.Context) error {

	conKey := strings.Split(util.GetEnv("CONDITION_GOV_RECOVER_SALE", ""), ",")
	var dFilter []string
	for _, v := range conKey {
		t := fmt.Sprintf(`INSTR(CONCAT_WS('|', department), '%s')`, v)
		dFilter = append(dFilter, t)
	}
	finalFilter := fmt.Sprintf(` %s `, strings.Join(dFilter, " OR "))
	if strings.TrimSpace(c.QueryParam(("staff_id"))) == "" || strings.TrimSpace(c.QueryParam("quarter")) == "" {
		return c.JSON(http.StatusBadRequest, m.Result{Message: "invalid staff id or quarter"})
	}

	listStaffId, err := CheckPermissionTeamLead(strings.TrimSpace(c.QueryParam(("staff_id"))), finalFilter)
	if err != nil {
		return echo.ErrInternalServerError
	}

	// page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	// if strings.TrimSpace(c.QueryParam("page")) == "" {
	// 	page = 1
	// }
	q := strings.TrimSpace(c.QueryParam("quarter"))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	today := time.Now()
	yearNow, _, _ := today.Date()
	yearBefore := yearNow
	var quarterBefore string
	var quarterBeforeNum int
	quarterNum, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("quarter")))
	var quarter string

	if q == "1" {
		quarter = "Q1"
		quarterBefore = "Q4"
		quarterBeforeNum = 4
		yearBefore = yearNow - 1
	} else if q == "2" {
		quarter = "Q2"
		quarterBefore = "Q1"
		quarterBeforeNum = 1
	} else if q == "3" {
		quarter = "Q3"
		quarterBefore = "Q2"
		quarterBeforeNum = 2
	} else {
		quarter = "Q4"
		quarterBefore = "Q3"
		quarterBeforeNum = 3
	}
	var report []m.OrgChart
	var invBefore []m.InvBefore
	sql := `select staff_id,fname,lname,nname,department,sum(inv_amount) as inv_amount,max(goal_total) as goal_total, count(staff_id) as checkdata,typestaff, 0 as inv_amount_old,
		(CASE
			WHEN goal_total is null or goal_total = 0 THEN 30
			WHEN (sum(inv_amount)/goal_total)*100 >= 176 THEN 30
			WHEN (sum(inv_amount)/goal_total)*100 >= 156 THEN 28
			WHEN (sum(inv_amount)/goal_total)*100 >= 126 THEN 25
			WHEN (sum(inv_amount)/goal_total)*100 >= 101 THEN 20
			WHEN (sum(inv_amount)/goal_total)*100 >= 76 THEN 15
			WHEN (sum(inv_amount)/goal_total)*100 >= 51 THEN 10
			WHEN (sum(inv_amount)/goal_total)*100 >= 1 THEN 5
			WHEN (sum(inv_amount)/goal_total)*100 = 0 THEN 0
			ELSE -5 END
		) as score_target,
		(CASE
			WHEN goal_total is null or goal_total = 0 THEN 999
			ELSE (sum(inv_amount)/goal_total)*100 END
		) as percent_target,
		(CASE
			WHEN sum(revenue)/sum(engcost) >= 1.2 THEN 25
			WHEN sum(revenue)/sum(engcost) >= 1.16 THEN 22
			WHEN sum(revenue)/sum(engcost) >= 1.1 THEN 19
			WHEN sum(revenue)/sum(engcost) >= 1.06 THEN 16
			WHEN sum(revenue)/sum(engcost) >= 1.0 THEN 13
			WHEN sum(revenue)/sum(engcost) >= 0.96 THEN 10
			WHEN sum(revenue)/sum(engcost) >= 0.9 THEN 7
			ELSE 0 END
		) as score_sf,
		(case
			when sum(revenue)/sum(engcost) is null then 0
			else sum(revenue)/sum(engcost) end
		) as sale_factor,
		(CASE
			WHEN sum(sum_if)/sum(total_so) >= 1 THEN 25
			WHEN sum(sum_if)/sum(total_so) >= 0.86 THEN 22
			WHEN sum(sum_if)/sum(total_so) >= 0.8 THEN 19
			WHEN sum(sum_if)/sum(total_so) >= 0.76 THEN 16
			WHEN sum(sum_if)/sum(total_so) >= 0.7 THEN 13
			WHEN sum(sum_if)/sum(total_so) >= 0.66 THEN 10
			WHEN sum(sum_if)/sum(total_so) >= 0.6 THEN 7
			ELSE 0 END
		) as score_if,
		(case
			when sum(sum_if)/sum(total_so) is null then 0
			else sum(sum_if)/sum(total_so) end
		) as in_factor,
		sum(revenue) as total_revenue,
		all_ranking.one_id, image,filename, -100 as growth_rate, -5 as score_growth,0 as score_all,quarter,year,position,job_months,staff_child
		from (
			select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
			(case
				when inv_amount is null then 0
				ELSE inv_amount END) as inv_amount
			from (
				select staff_id,fname,lname,nname,department,position,goal_total,
				'normal' as typestaff,
				sum((CASE
					WHEN TotalContractAmount is null THEN 0
					ELSE TotalContractAmount END
				)) as revenue,
				sum((CASE
					WHEN eng_cost is null THEN 0
					ELSE eng_cost END
				)) as engcost,
				sum((CASE
					WHEN in_factor is null THEN 0
					ELSE in_factor END
				)) as sum_if,
				sum((CASE
					WHEN sonumber is null THEN 0
					ELSE 1 END
				)) as total_so,
				one_id,
				quarter,year,job_months,staff_child
				from (
						select staff_id,fname,lname,nname,department,position,
						(CASE
							WHEN goal_total is null THEN 0
							ELSE goal_total END
						) as goal_total,
						staff_info.one_id,
						(CASE
							WHEN quarter is null THEN ?
							ELSE quarter END
						) as quarter,
						(CASE
							WHEN year is null THEN year(now())
							ELSE year END
						) as year,
						12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
						staff_child
						from staff_info
						left join
						(
							select * from goal_quarter where year = year(now()) and quarter = ?
						) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
						left join staff_start on staff_info.one_id = staff_start.one_id
						WHERE staff_id in (
							select staff_id from staff_info WHERE staff_child <> ''
						)
						group by staff_id
				) staff_detail
				LEFT JOIN (
					select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
					group by sonumber
				) total_so on total_so.sale_code = staff_detail.staff_id
				group by staff_id
			) tb_main
			LEFT join (
				select sum(PeriodAmount) as inv_amount, sale_code from (
					select PeriodAmount,sale_code
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
					group by sonumber
				) tb_inv group by sale_code
			) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_code
			where staff_id is not null and staff_id <> ''
			union
			select staff_id,fname,lname,nname,department,position,goal_total,typestaff,revenue,engcost,sum_if,total_so,one_id,quarter,year,job_months,staff_child,
			(case
				when inv_amount is null then 0
				ELSE inv_amount END) as inv_amount
			from (
				select staff_id,fname,lname,nname,department,position,goal_total,
				'normal' as typestaff,
				sum((CASE
					WHEN TotalContractAmount is null THEN 0
					ELSE TotalContractAmount END
				)) as revenue,
				sum((CASE
					WHEN eng_cost is null THEN 0
					ELSE eng_cost END
				)) as engcost,
				sum((CASE
					WHEN in_factor is null THEN 0
					ELSE in_factor END
				)) as sum_if,
				sum((CASE
					WHEN sonumber is null THEN 0
					ELSE 1 END
				)) as total_so,
				one_id,
				quarter,year,job_months,staff_child
				from (
						select staff_id,fname,lname,nname,department,position,
						(CASE
							WHEN goal_total is null THEN 0
							ELSE goal_total END
						) as goal_total,
						staff_info.one_id,
						(CASE
							WHEN quarter is null THEN ?
							ELSE quarter END
						) as quarter,
						(CASE
							WHEN year is null THEN year(now())
							ELSE year END
						) as year,
						12 * (YEAR(NOW()) - YEAR(start_job)) + (MONTH(NOW()) - MONTH(start_job)) AS job_months,
						staff_child
						from staff_info
						left join
						(
							select * from goal_quarter where year = year(now()) and quarter = ?
						) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
						left join staff_start on staff_info.one_id = staff_start.one_id
						WHERE staff_id in (
							select staff_id from staff_info WHERE staff_child <> ''
						)
						group by staff_id
				) staff_detail
				LEFT JOIN (
					select sale_lead,TotalContractAmount,sonumber,sale_code,sale_factor,in_factor,(TotalContractAmount/sale_factor) as eng_cost
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and Active_Inactive = 'Active'
					group by sonumber
				) total_so on total_so.sale_lead = staff_detail.staff_id
				group by staff_id
			) tb_main
			LEFT join (
				select sum(PeriodAmount) as inv_amount,sale_lead from (
					select PeriodAmount,sale_code,sale_lead
					from so_mssql
					WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = year(now()) and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
					group by sonumber
				) tb_inv group by sale_lead
			) tb_inv_now on tb_main.staff_id = tb_inv_now.sale_lead
			where staff_id is not null and staff_id <> ''
		) all_ranking LEFT JOIN staff_images ON all_ranking.one_id = staff_images.one_id
		WHERE staff_id in (?)
		group by staff_id;`

	sqlBefore := `select staff_id,count(staff_id) as checkdata,sum(inv_amount) as inv_amount
	from (
		select staff_id,sum(PeriodAmount) as inv_amount,count(sonumber) as total_so
		from (
			select staff_id from staff_info
			left join
			(
				select * from goal_quarter where year = ? and quarter = ?
			) goal_quarter on staff_info.staff_id = goal_quarter.ref_staff
			WHERE staff_id in (
				select staff_id from staff_info WHERE staff_child <> ''
			)
			group by staff_id
		) staff_detail
		LEFT JOIN (
			select PeriodAmount,sale_code,sonumber, type_sale
			from (
				select PeriodAmount,sale_code,sonumber , 'normal' as type_sale
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
				union
				select PeriodAmount,sale_lead as sale_code,sonumber, 'lead'
				from so_mssql
				WHERE quarter(ContractStartDate) = ? and year(ContractStartDate) = ? and so_refer = '' and Active_Inactive = 'Active' and SOWebStatus not like '%%Terminate%%'
				group by sonumber
			) tb_inv_old
		) total_new_so on total_new_so.sale_code = staff_detail.staff_id
		where staff_id is not null and staff_id <> '' and sale_code is not null
		group by staff_id
	) all_ranking
	WHERE staff_id in (?)
	group by staff_id;`

	sqlFilter := `select * from staff_info where staff_child <> '' AND  INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?)  `

	var staffInfo []m.StaffInfo
	mapCnStaff := map[string][]string{}
	hasErr := 0
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		if err := dbSale.Ctx().Raw(sql, quarter, quarter, quarterNum, quarterNum, quarter, quarter, quarterNum, quarterNum, listStaffId).Scan(&report).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlBefore, yearBefore, quarterBefore, quarterBeforeNum, yearBefore, quarterBeforeNum, yearBefore, listStaffId).Scan(&invBefore).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		if err := dbSale.Ctx().Raw(sqlFilter, filter).Scan(&staffInfo).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				hasErr += 1
				log.Errorln(pkgName, err, "select data error :-")
			}
		}
		wg.Done()
	}()
	go func() {
		var so []m.SOMssql
		if err := dbSale.Ctx().Model(&m.SOMssql{}).Where(`sale_code IN (?) AND INCSCDocNo = '' AND quarter(ContractStartDate) = ? AND year(ContractStartDate) = year(now()) AND DATEDIFF(NOW(),PeriodEndDate) > 60`, listStaffId, quarterNum-1).Group("Customer_ID").Find(&so).Error; err != nil {
			if !gorm.IsRecordNotFoundError(err) {
				log.Errorln(pkgName, err, "select data error :-")
				hasErr += 1
			}
		}
		for _, s := range so {
			mapCnStaff[s.SaleCode] = append(mapCnStaff[s.SaleCode], s.INCSCDocNo)
		}
		wg.Done()
	}()
	wg.Wait()

	if hasErr != 0 {
		return echo.ErrInternalServerError
	}
	var dataResult []m.OrgChart
	for index, r := range report {
		for _, i := range invBefore {
			if i.StaffID == r.StaffId {
				r.InvAmountOld = i.InvAmount
				r.GrowthRate = ((r.InvAmount - i.InvAmount) / i.InvAmount) * 100

				if r.GrowthRate >= 80 {
					r.ScoreGrowth = 50
				} else if r.GrowthRate >= 60 {
					r.ScoreGrowth = 44
				} else if r.GrowthRate >= 50 {
					r.ScoreGrowth = 38
				} else if r.GrowthRate >= 40 {
					r.ScoreGrowth = 32
				} else if r.GrowthRate >= 30 {
					r.ScoreGrowth = 25
				} else if r.GrowthRate >= 20 {
					r.ScoreGrowth = 18
				} else if r.GrowthRate >= 10 {
					r.ScoreGrowth = 11
				} else if r.GrowthRate >= 0 {
					r.ScoreGrowth = 4
				} else {
					r.ScoreGrowth = 0
				}
				if r.InvAmount > r.InvAmountOld {
					i := len(mapCnStaff[r.StaffId])
					x := 0
					if len(mapCnStaff[r.StaffId]) > 0 {
						x = (i * 1000) + ((i - 1) * 1000)
					}
					// wait cal aging & blacklist
					baseCal := r.InvAmountOld * 0.003
					growthCal := (r.InvAmount - r.InvAmountOld) * 0.03
					saleFactor := (baseCal + growthCal) * (r.SaleFactor * r.SaleFactor)
					r.Commission = saleFactor*(r.InFactor/0.7) - float64(x)
				}
			}
		}
		r.ScoreAll += r.ScoreSf + r.ScoreIf + r.ScoreGrowth
		report[index] = r
	}
	if len(report) > 1 {
		sort.SliceStable(report, func(i, j int) bool { return report[i].ScoreAll > report[j].ScoreAll })
	}
	for i, r := range report {
		report[i].Order = i + 1
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, report[i])
				}
			}
		}
	}
	var result m.Result
	// if len(dataResult) > (page * 10) {
	// 	start := (page - 1) * 10
	// 	end := (page * 10)
	// 	result.Data = dataResult[start:end]
	// 	result.Count = len(dataResult[start:end])
	// } else {
	// 	start := (page * 10) - (10)
	// 	result.Data = dataResult[start:]
	// 	result.Count = len(dataResult[start:])
	// }

	result.Data = dataResult
	result.Count = len(dataResult)

	result.Total = len(dataResult)
	var dataResultEx []m.OrgChart
	dataResultEx = dataResult
	// return c.JSON(http.StatusOK, result)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "rankinglead"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Order")
	f.SetCellValue(mode, "B1", "StaffId")
	f.SetCellValue(mode, "C1", "First Name")
	f.SetCellValue(mode, "D1", "Last Name")
	f.SetCellValue(mode, "E1", "Nick Name")
	f.SetCellValue(mode, "F1", "Position")
	f.SetCellValue(mode, "G1", "Department")
	f.SetCellValue(mode, "H1", "Staff Child")
	f.SetCellValue(mode, "I1", "Inv Amount")
	f.SetCellValue(mode, "J1", "Inv Amount Old")
	f.SetCellValue(mode, "K1", "Goal Total")
	f.SetCellValue(mode, "L1", "Score Target")
	f.SetCellValue(mode, "M1", "Score Sf")
	f.SetCellValue(mode, "N1", "Sale Factor")
	f.SetCellValue(mode, "O1", "Total So")
	f.SetCellValue(mode, "P1", "If Factor")
	f.SetCellValue(mode, "Q1", "Eng Cost")
	f.SetCellValue(mode, "R1", "Revenue")
	f.SetCellValue(mode, "S1", "Score If")
	f.SetCellValue(mode, "T1", "In Factor")
	f.SetCellValue(mode, "U1", "One Id")
	f.SetCellValue(mode, "V1", "Image")
	f.SetCellValue(mode, "W1", "File Name")
	f.SetCellValue(mode, "X1", "Growth Rate")
	f.SetCellValue(mode, "Y1", "Score Growth")
	f.SetCellValue(mode, "Z1", "Score All")
	f.SetCellValue(mode, "AA1", "Quarter")
	f.SetCellValue(mode, "AB1", "Year")
	f.SetCellValue(mode, "AC1", "Job Months")
	f.SetCellValue(mode, "AD1", "Commission")

	colOrder := "A"
	colStaffId := "B"
	colFname := "C"
	colLname := "D"
	colNname := "E"
	colPosition := "F"
	colDepartment := "G"
	colStaffChild := "H"
	colInvAmount := "I"
	colInvAmountOld := "J"
	colGoalTotal := "K"
	colScoreTarget := "L"
	colScoreSf := "M"
	colSaleFactor := "N"
	colTotalSo := "O"
	colIfFactor := "P"
	colEngCost := "Q"
	colRevenue := "R"
	colScoreIf := "S"
	colInFactor := "T"
	colOneId := "U"
	colImage := "V"
	colFileName := "W"
	colGrowthRate := "X"
	colScoreGrowth := "Y"
	colScoreAll := "Z"
	colQuarter := "AA"
	colYear := "AB"
	colJobMonths := "AC"
	colCommission := "AD"

	for k, v := range dataResultEx {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colOrder, k+2), v.Order)
		f.SetCellValue(mode, fmt.Sprint(colStaffId, k+2), v.StaffId)
		f.SetCellValue(mode, fmt.Sprint(colFname, k+2), v.Fname)
		f.SetCellValue(mode, fmt.Sprint(colLname, k+2), v.Lname)
		f.SetCellValue(mode, fmt.Sprint(colNname, k+2), v.Nname)
		f.SetCellValue(mode, fmt.Sprint(colPosition, k+2), v.Position)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colStaffChild, k+2), v.StaffChild)
		f.SetCellValue(mode, fmt.Sprint(colInvAmount, k+2), v.InvAmount)
		f.SetCellValue(mode, fmt.Sprint(colInvAmountOld, k+2), v.InvAmountOld)
		f.SetCellValue(mode, fmt.Sprint(colGoalTotal, k+2), v.GoalTotal)
		f.SetCellValue(mode, fmt.Sprint(colScoreTarget, k+2), v.ScoreTarget)
		f.SetCellValue(mode, fmt.Sprint(colScoreSf, k+2), v.ScoreSf)

		f.SetCellValue(mode, fmt.Sprint(colSaleFactor, k+2), v.SaleFactor)
		f.SetCellValue(mode, fmt.Sprint(colTotalSo, k+2), v.TotalSo)
		f.SetCellValue(mode, fmt.Sprint(colIfFactor, k+2), v.IfFactor)
		f.SetCellValue(mode, fmt.Sprint(colEngCost, k+2), v.EngCost)
		f.SetCellValue(mode, fmt.Sprint(colRevenue, k+2), v.Revenue)
		f.SetCellValue(mode, fmt.Sprint(colScoreIf, k+2), v.ScoreIf)

		f.SetCellValue(mode, fmt.Sprint(colInFactor, k+2), v.InFactor)
		f.SetCellValue(mode, fmt.Sprint(colOneId, k+2), v.OneId)
		f.SetCellValue(mode, fmt.Sprint(colImage, k+2), v.Image)
		f.SetCellValue(mode, fmt.Sprint(colFileName, k+2), v.FileName)
		f.SetCellValue(mode, fmt.Sprint(colGrowthRate, k+2), v.GrowthRate)
		f.SetCellValue(mode, fmt.Sprint(colScoreGrowth, k+2), v.ScoreGrowth)
		f.SetCellValue(mode, fmt.Sprint(colScoreAll, k+2), v.ScoreAll)
		f.SetCellValue(mode, fmt.Sprint(colQuarter, k+2), v.Quarter)
		f.SetCellValue(mode, fmt.Sprint(colYear, k+2), v.Year)
		f.SetCellValue(mode, fmt.Sprint(colJobMonths, k+2), v.JobMonths)
		f.SetCellValue(mode, fmt.Sprint(colCommission, k+2), v.Commission)
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

func GetReportExcelSaleFactorEndPoint(c echo.Context) error {
	accountId := strings.TrimSpace(c.Param("id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	check := checkPermissionUser(accountId)
	if !check {
		return echo.ErrNotFound
	}
	type SaleFactorPerson struct {
		TotalRevenue float64 `json:"total_revenue" gorm:"column:total_revenue"`
		CountSo      int     `json:"count_so" gorm:"column:count_so"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		EngCost      float64 `json:"eng_cost" gorm:"column:engcost"`
		RealSF       float64 `json:"real_sf" gorm:"column:real_sf"`
		Department   string  `json:"department" gorm:"column:department"`
		StaffId      string  `json:"staff_id" gorm:"column:staff_id"`
		StaffChild   string  `json:"staff_child" gorm:"column:staff_child"`
		Fname        string  `json:"fname" gorm:"column:fname"`
		Lname        string  `json:"lname" gorm:"column:lname"`
		Nname        string  `json:"nname" gorm:"column:nname"`
	}
	today := time.Now()
	year, month, _ := today.Date()
	var countSale []CountSoPerson
	var saleFac []SaleFactorPerson
	var dataRaw []SaleFactorPerson
	sql := `select
				sum(in_factor) as in_factor,
				sum(ex_factor) as ex_factor,
				sum(revenue) as total_revenue,
				sum(engcost) as engcost,
				sum(revenue)/sum(engcost) as real_sf,
				department,staff_id,fname,lname,nname,staff_child
			from (
				Select
						TotalContractAmount as revenue,
						(CASE
								WHEN TotalContractAmount is not null and sale_factor is not null and sale_factor != 0 THEN TotalContractAmount/sale_factor
								ELSE 0 END
						) as engcost,
						sale_factor,
						sale_code,
						in_factor,
						ex_factor
				from so_mssql where month(PeriodStartDate) = ? and year(PeriodStartDate) = ?
				group by sonumber
			) tb_so
			LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
			where department in (
				SELECT department FROM staff_info WHERE staff_child = '' and department <> 'Sale JV' GROUP BY department
			)
			GROUP BY staff_id ORDER BY real_sf desc`

	countCompany := `SELECT COUNT(Customer_ID) as count_so , department ,fname,lname,staff_id
			from (
					Select
									Customer_ID, Customer_name, sale_code
					from so_mssql where month(PeriodStartDate) = ? and year(PeriodStartDate) = ?
					group by Customer_ID
			) tb_so
			LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
			where department in (
					SELECT department FROM staff_info WHERE  department <> 'Sale JV' GROUP BY department
			) GROUP BY staff_id`
	sqlFactor := `select
					sum(in_factor)/count(sonumber) as in_fac,
					sum(ex_factor)/count(sonumber) as ex_fac

				from (
				Select
					TotalContractAmount as revenue,
					(CASE
						WHEN TotalContractAmount is not null and sale_factor is not null and sale_factor != 0 THEN TotalContractAmount/sale_factor
						ELSE 0 END
					) as engcost,
					sale_factor,
					in_factor,
					ex_factor,
					sale_code,
				sonumber
				from so_mssql
				group by sonumber
				) tb_so
				LEFT JOIN staff_info ON tb_so.sale_code = staff_info.staff_id
				where department in (
				SELECT department FROM staff_info WHERE department <> 'Sale JV' GROUP BY department)`
	sqlFilter := `select * from staff_info where INSTR(CONCAT_WS('|', staff_id, fname, lname, nname, position, department,one_id), ?) `

	var staffInfo []m.StaffInfo
	if err := dbSale.Ctx().Raw(sqlFilter, search).Scan(&staffInfo).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}

	if err := dbSale.Ctx().Raw(sql, month, year).Scan(&saleFac).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}
	if err := dbSale.Ctx().Raw(countCompany, month, year).Scan(&countSale).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}
	sumFac := struct {
		InFactor float64 `gorm:"column:in_fac"`
		ExFactor float64 `gorm:"column:ex_fac"`
	}{}
	if err := dbSale.Ctx().Raw(sqlFactor).Scan(&sumFac).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return echo.ErrInternalServerError
		}
	}

	sumInFac := 0.0
	sumExFac := 0.0
	sumSo := 0
	for _, v := range saleFac {
		for _, c := range countSale {
			if v.StaffId == c.StaffId {
				childCountSo := 0
				childInfac := 0.0
				childExfac := 0.0
				if v.StaffChild != "" {
					splitStaff := strings.Split(v.StaffChild, ",")
					for _, s := range splitStaff {
						for _, f := range saleFac {
							if s == f.StaffId {
								childInfac += v.InFactor
								childExfac += v.ExFactor
							}
						}
						for _, ch := range countSale {
							if v.StaffId == ch.StaffId {
								childCountSo += ch.CountSo
							}
						}
					}
				}
				sumSo += c.CountSo
				v.CountSo = c.CountSo + childCountSo
				v.InFactor = (v.InFactor + childInfac) / float64(c.CountSo+childCountSo)
				v.ExFactor = (v.ExFactor + childExfac) / float64(c.CountSo+childCountSo)

				sumInFac += (v.InFactor)
				sumExFac += (v.ExFactor)

				dataRaw = append(dataRaw, v)
			}
		}
	}
	// // fi"lter
	// var dataRe []SaleFactorPerson
	// for _, r := range dataRaw {
	// 	for _, s := range staffInfo {
	// 		if r.StaffId == s.StaffId {
	// 			dataRe = append(dataRe, r)
	// 		}
	// 	}
	// }

	// dataResult := map[string]interface{}{
	// 	"data":      dataRe,
	// 	"in_factor": sumFac.InFactor,
	// 	"ex_factor": sumFac.ExFactor,
	// }
	// return c.JSON(http.StatusOK, dataResult)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "salefactor"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "Total Revenue")
	f.SetCellValue(mode, "B1", "CountSo")
	f.SetCellValue(mode, "C1", "In Factor")
	f.SetCellValue(mode, "D1", "Ex Factor")
	f.SetCellValue(mode, "E1", "Eng Cost")
	f.SetCellValue(mode, "F1", "Real SF")
	f.SetCellValue(mode, "G1", "Department")
	f.SetCellValue(mode, "H1", "Staff Id")
	f.SetCellValue(mode, "I1", "Staff Child")
	f.SetCellValue(mode, "J1", "First Name")
	f.SetCellValue(mode, "K1", "Last Name")
	f.SetCellValue(mode, "L1", "Nick Name")

	colTotalRevenue := "A"
	colCountSo := "B"
	colInFactor := "C"
	colExFactor := "D"
	colEngCost := "E"
	colRealSF := "F"
	colDepartment := "G"
	colStaffId := "H"
	colStaffChild := "I"
	colFname := "J"
	colLname := "K"
	colNname := "L"

	for k, v := range dataRaw {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colTotalRevenue, k+2), v.TotalRevenue)
		f.SetCellValue(mode, fmt.Sprint(colCountSo, k+2), v.CountSo)
		f.SetCellValue(mode, fmt.Sprint(colInFactor, k+2), v.InFactor)
		f.SetCellValue(mode, fmt.Sprint(colExFactor, k+2), v.ExFactor)
		f.SetCellValue(mode, fmt.Sprint(colEngCost, k+2), v.EngCost)
		f.SetCellValue(mode, fmt.Sprint(colRealSF, k+2), v.RealSF)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colStaffId, k+2), v.StaffId)
		f.SetCellValue(mode, fmt.Sprint(colStaffChild, k+2), v.StaffChild)
		f.SetCellValue(mode, fmt.Sprint(colFname, k+2), v.Fname)
		f.SetCellValue(mode, fmt.Sprint(colLname, k+2), v.Lname)
		f.SetCellValue(mode, fmt.Sprint(colNname, k+2), v.Nname)
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

func GetExcelDetailReceiptEndPoint(c echo.Context) error {

	if strings.TrimSpace(c.QueryParam("sale_id")) == "" {
		return c.JSON(http.StatusBadRequest, server.Result{Message: "invalid sale id"})
	}

	saleId := strings.TrimSpace(c.QueryParam("sale_id"))
	search := strings.TrimSpace(c.QueryParam("search"))
	InvNumber := strings.TrimSpace(c.QueryParam("inv_number"))
	StaffId := strings.TrimSpace(c.QueryParam("staff_id"))

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

	type SOCus struct {
		BLSCDocNo   string  `json:"BLSCDocNo" gorm:"column:BLSCDocNo"`
		StaffID     string  `json:"staff_id" gorm:"column:staff_id"`
		Fname       string  `json:"fname" gorm:"column:fname"`
		Lname       string  `json:"lname" gorm:"column:lname"`
		Nname       string  `json:"nname" gorm:"column:nname"`
		Department  string  `json:"department" gorm:"column:department"`
		SoAmount    float64 `json:"so_amount" gorm:"column:so_amount"`
		Amount      float64 `json:"amount" gorm:"column:amount"`
		SOWebStatus string  `json:"so_web_status" gorm:"column:SOWebStatus"`
		CustomerID  string  `json:"Customer_ID" gorm:"column:Customer_ID"`
		CusnameThai string  `json:"Customer_Name" gorm:"column:Customer_Name"`
		// CusnameEng   string  `json:"Customer_Name" gorm:"column:Customer_Name"` //ไม่มีในcolumn so_ms
		BusinessType string  `json:"Business_type" gorm:"column:Business_type"`
		JobStatus    string  `json:"Job_Status" gorm:"column:Job_Status"`
		SoType       string  `json:"SO_Type" gorm:"column:SOType"`
		SaleFactor   float64 `json:"SaleFactors" gorm:"column:SaleFactors"`
		InFactor     float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor     float64 `json:"ex_factor" gorm:"column:ex_factor"`
		StartDate    string  `json:"PeriodStartDate" gorm:"column:PeriodStartDate"`
		EndDate      string  `json:"PeriodEndDate" gorm:"column:PeriodEndDate"`
		Detail       string  `json:"detail" gorm:"column:detail"`
	}
	type TrackInvoice struct {
		Amount     float64 `json:"amount" gorm:"column:amount"`
		InFactor   float64 `json:"in_factor" gorm:"column:in_factor"`
		ExFactor   float64 `json:"ex_factor" gorm:"column:ex_factor"`
		SaleFactor float64 `json:"sale_factor" gorm:"column:sale_factor"`
		ProRate    float64 `json:"pro_rate" gorm:"column:pro_rate"`
	}
	cus := []struct {
		CusnameThai string `json:"Customer_ID" gorm:"column:Customer_ID"`
	}{}
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
	dateFrom := time.Date(yearStart, monthStart, dayStart, 0, 0, 0, 0, time.Local)
	dateTo := time.Date(yearEnd, monthEnd, dayEnd, 0, 0, 0, 0, time.Local)
	so := []struct {
		InvNo string `json:"invoice_no" gorm:"column:invoice_no"`
	}{}

	sql := `
			 SELECT
			 invoice_no
				FROM (
					SELECT
					invoice_no,SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor
					FROM (
						SELECT *,CONCAT(BLSCDocNo) as invoice_no FROM so_mssql
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate
						and sale_code in (?)
					) sub_data
				) so_group
				GROUP by BLSCDocNo
			 `

	if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listId).Scan(&so).Error; err != nil {
		log.Errorln(pkgName, err, "select data error -:")
		return echo.ErrInternalServerError
	}

	listInv := []string{}
	for _, v := range so {
		listInv = append(listInv, v.InvNo)
	}
	var inv []m.Invoice
	if err := dbSale.Ctx().Model(&m.Invoice{}).Where(`invoice_no IN (?)`, listInv).Preload("InvStatus").Find(&inv).Error; err != nil {
		log.Errorln(pkgName, err, "select data invoice error -:")
	}
	hasBill := 0
	hadBill := 0
	nonBill := 0
	var listInvBilling []string
	for _, v := range inv {
		if len(v.InvStatus) != 0 {
			check := false
			for _, i := range v.InvStatus {
				if strings.TrimSpace(i.InvoiceStatusName) == "วางบิลแล้ว" {
					check = true
				} else {
					if !check {
						check = false
					}
				}
			}
			if check {
				listInvBilling = append(listInvBilling, v.InvoiceNo)
				hadBill += 1
			} else {
				hasBill += 1
			}
		} else {
			nonBill += 1
		}
	}

	hasErr := 0
	var sum []SOCus
	var soTotal []TrackInvoice
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {

		sqlSum := `		SELECT *,
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
					) as so_amount,
					sum(PeriodAmount) as amount
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) 
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
					;`

		if err := dbSale.Ctx().Raw(sqlSum, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, InvNumber, StaffId).Scan(&sum).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}

		wg.Done()
	}()
	go func() {
		sql := `SELECT
		sum(amount) as amount, 
		AVG(in_factor) as in_factor,
		AVG(ex_factor) as ex_factor,
		SUM(so_amount) as pro_rate,
		(sum(amount)/sum(amount_engcost)) as sale_factor
		from (
			SELECT
				sum(TotalContractAmount) as amount,
				sum(eng_cost) as amount_engcost,
				sale_factor,
				in_factor,sale_code,sale_name,ex_factor,PeriodAmount,so_amount
				FROM (
					SELECT
						SDPropertyCS28,sonumber,ContractStartDate,ContractEndDate,BLSCDocNo,PeriodStartDate,PeriodEndDate,GetCN,INCSCDocNo,Customer_ID,Customer_Name,
						sale_code,sale_name,sale_team,PeriodAmount, sale_factor, in_factor, ex_factor,TotalContractAmount,
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
						LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' and BLSCDocNo <> ''
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) 

						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
					) sub_data
				) so_group
				WHERE so_amount <> 0 group by BLSCDocNo
			) cust_group`

		if err := dbSale.Ctx().Raw(sql, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateTo, dateTo, dateFrom, dateTo, dateFrom, dateFrom, dateFrom, dateFrom, dateFrom, dateTo, dateTo, dateFrom, dateTo, dateFrom, listInv, listId, search, StaffId, InvNumber).Scan(&soTotal).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
			hasErr += 1
		}
		wg.Done()
	}()
	go func() {

		sqlSum := `	SELECT distinct Customer_ID
	 			FROM so_mssql
				 LEFT JOIN staff_info ON so_mssql.sale_code = staff_info.staff_id
						WHERE Active_Inactive = 'Active' 
						and PeriodStartDate <= ? and PeriodEndDate >= ?
						and PeriodStartDate <= PeriodEndDate and BLSCDocNo IN (?) 
						and sale_code in (?)
						and INSTR(CONCAT_WS('|', BLSCDocNo, staff_id, fname,lname,nname,department,SOWebStatus,Customer_ID,Customer_Name,SOType,detail), ?)
						and INSTR(CONCAT_WS('|', BLSCDocNo), ?)
						and INSTR(CONCAT_WS('|', sale_code), ?)
						group by BLSCDocNo
					;`

		if err := dbSale.Ctx().Raw(sqlSum, dateTo, dateFrom, listInv, listId, search, InvNumber, StaffId).Scan(&cus).Error; err != nil {
			log.Errorln(pkgName, err, "select data error -:")
		}

		wg.Done()
	}()
	wg.Wait()

	// dataReceipt := map[string]interface{}{
	// 	"has_receipt":    len(sum),
	// 	"customer_total": len(cus),
	// 	"total_so":       soTotal,
	// 	"detail":         sum,
	// }
	// return c.JSON(http.StatusOK, dataReceipt)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "detailreceipt"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "BLSCDocNo")
	f.SetCellValue(mode, "B1", "Staff ID")
	f.SetCellValue(mode, "C1", "First Name")
	f.SetCellValue(mode, "D1", "Last Name")
	f.SetCellValue(mode, "E1", "Nick Name")
	f.SetCellValue(mode, "F1", "Department")
	f.SetCellValue(mode, "G1", "So Amount")
	f.SetCellValue(mode, "H1", "Amount")
	f.SetCellValue(mode, "I1", "SOWebStatus")
	f.SetCellValue(mode, "J1", "Customer ID")
	f.SetCellValue(mode, "K1", "Cusname Thai")
	f.SetCellValue(mode, "L1", "Business Type")
	f.SetCellValue(mode, "M1", "Job Status")
	f.SetCellValue(mode, "N1", "So Type")
	f.SetCellValue(mode, "O1", "Sale Factor")
	f.SetCellValue(mode, "P1", "In Factor")
	f.SetCellValue(mode, "Q1", "Ex Factor")
	f.SetCellValue(mode, "R1", "Start Date")
	f.SetCellValue(mode, "S1", "End Date")
	f.SetCellValue(mode, "T1", "Detail")

	colBLSCDocNo := "A"
	colStaffID := "B"
	colFirstName := "C"
	colLastName := "D"
	colNickName := "E"
	colDepartment := "F"
	colSoAmount := "G"
	colAmount := "H"
	colSOWebStatus := "I"
	colCustomerID := "J"
	colCusnameThai := "K"
	colBusinessType := "L"
	colJobStatus := "M"
	colSoType := "N"
	colSaleFactor := "O"
	colInFactor := "P"
	colExFactor := "Q"
	colStartDate := "R"
	colEndDate := "S"
	colDetail := "T"

	for k, v := range sum {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
		f.SetCellValue(mode, fmt.Sprint(colBLSCDocNo, k+2), v.BLSCDocNo)
		f.SetCellValue(mode, fmt.Sprint(colStaffID, k+2), v.StaffID)
		f.SetCellValue(mode, fmt.Sprint(colFirstName, k+2), v.Fname)
		f.SetCellValue(mode, fmt.Sprint(colLastName, k+2), v.Lname)
		f.SetCellValue(mode, fmt.Sprint(colNickName, k+2), v.Nname)
		f.SetCellValue(mode, fmt.Sprint(colDepartment, k+2), v.Department)
		f.SetCellValue(mode, fmt.Sprint(colSoAmount, k+2), v.SoAmount)
		f.SetCellValue(mode, fmt.Sprint(colAmount, k+2), v.Amount)
		f.SetCellValue(mode, fmt.Sprint(colSOWebStatus, k+2), v.SOWebStatus)
		f.SetCellValue(mode, fmt.Sprint(colCustomerID, k+2), v.CustomerID)
		f.SetCellValue(mode, fmt.Sprint(colCusnameThai, k+2), v.CusnameThai)
		f.SetCellValue(mode, fmt.Sprint(colBusinessType, k+2), v.BusinessType)
		f.SetCellValue(mode, fmt.Sprint(colJobStatus, k+2), v.JobStatus)
		f.SetCellValue(mode, fmt.Sprint(colSoType, k+2), v.SoType)
		f.SetCellValue(mode, fmt.Sprint(colSaleFactor, k+2), v.SaleFactor)
		f.SetCellValue(mode, fmt.Sprint(colInFactor, k+2), v.InFactor)
		f.SetCellValue(mode, fmt.Sprint(colExFactor, k+2), v.ExFactor)
		f.SetCellValue(mode, fmt.Sprint(colStartDate, k+2), v.StartDate)
		f.SetCellValue(mode, fmt.Sprint(colEndDate, k+2), v.EndDate)
		f.SetCellValue(mode, fmt.Sprint(colDetail, k+2), v.Detail)

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
