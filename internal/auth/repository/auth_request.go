package repository

import (
	"context"

	"github.com/dennigogo/zitadel/internal/domain"
)

type AuthRequestRepository interface {
	CreateAuthRequest(ctx context.Context, request *domain.AuthRequest) (*domain.AuthRequest, error)
	AuthRequestByID(ctx context.Context, id, userAgentID string) (*domain.AuthRequest, error)
	AuthRequestByIDCheckLoggedIn(ctx context.Context, id, userAgentID string) (*domain.AuthRequest, error)
	AuthRequestByCode(ctx context.Context, code string) (*domain.AuthRequest, error)
	SaveAuthCode(ctx context.Context, id, code, userAgentID string) error
	DeleteAuthRequest(ctx context.Context, id string) error

	CheckLoginName(ctx context.Context, id, loginName, userAgentID string) error
	CheckExternalUserLogin(ctx context.Context, authReqID, userAgentID string, user *domain.ExternalUser, info *domain.BrowserInfo) error
	SetExternalUserLogin(ctx context.Context, authReqID, userAgentID string, user *domain.ExternalUser) error
	SelectUser(ctx context.Context, id, userID, userAgentID string) error
	SelectExternalIDP(ctx context.Context, authReqID, idpConfigID, userAgentID string) error
	VerifyPassword(ctx context.Context, id, userID, resourceOwner, password, userAgentID string, info *domain.BrowserInfo) error

	VerifyMFAOTP(ctx context.Context, authRequestID, userID, resourceOwner, code, userAgentID string, info *domain.BrowserInfo) error
	BeginMFAU2FLogin(ctx context.Context, userID, resourceOwner, authRequestID, userAgentID string) (*domain.WebAuthNLogin, error)
	VerifyMFAU2F(ctx context.Context, userID, resourceOwner, authRequestID, userAgentID string, credentialData []byte, info *domain.BrowserInfo) error
	BeginPasswordlessSetup(ctx context.Context, userID, resourceOwner string, preferredPlatformType domain.AuthenticatorAttachment) (login *domain.WebAuthNToken, err error)
	VerifyPasswordlessSetup(ctx context.Context, userID, resourceOwner, userAgentID, tokenName string, credentialData []byte) (err error)
	BeginPasswordlessInitCodeSetup(ctx context.Context, userID, resourceOwner, codeID, verificationCode string, preferredPlatformType domain.AuthenticatorAttachment) (login *domain.WebAuthNToken, err error)
	VerifyPasswordlessInitCodeSetup(ctx context.Context, userID, resourceOwner, userAgentID, tokenName, codeID, verificationCode string, credentialData []byte) (err error)
	BeginPasswordlessLogin(ctx context.Context, userID, resourceOwner, authRequestID, userAgentID string) (*domain.WebAuthNLogin, error)
	VerifyPasswordless(ctx context.Context, userID, resourceOwner, authRequestID, userAgentID string, credentialData []byte, info *domain.BrowserInfo) error

	LinkExternalUsers(ctx context.Context, authReqID, userAgentID string, info *domain.BrowserInfo) error
	AutoRegisterExternalUser(ctx context.Context, user *domain.Human, externalIDP *domain.UserIDPLink, orgMemberRoles []string, authReqID, userAgentID, resourceOwner string, metadatas []*domain.Metadata, info *domain.BrowserInfo) error
	ResetLinkingUsers(ctx context.Context, authReqID, userAgentID string) error
}
