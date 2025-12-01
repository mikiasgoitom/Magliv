package usecase

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"

	"github.com/mikiasgoitom/magliv/internal/domain"
)

// LoadBalancer orchestrates the Maglev logic.
type LoadBalancer struct {
    mu             sync.RWMutex
    lookupTable    []string
    allBackends    map[string]*domain.Backend // Master list of all possible backends
    activeBackends []*domain.Backend          // Currently active backends
}

// NewLoadBalancer creates a new load balancer instance.
// It takes a list of all possible backends and a list of which ones should be initially active.
func NewLoadBalancer(allBackends []*domain.Backend, initialActiveIDs []string) *LoadBalancer {
    lb := &LoadBalancer{
        allBackends:    make(map[string]*domain.Backend),
        activeBackends: make([]*domain.Backend, 0),
    }

    for _, b := range allBackends {
        lb.allBackends[b.ID] = b
    }

    for _, id := range initialActiveIDs {
        if backend, ok := lb.allBackends[id]; ok {
            lb.activeBackends = append(lb.activeBackends, backend)
        }
    }

    lb.updateLookupTable()
    return lb
}

// GetBackend selects a backend for a given key.
func (lb *LoadBalancer) GetBackend(key string) *domain.Backend {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    if len(lb.lookupTable) == 0 {
        return nil
    }

    hash := hash(key)
    index := hash % uint64(len(lb.lookupTable))

    backendID := lb.lookupTable[index]
    return lb.allBackends[backendID]
}

// updateLookupTable is an internal method to rebuild the lookup table.
func (lb *LoadBalancer) updateLookupTable() {
    lb.lookupTable = domain.PopulateLookupTable(lb.activeBackends)
    sort.Slice(lb.activeBackends, func(i, j int) bool {
        return lb.activeBackends[i].ID < lb.activeBackends[j].ID
    })
}

// DeactivateBackend removes a backend from the active pool.
func (lb *LoadBalancer) DeactivateBackend(id string) error {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    var found bool
    newActiveBackends := make([]*domain.Backend, 0)
    for _, b := range lb.activeBackends {
        if b.ID == id {
            found = true
            continue
        }
        newActiveBackends = append(newActiveBackends, b)
    }

    if !found {
        return fmt.Errorf("backend %s not found or already inactive", id)
    }

    lb.activeBackends = newActiveBackends
    lb.updateLookupTable()
    return nil
}

// ActivateBackend adds a backend to the active pool.
func (lb *LoadBalancer) ActivateBackend(id string) error {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    for _, b := range lb.activeBackends {
        if b.ID == id {
            return fmt.Errorf("backend %s is already active", id)
        }
    }

    backendToAdd, exists := lb.allBackends[id]
    if !exists {
        return fmt.Errorf("backend %s does not exist in the master list", id)
    }

    lb.activeBackends = append(lb.activeBackends, backendToAdd)
    lb.updateLookupTable()
    return nil
}

// GetActiveBackendIDs returns the IDs of the currently active backends.
func (lb *LoadBalancer) GetActiveBackendIDs() []string {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    ids := make([]string, len(lb.activeBackends))
    for i, b := range lb.activeBackends {
        ids[i] = b.ID
    }
    return ids
}

func hash(s string) uint64 {
    h := fnv.New64a()
    h.Write([]byte(s))
    return h.Sum64()
}