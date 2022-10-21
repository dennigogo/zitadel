package postgres

import (

	//sql import
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/dennigogo/zitadel/internal/database/dialect"
)

func init() {
	config := &Config{}
	dialect.Register(config, config, false)
}
