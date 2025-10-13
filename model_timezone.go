package main

import (
	"errors"
	"time"

	"github.com/labstack/echo"
)

type Timezone string

func (s *Timezone) validate(ctx echo.Context) error {
	_, err := time.LoadLocation(string(*s))
	if err != nil {
		return errors.New("Timezone is invalid")
	}

	return nil
}
