package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"load-balancer/lb"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: main <strategy> <weights>")
	}
	strategy, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid strategy")
	}
	weights := []int{}
	for _, w := range os.Args[2:] {
		weight, err := strconv.Atoi(w)
		if err != nil {
			log.Fatal("Invalid weight")
		}
		weights = append(weights, weight)
	}

	servers := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	rateLimiter := lb.NewRateLimiter(10, 1, time.Second)
	loadBalancer := lb.NewLoadBalancer(servers, rateLimiter, lb.Strategy(strategy), weights)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", loadBalancer)

	fmt.Println("Load Balancer started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
