package receiver

import (
	"context"
	"fmt"

	collectorv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

type Receiver struct {
	collectorv1.UnimplementedTraceServiceServer
}

func (r *Receiver) Export(ctx context.Context, req *collectorv1.ExportTraceServiceRequest) (*collectorv1.ExportTraceServiceResponse, error) {
	out := new(collectorv1.ExportTraceServiceResponse)
	for _, rs := range req.GetResourceSpans() {

		for _, ss := range rs.GetScopeSpans() {
			for _, span := range ss.GetSpans() {
				fmt.Println(span.GetName())
				fmt.Printf("traceID: %x\n", span.GetTraceId())
			}
		}
	}

	return out, nil
}
