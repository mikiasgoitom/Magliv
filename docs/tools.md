# Tools for Magliv Demo

This document outlines the tools chosen for the Magliv demo implementation. The selection prioritizes simplicity, speed of development, and effectiveness for a clear presentation.

### 1. Go (Golang)

- **Purpose**: The primary programming language for building the load balancer, the backend servers, and the real-time metrics WebSocket.
- **Why**:
  - **Standard Library**: Go's built-in `net/http` package is perfect for creating web servers, the reverse proxy, and WebSocket handlers.
  - **Concurrency**: Goroutines are essential for handling concurrent requests, managing WebSocket connections, and running multiple backend servers simultaneously.

### 2. `hey`

- **Purpose**: A simple, command-line load testing tool.
- **Why**:
  - **Ease of Use**: Generate significant load with a one-line command.
  - **Fast Setup**: Installs with a single `go install` command.
  - **Clear Output**: Provides a simple summary of requests per second, which is perfect for showing the load being applied.
- **Installation**:
  ```shell
  go install github.com/rakyll/hey@latest
  ```

### 3. WebSockets

- **Purpose**: To provide a real-time communication channel between the Go backend and the frontend graph.
- **Why**:
  - **Efficiency**: WebSockets offer a persistent, low-latency connection, allowing the server to push data to the browser instantly without the browser having to poll for updates.
  - **Simplicity**: We will use a well-regarded Go library (like `gorilla/websocket`) to make implementing the WebSocket server straightforward.

### 4. Chart.js

- **Purpose**: A JavaScript library for creating simple, animated, and interactive charts.
- **Why**:
  - **Lightweight**: It's a small library that can be included via a CDN, requiring no local installation.
  - **Easy to Learn**: The API is intuitive, and we can create a live-updating bar chart with just a few lines of JavaScript.
  - **Visually Appealing**: It will provide a clean and professional-looking graph for the presentation.
