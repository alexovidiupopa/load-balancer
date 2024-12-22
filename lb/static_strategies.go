package lb

import (
	"net/http"
	"net/url"
	"sync/atomic"
)

func (lb *LoadBalancer) getNextServer(r *http.Request) *url.URL {
	switch lb.strategy {
	case RoundRobin:
		return lb.roundRobin()
	case WeightedRoundRobin:
		return lb.weightedRoundRobin()
	case IPHash:
		return lb.ipHash(r)
	default:
		return lb.roundRobin()
	}
}

func (lb *LoadBalancer) roundRobin() *url.URL {
	for {
		index := atomic.AddUint32(&lb.index, 1)
		server := lb.servers[(int(index)-1)%len(lb.servers)]
		if lb.healthy[(int(index)-1)%len(lb.servers)] {
			return server
		}
	}
}

func (lb *LoadBalancer) weightedRoundRobin() *url.URL {
	totalWeight := 0
	for _, weight := range lb.weights {
		totalWeight += weight
	}
	for {
		index := atomic.AddUint32(&lb.index, 1)
		serverIndex := (int(index) - 1) % totalWeight
		for i, weight := range lb.weights {
			if serverIndex < weight {
				if lb.healthy[i] {
					return lb.servers[i]
				}
				break
			}
			serverIndex -= weight
		}
	}
}

func (lb *LoadBalancer) ipHash(r *http.Request) *url.URL {
	clientIP := r.RemoteAddr
	hashValue := hash(clientIP)
	serverIndex := int(hashValue) % len(lb.servers)
	for !lb.healthy[serverIndex] {
		serverIndex = (serverIndex + 1) % len(lb.servers)
	}
	return lb.servers[serverIndex]
}
