package sql

import (
	"database/sql"
	"os"
	"testing"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/cmd/initialise"
	"github.com/dennigogo/zitadel/internal/database"
	"github.com/dennigogo/zitadel/internal/database/cockroach"
)

var (
	testCRDBClient *sql.DB
)

func TestMain(m *testing.M) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		logging.WithFields("error", err).Fatal("unable to start db")
	}

	testCRDBClient, err = sql.Open("postgres", ts.PGURL().String())
	if err != nil {
		logging.WithFields("error", err).Fatal("unable to connect to db")
	}
	if err = testCRDBClient.Ping(); err != nil {
		logging.WithFields("error", err).Fatal("unable to ping db")
	}

	defer func() {
		testCRDBClient.Close()
		ts.Stop()
	}()

	if err = initDB(testCRDBClient); err != nil {
		logging.WithFields("error", err).Fatal("migrations failed")
	}

	os.Exit(m.Run())
}

func initDB(db *sql.DB) error {
	config := new(database.Config)
	config.SetConnector(&cockroach.Config{User: cockroach.User{Username: "zitadel"}, Database: "zitadel"})

	if err := initialise.ReadStmts("cockroach"); err != nil {
		return err
	}

	err := initialise.Init(db,
		initialise.VerifyUser(config.Username(), ""),
		initialise.VerifyDatabase(config.Database()),
		initialise.VerifyGrant(config.Database(), config.Username()))
	if err != nil {
		return err
	}

	return initialise.VerifyZitadel(db, *config)
}

func fillUniqueData(unique_type, field, instanceID string) error {
	_, err := testCRDBClient.Exec("INSERT INTO eventstore.unique_constraints (unique_type, unique_field, instance_id) VALUES ($1, $2, $3)", unique_type, field, instanceID)
	return err
}
