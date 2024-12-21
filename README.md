# Load Balancer with Rate Limiting, Circuit Breaker, and Prometheus Metrics

_Disclaimer: this project is a work in progress. I will be updating the readme as new features will be implemented._

This project implements a load balancer in Go with the following features:
- Request timeouts for backend servers
- Rate limiting using a token bucket algorithm
- Circuit breaker pattern for backend servers
- Prometheus metrics for monitoring

## Project Structure

- `main.go`: Contains the `main` function and initializes the load balancer.
- `load_balancer.go`: Contains the `LoadBalancer` struct and its methods.
- `rate_limiter.go`: Contains the `RateLimiter` struct and its methods.
- `circuit_breaker.go`: Contains the `CircuitBreaker` struct and its methods.
- `logging.go`: Contains the logging function.
- `metrics.go`: Contains the Prometheus metrics definitions and registration.

## Getting Started

### Prerequisites

- Go 1.16 or later
- Prometheus

### Installing

1. Clone the repository:
   ```sh
   git clone https://github.com/alexovidiupopa/load-balancer.git
   cd load-balancer
   ```

2. Install dependencies:
   ```sh
   go mod tidy
   ```

### Running the Load Balancer

1. Start the backend servers (e.g., on ports 8081, 8082, 8083).
```shell
go run backend/main.go 8081
go run backend/main.go 8082
go run backend/main.go 8083
```
2. Run the load balancer:
   ```sh
   go run lb/main.go
   ```

3. The load balancer will start on port 8080.

### Prometheus Metrics

- The metrics are exposed at `/metrics` endpoint.
- Configure Prometheus to scrape the metrics from the load balancer.

## Features

### Request Timeouts

The load balancer uses a custom `http.Transport` with timeouts for dialing and TLS handshake.

### Rate Limiting

The `RateLimiter` struct implements a token bucket algorithm to limit the rate of incoming requests.

### Circuit Breaker

The `CircuitBreaker` struct manages the state of each backend server (closed, open, half-open) to prevent sending requests to unhealthy servers.

### Prometheus Metrics

The load balancer tracks the total number of requests and the duration of each request using Prometheus metrics.
