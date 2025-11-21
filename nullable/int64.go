package nullable

import (
	"database/sql/driver"
	"encoding/json"
)

type Int64 struct {
	Int64 int64
	Valid bool
}

func NewInt64(i int64) Int64 {
	return Int64{Int64: i, Valid: true}
}

func (ni *Int64) Scan(value interface{}) error {
	if value == nil {
		ni.Int64, ni.Valid = 0, false
		return nil
	}
	ni.Int64 = value.(int64)
	ni.Valid = true
	return nil
}

func (ni Int64) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return ni.Int64, nil
}

func (ni Int64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	}
	return json.Marshal(nil)
}

func (ni *Int64) UnmarshalJSON(b []byte) error {
	var n *int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	if n == nil {
		ni.Int64, ni.Valid = 0, false
	} else {
		ni.Int64, ni.Valid = *n, true
	}
	return nil
}
