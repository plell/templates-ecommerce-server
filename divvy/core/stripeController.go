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

	"github.com/stripe/stripe-go/v72/webhook"
)

func getStripeKey() string {
	stripeKey := os.Getenv("STRIPE_API_KEY")
	return stripeKey
}

type CreateCheckoutSessionResponse struct {
	SessionID       string `json:"sessionId"`
	PaymentIntentID string `json:"paymentIntentId"`
	Price           int64  `json:"price"`
}

type CheckoutSessionRequest struct {
	Amount       int64  `json:"amount"`
	CustomAmount int64  `json:"customAmount"`
	PodSelector  string `json:"podSelector"`
	Currency     string `json:"currency"`
	CustomerID   string `json:"customerId"`
	UserSelector string `json:"userSelector"`
}

func CreateCheckoutSessionFromLinkByCustomer(c echo.Context) error {

	// the user uses this to get a checkout session from a link.
	// if the link isFixedAmount, amount check
	// if not, get "customAmount" from request body instead

	// you do not need a token to run this route
	// potential for abuse

	// here decode to get the customAmount, if there
	request := CheckoutSessionRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	// do amount
	amount := request.Amount

	fmt.Println("here is the amount: ")
	fmt.Println(amount)

	if amount < 100 {
		return c.String(http.StatusInternalServerError, "Amount minimum is 1USD")
	}

	stripe.Key = getStripeKey()
	params := &stripe.CheckoutSessionParams{
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{},
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
		// return c.Error(err)
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
