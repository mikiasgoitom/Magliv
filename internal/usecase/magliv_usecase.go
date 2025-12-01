package usecase

import (
	"hash/fnv"
	"sync"

	"github.com/mikiasgoitom/magliv/internal/domain"
)

// LoadBalancer orchestrates the Maglev logic. It holds the state of the
// lookup table and provides thread-safe methods to interact with it.
type LoadBalancer struct {
    mu          sync.RWMutex
    lookupTable []string
    backends    map[string]*domain.Backend
}

// NewLoadBalancer creates and initializes a new LoadBalancer instance.
func NewLoadBalancer(backends []*domain.Backend) *LoadBalancer {
    lb := &LoadBalancer{
        backends: make(map[string]*domain.Backend),
    }
    // Perform the initial population of the lookup table.
    lb.UpdateBackends(backends)
    return lb
}

// GetBackend selects a backend for a given key using the Maglev lookup table.
// It returns the chosen backend or nil if no backends are available.
func (lb *LoadBalancer) GetBackend(key string) *domain.Backend {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    if len(lb.lookupTable) == 0 {
        return nil
    }

    // Hash the key to find the position in the lookup table.
    hash := hash(key)
    index := hash % uint64(len(lb.lookupTable))

    backendID := lb.lookupTable[index]
    return lb.backends[backendID]
}

// UpdateBackends recalculates the lookup table with a new set of backends.
// This is used for the initial setup and for dynamically adding/removing backends.
func (lb *LoadBalancer) UpdateBackends(backends []*domain.Backend) {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    // Regenerate the lookup table using our core domain logic.
    lb.lookupTable = domain.PopulateLookupTable(backends)

    // Update the internal map for quick access to backend details by ID.
    newBackendsMap := make(map[string]*domain.Backend, len(backends))
    for _, b := range backends {
        newBackendsMap[b.ID] = b
    }
    lb.backends = newBackendsMap
}

// hash generates a uint64 hash for a given string using FNV-1a.
func hash(s string) uint64 {
    h := fnv.New64a()
    h.Write([]byte(s))
    return h.Sum64()
}