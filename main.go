package main

import (
	"net/http"

	"github.com/casbin/casbin/v2"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	eg, _ := casbin.NewEnforcer("auth_model.conf", "auth_policy.csv")

	e.Use(Middleware(eg))

	e.Any("*", func(c echo.Context) error {
		return c.String(http.StatusOK, http.StatusText(http.StatusOK))
	})

	e.Logger.Fatal(e.Start(":1323"))
}

type (
	// Config defines the config for CasbinAuth middleware.
	Config struct {
		// Enforcer CasbinAuth main rule.
		// Required.
		Enforcer *casbin.Enforcer
	}
)

func Middleware(ce *casbin.Enforcer) echo.MiddlewareFunc {
	config := Config{
		Enforcer: ce,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if pass, err := config.CheckPermission(c); err == nil && pass {
				return next(c)
			} else if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}

			return echo.ErrForbidden
		}
	}
}

// GetUserName gets the user name from the request.
// Currently, only HTTP basic authentication is supported
func (a *Config) GetUserName(c echo.Context) string {
	username, _, _ := c.Request().BasicAuth()
	return username
}

// CheckPermission checks the user/path/method combination from the request.
// Returns true (permission granted) or false (permission forbidden)
func (a *Config) CheckPermission(c echo.Context) (bool, error) {
	user := a.GetUserName(c)
	path := c.Request().URL.Path
	method := c.Request().Method
	return a.Enforcer.Enforce(user, path, method)
}
