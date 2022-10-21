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

func Test_AuthNKeyPrepares(t *testing.T) {
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
			name:    "prepareAuthNKeysQuery no result",
			prepare: prepareAuthNKeysQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
					nil,
					nil,
				),
			},
			object: &AuthNKeys{AuthNKeys: []*AuthNKey{}},
		},
		{
			name:    "prepareAuthNKeysQuery one result",
			prepare: prepareAuthNKeysQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
					[]string{
						"id",
						"creation_date",
						"resource_owner",
						"sequence",
						"expiration",
						"type",
						"count",
					},
					[][]driver.Value{
						{
							"id",
							testNow,
							"ro",
							uint64(20211109),
							testNow,
							1,
						},
					},
				),
			},
			object: &AuthNKeys{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				AuthNKeys: []*AuthNKey{
					{
						ID:            "id",
						CreationDate:  testNow,
						ResourceOwner: "ro",
						Sequence:      20211109,
						Expiration:    testNow,
						Type:          domain.AuthNKeyTypeJSON,
					},
				},
			},
		},
		{
			name:    "prepareAuthNKeysQuery multiple result",
			prepare: prepareAuthNKeysQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
					[]string{
						"id",
						"creation_date",
						"resource_owner",
						"sequence",
						"expiration",
						"type",
						"count",
					},
					[][]driver.Value{
						{
							"id-1",
							testNow,
							"ro",
							uint64(20211109),
							testNow,
							1,
						},
						{
							"id-2",
							testNow,
							"ro",
							uint64(20211109),
							testNow,
							1,
						},
					},
				),
			},
			object: &AuthNKeys{
				SearchResponse: SearchResponse{
					Count: 2,
				},
				AuthNKeys: []*AuthNKey{
					{
						ID:            "id-1",
						CreationDate:  testNow,
						ResourceOwner: "ro",
						Sequence:      20211109,
						Expiration:    testNow,
						Type:          domain.AuthNKeyTypeJSON,
					},
					{
						ID:            "id-2",
						CreationDate:  testNow,
						ResourceOwner: "ro",
						Sequence:      20211109,
						Expiration:    testNow,
						Type:          domain.AuthNKeyTypeJSON,
					},
				},
			},
		},
		{
			name:    "prepareAuthNKeysQuery sql err",
			prepare: prepareAuthNKeysQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
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
		{
			name:    "prepareAuthNKeysDataQuery no result",
			prepare: prepareAuthNKeysDataQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` projections.authn_keys.identifier,`+
						` projections.authn_keys.public_key,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
					nil,
					nil,
				),
			},
			object: &AuthNKeysData{AuthNKeysData: []*AuthNKeyData{}},
		},
		{
			name:    "prepareAuthNKeysDataQuery one result",
			prepare: prepareAuthNKeysDataQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` projections.authn_keys.identifier,`+
						` projections.authn_keys.public_key,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
					[]string{
						"id",
						"creation_date",
						"resource_owner",
						"sequence",
						"expiration",
						"type",
						"identifier",
						"public_key",
						"count",
					},
					[][]driver.Value{
						{
							"id",
							testNow,
							"ro",
							uint64(20211109),
							testNow,
							1,
							"identifier",
							[]byte("public"),
						},
					},
				),
			},
			object: &AuthNKeysData{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				AuthNKeysData: []*AuthNKeyData{
					{
						ID:            "id",
						CreationDate:  testNow,
						ResourceOwner: "ro",
						Sequence:      20211109,
						Expiration:    testNow,
						Type:          domain.AuthNKeyTypeJSON,
						Identifier:    "identifier",
						PublicKey:     []byte("public"),
					},
				},
			},
		},
		{
			name:    "prepareAuthNKeysDataQuery multiple result",
			prepare: prepareAuthNKeysDataQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` projections.authn_keys.identifier,`+
						` projections.authn_keys.public_key,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
					[]string{
						"id",
						"creation_date",
						"resource_owner",
						"sequence",
						"expiration",
						"type",
						"identifier",
						"public_key",
						"count",
					},
					[][]driver.Value{
						{
							"id-1",
							testNow,
							"ro",
							uint64(20211109),
							testNow,
							1,
							"identifier1",
							[]byte("public1"),
						},
						{
							"id-2",
							testNow,
							"ro",
							uint64(20211109),
							testNow,
							1,
							"identifier2",
							[]byte("public2"),
						},
					},
				),
			},
			object: &AuthNKeysData{
				SearchResponse: SearchResponse{
					Count: 2,
				},
				AuthNKeysData: []*AuthNKeyData{
					{
						ID:            "id-1",
						CreationDate:  testNow,
						ResourceOwner: "ro",
						Sequence:      20211109,
						Expiration:    testNow,
						Type:          domain.AuthNKeyTypeJSON,
						Identifier:    "identifier1",
						PublicKey:     []byte("public1"),
					},
					{
						ID:            "id-2",
						CreationDate:  testNow,
						ResourceOwner: "ro",
						Sequence:      20211109,
						Expiration:    testNow,
						Type:          domain.AuthNKeyTypeJSON,
						Identifier:    "identifier2",
						PublicKey:     []byte("public2"),
					},
				},
			},
		},
		{
			name:    "prepareAuthNKeysDataQuery sql err",
			prepare: prepareAuthNKeysDataQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type,`+
						` projections.authn_keys.identifier,`+
						` projections.authn_keys.public_key,`+
						` COUNT(*) OVER ()`+
						` FROM projections.authn_keys`),
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
		{
			name:    "prepareAuthNKeyQuery no result",
			prepare: prepareAuthNKeyQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type`+
						` FROM projections.authn_keys`),
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
			object: (*AuthNKey)(nil),
		},
		{
			name:    "prepareAuthNKeyQuery found",
			prepare: prepareAuthNKeyQuery,
			want: want{
				sqlExpectations: mockQuery(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type`+
						` FROM projections.authn_keys`),
					[]string{
						"id",
						"creation_date",
						"resource_owner",
						"sequence",
						"expiration",
						"type",
					},
					[]driver.Value{
						"id",
						testNow,
						"ro",
						uint64(20211109),
						testNow,
						1,
					},
				),
			},
			object: &AuthNKey{
				ID:            "id",
				CreationDate:  testNow,
				ResourceOwner: "ro",
				Sequence:      20211109,
				Expiration:    testNow,
				Type:          domain.AuthNKeyTypeJSON,
			},
		},
		{
			name:    "prepareAuthNKeyQuery sql err",
			prepare: prepareAuthNKeyQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.authn_keys.id,`+
						` projections.authn_keys.creation_date,`+
						` projections.authn_keys.resource_owner,`+
						` projections.authn_keys.sequence,`+
						` projections.authn_keys.expiration,`+
						` projections.authn_keys.type`+
						` FROM projections.authn_keys`),
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
		{
			name:    "prepareAuthNKeyPublicKeyQuery no result",
			prepare: prepareAuthNKeyPublicKeyQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.authn_keys.public_key`+
						` FROM projections.authn_keys`),
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
			object: ([]byte)(nil),
		},
		{
			name:    "prepareAuthNKeyPublicKeyQuery found",
			prepare: prepareAuthNKeyPublicKeyQuery,
			want: want{
				sqlExpectations: mockQuery(
					regexp.QuoteMeta(`SELECT projections.authn_keys.public_key`+
						` FROM projections.authn_keys`),
					[]string{
						"public_key",
					},
					[]driver.Value{
						[]byte("publicKey"),
					},
				),
			},
			object: []byte("publicKey"),
		},
		{
			name:    "prepareAuthNKeyPublicKeyQuery sql err",
			prepare: prepareAuthNKeyPublicKeyQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.authn_keys.public_key`+
						` FROM projections.authn_keys`),
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
