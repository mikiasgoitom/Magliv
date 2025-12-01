package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mikiasgoitom/magliv/internal/domain"
	"github.com/mikiasgoitom/magliv/internal/usecase"
	maglivhttp "github.com/mikiasgoitom/magliv/internal/handler/http"
)

// Define the addresses for our backend servers and the main load balancer.
const (
    loadBalancerAddr = "localhost:8080"
)

var backendAddresses = []string{
    "localhost:8081",
    "localhost:8082",
    "localhost:8083",
}

func main() {
    log.Printf("Starting Magliv Demo...")

    // --- Step 1: Create domain.Backend objects and start backend servers ---
    backends := make([]*domain.Backend, 0, len(backendAddresses))
    for i, addr := range backendAddresses {
        backendID := fmt.Sprintf("Backend-%d", i+1)

        // Create the domain object for the backend.
        backends = append(backends, &domain.Backend{
            ID:      backendID,
            Address: addr,
        })

        // Start a simple HTTP server for each backend in a goroutine.
        go startBackendServer(addr, backendID)
    }

    // --- Step 2: Initialize the use case (the load balancer core) ---
    loadBalancer := usecase.NewLoadBalancer(backends)
    log.Println("Successfully built the initial Maglev lookup table.")

    // --- Step 3: Initialize the HTTP handler ---
    handler := maglivhttp.NewMaglevHandler(loadBalancer)

    // --- Step 4: Start the main load balancer server ---
    server := &http.Server{
        Addr:    loadBalancerAddr,
        Handler: handler,
    }

    log.Printf("Load Balancer is listening on http://%s", loadBalancerAddr)
    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("Failed to start load balancer server: %v", err)
    }
}

// startBackendServer creates and starts a simple HTTP server that responds
// with its own name. This helps us see the load balancing in action.
func startBackendServer(addr, id string) {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] Received request", id)
        fmt.Fprintf(w, "Hello from %s\n", id)
    })

    log.Printf("Backend server '%s' starting on http://%s", id, addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatalf("Failed to start backend server %s: %v", id, err)
    }
}