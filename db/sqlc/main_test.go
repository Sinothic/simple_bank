package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/Sinothic/simplebank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDBConnection *sql.DB

func TestMain(m *testing.M) {
	var err error
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load conf ig:", err)
	}

	testDBConnection, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDBConnection)

	// Run the tests
	os.Exit(m.Run())
}
