package management

import (
	"github.com/dennigogo/zitadel/internal/domain"
	mgmt_pb "github.com/dennigogo/zitadel/pkg/grpc/management"
)

func AddLabelPolicyToDomain(p *mgmt_pb.AddCustomLabelPolicyRequest) *domain.LabelPolicy {
	return &domain.LabelPolicy{
		PrimaryColor:        p.PrimaryColor,
		BackgroundColor:     p.BackgroundColor,
		WarnColor:           p.WarnColor,
		PrimaryColorDark:    p.PrimaryColorDark,
		BackgroundColorDark: p.BackgroundColorDark,
		WarnColorDark:       p.WarnColorDark,
		HideLoginNameSuffix: p.HideLoginNameSuffix,
		DisableWatermark:    p.DisableWatermark,
	}
}

func updateLabelPolicyToDomain(p *mgmt_pb.UpdateCustomLabelPolicyRequest) *domain.LabelPolicy {
	return &domain.LabelPolicy{
		PrimaryColor:        p.PrimaryColor,
		BackgroundColor:     p.BackgroundColor,
		WarnColor:           p.WarnColor,
		FontColor:           p.FontColor,
		PrimaryColorDark:    p.PrimaryColorDark,
		BackgroundColorDark: p.BackgroundColorDark,
		WarnColorDark:       p.WarnColorDark,
		FontColorDark:       p.FontColorDark,
		HideLoginNameSuffix: p.HideLoginNameSuffix,
		DisableWatermark:    p.DisableWatermark,
	}
}
