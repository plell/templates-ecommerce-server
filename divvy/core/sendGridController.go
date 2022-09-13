// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package core

import (
	"fmt"
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stripe/stripe-go/v72"
)

type DynamicData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var SENDGRID_PAYMENT_RECEIVED_TEMPLATE = "d-5a26312bde7d41f8a292a654f9a60e8e"
var SENDGRID_PAYMENT_RECEIPT_TEMPLATE = "d-722ac28ca10b4ab7978698f7fcc4b1cc"

// payment received, to store
func SendPaymentReceivedEmail(c stripe.Charge) {
	log.Println("SendPaymentReceivedEmail")

	customerEmail := c.BillingDetails.Email
	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "amount",
		Value: FormatAmountToString(c.Amount, "$"),
	})
	dd = append(dd, DynamicData{
		Key:   "customerEmail",
		Value: customerEmail,
	})
	emails := []string{}
	SendEmail("payment", SENDGRID_PAYMENT_RECEIVED_TEMPLATE, emails, dd)
	log.Println("SENT PAYMENT EMAIL!")
	SendPaymentReceiptEmail(c, "Online Store")
}

// payment received, to customer
func SendPaymentReceiptEmail(c stripe.Charge, podName string) {
	log.Println("SendPaymentReceiptEmail")
	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "amount",
		Value: FormatAmountToString(int64(c.Amount), "$"),
	})
	dd = append(dd, DynamicData{
		Key:   "podName",
		Value: podName,
	})

	customerEmail := c.BillingDetails.Email
	emails := []string{}
	emails = append(emails, customerEmail)

	SendEmail("receipt", SENDGRID_PAYMENT_RECEIPT_TEMPLATE, emails, dd)
}

// general function used by all email routes
func SendEmail(sender string, templateId string, toEmails []string, dynamicData []DynamicData) {

	m := mail.NewV3Mail()

	address := sender + "@" + STORE_DOMAIN_NAME
	name := STORE_NAME
	e := mail.NewEmail(name, address)
	m.SetFrom(e)

	m.SetTemplateID(templateId)

	p := mail.NewPersonalization()
	tos := []*mail.Email{}

	for _, em := range toEmails {
		tos = append(tos, mail.NewEmail("", em))
	}

	p.AddTos(tos...)

	for _, dd := range dynamicData {
		p.SetDynamicTemplateData(dd.Key, dd.Value)
	}

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	response, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
