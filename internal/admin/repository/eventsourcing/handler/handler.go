package handler

import (
	"time"

	"github.com/dennigogo/zitadel/internal/admin/repository/eventsourcing/view"
	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/query"
	"github.com/dennigogo/zitadel/internal/static"
)

type Configs map[string]*Config

type Config struct {
	MinimumCycleDuration time.Duration
}

type handler struct {
	view                *view.View
	bulkLimit           uint64
	cycleDuration       time.Duration
	errorCountUntilSkip uint64

	es v1.Eventstore
}

func (h *handler) Eventstore() v1.Eventstore {
	return h.es
}

func Register(configs Configs, bulkLimit, errorCount uint64, view *view.View, es v1.Eventstore, static static.Storage) []query.Handler {
	handlers := []query.Handler{}
	if static != nil {
		handlers = append(handlers, newStyling(
			handler{view, bulkLimit, configs.cycleDuration("Styling"), errorCount, es},
			static))
	}
	return handlers
}

func (configs Configs) cycleDuration(viewModel string) time.Duration {
	c, ok := configs[viewModel]
	if !ok {
		return 3 * time.Minute
	}
	return c.MinimumCycleDuration
}

func (h *handler) MinimumCycleDuration() time.Duration {
	return h.cycleDuration
}

func (h *handler) LockDuration() time.Duration {
	return h.cycleDuration / 3
}

func (h *handler) QueryLimit() uint64 {
	return h.bulkLimit
}
