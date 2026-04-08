package graph

import (
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Time is a type alias for *timestamppb.Timestamp that implements GraphQL scalar marshaling.
// This allows direct binding of protobuf Timestamp fields to the GraphQL Time scalar.
type Time = *timestamppb.Timestamp

// MarshalTime marshals a protobuf Timestamp to a GraphQL Time scalar (RFC3339 string).
func MarshalTime(t Time) graphql.Marshaler {
	if t == nil {
		return graphql.Null
	}
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, `"`+t.AsTime().Format(time.RFC3339Nano)+`"`)
	})
}

// UnmarshalTime unmarshals a GraphQL Time scalar (RFC3339 string) to a protobuf Timestamp.
func UnmarshalTime(v interface{}) (Time, error) {
	switch v := v.(type) {
	case string:
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			// Try RFC3339 without nanoseconds
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, fmt.Errorf("Time must be RFC3339 formatted string: %w", err)
			}
		}
		return timestamppb.New(t), nil
	case time.Time:
		return timestamppb.New(v), nil
	default:
		return nil, fmt.Errorf("Time must be a string, got %T", v)
	}
}
