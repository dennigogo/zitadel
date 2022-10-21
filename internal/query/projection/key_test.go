package projection

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/handler"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/keypair"
)

func TestKeyProjection_reduces(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "reduceKeyPairAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(keypair.AddedEventType),
					keypair.AggregateType,
					keypairAddedEventData(domain.KeyUsageSigning, time.Now().Add(time.Hour)),
				), keypair.AddedEventMapper),
			},
			reduce: (&keyProjection{encryptionAlgorithm: crypto.CreateMockEncryptionAlg(gomock.NewController(t))}).reduceKeyPairAdded,
			want: wantReduce{
				projection:       KeyProjectionTable,
				aggregateType:    eventstore.AggregateType("key_pair"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.keys3 (id, creation_date, change_date, resource_owner, instance_id, sequence, algorithm, use) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
							expectedArgs: []interface{}{
								"agg-id",
								anyArg{},
								anyArg{},
								"ro-id",
								"instance-id",
								uint64(15),
								"algorithm",
								domain.KeyUsageSigning,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.keys3_private (id, instance_id, expiry, key) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								"agg-id",
								"instance-id",
								anyArg{},
								&crypto.CryptoValue{
									CryptoType: crypto.TypeEncryption,
									Algorithm:  "enc",
									KeyID:      "id",
									Crypted:    []byte("privateKey"),
								},
							},
						},
						{
							expectedStmt: "INSERT INTO projections.keys3_public (id, instance_id, expiry, key) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								"agg-id",
								"instance-id",
								anyArg{},
								[]byte("publicKey"),
							},
						},
					},
				},
			},
		},
		{
			name: "reduceKeyPairAdded expired",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(keypair.AddedEventType),
					keypair.AggregateType,
					keypairAddedEventData(domain.KeyUsageSigning, time.Now().Add(-time.Hour)),
				), keypair.AddedEventMapper),
			},
			reduce: (&keyProjection{}).reduceKeyPairAdded,
			want: wantReduce{
				projection:       KeyProjectionTable,
				aggregateType:    eventstore.AggregateType("key_pair"),
				sequence:         15,
				previousSequence: 10,
				executer:         &testExecuter{},
			},
		},
		{
			name: "reduceCertificateAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(keypair.AddedCertificateEventType),
					keypair.AggregateType,
					certificateAddedEventData(domain.KeyUsageSAMLMetadataSigning, time.Now().Add(time.Hour)),
				), keypair.AddedCertificateEventMapper),
			},
			reduce: (&keyProjection{certEncryptionAlgorithm: crypto.CreateMockEncryptionAlg(gomock.NewController(t))}).reduceCertificateAdded,
			want: wantReduce{
				projection:       KeyProjectionTable,
				aggregateType:    eventstore.AggregateType("key_pair"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.keys3_certificate (id, instance_id, expiry, certificate) VALUES ($1, $2, $3, $4)",
							expectedArgs: []interface{}{
								"agg-id",
								"instance-id",
								anyArg{},
								[]byte("privateKey"),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if !errors.IsErrorInvalidArgument(err) {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, tt.want)
		})
	}
}

func keypairAddedEventData(usage domain.KeyUsage, t time.Time) []byte {
	return []byte(`{"algorithm": "algorithm", "usage": ` + fmt.Sprintf("%d", usage) + `, "privateKey": {"key": {"cryptoType": 0, "algorithm": "enc", "keyID": "id", "crypted": "cHJpdmF0ZUtleQ=="}, "expiry": "` + t.Format(time.RFC3339) + `"}, "publicKey": {"key": {"cryptoType": 0, "algorithm": "enc", "keyID": "id", "crypted": "cHVibGljS2V5"}, "expiry": "` + t.Format(time.RFC3339) + `"}}`)
}

func certificateAddedEventData(usage domain.KeyUsage, t time.Time) []byte {
	return []byte(`{"algorithm": "algorithm", "usage": ` + fmt.Sprintf("%d", usage) + `, "certificate": {"key": {"cryptoType": 0, "algorithm": "enc", "keyID": "id", "crypted": "cHJpdmF0ZUtleQ=="}, "expiry": "` + t.Format(time.RFC3339) + `"}}`)
}
