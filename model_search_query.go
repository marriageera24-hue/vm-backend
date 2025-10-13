package main

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type SearchQuery struct {
	Query           string   `json:"q"`
	Gender          string   `json:"gender"`
	Caste           string   `json:"caste"`
	SubCastes       []string `json:"sub_castes"`
	FromAge         uint     `json:"from_age"`
	ToAge           uint     `json:"to_age"`
	MaritalStatuses []string `json:"marital_statuses"`
	Cities          []string `json:"cities"`
	Districts       []string `json:"districts"`
	States          []string `json:"states"`
	Countries       []string `json:"countries"`
	IncomeFrom      float64  `json:"income_from"`
	IncomeTo        float64  `json:"income_to"`
	Professions     []string `json:"profession"`
	Educations      []string `json:"educations"`
	Order           string   `json:"order"`
	OrderBy         string   `json:"order_by"`
	Page            uint     `json:"page"`
	Limit           int      `json:"limit"`

	UserUUID []string `json:"user_uuids"`
}

func (sq *SearchQuery) setDefault() {
	if strings.ToUpper(sq.Order) != "DESC" {
		sq.Order = "ASC"
	}

	if len(sq.OrderBy) == 0 {
		sq.OrderBy = "id"
	}

	sq.Limit = 10
}

func (sq *SearchQuery) Scan(value interface{}) error {
	asBytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	err := json.Unmarshal(asBytes, &sq)

	return err
}

func (sq SearchQuery) Value() (driver.Value, error) {

	return json.Marshal(sq)
}
