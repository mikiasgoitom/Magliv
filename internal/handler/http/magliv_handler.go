package http

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/mikiasgoitom/magliv/internal/usecase"
)

// MaglevHandler is the HTTP handler that uses the Maglev use case to proxy requests.
type MaglevHandler struct {
    LoadBalancer *usecase.LoadBalancer
    Hub          *Hub
}

// NewMaglevHandler creates a new instance of the MaglevHandler.
func NewMaglevHandler(lb *usecase.LoadBalancer, hub *Hub) *MaglevHandler {
    return &MaglevHandler{
        LoadBalancer: lb,
        Hub:          hub,
    }
}

// ServeHTTP is the entry point for incoming requests. It selects a backend
// and proxies the request to it.
func (h *MaglevHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Use the client's remote address as the key for consistent hashing.
    key := r.RemoteAddr

    // Get the backend from our use case.
    backend := h.LoadBalancer.GetBackend(key)
    if backend == nil {
        log.Printf("[ERROR] No available backends for request from %s", key)
        http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
        return
    }

    // Log the decision for our demo.
    log.Printf("[INFO] Request from %s -> %s (%s)", key, backend.ID, backend.Address)

    // Broadcast the request to the WebSocket clients.
    msg := map[string]string{
        "type":      "update",
        "backendId": backend.ID,
    }
    jsonMsg, err := json.Marshal(msg)
    if err == nil {
        h.Hub.Broadcast(jsonMsg)
    }


    // Parse the backend address to create the reverse proxy.
    targetUrl, err := url.Parse("http://" + backend.Address)
    if err != nil {
        log.Printf("[ERROR] Failed to parse backend URL %s: %v", backend.Address, err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Create a reverse proxy and serve the request.
    // httputil.NewSingleHostReverseProxy handles all the details of
    // forwarding the request and copying the response back.
    proxy := httputil.NewSingleHostReverseProxy(targetUrl)
    proxy.ServeHTTP(w, r)
}