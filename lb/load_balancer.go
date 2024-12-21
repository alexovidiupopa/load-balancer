package lb

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type LoadBalancer struct {
	servers         []*url.URL
	healthy         []bool
	index           uint32
	rateLimiter     *RateLimiter
	circuitBreakers []*CircuitBreaker
}

func NewLoadBalancer(servers []string, rateLimiter *RateLimiter) *LoadBalancer {
	urls := make([]*url.URL, len(servers))
	healthy := make([]bool, len(servers))
	circuitBreakers := make([]*CircuitBreaker, len(servers))
	for i, server := range servers {
		url, err := url.Parse(server)
		if err != nil {
			panic(err)
		}
		urls[i] = url
		healthy[i] = true                                         // Assume all servers are healthy initially
		circuitBreakers[i] = NewCircuitBreaker(3, 10*time.Second) // 3 failures, 10 seconds reset timeout
	}
	lb := &LoadBalancer{servers: urls, healthy: healthy, rateLimiter: rateLimiter, circuitBreakers: circuitBreakers}
	go lb.healthCheck()
	return lb
}

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

func (lb *LoadBalancer) getNextServer() *url.URL {
	for {
		index := atomic.AddUint32(&lb.index, 1)
		server := lb.servers[(int(index)-1)%len(lb.servers)]
		if lb.healthy[(int(index)-1)%len(lb.servers)] {
			return server
		}
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !lb.rateLimiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	logRequest(r)
	start := time.Now()
	target := lb.getNextServer()
	cb := lb.circuitBreakers[(int(lb.index)-1)%len(lb.servers)]
	if !cb.AllowRequest() {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		cb.OnFailure()
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			cb.OnSuccess()
		} else {
			cb.OnFailure()
		}
		return nil
	}
	proxy.ServeHTTP(w, r)
	duration := time.Since(start).Seconds()
	requestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
	requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
}
