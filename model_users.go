package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Model

	FirstName        string `json:"first_name"`
	MiddleName       string `json:"middle_name"`
	LastName         string `json:"last_name"`
	Phone            string `gorm:"index" json:"phone"`
	Email            Email  `json:"email"`
	Gender           string `gorm:"index" json:"gender"`
	MaritalStatus    string `gorm:"index" json:"marital_status"`
	DOB              string `gorm:"index" json:"date_of_birth"`
	Caste            string `json:"caste"`
	SubCaste         string `gorm:"index" json:"sub_caste"`
	AnnualIncome     uint   `gorm:"index" json:"annual_income"`
	ProfileCreatedBy string `json:"profile_created_by"`

	UserData

	Password     string `gorm:"-" json:"password,omitempty"`
	PasswordHash string `json:"-"`

	Consented          bool         `json:"consented"`
	ConsentedDate      sql.NullTime `json:"consented_at"`
	ConsentedIPAddress string       `json:"consented_ip_address"`

	LastLoginAt         sql.NullTime `json:"last_login_at"`
	LastLoginAttemptAt  sql.NullTime `json:"-"`
	FailedLoginAttempts uint         `json:"-"`
	LastLogoutAt        sql.NullTime `json:"-"`
	LastLoginIPAddress  string       `json:"last_login_ip_address"`

	LoginEmailTokenHash   string       `json:"-"`
	LoginEmailTriggeredAt sql.NullTime `json:"-"`

	ResetTokenHash   string       `json:"-"`
	ResetTriggeredAt sql.NullTime `json:"-"`

	SubscribeRegister bool `json:"subscribe_register"`

	Language         string `json:"language"`
	IsVerified       bool   `json:"is_verified"`
	VerificationHash string `json:"-"`

	UserMedia       []Media       `gorm:"-" json:"user_media"`
	UserWallet      []Wallet      `gorm:"-" json:"user_wallet"`
	InterestDetails *UserInterest `gorm:"-" json:"interest_details"`
}

func (u *User) sanitize(ctx echo.Context) {

	u.FirstName = sanitizeText(u.FirstName, 64)
	u.MiddleName = sanitizeText(u.MiddleName, 64)
	u.LastName = sanitizeText(u.LastName, 64)
	u.Phone = sanitizeText(u.Phone, 12)

	u.UserData.sanitize(ctx)
	u.Language = getLanguageFromContext(ctx)
}

func (u *User) validate(ctx echo.Context, skipRequiredCheck bool) error {

	if u.FirstName == "" && !skipRequiredCheck {
		return errors.New("First Name is required")
	}

	if u.MiddleName == "" && !skipRequiredCheck {
		return errors.New("Middle Name is required")
	}

	if u.LastName == "" && !skipRequiredCheck {
		return errors.New("Last Name is required")
	}

	if u.Phone == "" {
		if !skipRequiredCheck {
			return errors.New("Phone Number is required")
		}
	} else {
		// implement regex to validate phone no
	}

	err := u.UserData.validate(ctx, skipRequiredCheck)
	if err != nil {
		return err
	}

	if len(u.Email) == 0 {
		if !skipRequiredCheck {
			return errors.New(gettext("An email address is required", ctx))
		}
	} else {
		err = u.Email.validate(ctx)
		if err != nil {
			return err
		}
	}

	// Check and hash password
	if u.Password != "" {
		err = checkPasswordStrength(ctx, u.Password)
		if err != nil {
			return err
		}

		pw, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
		if err != nil {
			u.PasswordHash = ""
		} else {
			u.PasswordHash = string(pw)
		}

		u.Password = ""
		u.FailedLoginAttempts = 0
		u.LoginEmailTokenHash = ""
		u.LoginEmailTriggeredAt.Valid = false
		u.ResetTokenHash = ""
		u.ResetTriggeredAt.Valid = false
	}

	return nil
}

func (u *User) exists() bool {
	var count int

	db.Model(User{}).Where("uuid::text != ?", u.UUID).Where("( email IN(?) OR phone = ? )", u.Email, u.Phone).Count(&count).Limit(1)

	return count > 0
}

func (u *User) getExistsMessage(ctx echo.Context) string {
	var (
		ce User
	)

	defaultMessage := gettext("A user with this email/mobile already exists", ctx)

	if db.Model(&User{}).Where("uuid::text != ?", u.UUID).Where("email IN(?)", u.Email).First(&ce).RecordNotFound() {
		return defaultMessage
	}

	return gettext("A user account was created for you when you subscribed to our newsletter. Reset your password to log in.", ctx)
}

func (u *User) doSearch(params SearchQuery) (interface{}, error) {
	var uu []User

	dbQuery := db

	dbQuery = u.query(dbQuery, params)
	params.setDefault()

	dbQuery = dbQuery.Order(params.OrderBy + " " + params.Order)

	dbQuery = dbQuery.Limit(params.Limit)

	dbQuery = dbQuery.Offset(params.Limit * (int(params.Page) - 1))

	err := dbQuery.Find(&uu).Error

	for i, u := range uu {
		var ui UserInterest
		db.Model(&Media{}).Where("user_uuid = ?", u.UUID).Find(&u.UserMedia)
		db.Model(&Wallet{}).Where("to_user_uuid = ? AND from_user_uuid = ? ", u.UUID, LoggedInUser.UUID).Find(&ui)
		db.Table("user_interests").Where("to_user_uuid = ? AND from_user_uuid = ? ", u.UUID, LoggedInUser.UUID).Find(&ui)

		// if ui.ID > 0 {
		u.InterestDetails = &ui
		// }

		uu[i] = u
	}

	return uu, err
}

func (u *User) restore(ctx echo.Context, tx *gorm.DB) error {
	var count uint

	tx.Model(User{}).Where("email ILIKE ?", u.Email).Count(&count).Limit(1)

	if count > 0 {
		return fmt.Errorf(gettext("An active user already exists with the email address: %s", ctx), u.Email)
	}

	tx.Unscoped().Model(&User{}).Where("id = ?", u.ID).Update("deleted_at", nil)

	return nil
}

func (u *User) bulkRead(uuids []string) (interface{}, error) {
	var (
		uu []User
	)

	db.Where("uuid IN(?)", uuids).Find(&uu)

	return uu, nil
}

func (u *User) AfterFind() error {

	return nil
}

func (u *User) AfterSave(tx *gorm.DB) error {

	return nil
}

func (u *User) AfterUpdate(tx *gorm.DB) error {
	return nil
}

func (u *User) AfterDelete(tx *gorm.DB) (err error) {
	return
}

func (u *User) getLoginURL(e Email, host string, currentPath string, redirectPath string) (string, error) {
	// Build URL
	loginURL := os.Getenv("VM_PROFILE_URL")

	host = sanitizeHost(host)

	// Create, hash, and save token
	token, err := generateRandomString(16)
	if err != nil {
		return loginURL, errors.New("Unable to create email login token")
	}

	hash, err := getPasswordHash(token)
	if err != nil {
		return loginURL, errors.New("Unable to hash email login token")
	}

	db.Model(&u).Updates(map[string]interface{}{"login_email_triggered_at": time.Now(), "login_email_token_hash": string(hash)})

	if currentPath != "" {
		testURL, err := url.ParseRequestURI("https://" + host + currentPath)
		if err == nil {
			loginURL = "https://" + host + testURL.Path
		}
	}

	targetURL, err := url.ParseRequestURI(loginURL)
	if err != nil {
		return loginURL, errors.New("Unable to build login URL")
	}

	q := targetURL.Query()
	q.Set("email", string(u.Email))
	q.Set("login_token", token)

	if redirectPath != "" {
		redirect, err := url.ParseRequestURI("https://" + host + redirectPath)
		if err == nil {
			q.Set("redirect_to", redirect.Path)
		}
	}

	targetURL.RawQuery = q.Encode()

	return targetURL.String(), nil
}

func (u *User) sendLoginEmail(e Email, host string, currentPath string, redirectPath string) error {
	url, err := u.getLoginURL(e, host, currentPath, redirectPath)
	if err != nil {
		return err
	}

	// Send email
	vars := make(map[string]string)
	vars["login_url"] = url

	go u.notify("user-verify-link", string(e), vars, nil)

	return nil
}

func (u User) getKnownIPs() []string {
	var ss []string

	return ss
}

func (u *User) verify() {
	if u.IsVerified {
		return
	}

	db.Model(&u).Updates(map[string]interface{}{"is_verified": true})
}

func (u *User) getName() string {
	n := fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	n = strings.TrimSpace(n)
	return n
}

func (u *User) hasVerifiedEmail() bool {
	return u.IsVerified
}

func (u User) Count(ctx echo.Context) error {
	var (
		resp struct {
			Count uint   `json:"count"`
			Error string `json:"error"`
		}
		params SearchQuery
	)

	u, err := verifySession(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	} else {
		LoggedInUser = u
	}

	// Populate object from JSON
	err = ctx.Bind(&params)
	if err != nil {
		return returnInvalidData(ctx, err)
	}

	dbQuery := db.Model(&User{})
	dbQuery = u.query(dbQuery, params)

	err = dbQuery.Select("COUNT(id) count").Scan(&resp).Error

	if err != nil {
		resp.Error = err.Error()
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (u User) query(dbQuery *gorm.DB, params SearchQuery) *gorm.DB {

	dbQuery = dbQuery.Where("uuid != ?", LoggedInUser.UUID)

	if params.FromAge > 18 {
		dbQuery = dbQuery.Where("DATE_PART('Year', NOW()) - DATE_PART('Year', dob::date) >= ?", params.FromAge)
	}

	if params.ToAge > 18 {
		dbQuery = dbQuery.Where("DATE_PART('Year', NOW()) - DATE_PART('Year', dob::date) <= ?", params.ToAge)
	}

	if params.IncomeFrom > 0 {
		dbQuery = dbQuery.Where("annual_income >= ?", params.IncomeFrom)
	}

	if params.IncomeTo > 0 {
		dbQuery = dbQuery.Where("annual_income <= ?", params.IncomeTo)
	}

	if len(params.Gender) > 0 {
		dbQuery = dbQuery.Where("gender = ?", params.Gender)
	}

	if len(params.Caste) > 0 {
		dbQuery = dbQuery.Where("caste = ?", params.Caste)
	}

	if len(params.SubCastes) > 0 {
		dbQuery = dbQuery.Where("sub_caste IN (?)", params.SubCastes)
	}

	if len(params.MaritalStatuses) > 0 {
		dbQuery = dbQuery.Where("marital_status IN (?)", params.MaritalStatuses)
	}

	if len(params.Educations) > 0 {
		dbQuery = dbQuery.Where("educational_info->>'education' IN (?)", params.Educations)
	}

	if len(params.Cities) > 0 {
		dbQuery = dbQuery.Where("city IN (?)", params.Cities)
	}

	if len(params.Districts) > 0 {
		dbQuery = dbQuery.Where("district IN (?)", params.Districts)
	}

	if len(params.States) > 0 {
		dbQuery = dbQuery.Where("State IN (?)", params.States)
	}

	if len(params.Countries) > 0 {
		dbQuery = dbQuery.Where("country IN (?)", params.Countries)
	}

	return dbQuery
}
