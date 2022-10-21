package handler

import (
	"github.com/dennigogo/zitadel/internal/eventstore"
)

type HandlerConfig struct {
	Eventstore *eventstore.Eventstore
}
type Handler struct {
	Eventstore *eventstore.Eventstore
	Sub        *eventstore.Subscription
	EventQueue chan eventstore.Event
}

func NewHandler(config HandlerConfig) Handler {
	return Handler{
		Eventstore: config.Eventstore,
		EventQueue: make(chan eventstore.Event, 100),
	}
}

func (h *Handler) Subscribe(aggregates ...eventstore.AggregateType) {
	h.Sub = eventstore.SubscribeAggregates(h.EventQueue, aggregates...)
}

func (h *Handler) SubscribeEvents(types map[eventstore.AggregateType][]eventstore.EventType) {
	h.Sub = eventstore.SubscribeEventTypes(h.EventQueue, types)
}

func (h *Handler) Unsubscribe() {
	if h.Sub == nil {
		return
	}
	h.Sub.Unsubscribe()
}
