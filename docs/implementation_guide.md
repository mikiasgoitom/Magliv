# Magliv Implementation Guide

This document describes the architectural approach and implementation plan for the Magliv demo, following the principles of Clean Architecture.

## 1. System Architecture: Clean Architecture

We will structure the application into distinct layers to ensure separation of concerns. The dependency rule is paramount: **inner layers must not know anything about outer layers**.

- **`Domain` (Innermost)**: Contains the core business logic and entities. This is the heart of Maglev—the hashing algorithm and data structures. It has no dependencies on other layers.
- **`Usecase`**: Orchestrates the flow of data. It uses the `Domain` to perform its tasks. It knows about the `Domain`, but not about the delivery mechanism (like HTTP).
- **`Handler` (Outermost)**: The delivery mechanism. In our case, this is the HTTP reverse proxy that handles incoming requests. It depends on the `Usecase` layer to do the actual work.
- **`main`**: The composition root. It initializes and wires all the layers together.

```
+-----------------------------------------------------------------+
|  main (Composition Root)                                        |
|                                                                 |
|  +-----------------------+      +---------------------------+   |
|  | Handler (HTTP)        |----->| Usecase (LoadBalancer)    |   |
|  | - Reverse Proxy       |      | - GetBackend()            |   |
|  | - Request Keying      |      | - Add/Remove Backend()    |   |
|  +-----------------------+      +---------------------------+   |
|                                             |                   |
|                                             v                   |
|                                 +---------------------------+   |
|                                 | Domain (Maglev Core)      |   |
|                                 | - Backend Entity          |   |
|                                 | - Hashing & Table Gen     |   |
|                                 +---------------------------+   |
|                                                                 |
+-----------------------------------------------------------------+
```

## 2. Phase 1: Core Load Balancer with Terminal Logging

This phase focuses on getting the fundamental Maglev logic working and demonstrating it via clear terminal output.

### Component Implementation

- **`internal/domain/magliv.go`**:

  - `Backend`: A struct holding a backend's `ID` and `Address`.
  - `GeneratePermutation(...)`: The function to generate a backend's preference list.
  - `PopulateLookupTable(...)`: The core function to build the Maglev lookup table.

- **`internal/usecase/magliv_usecase.go`**:

  - `LoadBalancer`: A struct holding the backend list and the lookup table.
  - `NewLoadBalancer(backends)`: Constructor that generates the initial lookup table.
  - `GetBackend(key string)`: Hashes the key and uses the lookup table to select a backend.
  - `UpdateBackends(backends)`: Regenerates the lookup table when backends change.

- **`internal/handler/http/magliv_handler.go`**:

  - `MaglevHandler`: A struct holding a reference to the `usecase.LoadBalancer`.
  - `ServeHTTP(w, r)`:
    1.  Gets a key from the request (e.g., `r.RemoteAddr`).
    2.  Calls the use case to get the chosen backend.
    3.  **Logs the decision to the terminal**: `log.Printf("Request from %s routed to %s", key, backendID)`.
    4.  Uses `httputil.NewSingleHostReverseProxy` to forward the request.

- **`cmd/server/main.go`**:
  1.  Define backend server addresses.
  2.  Start simple HTTP servers for each backend in separate goroutines.
  3.  Initialize the `LoadBalancer` use case.
  4.  Initialize the `MaglevHandler`.
  5.  Start the main load balancer server, routing all `/` traffic to the `MaglevHandler`.

## 3. Phase 2: Adding Live Graphing

This phase adds the visual component on top of the working core from Phase 1.

### New Components & Modifications

- **`internal/usecase/hub.go` (New)**:

  - `Hub`: A struct to manage WebSocket clients (register, unregister, broadcast).
  - `NewHub()`: Creates and runs the hub in a goroutine.

- **`internal/handler/http/websocket_handler.go` (New)**:

  - A new handler to upgrade HTTP connections to WebSockets and register them with the hub.

- **`frontend/` (New Directory)**:
  - `index.html`: Contains the `<canvas>` for the chart and includes Chart.js and `app.js`.
  - `app.js`: Contains the JavaScript to connect to the WebSocket, listen for messages, and update the Chart.js graph.

### Modifications to Existing Files

- **`internal/usecase/magliv_usecase.go`**:

  - The `LoadBalancer` struct will be modified to hold a reference to the `Hub`.
  - The `NewLoadBalancer` constructor will accept the `Hub` as a parameter.

- **`internal/handler/http/magliv_handler.go`**:

  - The `ServeHTTP` method will be modified to call `hub.Broadcast` with the routing decision, in addition to logging it.

- **`cmd/server/main.go`**:
  - Will be updated to:
    1.  Initialize the `Hub` in a goroutine.
    2.  Pass the `Hub` to the `LoadBalancer` use case.
    3.  Add a new route `/ws` for the `websocket_handler`.
    4.  Add a new route `/graph` to serve the `frontend/index.html` file.
