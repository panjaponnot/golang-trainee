package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sale_ranking/core"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/requests"
	"sale_ranking/pkg/util"
	"strings"

	"github.com/labstack/echo/v4"
)

func GetUserOneThEndPoint(c echo.Context) error {
	token := c.QueryParam("token")

	if strings.TrimSpace(c.QueryParam("token")) == "" {
		return echo.ErrBadRequest
	}

	url := "https://chat-manage.one.th:8997/api/v1/getprofile"
	headers := map[string]string{
		echo.HeaderContentType:   "application/json",
		echo.HeaderAuthorization: fmt.Sprintf("Bearer %s", util.GetEnv("WEB_HOOK_CHAT_TOKEN", "")),
	}
	body, _ := json.Marshal(&struct {
		BotId  string `json:"bot_id"`
		Source string `json:"source"`
		Phone  string `json:"phone"`
	}{
		BotId:  util.GetEnv("WEB_HOOK_CHAT_BOT_ID", ""),
		Source: token,
		Phone:  "true",
	})
	r, err := requests.Request(http.MethodPost, url, headers, bytes.NewBuffer(body), 0)
	if err != nil {
		log.Errorln(pkgName, err, "service chat unavailable")
		return echo.ErrServiceUnavailable
	}
	type Raw struct {
		OneId           string `json:"one_id"`
		EmployeeeDetail string `josn:"employee_detail"`
	}
	data := struct {
		Data   Raw    `json:"data"`
		Status string `json:"status"`
	}{}
	dataByte, err := json.Marshal(r)
	if err := json.Unmarshal(dataByte, &data); err != nil {
		log.Errorln(pkgName, err, "Json unmarshall chat error")
		return echo.ErrInternalServerError
	}

	if data.Status == "success" {
		a := core.AttendantClient()
		d, _ := a.GetAccountByID(data.Data.OneId)
		// log.Infoln(d)
		fmt.Println(d)
	}
	// a := core.AttendantClient()
	// d, _ := a.GetAccountByID("aaa")
	return c.JSON(http.StatusOK, nil)
}
