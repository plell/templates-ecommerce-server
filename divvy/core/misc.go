package core

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/leekchan/accounting"
	_ "github.com/leekchan/accounting"
	_ "github.com/shopspring/decimal"
)

func Pong(c echo.Context) error {
	return c.String(http.StatusOK, "Pong")
}

func AbstractError(c echo.Context, message string) error {
	return c.String(http.StatusInternalServerError, message)
}

var pool = "abcdefghijklmnopqrstuvwxyzABCEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func MakeSelector(tableName string) string {
	rand.Seed(time.Now().UnixNano())
	l := 24
	bytes := make([]byte, l)

	randomSelector := ""
	// enter while loop, exited when n = 2
	n := 0
	for n < 1 {
		// create random string
		for i := 0; i < l; i++ {
			bytes[i] = pool[rand.Intn(len(pool))]
		}

		randomSelector = string(bytes)
		selector := Selector{}

		// create record in selectors to make sure only unique selector are made
		result := DB.Table(tableName).Where("selector = ?", randomSelector).First(&selector)
		if result.Error != nil {
			// good, this is a unique selector
			selector := Selector{
				Selector: randomSelector,
				Type:     tableName,
			}
			result := DB.Create(&selector) // pass pointer of data to Create
			if result.Error != nil {
				// db create failed
			}
			// leave loop
			log.Println("Made unique selector")
			n++
		} else {
			log.Println("Made duplicate selector, retry")
		}
	}

	return randomSelector
}

func MakeInviteCode() string {
	rand.Seed(time.Now().UnixNano())
	l := 24
	bytes := make([]byte, l)

	randomSelector := ""
	// create random string
	for i := 0; i < l; i++ {
		bytes[i] = pool[rand.Intn(len(pool))]
	}

	randomSelector = string(bytes)

	return randomSelector
}

func ContainsInt(arr []uint, val uint) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

func FormatAmountToString(amount int64, symbol string) string {
	// p := strconv.Itoa(int(amount))

	af := float64(amount) / 100

	ac := accounting.Accounting{Symbol: symbol, Precision: 2}

	a := ac.FormatMoney(af)

	return a
}

func FormatStringAmountNoSymbol(amount string) string {
	p, _ := strconv.Atoi(amount)

	af := float64(p) / 100

	ac := accounting.Accounting{Symbol: "", Precision: 2}

	a := ac.FormatMoney(af)

	return a
}

func stringArrayContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
