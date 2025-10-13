package main

import (
	"errors"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

type EmailTemplate struct {
	Model
	Language  string `json:"language"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
	Subject   string `json:"subject"`
	Preheader string `json:"preheader"`
	Body      string `json:"body"`
}

func getEmailTemplateSlugs() []string {
	return []string{"user-email-added", "user-email-removed", "user-email-verify", "user-forgot-password", "user-login-link", "user-password-changed", "user-update-personal-email", "user-verify-link", "user-welcome", "user-subscribe-welcome"}
}

func (e *EmailTemplate) sanitize(ctx echo.Context) {
	e.Name = sanitizeText(e.Name, 64)
	e.FromName = sanitizeText(e.FromName, 255)
	e.FromEmail = strings.ToLower(e.FromEmail)
	e.FromEmail = sanitizeText(e.FromEmail, 255)
	e.Subject = sanitizeText(e.Subject, 78)
	e.Preheader = sanitizeText(e.Preheader, 255)
	e.Body = sanitizeText(e.Body, 0)
}

func (e *EmailTemplate) validate(ctx echo.Context, skipRequiredCheck bool) error {
	if !isOneOf(e.Language, getSupportedLanguagesStrings()) {
		return errors.New("Language is invalid")
	}

	if !isOneOf(e.Slug, getEmailTemplateSlugs()) {
		return errors.New("Slug is invalid")
	}

	if e.Name == "" && !skipRequiredCheck {
		return errors.New("Name is required")
	}

	if e.FromName == "" && !skipRequiredCheck {
		return errors.New("From Name is required")
	}

	if e.FromEmail == "" && !skipRequiredCheck {
		return errors.New("From Email is required")
	}

	if e.FromEmail != "" {
		if !govalidator.IsEmail(e.FromEmail) {
			return errors.New("From Email is invalid")
		}
	}

	if e.Subject == "" && !skipRequiredCheck {
		return errors.New("Subject is required")
	}

	if e.Body == "" && !skipRequiredCheck {
		return errors.New("Subject is required")
	}

	return nil
}

func (e *EmailTemplate) exists() bool {
	var count int

	db.Model(EmailTemplate{}).Where("id != ?", e.ID).Where("language = ?", e.Language).Where("slug = ?", e.Slug).Count(&count).Limit(1)

	return count > 0
}

func (e *EmailTemplate) getExistsMessage(ctx echo.Context) string {
	return "An email template with in this language this slug already exists"
}

func (e *EmailTemplate) doSearch(params SearchQuery) (interface{}, error) {
	var ee []EmailTemplate

	dbQuery := db

	if params.Query != "" {
		dbQuery = dbQuery.Where("uuid::text = ?", params.Query)
		dbQuery = dbQuery.Where("slug = ?", params.Query)
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+params.Query+"%")
		dbQuery = dbQuery.Where("subject ILIKE ?", "%"+params.Query+"%")
	}

	dbQuery = dbQuery.Order(params.OrderBy + " " + params.Order)

	// dbQuery = dbQuery.Limit(params.Limit + 1)

	// dbQuery = dbQuery.Offset(params.Limit * (int(params.Page) - 1))

	err := dbQuery.Find(&ee).Error

	return ee, err
}

func (e *EmailTemplate) restore(ctx echo.Context, tx *gorm.DB) error {
	tx.Unscoped().Model(&EmailTemplate{}).Where("id = ?", e.ID).Update("deleted_at", nil)

	return nil
}

func (e *EmailTemplate) bulkRead(slugs []string) (interface{}, error) {
	var ee []EmailTemplate

	db.Where("slug IN(?)", slugs).Find(&ee)

	return ee, nil
}

func maybeCreateEmailTemplates() {
	var tt []EmailTemplate

	db.Find(&tt)

	for _, l := range getSupportedLanguagesStrings() {
		for _, s := range getEmailTemplateSlugs() {
			found := false

			for _, t := range tt {
				if t.Language != l {
					continue
				}

				if t.Slug != s {
					continue
				}

				found = true
			}

			if !found {
				e := EmailTemplate{}
				e.Language = l
				e.Slug = s
				apiCreate(nil, db, &e, true)
			}
		}
	}
}
