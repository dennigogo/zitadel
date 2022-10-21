package login

import (
	"net/http"

	"github.com/dennigogo/zitadel/internal/domain"
)

const (
	tmplChangeUsername     = "changeusername"
	tmplChangeUsernameDone = "changeusernamedone"
)

type changeUsernameData struct {
	Username string `schema:"username"`
}

func (l *Login) renderChangeUsername(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest, err error) {
	var errID, errMessage string
	if err != nil {
		errID, errMessage = l.getErrorMessage(r, err)
	}
	data := l.getUserData(r, authReq, "Change Username", errID, errMessage)
	l.renderer.RenderTemplate(w, r, l.getTranslator(r.Context(), authReq), l.renderer.Templates[tmplChangeUsername], data, nil)
}

func (l *Login) handleChangeUsername(w http.ResponseWriter, r *http.Request) {
	data := new(changeUsernameData)
	authReq, err := l.getAuthRequestAndParseData(r, data)
	if err != nil {
		l.renderError(w, r, authReq, err)
		return
	}
	_, err = l.command.ChangeUsername(setContext(r.Context(), authReq.UserOrgID), authReq.UserOrgID, authReq.UserID, data.Username)
	if err != nil {
		l.renderChangeUsername(w, r, authReq, err)
		return
	}
	l.renderChangeUsernameDone(w, r, authReq)
}

func (l *Login) renderChangeUsernameDone(w http.ResponseWriter, r *http.Request, authReq *domain.AuthRequest) {
	var errType, errMessage string
	data := l.getUserData(r, authReq, "Username Change Done", errType, errMessage)
	l.renderer.RenderTemplate(w, r, l.getTranslator(r.Context(), authReq), l.renderer.Templates[tmplChangeUsernameDone], data, nil)
}
