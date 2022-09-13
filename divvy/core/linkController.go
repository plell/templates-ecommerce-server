package core

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateCheckoutLink(c echo.Context) error {
	isFixedAmount := true

	request := Link{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	if request.Amount == 0 {
		// this is a variable donation, donors choose the price
		isFixedAmount = false
	} else if request.Amount < 100 {
		return c.String(http.StatusInternalServerError, "Amount minimum is 1USD")
	}

	link := Link{
		Amount:        request.Amount,
		Selector:      MakeSelector(LINK_TABLE),
		IsFixedAmount: isFixedAmount,
	}

	return c.JSON(http.StatusOK, link)
}
