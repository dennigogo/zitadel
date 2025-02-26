package policy

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/query"
	policy_pb "github.com/dennigogo/zitadel/pkg/grpc/policy"
)

func ModelLabelPolicyToPb(policy *query.LabelPolicy, assetPrefix string) *policy_pb.LabelPolicy {
	return &policy_pb.LabelPolicy{
		IsDefault:           policy.IsDefault,
		PrimaryColor:        policy.Light.PrimaryColor,
		BackgroundColor:     policy.Light.BackgroundColor,
		FontColor:           policy.Light.FontColor,
		WarnColor:           policy.Light.WarnColor,
		PrimaryColorDark:    policy.Dark.PrimaryColor,
		BackgroundColorDark: policy.Dark.BackgroundColor,
		WarnColorDark:       policy.Dark.WarnColor,
		FontColorDark:       policy.Dark.FontColor,
		FontUrl:             domain.AssetURL(assetPrefix, policy.ResourceOwner, policy.FontURL),
		LogoUrl:             domain.AssetURL(assetPrefix, policy.ResourceOwner, policy.Light.LogoURL),
		LogoUrlDark:         domain.AssetURL(assetPrefix, policy.ResourceOwner, policy.Dark.LogoURL),
		IconUrl:             domain.AssetURL(assetPrefix, policy.ResourceOwner, policy.Light.IconURL),
		IconUrlDark:         domain.AssetURL(assetPrefix, policy.ResourceOwner, policy.Dark.IconURL),

		DisableWatermark:    policy.WatermarkDisabled,
		HideLoginNameSuffix: policy.HideLoginNameSuffix,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
