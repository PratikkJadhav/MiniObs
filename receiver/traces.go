package receiver

import (
	"context"
	"fmt"

	"github.com/PratikkJadhav/MiniObs/storage"
	collectorv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

type Receiver struct {
	collectorv1.UnimplementedTraceServiceServer
	Store *storage.Store
}

func (r *Receiver) Export(ctx context.Context, req *collectorv1.ExportTraceServiceRequest) (*collectorv1.ExportTraceServiceResponse, error) {
	out := new(collectorv1.ExportTraceServiceResponse)
	for _, rs := range req.GetResourceSpans() {

		serviceName := "unknown_service"
		for _, attr := range rs.GetResource().GetAttributes() {
			if attr.GetKey() == "service.name" {
				serviceName = attr.GetValue().GetStringValue()
				break
			}
		}
		for _, ss := range rs.GetScopeSpans() {
			for _, span := range ss.GetSpans() {
				fmt.Println(span.GetName())
				fmt.Printf("traceID: %x\n", span.GetTraceId())

				if err := r.Store.Write(serviceName, span); err != nil {
					fmt.Printf("failed to write span: %v\n", err)
				}
			}
		}
	}

	return out, nil
}
