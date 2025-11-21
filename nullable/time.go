package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Time struct {
	Time  time.Time
	Valid bool
}

func NewTime(t time.Time) Time {
	return Time{Time: t, Valid: true}
}

func (nt *Time) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	nt.Time = value.(time.Time)
	nt.Valid = true
	return nil
}

func (nt Time) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func (nt Time) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.Time)
	}
	return json.Marshal(nil)
}

func (nt *Time) UnmarshalJSON(b []byte) error {
	var t *time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	if t == nil {
		nt.Time, nt.Valid = time.Time{}, false
	} else {
		nt.Time, nt.Valid = *t, true
	}
	return nil
}
