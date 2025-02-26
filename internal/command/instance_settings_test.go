package command

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/dennigogo/zitadel/internal/api/authz"

	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/instance"
)

func TestCommandSide_AddSecretGenerator(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx           context.Context
		generator     *crypto.GeneratorConfig
		generatorType domain.SecretGeneratorType
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "invalid empty type, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:           context.Background(),
				generator:     &crypto.GeneratorConfig{},
				generatorType: domain.SecretGeneratorTypeUnspecified,
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "secret generator config, error already exists",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSecretGeneratorAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode,
								4,
								time.Hour*1,
								true,
								true,
								true,
								true,
							),
						),
					),
				),
			},
			args: args{
				ctx: context.Background(),
				generator: &crypto.GeneratorConfig{
					Length:              4,
					Expiry:              1 * time.Hour,
					IncludeLowerLetters: true,
					IncludeUpperLetters: true,
					IncludeDigits:       true,
					IncludeSymbols:      true,
				},
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				err: caos_errs.IsErrorAlreadyExists,
			},
		},
		{
			name: "add secret generator, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusherWithInstanceID(
								"INSTANCE",
								instance.NewSecretGeneratorAddedEvent(
									context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									domain.SecretGeneratorTypeInitCode,
									4,
									time.Hour*1,
									true,
									true,
									true,
									true,
								),
							),
						},
						uniqueConstraintsFromEventConstraintWithInstanceID("INSTANCE", instance.NewAddSecretGeneratorTypeUniqueConstraint(domain.SecretGeneratorTypeInitCode)),
					),
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				generator: &crypto.GeneratorConfig{
					Length:              4,
					Expiry:              1 * time.Hour,
					IncludeLowerLetters: true,
					IncludeUpperLetters: true,
					IncludeDigits:       true,
					IncludeSymbols:      true,
				},
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.AddSecretGeneratorConfig(tt.args.ctx, tt.args.generatorType, tt.args.generator)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_ChangeSecretGenerator(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx           context.Context
		generator     *crypto.GeneratorConfig
		generatorType domain.SecretGeneratorType
		instanceID    string
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "empty generatortype, invalid error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:           context.Background(),
				generator:     &crypto.GeneratorConfig{},
				generatorType: domain.SecretGeneratorTypeUnspecified,
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "generator not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:           context.Background(),
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "generator removed, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSecretGeneratorAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode,
								4,
								time.Hour*1,
								true,
								true,
								true,
								true,
							),
						),
						eventFromEventPusher(
							instance.NewSecretGeneratorRemovedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode),
						),
					),
				),
			},
			args: args{
				ctx:           context.Background(),
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "no changes, precondition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSecretGeneratorAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode,
								4,
								time.Hour*1,
								true,
								true,
								true,
								true,
							),
						),
					),
				),
			},
			args: args{
				ctx: context.Background(),
				generator: &crypto.GeneratorConfig{
					Length:              4,
					Expiry:              1 * time.Hour,
					IncludeLowerLetters: true,
					IncludeUpperLetters: true,
					IncludeDigits:       true,
					IncludeSymbols:      true,
				},
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "secret generator change, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSecretGeneratorAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode,
								4,
								time.Hour*1,
								true,
								true,
								true,
								true,
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								newSecretGeneratorChangedEvent(context.Background(),
									domain.SecretGeneratorTypeInitCode,
									8,
									time.Hour*2,
									false,
									false,
									false,
									false),
							),
						},
					),
				),
			},
			args: args{
				ctx: context.Background(),
				generator: &crypto.GeneratorConfig{
					Length:              8,
					Expiry:              2 * time.Hour,
					IncludeLowerLetters: false,
					IncludeUpperLetters: false,
					IncludeDigits:       false,
					IncludeSymbols:      false,
				},
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.ChangeSecretGeneratorConfig(tt.args.ctx, tt.args.generatorType, tt.args.generator)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_RemoveSecretGenerator(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx           context.Context
		generatorType domain.SecretGeneratorType
		instanceID    string
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "empty type, invalid error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:           context.Background(),
				generatorType: domain.SecretGeneratorTypeUnspecified,
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "generator not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:           context.Background(),
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "generator removed, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSecretGeneratorAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode,
								4,
								time.Hour*1,
								true,
								true,
								true,
								true,
							),
						),
						eventFromEventPusher(
							instance.NewSecretGeneratorRemovedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode),
						),
					),
				),
			},
			args: args{
				ctx:           context.Background(),
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "generator config remove, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSecretGeneratorAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								domain.SecretGeneratorTypeInitCode,
								4,
								time.Hour*1,
								true,
								true,
								true,
								true,
							),
						),
					),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								instance.NewSecretGeneratorRemovedEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									domain.SecretGeneratorTypeInitCode),
							),
						},
						uniqueConstraintsFromEventConstraint(instance.NewRemoveSecretGeneratorTypeUniqueConstraint(domain.SecretGeneratorTypeInitCode)),
					),
				),
			},
			args: args{
				ctx:           context.Background(),
				generatorType: domain.SecretGeneratorTypeInitCode,
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.RemoveSecretGeneratorConfig(tt.args.ctx, tt.args.generatorType)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func newSecretGeneratorChangedEvent(ctx context.Context, generatorType domain.SecretGeneratorType, length uint, expiry time.Duration, lowerCase, upperCase, digits, symbols bool) *instance.SecretGeneratorChangedEvent {
	changes := []instance.SecretGeneratorChanges{
		instance.ChangeSecretGeneratorLength(length),
		instance.ChangeSecretGeneratorExpiry(expiry),
		instance.ChangeSecretGeneratorIncludeLowerLetters(lowerCase),
		instance.ChangeSecretGeneratorIncludeUpperLetters(upperCase),
		instance.ChangeSecretGeneratorIncludeDigits(digits),
		instance.ChangeSecretGeneratorIncludeSymbols(symbols),
	}
	event, _ := instance.NewSecretGeneratorChangeEvent(ctx,
		&instance.NewAggregate("INSTANCE").Aggregate,
		generatorType,
		changes,
	)
	return event
}
