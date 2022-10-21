package eventsourcing

import (
	"context"
	"database/sql"

	"github.com/dennigogo/zitadel/internal/authz/repository"
	"github.com/dennigogo/zitadel/internal/authz/repository/eventsourcing/eventstore"
	authz_view "github.com/dennigogo/zitadel/internal/authz/repository/eventsourcing/view"
	"github.com/dennigogo/zitadel/internal/crypto"
	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	"github.com/dennigogo/zitadel/internal/id"
	"github.com/dennigogo/zitadel/internal/query"
)

type EsRepository struct {
	eventstore.UserMembershipRepo
	eventstore.TokenVerifierRepo
}

func Start(queries *query.Queries, dbClient *sql.DB, keyEncryptionAlgorithm crypto.EncryptionAlgorithm, externalSecure bool) (repository.Repository, error) {
	es, err := v1.Start(dbClient)
	if err != nil {
		return nil, err
	}

	idGenerator := id.SonyFlakeGenerator()
	view, err := authz_view.StartView(dbClient, idGenerator, queries)
	if err != nil {
		return nil, err
	}

	return &EsRepository{
		eventstore.UserMembershipRepo{
			Queries: queries,
		},
		eventstore.TokenVerifierRepo{
			TokenVerificationKey: keyEncryptionAlgorithm,
			Eventstore:           es,
			View:                 view,
			Query:                queries,
			ExternalSecure:       externalSecure,
		},
	}, nil
}

func (repo *EsRepository) Health(ctx context.Context) error {
	if err := repo.TokenVerifierRepo.Health(); err != nil {
		return err
	}
	return nil
}
