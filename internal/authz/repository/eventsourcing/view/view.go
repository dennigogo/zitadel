package view

import (
	"database/sql"

	"github.com/dennigogo/zitadel/internal/id"
	"github.com/dennigogo/zitadel/internal/query"

	"github.com/jinzhu/gorm"
)

type View struct {
	Db          *gorm.DB
	Query       *query.Queries
	idGenerator id.Generator
}

func StartView(sqlClient *sql.DB, idGenerator id.Generator, queries *query.Queries) (*View, error) {
	gorm, err := gorm.Open("postgres", sqlClient)
	if err != nil {
		return nil, err
	}
	return &View{
		Db:          gorm,
		idGenerator: idGenerator,
		Query:       queries,
	}, nil
}

func (v *View) Health() (err error) {
	return v.Db.DB().Ping()
}
