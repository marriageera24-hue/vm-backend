package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
)

type Email string

func (e Email) validate(ctx echo.Context) error {

	if e == "" {
		return errors.New("Email Address is required")
	}

	parts := strings.Split(string(e), "@")
	if len(parts) != 2 {
		return fmt.Errorf("Invalid Email Address: %s", e)
	}

	if !govalidator.IsEmail(string(e)) {
		return fmt.Errorf("Invalid Email Address: %s", e)
	}

	mxrecords, _ := net.LookupMX(parts[1])
	if len(mxrecords) == 0 {
		return fmt.Errorf("We cannot send emails to '%s'. Please remove and add a valid email address. Contact support if you believe there has been an error.", parts[1])
	}

	return nil
}

func (e *User) getVerifyURL(host string, currentPath string) (string, error) {
	// Build URL
	loginURL := os.Getenv("VM_PROFILE_URL")

	host = sanitizeHost(host)

	if currentPath != "" {
		current, err := url.ParseRequestURI("https://" + host + currentPath)
		if err == nil {
			loginURL = "https://" + host + current.Path
		}
	}

	// Create, hash, and save token
	token, err := generateRandomString(16)
	if err != nil {
		return loginURL, errors.New("Unable to create email verify token")
	}

	hash, err := getPasswordHash(token)
	if err != nil {
		return loginURL, errors.New("Unable to hash email verify token")
	}

	db.Model(&e).Updates(map[string]interface{}{"verification_hash": string(hash)})

	targetURL, err := url.ParseRequestURI(loginURL)
	if err != nil {
		return loginURL, errors.New("Unable to create email verify URL")
	}

	q := targetURL.Query()
	q.Set("email", string(e.Email))
	q.Set("verification_token", token)
	targetURL.RawQuery = q.Encode()

	return targetURL.String(), nil
}
func (u *User) sendVerificationEmail(e Email, host string, currentPath string) {
	url, err := u.getVerifyURL(host, currentPath)
	if err != nil {
		return
	}

	vars := make(map[string]string)
	vars["verify_url"] = url
	u.notify("user-email-verify", string(e), vars, nil)
}
