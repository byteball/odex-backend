package services

import (
	"os"
	"testing"

	"github.com/byteball/odex-backend/daos"
	"github.com/byteball/odex-backend/utils/testutils"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var db *mongo.Client

func TestMain(m *testing.M) {
	code := m.Run()
	testutils.DropTestDB()
	os.Exit(code)
}

func init() {
	session := testutils.NewDBTestClient()
	if _, err := daos.InitSession(session); err != nil {
		panic(err)
	}
	db = session
}
