package auth

import (
	"context"

	"github.com/dennigogo/zitadel/internal/api/grpc/text"
	auth_pb "github.com/dennigogo/zitadel/pkg/grpc/auth"
)

func (s *Server) GetSupportedLanguages(ctx context.Context, req *auth_pb.GetSupportedLanguagesRequest) (*auth_pb.GetSupportedLanguagesResponse, error) {
	langs, err := s.query.Languages(ctx)
	if err != nil {
		return nil, err
	}
	return &auth_pb.GetSupportedLanguagesResponse{Languages: text.LanguageTagsToStrings(langs)}, nil
}
