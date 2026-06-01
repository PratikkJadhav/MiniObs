package storage

import "sync"

type location struct {
	fileID uint32
	offset int64
	size   uint32
}

type index struct {
	traces   map[string][]location
	services map[string][]string
	mu       sync.RWMutex
}

func (idx *index) Add(traceID, serviceName string, loc location) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.traces[traceID] = append(idx.traces[traceID], loc)

	found := false
	for _, id := range idx.services[serviceName] {
		if id == traceID {
			found = true
			break
		}
	}

	if !found {
		idx.services[serviceName] = append(idx.services[serviceName], traceID)
	}

}

func (idx *index) GetTrace(traceID string) []location {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.traces[traceID]
}
