package app

import (
	"net/http"
	"os"

	"github.com/labstack/echo"
	"google.golang.org/appengine"
)

// Bootstrap ...
func Bootstrap() {
	e := echo.New()

	// e.Logger.SetLevel(log.DEBUG)

	e.Use(authorizer)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		// log.Errorf("error:%v", err)                         // 標準出力へ
		c.JSON(http.StatusInternalServerError, err.Error()) // ブラウザ画面へ
	}
	registerTemplates(e)
	registerRoutes(e)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		e.Logger.Infof("Defaulting to port %s", port)
	}
	e.Logger.Infof("Listening on port %s", port)

	http.Handle("/", e)

	appengine.Main()

	e.Logger.Fatal(e.Start(":" + port))
}
