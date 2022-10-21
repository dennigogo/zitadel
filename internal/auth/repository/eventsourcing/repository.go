package eventsourcing

import (
	"context"
	"database/sql"

	"github.com/dennigogo/zitadel/internal/auth/repository/eventsourcing/eventstore"
	"github.com/dennigogo/zitadel/internal/auth/repository/eventsourcing/spooler"
	auth_view "github.com/dennigogo/zitadel/internal/auth/repository/eventsourcing/view"
	"github.com/dennigogo/zitadel/internal/auth_request/repository/cache"
	"github.com/dennigogo/zitadel/internal/command"
	sd "github.com/dennigogo/zitadel/internal/config/systemdefaults"
	"github.com/dennigogo/zitadel/internal/crypto"
	v1 "github.com/dennigogo/zitadel/internal/eventstore/v1"
	es_spol "github.com/dennigogo/zitadel/internal/eventstore/v1/spooler"
	"github.com/dennigogo/zitadel/internal/id"
	"github.com/dennigogo/zitadel/internal/query"
)

type Config struct {
	SearchLimit uint64
	Spooler     spooler.SpoolerConfig
}

type EsRepository struct {
	spooler    *es_spol.Spooler
	Eventstore v1.Eventstore
	eventstore.UserRepo
	eventstore.AuthRequestRepo
	eventstore.TokenRepo
	eventstore.RefreshTokenRepo
	eventstore.UserSessionRepo
	eventstore.OrgRepository
}

func Start(conf Config, systemDefaults sd.SystemDefaults, command *command.Commands, queries *query.Queries, dbClient *sql.DB, oidcEncryption crypto.EncryptionAlgorithm, userEncryption crypto.EncryptionAlgorithm) (*EsRepository, error) {
	es, err := v1.Start(dbClient)
	if err != nil {
		return nil, err
	}
	idGenerator := id.SonyFlakeGenerator()

	view, err := auth_view.StartView(dbClient, oidcEncryption, queries, idGenerator, es)
	if err != nil {
		return nil, err
	}

	authReq := cache.Start(dbClient)

	spool := spooler.StartSpooler(conf.Spooler, es, view, dbClient, systemDefaults, queries)

	userRepo := eventstore.UserRepo{
		SearchLimit:    conf.SearchLimit,
		Eventstore:     es,
		View:           view,
		Query:          queries,
		SystemDefaults: systemDefaults,
	}
	//TODO: remove as soon as possible
	queryView := queryViewWrapper{
		queries,
		view,
	}
	return &EsRepository{
		spool,
		es,
		userRepo,
		eventstore.AuthRequestRepo{
			PrivacyPolicyProvider:     queries,
			LabelPolicyProvider:       queries,
			Command:                   command,
			Query:                     queries,
			OrgViewProvider:           queries,
			AuthRequests:              authReq,
			View:                      view,
			Eventstore:                es,
			UserCodeAlg:               userEncryption,
			UserSessionViewProvider:   view,
			UserViewProvider:          view,
			UserCommandProvider:       command,
			UserEventProvider:         &userRepo,
			IDPProviderViewProvider:   view,
			IDPUserLinksProvider:      queries,
			LockoutPolicyViewProvider: queries,
			LoginPolicyViewProvider:   queries,
			UserGrantProvider:         queryView,
			ProjectProvider:           queryView,
			ApplicationProvider:       queries,
			IdGenerator:               idGenerator,
		},
		eventstore.TokenRepo{
			View:       view,
			Eventstore: es,
		},
		eventstore.RefreshTokenRepo{
			View:         view,
			Eventstore:   es,
			SearchLimit:  conf.SearchLimit,
			KeyAlgorithm: oidcEncryption,
		},
		eventstore.UserSessionRepo{
			View: view,
		},
		eventstore.OrgRepository{
			SearchLimit:    conf.SearchLimit,
			View:           view,
			SystemDefaults: systemDefaults,
			Eventstore:     es,
			Query:          queries,
		},
	}, nil
}

type queryViewWrapper struct {
	*query.Queries
	*auth_view.View
}

func (q queryViewWrapper) UserGrantsByProjectAndUserID(ctx context.Context, projectID, userID string) ([]*query.UserGrant, error) {
	userGrantProjectID, err := query.NewUserGrantProjectIDSearchQuery(projectID)
	if err != nil {
		return nil, err
	}
	userGrantUserID, err := query.NewUserGrantUserIDSearchQuery(userID)
	if err != nil {
		return nil, err
	}
	queries := &query.UserGrantsQueries{Queries: []query.SearchQuery{userGrantUserID, userGrantProjectID}}
	grants, err := q.Queries.UserGrants(ctx, queries)
	if err != nil {
		return nil, err
	}
	return grants.UserGrants, nil
}
func (repo *EsRepository) Health(ctx context.Context) error {
	if err := repo.UserRepo.Health(ctx); err != nil {
		return err
	}
	return repo.AuthRequestRepo.Health(ctx)
}
