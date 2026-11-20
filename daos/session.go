package daos

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"github.com/byteball/odex-backend/app"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Database struct contains the pointer to the mongo client
// It is a wrapper over the mongo driver to help utilize its connection pool
type Database struct {
	Session *mongo.Client
}

// Global instance of Database struct for singleton use
var db *Database
var logger = utils.Logger
var defaultTimeout = 10 * time.Second
var indexTimeout = 30 * time.Minute

func InitSession(session *mongo.Client) (*mongo.Client, error) {
	if db == nil {
		if session == nil {
			if app.Config.EnableTLS {
				session = NewTLSSession()
			} else {
				session = NewSession()
			}
		}

		db = &Database{session}
	}

	return db.Session, nil
}

func InitTLSSession() (*mongo.Client, error) {
	session := NewTLSSession()
	db = &Database{session}
	return db.Session, nil
}

func (d *Database) InitDatabase(session *mongo.Client) {
	d.Session = session
}

// connect creates a client and checks that the server is reachable
func connect(opts *options.ClientOptions, timeout time.Duration) *mongo.Client {
	opts.SetRegistry(types.BSONRegistry)

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	return client
}

func NewSession() *mongo.Client {
	timeout := 15 * time.Second
	opts := options.Client().
		SetHosts([]string{app.Config.MongoURL}).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout)

	return connect(opts, timeout)
}

func NewTLSSession() *mongo.Client {
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true

	timeout := 20 * time.Second
	opts := options.Client().
		SetHosts([]string{
			app.Config.MongoDBShardURL1,
			app.Config.MongoDBShardURL2,
			app.Config.MongoDBShardURL3,
		}).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout).
		SetMaxConnIdleTime(30 * time.Second).
		SetAuth(options.Credential{
			AuthSource: "admin",
			Username:   app.Config.MongoDBUsername,
			Password:   app.Config.MongoDBPassword,
		}).
		SetTLSConfig(tlsConfig)

	return connect(opts, timeout)
}

func NewSecureTLSSession() *mongo.Client {
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true

	timeout := 15 * time.Second
	opts := options.Client().
		SetHosts([]string{app.Config.MongoURL}).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout).
		SetAuth(options.Credential{
			AuthSource: "odex",
			Username:   app.Config.MongoDBUsername,
			Password:   app.Config.MongoDBPassword,
		}).
		SetTLSConfig(tlsConfig)

	return connect(opts, timeout)
}

// keySpec builds an ordered key document from field names, a "-" prefix meaning descending
func keySpec(fields ...string) bson.D {
	keys := make(bson.D, 0, len(fields))
	for _, f := range fields {
		if strings.HasPrefix(f, "-") {
			keys = append(keys, bson.E{Key: f[1:], Value: -1})
		} else {
			keys = append(keys, bson.E{Key: f, Value: 1})
		}
	}

	return keys
}

func (d *Database) collection(dbName, collection string) *mongo.Collection {
	return d.Session.Database(dbName).Collection(collection)
}

func readContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// EnsureIndex creates the index if it does not exist yet
func (d *Database) EnsureIndex(dbName, collection string, index mongo.IndexModel) error {
	ctx, cancel := context.WithTimeout(context.Background(), indexTimeout)
	defer cancel()

	_, err := d.collection(dbName, collection).Indexes().CreateOne(ctx, index)
	return err
}

// Create is a wrapper for the driver's insert functions.
func (d *Database) Create(dbName, collection string, data ...interface{}) (err error) {
	if len(data) == 0 {
		return nil
	}

	if len(data) == 1 {
		_, err = d.collection(dbName, collection).InsertOne(context.Background(), data[0])
		return
	}

	_, err = d.collection(dbName, collection).InsertMany(context.Background(), data)
	return
}

// GetByID is a wrapper for the driver's FindOne function.
func (d *Database) GetByID(dbName, collection string, id bson.ObjectID, response interface{}) (err error) {
	ctx, cancel := readContext()
	defer cancel()

	return d.collection(dbName, collection).FindOne(ctx, bson.M{"_id": id}).Decode(response)
}

// Get is a wrapper for the driver's Find function.
func (d *Database) Get(dbName, collection string, query interface{}, offset, limit int, response interface{}) (err error) {
	return d.find(dbName, collection, query, options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)), response)
}

func (d *Database) Query(dbName, collection string, query interface{}, selector interface{}, offset, limit int, response interface{}) (err error) {
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetProjection(selector)
	return d.find(dbName, collection, query, opts, response)
}

// GetAndSort is a wrapper for the driver's Find function with a sort option.
// Fields prefixed by "-" are sorted in descending order.
func (d *Database) GetAndSort(dbName, collection string, query interface{}, sort []string, offset, limit int, response interface{}) (err error) {
	opts := options.Find().SetSort(keySpec(sort...)).SetSkip(int64(offset)).SetLimit(int64(limit))
	return d.find(dbName, collection, query, opts, response)
}

func (d *Database) find(dbName, collection string, query interface{}, opts *options.FindOptionsBuilder, response interface{}) error {
	ctx, cancel := readContext()
	defer cancel()

	cursor, err := d.collection(dbName, collection).Find(ctx, query, opts)
	if err != nil {
		return err
	}

	return cursor.All(ctx, response)
}

func (d *Database) Count(dbName, collection string, query interface{}) (n int, err error) {
	count, err := d.collection(dbName, collection).CountDocuments(context.Background(), query)
	return int(count), err
}

// Update updates the first document matching the query with an update document
// (using operators such as $set). It returns mongo.ErrNoDocuments if nothing matched.
func (d *Database) Update(dbName, collection string, query interface{}, update interface{}) error {
	res, err := d.collection(dbName, collection).UpdateOne(context.Background(), query, update)
	if err == nil && res.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
	}

	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// Replace replaces the first document matching the query with a whole new document.
// It returns mongo.ErrNoDocuments if nothing matched and upsert is false.
func (d *Database) Replace(dbName, collection string, query interface{}, replacement interface{}, upsert bool) error {
	opts := options.Replace().SetUpsert(upsert)
	res, err := d.collection(dbName, collection).ReplaceOne(context.Background(), query, replacement, opts)
	if err == nil && res.MatchedCount == 0 && res.UpsertedCount == 0 {
		err = mongo.ErrNoDocuments
	}

	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (d *Database) UpdateAll(dbName, collection string, query interface{}, update interface{}) error {
	_, err := d.collection(dbName, collection).UpdateMany(context.Background(), query, update)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// Upsert applies an update document (using operators such as $set) to the first document
// matching the query, inserting a new document if there is none.
func (d *Database) Upsert(dbName, collection string, query interface{}, update interface{}) error {
	_, err := d.collection(dbName, collection).UpdateOne(context.Background(), query, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// FindAndModify upserts a document and decodes the resulting document into response.
func (d *Database) FindAndModify(dbName, collection string, query interface{}, update interface{}, response interface{}) error {
	ctx, cancel := readContext()
	defer cancel()

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := d.collection(dbName, collection).FindOneAndUpdate(ctx, query, update, opts).Decode(response)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// Aggregate is a wrapper for the driver's Aggregate function.
// It is used to make mongo aggregate pipeline queries
func (d *Database) Aggregate(dbName, collection string, query []bson.M, response interface{}) error {
	ctx := context.Background()

	cursor, err := d.collection(dbName, collection).Aggregate(ctx, query)
	if err == nil {
		err = cursor.All(ctx, response)
	}

	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// Remove removes one document matching a certain query
func (d *Database) Remove(dbName, collection string, query interface{}) error {
	res, err := d.collection(dbName, collection).DeleteOne(context.Background(), query)
	if err == nil && res.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}

	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// RemoveAll removes all the documents from a collection matching a certain query
func (d *Database) RemoveAll(dbName, collection string, query interface{}) error {
	_, err := d.collection(dbName, collection).DeleteMany(context.Background(), query)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// DropCollection drops all the documents in a collection
func (d *Database) DropCollection(dbName, collection string) error {
	err := d.collection(dbName, collection).Drop(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
