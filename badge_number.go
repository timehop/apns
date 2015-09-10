package apns

import "encoding/json"

// BadgeNumber is much a NullInt64 in
// database/sql except instead of using
// the nullable value for driver.Value
// encoding and decoding, this is specifically
// meant for JSON encoding and decoding
type BadgeNumber struct {
	Number uint
	IsSet  bool
}

// Unset will reset the BadgeNumber to
// both 0 and invalid
func (b *BadgeNumber) Unset() {
	b.Number = 0
	b.IsSet = false
}

// Set will set the BadgeNumber value to
// the number passed in, assuming it's >= 0.
// If so, the BadgeNumber will also be marked valid
func (b *BadgeNumber) Set(number uint) {
	b.Number = number
	b.IsSet = true
}

// MarshalJSON will marshall the numerical value of
// BadgeNumber
func (b BadgeNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Number)
}

// UnmarshalJSON will take any non-nil value and
// set BadgeNumber's numeric value to it. It assumes
// that if the unmarshaller gets here, there is a
// number to unmarshal and it's valid
func (b *BadgeNumber) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &b.Number)
	if err != nil {
		return err
	}

	// Since the point of this type is to
	// allow proper inclusion of 0 for int
	// types while respecting omitempty,
	// assume that set==true if there is
	// a value to unmarshal
	b.IsSet = true
	return nil
}
