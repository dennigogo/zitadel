package login

import (
	"net/http"

	"github.com/dennigogo/zitadel/internal/domain"

	http_mw "github.com/dennigogo/zitadel/internal/api/http/middleware"
)

const (
	tmplUserSelection = "userselection"
)

type userSelectionFormData struct {
	UserID string `schema:"userID"`
}

func (l *Login) renderUserSelection(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest, selectionData *domain.SelectUserStep) {
	data := userSelectionData{
		baseData: l.getBaseData(r, authReq, "Select User", "", ""),
		Users:    selectionData.Users,
		Linking:  len(authReq.LinkingUsers) > 0,
	}
	translator := l.getTranslator(r.Context(), authReq)
	l.renderer.RenderTemplate(w, r, translator, l.renderer.Templates[tmplUserSelection], data, nil)
}

func (l *Login) handleSelectUser(w http.ResponseWriter, r *http.Request) {
	data := new(userSelectionFormData)
	authSession, err := l.getAuthRequestAndParseData(r, data)
	if err != nil {
		l.renderError(w, r, authSession, err)
		return
	}
	if data.UserID == "0" {
		l.renderLogin(w, r, authSession, nil)
		return
	}
	userAgentID, _ := http_mw.UserAgentIDFromCtx(r.Context())
	err = l.authRepo.SelectUser(r.Context(), authSession.ID, data.UserID, userAgentID)
	if err != nil {
		l.renderError(w, r, authSession, err)
		return
	}
	l.renderNextStep(w, r, authSession)
}
