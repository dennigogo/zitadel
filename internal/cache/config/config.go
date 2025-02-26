package config

import (
	"encoding/json"

	"github.com/dennigogo/zitadel/internal/cache"
	"github.com/dennigogo/zitadel/internal/cache/bigcache"
	"github.com/dennigogo/zitadel/internal/cache/fastcache"
	"github.com/dennigogo/zitadel/internal/errors"
)

type CacheConfig struct {
	Type   string
	Config cache.Config
}

var caches = map[string]func() cache.Config{
	"bigcache":  func() cache.Config { return &bigcache.Config{} },
	"fastcache": func() cache.Config { return &fastcache.Config{} },
}

func (c *CacheConfig) UnmarshalJSON(data []byte) error {
	var rc struct {
		Type   string
		Config json.RawMessage
	}

	if err := json.Unmarshal(data, &rc); err != nil {
		return errors.ThrowInternal(err, "CONFI-98ejs", "unable to unmarshal config")
	}

	c.Type = rc.Type

	var err error
	c.Config, err = newCacheConfig(c.Type, rc.Config)
	if err != nil {
		return errors.ThrowInternal(err, "CONFI-do9es", "unable create config")
	}

	return nil
}

func newCacheConfig(cacheType string, configData []byte) (cache.Config, error) {
	t, ok := caches[cacheType]
	if !ok {
		return nil, errors.ThrowInternal(nil, "CONFI-di328s", "no config")
	}

	cacheConfig := t()
	if len(configData) == 0 {
		return cacheConfig, nil
	}

	if err := json.Unmarshal(configData, cacheConfig); err != nil {
		return nil, errors.ThrowInternal(nil, "CONFI-skei3", "could not read config")
	}

	return cacheConfig, nil
}
