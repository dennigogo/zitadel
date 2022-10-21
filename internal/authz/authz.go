package authz

import (
	"database/sql"

	"github.com/dennigogo/zitadel/internal/authz/repository"
	"github.com/dennigogo/zitadel/internal/authz/repository/eventsourcing"
	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/query"
)

func Start(queries *query.Queries, dbClient *sql.DB, keyEncryptionAlgorithm crypto.EncryptionAlgorithm, externalSecure bool) (repository.Repository, error) {
	return eventsourcing.Start(queries, dbClient, keyEncryptionAlgorithm, externalSecure)
}
