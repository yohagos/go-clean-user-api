package middleware

import (
	"sync"
	"time"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	duration time.Duration
}

type visitor struct {
	lastSeen time.Time
	count    int
	mu       sync.Mutex
}

func NewRateLimiter(rate int, duration time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		duration: duration,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) Allow(ip string) bool {
	rl.mu.RLock()
	v, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		v = &visitor{lastSeen: time.Now(), count: 0}
		rl.visitors[ip] = v
		rl.mu.Unlock()
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	if now.Sub(v.lastSeen) > rl.duration {
		v.count = 0
		v.lastSeen = now
	}

	v.count++
	return v.count <= rl.rate
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if time.Since(v.lastSeen) > rl.duration {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
