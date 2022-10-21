package log

import (
	"fmt"

	"github.com/k3a/html2text"

	"github.com/dennigogo/zitadel/internal/notification/channels"
	"github.com/zitadel/logging"
)

func InitStdoutChannel(config LogConfig) channels.NotificationChannel {

	logging.Log("NOTIF-D0164").Debug("successfully initialized stdout email and sms channel")

	return channels.HandleMessageFunc(func(message channels.Message) error {

		content := message.GetContent()
		if config.Compact {
			content = html2text.HTML2Text(content)
		}

		logging.Log("NOTIF-c73ba").WithFields(map[string]interface{}{
			"type":    fmt.Sprintf("%T", message),
			"content": content,
		}).Info("handling notification message")
		return nil
	})
}
