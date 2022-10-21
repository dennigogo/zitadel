package eventstore

import (
	"database/sql"

	z_sql "github.com/dennigogo/zitadel/internal/eventstore/repository/sql"
)

func Start(sqlClient *sql.DB) (*Eventstore, error) {
	return NewEventstore(z_sql.NewCRDB(sqlClient)), nil
}
