package admin

import (
	"context"

	"github.com/dennigogo/zitadel/internal/query"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func (s *Server) ListFailedEvents(ctx context.Context, req *admin_pb.ListFailedEventsRequest) (*admin_pb.ListFailedEventsResponse, error) {
	failedEventsOld, err := s.administrator.GetFailedEvents(ctx)
	if err != nil {
		return nil, err
	}
	convertedOld := FailedEventsViewToPb(failedEventsOld)

	failedEvents, err := s.query.SearchFailedEvents(ctx, new(query.FailedEventSearchQueries))
	if err != nil {
		return nil, err
	}
	convertedNew := FailedEventsToPb(s.database, failedEvents)
	return &admin_pb.ListFailedEventsResponse{Result: append(convertedOld, convertedNew...)}, nil
}

func (s *Server) RemoveFailedEvent(ctx context.Context, req *admin_pb.RemoveFailedEventRequest) (*admin_pb.RemoveFailedEventResponse, error) {
	var err error
	if req.Database != s.database {
		err = s.administrator.RemoveFailedEvent(ctx, RemoveFailedEventRequestToModel(req))
	} else {
		err = s.query.RemoveFailedEvent(ctx, req.ViewName, req.FailedSequence)
	}
	if err != nil {
		return nil, err
	}
	return &admin_pb.RemoveFailedEventResponse{}, nil
}
