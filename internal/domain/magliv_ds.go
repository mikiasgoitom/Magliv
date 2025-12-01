package domain

import (
    "hash/fnv"
    "sort"
)

// Backend holds the information for a single backend server.
type Backend struct {
    ID      string
    Address string
}

// M is the size of the lookup table. The Maglev paper suggests this should be
// a large prime number. For our demo, 65537 is a good choice.
const M = 65537

// PopulateLookupTable generates the Maglev lookup table. This is the heart of the algorithm.
// It takes a slice of active backends and returns a lookup table of size M, where each
// entry contains the ID of a backend.
func PopulateLookupTable(backends []*Backend) []string {
    if len(backends) == 0 {
        return nil
    }

    // Step 1: Generate the permutation table for all backends.
    // This table holds the preferred lookup table slots for each backend.
    permutation := generatePermutationTable(backends)

    // Step 2: Populate the lookup table.
    lookup := make([]int, M)
    for i := range lookup {
        lookup[i] = -1 // Initialize with -1 to mark as empty.
    }

    next := make([]int, len(backends)) // `next` tracks the next slot to try for each backend.
    filledCount := 0

    for filledCount < M {
        for i := range backends {
            // For each backend, find its next preferred, empty slot.
            c := permutation[i][next[i]]
            for lookup[c] >= 0 {
                next[i]++
                c = permutation[i][next[i]]
            }

            // Assign the backend's index (i) to the slot.
            lookup[c] = i
            next[i]++
            filledCount++

            if filledCount == M {
                break
            }
        }
    }

    // Step 3: Convert the lookup table of indices into a table of backend IDs.
    result := make([]string, M)
    for i, backendIndex := range lookup {
        if backendIndex != -1 {
            result[i] = backends[backendIndex].ID
        }
    }

    return result
}

// generatePermutationTable creates the preference list for all backends.
func generatePermutationTable(backends []*Backend) [][]int {
    // Sort backends by ID to ensure deterministic behavior.
    // This is important for consistent hashing when the backend set changes.
    sortedBackends := make([]*Backend, len(backends))
    copy(sortedBackends, backends)
    sort.Slice(sortedBackends, func(i, j int) bool {
        return sortedBackends[i].ID < sortedBackends[j].ID
    })

    table := make([][]int, len(sortedBackends))

    for i, backend := range sortedBackends {
        // As per the paper, we use two different hash functions to generate the permutation.
        // We can simulate this by using FNV-1a with two different seeds (the backend ID itself,
        // and the backend ID plus a separator).
        offset := hash(backend.ID) % M
        skip := (hash(backend.ID+"|") % (M - 1)) + 1

        row := make([]int, M)
        for j := 0; j < M; j++ {
            row[j] = int((offset + uint64(j)*skip) % M)
        }
        table[i] = row
    }
    return table
}

// hash generates a uint64 hash for a given string using FNV-1a.
func hash(s string) uint64 {
    h := fnv.New64a()
    h.Write([]byte(s))
    return h.Sum64()
}