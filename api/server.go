package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
)

var (
	once sync.Once
	e    *echo.Echo
)

func getEcho() *echo.Echo {
	once.Do(func() {
		e = echo.New()
		e.HideBanner = true
		registerStats(e)
		registerTopLangs(e)
	})
	return e
}

// Handler allows platforms expecting an http.Handler-like entrypoint to serve via Echo
func Handler(w http.ResponseWriter, r *http.Request) {
	getEcho().ServeHTTP(w, r)
}

func main() {
	if os.Getenv("GO_LOCAL_SERVER") == "1" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
		_ = getEcho().Start(":" + port)
	}
}
