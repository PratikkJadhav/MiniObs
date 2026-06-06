package query

import (
	"sort"

	"github.com/PratikkJadhav/MiniObs/storage"
)

func ComputeMetrics(store *storage.Store, serviceName string) (p50, p95, p99 float64, errorRate float64, err error) {
	var durations []float64
	totalSpans := 0
	errorCount := 0

	traceIDs := store.GetTraceIDs(serviceName)

	for _, traceID := range traceIDs {
		spans, err := store.GetTraceByID(traceID)
		if err != nil {
			return 0, 0, 0, 0, err
		}

		for _, s := range spans {

			startTime := s.GetStartTimeUnixNano()
			endTime := s.GetEndTimeUnixNano()
			latencyMs := float64(endTime-startTime) / 1e6
			durations = append(durations, latencyMs)

			ers := s.GetStatus().GetCode()

			if ers == 2 {
				errorCount++
			}
			totalSpans++
		}
	}

	sort.Float64s(durations)

	if totalSpans == 0 {
		return 0, 0, 0, 0, nil
	}

	n := len(durations)
	p50 = durations[n*50/100]
	p95 = durations[n*95/100]
	p99 = durations[n*99/100]

	errorRate = float64(errorCount) / float64(totalSpans) * 100

	return p50, p95, p99, errorRate, nil
}
