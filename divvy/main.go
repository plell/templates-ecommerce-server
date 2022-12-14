package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"

	core "github.com/plell/divvygo/divvy/core"
)

func main() {

	// Echo instance
	e := echo.New()

	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	// Make Routes
	core.MakeRoutes(e)

	// DB connect
	// core.ConnectDB()
	// DB Automigrate
	// core.MigrateUp()

	// do token refresh on start
	core.GoogleRefreshTokenIfExists()
	// cron stuff
	c := cron.New()
	c.AddFunc("@every 30m", func() {
		// test refresh tokens expire after 7d, tokens expire after 60m
		core.GoogleRefreshTokenIfExists()
	})
	c.Start()

	// client webhooks
	// go core.RunWebsocketBroker()

	// logger
	// core.StartDNALogger()

	// Start server
	fmt.Println("start http 8000 server!")
	e.Logger.Fatal(e.Start(":8000"))

}
