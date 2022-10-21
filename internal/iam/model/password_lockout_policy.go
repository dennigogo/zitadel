package model

import (
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
)

type LockoutPolicy struct {
	models.ObjectRoot

	State               PolicyState
	MaxPasswordAttempts uint64
	ShowLockOutFailures bool
}
