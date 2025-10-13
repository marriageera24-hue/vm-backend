package main

import (
	"github.com/labstack/echo"
	"github.com/pariz/gountries"
)

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (c *Country) sanitize(ctx echo.Context) {
	code := c.Code

	c.Code = ""
	c.Name = ""

	if code == "" {
		return
	}

	query := gountries.New()
	country, err := query.FindCountryByAlpha(code)
	if err != nil {
		return
	}

	c.Code = country.Codes.Alpha2
	c.Name = country.Name.Common
}
