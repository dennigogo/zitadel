package spooler

import (
	"database/sql"
	"time"

	es_locker "github.com/dennigogo/zitadel/internal/eventstore/v1/locker"
)

const (
	lockTable = "auth.locks"
)

type locker struct {
	dbClient *sql.DB
}

func NewLocker(client *sql.DB) *locker {
	return &locker{dbClient: client}
}

func (l *locker) Renew(lockerID, viewModel, instanceID string, waitTime time.Duration) error {
	return es_locker.Renew(l.dbClient, lockTable, lockerID, viewModel, instanceID, waitTime)
}
