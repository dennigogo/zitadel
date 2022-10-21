package query

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/dennigogo/zitadel/internal/domain"
)

var (
	loginPolicyIDPLinksQuery = regexp.QuoteMeta(`SELECT projections.idp_login_policy_links3.idp_id,` +
		` projections.idps2.name,` +
		` projections.idps2.type,` +
		` COUNT(*) OVER ()` +
		` FROM projections.idp_login_policy_links3` +
		` LEFT JOIN projections.idps2 ON projections.idp_login_policy_links3.idp_id = projections.idps2.id`)
	loginPolicyIDPLinksCols = []string{
		"idp_id",
		"name",
		"type",
		"count",
	}
)

func Test_IDPLoginPolicyLinkPrepares(t *testing.T) {
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
			name:    "prepareIDPsQuery found",
			prepare: prepareIDPLoginPolicyLinksQuery,
			want: want{
				sqlExpectations: mockQueries(
					loginPolicyIDPLinksQuery,
					loginPolicyIDPLinksCols,
					[][]driver.Value{
						{
							"idp-id",
							"idp-name",
							domain.IDPConfigTypeJWT,
						},
					},
				),
			},
			object: &IDPLoginPolicyLinks{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				Links: []*IDPLoginPolicyLink{
					{
						IDPID:   "idp-id",
						IDPName: "idp-name",
						IDPType: domain.IDPConfigTypeJWT,
					},
				},
			},
		},
		{
			name:    "prepareIDPsQuery no idp",
			prepare: prepareIDPLoginPolicyLinksQuery,
			want: want{
				sqlExpectations: mockQueries(
					loginPolicyIDPLinksQuery,
					loginPolicyIDPLinksCols,
					[][]driver.Value{
						{
							"idp-id",
							nil,
							nil,
						},
					},
				),
			},
			object: &IDPLoginPolicyLinks{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				Links: []*IDPLoginPolicyLink{
					{
						IDPID:   "idp-id",
						IDPName: "",
						IDPType: domain.IDPConfigTypeUnspecified,
					},
				},
			},
		},
		{
			name:    "prepareIDPsQuery sql err",
			prepare: prepareIDPLoginPolicyLinksQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					loginPolicyIDPLinksQuery,
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
