# Magliv

This project is a demonstration of a user-space network load balancer inspired by Google's Maglev paper. It is implemented in Go, follows Clean Architecture principles, and features a live dashboard to visualize the load distribution in real-time.

The primary goal is to demonstrate Maglev's key feature: **consistent hashing with minimal disruption**. When a backend server is added or removed, only a small fraction of connections are re-routed, while the vast majority remain mapped to their original backends.

## Features

- **Maglev Consistent Hashing**: Core implementation of the Maglev hashing algorithm for backend selection.
- **HTTP Reverse Proxy**: A functional load balancer that distributes incoming HTTP requests across a set of active backend servers.
- **Live Dashboard**: A web-based dashboard (`/dashboard`) that visualizes the distribution of requests across all backends in real-time using WebSockets and Chart.js.
- **Dynamic Admin Panel**: A simple admin interface (`/admin`) to activate and deactivate backend servers on the fly, allowing for a live demonstration of Maglev's resilience.

## Architecture

The application is built using a simplified **Clean Architecture** approach to ensure a clear separation of concerns.

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

- **Domain**: Contains the pure, dependency-free logic of the Maglev algorithm.
- **Usecase**: Orchestrates the domain logic, managing the state of active vs. inactive backends.
- **Handler**: The outermost layer that handles all communication (HTTP requests, WebSocket connections).
- **Main**: The entry point that initializes and wires all the components together.

## Getting Started

### Prerequisites

- **Go**: Version 1.18 or higher.
- **`hey`**: A command-line load testing tool. Install it with:
  ```shell
  go install github.com/rakyll/hey@latest
  ```

### Running the Application

1.  **Clone the repository** (if you haven't already).

2.  **Install dependencies**:

    ```shell
    go mod tidy
    ```

3.  **Run the server**:

    ```shell
    go run ./cmd/server/main.go
    ```

    You will see output indicating that the servers are running:

    ```
    Load Balancer is listening on http://localhost:8080
    Dashboard available at http://localhost:8080/dashboard
    Admin panel available at http://localhost:8080/admin
    ```

## How to Use and Demonstrate

This demo is best experienced with three browser tabs and a terminal window.

1.  **Open the Live Dashboard**:

    - Navigate to `http://localhost:8080/dashboard`
    - You will see a bar chart with the initially active backends (Backend-1, Backend-2, Backend-3).

2.  **Open the Admin Panel**:

    - Navigate to `http://localhost:8080/admin`
    - This panel lists all 10 available backends with `+` (Activate) and `-` (Deactivate) buttons.

3.  **Generate Load**:

    - In your terminal, run `hey` to send a continuous stream of requests to the load balancer.
    - ```shell
      hey -z 30s -c 50 http://localhost:8080
      ```

4.  **Observe the Distribution**:
    - Switch back to the **Dashboard** tab. You will see the bars for the active backends growing in real-time as they receive requests.

### Demonstrating Resilience (The "Aha!" Moment)

This is the most important part of the demo.

1.  While the load test (`hey`) is still running, go to the **Admin Panel** tab.
2.  Click the **`-`** button next to **Backend-2**.
3.  Quickly switch back to the **Dashboard** tab.

You will observe the following behavior, which demonstrates Maglev's core strength:

- The bar for **Backend-2** will immediately stop growing.
- The bars for **Backend-1** and **Backend-3** will start growing faster, as they have absorbed the traffic that was previously going to Backend-2.
- Crucially, the _total number of requests_ that were re-routed is minimized.

You can then go back to the admin panel and use the **`+`** button to reactivate Backend-2 (or activate a new one like Backend-4) and watch it start receiving traffic again.
