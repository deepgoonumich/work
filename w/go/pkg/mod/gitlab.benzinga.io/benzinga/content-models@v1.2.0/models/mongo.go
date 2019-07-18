package models

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

type idWrapper struct {
	ID bson.ObjectId `json:"$id"`
}

type drupalField struct {
	Value interface{} `json:"value" bson:"value"`
}

// PHPObjID is a Mongo object ID as represented by PHP mongo.
type PHPObjID bson.ObjectId

// UnmarshalJSON overrides bson.ObjectId's JSON unmarshaller, to support
// different formatting for object IDs from PHP. e.g.:
//   { "$id": "..." }
func (id *PHPObjID) UnmarshalJSON(data []byte) error {
	if len(data) == 26 {
		return (*bson.ObjectId)(id).UnmarshalJSON(data)
	}

	obj := idWrapper{}
	err := json.Unmarshal(data, &obj)
	*id = PHPObjID(obj.ID)
	return err
}

func tryStr(v interface{}) string {
	s, _ := v.(string)
	return s
}

func tryInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float32:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}

// MarshalJSON overrides bson.ObjectID's JSON marshaller, to output back to the
// PHP mongo format.
func (id PHPObjID) MarshalJSON() ([]byte, error) {
	return json.Marshal(idWrapper{bson.ObjectId(id)})
}

// GetBSON implements bson.Getter
func (id PHPObjID) GetBSON() (interface{}, error) {
	return bson.ObjectId(id), nil
}

// Sentiment represents a drupal field of type int.
type Sentiment int

// UnmarshalJSON implements json.Unmarshaler
func (n *Sentiment) UnmarshalJSON(data []byte) error {
	obj := []drupalField{}
	err := json.Unmarshal(data, &obj)
	if err != nil || len(obj) == 0 {
		*n = 0
		// TODO: should we return the error?
		return nil
	}
	*n = Sentiment(tryInt(obj[0].Value))
	return nil
}

// MarshalJSON implements json.Marshaler
func (n Sentiment) MarshalJSON() ([]byte, error) {
	return json.Marshal([]drupalField{{Value: int(n)}})
}

// SetBSON implements bson.Setter
func (n *Sentiment) SetBSON(raw bson.Raw) error {
	// This only marginally less horrible than writing our own BSON library.
	obj := struct {
		Z drupalField `bson:"0"`
	}{}

	err := bson.Unmarshal(raw.Data, &obj)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		*n = 0
		return nil
	}
	*n = Sentiment(tryInt(obj.Z.Value))
	return nil
}

// GetBSON implements bson.Getter
func (n Sentiment) GetBSON() (interface{}, error) {
	return []drupalField{{Value: int(n)}}, nil
}

// DrupalInt represents a drupal field of type int.
type DrupalInt int

// UnmarshalJSON implements json.Unmarshaler
func (n *DrupalInt) UnmarshalJSON(data []byte) error {
	obj := drupalField{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		*n = 0
		// TODO: should we return the error?
		fmt.Println(err)
		return nil
	}

	if str := tryStr(obj.Value); str != "" {
		if num, err := strconv.Atoi(str); err == nil {
			*n = DrupalInt(num)
			return nil
		}
	}

	*n = DrupalInt(tryInt(obj.Value))
	return nil
}

// MarshalJSON implements json.Marshaler
func (n DrupalInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(drupalField{Value: int(n)})
}

// SetBSON implements bson.Setter
func (n *DrupalInt) SetBSON(raw bson.Raw) error {
	// This only marginally less horrible than writing our own BSON library.
	obj := struct {
		Z drupalField `bson:"0"`
	}{}

	err := bson.Unmarshal(raw.Data, &obj)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		*n = 0
		return nil
	}
	*n = DrupalInt(tryInt(obj.Z.Value))
	return nil
}

// GetBSON implements bson.Getter
func (n DrupalInt) GetBSON() (interface{}, error) {
	return []drupalField{{Value: int(n)}}, nil
}

// DrupalStr represents a drupal field of type string.
type DrupalStr string

// UnmarshalJSON implements json.Unmarshaler
func (n *DrupalStr) UnmarshalJSON(data []byte) error {
	obj := []drupalField{}
	err := json.Unmarshal(data, &obj)
	if err != nil || len(obj) == 0 {
		*n = DrupalStr("")
		return nil
	}
	*n = DrupalStr(tryStr(obj[0].Value))
	return nil
}

// MarshalJSON implements json.Marshaler
func (n DrupalStr) MarshalJSON() ([]byte, error) {
	return json.Marshal([]drupalField{{Value: string(n)}})
}

// SetBSON implements bson.Setter
func (n *DrupalStr) SetBSON(raw bson.Raw) error {
	// This only marginally less horrible than writing our own BSON library.
	obj := struct {
		Z drupalField `bson:"0"`
	}{}

	err := bson.Unmarshal(raw.Data, &obj)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		*n = DrupalStr("")
		return nil
	}
	*n = DrupalStr(tryStr(obj.Z.Value))
	return nil
}

// GetBSON implements bson.Getter
func (n DrupalStr) GetBSON() (interface{}, error) {
	return []drupalField{{Value: string(n)}}, nil
}

// IDToBytes converts an object ID to raw bytes.
func IDToBytes(id bson.ObjectId) []byte {
	return []byte(string(id))
}

// IDFromBytes converts raw bytes to an object ID.
func IDFromBytes(id []byte) bson.ObjectId {
	return bson.ObjectId(string(id))
}
