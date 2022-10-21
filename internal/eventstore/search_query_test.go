package eventstore

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/dennigogo/zitadel/internal/database"
	"github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/eventstore/repository"
)

func testAddQuery(queryFuncs ...func(*SearchQuery) *SearchQuery) func(*SearchQueryBuilder) *SearchQueryBuilder {
	return func(builder *SearchQueryBuilder) *SearchQueryBuilder {
		query := builder.AddQuery()
		for _, queryFunc := range queryFuncs {
			queryFunc(query)
		}
		return query.Builder()
	}
}

func testSetColumns(columns Columns) func(factory *SearchQueryBuilder) *SearchQueryBuilder {
	return func(factory *SearchQueryBuilder) *SearchQueryBuilder {
		factory = factory.Columns(columns)
		return factory
	}
}

func testSetLimit(limit uint64) func(builder *SearchQueryBuilder) *SearchQueryBuilder {
	return func(builder *SearchQueryBuilder) *SearchQueryBuilder {
		builder = builder.Limit(limit)
		return builder
	}
}

func testOr(queryFuncs ...func(*SearchQuery) *SearchQuery) func(*SearchQuery) *SearchQuery {
	return func(query *SearchQuery) *SearchQuery {
		subQuery := query.Or()
		for _, queryFunc := range queryFuncs {
			queryFunc(subQuery)
		}
		return subQuery
	}
}

func testSetAggregateTypes(types ...AggregateType) func(*SearchQuery) *SearchQuery {
	return func(query *SearchQuery) *SearchQuery {
		query = query.AggregateTypes(types...)
		return query
	}
}

func testSetSequenceGreater(sequence uint64) func(*SearchQuery) *SearchQuery {
	return func(query *SearchQuery) *SearchQuery {
		query = query.SequenceGreater(sequence)
		return query
	}
}

func testSetSequenceLess(sequence uint64) func(*SearchQuery) *SearchQuery {
	return func(query *SearchQuery) *SearchQuery {
		query = query.SequenceLess(sequence)
		return query
	}
}

func testSetAggregateIDs(aggregateIDs ...string) func(*SearchQuery) *SearchQuery {
	return func(query *SearchQuery) *SearchQuery {
		query = query.AggregateIDs(aggregateIDs...)
		return query
	}
}

func testSetEventTypes(eventTypes ...EventType) func(*SearchQuery) *SearchQuery {
	return func(query *SearchQuery) *SearchQuery {
		query = query.EventTypes(eventTypes...)
		return query
	}
}

func testSetResourceOwner(resourceOwner string) func(*SearchQueryBuilder) *SearchQueryBuilder {
	return func(builder *SearchQueryBuilder) *SearchQueryBuilder {
		builder = builder.ResourceOwner(resourceOwner)
		return builder
	}
}

func testSetCreationDateAfter(date time.Time) func(*SearchQuery) *SearchQuery {
	return func(query *SearchQuery) *SearchQuery {
		query = query.CreationDateAfter(date)
		return query
	}
}

func testSetSortOrder(asc bool) func(*SearchQueryBuilder) *SearchQueryBuilder {
	return func(query *SearchQueryBuilder) *SearchQueryBuilder {
		if asc {
			query = query.OrderAsc()
		} else {
			query = query.OrderDesc()
		}
		return query
	}
}

func TestSearchQuerybuilderSetters(t *testing.T) {
	type args struct {
		columns Columns
		setters []func(*SearchQueryBuilder) *SearchQueryBuilder
	}
	tests := []struct {
		name string
		args args
		res  *SearchQueryBuilder
	}{
		{
			name: "New builder",
			args: args{
				columns: ColumnsEvent,
			},
			res: &SearchQueryBuilder{
				columns: repository.Columns(ColumnsEvent),
			},
		},
		{
			name: "set columns",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testSetColumns(repository.ColumnsMaxSequence)},
			},
			res: &SearchQueryBuilder{
				columns: repository.ColumnsMaxSequence,
			},
		},
		{
			name: "set limit",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testSetLimit(100)},
			},
			res: &SearchQueryBuilder{
				limit: 100,
			},
		},
		{
			name: "set sequence greater",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testAddQuery(testSetSequenceGreater(90))},
			},
			res: &SearchQueryBuilder{
				queries: []*SearchQuery{
					{
						eventSequenceGreater: 90,
					},
				},
			},
		},
		{
			name: "set sequence less",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testAddQuery(testSetSequenceLess(90))},
			},
			res: &SearchQueryBuilder{
				queries: []*SearchQuery{
					{
						eventSequenceLess: 90,
					},
				},
			},
		},
		{
			name: "set aggregateIDs",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testAddQuery(testSetAggregateIDs("1235", "09824"))},
			},
			res: &SearchQueryBuilder{
				queries: []*SearchQuery{
					{
						aggregateIDs: []string{"1235", "09824"},
					},
				},
			},
		},
		{
			name: "set eventTypes",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testAddQuery(testSetEventTypes("user.created", "user.updated"))},
			},
			res: &SearchQueryBuilder{
				queries: []*SearchQuery{
					{
						eventTypes: []EventType{"user.created", "user.updated"},
					},
				},
			},
		},
		{
			name: "set resource owner",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testSetResourceOwner("hodor")},
			},
			res: &SearchQueryBuilder{
				resourceOwner: "hodor",
			},
		},
		{
			name: "default search query",
			args: args{
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{testAddQuery(testSetAggregateTypes("user"), testSetAggregateIDs("1235", "024")), testSetSortOrder(false)},
			},
			res: &SearchQueryBuilder{
				desc: true,
				queries: []*SearchQuery{
					{
						aggregateTypes: []AggregateType{"user"},
						aggregateIDs:   []string{"1235", "024"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSearchQueryBuilder(tt.args.columns)
			for _, setter := range tt.args.setters {
				builder = setter(builder)
			}

			assertBuilder(t, tt.res, builder)
		})
	}
}

func TestSearchQuerybuilderBuild(t *testing.T) {
	testNow := time.Now()
	type args struct {
		columns    Columns
		setters    []func(*SearchQueryBuilder) *SearchQueryBuilder
		instanceID string
	}
	type res struct {
		isErr func(err error) bool
		query *repository.SearchQuery
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "no aggregate types",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{},
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
				query: nil,
			},
		},
		{
			name: "invalid column (too low)",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testSetColumns(Columns(-1)),
					testAddQuery(testSetAggregateTypes("user")),
				},
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
		{
			name: "invalid column (too high)",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testSetColumns(math.MaxInt32),
					testAddQuery(testSetAggregateTypes("uesr")),
				},
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
		{
			name: "filter aggregate type",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(testSetAggregateTypes("user")),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate types",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(testSetAggregateTypes("user", "org")),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, database.StringArray{"user", "org"}, repository.OperationIn),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type, limit, desc",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testSetLimit(5),
					testSetSortOrder(false),
					testAddQuery(
						testSetSequenceGreater(100),
						testSetAggregateTypes("user"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    true,
					Limit:   5,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldSequence, uint64(100), repository.OperationLess),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type, limit, asc",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testSetLimit(5),
					testSetSortOrder(true),
					testAddQuery(
						testSetSequenceGreater(100),
						testSetAggregateTypes("user"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   5,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldSequence, uint64(100), repository.OperationGreater),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type, limit, desc, max event sequence cols",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testSetLimit(5),
					testSetSortOrder(false),
					testSetColumns(repository.ColumnsMaxSequence),
					testAddQuery(
						testSetSequenceGreater(100),
						testSetAggregateTypes("user"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsMaxSequence,
					Desc:    true,
					Limit:   5,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldSequence, uint64(100), repository.OperationLess),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type and aggregate id",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetAggregateIDs("1234"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldAggregateID, "1234", repository.OperationEquals),
						},
					},
				},
			},
		},
		{
			name: "filter multiple aggregate type and aggregate id",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetAggregateIDs("1234"),
						testOr(
							testSetAggregateTypes("org"),
							testSetAggregateIDs("izu"),
						),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldAggregateID, "1234", repository.OperationEquals),
						},
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("org"), repository.OperationEquals),
							repository.NewFilter(repository.FieldAggregateID, "izu", repository.OperationEquals),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type and aggregate ids",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetAggregateIDs("1234", "0815"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldAggregateID, database.StringArray{"1234", "0815"}, repository.OperationIn),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type and sequence greater",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetSequenceGreater(8),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldSequence, uint64(8), repository.OperationGreater),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type and event type",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetEventTypes("user.created"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldEventType, repository.EventType("user.created"), repository.OperationEquals),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type and event types",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetEventTypes("user.created", "user.changed"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldEventType, database.StringArray{"user.created", "user.changed"}, repository.OperationIn),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type resource owner",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testSetResourceOwner("hodor"),
					testAddQuery(
						testSetAggregateTypes("user"),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldResourceOwner, "hodor", repository.OperationEquals),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type and sequence between",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetSequenceGreater(8),
						testSetSequenceLess(16),
					),
				},
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldSequence, uint64(8), repository.OperationGreater),
							repository.NewFilter(repository.FieldSequence, uint64(16), repository.OperationLess),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type and instanceID",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
					),
				},
				instanceID: "instanceID",
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldInstanceID, "instanceID", repository.OperationEquals),
						},
					},
				},
			},
		},
		{
			name: "filter aggregate type, instanceID and creation date after",
			args: args{
				columns: ColumnsEvent,
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
						testSetCreationDateAfter(testNow),
					),
				},
				instanceID: "instanceID",
			},
			res: res{
				isErr: nil,
				query: &repository.SearchQuery{
					Columns: repository.ColumnsEvent,
					Desc:    false,
					Limit:   0,
					Filters: [][]*repository.Filter{
						{
							repository.NewFilter(repository.FieldAggregateType, repository.AggregateType("user"), repository.OperationEquals),
							repository.NewFilter(repository.FieldCreationDate, testNow, repository.OperationGreater),
							repository.NewFilter(repository.FieldInstanceID, "instanceID", repository.OperationEquals),
						},
					},
				},
			},
		},
		{
			name: "column invalid",
			args: args{
				columns: Columns(-1),
				setters: []func(*SearchQueryBuilder) *SearchQueryBuilder{
					testAddQuery(
						testSetAggregateTypes("user"),
					),
				},
			},
			res: res{
				isErr: errors.IsPreconditionFailed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSearchQueryBuilder(tt.args.columns)
			for _, f := range tt.args.setters {
				builder = f(builder)
			}
			query, err := builder.build(tt.args.instanceID)
			if tt.res.isErr != nil && !tt.res.isErr(err) {
				t.Errorf("wrong error(%T): %v", err, err)
				return
			}
			if err != nil && tt.res.isErr == nil {
				t.Errorf("no error expected: %v", err)
				return
			}

			assertRepoQuery(t, tt.res.query, query)
		})
	}
}

func assertBuilder(t *testing.T, want, got *SearchQueryBuilder) {
	t.Helper()

	if got.columns != want.columns {
		t.Errorf("wrong column: got: %v want: %v", got.columns, want.columns)
	}
	if got.desc != want.desc {
		t.Errorf("wrong desc: got: %v want: %v", got.desc, want.desc)
	}
	if got.limit != want.limit {
		t.Errorf("wrong limit: got: %v want: %v", got.limit, want.limit)
	}
	if got.resourceOwner != want.resourceOwner {
		t.Errorf("wrong : got: %v want: %v", got.resourceOwner, want.resourceOwner)
	}
	if len(got.queries) != len(want.queries) {
		t.Errorf("wrong length of queries: got: %v want: %v", len(got.queries), len(want.queries))
	}

	for i, query := range got.queries {
		assertQuery(t, i, want.queries[i], query)
	}
}

func assertQuery(t *testing.T, i int, want, got *SearchQuery) {
	t.Helper()

	if !reflect.DeepEqual(got.aggregateIDs, want.aggregateIDs) {
		t.Errorf("wrong aggregateIDs in query %d : got: %v want: %v", i, got.aggregateIDs, want.aggregateIDs)
	}
	if !reflect.DeepEqual(got.aggregateTypes, want.aggregateTypes) {
		t.Errorf("wrong aggregateTypes in query %d : got: %v want: %v", i, got.aggregateTypes, want.aggregateTypes)
	}
	if !reflect.DeepEqual(got.eventData, want.eventData) {
		t.Errorf("wrong eventData in query %d : got: %v want: %v", i, got.eventData, want.eventData)
	}
	if got.eventSequenceLess != want.eventSequenceLess {
		t.Errorf("wrong eventSequenceLess in query %d : got: %v want: %v", i, got.eventSequenceLess, want.eventSequenceLess)
	}
	if got.eventSequenceGreater != want.eventSequenceGreater {
		t.Errorf("wrong eventSequenceGreater in query %d : got: %v want: %v", i, got.eventSequenceGreater, want.eventSequenceGreater)
	}
	if !reflect.DeepEqual(got.eventTypes, want.eventTypes) {
		t.Errorf("wrong eventTypes in query %d : got: %v want: %v", i, got.eventTypes, want.eventTypes)
	}
}

func assertRepoQuery(t *testing.T, want, got *repository.SearchQuery) {
	t.Helper()

	if want == nil && got == nil {
		return
	}

	if !reflect.DeepEqual(got.Columns, want.Columns) {
		t.Errorf("wrong columns in query: got: %v want: %v", got.Columns, want.Columns)
	}
	if got.Desc != want.Desc {
		t.Errorf("wrong desc in query: got: %v want: %v", got.Desc, want.Desc)
	}
	if got.Limit != want.Limit {
		t.Errorf("wrong limit in query: got: %v want: %v", got.Limit, want.Limit)
	}

	if len(got.Filters) != len(want.Filters) {
		t.Errorf("wrong length of filters: got: %v want: %v", len(got.Filters), len(want.Filters))
	}

	for filterIdx, filter := range got.Filters {
		if len(got.Filters) != len(want.Filters) {
			t.Errorf("wrong length of subfilters: got: %v want: %v", len(filter), len(want.Filters[filterIdx]))
		}
		for subFilterIdx, f := range filter {
			assertFilters(t, subFilterIdx, want.Filters[filterIdx][subFilterIdx], f)
		}
	}
}

func assertFilters(t *testing.T, i int, want, got *repository.Filter) {
	t.Helper()

	if want.Field != got.Field {
		t.Errorf("wrong field in filter %d : got: %v want: %v", i, got.Field, want.Field)
	}
	if want.Operation != got.Operation {
		t.Errorf("wrong operation in filter %d : got: %v want: %v", i, got.Operation, want.Operation)
	}
	if !reflect.DeepEqual(want.Value, got.Value) {
		t.Errorf("wrong value in filter %d : got: %v want: %v", i, got.Value, want.Value)
	}
}

func TestSearchQuery_matches(t *testing.T) {
	type args struct {
		event Event
	}
	tests := []struct {
		name  string
		query *SearchQuery
		event Event
		want  bool
	}{
		{
			name:  "sequence too low",
			query: NewSearchQueryBuilder(ColumnsEvent).AddQuery().SequenceLess(10),
			event: &BaseEvent{
				sequence: 10,
			},
			want: false,
		},
		{
			name:  "sequence too high",
			query: NewSearchQueryBuilder(ColumnsEvent).AddQuery().SequenceGreater(60),
			event: &BaseEvent{
				sequence: 60,
			},
			want: false,
		},
		{
			name:  "wrong aggregate type",
			query: NewSearchQueryBuilder(ColumnsEvent).AddQuery().AggregateTypes("searched"),
			event: &BaseEvent{
				aggregate: Aggregate{
					Type: "found",
				},
			},
			want: false,
		},
		{
			name:  "wrong aggregate id",
			query: NewSearchQueryBuilder(ColumnsEvent).AddQuery().AggregateIDs("1", "10", "100"),
			event: &BaseEvent{
				aggregate: Aggregate{
					ID: "2",
				},
			},
			want: false,
		},
		{
			name:  "wrong event type",
			query: NewSearchQueryBuilder(ColumnsEvent).AddQuery().EventTypes("event.searched.type"),
			event: &BaseEvent{
				EventType: "event.actual.type",
			},
			want: false,
		},
		{
			name: "matching",
			query: NewSearchQueryBuilder(ColumnsEvent).AddQuery().
				SequenceLess(100).SequenceGreater(50).AggregateIDs("2").AggregateTypes("actual").EventTypes("event.actual.type"),
			event: &BaseEvent{
				sequence: 55,
				aggregate: Aggregate{
					ID:   "2",
					Type: "actual",
				},
				EventType: "event.actual.type",
			},
			want: true,
		},
		{
			name:  "matching empty query",
			query: NewSearchQueryBuilder(ColumnsEvent).AddQuery(),
			event: &BaseEvent{
				sequence: 55,
				aggregate: Aggregate{
					ID:   "2",
					Type: "actual",
				},
				EventType: "event.actual.type",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &SearchQuery{
				aggregateTypes:       tt.query.aggregateTypes,
				aggregateIDs:         tt.query.aggregateIDs,
				eventSequenceGreater: tt.query.eventSequenceGreater,
				eventSequenceLess:    tt.query.eventSequenceLess,
				eventTypes:           tt.query.eventTypes,
				eventData:            tt.query.eventData,
			}
			if got := query.matches(tt.event); got != tt.want {
				t.Errorf("SearchQuery.matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchQueryBuilder_Matches(t *testing.T) {
	type args struct {
		event       Event
		existingLen int
	}
	tests := []struct {
		name    string
		builder *SearchQueryBuilder
		args    args
		want    bool
	}{
		{
			name:    "limit exeeded",
			builder: NewSearchQueryBuilder(ColumnsEvent).Limit(100),
			args: args{
				event:       &BaseEvent{},
				existingLen: 100,
			},
			want: false,
		},
		{
			name:    "wrong resource owner",
			builder: NewSearchQueryBuilder(ColumnsEvent).ResourceOwner("query"),
			args: args{
				event: &BaseEvent{
					aggregate: Aggregate{
						ResourceOwner: "ro",
					},
				},
				existingLen: 0,
			},
			want: false,
		},
		{
			name:    "wrong instance",
			builder: NewSearchQueryBuilder(ColumnsEvent).InstanceID("instance"),
			args: args{
				event: &BaseEvent{
					aggregate: Aggregate{
						InstanceID: "different instance",
					},
				},
				existingLen: 0,
			},
			want: false,
		},
		{
			name: "query failed",
			builder: NewSearchQueryBuilder(ColumnsEvent).
				AddQuery().
				SequenceGreater(1000).
				Builder(),
			args: args{
				event: &BaseEvent{
					sequence: 999,
				},
				existingLen: 0,
			},
			want: false,
		},
		{
			name: "matching",
			builder: NewSearchQueryBuilder(ColumnsEvent).
				Limit(1000).
				ResourceOwner("ro").
				InstanceID("instance").
				AddQuery().
				SequenceGreater(1000).
				Builder(),
			args: args{
				event: &BaseEvent{
					aggregate: Aggregate{
						ResourceOwner: "ro",
						InstanceID:    "instance",
					},
					sequence: 1001,
				},
				existingLen: 999,
			},
			want: true,
		},
		{
			name:    "matching builder resourceOwner and Instance",
			builder: NewSearchQueryBuilder(ColumnsEvent),
			args: args{
				event: &BaseEvent{
					aggregate: Aggregate{
						ResourceOwner: "ro",
						InstanceID:    "instance",
					},
					sequence: 1001,
				},
				existingLen: 999,
			},
			want: true,
		},
		{
			name:    "matching builder resourceOwner only",
			builder: NewSearchQueryBuilder(ColumnsEvent),
			args: args{
				event: &BaseEvent{
					aggregate: Aggregate{
						ResourceOwner: "ro",
					},
					sequence: 1001,
				},
				existingLen: 999,
			},
			want: true,
		},
		{
			name:    "matching builder instanceID only",
			builder: NewSearchQueryBuilder(ColumnsEvent),
			args: args{
				event: &BaseEvent{
					aggregate: Aggregate{
						InstanceID: "instance",
					},
					sequence: 1001,
				},
				existingLen: 999,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.builder.Matches(tt.args.event, tt.args.existingLen); got != tt.want {
				t.Errorf("SearchQueryBuilder.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
