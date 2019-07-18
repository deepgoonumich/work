package models

import (
	"bytes"
	"errors"
	"math"
	"time"

	"github.com/tinylib/msgp/msgp"
)

// Time is a wrapper around time.Time with different serialization behavior.
type Time struct{ time.Time }

// UnmarshalJSON implements json.Unmarshaler
func (t *Time) UnmarshalJSON(data []byte) (err error) {
	// Handle empty cases gracefully.
	if bytes.Equal(data, []byte(`null`)) || bytes.Equal(data, []byte(`""`)) {
		*t = Time{}
		return nil
	}

	t.Time, err = time.Parse(`"`+time.RFC3339Nano+`"`, string(data))
	return
}

// MarshalJSON implements json.Marshaler
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte(t.Format(`null`)), nil
	}

	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	return []byte(t.Format(`"` + time.RFC3339Nano + `"`)), nil
}

// TimeFromFloat converts a floating-point UNIX timestamp to a Time object.
func TimeFromFloat(ts float64) Time {
	if ts == 0 {
		return Time{}
	}

	i, f := math.Modf(ts)
	return Time{time.Unix(int64(i), int64(f*1000000000.0))}
}

// Float converts a Time object to a floating-point UNIX timestamp.
func (t Time) Float() float64 {
	if t.IsZero() {
		return 0
	}

	return float64(t.UnixNano()) / float64(1000000000.0)
}

// DecodeMsg implements msgp.Decodable
func (t *Time) DecodeMsg(dc *msgp.Reader) (err error) {
	t.Time, err = dc.ReadTime()

	// Handle nil time properly.
	if t.Time.IsZero() {
		t.Time = time.Time{}
	}

	return
}

// EncodeMsg implements msgp.Encodable
func (t Time) EncodeMsg(en *msgp.Writer) error {
	return en.WriteTime(t.Time)
}

// MarshalMsg implements msgp.Marshaler
func (t Time) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, t.Msgsize())
	o = msgp.AppendTime(o, t.Time)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (t *Time) UnmarshalMsg(bts []byte) (o []byte, err error) {
	t.Time, o, err = msgp.ReadTimeBytes(bts)

	// Handle nil time properly.
	if t.Time.IsZero() {
		t.Time = time.Time{}
	}

	return
}

// Msgsize implements msgp.Sizer
func (t Time) Msgsize() int {
	return msgp.TimeSize
}
