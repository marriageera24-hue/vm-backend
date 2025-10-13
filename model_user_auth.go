package main

import (
	"errors"

	"github.com/labstack/echo"
)

type UserAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *UserAuth) validate(ctx echo.Context) (User, error) {
	var (
		u   User
		err error
	)

	err = db.Where("email ILIKE ? OR phone = ? ", a.Username, a.Username).First(&u).Error
	if err != nil {
		return User{}, errors.New(gettext("No user found with that email address or phone number", ctx))
	}

	err = verifyPassword(ctx, u, a.Password)
	if err != nil {
		return u, errors.New(gettext("The password you entered is incorrect", ctx))
	}

	return u, nil
}
