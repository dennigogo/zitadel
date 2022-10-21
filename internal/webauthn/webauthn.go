package webauthn

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/http"
	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
)

type Config struct {
	DisplayName    string
	ExternalSecure bool
}

type webUser struct {
	*domain.Human
	accountName string
	credentials []webauthn.Credential
}

func (u *webUser) WebAuthnID() []byte {
	return []byte(u.AggregateID)
}

func (u *webUser) WebAuthnName() string {
	if u.accountName != "" {
		return u.accountName
	}
	return u.GetUsername()
}

func (u *webUser) WebAuthnDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.GetUsername()
}

func (u *webUser) WebAuthnIcon() string {
	return ""
}

func (u *webUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (w *Config) BeginRegistration(ctx context.Context, user *domain.Human, accountName string, authType domain.AuthenticatorAttachment, userVerification domain.UserVerificationRequirement, isLoginUI bool, webAuthNs ...*domain.WebAuthNToken) (*domain.WebAuthNToken, error) {
	webAuthNServer, err := w.serverFromContext(ctx)
	if err != nil {
		return nil, err
	}
	creds := WebAuthNsToCredentials(webAuthNs)
	existing := make([]protocol.CredentialDescriptor, len(creds))
	for i, cred := range creds {
		existing[i] = protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.ID,
		}
	}
	credentialOptions, sessionData, err := webAuthNServer.BeginRegistration(
		&webUser{
			Human:       user,
			accountName: accountName,
			credentials: creds,
		},
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			UserVerification:        UserVerificationFromDomain(userVerification),
			AuthenticatorAttachment: AuthenticatorAttachmentFromDomain(authType),
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
		webauthn.WithExclusions(existing),
	)
	if err != nil {
		return nil, caos_errs.ThrowInternal(err, "WEBAU-bM8sd", "Errors.User.WebAuthN.BeginRegisterFailed")
	}
	cred, err := json.Marshal(credentialOptions)
	if err != nil {
		return nil, caos_errs.ThrowInternal(err, "WEBAU-D7cus", "Errors.User.WebAuthN.MarshalError")
	}
	return &domain.WebAuthNToken{
		Challenge:              sessionData.Challenge,
		CredentialCreationData: cred,
		AllowedCredentialIDs:   sessionData.AllowedCredentialIDs,
		UserVerification:       UserVerificationToDomain(sessionData.UserVerification),
	}, nil
}

func (w *Config) FinishRegistration(ctx context.Context, user *domain.Human, webAuthN *domain.WebAuthNToken, tokenName string, credData []byte, isLoginUI bool) (*domain.WebAuthNToken, error) {
	if webAuthN == nil {
		return nil, caos_errs.ThrowInternal(nil, "WEBAU-5M9so", "Errors.User.WebAuthN.NotFound")
	}
	credentialData, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(credData))
	if err != nil {
		e := *err.(*protocol.Error)
		logging.WithFields("error", e).Error("webauthn credential could not be parsed")
		return nil, caos_errs.ThrowInternal(err, "WEBAU-sEr8c", "Errors.User.WebAuthN.ErrorOnParseCredential")
	}
	sessionData := WebAuthNToSessionData(webAuthN)
	webAuthNServer, err := w.serverFromContext(ctx)
	if err != nil {
		return nil, err
	}
	credential, err := webAuthNServer.CreateCredential(
		&webUser{
			Human: user,
		},
		sessionData,
		credentialData)
	if err != nil {
		return nil, caos_errs.ThrowInternal(err, "WEBAU-3Vb9s", "Errors.User.WebAuthN.CreateCredentialFailed")
	}

	webAuthN.KeyID = credential.ID
	webAuthN.PublicKey = credential.PublicKey
	webAuthN.AttestationType = credential.AttestationType
	webAuthN.AAGUID = credential.Authenticator.AAGUID
	webAuthN.SignCount = credential.Authenticator.SignCount
	webAuthN.WebAuthNTokenName = tokenName
	return webAuthN, nil
}

func (w *Config) BeginLogin(ctx context.Context, user *domain.Human, userVerification domain.UserVerificationRequirement, webAuthNs ...*domain.WebAuthNToken) (*domain.WebAuthNLogin, error) {
	webAuthNServer, err := w.serverFromContext(ctx)
	if err != nil {
		return nil, err
	}
	assertion, sessionData, err := webAuthNServer.BeginLogin(&webUser{
		Human:       user,
		credentials: WebAuthNsToCredentials(webAuthNs),
	}, webauthn.WithUserVerification(UserVerificationFromDomain(userVerification)))
	if err != nil {
		return nil, caos_errs.ThrowInternal(err, "WEBAU-4G8sw", "Errors.User.WebAuthN.BeginLoginFailed")
	}
	cred, err := json.Marshal(assertion)
	if err != nil {
		return nil, caos_errs.ThrowInternal(err, "WEBAU-2M0s9", "Errors.User.WebAuthN.MarshalError")
	}
	return &domain.WebAuthNLogin{
		Challenge:               sessionData.Challenge,
		CredentialAssertionData: cred,
		AllowedCredentialIDs:    sessionData.AllowedCredentialIDs,
		UserVerification:        userVerification,
	}, nil
}

func (w *Config) FinishLogin(ctx context.Context, user *domain.Human, webAuthN *domain.WebAuthNLogin, credData []byte, webAuthNs ...*domain.WebAuthNToken) ([]byte, uint32, error) {
	assertionData, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(credData))
	if err != nil {
		return nil, 0, caos_errs.ThrowInternal(err, "WEBAU-ADgv4", "Errors.User.WebAuthN.ValidateLoginFailed")
	}
	webUser := &webUser{
		Human:       user,
		credentials: WebAuthNsToCredentials(webAuthNs),
	}
	webAuthNServer, err := w.serverFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	credential, err := webAuthNServer.ValidateLogin(webUser, WebAuthNLoginToSessionData(webAuthN), assertionData)
	if err != nil {
		return nil, 0, caos_errs.ThrowInternal(err, "WEBAU-3M9si", "Errors.User.WebAuthN.ValidateLoginFailed")
	}

	if credential.Authenticator.CloneWarning {
		return credential.ID, credential.Authenticator.SignCount, caos_errs.ThrowInternal(err, "WEBAU-4M90s", "Errors.User.WebAuthN.CloneWarning")
	}
	return credential.ID, credential.Authenticator.SignCount, nil
}

func (w *Config) serverFromContext(ctx context.Context) (*webauthn.WebAuthn, error) {
	instance := authz.GetInstance(ctx)
	return webauthn.New(&webauthn.Config{
		RPDisplayName: w.DisplayName,
		RPID:          instance.RequestedDomain(),
		RPOrigin:      http.BuildOrigin(instance.RequestedHost(), w.ExternalSecure),
	})
}
