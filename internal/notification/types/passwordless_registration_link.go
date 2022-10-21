package types

import (
	"github.com/dennigogo/zitadel/internal/api/ui/login"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/query"
)

func (notify Notify) SendPasswordlessRegistrationLink(user *query.NotifyUser, origin, code, codeID string) error {
	url := domain.PasswordlessInitCodeLink(origin+login.HandlerPrefix+login.EndpointPasswordlessRegistration, user.ID, user.ResourceOwner, codeID, code)
	return notify(url, nil, domain.PasswordlessRegistrationMessageType, true)
}
