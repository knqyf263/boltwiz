package routes

import (
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/knqyf263/boltwiz/ui"
)

func RegisterStaticRoutes(e *echo.Echo) {
	staticPages := []string{
		"css",
		"styles",
		"img",
		"js",
		"app",
		"maps",
		"ico",
		"fonts",
		"video",
		"icons",
	}

	sub, err := fs.Sub(ui.WebContent, "dist")
	if err != nil {
		return
	}

	ac := http.FS(sub)

	var contentHandler = echo.WrapHandler(http.FileServer(ac))
	// set the root route (serving index.html)
	e.GET("/", contentHandler)

	// configure static routes
	for _, path := range staticPages {
		e.GET("/"+path+"/*", contentHandler, middleware.Gzip())
	}
}
