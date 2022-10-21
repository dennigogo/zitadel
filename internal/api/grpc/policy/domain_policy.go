package policy

import (
	"github.com/dennigogo/zitadel/internal/api/grpc/object"
	"github.com/dennigogo/zitadel/internal/query"
	policy_pb "github.com/dennigogo/zitadel/pkg/grpc/policy"
)

func DomainPolicyToPb(policy *query.DomainPolicy) *policy_pb.DomainPolicy {
	return &policy_pb.DomainPolicy{
		UserLoginMustBeDomain:                  policy.UserLoginMustBeDomain,
		ValidateOrgDomains:                     policy.ValidateOrgDomains,
		SmtpSenderAddressMatchesInstanceDomain: policy.SMTPSenderAddressMatchesInstanceDomain,
		IsDefault:                              policy.IsDefault,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}

func DomainPolicyToOrgIAMPb(policy *query.DomainPolicy) *policy_pb.OrgIAMPolicy {
	return &policy_pb.OrgIAMPolicy{
		UserLoginMustBeDomain: policy.UserLoginMustBeDomain,
		IsDefault:             policy.IsDefault,
		Details: object.ToViewDetailsPb(
			policy.Sequence,
			policy.CreationDate,
			policy.ChangeDate,
			policy.ResourceOwner,
		),
	}
}
