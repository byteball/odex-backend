package types

import (
	"bytes"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BSONRegistry decodes embedded documents and arrays into plain Go maps and slices
// instead of bson.D and bson.A.
var BSONRegistry = newBSONRegistry()

func newBSONRegistry() *bson.Registry {
	r := bson.NewRegistry()
	r.RegisterTypeMapEntry(bson.TypeEmbeddedDocument, reflect.TypeOf(map[string]interface{}{}))
	r.RegisterTypeMapEntry(bson.TypeArray, reflect.TypeOf([]interface{}{}))
	return r
}

func unmarshalBSON(data []byte, v interface{}) error {
	dec := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(data)))
	dec.SetRegistry(BSONRegistry)
	return dec.Decode(v)
}
