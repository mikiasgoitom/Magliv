package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mikiasgoitom/magliv/internal/usecase"
)

// AdminHandler handles administrative tasks like activating/deactivating backends.
type AdminHandler struct {
    LoadBalancer *usecase.LoadBalancer
    Hub          *Hub
}

func NewAdminHandler(lb *usecase.LoadBalancer, hub *Hub) *AdminHandler {
    return &AdminHandler{
        LoadBalancer: lb,
        Hub:          hub,
    }
}

func (h *AdminHandler) Activate(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
        return
    }

    if err := h.LoadBalancer.ActivateBackend(id); err != nil {
        http.Error(w, err.Error(), http.StatusConflict)
        return
    }

    h.Hub.UpdateBackendIDs(h.LoadBalancer.GetActiveBackendIDs())
    log.Printf("[ADMIN] Activated backend: %s", id)
    fmt.Fprintf(w, "Activated backend %s.", id)
}

func (h *AdminHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
        return
    }

    if err := h.LoadBalancer.DeactivateBackend(id); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    h.Hub.UpdateBackendIDs(h.LoadBalancer.GetActiveBackendIDs())
    log.Printf("[ADMIN] Deactivated backend: %s", id)
    fmt.Fprintf(w, "Deactivated backend %s.", id)
}