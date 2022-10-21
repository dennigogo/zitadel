package settings

import (
	obj_pb "github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/query"
	settings_pb "github.com/dennigogo/zitadel/pkg/grpc/settings"
)

func NotificationProviderToPb(provider *query.DebugNotificationProvider) *settings_pb.DebugNotificationProvider {
	mapped := &settings_pb.DebugNotificationProvider{
		Compact: provider.Compact,
		Details: obj_pb.ToViewDetailsPb(provider.Sequence, provider.CreationDate, provider.ChangeDate, provider.AggregateID),
	}
	return mapped
}
