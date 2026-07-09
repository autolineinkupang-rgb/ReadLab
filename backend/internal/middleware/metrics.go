package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	metricsMu    sync.RWMutex
	requestCount = map[string]int64{}
	requestDur   = map[string]float64{}
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		method := c.Request.Method
		path := normalizePath(c.FullPath())
		key := fmt.Sprintf("%s %s %d", method, path, status)

		duration := time.Since(start).Seconds()

		metricsMu.Lock()
		requestCount[key]++
		requestDur[key] += duration
		metricsMu.Unlock()
	}
}

func normalizePath(fullPath string) string {
	if fullPath == "" {
		return "/unknown"
	}
	return fullPath
}

func MetricsHandler(c *gin.Context) {
	metricsMu.RLock()
	defer metricsMu.RUnlock()

	var sb strings.Builder
	sb.WriteString("# HELP readlab_requests_total Total request count\n")
	sb.WriteString("# TYPE readlab_requests_total counter\n")

	for key, count := range requestCount {
		parts := strings.SplitN(key, " ", 3)
		if len(parts) == 3 {
			method := parts[0]
			path := parts[1]
			status := parts[2]
			sb.WriteString(fmt.Sprintf("readlab_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				method, path, status, count))
		}
	}

	sb.WriteString("\n# HELP readlab_request_duration_seconds Total request duration in seconds\n")
	sb.WriteString("# TYPE readlab_request_duration_seconds counter\n")

	for key, dur := range requestDur {
		parts := strings.SplitN(key, " ", 3)
		if len(parts) == 3 {
			method := parts[0]
			path := parts[1]
			status := parts[2]
			sb.WriteString(fmt.Sprintf("readlab_request_duration_seconds{method=\"%s\",path=\"%s\",status=\"%s\"} %s\n",
				method, path, status, strconv.FormatFloat(dur, 'f', 6, 64)))
		}
	}

	c.String(200, sb.String())
}
