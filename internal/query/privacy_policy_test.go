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

func Test_PrivacyPolicyPrepares(t *testing.T) {
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
			name:    "preparePrivacyPolicyQuery no result",
			prepare: preparePrivacyPolicyQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.privacy_policies.id,`+
						` projections.privacy_policies.sequence,`+
						` projections.privacy_policies.creation_date,`+
						` projections.privacy_policies.change_date,`+
						` projections.privacy_policies.resource_owner,`+
						` projections.privacy_policies.privacy_link,`+
						` projections.privacy_policies.tos_link,`+
						` projections.privacy_policies.help_link,`+
						` projections.privacy_policies.is_default,`+
						` projections.privacy_policies.state`+
						` FROM projections.privacy_policies`),
					nil,
					nil,
				),
				err: func(err error) (error, bool) {
					if !errs.IsNotFound(err) {
						return fmt.Errorf("err should be NotFoundError got: %w", err), false
					}
					return nil, true
				},
			},
			object: (*PrivacyPolicy)(nil),
		},
		{
			name:    "preparePrivacyPolicyQuery found",
			prepare: preparePrivacyPolicyQuery,
			want: want{
				sqlExpectations: mockQuery(
					regexp.QuoteMeta(`SELECT projections.privacy_policies.id,`+
						` projections.privacy_policies.sequence,`+
						` projections.privacy_policies.creation_date,`+
						` projections.privacy_policies.change_date,`+
						` projections.privacy_policies.resource_owner,`+
						` projections.privacy_policies.privacy_link,`+
						` projections.privacy_policies.tos_link,`+
						` projections.privacy_policies.help_link,`+
						` projections.privacy_policies.is_default,`+
						` projections.privacy_policies.state`+
						` FROM projections.privacy_policies`),
					[]string{
						"id",
						"sequence",
						"creation_date",
						"change_date",
						"resource_owner",
						"privacy_link",
						"tos_link",
						"help_link",
						"is_default",
						"state",
					},
					[]driver.Value{
						"pol-id",
						uint64(20211109),
						testNow,
						testNow,
						"ro",
						"privacy.ch",
						"tos.ch",
						"help.ch",
						true,
						domain.PolicyStateActive,
					},
				),
			},
			object: &PrivacyPolicy{
				ID:            "pol-id",
				CreationDate:  testNow,
				ChangeDate:    testNow,
				Sequence:      20211109,
				ResourceOwner: "ro",
				State:         domain.PolicyStateActive,
				PrivacyLink:   "privacy.ch",
				TOSLink:       "tos.ch",
				HelpLink:      "help.ch",
				IsDefault:     true,
			},
		},
		{
			name:    "preparePrivacyPolicyQuery sql err",
			prepare: preparePrivacyPolicyQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.privacy_policies.id,`+
						` projections.privacy_policies.sequence,`+
						` projections.privacy_policies.creation_date,`+
						` projections.privacy_policies.change_date,`+
						` projections.privacy_policies.resource_owner,`+
						` projections.privacy_policies.privacy_link,`+
						` projections.privacy_policies.tos_link,`+
						` projections.privacy_policies.help_link,`+
						` projections.privacy_policies.is_default,`+
						` projections.privacy_policies.state`+
						` FROM projections.privacy_policies`),
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
