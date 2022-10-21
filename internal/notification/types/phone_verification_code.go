package types

import (
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/query"
)

func (notify Notify) SendPhoneVerificationCode(user *query.NotifyUser, origin, code string) error {
	args := make(map[string]interface{})
	args["Code"] = code
	return notify("", args, domain.VerifyPhoneMessageType, true)
}
