package kf

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Handler manages graceful shutdown
type Handler struct {
	shutdownFuncs []func() error
	timeout       time.Duration
}

// NewHandler creates a new shutdown handler
func NewHandler(timeout time.Duration) *Handler {
	return &Handler{
		shutdownFuncs: make([]func() error, 0),
		timeout:       timeout,
	}
}

// AddShutdownFunc adds a function to be called during shutdown
func (h *Handler) AddShutdownFunc(fn func() error) {
	h.shutdownFuncs = append(h.shutdownFuncs, fn)
}

// WaitForShutdown waits for shutdown signal and executes shutdown functions
func (h *Handler) WaitForShutdown() {
	// Create channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("🔄 Received shutdown signal: %v", sig)

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	// Execute shutdown functions
	h.executeShutdown(ctx)

	log.Println("👋 Application shutdown complete")
}

// executeShutdown executes all shutdown functions with timeout
func (h *Handler) executeShutdown(ctx context.Context) {
	done := make(chan bool, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ Panic during shutdown: %v", r)
			}
			done <- true
		}()

		for i, fn := range h.shutdownFuncs {
			if err := fn(); err != nil {
				log.Printf("❌ Error in shutdown function %d: %v", i, err)
			}
		}
	}()

	select {
	case <-done:
		log.Println("✅ Graceful shutdown completed")
	case <-ctx.Done():
		log.Println("⏰ Shutdown timeout exceeded, forcing exit")
	}
}

// HandlePanic handles panics and ensures graceful shutdown
func HandlePanic() {
	if r := recover(); r != nil {
		log.Printf("❌ Panic occurred: %v", r)
		// You can add additional panic handling logic here
		// like sending alerts, logging to external systems, etc.
	}
}
