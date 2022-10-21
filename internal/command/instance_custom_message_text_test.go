package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"github.com/dennigogo/zitadel/internal/domain"
	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
	"github.com/dennigogo/zitadel/internal/repository/instance"
)

func TestCommandSide_SetDefaultMessageText(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx        context.Context
		instanceID string
		config     *domain.CustomMessageText
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
			name: "invalid custom text, error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:        context.Background(),
				instanceID: "INSTANCE",
				config:     &domain.CustomMessageText{},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "custom text set all fields, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
					expectPush(
						[]*repository.Event{
							eventFromEventPusher(
								instance.NewCustomTextSetEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"Template",
									domain.MessageGreeting,
									"Greeting",
									language.English,
								),
							),
							eventFromEventPusher(
								instance.NewCustomTextSetEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"Template",
									domain.MessageSubject,
									"Subject",
									language.English,
								),
							),
							eventFromEventPusher(
								instance.NewCustomTextSetEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"Template",
									domain.MessageTitle,
									"Title",
									language.English,
								),
							),
							eventFromEventPusher(
								instance.NewCustomTextSetEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"Template",
									domain.MessagePreHeader,
									"PreHeader",
									language.English,
								),
							),
							eventFromEventPusher(
								instance.NewCustomTextSetEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"Template",
									domain.MessageText,
									"Text",
									language.English,
								),
							),
							eventFromEventPusher(
								instance.NewCustomTextSetEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"Template",
									domain.MessageButtonText,
									"ButtonText",
									language.English,
								),
							),
							eventFromEventPusher(
								instance.NewCustomTextSetEvent(context.Background(),
									&instance.NewAggregate("INSTANCE").Aggregate,
									"Template",
									domain.MessageFooterText,
									"Footer",
									language.English,
								),
							),
						},
					),
				),
			},
			args: args{
				ctx:        context.Background(),
				instanceID: "INSTANCE",
				config: &domain.CustomMessageText{
					MessageTextType: "Template",
					Language:        language.English,
					Greeting:        "Greeting",
					Subject:         "Subject",
					Title:           "Title",
					PreHeader:       "PreHeader",
					Text:            "Text",
					ButtonText:      "ButtonText",
					FooterText:      "Footer",
				},
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
			got, err := r.SetDefaultMessageText(tt.args.ctx, tt.args.instanceID, tt.args.config)
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
