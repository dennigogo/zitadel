package query

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"

	"github.com/rakyll/statik/fs"
	"golang.org/x/text/language"

	"github.com/dennigogo/zitadel/internal/api/authz"
	sd "github.com/dennigogo/zitadel/internal/config/systemdefaults"
	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/query/projection"
	"github.com/dennigogo/zitadel/internal/repository/action"
	iam_repo "github.com/dennigogo/zitadel/internal/repository/instance"
	"github.com/dennigogo/zitadel/internal/repository/keypair"
	"github.com/dennigogo/zitadel/internal/repository/org"
	"github.com/dennigogo/zitadel/internal/repository/project"
	usr_repo "github.com/dennigogo/zitadel/internal/repository/user"
	"github.com/dennigogo/zitadel/internal/repository/usergrant"
)

type Queries struct {
	eventstore *eventstore.Eventstore
	client     *sql.DB

	idpConfigEncryption crypto.EncryptionAlgorithm

	DefaultLanguage                     language.Tag
	LoginDir                            http.FileSystem
	NotificationDir                     http.FileSystem
	mutex                               sync.Mutex
	LoginTranslationFileContents        map[string][]byte
	NotificationTranslationFileContents map[string][]byte
	supportedLangs                      []language.Tag
	zitadelRoles                        []authz.RoleMapping
	multifactors                        domain.MultifactorConfigs
}

func StartQueries(ctx context.Context, es *eventstore.Eventstore, sqlClient *sql.DB, projections projection.Config, defaults sd.SystemDefaults, idpConfigEncryption, otpEncryption, keyEncryptionAlgorithm crypto.EncryptionAlgorithm, certEncryptionAlgorithm crypto.EncryptionAlgorithm, zitadelRoles []authz.RoleMapping) (repo *Queries, err error) {
	statikLoginFS, err := fs.NewWithNamespace("login")
	if err != nil {
		return nil, fmt.Errorf("unable to start login statik dir")
	}

	statikNotificationFS, err := fs.NewWithNamespace("notification")
	if err != nil {
		return nil, fmt.Errorf("unable to start notification statik dir")
	}

	repo = &Queries{
		eventstore:                          es,
		client:                              sqlClient,
		DefaultLanguage:                     language.Und,
		LoginDir:                            statikLoginFS,
		NotificationDir:                     statikNotificationFS,
		LoginTranslationFileContents:        make(map[string][]byte),
		NotificationTranslationFileContents: make(map[string][]byte),
		zitadelRoles:                        zitadelRoles,
	}
	iam_repo.RegisterEventMappers(repo.eventstore)
	usr_repo.RegisterEventMappers(repo.eventstore)
	org.RegisterEventMappers(repo.eventstore)
	project.RegisterEventMappers(repo.eventstore)
	action.RegisterEventMappers(repo.eventstore)
	keypair.RegisterEventMappers(repo.eventstore)
	usergrant.RegisterEventMappers(repo.eventstore)

	repo.idpConfigEncryption = idpConfigEncryption
	repo.multifactors = domain.MultifactorConfigs{
		OTP: domain.OTPConfig{
			CryptoMFA: otpEncryption,
			Issuer:    defaults.Multifactors.OTP.Issuer,
		},
	}

	err = projection.Start(ctx, sqlClient, es, projections, keyEncryptionAlgorithm, certEncryptionAlgorithm)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (q *Queries) Health(ctx context.Context) error {
	return q.client.Ping()
}
