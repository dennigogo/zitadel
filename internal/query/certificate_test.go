package query

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
	errs "github.com/dennigogo/zitadel/internal/errors"
)

func Test_CertificatePrepares(t *testing.T) {
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
			name:    "prepareCertificateQuery no result",
			prepare: prepareCertificateQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.keys3.id,`+
						` projections.keys3.creation_date,`+
						` projections.keys3.change_date,`+
						` projections.keys3.sequence,`+
						` projections.keys3.resource_owner,`+
						` projections.keys3.algorithm,`+
						` projections.keys3.use,`+
						` projections.keys3_certificate.expiry,`+
						` projections.keys3_certificate.certificate,`+
						` projections.keys3_private.key,`+
						` COUNT(*) OVER ()`+
						` FROM projections.keys3`+
						` LEFT JOIN projections.keys3_certificate ON projections.keys3.id = projections.keys3_certificate.id`+
						` LEFT JOIN projections.keys3_private ON projections.keys3.id = projections.keys3_private.id`),
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
			object: &Certificates{Certificates: []Certificate{}},
		},
		{
			name:    "prepareCertificateQuery found",
			prepare: prepareCertificateQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(`SELECT projections.keys3.id,`+
						` projections.keys3.creation_date,`+
						` projections.keys3.change_date,`+
						` projections.keys3.sequence,`+
						` projections.keys3.resource_owner,`+
						` projections.keys3.algorithm,`+
						` projections.keys3.use,`+
						` projections.keys3_certificate.expiry,`+
						` projections.keys3_certificate.certificate,`+
						` projections.keys3_private.key,`+
						` COUNT(*) OVER ()`+
						` FROM projections.keys3`+
						` LEFT JOIN projections.keys3_certificate ON projections.keys3.id = projections.keys3_certificate.id`+
						` LEFT JOIN projections.keys3_private ON projections.keys3.id = projections.keys3_private.id`),
					[]string{
						"id",
						"creation_date",
						"change_date",
						"sequence",
						"resource_owner",
						"algorithm",
						"use",
						"expiry",
						"certificate",
						"key",
						"count",
					},
					[][]driver.Value{
						{
							"key-id",
							testNow,
							testNow,
							uint64(20211109),
							"ro",
							"",
							1,
							testNow,
							[]byte(`privateKey`),
							[]byte(`{"Algorithm": "enc", "Crypted": "cHJpdmF0ZUtleQ==", "CryptoType": 0, "KeyID": "id"}`),
						},
					},
				),
			},
			object: &Certificates{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				Certificates: []Certificate{
					&rsaCertificate{
						key: key{
							id:            "key-id",
							creationDate:  testNow,
							changeDate:    testNow,
							sequence:      20211109,
							resourceOwner: "ro",
							algorithm:     "",
							use:           domain.KeyUsageSAMLMetadataSigning,
						},
						expiry:      testNow,
						certificate: []byte("privateKey"),
						privateKey: &crypto.CryptoValue{
							CryptoType: crypto.TypeEncryption,
							Algorithm:  "enc",
							KeyID:      "id",
							Crypted:    []byte("privateKey"),
						},
					},
				},
			},
		},
		{
			name:    "prepareCertificateQuery sql err",
			prepare: prepareCertificateQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(`SELECT projections.keys3.id,`+
						` projections.keys3.creation_date,`+
						` projections.keys3.change_date,`+
						` projections.keys3.sequence,`+
						` projections.keys3.resource_owner,`+
						` projections.keys3.algorithm,`+
						` projections.keys3.use,`+
						` projections.keys3_certificate.expiry,`+
						` projections.keys3_certificate.certificate,`+
						` projections.keys3_private.key,`+
						` COUNT(*) OVER ()`+
						` FROM projections.keys3`+
						` LEFT JOIN projections.keys3_certificate ON projections.keys3.id = projections.keys3_certificate.id`+
						` LEFT JOIN projections.keys3_private ON projections.keys3.id = projections.keys3_private.id`),
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
