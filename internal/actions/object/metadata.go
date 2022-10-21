package object

import (
	"encoding/json"
	"time"

	"github.com/dop251/goja"
	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/actions"
	"github.com/dennigogo/zitadel/internal/query"
)

func UserMetadataListFromQuery(c *actions.FieldConfig, metadata *query.UserMetadataList) goja.Value {
	result := &userMetadataList{
		Count:     metadata.Count,
		Sequence:  metadata.Sequence,
		Timestamp: metadata.Timestamp,
		Metadata:  make([]*userMetadata, len(metadata.Metadata)),
	}

	for i, md := range metadata.Metadata {
		var value interface{}
		err := json.Unmarshal(md.Value, &value)
		if err != nil {
			logging.WithError(err).Debug("unable to unmarshal into map")
			panic(err)
		}
		result.Metadata[i] = &userMetadata{
			CreationDate:  md.CreationDate,
			ChangeDate:    md.ChangeDate,
			ResourceOwner: md.ResourceOwner,
			Sequence:      md.Sequence,
			Key:           md.Key,
			Value:         c.Runtime.ToValue(value),
		}
	}

	return c.Runtime.ToValue(result)
}

type userMetadataList struct {
	Count     uint64
	Sequence  uint64
	Timestamp time.Time
	Metadata  []*userMetadata
}

type userMetadata struct {
	CreationDate  time.Time
	ChangeDate    time.Time
	ResourceOwner string
	Sequence      uint64
	Key           string
	Value         goja.Value
}
