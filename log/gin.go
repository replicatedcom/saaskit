package log

import (
	"fmt"
	"io"
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
		Debugf(logMessage)
	}
}

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

// GinLoggerWithWriter instance a Logger middleware with the specified writter buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func GinLoggerWithWriter(out io.Writer, notlogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			// Stop timer
			end := time.Now()
			latency := end.Sub(start)

			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()
			statusColor := colorForStatus(statusCode)
			methodColor := colorForMethod(method)
			comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

			fmt.Fprintf(out, "[GIN] %v |%s %3d %s| %13v | %s |%s  %s %-7s %s\n%s",
				end.Format("2006/01/02 - 15:04:05"),
				statusColor, statusCode, reset,
				latency,
				clientIP,
				methodColor, reset, method,
				path,
				comment,
			)
		}
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}
