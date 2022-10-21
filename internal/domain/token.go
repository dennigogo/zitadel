package domain

import (
	"context"
	"strings"
	"time"

	"github.com/dennigogo/zitadel/internal/api/authz"
	es_models "github.com/dennigogo/zitadel/internal/eventstore/v1/models"
)

type Token struct {
	es_models.ObjectRoot

	TokenID           string
	ApplicationID     string
	UserAgentID       string
	RefreshTokenID    string
	Audience          []string
	Expiration        time.Time
	Scopes            []string
	PreferredLanguage string
}

func AddAudScopeToAudience(ctx context.Context, audience, scopes []string) []string {
	for _, scope := range scopes {
		if !(strings.HasPrefix(scope, ProjectIDScope) && strings.HasSuffix(scope, AudSuffix)) {
			continue
		}
		projectID := strings.TrimSuffix(strings.TrimPrefix(scope, ProjectIDScope), AudSuffix)
		if projectID == ProjectIDScopeZITADEL {
			projectID = authz.GetInstance(ctx).ProjectID()
		}
		audience = append(audience, projectID)
	}
	return audience
}
