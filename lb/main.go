package lb

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	servers := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	rateLimiter := NewRateLimiter(10, 1, time.Second) // 10 tokens max, 1 token per second
	lb := NewLoadBalancer(servers, rateLimiter)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", lb)

	fmt.Println("Load Balancer started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
