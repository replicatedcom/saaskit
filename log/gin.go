package log

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type StatusRange struct {
	Min int
	Max int
}

func (sr *StatusRange) Includes(n int) bool {
	return n >= sr.Min && n <= sr.Max
}

func SaaskitLogger(statusBlacklistRanges []StatusRange) gin.HandlerFunc {
	return func(c *gin.Context) {
		bl := statusBlacklistRanges

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		for _, r := range bl {
			if r.Includes(statusCode) {
				return
			}
		}

		logMessage := fmt.Sprintf("[GIN] %3d | %12v | %s | %-7s %s", statusCode, latency, clientIP, method, c.Request.URL.Path)

		switch {
		case statusCode >= 200 && statusCode <= 299:
			fallthrough
		case statusCode >= 300 && statusCode <= 399:
			Debugf(logMessage)
		case statusCode >= 400 && statusCode <= 499:
			fallthrough
		default:
			Errorf(logMessage)
		}
	}
}
