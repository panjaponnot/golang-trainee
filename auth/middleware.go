package auth

import (
	"context"
	"fmt"
	"net/http"
	"sale_ranking/core"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Skipper middleware.Skipper
}

var DefaultConfig = Config{
	Skipper: middleware.DefaultSkipper,
}

func UserAuthMiddleware(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			token := GetTokenFromHeader(c, "Bearer", echo.HeaderAuthorization)
			if token == "" {
				return echo.ErrUnauthorized
			}
			sToken := strings.Split(token, ",")
			claims, err := core.DecodeAccessToken(sToken[1])
			if err != nil {
				return echo.ErrUnauthorized
			}
			key := fmt.Sprintf("%s:%s:%s", sessionKey, claims.Id, claims.Subject)
			var session CacheSession
			if err := redis.Get(key, &session); err != nil {
				return echo.ErrUnauthorized
			}
			// log.Infoln("uid", claims.Id)
			if session.AccountId == claims.Subject && session.AccessToken == sToken[0] && session.Uid.String() == claims.Id && session.Ip == c.RealIP() && session.Agent == c.Request().UserAgent() {
				// _ = redis.Set(key, session, sessionTimeOut)
				request := c.Request()
				authCtx := context.WithValue(request.Context(), authorizedContext, session)
				c.SetRequest(request.WithContext(authCtx))
				return next(c)
			}
			return echo.ErrUnauthorized
		}
	}
}

// client
func AuthMiddlewareWithConfig(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			token := GetTokenFromHeader(c, "Bearer", echo.HeaderAuthorization)
			if token == "" {
				return echo.ErrUnauthorized
			}
			claims, err := core.IsClientAccess(token)
			if err != nil {
				return echo.ErrUnauthorized
			}
			clientClaim := core.ClientClaims{Name: claims.Name}
			request := c.Request()
			authCtx := context.WithValue(request.Context(), authorizedContext, clientClaim)
			c.SetRequest(request.WithContext(authCtx))
			return next(c)
		}
	}
}

// TODO: User permission middleware
func UserPermissionMiddleware(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			return c.JSON(http.StatusForbidden, nil)
		}
	}
}

// middleware use //
func UseAuthMiddleware(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			claims, err := core.DecodeAccessToken(GetTokenFromHeader(c, "Bearer", echo.HeaderAuthorization))
			if err != nil {
				return echo.ErrUnauthorized
			}
			key := fmt.Sprintf("%s:%s", sessionKey, claims.Id)
			var session CacheSession
			if err := redis.Get(key, &session); err != nil {
				return echo.ErrUnauthorized
			} else {
				return next(c)
			}
			// if session.AccountId == claims.Subject && session.Ip == c.RealIP() && session.Agent == c.Request().UserAgent() {
			// 	_ = redis.Set(key, session, sessionTimeOut)
			// 	request := c.Request()
			// 	authCtx := context.WithValue(request.Context(), authorizedContext, session)
			// 	c.SetRequest(request.WithContext(authCtx))
			// 	return next(c)
			// }
			// return echo.ErrUnauthorized
		}
	}
}
