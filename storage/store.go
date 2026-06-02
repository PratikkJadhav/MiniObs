package storage

import (
	"encoding/hex"
	"os"
	"path/filepath"

	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

type Store struct {
	segment *segment
	index   *index
}

func NewStore(dir string) (*Store, error) {

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dir, "seg-000001.dat")

	seg, err := openSegment(1, filePath)
	if err != nil {
		return nil, err
	}

	idx := &index{
		traces:   make(map[string][]location),
		services: make(map[string][]string),
	}

	return &Store{
		segment: seg,
		index:   idx,
	}, nil

}

func (s *Store) Write(serviceName string, span *tracepb.Span) error {
	spanBytes, err := proto.Marshal(span)
	if err != nil {
		return err
	}

	loc, err := s.segment.Write(spanBytes)
	if err != nil {
		return err
	}
	traceID := hex.EncodeToString(span.GetTraceId())
	s.index.Add(traceID, serviceName, loc)

	return nil

}

func (s *Store) Read(traceID string) ([]*tracepb.Span, error) {
	locs := s.index.GetTrace(traceID)

	var spans []*tracepb.Span

	for _, loc := range locs {
		spanBytes, err := s.segment.Read(loc)
		if err != nil {
			return nil, err
		}

		span := &tracepb.Span{}

		err = proto.Unmarshal(spanBytes, span)
		if err != nil {
			return nil, err
		}

		spans = append(spans, span)

	}

	return spans, nil

}

func (s *Store) GetServices() []string {
	var slice []string

	slice = s.index.GetServices()

	return slice
}

func (s *Store) GetTraceIDs(serviceName string) []string {

	return s.index.GetTracesByService(serviceName)
}
