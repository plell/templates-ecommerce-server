package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func getTokenPath() string {
	tp := "tokens/google-token.json"
	return tp
}

// FIXME GO THROUGH ALL TOKENS IN token folder
func GoogleRefreshTokenIfExists() {

	// this will grab tokens from the db and refresh all,
	// but for now we store tokens in file
	// or store one token
	log.Println("GoogleRefreshTokenIfExists")

	tokenPath := getTokenPath()

	token, err := tokenFromFile(tokenPath)

	if err != nil {
		log.Println("No token to refresh")
		return
	}

	log.Println("AccessToken", token.AccessToken)
	log.Println("RefreshToken", token.RefreshToken)
	log.Println("TokenType", token.TokenType)
	log.Println("Expiry", token.Expiry)

	now := time.Now()

	timeUntilExpire := token.Expiry.Unix() - now.Unix()

	log.Println("timeUntilExpire", timeUntilExpire)

	config := SetupConfig()

	tokenSource := config.TokenSource(context.TODO(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Println("Failed to get token from Token source")
		return
	}

	if newToken.AccessToken != token.AccessToken {
		saveToken(tokenPath, newToken)
		log.Println("Saved new token")
	}

}

func GoogleLogin(c echo.Context) error {
	// vendor := c.Param("vendor")
	googleConfig := SetupConfig()

	// put vendor id into state! it will already be in the db, so on callback check the db
	url := googleConfig.AuthCodeURL(STORE_VENDOR_ID, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	c.Redirect(http.StatusSeeOther, url)
	return nil
}

func GoogleCallback(c echo.Context) error {
	state := c.Request().URL.Query()["state"][0]

	log.Println("GoogleCallback")

	// here is where we can get the vendor id back
	// to save the token in the right place
	// after getting record from db by vendor id
	if state != STORE_VENDOR_ID {
		log.Println(c.Response().Writer, " states dont match")
		return nil
	}

	code := c.Request().URL.Query()["code"][0]

	// vendor := c.Param("vendor")
	googleConfig := SetupConfig()

	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println(c.Response().Writer, " Code-Token Exchange Failed")
		return nil
	}

	log.Println("AccessToken", token.AccessToken)
	log.Println("RefreshToken", token.RefreshToken)
	log.Println("TokenType", token.TokenType)
	log.Println("Expiry", token.Expiry)

	tokenPath := getTokenPath()

	// store token
	saveToken(tokenPath, token)

	fmt.Fprintln(c.Response().Writer, "Authentication successful!")

	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(c echo.Context) *http.Client {
	// The file google-token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	// vendor := c.Param("vendor")
	config := SetupConfig()

	tokenPath := getTokenPath()

	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenPath, tok)
	}

	client := config.Client(context.Background(), tok)

	// Requests offline access.

	// client.setState({"access_type", "offline"})

	// Consent prompt is required to ensure a refresh token is always
	// returned when requesting offline access.
	// client.setParam("prompt", "consent")
	return client
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

func SendSubscriberEmail(c echo.Context) error {
	sendGoogleMail(c, STORE_EMAIL, "SendSubscriberEmail", "")
	return c.String(http.StatusOK, "ok")
}

func SendContactFormEmail(c echo.Context) error {
	sendGoogleMail(c, STORE_EMAIL, "SendContactFormEmail", "")
	return c.String(http.StatusOK, "ok")
}

// fixme there wont always be data!
func sendGoogleMail(c echo.Context, to string, subject string, bodyhtml string) {
	log.Println("sendGoogleMail")
	ctx := context.Background()
	client := getClient(c)

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return
	}

	_to := "To: " + to + "\n"
	_subject := "Subject: " + subject + "\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := "<html><body>" + bodyhtml + "</body></html>"
	byteMsg := []byte(_to + _subject + mime + body)

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

func createGoogleCalendarEvent(c echo.Context, orderData SessionOrderData) {
	log.Println("createGoogleCalendarEvent")
	ctx := context.Background()
	client := getClient(c)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Calendar client: %v", err)
		return
	}

	metadata := orderData.MetaData

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

	description += makeOrderDescriptionFromMetadata(metadata)

	description += "\n\n\nSHOP ORDER DETAILS \n\n"

	lineItems := orderData.LineItems

	for _, item := range lineItems {
		quantity := strconv.Itoa(int(item.StripeLineItem.Quantity))
		itemName := item.Name
		description += itemName + " (" + quantity + ") \n"
	}

	summary := "Order: " + firstName + " " + lastName

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
	log.Printf("Event created: %s\n", event.HtmlLink)

}
