package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func returnInvalidData(ctx echo.Context, err error) error {
	text := fmt.Sprintf("Path: %s <br> Method: %s <br> Error: %s", ctx.Path(), ctx.Request().Method, err)

	go sendDebugEmail("Invalid Data Submitted", text)

	return ctx.JSON(http.StatusBadRequest, "Invalid data")
}
