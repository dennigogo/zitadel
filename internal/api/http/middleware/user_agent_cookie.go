package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	http_utils "github.com/dennigogo/zitadel/internal/api/http"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/id"
)

type cookieKey int

var (
	userAgentKey cookieKey = 0
)

func UserAgentIDFromCtx(ctx context.Context) (string, bool) {
	userAgentID, ok := ctx.Value(userAgentKey).(string)
	return userAgentID, ok
}

type UserAgent struct {
	ID string
}

type userAgentHandler struct {
	cookieHandler   *http_utils.CookieHandler
	cookieName      string
	idGenerator     id.Generator
	nextHandler     http.Handler
	ignoredPrefixes []string
}

type UserAgentCookieConfig struct {
	Name   string
	MaxAge time.Duration
}

func NewUserAgentHandler(config *UserAgentCookieConfig, cookieKey []byte, idGenerator id.Generator, externalSecure bool, ignoredPrefixes ...string) (func(http.Handler) http.Handler, error) {
	opts := []http_utils.CookieHandlerOpt{
		http_utils.WithEncryption(cookieKey, cookieKey),
		http_utils.WithMaxAge(int(config.MaxAge.Seconds())),
	}
	if !externalSecure {
		opts = append(opts, http_utils.WithUnsecure())
	}
	return func(handler http.Handler) http.Handler {
		return &userAgentHandler{
			nextHandler:     handler,
			cookieName:      config.Name,
			cookieHandler:   http_utils.NewCookieHandler(opts...),
			idGenerator:     idGenerator,
			ignoredPrefixes: ignoredPrefixes,
		}
	}, nil
}

func (ua *userAgentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range ua.ignoredPrefixes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			ua.nextHandler.ServeHTTP(w, r)
			return
		}
	}
	agent, err := ua.getUserAgent(r)
	if err != nil {
		agent, err = ua.newUserAgent()
	}
	if err == nil {
		ctx := context.WithValue(r.Context(), userAgentKey, agent.ID)
		r = r.WithContext(ctx)
		ua.setUserAgent(w, r.Host, agent)
	}
	ua.nextHandler.ServeHTTP(w, r)
}

func (ua *userAgentHandler) newUserAgent() (*UserAgent, error) {
	agentID, err := ua.idGenerator.Next()
	if err != nil {
		return nil, err
	}
	return &UserAgent{ID: agentID}, nil
}

func (ua *userAgentHandler) getUserAgent(r *http.Request) (*UserAgent, error) {
	userAgent := new(UserAgent)
	err := ua.cookieHandler.GetEncryptedCookieValue(r, ua.cookieName, userAgent)
	if err != nil {
		return nil, errors.ThrowPermissionDenied(err, "HTTP-YULqH4", "cannot read user agent cookie")
	}
	return userAgent, nil
}

func (ua *userAgentHandler) setUserAgent(w http.ResponseWriter, host string, agent *UserAgent) error {
	err := ua.cookieHandler.SetEncryptedCookie(w, ua.cookieName, host, agent)
	if err != nil {
		return errors.ThrowPermissionDenied(err, "HTTP-AqgqdA", "cannot set user agent cookie")
	}
	return nil
}
