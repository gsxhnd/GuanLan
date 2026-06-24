package server

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func tsPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func floatPtr(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func dateStr(t time.Time) string {
	return t.Format("2006-01-02")
}
