package types

import (
	"bytes"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BSONRegistry decodes embedded documents and arrays into plain Go maps and slices
// instead of bson.D and bson.A.
var BSONRegistry = newBSONRegistry()

func newBSONRegistry() *bson.Registry {
	r := bson.NewRegistry()
	r.RegisterTypeMapEntry(bson.TypeEmbeddedDocument, reflect.TypeOf(map[string]interface{}{}))
	r.RegisterTypeMapEntry(bson.TypeArray, reflect.TypeOf([]interface{}{}))

	timeType := reflect.TypeOf(time.Time{})
	defaultTimeDecoder, err := r.LookupDecoder(timeType)
	if err != nil {
		panic(err)
	}
	r.RegisterTypeDecoder(timeType, bson.ValueDecoderFunc(func(dc bson.DecodeContext, vr bson.ValueReader, val reflect.Value) error {
		if vr.Type() != bson.TypeString {
			return defaultTimeDecoder.DecodeValue(dc, vr, val)
		}
		s, err := vr.ReadString()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(parseLegacyTime(s)))
		return nil
	}))
	return r
}

// parseLegacyTime parses dates stored as strings (RFC 3339 or JavaScript Date.toString()).
// Like the old mgo driver, it yields the zero time for unparseable values.
func parseLegacyTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC()
	}
	// "Wed Feb 12 2020 23:28:47 GMT+0000 (Coordinated Universal Time)"
	if i := strings.Index(s, " ("); i >= 0 {
		s = s[:i]
	}
	if t, err := time.Parse("Mon Jan 02 2006 15:04:05 GMT-0700", s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

func unmarshalBSON(data []byte, v interface{}) error {
	dec := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(data)))
	dec.SetRegistry(BSONRegistry)
	return dec.Decode(v)
}
