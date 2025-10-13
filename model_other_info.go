package main

import (
	"database/sql/driver"
	"encoding/json"
)

type OtherInfo map[string]interface{}

func (info *OtherInfo) Scan(value interface{}) error {
	asBytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	err := json.Unmarshal(asBytes, &info)

	return err
}

func (info OtherInfo) Value() (driver.Value, error) {
	return json.Marshal(info)
}
