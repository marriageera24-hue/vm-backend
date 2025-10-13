package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/tomasen/realip"
)

func userGetHandler(ctx echo.Context) error {
	var u User

	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	db.Model(&Media{}).Where("user_uuid = ?", u.UUID).Find(&u.UserMedia)
	db.Model(&Wallet{}).Where("user_uuid = ?", u.UUID).Find(&u.UserWallet)
	return ctx.JSON(http.StatusOK, u)
}

func userDeleteHandler(ctx echo.Context) error {
	var (
		u   User
		err error
	)

	u, err = verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	db.Model(&u).Updates(map[string]interface{}{"deleted_at": time.Now()})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, gettext("User deleted", ctx))
}
func userRegisterHandler(ctx echo.Context) error {
	var (
		data struct {
			User
			// Email        Email  `json:"email"`
			// Phone        string `json:"phone"`
			Host         string `json:"host"`
			CurrentPath  string `json:"current_path"`
			RedirectPath string `json:"redirect_path"`
		}
		u User
	)

	// _, err := verifySession(ctx)
	// if err == nil {
	// 	return ctx.JSON(http.StatusUnauthorized, gettext("You are already logged in", ctx))
	// }

	// Populate object from JSON
	err := ctx.Bind(&data)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if data.Phone == "" {
		return ctx.JSON(http.StatusBadRequest, "Phone is required")
	}

	if len(data.Phone) < 10 && len(data.Phone) > 12 && data.Email == "" {
		return ctx.JSON(http.StatusBadRequest, gettext("Email is required", ctx))
	}

	u = data.User

	u.Email = data.Email
	u.Phone = data.Phone

	if u.exists() {
		return ctx.JSON(http.StatusConflict, gettext("Email or Phone is already registered", ctx))
	}

	if len(u.Password) < 1 {
		p, err := generateRandomString(16)
		if err == nil {
			u.Password = p
		}
	}

	u.IsVerified = true
	u.Consented = true
	u.ConsentedDate.Valid = true
	u.ConsentedDate.Time = time.Now()
	u.ConsentedIPAddress = realip.FromRequest(ctx.Request())

	if strings.Contains(u.FirstName, "https://") {
		return ctx.JSON(http.StatusOK, u)
	}

	if strings.Contains(u.FirstName, "http://") {
		return ctx.JSON(http.StatusOK, u)
	}

	if strings.Contains(u.LastName, "http://") {
		return ctx.JSON(http.StatusOK, u)
	}

	if strings.Contains(u.LastName, "http://") {
		return ctx.JSON(http.StatusOK, u)
	}

	tx := db.Begin()

	_, err = apiCreate(ctx, tx, &u, false)
	token, e2 := createSession(ctx, u)
	if err != nil || e2 != nil {
		tx.Rollback()

		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	tx.Commit()

	// vars := make(map[string]string)
	// if len(u.Email) > 0 {
	// 	vars["user_email"] = string(u.Email)
	// 	url, err := u.getLoginURL(u.Email, data.Host, data.CurrentPath, data.RedirectPath)
	// 	if err == nil {
	// 		vars["login_url"] = url
	// 	}
	// }

	// go u.notify("user-welcome", data.Email, vars, nil)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user":  u,
		"token": token,
	})
}

func userUpdateHandler(ctx echo.Context) error {
	var (
		data struct {
			User
			Password  string `json:"password"`
			Password2 string `json:"password_2"`
			Host      string `json:"host"`
		}
		old User
	)

	new, err := verifySession(ctx)

	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, gettext("You are not authorized to do this", ctx))
	}

	bb, err := json.Marshal(new)
	if err == nil {
		json.Unmarshal(bb, &old)
	}
	old.ID = new.ID

	// Populate object from JSON
	err = ctx.Bind(&data)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if data.Password != "" {
		if data.Password != data.Password2 {
			return ctx.JSON(http.StatusBadRequest, gettext("Your passwords don't match", ctx))
		}

		new.Password = data.Password
	}

	new = data.User
	new.ID = old.ID

	if !new.Consented {
		new.Consented = true
		new.ConsentedDate.Valid = true
		new.ConsentedDate.Time = time.Now()
		new.ConsentedIPAddress = realip.FromRequest(ctx.Request())
	}

	tx := db.Begin()

	_, err = apiUpdate(ctx, tx, &new, false)
	if err != nil {
		tx.Rollback()

		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	tx.Commit()

	return ctx.JSON(http.StatusOK, new)
}

func userPatchHandler(ctx echo.Context) error {
	var (
		data struct {
			BusinessEmail string `json:"business_email"`
			Host          string `json:"host"`
			CurrentPath   string `json:"current_path"`
			User
		}
		old User
	)

	new, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, gettext("You are not authorized to do this", ctx))
	}

	bb, err := json.Marshal(new)
	if err == nil {
		json.Unmarshal(bb, &old)
	}

	data.User = new

	// Populate object from JSON
	err = ctx.Bind(&data)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	new = data.User

	if !new.Consented {
		new.Consented = true
		new.ConsentedDate.Valid = true
		new.ConsentedDate.Time = time.Now()
		new.ConsentedIPAddress = realip.FromRequest(ctx.Request())
	}

	tx := db.Begin()

	_, err = apiUpdate(ctx, tx, &new, false)
	if err != nil {
		tx.Rollback()

		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	tx.Commit()

	return ctx.JSON(http.StatusOK, new)
}

func userVerifyHandler(ctx echo.Context) error {
	var (
		u   User
		err error
		// data struct {
		// 	Email string `json:"email"`
		// 	Token string `json:"token"`
		// }
		data struct {
			Uuid string `json:"uuid"`
		}
	)

	// Populate object from JSON
	err = ctx.Bind(&data)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	// if db.Where("email ILIKE ?", data.Email).First(&u).RecordNotFound() {
	// 	return ctx.JSON(http.StatusBadRequest, gettext("User not found", ctx))
	// }

	if u.IsVerified {
		return ctx.JSON(http.StatusOK, gettext("You have already verified this user", ctx))
	}

	// err = bcrypt.CompareHashAndPassword([]byte(u.VerificationHash), []byte(data.Token))
	// if err != nil {

	// 	return ctx.JSON(http.StatusBadRequest, gettext("Verification token is invalid", ctx))
	// }

	// Start with db.Model(&User{}) to ensure a clean slate
	result := db.Exec("UPDATE users SET is_verified = ? WHERE uuid = ?", true, data.Uuid)
	if result.Error != nil {
		// There was a database execution error.
		return ctx.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return ctx.JSON(http.StatusOK, gettext("User has been verified", ctx))
}

func userUnverifyHandler(ctx echo.Context) error {
	var (
		u   User
		err error
		// data struct {
		// 	Email string `json:"email"`
		// 	Token string `json:"token"`
		// }
		data struct {
			Uuid string `json:"uuid"`
		}
	)

	// Populate object from JSON
	err = ctx.Bind(&data)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	// if db.Where("email ILIKE ?", data.Email).First(&u).RecordNotFound() {
	// 	return ctx.JSON(http.StatusBadRequest, gettext("User not found", ctx))
	// }

	if u.IsVerified {
		return ctx.JSON(http.StatusOK, gettext("You have already verified this user", ctx))
	}

	// err = bcrypt.CompareHashAndPassword([]byte(u.VerificationHash), []byte(data.Token))
	// if err != nil {

	// 	return ctx.JSON(http.StatusBadRequest, gettext("Verification token is invalid", ctx))
	// }

	// Start with db.Model(&User{}) to ensure a clean slate
	result := db.Exec("UPDATE users SET is_verified = ? WHERE uuid = ?", false, data.Uuid)
	if result.Error != nil {
		// There was a database execution error.
		return ctx.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return ctx.JSON(http.StatusOK, gettext("User has been verified", ctx))
}

func userSendVerifyHandler(ctx echo.Context) error {
	var (
		e    User
		err  error
		data struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
			Host  string `json:"host"`
		}
	)

	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, gettext("You are not authorized to do this", ctx))
	}

	reqType := ctx.Param("type") // tipe of verification request

	// Populate object from JSON
	err = ctx.Bind(&data)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	if reqType == "email" {

		if db.Where("id = ?", u.ID).Where("email ILIKE ?", data.Email).First(&e).RecordNotFound() {
			return ctx.JSON(http.StatusBadRequest, gettext("Email not found", ctx))
		}

		if e.IsVerified {
			return ctx.JSON(http.StatusOK, gettext("You have already verified this email address", ctx))
		}

		u.sendVerificationEmail(u.Email, data.Host, "")

		return ctx.JSON(http.StatusOK, gettext("Verification link has been sent", ctx))
	} else if reqType == "otp" {

		if db.Where("id = ?", u.ID).Where("phone ILIKE ?", data.Phone).First(&e).RecordNotFound() {
			return ctx.JSON(http.StatusBadRequest, gettext("Phone not found", ctx))
		}

		err = generateOTP(u, data.Phone)

		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		return ctx.JSON(http.StatusOK, gettext("Verification generated successfully", ctx))
	}

	return ctx.JSON(http.StatusBadRequest, gettext("Invalid request", ctx))
}

func userExistsHandler(ctx echo.Context) error {
	var u User
	i := ctx.Param("info")

	if db.Where("(uuid::text = ? ) OR (email = ?) OR (phone = ?) ", i, i, i).First(&u).RecordNotFound() {
		return ctx.NoContent(http.StatusNotFound)
	}

	return ctx.NoContent(http.StatusOK)
}

func userResubscribeHandler(ctx echo.Context) error {
	var old User

	new, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, gettext("You are not authorized to do this", ctx))
	}

	bb, err := json.Marshal(new)
	if err == nil {
		json.Unmarshal(bb, &old)
	}

	new.Unsubscribed = false

	tx := db.Begin()

	_, err = apiUpdate(ctx, tx, &new, false)
	if err != nil {
		tx.Rollback()

		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	tx.Commit()

	return ctx.JSON(http.StatusOK, new)
}
