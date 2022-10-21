package model

import (
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
)

type DomainPolicy struct {
	models.ObjectRoot

	State                 PolicyState
	UserLoginMustBeDomain bool
	Default               bool
}
