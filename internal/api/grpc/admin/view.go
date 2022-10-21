package admin

import (
	"context"

	"github.com/dennigogo/zitadel/internal/query"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func (s *Server) ListViews(ctx context.Context, _ *admin_pb.ListViewsRequest) (*admin_pb.ListViewsResponse, error) {
	currentSequences, err := s.query.SearchCurrentSequences(ctx, new(query.CurrentSequencesSearchQueries))
	if err != nil {
		return nil, err
	}
	convertedCurrentSequences := CurrentSequencesToPb(s.database, currentSequences)
	views, err := s.administrator.GetViews()
	if err != nil {
		return nil, err
	}
	convertedViews := ViewsToPb(views)

	convertedCurrentSequences = append(convertedCurrentSequences, convertedViews...)
	return &admin_pb.ListViewsResponse{Result: convertedCurrentSequences}, nil
}
