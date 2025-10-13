package main

import (
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

type apiObject interface {
	zeroID()
	sanitize(echo.Context)
	validate(echo.Context, bool) error
	exists() bool
	getExistsMessage(echo.Context) string
	doSearch(SearchQuery) (interface{}, error)
	bulkRead([]string) (interface{}, error)
	restore(echo.Context, *gorm.DB) error

	postRead()
	postSearch()
}

func getAPIItem(ctx echo.Context) (apiObject, error) {
	var item apiObject

	if strings.HasPrefix(ctx.Path(), "/email_templates") {
		return &EmailTemplate{}, nil
	}

	if strings.HasPrefix(ctx.Path(), "/users") {
		return &User{}, nil
	}

	return item, errors.New("invalid endpoint")
}
