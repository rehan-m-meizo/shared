package nullable

import (
	"database/sql/driver"
	"encoding/json"
)

type Bool struct {
	Bool  bool
	Valid bool
}

func NewBool(b bool) Bool {
	return Bool{Bool: b, Valid: true}
}

func (nb *Bool) Scan(value interface{}) error {
	if value == nil {
		nb.Bool, nb.Valid = false, false
		return nil
	}
	nb.Bool = value.(bool)
	nb.Valid = true
	return nil
}

func (nb Bool) Value() (driver.Value, error) {
	if !nb.Valid {
		return nil, nil
	}
	return nb.Bool, nil
}

func (nb Bool) MarshalJSON() ([]byte, error) {
	if nb.Valid {
		return json.Marshal(nb.Bool)
	}
	return json.Marshal(nil)
}

func (nb *Bool) UnmarshalJSON(b []byte) error {
	var boolVal *bool
	if err := json.Unmarshal(b, &boolVal); err != nil {
		return err
	}
	if boolVal == nil {
		nb.Bool, nb.Valid = false, false
	} else {
		nb.Bool, nb.Valid = *boolVal, true
	}
	return nil
}
