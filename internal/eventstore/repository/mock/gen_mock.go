package mock

//go:generate mockgen -package mock -destination ./repository.mock.go github.com/dennigogo/zitadel/internal/eventstore/repository Repository
