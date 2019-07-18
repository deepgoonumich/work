package models

import (
	"encoding/json"
	"reflect"

	"github.com/tinylib/msgp/msgp"
	"gopkg.in/mgo.v2/bson"
)

// Meta contains the metadata of a content object. This structure is special;
// All of the fields must be pointers to other structures. Unknown Meta fields
// will automatically be deserialized into and serialized from the map, Ext.
type Meta struct {
	SectorV2        *SectorMeta
	Partner         *PartnerMeta
	PartnerTaxonomy *PartnerTaxonomyMeta
	SEC             *SECMeta

	// Ext contains any unknown fields. Can also be used to store arbitrary
	// metadata. These should never be anything other than
	// map[string]interface{}.
	Ext map[string]interface{}
}

// You might be wondering, "What the hell is going on here?" Time to go grab
// a coffee, you're in for a long day. Basically, our `Content` structure
// goes across three different libraries: tinylib/msgp, labix/gobson, and Go's
// own `encoding/json` library. These libraries all have some level of
// interoperability, because they are all modelled to work like encoding/json.
// However, they all have different opinions in many small edge cases. This is
// horrible, but there isn't really a single library that can do all of these
// serialization formats in one + with optimal overhead.
//
// What we want for Meta is:
// - Type-safe access to structures we recognize
// - Interface access to structures we don't
//
// To do this, we want to have various structure fields with exact types, and
// those would be serialized and deserialized just like any other field; but
// then we want a special field at the end that grabs any fields we didn't
// explicitly.
//
// With labix's gobson, this is very easy: you can just use the `inline` flag.
// However, this functionality is _not_ present in any of the other libraries.
// Also, it'd be best to avoid magical struct tags here where we want to be
// very careful about our behavior. We already have some stuff relying on
// the capitalization of fields, so we need to make sure the behavior remains
// exactly as-is.
//
// One thing all of these libraries have in common is that they all support
// custom serialization methods for types you own, via various interfaces.
// Unfortunately, and evidential based on the several hundred lines of this
// file, each library does it differently. For efficiency reasons, msgp goes
// as far as to have 5 interfaces to implement in order to make things work.
//
// Fortunately, encoding/json and labix/gobson are much simpler, and in fact
// the serializers for those libraries are very, very similar. Unfortunately,
// though, they are just different enough to necessitate code duplication.
//
// So to be clear, this `Meta` type implements 9 different interfaces to
// provide flawless interoperability between 3 different serialization
// libraries.
//
// Let's hope nobody ever reads this.
//
// P.S.: If you're trying to simplify this code, I'd advise you to quit now.

// This code sets up a map of fields in the Meta structure.
var metaFields = map[string]int{}

func init() {
	t := reflect.ValueOf(Meta{}).Type()
	l := t.NumField()
	for i := 0; i < l; i++ {
		f := t.Field(i)

		// Do not index Ext field.
		if f.Name == "Ext" {
			continue
		}

		metaFields[f.Name] = i
	}
}

// mergeMeta returns all of the meta in a single map[string]interface{}.
func (m Meta) mergeMeta() (all map[string]interface{}) {
	all = map[string]interface{}{}

	// Get ext keys.
	for key, value := range m.Ext {
		if _, ok := metaFields[key]; ok {
			panic("programming error; meta key conflict")
		}
		all[key] = value
	}

	// Get struct keys.
	for key, index := range metaFields {
		field := reflect.ValueOf(m).Field(index)
		if !field.IsNil() {
			all[key] = field.Interface()
		}
	}

	return
}

// clear deletes everything in the Meta structure.
func (m *Meta) clear() {
	m.Ext = map[string]interface{}{}

	for _, index := range metaFields {
		field := reflect.ValueOf(m).Elem().Field(index)
		field.Set(reflect.Zero(field.Type()))
	}
}

// emptyToNil zeros out Ext if it has no data.
func (m *Meta) emptyToNil() {
	if len(m.Ext) == 0 {
		m.Ext = nil
	}
}

// DecodeMsg implements msgp.Decodable
func (m *Meta) DecodeMsg(r *msgp.Reader) error {
	maplen, err := r.ReadMapHeader()
	if err != nil {
		return err
	}

	m.clear()

	for maplen > 0 {
		var key string

		maplen--

		key, err = r.ReadString()
		if err != nil {
			return err
		}

		if si, ok := metaFields[key]; ok {
			field := reflect.ValueOf(m).Elem().Field(si)

			// Create new instance of meta field type.
			field.Set(reflect.New(field.Type().Elem()))

			// Get field as msgp.Decodable.
			decodable, ok := field.Interface().(msgp.Decodable)
			if !ok {
				panic("programmer error: field should be msgp.Decodable")
			}

			// Decode field.
			err = decodable.DecodeMsg(r)
			if err != nil {
				return err
			}
		} else {
			intf := interface{}(nil)

			intf, err = r.ReadIntf()
			if err != nil {
				return err
			}

			m.Ext[key] = intf
		}
	}

	m.emptyToNil()

	return nil
}

// EncodeMsg implements msgp.Encodable
func (m *Meta) EncodeMsg(w *msgp.Writer) error {
	all := m.mergeMeta()

	err := w.WriteMapHeader(uint32(len(all)))
	if err != nil {
		return err
	}

	for key, value := range all {
		err = w.WriteString(key)
		if err != nil {
			return err
		}

		err = w.WriteIntf(value)
		if err != nil {
			return err
		}
	}

	return nil
}

// MarshalMsg implements msgp.Marshaler
func (m *Meta) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, m.Msgsize())

	all := m.mergeMeta()

	o = msgp.AppendMapHeader(o, uint32(len(all)))
	for key, value := range all {
		o = msgp.AppendString(o, key)

		o, err = msgp.AppendIntf(o, value)
		if err != nil {
			return
		}
	}

	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (m *Meta) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var maplen uint32
	maplen, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}

	m.clear()

	for maplen > 0 {
		var key string

		maplen--

		key, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}

		if si, ok := metaFields[key]; ok {
			field := reflect.ValueOf(m).Elem().Field(si)

			// Create new instance of meta field type.
			field.Set(reflect.New(field.Type().Elem()))

			// Get field as unmarshaler.
			unmarshaler, ok := field.Interface().(msgp.Unmarshaler)
			if !ok {
				panic("programmer error: field should be msgp.Unmarshaler")
			}

			// Unmarshal field.
			bts, err = unmarshaler.UnmarshalMsg(bts)
			if err != nil {
				return nil, err
			}
		} else {
			intf := interface{}(nil)

			intf, bts, err = msgp.ReadIntfBytes(bts)
			if err != nil {
				return nil, err
			}

			m.Ext[key] = intf
		}
	}

	m.emptyToNil()

	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (m Meta) Msgsize() (s int) {
	s += msgp.MapHeaderSize
	for key, value := range m.mergeMeta() {
		s += msgp.StringPrefixSize + len(key) + msgp.GuessSize(value)
	}
	return
}

// GetBSON implements bson.Getter
func (m Meta) GetBSON() (interface{}, error) {
	return m.mergeMeta(), nil
}

// SetBSON implements bson.Setter
func (m *Meta) SetBSON(raw bson.Raw) error {
	m.clear()

	all := map[string]bson.Raw{}

	err := raw.Unmarshal(all)
	if err != nil {
		return err
	}

	for key, value := range all {
		if si, ok := metaFields[key]; ok {
			field := reflect.ValueOf(m).Elem().Field(si)

			// Create new instance of meta field type.
			field.Set(reflect.New(field.Type().Elem()))

			// Decode into new instance.
			err = value.Unmarshal(field.Interface())
			if err != nil {
				return err
			}
		} else {
			intf := interface{}(nil)

			err = value.Unmarshal(&intf)
			if err != nil {
				return err
			}

			m.Ext[key] = intf
		}
	}

	m.emptyToNil()

	return nil
}

// MarshalJSON implements json.Marshaler
func (m Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.mergeMeta())
}

// UnmarshalJSON implements json.Unmarshaler
func (m *Meta) UnmarshalJSON(data []byte) error {
	m.clear()

	all := map[string]json.RawMessage{}

	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}

	for key, value := range all {
		if si, ok := metaFields[key]; ok {
			field := reflect.ValueOf(m).Elem().Field(si)

			// Create new instance of meta field type.
			field.Set(reflect.New(field.Type().Elem()))

			// Decode into new instance.
			err = json.Unmarshal(value, field.Interface())
			if err != nil {
				return err
			}
		} else {
			intf := interface{}(nil)

			err = json.Unmarshal(value, &intf)
			if err != nil {
				return err
			}

			m.Ext[key] = intf
		}
	}

	m.emptyToNil()

	return nil
}
