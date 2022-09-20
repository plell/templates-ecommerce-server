package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/product"
	"github.com/stripe/stripe-go/v72/webhook"
)

func getStripeKey() string {
	stripeKey := os.Getenv("STRIPE_API_KEY")
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

	params := &stripe.ProductListParams{}

	// params.Filters.AddFilter("limit", "", "3")
	i := product.List(params)

	products := []*stripe.Product{}

	for i.Next() {
		p := i.Product()
		products = append(products, p)
	}

	return c.JSON(http.StatusOK, products)
}

type CustomerDetails struct {
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	PickupDate     string `json:"pickup_date"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	AddressLine1   string `json:"address_line_1"`
	AddressLine2   string `json:"address_line_2"`
	AddressState   string `json:"address_state"`
	AddressZip     string `json:"address_zip"`
	AddressCountry string `json:"address_country"`
}

type CheckoutSessionRequest struct {
	Amount   int64            `json:"amount"`
	Products map[string]int64 `json:"products"`
	Currency string           `json:"currency"`
	Customer CustomerDetails  `json:"customer"`
}

func makeOrderMetaData(request CheckoutSessionRequest) map[string]string {
	var metaDataPack = make(map[string]string)
	metaDataPack["email"] = request.Customer.Email
	metaDataPack["name"] = request.Customer.FirstName + " " + request.Customer.LastName
	metaDataPack["pickup_date"] = request.Customer.PickupDate
	metaDataPack["phone"] = request.Customer.Phone
	metaDataPack["address_line1"] = request.Customer.AddressLine1
	metaDataPack["address_line2"] = request.Customer.AddressLine2
	metaDataPack["address_state"] = request.Customer.AddressState
	metaDataPack["address_country"] = request.Customer.AddressCountry
	metaDataPack["address_zip"] = request.Customer.AddressZip
	return metaDataPack
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

	// set up line items
	lineItems := []*stripe.CheckoutSessionLineItemParams{}

	for product_id, quantity := range request.Products {
		item := &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Product:  stripe.String(string(product_id)),
				Currency: stripe.String(string(stripe.CurrencyUSD)),
			},
			Quantity: stripe.Int64(quantity),
		}
		lineItems = append(lineItems, item)
	}

	metaData := makeOrderMetaData(request)

	params := &stripe.CheckoutSessionParams{
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metaData,
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

type AmountCheckoutSessionRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
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

	metaData := makeOrderMetaData(request)

	stripe.Key = getStripeKey()
	params := &stripe.CheckoutSessionParams{
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metaData,
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
						Name: stripe.String(STORE_PRODUCT_NAME),
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
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	data := CreateCheckoutSessionResponse{
		SessionID: session.ID,
	}

	return c.JSON(http.StatusOK, data)
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

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

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
		handleCompletedCheckoutSession(session)
	}

	return c.String(http.StatusOK, "ok")
}

func handleCompletedCheckoutSession(session stripe.CheckoutSession) {
	// Fulfill the purchase.
	// here is where the transaction record is updated, with a completed status
	log.Println("handleCompletedCheckoutSession")

	WebsocketWriter(&SocketMessage{
		Amount:          session.AmountTotal,
		PaymentIntentID: session.PaymentIntent.ID,
	})
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

	newcharge := Charge{
		ChargeID: ch.ID,
		Amount:   int64(amount),
	}

	DB.Create(&newcharge)

	SendPaymentReceivedEmail(ch)
}
