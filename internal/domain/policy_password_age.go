package domain

import (
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
)

type PasswordAgePolicy struct {
	models.ObjectRoot

	MaxAgeDays     uint64
	ExpireWarnDays uint64
}
