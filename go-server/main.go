package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	slogotel "github.com/remychantenay/slog-otel"
	"go.opentelemetry.io/otel"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *slog.Logger

func main() {

	err := SetupLogging()
	if err != nil {
		panic("Failed to setup logging: " + err.Error())
	}
	slog.Info("Server starting")
	for i := 0; i < 1000; i++ {
		log.Println("tests log")
	}
	r := gin.Default()
	r.Use(GinLoggingMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(":9000")
}

func SetupLogging() error {
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, "server.log")

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath, // Log file name (absolute or relative)
		MaxSize:    100,     // megabytes before rotation
		MaxBackups: 7,       // number of rotated backups to keep
		MaxAge:     7,       // days before old logs are deleted
		Compress:   true,    // compress rotated files
	}

	jsonHandler := slog.NewJSONHandler(lumberjackLogger, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	otelHandler := slogotel.OtelHandler{
		Next: jsonHandler,
	}

	logger := slog.New(otelHandler)
	slog.SetDefault(logger)

	return nil
}

func GinLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Start a span for this request
		tracer := otel.Tracer("server")
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()

		// Update request context with span
		c.Request = c.Request.WithContext(ctx)

		// Log with context - trace/span IDs automatically added
		slog.InfoContext(ctx, "Request started",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("client_ip", c.ClientIP()),
		)

		c.Next()

		duration := time.Since(start)

		// Check for errors
		if len(c.Errors) > 0 {
			// Log errors with trace context
			for _, err := range c.Errors {
				slog.ErrorContext(ctx, "Request error",
					slog.String("error", err.Error()),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
				)
			}
			span.RecordError(c.Errors.Last())
		}

		// Log completion with trace context
		slog.InfoContext(ctx, "Request completed",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", duration),
		)
	}
}
