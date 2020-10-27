package auth

import (
	"fmt"
	"net/http"
	"sale_ranking/core"
	m "sale_ranking/model"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"
	"sale_ranking/pkg/util/crypto"
	"strings"
	"time"
	"unicode"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

var (
	excludeReCaptCha = []string{
		"127.0.0.1",
		"203.150.21.70",
		"203.150.21.74",
	}
)

func newLoginTicketEndpoint(c echo.Context) error {
	salt := strings.TrimSpace(c.QueryParam("salt"))
	runes := []func(rune) bool{
		unicode.IsUpper,
		unicode.IsLower,
		unicode.IsDigit,
	}
	if len(salt) != 32 || !crypto.ValidateString(salt, runes) {
		return echo.ErrBadRequest
	}
	id := uuid.NewV4()
	key := fmt.Sprintf("%s:%s", ticketKey, id.String())
	ticket := CacheTicket{
		Id:       id,
		Salt:     salt,
		ClientIp: c.RealIP(),
		Agent:    c.Request().UserAgent(),
	}
	if err := redis.Set(key, ticket, time.Minute); err != nil {
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, server.Result{Data: map[string]interface{}{
		"ticket":     id,
		"expires_at": time.Now().Add(time.Minute),
	}})
}

func submitLoginEndpoint(c echo.Context) error {
	data := struct {
		Secret         string    `json:"secret"`
		Ticket         uuid.UUID `json:"ticket"`
		ReCaptChaToken string    `json:"token"`
	}{}
	if err := c.Bind(&data); err != nil || data.Ticket == uuid.Nil {
		return echo.ErrBadRequest
	}
	var ticket CacheTicket
	key := fmt.Sprintf("%s:%s", ticketKey, data.Ticket)
	if err := redis.Get(key, &ticket); err != nil || ticket.ClientIp != c.RealIP() || ticket.Agent != c.Request().UserAgent() {
		return echo.ErrForbidden
	}
	_ = redis.Del(key)
	authCode, err := core.AESDecryptCipherByKey([]byte(data.Secret), ticket.Salt)
	if err != nil {
		log.Errorln(pkgName, err, "Decrypt secret error")
		return echo.ErrBadRequest
	}
	if !util.Contains(excludeReCaptCha, c.RealIP()) {
		if err := core.ReCaptCha().VerifyWithOptions(data.ReCaptChaToken, recaptcha.VerifyOption{Action: "SSOAuth", Threshold: core.ReCaptChaTrustScore()}); err != nil {
			log.Errorln(pkgName, err, "reCaptCah verify error for SSO from", c.RealIP(), c.Request().UserAgent())
			return echo.ErrForbidden
		}
	}
	account, err := core.IdentityClient().VerifyAuthorizedCode(string(authCode))
	if err != nil {
		log.Errorln(pkgName, err, "One Identity service not available")
		return c.JSON(http.StatusServiceUnavailable, server.Result{Error: "one id service not available"})
	}

	// Check attendant employee account
	attendantProfile, err := core.AttendantClient().GetAccountByID(account.AccountID)
	if err != nil {
		log.Errorln(pkgName, err, "Get account to attendant client error")
		return c.JSON(http.StatusUnauthorized, server.Result{Error: "Unauthorized access restrict"})
	}

	var user m.User
	var permissions []m.Permission
	if err := db.Ctx().Model(&m.User{}).Where(m.User{AccountId: account.AccountID}).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			user = m.User{
				AccountId: attendantProfile.AccountID,
				Status:    "",
			}
			if err := db.Ctx().Model(&m.User{}).Create(&user).Error; err != nil {
				log.Errorln(pkgName, err, "Register new account error")
				return c.JSON(http.StatusInternalServerError, server.Result{Error: "register new account error"})
			}
		} else {
			log.Errorln(pkgName, err, "Get login user error")
			return c.JSON(http.StatusInternalServerError, server.Result{Error: "server error"})
		}
	}

	// Select permission
	if err := db.Ctx().Model(&m.Permission{}).Where(m.Permission{AccountId: account.AccountID}).Preload("Company").Preload("Role").Find(&permissions).Error; err != nil {
		log.Errorln(pkgName, err, "Get user permission error")
		return c.JSON(http.StatusInternalServerError, server.Result{Error: "get user permission error"})
	}

	// Write session cache
	key = fmt.Sprintf("%s:%s", sessionKey, user.Uid)
	session := CacheSession{
		UserUid:     user.Uid,
		AccountId:   account.AccountID,
		Ip:          c.RealIP(),
		Agent:       c.Request().UserAgent(),
		Profile:     attendantProfile,
		Permissions: permissions,
	}
	_ = redis.Set(key, &session, sessionTimeOut)
	token, err := core.EncodeAccessToken(user.Uid, account.AccountID, nil)
	if err != nil {
		log.Errorln(pkgName, err, "Encode access token error")
		return c.JSON(http.StatusInternalServerError, server.Result{Error: "Generate session token error"})
	}
	resourceKey, _ := core.EncryptWithServerKey([]byte(fmt.Sprintf("%s|%s", user.Uid, account.AccountID)))
	resultData := map[string]interface{}{
		"resource_key": resourceKey,
		"token":        token,
		"type":         "Bearer",
		"profile":      attendantProfile,
		"permissions":  permissions,
	}

	switch user.Status {
	case StatusActivated:
		// activated account
		return c.JSON(http.StatusOK, server.Result{Data: resultData})
	case StatusWaiting:
		// requested to access
		return c.JSON(http.StatusForbidden, server.Result{Error: "waiting account", Data: resultData})
	case StatusSuspended:
		// suspended by admin
		return c.JSON(http.StatusForbidden, server.Result{Error: "suspended account", Data: resultData})
	}
	// valid but no action
	return c.JSON(http.StatusForbidden, server.Result{Error: "valid account", Data: resultData})
}

func isLoggedInEndpoint(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func getLogInByOneIdEndpoint(c echo.Context) error {
	return c.JSON(http.StatusOK, server.Result{Data: core.IdentityClient().GetLoginUrl()})
}
