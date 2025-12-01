package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/mikiasgoitom/magliv/internal/domain"
    maglivhttp "github.com/mikiasgoitom/magliv/internal/handler/http"
    "github.com/mikiasgoitom/magliv/internal/usecase"
)

const loadBalancerAddr = "localhost:8080"

func main() {
    log.Printf("Starting Magliv Demo...")

    // --- Step 1: Start all 10 backend servers ---
    allBackends := make([]*domain.Backend, 0, 10)
    for i := 0; i < 10; i++ {
        id := fmt.Sprintf("Backend-%d", i+1)
        addr := fmt.Sprintf("localhost:%d", 8081+i)
        allBackends = append(allBackends, &domain.Backend{ID: id, Address: addr})
        go startBackendServer(addr, id)
    }

    // --- Step 2: Initialize Usecase, Hub, and Handlers ---
    // Start with the first 3 backends active.
    initialActiveIDs := []string{"Backend-1", "Backend-2", "Backend-3"}
    loadBalancer := usecase.NewLoadBalancer(allBackends, initialActiveIDs)

    hub := maglivhttp.NewHub(loadBalancer.GetActiveBackendIDs())
    go hub.Run()

    maglevHandler := maglivhttp.NewMaglevHandler(loadBalancer, hub)
    adminHandler := maglivhttp.NewAdminHandler(loadBalancer, hub)

    // --- Step 3: Setup server routes ---
    mux := http.NewServeMux()
    mux.Handle("/", maglevHandler)
    mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "frontend/index.html")
    })
    mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "frontend/admin.html")
    })
    mux.HandleFunc("/ws", hub.ServeWs)
    mux.HandleFunc("/admin/activate", adminHandler.Activate)
    mux.HandleFunc("/admin/deactivate", adminHandler.Deactivate)

    // --- Step 4: Start the server ---
    server := &http.Server{Addr: loadBalancerAddr, Handler: mux}
    log.Printf("Load Balancer is listening on http://%s", loadBalancerAddr)
    log.Printf("Dashboard available at http://%s/dashboard", loadBalancerAddr)
    log.Printf("Admin panel available at http://%s/admin", loadBalancerAddr)
    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("Failed to start load balancer server: %v", err)
    }
}

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