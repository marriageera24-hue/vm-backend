package main

import (
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/labstack/echo"
	//"golang.org/x/crypto/bcrypt"
)

func passwordForgotHandler(ctx echo.Context) error {
	var (
		err error
		u   User
		r   struct {
			Email string `json:"email"`
		}
	)

	// Populate object from JSON
	err = ctx.Bind(&r)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if r.Email == "" {
		return ctx.JSON(http.StatusBadRequest, gettext("An email address is required", ctx))
	}

	err = db.Where("email ILIKE ?", r.Email).First(&u).Error
	if err != nil {
		return ctx.JSON(http.StatusNotFound, gettext("User not found", ctx))
	}

	resetToken, err := generateRandomString(16)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, gettext("Unable to create reset key", ctx))
	}

	hash, err := getPasswordHash(resetToken)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, gettext("Unable to hash reset key", ctx))
	}

	db.Model(&u).Updates(map[string]interface{}{"reset_triggered_at": time.Now(), "reset_token_hash": string(hash)})

	profileURL := os.Getenv("VM_PROFILE_URL")
	if u.Language == "ja" {
		profileURL = os.Getenv("VM_JP_PROFILE_URL")
	}

	url, err := url.ParseRequestURI(profileURL)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, gettext("Unable to build reset URL", ctx))
	}

	q := url.Query()
	q.Set("email", string(u.Email))
	q.Set("reset_token", resetToken)
	url.RawQuery = q.Encode()

	vars := make(map[string]string)
	vars["reset_url"] = url.String()
	go u.notify("user-forgot-password", string(u.Email), vars, nil)

	return ctx.JSON(http.StatusOK, gettext("Password reset email sent", ctx))
}

func passwordResetHandler(ctx echo.Context) error {
	var (
		err error
		u   User
		r   struct {
			Username      string `json:"username"`
			//ResetToken string `json:"token"`
			Password   string `json:"password"`
			//Password2  string `json:"password_2"`
		}
	)

	// Populate object from JSON
	err = ctx.Bind(&r)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	/*if r.Password != r.Password2 {
		return ctx.JSON(http.StatusBadRequest, gettext("Your passwords don't match22", ctx))
	}*/

	if r.Username == "" {
		return ctx.JSON(http.StatusUnauthorized, gettext("An username is required", ctx))
	}
	err = db.Where("email ILIKE ? OR phone = ? ", r.Username, r.Username).First(&u).Error
	if err != nil {
		return ctx.JSON(http.StatusNotFound, gettext("User not found", ctx))
	}


	err = checkPasswordStrength(ctx, r.Password)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	/*if r.Email == "" {
		return ctx.JSON(http.StatusBadRequest, gettext("An email address is required", ctx))
	}

	if db.Where("email ILIKE ?", r.Email).First(&u).RecordNotFound() {
		return ctx.JSON(http.StatusNotFound, gettext("User not found", ctx))
	}

	if db.First(&u, u.UUID).RecordNotFound() {
		return ctx.JSON(http.StatusNotFound, gettext("User not found", ctx))
	}

	if u.ResetTokenHash == "" {
		return ctx.JSON(http.StatusBadRequest, gettext("Incorrect reset key", ctx))
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.ResetTokenHash), []byte(r.ResetToken))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, gettext("Incorrect reset key", ctx))
	}

	if u.ResetTriggeredAt.Valid && u.ResetTriggeredAt.Time.AddDate(0, 0, 1).Before(time.Now()) {
		return ctx.JSON(http.StatusBadRequest, gettext("Your reset key has expired. Please complete your reset within a day of triggering the reset email.", ctx))
	}*/

	hash, err := getPasswordHash(r.Password)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, gettext("Unable to hash password", ctx))
	}

	db.Model(&u).Updates(
		map[string]interface{}{
			"password_hash":            string(hash),
			"failed_login_attempts":    0,
			"login_email_token_hash":   "",
			"login_email_triggered_at": nil,
			"reset_token_hash":         "",
			"reset_triggered_at":       nil,
		},
	)

	u.verify()

	return ctx.JSON(http.StatusOK, gettext("Password reset successfully", ctx))
}
