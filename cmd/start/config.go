package start

import (
	"time"

	"github.com/dennigogo/zitadel/internal/api/saml"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/actions"
	admin_es "github.com/dennigogo/zitadel/internal/admin/repository/eventsourcing"
	internal_authz "github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/http/middleware"
	"github.com/dennigogo/zitadel/internal/api/oidc"
	"github.com/dennigogo/zitadel/internal/api/ui/console"
	"github.com/dennigogo/zitadel/internal/api/ui/login"
	auth_es "github.com/dennigogo/zitadel/internal/auth/repository/eventsourcing"
	"github.com/dennigogo/zitadel/internal/command"
	"github.com/dennigogo/zitadel/internal/config/hook"
	"github.com/dennigogo/zitadel/internal/config/network"
	"github.com/dennigogo/zitadel/internal/config/systemdefaults"
	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/database"
	"github.com/dennigogo/zitadel/internal/id"
	"github.com/dennigogo/zitadel/internal/query/projection"
	static_config "github.com/dennigogo/zitadel/internal/static/config"
	metrics "github.com/dennigogo/zitadel/internal/telemetry/metrics/config"
	tracing "github.com/dennigogo/zitadel/internal/telemetry/tracing/config"
)

type Config struct {
	Log               *logging.Config
	Port              uint16
	ExternalPort      uint16
	ExternalDomain    string
	ExternalSecure    bool
	TLS               network.TLS
	HTTP2HostHeader   string
	HTTP1HostHeader   string
	WebAuthNName      string
	Database          database.Config
	Tracing           tracing.Config
	Metrics           metrics.Config
	Projections       projection.Config
	Auth              auth_es.Config
	Admin             admin_es.Config
	UserAgentCookie   *middleware.UserAgentCookieConfig
	OIDC              oidc.Config
	SAML              saml.Config
	Login             login.Config
	Console           console.Config
	AssetStorage      static_config.AssetStorageConfig
	InternalAuthZ     internal_authz.Config
	SystemDefaults    systemdefaults.SystemDefaults
	EncryptionKeys    *encryptionKeyConfig
	DefaultInstance   command.InstanceSetup
	AuditLogRetention time.Duration
	SystemAPIUsers    map[string]*internal_authz.SystemAPIUser
	CustomerPortal    string
	Machine           *id.Config
	Actions           *actions.Config
}

func MustNewConfig(v *viper.Viper) *Config {
	config := new(Config)

	err := v.Unmarshal(config,
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			hook.Base64ToBytesHookFunc(),
			hook.TagToLanguageHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			database.DecodeHook,
			actions.HTTPConfigDecodeHook,
		)),
	)
	logging.OnError(err).Fatal("unable to read config")

	err = config.Log.SetLogger()
	logging.OnError(err).Fatal("unable to set logger")

	err = config.Tracing.NewTracer()
	logging.OnError(err).Fatal("unable to set tracer")

	err = config.Metrics.NewMeter()
	logging.OnError(err).Fatal("unable to set meter")

	id.Configure(config.Machine)
	actions.SetHTTPConfig(&config.Actions.HTTP)

	return config
}

type encryptionKeyConfig struct {
	DomainVerification   *crypto.KeyConfig
	IDPConfig            *crypto.KeyConfig
	OIDC                 *crypto.KeyConfig
	SAML                 *crypto.KeyConfig
	OTP                  *crypto.KeyConfig
	SMS                  *crypto.KeyConfig
	SMTP                 *crypto.KeyConfig
	User                 *crypto.KeyConfig
	CSRFCookieKeyID      string
	UserAgentCookieKeyID string
}
