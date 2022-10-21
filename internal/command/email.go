package command

import (
	"context"
	"time"

	"github.com/dennigogo/zitadel/internal/command/preparation"
	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
)

type Email struct {
	Address  string
	Verified bool
}

func (e *Email) Valid() bool {
	return e.Address != "" && domain.EmailRegex.MatchString(e.Address)
}

func newEmailCode(ctx context.Context, filter preparation.FilterToQueryReducer, alg crypto.EncryptionAlgorithm) (value *crypto.CryptoValue, expiry time.Duration, err error) {
	return newCryptoCodeWithExpiry(ctx, filter, domain.SecretGeneratorTypeVerifyEmailCode, alg)
}
