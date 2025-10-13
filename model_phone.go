package main

import (
	"database/sql/driver"
	"encoding/json"
)

type Phone struct {
	CountryCode         string `json:"country_code"`
	Number              string `json:"number"`
	NationalFormat      string `json:"national_format"`
	InternationalFormat string `json:"international_format"`
	E164Format          string `json:"e164_format"`
}

func (p *Phone) Scan(value interface{}) error {
	asBytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	err := json.Unmarshal(asBytes, &p)

	return err
}

func (p Phone) Value() (driver.Value, error) {
	if p.Number == "" {
		return nil, nil
	}

	return json.Marshal(p)
}
