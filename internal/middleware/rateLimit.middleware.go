package middleware

import (
	"go-app/global"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimit struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

func NewRateLimitMiddleware() *RateLimit {
	config := global.Config.RateLimit
	interval, _ := time.ParseDuration(config.Interval)              // chuyển từ string thành duration
	limitRate := rate.Every(interval / time.Duration(config.Limit)) // tính toán thời gian nạp token

	m := &RateLimit{
		visitors: make(map[string]*visitor),
		rate:     limitRate,
		burst:    config.Burst,
	}

	go m.cleanupVisitors()
	return m
}

func (m *RateLimit) getVisitor(ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, exists := m.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(m.rate, m.burst)

		m.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}

		return limiter
	}

	v.lastSeen = time.Now()

	return v.limiter
}

func (m *RateLimit) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		m.mu.Lock()

		for ip, v := range m.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(m.visitors, ip)
			}
		}

		m.mu.Unlock()
	}
}

func (m *RateLimit) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := m.getVisitor(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"code":    "TOO_MANY_REQUESTS",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
