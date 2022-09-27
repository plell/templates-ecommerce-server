package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v72"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func getGoogleCalendarKey() string {
	googleKey := os.Getenv("GOOGLE_CALENDAR_API_KEY")
	return googleKey
}

func createGoogleCalendarEvent(session stripe.CheckoutSession) {
	ctx := context.Background()
	googleApiKey := getGoogleCalendarKey()
	srv, err := calendar.NewService(ctx, option.WithAPIKey(googleApiKey))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	metadata := session.PaymentIntent.Metadata

	eventTime := time.Now().String()

	if _, ok := metadata["pickup_date"]; ok {
		eventTime = metadata["pickup_date"]
	}

	description := "FORM ORDER DETAILS \n\n"

	for _, md := range session.PaymentIntent.Metadata {
		description += md + ", \n"
	}

	description = "\n\n\nSHOP ORDER DETAILS \n\n"

	items := session.LineItems.Data
	for _, item := range items {
		quantity := strconv.Itoa(int(item.Quantity))
		itemName := item.Price.Product.Name
		description += itemName + "(" + quantity + ") - Item ID: " + item.ID + " \n"
	}

	summary := "New Order"

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: eventTime,
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: eventTime,
			TimeZone: "America/Los_Angeles",
		},
	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)
}
