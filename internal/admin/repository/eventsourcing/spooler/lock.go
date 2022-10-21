package spooler

import (
	"database/sql"
	"time"

	es_locker "github.com/dennigogo/zitadel/internal/eventstore/v1/locker"
)

const (
	lockTable = "adminapi.locks"
)

type locker struct {
	dbClient *sql.DB
}

func (l *locker) Renew(lockerID, viewModel, instanceID string, waitTime time.Duration) error {
	return es_locker.Renew(l.dbClient, lockTable, lockerID, viewModel, instanceID, waitTime)
}
