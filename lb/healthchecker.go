package lb

import (
	"net/http"
	"time"
)

func (lb *LoadBalancer) healthCheck() {
	for {
		for i, server := range lb.servers {
			resp, err := http.Get(server.String() + "/health")
			if err != nil || resp.StatusCode != http.StatusOK {
				lb.healthy[i] = false
			} else {
				lb.healthy[i] = true
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
		time.Sleep(30 * time.Second) // Check every 30 seconds
	}
}
