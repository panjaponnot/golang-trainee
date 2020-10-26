package auth

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func logOutEndpoint(c echo.Context) error {
	session, err := GetAuthorizedContext(c)
	if err != nil {
		return echo.ErrUnauthorized
	}
	key := fmt.Sprintf("%s:%s", sessionKey, session.UserUid)
	_ = redis.Del(key)
	return c.JSON(http.StatusNoContent, nil)
}
