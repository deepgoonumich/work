package models

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinylib/msgp/msgp"
)

// These are some helpers for asserting that the generated code works.
// You probably don't want to modify this if your tests are failing.

// AssertRoundTripSafe ensures the object passed in can survive roundtrip.
func AssertRoundTripSafe(t *testing.T, x interface{}) {
	JSONRoundTrip(t, x)
	BSONRoundTrip(t, x)
	MSGPRoundTrip(t, x)
}

// JSONRoundTrip tests JSON serializability and deserializability on an object
func JSONRoundTrip(t *testing.T, x interface{}) {
	data, err := json.Marshal(x)
	require.NoError(t, err)

	y := reflect.New(reflect.TypeOf(x).Elem()).Interface()
	err = json.Unmarshal(data, y)
	require.NoError(t, err)

	assert.EqualValues(t, x, reflect.ValueOf(y).Interface())
}

// BSONRoundTrip tests BSON serializability and deserializability on an object
func BSONRoundTrip(t *testing.T, x interface{}) {
	data, err := bson.Marshal(x)
	require.NoError(t, err)

	y := reflect.New(reflect.TypeOf(x).Elem()).Interface()
	err = bson.Unmarshal(data, y)
	require.NoError(t, err)

	assert.EqualValues(t, x, reflect.ValueOf(y).Interface())
}

// MSGPRoundTrip tests MSGP serializability and deserailizability on an object
func MSGPRoundTrip(t *testing.T, x interface{}) {
	data, err := x.(msgp.Marshaler).MarshalMsg(nil)
	require.NoError(t, err)

	y := reflect.New(reflect.TypeOf(x).Elem()).Interface()
	_, err = y.(msgp.Unmarshaler).UnmarshalMsg(data)
	require.NoError(t, err)

	// assert.EqualValues(t, x, reflect.ValueOf(y).Interface())

	xj, err := json.Marshal(&x)
	require.NoError(t, err)
	yj, err := json.Marshal(&y)
	require.NoError(t, err)
	assert.Equal(t, xj, yj)
}

// IsMsgpable is an interface that allows one to statically assert that a type
// has `msgp` generated code associated with it.
type IsMsgpable interface {
	msgp.Marshaler
	msgp.Unmarshaler
	msgp.Encodable
	msgp.Decodable
	msgp.Sizer
}
