package auth

import (
	"errors"
	"fmt"
	"net/http"
	"sale_ranking/core"
	"sale_ranking/pkg/attendant"
	"sale_ranking/pkg/cache"
	"sale_ranking/pkg/database"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/server"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
)

type CacheSession struct {
	Uid          uuid.UUID                `json:"user_uid"`
	AccountId    string                   `json:"account_id"`
	Ip           string                   `json:"ip"`
	Agent        string                   `json:"agent"`
	RefreshToken string                   `json:"refresh_token"`
	AccessToken  string                   `json:"access_token"`
	TokenType    string                   `json:"token_type"`
	Profile      attendant.AccountProfile `json:"profile"`
	Role         string                   `json:"role"`
	SubRole      string                   `json:"sub_role"`
	Username     string                   `json:"username"`
	GenToken     string                   `json:"gen_token"`
}

type CacheTicket struct {
	Id       uuid.UUID `json:"id"`
	ClientIp string    `json:"client_ip"`
	Agent    string    `json:"agent"`
	Salt     string    `json:"salt"`
}

const (
	pkgName = "AUTH"

	sessionKey = "session"
	ticketKey  = "ticket"

	authorizedContext = "auth"
	sessionTimeOut    = 24 * time.Hour

	// User status
	StatusNone      = ""          // valid but no any requests to any companies
	StatusWaiting   = "waiting"   // valid and requested to access some company
	StatusActivated = "activated" // valid and success login
	StatusSuspended = "suspended" // valid but suspend for something cause
)

var (
	db    database.Database
	redis cache.Redis
)

func InitApiRouter(g *echo.Group) error {
	if err := initDataStore(); err != nil {
		return err
	}
	// Skipper
	g.Use(UserAuthMiddleware(Config{Skipper: func(c echo.Context) bool {
		skipper := server.NewSkipperPath("")
		skipper.Add("/api/v2/auth/ticket", http.MethodGet)
		skipper.Add("/api/v2/auth/login", http.MethodPost)
		return skipper.Test(c)
	}}))
	// Router
	g.GET("", isLoggedInEndpoint)
	g.GET("/ticket", newLoginTicketEndpoint)
	g.POST("/login", submitLoginEndpoint)
	g.GET("/login", getLogInByOneIdEndpoint)
	g.GET("/logout", logOutEndpoint)
	g.GET("/test", getStateEndPoint)

	return nil
}

func initDataStore() error {
	// Database
	db = core.NewDatabase(pkgName, "salerank")
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

func GetAuthorizedContext(c echo.Context) (CacheSession, error) {
	session, ok := c.Request().Context().Value(authorizedContext).(*CacheSession)
	if !ok {
		return CacheSession{}, errors.New("unauthorized session")
	}
	return *session, nil
}

// Extract token from http header
func GetTokenFromHeader(c echo.Context, tokenType string, header string) string {
	token := c.Request().Header.Get(header)
	token = strings.TrimSpace(token)
	tokenType += " "
	if token == "" || len(token) < (len(tokenType)+1) || strings.ToLower(token[:len(tokenType)]) != strings.ToLower(tokenType) {
		return ""
	}
	token = strings.TrimSpace(token[len(tokenType):])
	return token
}

// Verify resource key
func VerifyResourceKey(resourceKey string, c echo.Context) (CacheSession, error) {
	resourceByte, err := core.DecryptWithServerKey(c.QueryParam("resource_key"))
	if err != nil {
		log.Errorln(pkgName, err, "Decrypt resource key error")
		return CacheSession{}, err
	}
	sessionCacheKey := fmt.Sprintf("%s:%s", sessionKey, strings.Split(string(resourceByte), "|")[0])
	var session CacheSession
	if err := redis.Get(sessionCacheKey, &session); err != nil || session.AccountId != strings.Split(string(resourceByte), "|")[1] {
		return CacheSession{}, err
	}
	return session, nil
}
