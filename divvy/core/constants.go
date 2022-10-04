package core

import (
	"os"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var IS_DEV = os.Getenv("IS_DEV")

var PAYMENT_SUCCESS_URL = "https://www.lazycowbakeryseattle.com/success"
var PAYMENT_CANCEL_URL = "https://www.lazycowbakeryseattle.com/"
var STORE_NAME = "Lazy Cow Bakery"
var STORE_DOMAIN_NAME = "lazycowbakeryseattle.com"
var STORE_PRODUCT_NAME = "Custom Cake"
var STORE_EMAIL = "plelldavid@gmail.com" // "lazycowbakery@gmail.com"
var STORE_VENDOR_ID = "ZALdmzb6swk0wt"
var AUTH_TOKEN_PATH = "tokens/google-token-" + STORE_VENDOR_ID + ".json"

func getGoogleKeys() (string, string) {
	clientId := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	return clientId, clientSecret
}

func SetupConfig() *oauth2.Config {

	// log.Println("VENDOR ID IS ", vendorId)
	clientId, clientSecret := getGoogleKeys()
	conf := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8000/google/callback", //"https://api.plellworks.com/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Endpoint: google.Endpoint,
	}

	return conf
}

func MakeReceiptEmailTemplate(data SessionOrderData) string {
	template := ""
	customerName := ""
	date := ""
	orderMeta := makeOrderMetaFromMetadata(data.MetaData)

	if _, ok := data.MetaData["pickup_date"]; ok {
		date = data.MetaData["pickup_date"]
	}

	if _, ok := data.MetaData["first_name"]; ok {
		customerName += data.MetaData["first_name"]
	}

	if _, ok := data.MetaData["last_name"]; ok {
		customerName += " " + data.MetaData["last_name"]
	}

	template += `<div
style="
  padding: 50px;"
>
<div style="font-size: 48px; line-height: 72px; margin-bottom:40px" >
  Thank you for your order!
</div>

<img src="https://ci4.googleusercontent.com/proxy/xftfRJ8I_WvHOC42Ik0n_0ttaKmgsK-mQoX_a_IakOfuoemqHguusNh2RxnNnjhrWhmCGSzGSwsuiK-COR107PQVS58JOaOY4q9Gz0XSXoW-eF5yA8Gpanooe-xcXYKs0js_hjfh-0Q-E0rkBv7yPtYQAmQxNTo8WnYcU-HgqIBstCZmGiOHERZR3ejWpSXVVBIK4UcSvco_kV1wh8IP0wvquDfJQRuCOXOsVNlc4vhOivJJ=s0-d-e1-ft#https://static.wixstatic.com/media/d28008_0e517d22a8db47f6baeb27e2c71032d5~mv2.jpg/v1/fit/w_680,h_2000,al_c,q_85/d28008_0e517d22a8db47f6baeb27e2c71032d5~mv2.jpg" width="680" style="display:block;max-width:100%;vertical-align:middle;width:auto" alt="" class="CToWUd a6T" data-bit="iit" tabindex="0">

<div style="margin:50px 0;">
<h1 style="margin-bottom:30px;">Order Details</h1>
  <ul style="margin-bottom: 30px;">`

	// do cake
	for k, v := range orderMeta {
		template += "<li>" + k + ": " + v + "</li>"
	}

	// do shop
	for _, v := range data.LineItems {
		template += "<li>" + v.Name + " (" + strconv.Itoa(int(v.StripeLineItem.Quantity)) + ")</li>"
	}

	template += `
  </ul>

<div>
  Please pick up your order on ` + date + ` between 1-4pm at Lazy Cow Bakery 3418 Fremont
  Ave N Seattle, WA 98103
  </div>
  </div>` + TEMPLATE_FOOTER

	return template
}

func MakeNewOrderEmailTemplate(data SessionOrderData) string {
	template := ""
	customerName := ""
	email := ""
	date := ""
	orderMeta := makeOrderMetaFromMetadata(data.MetaData)

	if _, ok := data.MetaData["pickup_date"]; ok {
		date = data.MetaData["pickup_date"]
	}

	if _, ok := data.MetaData["first_name"]; ok {
		customerName += data.MetaData["first_name"]
	}

	if _, ok := data.MetaData["last_name"]; ok {
		customerName += " " + data.MetaData["last_name"]
	}

	if _, ok := data.MetaData["email"]; ok {
		email = data.MetaData["email"]
	}

	template += `<div
	style="
	  padding: 50px;
	>
	<div>
	<h1 style="margin-bottom:30px;">Order for ` + customerName + ` (` + email + `)</h1>
	
	  <ul style="margin-bottom: 30px;">`

	// do cake
	for k, v := range orderMeta {
		template += "<li>" + k + ": " + v + "</li>"
	}

	// do shop
	for _, v := range data.LineItems {
		template += "<li>" + v.Name + " (" + strconv.Itoa(int(v.StripeLineItem.Quantity)) + ")</li>"
	}

	template += `</ul> 
	
	<div style="margin:20px 0">
	Pickup date is ` + date + `
	</div>
	` + TEMPLATE_FOOTER

	return template
}

var TEMPLATE_FOOTER = `
<img src="https://ci3.googleusercontent.com/proxy/rlDOKoI1LxmxbVHip-jv1aiKyoEIZrnXzkoypnSRkcR1SzhJLcnOoQDgr0nA4ae5fgX8LNim7ArsdlRQLTnAr8tqD6hS4Kq7kKbyPo2_4T7DWojB-MwNIanZVuiYMCEueYXhJXbEKppagMR5y28uw0Y04pw98HxmZZxTFSKPqlMnmIs5tEp4v3AxvsqUdH8F19hGECRDeDm69GDKpkOUC9DNd5XHH0_gmGH5v6eXeF90FmmM=s0-d-e1-ft#https://static.wixstatic.com/media/d28008_3324608835144a0a809cb40c7812802b~mv2.png/v1/fit/w_680,h_2000,al_c,q_85/d28008_3324608835144a0a809cb40c7812802b~mv2.png" width="680" style="display:block;max-width:100%;vertical-align:middle;width:auto" alt="" class="CToWUd" data-bit="iit">

<div style="width:100%;text-align:center;>
	<div>
	3418 Fremont Avenue North, Seattle, WA, USA
	</div>
	<a href="https://lazycowbakeryseattle.com/">Check out our site</a>
	</div>
</div>`
