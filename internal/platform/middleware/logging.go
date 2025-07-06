package middleware

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/niiniyare/erp/internal/shared/logger"
)

// RequestLogger creates a middleware for logging HTTP requests
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method
        
        // Process request
        c.Next()
        
        // Log request details
        duration := time.Since(start)
        statusCode := c.Writer.Status()
        clientIP := c.ClientIP()
        userAgent := c.Request.UserAgent()
        
        logLevel := logger.InfoLevel
        if statusCode >= 400 && statusCode < 500 {
            logLevel = logger.WarnLevel
        } else if statusCode >= 500 {
            logLevel = logger.ErrorLevel
        }
        
        fields := logger.Fields{
            "method":      method,
            "path":        path,
            "status_code": statusCode,
            "duration_ms": float64(duration.Nanoseconds()) / 1e6,
            "client_ip":   clientIP,
            "user_agent":  userAgent,
        }
        
        // Add error information if present
        if len(c.Errors) > 0 {
            fields["errors"] = c.Errors.String()
        }
        
        message := "HTTP request completed"
        switch logLevel {
        case logger.WarnLevel:
            logger.WarnContext(c.Request.Context(), message, fields)
        case logger.ErrorLevel:
            logger.ErrorContext(c.Request.Context(), message, fields)
        default:
            logger.InfoContext(c.Request.Context(), message, fields)
        }
    }
}
