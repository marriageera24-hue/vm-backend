package main

import (
	"database/sql/driver"
	"encoding/json"
)

type FamilyInfo map[string]interface{}

func (info *FamilyInfo) Scan(value interface{}) error {
	asBytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	err := json.Unmarshal(asBytes, &info)

	return err
}

func (info FamilyInfo) Value() (driver.Value, error) {
	return json.Marshal(info)
}
