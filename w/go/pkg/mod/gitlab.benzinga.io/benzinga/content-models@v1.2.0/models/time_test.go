package models

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"
)

func TestTimeJSON(t *testing.T) {
	n := Time{}

	// Test marshal nil time
	data, err := n.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, `null`, string(data))

	// Test unmarshal nil time
	err = n.UnmarshalJSON([]byte(`null`))
	assert.Nil(t, err)
	assert.True(t, n.IsZero())

	// Test marshal arbitrary time
	n.Time, _ = time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	data, err = n.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, `"2006-01-02T15:04:05Z"`, string(data))

	// Test unmarshal arbitrary time
	err = n.UnmarshalJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, int64(1136214245000000000), n.UnixNano())
}

func TestTimeMsgPack(t *testing.T) {
	// Message size is constant.
	n := Time{}
	assert.Equal(t, 15, n.Msgsize())

	// Test zero value representation
	data, err := n.MarshalMsg([]byte{})
	assert.Nil(t, err)
	assert.Equal(t, []byte{
		0xc7, 0x0c, 0x05, 0xff, 0xff, 0xff, 0xf1, 0x88,
		0x6e, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00,
	}, data)

	// Test zero value unmarshal
	n2 := Time{}
	_, err = n2.UnmarshalMsg(data)
	assert.Nil(t, err)
	assert.Equal(t, n, n2)

	// Test arbitrary marshal+unmarshal
	n3 := Time{time.Now().In(time.Local)}
	n4 := Time{}

	data, err = n3.MarshalMsg([]byte{})
	assert.Nil(t, err)

	_, err = n4.UnmarshalMsg(data)
	assert.Nil(t, err)
	assert.Equal(t, n3, n4)
}

func TestTimeMsgPackStream(t *testing.T) {
	b := bytes.Buffer{}
	w := msgp.NewWriter(&b)

	n := Time{}
	err := n.EncodeMsg(w)
	assert.Nil(t, err)
	w.Flush()

	assert.Equal(t, []byte{
		0xc7, 0x0c, 0x05, 0xff, 0xff, 0xff, 0xf1, 0x88,
		0x6e, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00,
	}, b.Bytes())

	d := bytes.NewBuffer(b.Bytes())
	r := msgp.NewReader(d)

	n2 := Time{}
	err = n.DecodeMsg(r)
	assert.Nil(t, err)
	assert.Equal(t, n, n2)
}

func TestTimeFromFloat(t *testing.T) {
	tests := []struct {
		val float64
		rfc string
	}{
		{0, "0001-01-01T00:00:00Z"},
		{0.5, "1970-01-01T00:00:00.5Z"},
		{1, "1970-01-01T00:00:01Z"},
		{1463171895, "2016-05-13T20:38:15Z"},
	}

	for _, test := range tests {
		tim := TimeFromFloat(test.val)
		tim.Time = tim.Time.UTC()

		rfc := tim.Format(time.RFC3339Nano)
		assert.Equal(t, test.rfc, rfc)

		assert.InDelta(t, test.val, tim.Float(), 0.1)
	}
}

func TestMarshalOutOfRangeYear(t *testing.T) {
	tim := TimeFromFloat(253402300800).UTC()
	assert.Equal(t, 10000, tim.Year())
	_, err := tim.MarshalJSON()
	require.NotNil(t, err)
	assert.Equal(t, "Time.MarshalJSON: year outside of range [0,9999]", err.Error())
}
