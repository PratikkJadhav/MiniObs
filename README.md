# MiniObs

A minimal observability backend built from scratch in Go. Accepts real OpenTelemetry traces over gRPC, stores them on disk using a custom storage engine, and exposes a query API with a web UI. No ClickHouse, no external databases, no frameworks — pure Go.

Built to understand observability internals deeply and as a portfolio project for outreach to infra/observability startups.

---

## Architecture

```
Instrumented Go App
        │
        │ OTLP/gRPC (port 4317)
        ▼
┌─────────────────┐
│  gRPC Receiver  │   implements TraceServiceServer
│  receiver/      │   extracts spans from ExportTraceServiceRequest
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Storage Engine  │   append-only segment files + in-memory index
│ storage/        │   length-prefix framing: [4B length][N bytes protobuf]
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Query Layer   │   p50/p95/p99 latency, error rate per service
│   query/        │   computed directly from stored spans
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   HTTP API      │   Chi router, JSON responses (port 8080)
│   api/          │   + static web UI
└─────────────────┘
```

---

## Storage Design

No external database. Everything is custom.

**Segment files** — spans are appended sequentially to binary files:

```
seg-000001.dat  [4B len][span bytes][4B len][span bytes]...
seg-000002.dat  ← rotated at 64MB
```

**In-memory index** — built on startup, updated on every write:

```
traceID     → []Location{fileID, offset, size}
serviceName → []traceID
```

**Hint files** — index persisted to disk on shutdown, loaded on restart:

```
hint.dat: traceID|fileID|offset|size|serviceName
```

On restart, the index is rebuilt from `hint.dat` in O(n) without scanning segment files.

---

## HTTP API

```
GET /api/services                         list all seen services
GET /api/traces?service=<name>            list traces for a service
GET /api/traces/:traceID                  all spans for a trace
GET /api/metrics?service=<name>           p50/p95/p99 latency + error rate
```

---

## Web UI

Served at `http://localhost:8080`. Shows live trace feed, latency distribution chart, duration histogram, and per-service metrics.


---

## How to Run

**Prerequisites:** Go 1.21+

```bash
# clone
git clone https://github.com/PratikkJadhav/MiniObs
cd MiniObs

# install dependencies
go mod tidy

# start the server (gRPC on :4317, HTTP on :8080)
go run cmd/main.go

# in another terminal, send test spans
go run cmd/testclient/main.go

# open the UI
open http://localhost:8080
```

**Query via curl:**

```bash
# list services
curl http://localhost:8080/api/services

# list traces
curl "http://localhost:8080/api/traces?service=miniobs-server"

# get latency metrics
curl "http://localhost:8080/api/metrics?service=miniobs-server"
```



## Project Structure

```
miniobs/
├── cmd/
│   ├── main.go              entry point — wires gRPC + HTTP servers
│   └── testclient/main.go   sends test OTLP spans
├── receiver/
│   └── traces.go            gRPC server implementing TraceServiceServer
├── storage/
│   ├── segment.go           append-only file writer + reader
│   ├── index.go             in-memory index + hint file persistence
│   └── store.go             public API: Write(), Read(), GetTraceSummaries()
├── query/
│   └── traces.go            p50/p95/p99 latency, error rate computation
├── api/
│   └── http.go              Chi HTTP handlers, CORS middleware
└── ui/
    └── index.html           single-file web UI (vanilla JS + Chart.js)
```



