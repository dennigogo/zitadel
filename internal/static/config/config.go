package config

import (
	"database/sql"

	"github.com/dennigogo/zitadel/internal/api/http/middleware"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/static"
	"github.com/dennigogo/zitadel/internal/static/database"
	"github.com/dennigogo/zitadel/internal/static/s3"
)

type AssetStorageConfig struct {
	Type   string
	Cache  middleware.CacheConfig
	Config map[string]interface{} `mapstructure:",remain"`
}

func (a *AssetStorageConfig) NewStorage(client *sql.DB) (static.Storage, error) {
	t, ok := storage[a.Type]
	if !ok {
		return nil, errors.ThrowInternalf(nil, "STATIC-dsbjh", "config type %s not supported", a.Type)
	}

	return t(client, a.Config)
}

var storage = map[string]static.CreateStorage{
	"db": database.NewStorage,
	"":   database.NewStorage,
	"s3": s3.NewStorage,
}
