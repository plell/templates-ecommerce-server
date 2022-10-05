package core

import (
	"log"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type EmptyJSON struct {
}

var (
	IS_DEV     = os.Getenv("IS_DEV")
	EMPTY_JSON = EmptyJSON{}
	// prod
	PAYMENT_SUCCESS_URL = "https://www.lazycowbakeryseattle.com/success"
	PAYMENT_CANCEL_URL  = "https://www.lazycowbakeryseattle.com/"
	STORE_EMAIL         = "lazycowbakery@gmail.com"
	GOOGLE_REDIRECT_URL = "https://api.plellworks.com/google/callback"

	// dev
	// STORE_EMAIL         = "plelldavid@gmail.com"
	// GOOGLE_REDIRECT_URL = "http://localhost:8000/google/callback"
	// PAYMENT_SUCCESS_URL = "http://localhost:3000/success"
	// PAYMENT_CANCEL_URL  = "http://localhost:3000"
	STORE_NAME         = "Lazy Cow Bakery"
	STORE_DOMAIN_NAME  = "lazycowbakeryseattle.com"
	STORE_PRODUCT_NAME = "Custom Cake"
	// prod

	STORE_VENDOR_ID = "ZALdmzb6swk0wt"
	AUTH_TOKEN_PATH = "tokens/google-token-" + STORE_VENDOR_ID + ".json"
)

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
		RedirectURL:  GOOGLE_REDIRECT_URL,
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Endpoint: google.Endpoint,
	}

	return conf
}

func getReadableDate(date string) string {
	if date == "" {
		return date
	}

	formattedDate := date

	parsedDate, err := time.Parse(time.RFC3339, formattedDate)
	if err != nil {
		log.Println("could not parse date")
	} else {
		month := parsedDate.Month().String()
		day := strconv.Itoa(parsedDate.Day())
		year := strconv.Itoa(parsedDate.Year())
		formattedDate = month + " " + day + ", " + year
	}

	return formattedDate
}

func MakeReceiptEmailTemplate(data SessionOrderData) string {
	template := ""
	customerName := ""
	date := ""
	orderMeta := makeOrderMetaFromMetadata(data.MetaData)

	if _, ok := data.MetaData["pickup_date"]; ok {
		date = data.MetaData["pickup_date"]
		date = getReadableDate(date)
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
<h3 style="margin-bottom:20px;">Order Details</h3>
<div style="margin-bottom: 30px;">`

	// do cake
	for k, v := range orderMeta {
		template += "<div><i>" + k + ": " + v + "</i></div>"
	}

	// do shop
	for _, v := range data.LineItems {
		template += "<div><i>" + v.Name + " (" + strconv.Itoa(int(v.StripeLineItem.Quantity)) + ")</i></div>"
	}

	template += `
  </div>

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
		date = getReadableDate(date)
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
	<div style="margin-bottom:20px;">
		<h3 >Order from ` + customerName + `</h3>
		<div>(` + email + `)</div>
	</div>
	  <div style="margin: 20px 0;">`

	// do cake
	for k, v := range orderMeta {
		template += "<div><i>" + k + ": " + v + "</i></div>"
	}

	// do shop
	for _, v := range data.LineItems {
		template += "<div><i>" + v.Name + " (" + strconv.Itoa(int(v.StripeLineItem.Quantity)) + ")</i></div>"
	}

	template += `</div> 
	
	<div style="margin:20px 0">
	Pickup date is ` + date + `
	</div>
	` + TEMPLATE_FOOTER

	return template
}

var TEMPLATE_FOOTER = `
<img src="https://ci3.googleusercontent.com/proxy/rlDOKoI1LxmxbVHip-jv1aiKyoEIZrnXzkoypnSRkcR1SzhJLcnOoQDgr0nA4ae5fgX8LNim7ArsdlRQLTnAr8tqD6hS4Kq7kKbyPo2_4T7DWojB-MwNIanZVuiYMCEueYXhJXbEKppagMR5y28uw0Y04pw98HxmZZxTFSKPqlMnmIs5tEp4v3AxvsqUdH8F19hGECRDeDm69GDKpkOUC9DNd5XHH0_gmGH5v6eXeF90FmmM=s0-d-e1-ft#https://static.wixstatic.com/media/d28008_3324608835144a0a809cb40c7812802b~mv2.png/v1/fit/w_680,h_2000,al_c,q_85/d28008_3324608835144a0a809cb40c7812802b~mv2.png" width="680" style="display:block;max-width:100%;vertical-align:middle;width:auto" alt="" class="CToWUd" data-bit="iit">

<div style="width:100%;text-align:center;>
	<div style="margin-bottom:20px;">
	3418 Fremont Avenue North, Seattle, WA, USA
	</div>
	<a href="https://lazycowbakeryseattle.com/">Check out our site</a>
	</div>
</div>`

var EMAIL_HEADER = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html data-editor-version="2" class="sg-campaigns" xmlns="http://www.w3.org/1999/xhtml">
    <head>
      <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
      <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=1">
      <!--[if !mso]><!-->
      <meta http-equiv="X-UA-Compatible" content="IE=Edge">
      <!--<![endif]-->
      <!--[if (gte mso 9)|(IE)]>
      <xml>
        <o:OfficeDocumentSettings>
          <o:AllowPNG/>
          <o:PixelsPerInch>96</o:PixelsPerInch>
        </o:OfficeDocumentSettings>
      </xml>
      <![endif]-->
      <!--[if (gte mso 9)|(IE)]>
  <style type="text/css">
    body {width: 600px;margin: 0 auto;}
    table {border-collapse: collapse;}
    table, td {mso-table-lspace: 0pt;mso-table-rspace: 0pt;}
    img {-ms-interpolation-mode: bicubic;}
  </style>
<![endif]-->
      <style type="text/css">
    body, p, div {
      font-family: inherit;
      font-size: 14px;
    }
    body {
      color: #000000;
    }
    body a {
      color: #1188E6;
      text-decoration: none;
    }
    p { margin: 0; padding: 0; }
    table.wrapper {
      width:100% !important;
      table-layout: fixed;
      -webkit-font-smoothing: antialiased;
      -webkit-text-size-adjust: 100%;
      -moz-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
    }
    img.max-width {
      max-width: 100% !important;
    }
    .column.of-2 {
      width: 50%;
    }
    .column.of-3 {
      width: 33.333%;
    }
    .column.of-4 {
      width: 25%;
    }
    ul ul ul ul  {
      list-style-type: disc !important;
    }
    ol ol {
      list-style-type: lower-roman !important;
    }
    ol ol ol {
      list-style-type: lower-latin !important;
    }
    ol ol ol ol {
      list-style-type: decimal !important;
    }
    @media screen and (max-width:480px) {
      .preheader .rightColumnContent,
      .footer .rightColumnContent {
        text-align: left !important;
      }
      .preheader .rightColumnContent div,
      .preheader .rightColumnContent span,
      .footer .rightColumnContent div,
      .footer .rightColumnContent span {
        text-align: left !important;
      }
      .preheader .rightColumnContent,
      .preheader .leftColumnContent {
        font-size: 80% !important;
        padding: 5px 0;
      }
      table.wrapper-mobile {
        width: 100% !important;
        table-layout: fixed;
      }
      img.max-width {
        height: auto !important;
        max-width: 100% !important;
      }
      a.bulletproof-button {
        display: block !important;
        width: auto !important;
        font-size: 80%;
        padding-left: 0 !important;
        padding-right: 0 !important;
      }
      .columns {
        width: 100% !important;
      }
      .column {
        display: block !important;
        width: 100% !important;
        padding-left: 0 !important;
        padding-right: 0 !important;
        margin-left: 0 !important;
        margin-right: 0 !important;
      }
      .social-icon-column {
        display: inline-block !important;
      }
    }
  </style>
      <!--user entered Head Start--><link href="https://fonts.googleapis.com/css?family=Viga&display=swap" rel="stylesheet"><style>
    body {font-family: 'Viga', sans-serif;}
</style><!--End Head user entered-->
    </head>
    <body>
      <center class="wrapper" data-link-color="#1188E6" data-body-style="font-size:14px; font-family:inherit; color:#000000; background-color:#f0f0f0;">
        <div class="webkit">`

var EMAIL_FOOTER = `
		</div>
      </center>
	  </body>
	  </html>
        `
