package storage

import (
	"encoding/hex"
	"os"
	"path/filepath"

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

func (s *Store) Write(span *tracepb.Span) error {
	spanBytes, err := proto.Marshal(span)
	if err != nil {
		return err
	}

	loc := s.Write(spanBytes)
	traceID := hex.EncodeToString(span.GetTraceId())
	err := s.index.Add(traceID, serviceName, loc)

	return nil

}
