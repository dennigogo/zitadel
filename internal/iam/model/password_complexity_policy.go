package model

import (
	"regexp"

	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/v1/models"
)

var (
	hasStringLowerCase = regexp.MustCompile(`[a-z]`).MatchString
	hasStringUpperCase = regexp.MustCompile(`[A-Z]`).MatchString
	hasNumber          = regexp.MustCompile(`[0-9]`).MatchString
	hasSymbol          = regexp.MustCompile(`[^A-Za-z0-9]`).MatchString
)

type PasswordComplexityPolicy struct {
	models.ObjectRoot

	State        PolicyState
	MinLength    uint64
	HasLowercase bool
	HasUppercase bool
	HasNumber    bool
	HasSymbol    bool

	Default bool
}

func (p *PasswordComplexityPolicy) IsValid() error {
	if p.MinLength == 0 || p.MinLength > 72 {
		return caos_errs.ThrowInvalidArgument(nil, "MODEL-Lsp0e", "Errors.User.PasswordComplexityPolicy.MinLengthNotAllowed")
	}
	return nil
}

func (p *PasswordComplexityPolicy) Check(password string) error {
	if p.MinLength != 0 && uint64(len(password)) < p.MinLength {
		return caos_errs.ThrowInvalidArgument(nil, "MODEL-HuJf6", "Errors.User.PasswordComplexityPolicy.MinLength")
	}

	if p.HasLowercase && !hasStringLowerCase(password) {
		return caos_errs.ThrowInvalidArgument(nil, "MODEL-co3Xw", "Errors.User.PasswordComplexityPolicy.HasLower")
	}

	if p.HasUppercase && !hasStringUpperCase(password) {
		return caos_errs.ThrowInvalidArgument(nil, "MODEL-VoaRj", "Errors.User.PasswordComplexityPolicy.HasUpper")
	}

	if p.HasNumber && !hasNumber(password) {
		return caos_errs.ThrowInvalidArgument(nil, "MODEL-ZBv4H", "Errors.User.PasswordComplexityPolicy.HasNumber")
	}

	if p.HasSymbol && !hasSymbol(password) {
		return caos_errs.ThrowInvalidArgument(nil, "MODEL-ZDLwA", "Errors.User.PasswordComplexityPolicy.HasSymbol")
	}
	return nil
}
