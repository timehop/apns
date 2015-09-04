package apns

import (
	"errors"
	"strconv"
)

// BadgeNumber is much a NullInt64 in
// database/sql except instead of using
// the nullable value for driver.Value
// encoding and decoding, this is specifically
// meant for JSON encoding and decoding
type BadgeNumber struct {
	Number int
	Valid  bool
}

// NewBadgeNumber returns a valid BadgeNumber
// with the value passed in
func NewBadgeNumber(number int) BadgeNumber {
	return BadgeNumber{
		Number: number,
		Valid:  true,
	}
}

// Unset will reset the BadgeNumber to
// both 0 and invalid
func (b *BadgeNumber) Unset() {
	b.Number = 0
	b.Valid = false
}

// Set will set the BadgeNumber value to
// the number passed in, assuming it's >= 0.
// If so, the BadgeNumber will also be marked valid
func (b *BadgeNumber) Set(number int) error {
	if number < 0 {
		return errors.New("Number must be >= 0")
	}

	b.Number = number
	b.Valid = true
	return nil
}

// MarshalJSON will marshall the numerical value of
// BadgeNumber
func (b BadgeNumber) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(b.Number)), nil
}

// UnmarshalJSON will take any non-nil value and
// set BadgeNumber's numeric value to it. It assumes
// that if the unmarshaller gets here, there is a
// number to unmarshal and it's valid
func (b *BadgeNumber) UnmarshalJSON(data []byte) error {
	if val, err := strconv.ParseInt(string(data), 10, 32); err != nil {
		return err
	} else {
		b.Number = int(val)

		// Since the point of this type is to
		// allow proper inclusion of 0 for int
		// types while respecting omitempty,
		// assume that set==true if there is
		// a value to unmarshal
		b.Valid = true
		return nil
	}
}
