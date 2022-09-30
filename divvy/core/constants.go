package core

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var PAYMENT_SUCCESS_URL = "https://www.lazycowbakeryseattle.com/success"
var PAYMENT_CANCEL_URL = "https://www.lazycowbakeryseattle.com/"
var STORE_NAME = "Lazy Cow Bakery"
var STORE_DOMAIN_NAME = "lazycowbakeryseattle.com"
var STORE_PRODUCT_NAME = "Custom Cake"
var STORE_EMAIL = "plelldavid@gmail.com"

func getGoogleKeys() (string, string) {
	clientId := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	return clientId, clientSecret
}

func SetupConfig() *oauth2.Config {
	clientId, clientSecret := getGoogleKeys()
	conf := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8000/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Endpoint: google.Endpoint,
	}

	return conf
}

var RECIEPT_EMAIL_HTML = `<div>
Product Checkout created!
<div style="margin:80px">
MARGIN BABY
</div>
</div>`

var ORDER_EMAIL_HTML = `<div>
Product Checkout created!
<div style="margin:80px">
MARGIN BABY
</div>
</div>`
