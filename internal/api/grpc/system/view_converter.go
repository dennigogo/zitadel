package system

import (
	"github.com/dennigogo/zitadel/internal/query"
	"github.com/dennigogo/zitadel/internal/view/model"
	system_pb "github.com/dennigogo/zitadel/pkg/grpc/system"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ViewsToPb(views []*model.View) []*system_pb.View {
	v := make([]*system_pb.View, len(views))
	for i, view := range views {
		v[i] = ViewToPb(view)
	}
	return v
}

func ViewToPb(view *model.View) *system_pb.View {
	return &system_pb.View{
		Database:                 view.Database,
		ViewName:                 view.ViewName,
		LastSuccessfulSpoolerRun: timestamppb.New(view.LastSuccessfulSpoolerRun),
		ProcessedSequence:        view.CurrentSequence,
		EventTimestamp:           timestamppb.New(view.EventTimestamp),
	}
}

func CurrentSequencesToPb(database string, currentSequences *query.CurrentSequences) []*system_pb.View {
	v := make([]*system_pb.View, len(currentSequences.CurrentSequences))
	for i, currentSequence := range currentSequences.CurrentSequences {
		v[i] = CurrentSequenceToPb(database, currentSequence)
	}
	return v
}

func CurrentSequenceToPb(database string, currentSequence *query.CurrentSequence) *system_pb.View {
	return &system_pb.View{
		Database:          database,
		ViewName:          currentSequence.ProjectionName,
		ProcessedSequence: currentSequence.CurrentSequence,
		EventTimestamp:    timestamppb.New(currentSequence.Timestamp),
	}
}
