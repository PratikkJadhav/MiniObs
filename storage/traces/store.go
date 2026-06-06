package storage

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

type Store struct {
	segment       *segment
	index         *index
	currentFileID uint32
	currentSize   int64
	dir           string
}

type TraceSummary struct {
	TraceID           string `json:"trace_id"`
	ServiceName       string `json:"service_name"`
	Name              string `json:"name"`
	DurationNs        uint64 `json:"duration_ns"`
	StartTimeUnixNano uint64 `json:"start_time_unix_nano"`
	Status            int    `json:"status"`
}

func NewStore(dir string) (*Store, error) {

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	fileID := uint32(1)
	file := fmt.Sprintf("seg-%06d.dat", fileID)
	filePath := filepath.Join(dir, file)

	seg, err := openSegment(1, filePath)
	if err != nil {
		return nil, err
	}

	idx := &index{
		traces:   make(map[string][]location),
		services: make(map[string][]string),
	}
	s := &Store{
		segment:       seg,
		index:         idx,
		currentFileID: fileID,
		currentSize:   0,
		dir:           dir,
	}

	if err := s.loadHint(); err != nil {
		return nil, err
	}

	return s, nil
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

	s.currentSize += int64(4 + len(spanBytes))

	if s.currentSize >= 64*1024*1024 {
		s.currentFileID++
		s.currentSize = 0

		newFile := fmt.Sprintf("seg-%06d.dat", s.currentFileID)
		newPath := filepath.Join(s.dir, newFile)

		newSeg, err := openSegment(s.currentFileID, newPath)
		if err != nil {
			return err
		}
		s.segment = newSeg
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

func (s *Store) GetTraceByID(traceID string) ([]*tracepb.Span, error) {
	return s.Read(traceID)
}

func (idx *index) All() map[string][]location {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.traces
}

func (idx *index) GetServiceForTrace(traceID string) string {
	for serviceName, traceIDs := range idx.services {
		for _, id := range traceIDs {
			if id == traceID {
				return serviceName
			}
		}
	}
	return ""
}

func (s *Store) SaveHint() error {
	hintPath := filepath.Join(s.dir, "hint.dat")
	f, err := os.Create(hintPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for traceID, locs := range s.index.All() {
		serviceName := s.index.GetServiceForTrace(traceID)
		for _, loc := range locs {
			fmt.Fprintf(f, "%s|%d|%d|%d|%s\n",
				traceID, loc.fileID, loc.offset, loc.size, serviceName)
		}
	}
	return nil
}

func (s *Store) loadHint() error {
	hintPath := filepath.Join(s.dir, "hint.dat")

	f, err := os.Open(hintPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // first startup, no hint file yet, that's fine
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")

		if len(parts) != 5 {
			continue // skip malformed lines
		}

		traceID := parts[0]
		fileID, _ := strconv.ParseUint(parts[1], 10, 32)
		offset, _ := strconv.ParseInt(parts[2], 10, 64)
		size, _ := strconv.ParseUint(parts[3], 10, 32)

		serviceName := parts[4]

		loc := location{
			fileID: uint32(fileID),
			offset: offset,
			size:   uint32(size),
		}

		s.index.Add(traceID, serviceName, loc)

	}

	return scanner.Err()

}

func (s *Store) GetTraceSummaries(serviceName string) ([]TraceSummary, error) {
	traceIDs := s.GetTraceIDs(serviceName)
	var summaries []TraceSummary

	for _, traceID := range traceIDs {
		spans, err := s.Read(traceID)
		if err != nil {
			return nil, err
		}
		if len(spans) == 0 {
			continue
		}

		rootSpan := spans[0]
		summary := TraceSummary{
			TraceID:           traceID,
			ServiceName:       serviceName,
			Name:              rootSpan.GetName(),
			StartTimeUnixNano: rootSpan.GetStartTimeUnixNano(),
			DurationNs:        rootSpan.GetEndTimeUnixNano() - rootSpan.GetStartTimeUnixNano(),
			Status:            int(rootSpan.GetStatus().GetCode()),
		}

		summaries = append(summaries, summary)

	}

	return summaries, nil
}
