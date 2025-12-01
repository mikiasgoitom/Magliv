# Tools for Magliv Demo

This document outlines the tools used for the Magliv demo implementation. The selection prioritizes simplicity, speed of development, and effectiveness for a clear presentation.

### 1. Go (Golang)

*   **Purpose**: The primary programming language for building the load balancer, the backend servers, and the real-time metrics WebSocket.
*   **Why**: Go's strong standard library (`net/http`), built-in concurrency (goroutines), and performance make it ideal for this networking-focused application.

### 2. `hey`

*   **Purpose**: A simple, command-line load testing tool.
*   **Why**: It allows for easy generation of HTTP load to demonstrate the load balancer's distribution and resilience characteristics in real-time.
*   **Installation**: `go install github.com/rakyll/hey@latest`

### 3. WebSockets (`gorilla/websocket`)

*   **Purpose**: To provide a real-time, bidirectional communication channel between the Go backend and the frontend dashboard.
*   **Why**: WebSockets allow the server to push routing updates to the dashboard instantly, enabling the live graph.

### 4. Chart.js

*   **Purpose**: A JavaScript library for creating the live bar chart on the frontend dashboard.
*   **Why**: It is lightweight, easy to use, and creates a clean, professional-looking visualization of the request distribution.