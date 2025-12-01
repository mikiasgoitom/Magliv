# Magliv Implementation Guide

This document describes the final architecture and implementation of the Magliv demo, following the principles of Clean Architecture.

## 1. System Architecture

The application is structured into distinct layers to ensure a clear separation of concerns. The core principle is that inner layers are independent of outer layers.

*   **`Domain`**: Contains the core business logic—the Maglev hashing algorithm and the `Backend` entity. It has no external dependencies.
*   **`Usecase`**: Orchestrates the domain logic. It contains the `LoadBalancer` which manages the state of active and inactive backends.
*   **`Handler`**: The outermost layer, responsible for communication. This includes the `MaglevHandler` (reverse proxy), `AdminHandler` (backend activation/deactivation), and the `Hub` (WebSocket communication).
*   **`main`**: The composition root that initializes and wires all the layers together.

```
+-------------------------------------------------------------------------+
|  main (Composition Root)                                                |
|                                                                         |
|  +-----------------+   +------------------+   +-----------------------+ |
|  | AdminHandler    |   | MaglevHandler    |   | Hub (WebSocket)       | |
|  | - /admin/act... |   | - / (Proxy)      |   | - /ws                 | |
|  +-----------------+   +------------------+   +-----------------------+ |
|          |                   |                        |                 |
|          +-------------------+------------------------+                 |
|                              |                                          |
|                              v                                          |
|                  +---------------------------+                          |
|                  | Usecase (LoadBalancer)    |                          |
|                  | - Activate/Deactivate     |                          |
|                  | - GetBackend()            |                          |
|                  +---------------------------+                          |
|                              |                                          |
|                              v                                          |
|                  +---------------------------+                          |
|                  | Domain (Maglev Core)      |                          |
|                  | - Hashing & Table Gen     |                          |
|                  +---------------------------+                          |
|                                                                         |
+-------------------------------------------------------------------------+
```

## 2. Component Breakdown

### Core Logic

*   **`internal/domain/magliv_ds.go`**: Defines the `Backend` entity and contains the `PopulateLookupTable` function, which is the pure implementation of the Maglev hashing algorithm.
*   **`internal/usecase/magliv_usecase.go`**: Defines the `LoadBalancer`, which holds the master list of all 10 backends and a list of currently active ones. It contains the logic to `ActivateBackend`, `DeactivateBackend`, and `GetBackend` for a given request key.

### Handlers

*   **`internal/handler/http/magliv_handler.go`**: The main reverse proxy. For every incoming request, it calls the `LoadBalancer` use case to select an active backend, forwards the request, and sends an "update" message to the `Hub`.
*   **`internal/handler/http/admin_handler.go`**: Exposes the `/admin/activate` and `/admin/deactivate` endpoints. It calls the corresponding methods on the `LoadBalancer` use case and instructs the `Hub` to notify dashboards of the state change.
*   **`internal/handler/http/hub.go`**: Manages all WebSocket connections. It registers new clients, unregisters them on disconnect, and broadcasts messages. When a client connects or the backend state changes, it sends an `init` message to reset the dashboards.

### Frontend

*   **`frontend/index.html`**: The live dashboard. It uses Chart.js to render a bar chart and connects to the `/ws` endpoint. It listens for `init` messages to set up the chart and `update` messages to increment the request counts for each backend.
*   **`frontend/admin.html`**: The admin control panel. It provides a UI with `+` and `-` buttons to hit the `/admin/activate` and `/admin/deactivate` endpoints, allowing for live demonstration of Maglev's resilience.

### Entrypoint

*   **`cmd/server/main.go`**:
    1.  Initializes and starts all 10 backend HTTP servers in the background.
    2.  Initializes the `LoadBalancer` use case, telling it which backends are initially active.
    3.  Initializes the `Hub`, `MaglevHandler`, and `AdminHandler`.
    4.  Sets up all the routes (`/`, `/dashboard`, `/admin`, `/ws`, etc.).
    5.  Starts the main HTTP server.