package admin

import (
	"github.com/dennigogo/zitadel/internal/domain"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func updateLabelPolicyToDomain(policy *admin_pb.UpdateLabelPolicyRequest) *domain.LabelPolicy {
	return &domain.LabelPolicy{
		PrimaryColor:        policy.PrimaryColor,
		BackgroundColor:     policy.BackgroundColor,
		WarnColor:           policy.WarnColor,
		FontColor:           policy.FontColor,
		PrimaryColorDark:    policy.PrimaryColorDark,
		BackgroundColorDark: policy.BackgroundColorDark,
		WarnColorDark:       policy.WarnColorDark,
		FontColorDark:       policy.FontColorDark,
		HideLoginNameSuffix: policy.HideLoginNameSuffix,
		DisableWatermark:    policy.DisableWatermark,
	}
}
