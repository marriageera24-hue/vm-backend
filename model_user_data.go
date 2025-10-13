package main

import (
	"errors"

	"github.com/labstack/echo"
)

type UserData struct {
	OrganizationName string `json:"organization_name"`
	Industry         string `json:"industry"`
	JobTitle         string `json:"job_title"`

	Address1   string `json:"address_1"`
	Address2   string `json:"address_2"`
	City       string `json:"city"`
	District   string `json:"district"`
	State      string `json:"State"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`

	Unsubscribed bool `json:"unsubscribed"`

	FamilyInfo      FamilyInfo      `gorm:"type:jsonb" json:"family_info"`
	EducationalInfo EducationalInfo `gorm:"type:jsonb" json:"educationl_info"`
	OtherInfo       OtherInfo       `gorm:"type:jsonb" json:"other_info"`
}

func (u *UserData) sanitize(ctx echo.Context) {
	u.OrganizationName = ""
	u.Industry = ""

	u.JobTitle = sanitizeText(u.JobTitle, 100)

	u.Country = sanitizeText(u.Country, 100)

	u.Address1 = sanitizeText(u.Address1, 500)
	u.Address2 = sanitizeText(u.Address2, 500)
	u.City = sanitizeText(u.City, 100)
	u.District = sanitizeText(u.District, 100)

	u.State = sanitizeText(u.State, 100)

	u.PostalCode = sanitizeText(u.PostalCode, 16)
}

func (u *UserData) validate(ctx echo.Context, skipRequiredCheck bool) error {

	if u.JobTitle == "" && !skipRequiredCheck {
		return errors.New("Job Title is required")
	}

	if u.Country == "" && !skipRequiredCheck {
		return errors.New("Country is required")
	}

	if u.City == "" && !skipRequiredCheck {
		return errors.New("City is required")
	}

	// if u.District == "" && !skipRequiredCheck {
	// 	return errors.New("District is required")
	// }

	if u.State == "" && !skipRequiredCheck {
		return errors.New("State / Province is required")
	}

	if u.PostalCode == "" && !skipRequiredCheck {
		return errors.New("Postal Code is required")
	}

	return nil
}
