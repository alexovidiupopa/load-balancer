package lb

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type LoadBalancer struct {
	servers         []*url.URL
	healthy         []bool
	index           uint32
	rateLimiter     *RateLimiter
	circuitBreakers []*CircuitBreaker
	strategy        Strategy
	weights         []int
}

func NewLoadBalancer(servers []string, rateLimiter *RateLimiter, strategy Strategy, weights []int) *LoadBalancer {
	urls := make([]*url.URL, len(servers))
	healthy := make([]bool, len(servers))
	circuitBreakers := make([]*CircuitBreaker, len(servers))
	for i, server := range servers {
		url, err := url.Parse(server)
		if err != nil {
			panic(err)
		}
		urls[i] = url
		healthy[i] = true
		circuitBreakers[i] = NewCircuitBreaker(3, 10*time.Second)
	}
	lb := &LoadBalancer{servers: urls, healthy: healthy, rateLimiter: rateLimiter, circuitBreakers: circuitBreakers, strategy: strategy, weights: weights}
	go lb.healthCheck()
	return lb
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !lb.rateLimiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	logRequest(r)
	start := time.Now()
	target := lb.getNextServer(r)
	cb := lb.circuitBreakers[(int(lb.index))%len(lb.servers)]
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
