package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func pingHandler(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}
