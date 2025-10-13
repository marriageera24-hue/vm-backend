package main

import (
	"database/sql/driver"
	"encoding/json"
)

type EducationalInfo map[string]interface{}

func (info *EducationalInfo) Scan(value interface{}) error {
	asBytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	err := json.Unmarshal(asBytes, &info)

	return err
}

func (info EducationalInfo) Value() (driver.Value, error) {
	return json.Marshal(info)
}
