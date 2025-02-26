package middleware

import (
	"net/http"

	"github.com/dennigogo/zitadel/internal/telemetry"

	http_utils "github.com/dennigogo/zitadel/internal/api/http"
)

func DefaultTelemetryHandler(handler http.Handler) http.Handler {
	return TelemetryHandler(http_utils.Probes...)(handler)
}

func TelemetryHandler(ignoredMethods ...string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return telemetry.TelemetryHandler(handler, ignoredMethods...)
	}
}
