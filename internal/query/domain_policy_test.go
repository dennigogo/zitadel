package query

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/dennigogo/zitadel/internal/domain"
	errs "github.com/dennigogo/zitadel/internal/errors"
)

func Test_DomainPolicyPrepares(t *testing.T) {
	type want struct {
		sqlExpectations sqlExpectation
		err             checkErr
	}
	tests := []struct {
		name    string
		prepare interface{}
		want    want
		object  interface{}
	}{
		{
			name:    "prepareDomainPolicyQuery no result",
			prepare: prepareDomainPolicyQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.domain_policies.id,`+
						` projections.domain_policies.sequence,`+
						` projections.domain_policies.creation_date,`+
						` projections.domain_policies.change_date,`+
						` projections.domain_policies.resource_owner,`+
						` projections.domain_policies.user_login_must_be_domain,`+
						` projections.domain_policies.validate_org_domains,`+
						` projections.domain_policies.smtp_sender_address_matches_instance_domain,`+
						` projections.domain_policies.is_default,`+
						` projections.domain_policies.state`+
						` FROM projections.domain_policies`),
					nil,
					nil,
				),
				err: func(err error) (error, bool) {
					if !errs.IsNotFound(err) {
						return fmt.Errorf("err should be zitadel.NotFoundError got: %w", err), false
					}
					return nil, true
				},
			},
			object: (*DomainPolicy)(nil),
		},
		{
			name:    "prepareDomainPolicyQuery found",
			prepare: prepareDomainPolicyQuery,
			want: want{
				sqlExpectations: mockQuery(
					regexp.QuoteMeta(`SELECT projections.domain_policies.id,`+
						` projections.domain_policies.sequence,`+
						` projections.domain_policies.creation_date,`+
						` projections.domain_policies.change_date,`+
						` projections.domain_policies.resource_owner,`+
						` projections.domain_policies.user_login_must_be_domain,`+
						` projections.domain_policies.validate_org_domains,`+
						` projections.domain_policies.smtp_sender_address_matches_instance_domain,`+
						` projections.domain_policies.is_default,`+
						` projections.domain_policies.state`+
						` FROM projections.domain_policies`),
					[]string{
						"id",
						"sequence",
						"creation_date",
						"change_date",
						"resource_owner",
						"user_login_must_be_domain",
						"validate_org_domains",
						"smtp_sender_address_matches_instance_domain",
						"is_default",
						"state",
					},
					[]driver.Value{
						"pol-id",
						uint64(20211109),
						testNow,
						testNow,
						"ro",
						true,
						true,
						true,
						true,
						domain.PolicyStateActive,
					},
				),
			},
			object: &DomainPolicy{
				ID:                                     "pol-id",
				CreationDate:                           testNow,
				ChangeDate:                             testNow,
				Sequence:                               20211109,
				ResourceOwner:                          "ro",
				State:                                  domain.PolicyStateActive,
				UserLoginMustBeDomain:                  true,
				ValidateOrgDomains:                     true,
				SMTPSenderAddressMatchesInstanceDomain: true,
				IsDefault:                              true,
			},
		},
		{
			name:    "prepareDomainPolicyQuery sql err",
			prepare: prepareDomainPolicyQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.domain_policies.id,`+
						` projections.domain_policies.sequence,`+
						` projections.domain_policies.creation_date,`+
						` projections.domain_policies.change_date,`+
						` projections.domain_policies.resource_owner,`+
						` projections.domain_policies.user_login_must_be_domain,`+
						` projections.domain_policies.validate_org_domains,`+
						` projections.domain_policies.smtp_sender_address_matches_instance_domain,`+
						` projections.domain_policies.is_default,`+
						` projections.domain_policies.state`+
						` FROM projections.domain_policies`),
					sql.ErrConnDone,
				),
				err: func(err error) (error, bool) {
					if !errors.Is(err, sql.ErrConnDone) {
						return fmt.Errorf("err should be sql.ErrConnDone got: %w", err), false
					}
					return nil, true
				},
			},
			object: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPrepare(t, tt.prepare, tt.object, tt.want.sqlExpectations, tt.want.err)
		})
	}
}
