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
		PayTypeChange    	string  `json:"pay_type_change"`
		SoTypeChange      string  `json:"so_type_change"`
		Reason            string  `json:"reason"`
		Remark            string  `json:"remark"`
	}{}

	if err := dbSale.Ctx().Raw(`
	SELECT Active_Inactive,has_refer,tb_ch_so.sonumber,Customer_ID,Customer_Name,DATE_FORMAT(ContractStartDate, '%Y-%m-%d') as ContractStartDate,
	DATE_FORMAT(ContractEndDate, '%Y-%m-%d') as ContractEndDate,so_refer,sale_code,sale_lead,DATEDIFF(ContractEndDate, NOW()) as days,
	month(ContractEndDate) as so_month, SOWebStatus,pricesale,PeriodAmount, SUM(PeriodAmount) as TotalAmount,staff_id,prefix,fname,lname,nname,position,
	department,so_type_change,pay_type_change,
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
	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
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
	if len(dataRaw) > (page * 20) {
		start := (page - 1) * 20
		end := (page * 20)
		dataResult.Detail = map[string]interface{}{
			"data":  dataRaw[start:end],
			"count": len(dataRaw[start:end]),
		}
		dataRawRes = dataRaw[start:end]
	} else {
		start := (page * 20) - (20)
		dataResult.Detail = map[string]interface{}{
			"data":  dataRaw[start:],
			"count": len(dataRaw[start:]),
		}
		dataRawRes = dataRaw[start:]
	}
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
	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
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
	for _, r := range report {
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

		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, r)
				}
			}
		}
	}

	if len(dataResult) > 1 {
		sort.SliceStable(dataResult, func(i, j int) bool { return dataResult[i].ScoreAll > dataResult[j].ScoreAll })
	}

	var dataResultEx []m.OrgChart
	var result m.Result
	if len(dataResult) > (page * 10) {
		start := (page - 1) * 10
		end := (page * 10)
		result.Data = dataResult[start:end]
		result.Count = len(dataResult[start:end])
		dataResultEx = dataResult[start:end]
	} else {
		start := (page * 10) - (10)
		result.Data = dataResult[start:]
		result.Count = len(dataResult[start:])
		dataResultEx = dataResult[start:]
	}
	result.Total = len(dataResult)
	// return c.JSON(http.StatusOK, result)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "rankingbase"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "StaffId")
	f.SetCellValue(mode, "B1", "First Name")
	f.SetCellValue(mode, "C1", "Last Name")
	f.SetCellValue(mode, "D1", "Nick Name")
	f.SetCellValue(mode, "E1", "Position")
	f.SetCellValue(mode, "F1", "Department")
	f.SetCellValue(mode, "G1", "Staff Child")
	f.SetCellValue(mode, "H1", "Inv Amount")
	f.SetCellValue(mode, "I1", "Inv Amount Old")
	f.SetCellValue(mode, "J1", "Goal Total")
	f.SetCellValue(mode, "K1", "Score Target")
	f.SetCellValue(mode, "L1", "Score Sf")
	f.SetCellValue(mode, "M1", "Sale Factor")
	f.SetCellValue(mode, "N1", "Total So")
	f.SetCellValue(mode, "O1", "If Factor")
	f.SetCellValue(mode, "P1", "Eng Cost")
	f.SetCellValue(mode, "Q1", "Revenue")
	f.SetCellValue(mode, "R1", "Score If")
	f.SetCellValue(mode, "S1", "In Factor")
	f.SetCellValue(mode, "T1", "One Id")
	f.SetCellValue(mode, "U1", "Image")
	f.SetCellValue(mode, "V1", "File Name")
	f.SetCellValue(mode, "W1", "Growth Rate")
	f.SetCellValue(mode, "X1", "Score Growth")
	f.SetCellValue(mode, "Y1", "Score All")
	f.SetCellValue(mode, "Z1", "Quarter")
	f.SetCellValue(mode, "AA1", "Year")
	f.SetCellValue(mode, "AB1", "Job Months")
	f.SetCellValue(mode, "AC1", "Commission")

	colStaffId := "A"
	colFname := "B"
	colLname := "C"
	colNname := "D"
	colPosition := "E"
	colDepartment := "F"
	colStaffChild := "G"
	colInvAmount := "H"
	colInvAmountOld := "I"
	colGoalTotal := "J"
	colScoreTarget := "K"
	colScoreSf := "L"
	colSaleFactor := "M"
	colTotalSo := "N"
	colIfFactor := "O"
	colEngCost := "P"
	colRevenue := "Q"
	colScoreIf := "R"
	colInFactor := "S"
	colOneId := "T"
	colImage := "U"
	colFileName := "V"
	colGrowthRate := "W"
	colScoreGrowth := "X"
	colScoreAll := "Y"
	colQuarter := "Z"
	colYear := "AA"
	colJobMonths := "AB"
	colCommission := "AC"

	for k, v := range dataResultEx {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
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

	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
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
	for _, r := range report {
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
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, r)
				}
			}
		}
	}
	if len(dataResult) > 1 {
		sort.SliceStable(dataResult, func(i, j int) bool { return dataResult[i].ScoreAll > dataResult[j].ScoreAll })
	}

	var dataResultEx []m.OrgChart
	var result m.Result
	if len(dataResult) > (page * 10) {
		start := (page - 1) * 10
		end := (page * 10)
		result.Data = dataResult[start:end]
		result.Count = len(dataResult[start:end])
		dataResultEx = dataResult[start:end]
	} else {
		start := (page * 10) - (10)
		result.Data = dataResult[start:]
		result.Count = len(dataResult[start:])
		dataResultEx = dataResult[start:]
	}
	result.Total = len(dataResult)
	// return c.JSON(http.StatusOK, result)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "rankingkey"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "StaffId")
	f.SetCellValue(mode, "B1", "First Name")
	f.SetCellValue(mode, "C1", "Last Name")
	f.SetCellValue(mode, "D1", "Nick Name")
	f.SetCellValue(mode, "E1", "Position")
	f.SetCellValue(mode, "F1", "Department")
	f.SetCellValue(mode, "G1", "Staff Child")
	f.SetCellValue(mode, "H1", "Inv Amount")
	f.SetCellValue(mode, "I1", "Inv Amount Old")
	f.SetCellValue(mode, "J1", "Goal Total")
	f.SetCellValue(mode, "K1", "Score Target")
	f.SetCellValue(mode, "L1", "Score Sf")
	f.SetCellValue(mode, "M1", "Sale Factor")
	f.SetCellValue(mode, "N1", "Total So")
	f.SetCellValue(mode, "O1", "If Factor")
	f.SetCellValue(mode, "P1", "Eng Cost")
	f.SetCellValue(mode, "Q1", "Revenue")
	f.SetCellValue(mode, "R1", "Score If")
	f.SetCellValue(mode, "S1", "In Factor")
	f.SetCellValue(mode, "T1", "One Id")
	f.SetCellValue(mode, "U1", "Image")
	f.SetCellValue(mode, "V1", "File Name")
	f.SetCellValue(mode, "W1", "Growth Rate")
	f.SetCellValue(mode, "X1", "Score Growth")
	f.SetCellValue(mode, "Y1", "Score All")
	f.SetCellValue(mode, "Z1", "Quarter")
	f.SetCellValue(mode, "AA1", "Year")
	f.SetCellValue(mode, "AB1", "Job Months")
	f.SetCellValue(mode, "AC1", "Commission")

	colStaffId := "A"
	colFname := "B"
	colLname := "C"
	colNname := "D"
	colPosition := "E"
	colDepartment := "F"
	colStaffChild := "G"
	colInvAmount := "H"
	colInvAmountOld := "I"
	colGoalTotal := "J"
	colScoreTarget := "K"
	colScoreSf := "L"
	colSaleFactor := "M"
	colTotalSo := "N"
	colIfFactor := "O"
	colEngCost := "P"
	colRevenue := "Q"
	colScoreIf := "R"
	colInFactor := "S"
	colOneId := "T"
	colImage := "U"
	colFileName := "V"
	colGrowthRate := "W"
	colScoreGrowth := "X"
	colScoreAll := "Y"
	colQuarter := "Z"
	colYear := "AA"
	colJobMonths := "AB"
	colCommission := "AC"

	for k, v := range dataResultEx {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
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
	page, _ := strconv.Atoi(strings.TrimSpace(c.QueryParam("page")))
	if strings.TrimSpace(c.QueryParam("page")) == "" {
		page = 1
	}
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
	for _, r := range report {
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
		if len(staffInfo) != 0 {
			for _, st := range staffInfo {
				if st.StaffId == r.StaffId {
					dataResult = append(dataResult, r)
				}
			}
		}
	}
	if len(dataResult) != 0 {
		sort.SliceStable(dataResult, func(i, j int) bool { return dataResult[i].ScoreAll > dataResult[j].ScoreAll })
	}

	var dataResultEx []m.OrgChart
	var result m.Result
	if len(dataResult) > (page * 10) {
		start := (page - 1) * 10
		end := (page * 10)
		result.Data = dataResult[start:end]
		result.Count = len(dataResult[start:end])
		dataResultEx = dataResult[start:end]
	} else {
		start := (page * 10) - (10)
		result.Data = dataResult[start:]
		result.Count = len(dataResult[start:])
		dataResultEx = dataResult[start:]
	}
	result.Total = len(dataResult)
	// return c.JSON(http.StatusOK, result)

	log.Infoln(pkgName, "====  create excel ====")
	f := excelize.NewFile()
	// Create a new sheet.
	mode := "rankingrecovery"
	index := f.NewSheet(mode)
	// Set value of a cell.

	f.SetCellValue(mode, "A1", "StaffId")
	f.SetCellValue(mode, "B1", "First Name")
	f.SetCellValue(mode, "C1", "Last Name")
	f.SetCellValue(mode, "D1", "Nick Name")
	f.SetCellValue(mode, "E1", "Position")
	f.SetCellValue(mode, "F1", "Department")
	f.SetCellValue(mode, "G1", "Staff Child")
	f.SetCellValue(mode, "H1", "Inv Amount")
	f.SetCellValue(mode, "I1", "Inv Amount Old")
	f.SetCellValue(mode, "J1", "Goal Total")
	f.SetCellValue(mode, "K1", "Score Target")
	f.SetCellValue(mode, "L1", "Score Sf")
	f.SetCellValue(mode, "M1", "Sale Factor")
	f.SetCellValue(mode, "N1", "Total So")
	f.SetCellValue(mode, "O1", "If Factor")
	f.SetCellValue(mode, "P1", "Eng Cost")
	f.SetCellValue(mode, "Q1", "Revenue")
	f.SetCellValue(mode, "R1", "Score If")
	f.SetCellValue(mode, "S1", "In Factor")
	f.SetCellValue(mode, "T1", "One Id")
	f.SetCellValue(mode, "U1", "Image")
	f.SetCellValue(mode, "V1", "File Name")
	f.SetCellValue(mode, "W1", "Growth Rate")
	f.SetCellValue(mode, "X1", "Score Growth")
	f.SetCellValue(mode, "Y1", "Score All")
	f.SetCellValue(mode, "Z1", "Quarter")
	f.SetCellValue(mode, "AA1", "Year")
	f.SetCellValue(mode, "AB1", "Job Months")
	f.SetCellValue(mode, "AC1", "Commission")

	colStaffId := "A"
	colFname := "B"
	colLname := "C"
	colNname := "D"
	colPosition := "E"
	colDepartment := "F"
	colStaffChild := "G"
	colInvAmount := "H"
	colInvAmountOld := "I"
	colGoalTotal := "J"
	colScoreTarget := "K"
	colScoreSf := "L"
	colSaleFactor := "M"
	colTotalSo := "N"
	colIfFactor := "O"
	colEngCost := "P"
	colRevenue := "Q"
	colScoreIf := "R"
	colInFactor := "S"
	colOneId := "T"
	colImage := "U"
	colFileName := "V"
	colGrowthRate := "W"
	colScoreGrowth := "X"
	colScoreAll := "Y"
	colQuarter := "Z"
	colYear := "AA"
	colJobMonths := "AB"
	colCommission := "AC"

	for k, v := range dataResultEx {
		// log.Infoln(pkgName, "====>", fmt.Sprint(colSaleId, k+2))
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
