package login

import (
	"net/http"

	http_mw "github.com/dennigogo/zitadel/internal/api/http/middleware"
	"github.com/dennigogo/zitadel/internal/domain"
)

const (
	tmplMFAVerify = "mfaverify"
)

type mfaVerifyFormData struct {
	MFAType          domain.MFAType `schema:"mfaType"`
	Code             string         `schema:"code"`
	SelectedProvider domain.MFAType `schema:"provider"`
}

func (l *Login) handleMFAVerify(w http.ResponseWriter, r *http.Request) {
	data := new(mfaVerifyFormData)
	authReq, err := l.getAuthRequestAndParseData(r, data)
	if err != nil {
		l.renderError(w, r, authReq, err)
		return
	}
	step, ok := authReq.PossibleSteps[0].(*domain.MFAVerificationStep)
	if !ok {
		l.renderError(w, r, authReq, err)
		return
	}
	if data.Code == "" {
		l.renderMFAVerifySelected(w, r, authReq, step, data.SelectedProvider, nil)
		return
	}
	if data.MFAType == domain.MFATypeOTP {
		userAgentID, _ := http_mw.UserAgentIDFromCtx(r.Context())
		err = l.authRepo.VerifyMFAOTP(setContext(r.Context(), authReq.UserOrgID), authReq.ID, authReq.UserID, authReq.UserOrgID, data.Code, userAgentID, domain.BrowserInfoFromRequest(r))
		if err != nil {
			l.renderMFAVerifySelected(w, r, authReq, step, domain.MFATypeOTP, err)
			return
		}
	}
	l.renderNextStep(w, r, authReq)
}

func (l *Login) renderMFAVerify(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest, verificationStep *domain.MFAVerificationStep, err error) {
	if verificationStep == nil {
		l.renderError(w, r, authReq, err)
		return
	}
	provider := verificationStep.MFAProviders[len(verificationStep.MFAProviders)-1]
	l.renderMFAVerifySelected(w, r, authReq, verificationStep, provider, err)
}

func (l *Login) renderMFAVerifySelected(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest, verificationStep *domain.MFAVerificationStep, selectedProvider domain.MFAType, err error) {
	var errID, errMessage string
	if err != nil {
		errID, errMessage = l.getErrorMessage(r, err)
	}
	data := l.getUserData(r, authReq, "MFA Verify", errID, errMessage)
	if verificationStep == nil {
		l.renderError(w, r, authReq, err)
		return
	}
	switch selectedProvider {
	case domain.MFATypeU2F:
		l.renderU2FVerification(w, r, authReq, removeSelectedProviderFromList(verificationStep.MFAProviders, domain.MFATypeU2F), nil)
		return
	case domain.MFATypeOTP:
		data.MFAProviders = removeSelectedProviderFromList(verificationStep.MFAProviders, domain.MFATypeOTP)
		data.SelectedMFAProvider = domain.MFATypeOTP
	default:
		l.renderError(w, r, authReq, err)
		return
	}
	l.renderer.RenderTemplate(w, r, l.getTranslator(r.Context(), authReq), l.renderer.Templates[tmplMFAVerify], data, nil)
}

func removeSelectedProviderFromList(providers []domain.MFAType, selected domain.MFAType) []domain.MFAType {
	for i := len(providers) - 1; i >= 0; i-- {
		if providers[i] == selected {
			copy(providers[i:], providers[i+1:])
			return providers[:len(providers)-1]
		}
	}
	return providers
}
