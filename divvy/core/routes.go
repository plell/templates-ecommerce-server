package core

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// var SocketServer *socketio.Server

func MakeRoutes(e *echo.Echo) {

	e.Use(LogPathAndIp)
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
			<h1>Welcome to Echo!</h1>
			<h3>TLS certificates automatically installed from Let's Encrypt :)</h3>
		`)
	})

	// stripe webhook listener
	e.Any("/webhook", echo.HandlerFunc(HandleStripeWebhook))

	// token not required group
	e.POST("/stripe/checkoutLink", CreateCheckoutLink)
	e.POST("/stripe/createSessionByCustomer", CreateCheckoutSessionByCustomer)

	// client webhooks, userSelector required
	// u := e.Group("")
	// u.Use(UserExists)
	// u.Any("/ws/:userSelector", echo.HandlerFunc(WsEndpoint))

}
