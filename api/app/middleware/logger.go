package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)

		log.Printf(`{"time":"%s", "status":%d, "latency":"%s", "client_ip":"%s", "method":"%s", "path":"%s", "error":"%s"}`+"\n",
			startTime.Format(time.RFC3339),
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}
