package cockroach

import (
	"github.com/dennigogo/zitadel/internal/database/dialect"
)

func init() {
	config := &Config{}
	dialect.Register(config, config, true)
}
