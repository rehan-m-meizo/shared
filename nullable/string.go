package nullable

import (
	"database/sql/driver"
	"encoding/json"
)

type String struct {
	String string
	Valid  bool
}

func NewString(s string) String {
	return String{String: s, Valid: true}
}

func (ns *String) Scan(value interface{}) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}
	ns.String = value.(string)
	ns.Valid = true
	return nil
}

func (ns String) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

func (ns String) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

func (ns *String) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == nil {
		ns.String, ns.Valid = "", false
	} else {
		ns.String, ns.Valid = *s, true
	}
	return nil
}
