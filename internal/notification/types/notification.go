package types

import (
	"context"

	"github.com/dennigogo/zitadel/internal/i18n"
	"github.com/dennigogo/zitadel/internal/notification/channels/fs"
	"github.com/dennigogo/zitadel/internal/notification/channels/log"
	"github.com/dennigogo/zitadel/internal/notification/channels/smtp"
	"github.com/dennigogo/zitadel/internal/notification/channels/twilio"
	"github.com/dennigogo/zitadel/internal/notification/templates"
	"github.com/dennigogo/zitadel/internal/query"
)

type Notify func(
	url string,
	args map[string]interface{},
	messageType string,
	allowUnverifiedNotificationChannel bool,
) error

func SendEmail(
	ctx context.Context,
	mailhtml string,
	translator *i18n.Translator,
	user *query.NotifyUser,
	emailConfig func(ctx context.Context) (*smtp.EmailConfig, error),
	getFileSystemProvider func(ctx context.Context) (*fs.FSConfig, error),
	getLogProvider func(ctx context.Context) (*log.LogConfig, error),
	colors *query.LabelPolicy,
	assetsPrefix string,
) Notify {
	return func(
		url string,
		args map[string]interface{},
		messageType string,
		allowUnverifiedNotificationChannel bool,
	) error {
		args = mapNotifyUserToArgs(user, args)
		data := GetTemplateData(translator, args, assetsPrefix, url, messageType, user.PreferredLanguage.String(), colors)
		template, err := templates.GetParsedTemplate(mailhtml, data)
		if err != nil {
			return err
		}
		return generateEmail(ctx, user, data.Subject, template, emailConfig, getFileSystemProvider, getLogProvider, allowUnverifiedNotificationChannel)
	}
}

func SendSMSTwilio(
	ctx context.Context,
	translator *i18n.Translator,
	user *query.NotifyUser,
	twilioConfig func(ctx context.Context) (*twilio.TwilioConfig, error),
	getFileSystemProvider func(ctx context.Context) (*fs.FSConfig, error),
	getLogProvider func(ctx context.Context) (*log.LogConfig, error),
	colors *query.LabelPolicy,
	assetsPrefix string,
) Notify {
	return func(
		url string,
		args map[string]interface{},
		messageType string,
		allowUnverifiedNotificationChannel bool,
	) error {
		args = mapNotifyUserToArgs(user, args)
		data := GetTemplateData(translator, args, assetsPrefix, url, messageType, user.PreferredLanguage.String(), colors)
		return generateSms(ctx, user, data.Text, twilioConfig, getFileSystemProvider, getLogProvider, allowUnverifiedNotificationChannel)
	}
}

func externalLink(origin string) string {
	return origin + "/ui/login"
}
