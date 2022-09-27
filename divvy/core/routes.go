package core

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

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
	e.POST("/stripe/createAmountSessionByCustomer", CreateAmountCheckoutSessionByCustomer)
	e.POST("/stripe/createProductSessionByCustomer", CreateProductCheckoutSessionByCustomer)
	e.GET("/stripe/products", GetProductsFromStripe)
	e.POST("/email/subscribe", SendSubscriberEmail)
	e.POST("/email/contact", SendContactFormEmail)

	// client webhooks, userSelector required
	// u := e.Group("")
	// u.Use(UserExists)
	// u.Any("/ws/:userSelector", echo.HandlerFunc(WsEndpoint))

}
