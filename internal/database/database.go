package database

import (
	"database/sql"
	"reflect"

	_ "github.com/dennigogo/zitadel/internal/database/cockroach"
	"github.com/dennigogo/zitadel/internal/database/dialect"
	_ "github.com/dennigogo/zitadel/internal/database/postgres"
	"github.com/dennigogo/zitadel/internal/errors"
)

type Config struct {
	Dialects  map[string]interface{} `mapstructure:",remain"`
	connector dialect.Connector
}

func (c *Config) SetConnector(connector dialect.Connector) {
	c.connector = connector
}

func Connect(config Config, useAdmin bool) (*sql.DB, error) {
	client, err := config.connector.Connect(useAdmin)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(); err != nil {
		return nil, errors.ThrowPreconditionFailed(err, "DATAB-0pIWD", "Errors.Database.Connection.Failed")
	}

	return client, nil
}

func DecodeHook(from, to reflect.Value) (interface{}, error) {
	if to.Type() != reflect.TypeOf(Config{}) {
		return from.Interface(), nil
	}

	configuredDialects, ok := from.Interface().(map[string]interface{})
	if !ok {
		return from.Interface(), nil
	}

	configuredDialect := dialect.SelectByConfig(configuredDialects)
	configs := make([]interface{}, 0, len(configuredDialects)-1)

	for name, dialectConfig := range configuredDialects {
		if !configuredDialect.Matcher.MatchName(name) {
			continue
		}

		configs = append(configs, dialectConfig)
	}

	connector, err := configuredDialect.Matcher.Decode(configs)
	if err != nil {
		return nil, err
	}

	return Config{connector: connector}, nil
}

func (c Config) Database() string {
	return c.connector.DatabaseName()
}

func (c Config) Username() string {
	return c.connector.Username()
}

func (c Config) Password() string {
	return c.connector.Password()
}

func (c Config) Type() string {
	return c.connector.Type()
}
