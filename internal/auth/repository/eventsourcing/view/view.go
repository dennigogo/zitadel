package view

import (
	"database/sql"

	"github.com/jinzhu/gorm"

	"github.com/dennigogo/zitadel/internal/crypto"
	eventstore "github.com/dennigogo/zitadel/internal/eventstore/v1"
	"github.com/dennigogo/zitadel/internal/id"
	"github.com/dennigogo/zitadel/internal/query"
)

type View struct {
	Db           *gorm.DB
	keyAlgorithm crypto.EncryptionAlgorithm
	idGenerator  id.Generator
	query        *query.Queries
	es           eventstore.Eventstore
}

func StartView(sqlClient *sql.DB, keyAlgorithm crypto.EncryptionAlgorithm, queries *query.Queries, idGenerator id.Generator, es eventstore.Eventstore) (*View, error) {
	gorm, err := gorm.Open("postgres", sqlClient)
	if err != nil {
		return nil, err
	}
	return &View{
		Db:           gorm,
		keyAlgorithm: keyAlgorithm,
		idGenerator:  idGenerator,
		query:        queries,
		es:           es,
	}, nil
}

func (v *View) Health() (err error) {
	return v.Db.DB().Ping()
}
