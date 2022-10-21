package eventstore

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/dennigogo/zitadel/internal/api/authz"
	"github.com/dennigogo/zitadel/internal/api/service"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
)

// testEvent implements the Event interface
type testEvent struct {
	BaseEvent

	description         string
	shouldCheckPrevious bool
	data                func() interface{}
}

func newTestEvent(id, description string, data func() interface{}, checkPrevious bool) *testEvent {
	return &testEvent{
		description:         description,
		data:                data,
		shouldCheckPrevious: checkPrevious,
		BaseEvent: *NewBaseEventForPush(
			service.WithService(authz.NewMockContext("instanceID", "resourceOwner", "editorUser"), "editorService"),
			NewAggregate(authz.NewMockContext("zitadel", "caos", "adlerhurst"), id, "test.aggregate", "v1"),
			"test.event",
		),
	}
}

func (e *testEvent) Data() interface{} {
	return e.data()
}

func (e *testEvent) UniqueConstraints() []*EventUniqueConstraint {
	return nil
}

func (e *testEvent) Assets() []*Asset {
	return nil
}

func testFilterMapper(event *repository.Event) (Event, error) {
	if event == nil {
		return newTestEvent("testID", "hodor", nil, false), nil
	}
	return &testEvent{description: "hodor", BaseEvent: *BaseEventFromRepo(event)}, nil
}

func Test_eventstore_RegisterFilterEventMapper(t *testing.T) {
	type fields struct {
		eventMapper map[EventType]eventTypeInterceptors
	}
	type args struct {
		eventType EventType
		mapper    func(*repository.Event) (Event, error)
	}
	type res struct {
		event       Event
		mapperCount int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "no event type",
			args: args{
				eventType: "",
				mapper:    testFilterMapper,
			},
			fields: fields{
				eventMapper: map[EventType]eventTypeInterceptors{},
			},
			res: res{
				mapperCount: 0,
			},
		},
		{
			name: "no event mapper",
			args: args{
				eventType: "event.type",
				mapper:    nil,
			},
			fields: fields{
				eventMapper: map[EventType]eventTypeInterceptors{},
			},
			res: res{
				mapperCount: 0,
			},
		},
		{
			name: "new interceptor",
			fields: fields{
				eventMapper: map[EventType]eventTypeInterceptors{},
			},
			args: args{
				eventType: "event.type",
				mapper:    testFilterMapper,
			},
			res: res{
				event:       newTestEvent("testID", "hodor", nil, false),
				mapperCount: 1,
			},
		},
		{
			name: "existing interceptor new filter mapper",
			fields: fields{
				eventMapper: map[EventType]eventTypeInterceptors{
					"event.type": {},
				},
			},
			args: args{
				eventType: "new.event",
				mapper:    testFilterMapper,
			},
			res: res{
				event:       newTestEvent("testID", "hodor", nil, false),
				mapperCount: 2,
			},
		},
		{
			name: "existing interceptor existing filter mapper",
			fields: fields{
				eventMapper: map[EventType]eventTypeInterceptors{
					"event.type": {
						eventMapper: func(*repository.Event) (Event, error) {
							return nil, errors.ThrowUnimplemented(nil, "V2-1qPvn", "unimplemented")
						},
					},
				},
			},
			args: args{
				eventType: "new.event",
				mapper:    testFilterMapper,
			},
			res: res{
				event:       newTestEvent("testID", "hodor", nil, false),
				mapperCount: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Eventstore{
				eventInterceptors: tt.fields.eventMapper,
			}
			es = es.RegisterFilterEventMapper(tt.args.eventType, tt.args.mapper)
			if len(es.eventInterceptors) != tt.res.mapperCount {
				t.Errorf("unexpected mapper count: want %d, got %d", tt.res.mapperCount, len(es.eventInterceptors))
				return
			}

			if tt.res.mapperCount == 0 {
				return
			}

			mapper := es.eventInterceptors[tt.args.eventType]
			event, err := mapper.eventMapper(nil)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if !reflect.DeepEqual(tt.res.event, event) {
				t.Errorf("events should be deep equal. \ngot %#v\nwant %#v", event, tt.res.event)
			}
		})
	}
}

func Test_eventData(t *testing.T) {
	type args struct {
		event Command
	}
	type res struct {
		jsonText []byte
		wantErr  bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "data as json bytes",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						return []byte(`{"piff":"paff"}`)
					},
					false),
			},
			res: res{
				jsonText: []byte(`{"piff":"paff"}`),
				wantErr:  false,
			},
		},
		{
			name: "data as invalid json bytes",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						return []byte(`{"piffpaff"}`)
					},
					false),
			},
			res: res{
				jsonText: []byte(nil),
				wantErr:  true,
			},
		},
		{
			name: "data as struct",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						return struct {
							Piff string `json:"piff"`
						}{Piff: "paff"}
					},
					false),
			},
			res: res{
				jsonText: []byte(`{"piff":"paff"}`),
				wantErr:  false,
			},
		},
		{
			name: "data as ptr to struct",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						return &struct {
							Piff string `json:"piff"`
						}{Piff: "paff"}
					},
					false),
			},
			res: res{
				jsonText: []byte(`{"piff":"paff"}`),
				wantErr:  false,
			},
		},
		{
			name: "no data",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						return nil
					},
					false),
			},
			res: res{
				jsonText: []byte(nil),
				wantErr:  false,
			},
		},
		{
			name: "invalid because primitive",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						return ""
					},
					false),
			},
			res: res{
				jsonText: []byte(nil),
				wantErr:  true,
			},
		},
		{
			name: "invalid because pointer to primitive",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						var s string
						return &s
					},
					false),
			},
			res: res{
				jsonText: []byte(nil),
				wantErr:  true,
			},
		},
		{
			name: "invalid because invalid struct for json",
			args: args{
				event: newTestEvent(
					"id",
					"hodor",
					func() interface{} {
						return struct {
							Field chan string `json:"field"`
						}{}
					},
					false),
			},
			res: res{
				jsonText: []byte(nil),
				wantErr:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EventData(tt.args.event)
			if (err != nil) != tt.res.wantErr {
				t.Errorf("EventData() error = %v, wantErr %v", err, tt.res.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.res.jsonText) {
				t.Errorf("EventData() = %v, want %v", string(got), string(tt.res.jsonText))
			}
		})
	}
}

func TestEventstore_aggregatesToEvents(t *testing.T) {
	type args struct {
		instanceID string
		events     []Command
	}
	type res struct {
		wantErr bool
		events  []*repository.Event
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "one aggregate one event",
			args: args{
				instanceID: "instanceID",
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
				},
			},
			res: res{
				wantErr: false,
				events: []*repository.Event{
					{
						AggregateID:   "1",
						AggregateType: "test.aggregate",
						Data:          []byte(nil),
						EditorService: "editorService",
						EditorUser:    "editorUser",
						ResourceOwner: sql.NullString{String: "caos", Valid: true},
						InstanceID:    "instanceID",
						Type:          "test.event",
						Version:       "v1",
					},
				},
			},
		},
		{
			name: "one aggregate multiple events",
			args: args{
				instanceID: "instanceID",
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
				},
			},
			res: res{
				wantErr: false,
				events: []*repository.Event{
					{
						AggregateID:   "1",
						AggregateType: "test.aggregate",
						Data:          []byte(nil),
						EditorService: "editorService",
						EditorUser:    "editorUser",
						ResourceOwner: sql.NullString{String: "caos", Valid: true},
						InstanceID:    "instanceID",
						Type:          "test.event",
						Version:       "v1",
					},
					{
						AggregateID:   "1",
						AggregateType: "test.aggregate",
						Data:          []byte(nil),
						EditorService: "editorService",
						EditorUser:    "editorUser",
						ResourceOwner: sql.NullString{String: "caos", Valid: true},
						InstanceID:    "instanceID",
						Type:          "test.event",
						Version:       "v1",
					},
				},
			},
		},
		{
			name: "invalid data",
			args: args{
				instanceID: "instanceID",
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return `{"data":""`
						},
						false),
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no aggregate id",
			args: args{
				events: []Command{
					&testEvent{
						BaseEvent: *NewBaseEventForPush(
							service.WithService(authz.NewMockContext("instanceID", "resourceOwner", "editorUser"), "editorService"),
							NewAggregate(
								authz.NewMockContext("zitadel", "caos", "adlerhurst"),
								"",
								"test.aggregate",
								"v1",
							),
							"test.event",
						),
						data: func() interface{} {
							return nil
						},
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no aggregate type",
			args: args{
				events: []Command{
					&testEvent{
						BaseEvent: *NewBaseEventForPush(
							service.WithService(authz.NewMockContext("instanceID", "resourceOwner", "editorUser"), "editorService"),
							NewAggregate(
								authz.NewMockContext("zitadel", "caos", "adlerhurst"),
								"id",
								"",
								"v1",
							),
							"test.event",
						),
						data: func() interface{} {
							return nil
						},
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no aggregate version",
			args: args{
				events: []Command{
					&testEvent{
						BaseEvent: *NewBaseEventForPush(
							service.WithService(authz.NewMockContext("instanceID", "resourceOwner", "editorUser"), "editorService"),
							NewAggregate(
								authz.NewMockContext("zitadel", "caos", "adlerhurst"),
								"id",
								"test.aggregate",
								"",
							),
							"test.event",
						),
						data: func() interface{} {
							return nil
						},
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no event type",
			args: args{
				events: []Command{
					&testEvent{
						BaseEvent: *NewBaseEventForPush(
							service.WithService(authz.NewMockContext("instanceID", "resourceOwner", "editorUser"), "editorService"),
							NewAggregate(
								authz.NewMockContext("zitadel", "caos", "adlerhurst"),
								"id",
								"test.aggregate",
								"v1",
							),
							"",
						),
						data: func() interface{} {
							return nil
						},
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no resourceowner",
			args: args{
				events: []Command{
					&testEvent{
						BaseEvent: *NewBaseEventForPush(
							service.WithService(authz.NewMockContext("instanceID", "", "editorUser"), "editorService"),
							NewAggregate(
								authz.NewMockContext("zitadel", "", "adlerhurst"),
								"id",
								"test.aggregate",
								"v1",
							),
							"test.event",
						),
						data: func() interface{} {
							return nil
						},
					},
				},
			},
			res: res{
				wantErr: false,
				events: []*repository.Event{
					{
						AggregateID:   "id",
						AggregateType: "test.aggregate",
						Data:          []byte(nil),
						EditorService: "editorService",
						EditorUser:    "editorUser",
						ResourceOwner: sql.NullString{String: "", Valid: false},
						InstanceID:    "zitadel",
						Type:          "test.event",
						Version:       "v1",
					},
				},
			},
		},
		{
			name: "multiple aggregates",
			args: args{
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
					newTestEvent(
						"2",
						"",
						func() interface{} {
							return nil
						},
						true),
				},
			},
			res: res{
				wantErr: false,
				events: combineEventLists(
					[]*repository.Event{
						{
							AggregateID:   "1",
							AggregateType: "test.aggregate",
							Data:          []byte(nil),
							EditorService: "editorService",
							EditorUser:    "editorUser",
							ResourceOwner: sql.NullString{String: "caos", Valid: true},
							InstanceID:    "zitadel",
							Type:          "test.event",
							Version:       "v1",
						},
						{
							AggregateID:   "1",
							AggregateType: "test.aggregate",
							Data:          []byte(nil),
							EditorService: "editorService",
							EditorUser:    "editorUser",
							ResourceOwner: sql.NullString{String: "caos", Valid: true},
							InstanceID:    "zitadel",
							Type:          "test.event",
							Version:       "v1",
						},
					},
					[]*repository.Event{
						{
							AggregateID:   "2",
							AggregateType: "test.aggregate",
							Data:          []byte(nil),
							EditorService: "editorService",
							EditorUser:    "editorUser",
							ResourceOwner: sql.NullString{String: "caos", Valid: true},
							InstanceID:    "zitadel",
							Type:          "test.event",
							Version:       "v1",
						},
					},
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, _, err := commandsToRepository(tt.args.instanceID, tt.args.events)
			if (err != nil) != tt.res.wantErr {
				t.Errorf("Eventstore.aggregatesToEvents() error = %v, wantErr %v", err, tt.res.wantErr)
				return
			}

			if err != nil {
				return
			}

			if len(tt.res.events) != len(events) {
				t.Errorf("length of events unequal want: %d got %d", len(tt.res.events), len(events))
				return
			}

			for i := 0; i < len(tt.res.events); i++ {
				compareEvents(t, tt.res.events[i], events[i])
			}
		})
	}
}

type testRepo struct {
	events    []*repository.Event
	sequence  uint64
	instances []string
	err       error
	t         *testing.T
}

func (repo *testRepo) Health(ctx context.Context) error {
	return nil
}

func (repo *testRepo) CreateInstance(ctx context.Context, instance string) error {
	return nil
}

func (repo *testRepo) Step20(context.Context, uint64) error { return nil }

func (repo *testRepo) Push(ctx context.Context, events []*repository.Event, uniqueConstraints ...*repository.UniqueConstraint) error {
	if repo.err != nil {
		return repo.err
	}

	if len(repo.events) != len(events) {
		repo.t.Errorf("length of events unequal want: %d got %d", len(repo.events), len(events))
		return fmt.Errorf("")
	}

	for i := 0; i < len(repo.events); i++ {
		compareEvents(repo.t, repo.events[i], events[i])
	}

	return nil
}

func (repo *testRepo) Filter(ctx context.Context, searchQuery *repository.SearchQuery) ([]*repository.Event, error) {
	if repo.err != nil {
		return nil, repo.err
	}
	return repo.events, nil
}

func (repo *testRepo) LatestSequence(ctx context.Context, queryFactory *repository.SearchQuery) (uint64, error) {
	if repo.err != nil {
		return 0, repo.err
	}
	return repo.sequence, nil
}

func (repo *testRepo) InstanceIDs(ctx context.Context, queryFactory *repository.SearchQuery) ([]string, error) {
	if repo.err != nil {
		return nil, repo.err
	}
	return repo.instances, nil
}

func TestEventstore_Push(t *testing.T) {
	type args struct {
		events []Command
	}
	type fields struct {
		repo        *testRepo
		eventMapper map[EventType]func(*repository.Event) (Event, error)
	}
	type res struct {
		wantErr bool
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		res    res
	}{
		{
			name: "one aggregate one event",
			args: args{
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
				},
			},
			fields: fields{
				repo: &testRepo{
					t: t,
					events: []*repository.Event{
						{
							AggregateID:   "1",
							AggregateType: "test.aggregate",
							Data:          []byte(nil),
							EditorService: "editorService",
							EditorUser:    "editorUser",
							ResourceOwner: sql.NullString{String: "caos", Valid: true},
							InstanceID:    "zitadel",
							Type:          "test.event",
							Version:       "v1",
						},
					},
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
		},
		{
			name: "one aggregate multiple events",
			args: args{
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
				},
			},
			fields: fields{
				repo: &testRepo{
					t: t,
					events: []*repository.Event{
						{
							AggregateID:   "1",
							AggregateType: "test.aggregate",
							Data:          []byte(nil),
							EditorService: "editorService",
							EditorUser:    "editorUser",
							ResourceOwner: sql.NullString{String: "caos", Valid: true},
							InstanceID:    "zitadel",
							Type:          "test.event",
							Version:       "v1",
						},
						{
							AggregateID:   "1",
							AggregateType: "test.aggregate",
							Data:          []byte(nil),
							EditorService: "editorService",
							EditorUser:    "editorUser",
							ResourceOwner: sql.NullString{String: "caos", Valid: true},
							InstanceID:    "zitadel",
							Type:          "test.event",
							Version:       "v1",
						},
					},
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: false,
			},
		},
		{
			name: "multiple aggregates",
			args: args{
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
					newTestEvent(
						"2",
						"",
						func() interface{} {
							return nil
						},
						true),
				},
			},
			fields: fields{
				repo: &testRepo{
					t: t,
					events: combineEventLists(
						[]*repository.Event{
							{
								AggregateID:   "1",
								AggregateType: "test.aggregate",
								Data:          []byte(nil),
								EditorService: "editorService",
								EditorUser:    "editorUser",
								ResourceOwner: sql.NullString{String: "caos", Valid: true},
								InstanceID:    "zitadel",
								Type:          "test.event",
								Version:       "v1",
							},
							{
								AggregateID:   "1",
								AggregateType: "test.aggregate",
								Data:          []byte(nil),
								EditorService: "editorService",
								EditorUser:    "editorUser",
								ResourceOwner: sql.NullString{String: "caos", Valid: true},
								InstanceID:    "zitadel",
								Type:          "test.event",
								Version:       "v1",
							},
						},
						[]*repository.Event{
							{
								AggregateID:   "2",
								AggregateType: "test.aggregate",
								Data:          []byte(nil),
								EditorService: "editorService",
								EditorUser:    "editorUser",
								ResourceOwner: sql.NullString{String: "caos", Valid: true},
								InstanceID:    "zitadel",
								Type:          "test.event",
								Version:       "v1",
							},
						},
					),
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: false,
			},
		},
		{
			name: "push fails",
			args: args{
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return nil
						},
						false),
				},
			},
			fields: fields{
				repo: &testRepo{
					t:   t,
					err: errors.ThrowInternal(nil, "V2-qaa4S", "test err"),
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "aggreagtes to events mapping fails",
			args: args{
				events: []Command{
					newTestEvent(
						"1",
						"",
						func() interface{} {
							return `{"data":""`
						},
						false),
				},
			},
			fields: fields{
				repo: &testRepo{
					t:   t,
					err: errors.ThrowInternal(nil, "V2-qaa4S", "test err"),
				},
			},
			res: res{
				wantErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Eventstore{
				repo:              tt.fields.repo,
				interceptorMutex:  sync.Mutex{},
				eventInterceptors: map[EventType]eventTypeInterceptors{},
			}
			for eventType, mapper := range tt.fields.eventMapper {
				es = es.RegisterFilterEventMapper(eventType, mapper)
			}
			if len(es.eventInterceptors) != len(tt.fields.eventMapper) {
				t.Errorf("register event mapper failed expected mapper amount: %d, got: %d", len(tt.fields.eventMapper), len(es.eventInterceptors))
				t.FailNow()
			}

			_, err := es.Push(context.Background(), tt.args.events...)
			if (err != nil) != tt.res.wantErr {
				t.Errorf("Eventstore.aggregatesToEvents() error = %v, wantErr %v", err, tt.res.wantErr)
			}
		})
	}
}

func TestEventstore_FilterEvents(t *testing.T) {
	type args struct {
		query *SearchQueryBuilder
	}
	type fields struct {
		repo        *testRepo
		eventMapper map[EventType]func(*repository.Event) (Event, error)
	}
	type res struct {
		wantErr bool
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		res    res
	}{
		{
			name: "invalid factory",
			args: args{
				query: nil,
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no events",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsEvent,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"no.aggregates"},
						},
					},
				},
			},
			fields: fields{
				repo: &testRepo{
					events: []*repository.Event{},
					t:      t,
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: false,
			},
		},
		{
			name: "repo error",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsEvent,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"no.aggregates"},
						},
					},
				},
			},
			fields: fields{
				repo: &testRepo{
					t:   t,
					err: errors.ThrowInternal(nil, "V2-RfkBa", "test err"),
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "found events",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsEvent,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"test.aggregate"},
						},
					},
				},
			},
			fields: fields{
				repo: &testRepo{
					events: []*repository.Event{
						{
							AggregateID: "test.aggregate",
							Type:        "test.event",
						},
					},
					t: t,
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Eventstore{
				repo:              tt.fields.repo,
				interceptorMutex:  sync.Mutex{},
				eventInterceptors: map[EventType]eventTypeInterceptors{},
			}

			for eventType, mapper := range tt.fields.eventMapper {
				es = es.RegisterFilterEventMapper(eventType, mapper)
			}
			if len(es.eventInterceptors) != len(tt.fields.eventMapper) {
				t.Errorf("register event mapper failed expected mapper amount: %d, got: %d", len(tt.fields.eventMapper), len(es.eventInterceptors))
				t.FailNow()
			}

			_, err := es.Filter(context.Background(), tt.args.query)
			if (err != nil) != tt.res.wantErr {
				t.Errorf("Eventstore.aggregatesToEvents() error = %v, wantErr %v", err, tt.res.wantErr)
			}
		})
	}
}

func TestEventstore_LatestSequence(t *testing.T) {
	type args struct {
		query *SearchQueryBuilder
	}
	type fields struct {
		repo *testRepo
	}
	type res struct {
		wantErr bool
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		res    res
	}{
		{
			name: "invalid factory",
			args: args{
				query: nil,
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no events",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsMaxSequence,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"no.aggregates"},
						},
					},
				},
			},
			fields: fields{
				repo: &testRepo{
					events: []*repository.Event{},
					t:      t,
				},
			},
			res: res{
				wantErr: false,
			},
		},
		{
			name: "repo error",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsMaxSequence,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"no.aggregates"},
						},
					},
				},
			},
			fields: fields{
				repo: &testRepo{
					t:   t,
					err: errors.ThrowInternal(nil, "V2-RfkBa", "test err"),
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "found events",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsMaxSequence,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"test.aggregate"},
						},
					},
				},
			},
			fields: fields{
				repo: &testRepo{
					sequence: 50,
					t:        t,
				},
			},
			res: res{
				wantErr: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Eventstore{
				repo: tt.fields.repo,
			}

			_, err := es.LatestSequence(context.Background(), tt.args.query)
			if (err != nil) != tt.res.wantErr {
				t.Errorf("Eventstore.aggregatesToEvents() error = %v, wantErr %v", err, tt.res.wantErr)
			}
		})
	}
}

type testReducer struct {
	t              *testing.T
	events         []Event
	expectedLength int
	err            error
}

func (r *testReducer) Reduce() error {
	r.t.Helper()
	if len(r.events) != r.expectedLength {
		r.t.Errorf("wrong amount of append events wanted: %d, got %d", r.expectedLength, len(r.events))
	}
	if r.err != nil {
		return r.err
	}
	return nil
}

func (r *testReducer) AppendEvents(e ...Event) {
	r.events = append(r.events, e...)
}

func TestEventstore_FilterToReducer(t *testing.T) {
	type args struct {
		query     *SearchQueryBuilder
		readModel reducer
	}
	type fields struct {
		repo        *testRepo
		eventMapper map[EventType]func(*repository.Event) (Event, error)
	}
	type res struct {
		wantErr bool
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		res    res
	}{
		{
			name: "invalid factory",
			args: args{
				query: nil,
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "no events",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsEvent,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"no.aggregates"},
						},
					},
				},
				readModel: &testReducer{
					t:              t,
					expectedLength: 0,
				},
			},
			fields: fields{
				repo: &testRepo{
					events: []*repository.Event{},
					t:      t,
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: false,
			},
		},
		{
			name: "repo error",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsEvent,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"no.aggregates"},
						},
					},
				},
				readModel: &testReducer{
					t:              t,
					expectedLength: 0,
				},
			},
			fields: fields{
				repo: &testRepo{
					t:   t,
					err: errors.ThrowInternal(nil, "V2-RfkBa", "test err"),
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "found events",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsEvent,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"test.aggregate"},
						},
					},
				},
				readModel: &testReducer{
					t:              t,
					expectedLength: 1,
				},
			},
			fields: fields{
				repo: &testRepo{
					events: []*repository.Event{
						{
							AggregateID: "test.aggregate",
							Type:        "test.event",
						},
					},
					t: t,
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
		},
		{
			name: "append in reducer fails",
			args: args{
				query: &SearchQueryBuilder{
					columns: repository.ColumnsEvent,
					queries: []*SearchQuery{
						{
							builder:        &SearchQueryBuilder{},
							aggregateTypes: []AggregateType{"test.aggregate"},
						},
					},
				},
				readModel: &testReducer{
					t:              t,
					err:            errors.ThrowInvalidArgument(nil, "V2-W06TG", "test err"),
					expectedLength: 1,
				},
			},
			fields: fields{
				repo: &testRepo{
					events: []*repository.Event{
						{
							AggregateID: "test.aggregate",
							Type:        "test.event",
						},
					},
					t: t,
				},
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(e *repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Eventstore{
				repo:              tt.fields.repo,
				interceptorMutex:  sync.Mutex{},
				eventInterceptors: map[EventType]eventTypeInterceptors{},
			}
			for eventType, mapper := range tt.fields.eventMapper {
				es = es.RegisterFilterEventMapper(eventType, mapper)
			}
			if len(es.eventInterceptors) != len(tt.fields.eventMapper) {
				t.Errorf("register event mapper failed expected mapper amount: %d, got: %d", len(tt.fields.eventMapper), len(es.eventInterceptors))
				t.FailNow()
			}

			err := es.FilterToReducer(context.Background(), tt.args.query, tt.args.readModel)
			if (err != nil) != tt.res.wantErr {
				t.Errorf("Eventstore.aggregatesToEvents() error = %v, wantErr %v", err, tt.res.wantErr)
			}
		})
	}
}

func combineEventLists(lists ...[]*repository.Event) []*repository.Event {
	events := []*repository.Event{}
	for _, list := range lists {
		events = append(events, list...)
	}
	return events
}

func compareEvents(t *testing.T, want, got *repository.Event) {
	t.Helper()

	if want.AggregateID != got.AggregateID {
		t.Errorf("wrong aggregateID got %q want %q", got.AggregateID, want.AggregateID)
	}
	if want.AggregateType != got.AggregateType {
		t.Errorf("wrong aggregateType got %q want %q", got.AggregateType, want.AggregateType)
	}
	if !reflect.DeepEqual(want.Data, got.Data) {
		t.Errorf("wrong data got %s want %s", string(got.Data), string(want.Data))
	}
	if want.EditorService != got.EditorService {
		t.Errorf("wrong editor service got %q want %q", got.EditorService, want.EditorService)
	}
	if want.EditorUser != got.EditorUser {
		t.Errorf("wrong editor user got %q want %q", got.EditorUser, want.EditorUser)
	}
	if want.ResourceOwner != got.ResourceOwner {
		t.Errorf("wrong resource owner got %q want %q", got.ResourceOwner.String, want.ResourceOwner.String)
	}
	if want.Type != got.Type {
		t.Errorf("wrong event type got %q want %q", got.Type, want.Type)
	}
	if want.Version != got.Version {
		t.Errorf("wrong version got %q want %q", got.Version, want.Version)
	}
	if want.PreviousAggregateSequence != got.PreviousAggregateSequence {
		t.Errorf("wrong previous sequence got %d want %d", got.PreviousAggregateSequence, want.PreviousAggregateSequence)
	}
}

func TestEventstore_mapEvents(t *testing.T) {
	type fields struct {
		eventMapper map[EventType]func(*repository.Event) (Event, error)
	}
	type args struct {
		events []*repository.Event
	}
	type res struct {
		events  []Event
		wantErr bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		//{
		//	name: "no mapper",
		//	args: args{
		//		events: []*repository.Event{
		//			{
		//				Type: "no.mapper.found",
		//			},
		//		},
		//	},
		//	fields: fields{
		//		eventMapper: map[EventType]func(*repository.Event) (EventReader, error){},
		//	},
		//	res: res{
		//		//TODO: as long as not all events are implemented in v2 eventstore doesn't return an error
		//		// afterwards it will return an error on un
		//		wantErr: true,
		//	},
		//},
		{
			name: "mapping failed",
			args: args{
				events: []*repository.Event{
					{
						Type: "test.event",
					},
				},
			},
			fields: fields{
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(*repository.Event) (Event, error) {
						return nil, errors.ThrowInternal(nil, "V2-8FbQk", "test err")
					},
				},
			},
			res: res{
				wantErr: true,
			},
		},
		{
			name: "mapping succeeded",
			args: args{
				events: []*repository.Event{
					{
						Type: "test.event",
					},
				},
			},
			fields: fields{
				eventMapper: map[EventType]func(*repository.Event) (Event, error){
					"test.event": func(*repository.Event) (Event, error) {
						return &testEvent{}, nil
					},
				},
			},
			res: res{
				events: []Event{
					&testEvent{},
				},
				wantErr: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Eventstore{
				interceptorMutex:  sync.Mutex{},
				eventInterceptors: map[EventType]eventTypeInterceptors{},
			}
			for eventType, mapper := range tt.fields.eventMapper {
				es = es.RegisterFilterEventMapper(eventType, mapper)
			}
			if len(es.eventInterceptors) != len(tt.fields.eventMapper) {
				t.Errorf("register event mapper failed expected mapper amount: %d, got: %d", len(tt.fields.eventMapper), len(es.eventInterceptors))
				t.FailNow()
			}

			gotMappedEvents, err := es.mapEvents(tt.args.events)
			if (err != nil) != tt.res.wantErr {
				t.Errorf("Eventstore.mapEvents() error = %v, wantErr %v", err, tt.res.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMappedEvents, tt.res.events) {
				t.Errorf("Eventstore.mapEvents() = %v, want %v", gotMappedEvents, tt.res.events)
			}
		})
	}
}
