package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/tomasen/realip"
	"golang.org/x/crypto/bcrypt"
)

func getCookieName(name string) string {
	env := os.Getenv("VM_ENVIRONMENT")
	if env == "production" {
		return name
	}

	return name + "_" + env
}

func createSession(ctx echo.Context, u User) (string, error) {
	var s Session

	// Delete other sessions for user
	db.Where("user_id = ?", u.ID).Delete(&Session{})

	et, err := strconv.ParseInt(os.Getenv("JWT_EXPIRATION_PERIOD"), 0, 64)

	if err != nil {
		et = 12
	}

	etd := time.Duration(et)
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.ID
	claims["name"] = u.getName()
	claims["admin"] = false
	// claims["user_id"] = u.ID
	// d, err := json.Marshal(u)
	// if err == nil {
	// claims["user_data"] = string(d)
	// }

	exp := time.Now().Add(time.Hour * etd)
	claims["exp"] = exp.Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	// Setup new session
	s.UserID = u.ID
	s.Token = t
	s.IPAddress = realip.FromRequest(ctx.Request())
	s.ExpiresAt = exp

	// Save session
	db.NewRecord(s)
	err = db.Create(&s).Error
	if err != nil {
		return "", errors.New(gettext("Unable to create session. Please try again.", ctx))
	}

	// Update last login
	db.Model(&u).Updates(
		map[string]interface{}{
			"last_login_at":            time.Now(),
			"last_login_attempt_at":    time.Now(),
			"failed_login_attempts":    0,
			"last_login_ip_address":    s.IPAddress,
			"login_email_token_hash":   "",
			"login_email_triggered_at": nil,
			"reset_token_hash":         "",
			"reset_triggered_at":       nil,
		},
	)

	return t, nil
}

func sessionsLoginPasswordHandler(ctx echo.Context) error {
	var (
		ua  UserAuth
		u   User
		err error
	)

	// Populate object from JSON
	err = ctx.Bind(&ua)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if ua.Username == "" {
		return ctx.JSON(http.StatusUnauthorized, gettext("An username is required", ctx))
	}

	u, err = ua.validate(ctx)

	// Rate limit
	if u.ID > 0 {
		if u.FailedLoginAttempts >= 5 {
			if u.LastLoginAttemptAt.Valid && time.Now().Before(u.LastLoginAttemptAt.Time.Add(time.Minute)) {
				db.Model(&u).Updates(map[string]interface{}{"last_login_attempt_at": time.Now()})

				return ctx.JSON(http.StatusTooManyRequests, gettext("You've made too many login attempts. Please wait a minute before trying again.", ctx))
			}
		}
	}

	if err != nil {
		if u.ID > 0 {
			db.Model(&u).Updates(map[string]interface{}{"last_login_attempt_at": time.Now(), "failed_login_attempts": u.FailedLoginAttempts + 1})
		}

		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	// if !u.hasVerifiedEmail() {
	// 	return ctx.JSON(http.StatusUnauthorized, gettext("You must verify your email before logging in.", ctx))
	// }

	token, err := createSession(ctx, u)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user":  u,
		"token": token,
	})
}

func sessionsLoginSendEmailHandler(ctx echo.Context) error {
	var (
		ua struct {
			Email        string `json:"email"`
			Host         string `json:"host"`
			CurrentPath  string `json:"current_path"`
			RedirectPath string `json:"redirect_path"`
		}
		u   User
		err error
	)

	// Populate object from JSON
	err = ctx.Bind(&ua)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if ua.Email == "" {
		return ctx.JSON(http.StatusUnauthorized, gettext("An email address is required", ctx))
	}

	if db.Where("email ILIKE ?", ua.Email).First(&u).RecordNotFound() {
		return ctx.JSON(http.StatusUnauthorized, gettext("No user found with this email address.", ctx))
	}

	go u.sendLoginEmail(u.Email, ua.Host, ua.CurrentPath, ua.RedirectPath)

	return ctx.JSON(http.StatusOK, gettext("Your login link has been emailed to you.", ctx))
}

func sessionsLoginEmailHandler(ctx echo.Context) error {
	var (
		ua struct {
			Email string `json:"email"`
			Token string `json:"token"`
		}
		u   User
		err error
	)

	message := gettext("Invalid login URL. This link has expired or there was an issue. You may use your password as a backup method. If you need assistance with our site, please contact support@sdxcentral.com.", ctx)

	// Populate object from JSON
	err = ctx.Bind(&ua)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if ua.Email == "" {
		return ctx.JSON(http.StatusUnauthorized, gettext("An email address is required", ctx))
	}

	if db.Where("email ILIKE ?", ua.Email).First(&u).RecordNotFound() {
		return ctx.JSON(http.StatusBadRequest, message)
	}

	if u.LoginEmailTokenHash == "" {
		return ctx.JSON(http.StatusBadRequest, message)
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.LoginEmailTokenHash), []byte(ua.Token))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, message)
	}

	if u.LoginEmailTriggeredAt.Valid && u.LoginEmailTriggeredAt.Time.AddDate(0, 0, 2).Before(time.Now()) {
		return ctx.JSON(http.StatusBadRequest, message)
	}

	t, err := createSession(ctx, u)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	u.verify()

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user":  u,
		"token": t,
	})
}

func sessionsLogoutHandler(ctx echo.Context) error {
	u, err := verifySession(ctx)
	if err == nil {
		db.Where("user_id = ?", u.ID).Delete(&Session{})
		db.Model(&u).Updates(map[string]interface{}{"last_logout_at": time.Now()})
	}

	return ctx.NoContent(http.StatusNoContent)
}
