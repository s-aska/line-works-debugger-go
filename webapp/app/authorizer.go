package app

import (
	"github.com/labstack/echo"
)

func authorizer(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// r := c.Request()
		// ctx := r.Context()
		// userID := perm.Username(ctx, r)
		//
		// if !strings.HasPrefix(r.URL.Path, "/signin/") && userID == "" {
		// 	return c.Redirect(http.StatusFound, "/signin/azure/")
		// }

		return next(c)
	}
}
