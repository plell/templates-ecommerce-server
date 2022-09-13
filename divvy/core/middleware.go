package core

import (
	"github.com/labstack/echo/v4"
)

func LogPathAndIp(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		m := c.Request().Method
		if m == "DELETE" || m == "PATCH" {
			ip := c.RealIP()
			LogInfo(ip + " requesting at " + c.Path())
		}
		return next(c)
	}
}
