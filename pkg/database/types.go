package database

import (
	"database/sql/driver"
	"encoding/json"
)

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v := value.(string)
	return json.Unmarshal([]byte(v), s)
}
