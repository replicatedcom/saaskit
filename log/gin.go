package log

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func SaaskitLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		logMessage := fmt.Sprintf("[GIN] %3d | %12v | %s | %-7s %s", statusCode, latency, clientIP, method, c.Request.URL.Path)

		switch {
		case statusCode >= 200 && statusCode <= 299:
			fallthrough
		case statusCode >= 300 && statusCode <= 399:
			Infof(logMessage)
		case statusCode >= 400 && statusCode <= 499:
			fallthrough
		default:
			Errorf(logMessage)
		}
	}
}
