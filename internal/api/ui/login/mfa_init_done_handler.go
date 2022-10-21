package login

import (
	"net/http"

	"github.com/dennigogo/zitadel/internal/domain"
)

const (
	tmplMFAInitDone = "mfainitdone"
)

type mfaInitDoneData struct {
}

func (l *Login) renderMFAInitDone(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest, data *mfaDoneData) {
	var errType, errMessage string
	data.baseData = l.getBaseData(r, authReq, "MFA Init Done", errType, errMessage)
	data.profileData = l.getProfileData(authReq)
	translator := l.getTranslator(r.Context(), authReq)
	l.renderer.RenderTemplate(w, r, translator, l.renderer.Templates[tmplMFAInitDone], data, nil)
}
