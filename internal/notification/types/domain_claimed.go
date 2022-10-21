package types

import (
	"strings"

	"github.com/dennigogo/zitadel/internal/api/ui/login"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/query"
)

func (notify Notify) SendDomainClaimed(user *query.NotifyUser, origin, username string) error {
	url := login.LoginLink(origin, user.ResourceOwner)
	args := make(map[string]interface{})
	args["TempUsername"] = username
	args["Domain"] = strings.Split(user.LastEmail, "@")[1]
	return notify(url, args, domain.DomainClaimedMessageType, true)
}
