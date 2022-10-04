package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"github.com/stripe/stripe-go/v72/webhook"
)

func getStripeKey() string {
	stripeKey := os.Getenv("STRIPE_API_KEY")
	return stripeKey
}

func getStripeWebhookKey() string {
	stripeKey := os.Getenv("STRIPE_WEBHOOK_SECRET")
	return stripeKey
}

type CreateCheckoutSessionResponse struct {
	SessionID string `json:"sessionId"`
}

type StripeProduct struct {
	SessionID string `json:"sessionId"`
}

func GetProductsFromStripe(c echo.Context) error {
	stripe.Key = getStripeKey()

	// get prices
	priceParams := &stripe.PriceListParams{}
	price_i := price.List(priceParams)

	var prices = make(map[string]string)

	for price_i.Next() {
		p := price_i.Price()
		prod_id := stripe.String(p.Product.ID)
		prices[*prod_id] = strconv.Itoa(int(p.UnitAmount))
	}

	params := &stripe.ProductListParams{}

	// get products
	i := product.List(params)
	products := []*stripe.Product{}

	for i.Next() {
		p := i.Product()
		var metaData = make(map[string]string)
		if _, ok := prices[p.ID]; ok {
			metaData["price"] = prices[p.ID]
		}
		p.Metadata = metaData
		products = append(products, p)
	}

	return c.JSON(http.StatusOK, products)
}

type CheckoutSessionRequest struct {
	Amount   int64             `json:"amount"`
	Products map[string]int64  `json:"products"`
	Currency string            `json:"currency"`
	Form     map[string]string `json:"form"`
}

func makeOrderMetaData(request CheckoutSessionRequest) map[string]string {
	var metaDataPack = make(map[string]string)
	for key, value := range request.Form {
		metaDataPack[key] = value
	}
	return metaDataPack
}

var skipList = []string{"pickup_date", "first_name", "last_name", "phone", "email", "cake_type"}

func makeOrderDescriptionFromMetadata(metadata map[string]string) string {
	description := ""

	for key, value := range metadata {
		if stringArrayContains(skipList, key) {
			continue
		}
		// replace _ and make capitalize words
		newKey := strings.ToUpper(strings.ReplaceAll(key, "_", " "))
		description += newKey + ": " + value + " / \n"
	}

	return description
}

func makeOrderMetaFromMetadata(metadata map[string]string) map[string]string {
	meta := make(map[string]string)

	for key, value := range metadata {
		if stringArrayContains(skipList, key) {
			continue
		}
		// replace _ and make capitalize words
		newKey := strings.Title(strings.ReplaceAll(key, "_", " "))
		meta[newKey] = strings.ToUpper(value)
	}

	return meta
}

func CreateProductCheckoutSessionByCustomer(c echo.Context) error {
	// you do not need a token to run this route
	// potential for abuse

	// here decode to get the customAmount, if there
	request := CheckoutSessionRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	stripe.Key = getStripeKey()

	// get prices

	var prices = make(map[string]int64)

	priceParams := &stripe.PriceListParams{}
	price_i := price.List(priceParams)

	for price_i.Next() {
		p := price_i.Price()
		prod_id := stripe.String(p.Product.ID)
		prices[*prod_id] = p.UnitAmount
	}

	// set up line items
	lineItems := []*stripe.CheckoutSessionLineItemParams{}

	for product_id, quantity := range request.Products {
		unitAmount := stripe.Int64(prices[product_id])
		item := &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:   stripe.String(string(stripe.CurrencyUSD)),
				Product:    stripe.String(string(product_id)),
				UnitAmount: unitAmount,
			},

			Quantity: stripe.Int64(quantity),
		}
		lineItems = append(lineItems, item)
	}

	metaData := makeOrderMetaData(request)

	customerEmail := ""
	if _, ok := metaData["email"]; ok {
		customerEmail = metaData["email"]
	}

	humanReadableMetaDataString := makeOrderDescriptionFromMetadata(metaData)

	params := &stripe.CheckoutSessionParams{
		CustomerEmail: &customerEmail,
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			ReceiptEmail: &customerEmail,
			Metadata:     metaData,
			Description:  stripe.String(humanReadableMetaDataString),
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems:  lineItems,
		SuccessURL: stripe.String(PAYMENT_SUCCESS_URL),
		CancelURL:  stripe.String(PAYMENT_CANCEL_URL),
	}

	session, err := session.New(params)

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	data := CreateCheckoutSessionResponse{
		SessionID: session.ID,
	}

	return c.JSON(http.StatusOK, data)
}

func CreateAmountCheckoutSessionByCustomer(c echo.Context) error {

	// you do not need a token to run this route
	// potential for abuse
	request := CheckoutSessionRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	// do amount
	amount := request.Amount

	if amount < 100 {
		return c.String(http.StatusInternalServerError, "Amount minimum is 1USD")
	}

	stripe.Key = getStripeKey()

	metaData := makeOrderMetaData(request)

	customerEmail := ""
	if _, ok := metaData["email"]; ok {
		customerEmail = metaData["email"]
	}

	humanReadableMetaDataString := makeOrderDescriptionFromMetadata(metaData)

	params := &stripe.CheckoutSessionParams{
		CustomerEmail: &customerEmail,
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			ReceiptEmail: &customerEmail,
			Metadata:     metaData,
			Description:  stripe.String(humanReadableMetaDataString),
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(STORE_PRODUCT_NAME),
						Metadata:    metaData,
						Description: stripe.String(humanReadableMetaDataString),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(PAYMENT_SUCCESS_URL),
		CancelURL:  stripe.String(PAYMENT_CANCEL_URL),
	}

	session, err := session.New(params)

	if err != nil {
		log.Println("session error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	data := CreateCheckoutSessionResponse{
		SessionID: session.ID,
	}

	return c.JSON(http.StatusOK, data)
}

func getSessionOrderData(sess stripe.CheckoutSession, c echo.Context) SessionOrderData {

	data := SessionOrderData{}

	// first get products, because lineItems dont include names!
	stripe.Key = getStripeKey()

	products := []*stripe.Product{}
	params := &stripe.ProductListParams{}

	i := product.List(params)
	for i.Next() {
		p := i.Product()
		products = append(products, p)
	}

	lineItems := []OrderLineItem{}

	ii := session.ListLineItems(*stripe.String(sess.ID), nil)
	for ii.Next() {
		li := ii.LineItem()

		thisLineItem := OrderLineItem{}
		thisLineItem.StripeLineItem = li

		// get name
		for _, p := range products {
			if p.ID == li.Price.Product.ID {
				thisLineItem.Name = p.Name
			}
		}

		lineItems = append(lineItems, thisLineItem)
	}

	pi, _ := paymentintent.Get(
		*stripe.String(sess.PaymentIntent.ID),
		nil,
	)

	data.LineItems = lineItems
	data.MetaData = pi.Metadata
	data.PaymentIntent = pi

	return data

}

type ChargeList struct {
	Amount int64  `json:"amount"`
	ID     string `json:"id"`
}

type ChargeListItem struct {
	ID                string            `json:"id"`
	PaymentMethodCard PaymentMethodCard `json:"paymentMethodCard"`
	Amount            int64             `json:"amount"`
	Refunded          bool              `json:"refunded"`
	Metadata          map[string]string `json:"metadata"`
	Created           int64             `json:"created"`
	Paid              bool              `json:"paid"`
	HasMore           bool              `json:"hasMore"`
}

type PaymentMethodCard struct {
	Network stripe.PaymentMethodCardNetwork `json:"network"`
	Last4   string                          `json:"last4"`
}

type ListNav struct {
	StartingAfterID string `json:"startingAfterId"`
	EndingBeforeID  string `json:"endingBeforeId"`
}

// Payment Intents API
// When using the Payment Intents API with Stripe’s client libraries and SDKs, ensure that:

// Authentication flows are triggered when required (use the regulatory test card numbers and PaymentMethods.)
// No authentication (default U.S. card): 4242 4242 4242 4242.
// Authentication required: 4000 0027 6000 3184.
// The PaymentIntent is created with an idempotency key to avoid erroneously creating duplicate PaymentIntents for the same purchase.
// Errors are caught and displayed properly in the UI.

// webhooks
// session checkout complete

// stripe listen --forward-to localhost:8000/webhook
func HandleStripeWebhook(c echo.Context) error {
	w := c.Response().Writer
	req := c.Request()
	// w http.ResponseWriter, req *http.Request
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return c.String(http.StatusOK, "ok")
	}

	webhookSecret := getStripeWebhookKey()

	// Verify webhook signature and extract the event.
	// See https://stripe.com/docs/webhooks/signatures for more information.
	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), webhookSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature.
		return c.String(http.StatusOK, "ok")
	}

	log.Println(event.Type)

	if event.Type == "charge.succeeded" {
		var ch stripe.Charge
		err := json.Unmarshal(event.Data.Raw, &ch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return c.String(http.StatusOK, "ok")
		}
		handleSuccessfulCharge(ch)
	}
	if event.Type == "payment_intent.succeeded" {
		var intent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &intent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return c.String(http.StatusOK, "ok")
		}
		handleSuccessfulPaymentIntent(intent)
	}
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return c.String(http.StatusOK, "ok")
		}
		handleCompletedCheckoutSession(session, c)
	}

	return c.String(http.StatusOK, "ok")
}

type OrderLineItem struct {
	StripeLineItem *stripe.LineItem `json:"stripeLineItem"`
	Name           string           `json:"name"`
}

type SessionOrderData struct {
	LineItems     []OrderLineItem       `json:"lineItems"`
	MetaData      map[string]string     `json:"metadata"`
	PaymentIntent *stripe.PaymentIntent `json:"paymentIntent"`
}

func handleCompletedCheckoutSession(sess stripe.CheckoutSession, c echo.Context) {
	// Fulfill the purchase.
	// here is where the transaction record is updated, with a completed status
	log.Println("handleCompletedCheckoutSession")

	WebsocketWriter(&SocketMessage{
		Amount:          sess.AmountTotal,
		PaymentIntentID: sess.PaymentIntent.ID,
	})

	sessionData := getSessionOrderData(sess, c)

	newOrderEmailTemplate := MakeNewOrderEmailTemplate(sessionData)
	receiptEmailTemplate := MakeReceiptEmailTemplate(sessionData)

	sendGoogleMail(c, STORE_EMAIL, "New order!", newOrderEmailTemplate)
	sendGoogleMail(c, sessionData.PaymentIntent.ReceiptEmail, "Thank you for your order!", receiptEmailTemplate)

	createGoogleCalendarEvent(c, sessionData)
}

func handleSuccessfulPaymentIntent(intent stripe.PaymentIntent) {
	// here is where the transaction record is updated, with a completed status
	log.Println("handleSuccessfulPaymentIntent")
	amount := intent.Amount

	WebsocketWriter(&SocketMessage{
		Amount:          amount,
		PaymentIntentID: intent.ID,
	})
}

func handleSuccessfulCharge(ch stripe.Charge) {
	// here is where the transaction record is updated, with a completed status
	log.Println("handleSuccessfulCharge")
	amount := ch.Amount

	WebsocketWriter(&SocketMessage{
		Amount:          amount,
		PaymentIntentID: ch.PaymentIntent.ID,
	})

	// newcharge := Charge{
	// 	ChargeID: ch.ID,
	// 	Amount:   int64(amount),
	// }

	// DB.Create(&newcharge)

	// SendPaymentReceivedEmail(ch)
}
