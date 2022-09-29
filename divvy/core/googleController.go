package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// func GoogleRefreshToken() {
// 	tok, err := tokenFromFile(tokenFileName)
// 	if err != nil {
// 		log.Println("No token to refresh")
// 		return
// 	}

// 	config := SetupConfig()

// 	// refresh token
// 	// tok.RefreshToken
// 	updatedToken, err := config.TokenSource(context.TODO(), token).Token()

// 	token, err := googleConfig.Exchange(context.Background(), code)
// 	if err != nil {
// 		log.Println("Failed to refresh!")
// 		return
// 	}

// 	// store new token
// 	saveToken(tokenFileName, token)
// }

const RFC3339 = "2006-01-02T15:04:05Z07:00"

var tokenFileName = "google-token.json"

func GoogleLogin(c echo.Context) error {
	googleConfig := SetupConfig()
	url := googleConfig.AuthCodeURL("randomstate")
	c.Redirect(http.StatusSeeOther, url)
	return nil
}

func GoogleCallback(c echo.Context) error {
	state := c.Request().URL.Query()["state"][0]
	if state != "randomstate" {
		log.Println(c.Response().Writer, " states dont match")
		return nil
	}

	code := c.Request().URL.Query()["code"][0]

	googleConfig := SetupConfig()

	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println(c.Response().Writer, " Code-Token Exchange Failed")
		return nil
	}

	// store token
	saveToken(tokenFileName, token)

	fmt.Fprintln(c.Response().Writer, "Authentication successful!")

	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient() *http.Client {
	// The file google-token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	config := SetupConfig()
	tok, err := tokenFromFile(tokenFileName)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFileName, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Printf("Unable to read authorization code: %v", err)
		return nil
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Printf("Unable to retrieve token from web: %v", err)
		return nil
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Unable to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func sendGoogleMail(to string, bodyhtml string) {
	ctx := context.Background()
	client := getClient()

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return
	}

	_to := "To: " + to + "\n"
	subject := "Subject: Test email from Go!\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := "<html><body>" + bodyhtml + "</body></html>"
	byteMsg := []byte(_to + subject + mime + body)

	encodedMessage := base64.StdEncoding.EncodeToString(byteMsg)

	msgbody := gmail.MessagePartBody{
		Data: encodedMessage,
	}

	headers := []*gmail.MessagePartHeader{&gmail.MessagePartHeader{
		Name:  "To",
		Value: to,
	}}

	payload := gmail.MessagePart{
		Body:    &msgbody,
		Headers: headers,
	}

	msg := gmail.Message{
		Payload: &payload,
		Raw:     encodedMessage,
	}

	res := gmailService.Users.Messages.Send("me", &msg)

	ok, err := res.Do()

	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return
	}

	fmt.Printf("ok.HTTPStatusCode: %v\n", ok.HTTPStatusCode)
}

func createGoogleCalendarEvent(metadata map[string]string, session stripe.CheckoutSession) {
	ctx := context.Background()
	client := getClient()

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Calendar client: %v", err)
		return
	}

	eventTime := ""

	log.Println("metadata", metadata)

	if _, ok := metadata["pickup_date"]; ok {
		eventTime = metadata["pickup_date"]
	}

	date := time.Now()
	if eventTime != "" {
		t, err := time.Parse(time.RFC3339, eventTime)
		if err != nil {
			return
		}
		date = t
	}

	dateString := date.Format(time.RFC3339)

	firstName := ""
	lastName := ""

	if _, ok := metadata["first_name"]; ok {
		firstName = metadata["first_name"]
	}
	if _, ok := metadata["last_name"]; ok {
		lastName = metadata["last_name"]
	}

	description := "FORM ORDER DETAILS \n\n"

	for key, md := range metadata {
		description += key + ": " + md + " \n"
	}

	description += "\n\n\nSHOP ORDER DETAILS \n\n"

	// for k, _ := range lineItems {
	// 	log.Println("k ", k)
	// }

	// for k, item := range lineItems {
	// 	log.Println("k ", k)
	// 	log.Println("item ", item)
	// 	quantity := strconv.Itoa(int(item.Quantity))
	// 	itemName := item.Price.Product.Name
	// 	description += itemName + "(" + quantity + ") - Item ID: " + item.ID + " \n"
	// }

	summary := "New Order: " + firstName + " " + lastName

	event := &calendar.Event{
		Summary:     summary,
		Location:    "",
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: dateString,
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: dateString,
			TimeZone: "America/Los_Angeles",
		},
		Recurrence: []string{},
		Attendees:  []*calendar.EventAttendee{},
	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Printf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)
	fmt.Printf("Event created: %s\n", event.HtmlLink)
}
