// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stripe/stripe-go/v72"
)

type DynamicData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var SENDGRID_PAYMENT_RECEIVED_TEMPLATE = "d-7026b77b41ed47f08363f2c077bf9e21"
var SENDGRID_ORDER_RECEIVED_TEMPLATE = "d-cd8f45b8cb254e76b9eff4b42048ba9b"

var SENDGRID_CONTACT_FORM_EMAIL = "d-2976a799a7a54c5fa54031c67ee79615"
var SENDGRID_SUBSCRIBER_EMAIL = "d-b0666a75d01e46dc8621aaf35a2287f6"

func addMetaData(paymentIntent stripe.PaymentIntent, session stripe.CheckoutSession) []DynamicData {
	metaData := paymentIntent.Metadata
	lineItems := session.LineItems.Data
	dd := []DynamicData{}

	for key, value := range metaData {
		dd = append(dd, DynamicData{
			Key:   key,
			Value: value,
		})
	}

	for _, item := range lineItems {
		dd = append(dd, DynamicData{
			Key:   "Item ID: " + item.ID,
			Value: strconv.Itoa(int(item.Quantity)),
		})
	}

	return dd
}

// order received email to owner
func SendOrderReceivedEmail(session stripe.CheckoutSession) {
	log.Println("SendOrderReceivedEmail")

	dd := addMetaData(*session.PaymentIntent, session)

	emails := []string{}
	emails = append(emails, STORE_EMAIL)

	SendEmail("orders", "You got an order!", SENDGRID_ORDER_RECEIVED_TEMPLATE, emails, dd)
}

// payment received, to store
func SendPaymentReceivedEmail(c stripe.Charge, s stripe.CheckoutSession) {
	log.Println("SendPaymentReceivedEmail")
}

type ContactRequest struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// contact form to owner
func SendContactFormEmail(c echo.Context) error {
	// you do not need a token to run this route
	// potential for abuse
	log.Println("SendContactFormEmail")

	request := ContactRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	emails := []string{STORE_EMAIL}

	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "email",
		Value: request.Email,
	})
	dd = append(dd, DynamicData{
		Key:   "name",
		Value: request.Name,
	})
	dd = append(dd, DynamicData{
		Key:   "message",
		Value: request.Message,
	})

	SendEmail("contact", "You got a message!", SENDGRID_CONTACT_FORM_EMAIL, emails, dd)

	return c.String(http.StatusOK, "Success")
}

type SubscriberRequest struct {
	Email string `json:"email"`
}

// mailing list subscriber request to owner
func SendSubscriberEmail(c echo.Context) error {
	log.Println("SendMailingListSubscriberEmail")

	// here decode to get the customAmount, if there
	request := SubscriberRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	emails := []string{STORE_EMAIL}

	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "email",
		Value: request.Email,
	})

	SendEmail("subscribers", "You got a subscriber!", SENDGRID_CONTACT_FORM_EMAIL, emails, dd)

	return c.String(http.StatusOK, "Success")
}

// general function used by all email routes
func SendEmail(sender string, subject string, templateId string, toEmails []string, dynamicData []DynamicData) {

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

	p.Subject = subject

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
