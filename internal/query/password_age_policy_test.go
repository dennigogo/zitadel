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

func Test_PasswordAgePolicyPrepares(t *testing.T) {
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
			name:    "preparePasswordAgePolicyQuery no result",
			prepare: preparePasswordAgePolicyQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.password_age_policies.id,`+
						` projections.password_age_policies.sequence,`+
						` projections.password_age_policies.creation_date,`+
						` projections.password_age_policies.change_date,`+
						` projections.password_age_policies.resource_owner,`+
						` projections.password_age_policies.expire_warn_days,`+
						` projections.password_age_policies.max_age_days,`+
						` projections.password_age_policies.is_default,`+
						` projections.password_age_policies.state`+
						` FROM projections.password_age_policies`),
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
			object: (*PasswordAgePolicy)(nil),
		},
		{
			name:    "preparePasswordAgePolicyQuery found",
			prepare: preparePasswordAgePolicyQuery,
			want: want{
				sqlExpectations: mockQuery(
					regexp.QuoteMeta(`SELECT projections.password_age_policies.id,`+
						` projections.password_age_policies.sequence,`+
						` projections.password_age_policies.creation_date,`+
						` projections.password_age_policies.change_date,`+
						` projections.password_age_policies.resource_owner,`+
						` projections.password_age_policies.expire_warn_days,`+
						` projections.password_age_policies.max_age_days,`+
						` projections.password_age_policies.is_default,`+
						` projections.password_age_policies.state`+
						` FROM projections.password_age_policies`),
					[]string{
						"id",
						"sequence",
						"creation_date",
						"change_date",
						"resource_owner",
						"expire_warn_days",
						"max_age_days",
						"is_default",
						"state",
					},
					[]driver.Value{
						"pol-id",
						uint64(20211109),
						testNow,
						testNow,
						"ro",
						10,
						20,
						true,
						domain.PolicyStateActive,
					},
				),
			},
			object: &PasswordAgePolicy{
				ID:             "pol-id",
				CreationDate:   testNow,
				ChangeDate:     testNow,
				Sequence:       20211109,
				ResourceOwner:  "ro",
				State:          domain.PolicyStateActive,
				ExpireWarnDays: 10,
				MaxAgeDays:     20,
				IsDefault:      true,
			},
		},
		{
			name:    "preparePasswordAgePolicyQuery sql err",
			prepare: preparePasswordAgePolicyQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.password_age_policies.id,`+
						` projections.password_age_policies.sequence,`+
						` projections.password_age_policies.creation_date,`+
						` projections.password_age_policies.change_date,`+
						` projections.password_age_policies.resource_owner,`+
						` projections.password_age_policies.expire_warn_days,`+
						` projections.password_age_policies.max_age_days,`+
						` projections.password_age_policies.is_default,`+
						` projections.password_age_policies.state`+
						` FROM projections.password_age_policies`),
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
