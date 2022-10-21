package admin

import (
	"testing"

	"github.com/dennigogo/zitadel/internal/test"
	"github.com/dennigogo/zitadel/internal/view/model"
	admin_pb "github.com/dennigogo/zitadel/pkg/grpc/admin"
)

func TestFailedEventsToPbFields(t *testing.T) {
	type args struct {
		failedEvents []*model.FailedEvent
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "all fields",
			args: args{
				failedEvents: []*model.FailedEvent{
					{
						Database:       "admin",
						ViewName:       "users",
						FailedSequence: 456,
						FailureCount:   5,
						ErrMsg:         "some error",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FailedEventsViewToPb(tt.args.failedEvents)
			for _, g := range got {
				test.AssertFieldsMapped(t, g)
			}
		})
	}
}

func TestFailedEventToPbFields(t *testing.T) {
	type args struct {
		failedEvent *model.FailedEvent
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"all fields",
			args{
				failedEvent: &model.FailedEvent{
					Database:       "admin",
					ViewName:       "users",
					FailedSequence: 456,
					FailureCount:   5,
					ErrMsg:         "some error",
				},
			},
		},
	}
	for _, tt := range tests {
		converted := FailedEventViewToPb(tt.args.failedEvent)
		test.AssertFieldsMapped(t, converted)
	}
}

func TestRemoveFailedEventRequestToModelFields(t *testing.T) {
	type args struct {
		req *admin_pb.RemoveFailedEventRequest
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"all fields",
			args{
				req: &admin_pb.RemoveFailedEventRequest{
					Database:       "admin",
					ViewName:       "users",
					FailedSequence: 456,
				},
			},
		},
	}
	for _, tt := range tests {
		converted := RemoveFailedEventRequestToModel(tt.args.req)
		test.AssertFieldsMapped(t, converted, "FailureCount", "ErrMsg")
	}
}
