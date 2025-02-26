package senders

import (
	"context"

	"github.com/dennigogo/zitadel/internal/notification/channels"
	"github.com/dennigogo/zitadel/internal/notification/channels/fs"
	"github.com/dennigogo/zitadel/internal/notification/channels/log"
)

func debugChannels(ctx context.Context, getFileSystemProvider func(ctx context.Context) (*fs.FSConfig, error), getLogProvider func(ctx context.Context) (*log.LogConfig, error)) []channels.NotificationChannel {
	var (
		providers []channels.NotificationChannel
	)

	if fsProvider, err := getFileSystemProvider(ctx); err == nil {
		p, err := fs.InitFSChannel(*fsProvider)
		if err == nil {
			providers = append(providers, p)
		}
	}

	if logProvider, err := getLogProvider(ctx); err == nil {
		providers = append(providers, log.InitStdoutChannel(*logProvider))
	}

	return providers
}
