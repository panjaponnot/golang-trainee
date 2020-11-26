package auth

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func logOutEndpoint(c echo.Context) error {
	session, err := GetAuthorizedContext(c)
	if err != nil {
		return echo.ErrUnauthorized
	}
	key := fmt.Sprintf("%s:%s", sessionKey, session.Uid)
	_ = redis.Del(key)
	return c.JSON(http.StatusNoContent, nil)
}
