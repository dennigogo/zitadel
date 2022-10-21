package types

import (
	"github.com/dennigogo/zitadel/internal/api/ui/login"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/query"
)

func (notify Notify) SendPasswordCode(user *query.NotifyUser, origin, code string) error {
	url := login.InitPasswordLink(origin, user.ID, code, user.ResourceOwner)
	args := make(map[string]interface{})
	args["Code"] = code
	return notify(url, args, domain.PasswordResetMessageType, true)
}
