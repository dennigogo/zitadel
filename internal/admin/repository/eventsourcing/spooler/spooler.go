package spooler

import (
	"database/sql"

	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	"github.com/dennigogo/zitadel/internal/static"

	"github.com/dennigogo/zitadel/internal/admin/repository/eventsourcing/handler"
	"github.com/dennigogo/zitadel/internal/admin/repository/eventsourcing/view"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/spooler"
)

type SpoolerConfig struct {
	BulkLimit             uint64
	FailureCountUntilSkip uint64
	ConcurrentWorkers     int
	ConcurrentInstances   int
	Handlers              handler.Configs
}

func StartSpooler(c SpoolerConfig, es v1.Eventstore, view *view.View, sql *sql.DB, static static.Storage) *spooler.Spooler {
	spoolerConfig := spooler.Config{
		Eventstore:          es,
		Locker:              &locker{dbClient: sql},
		ConcurrentWorkers:   c.ConcurrentWorkers,
		ConcurrentInstances: c.ConcurrentInstances,
		ViewHandlers:        handler.Register(c.Handlers, c.BulkLimit, c.FailureCountUntilSkip, view, es, static),
	}
	spool := spoolerConfig.New()
	spool.Start()
	return spool
}
