package app

import (
	"github.com/labstack/echo"
	"github.com/s-aska/line-works-debugger-go/webapp/app/controllers/calendar"
	"github.com/s-aska/line-works-debugger-go/webapp/app/controllers/top"
)

func registerRoutes(e *echo.Echo) {
	e.GET("/", top.Root)
	e.POST("/", top.Post)
	e.GET("/callback", top.Callback)
	e.GET("/calendar/", calendar.Events)
	e.POST("/callback", top.Callback)
}
