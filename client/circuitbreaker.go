package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

type CircuitBreakerClient struct {
	breakers map[string]*gobreaker.CircuitBreaker
	client   *http.Client
	mu       sync.Mutex
}

var breakerClient *CircuitBreakerClient

func init() {
	breakerClient = &CircuitBreakerClient{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func GetBreakerClient() *CircuitBreakerClient {
	return breakerClient
}

func (c *CircuitBreakerClient) getBreaker(serviceName string) *gobreaker.CircuitBreaker {
	c.mu.Lock()
	defer c.mu.Unlock()

	if brk, exists := c.breakers[serviceName]; exists {
		return brk
	}

	brk := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: 2,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("[CIRCUIT] %s state changed from %s to %s\n", name, from.String(), to.String())
		},
	})

	c.breakers[serviceName] = brk
	return brk
}

func (c *CircuitBreakerClient) execute(serviceName, method, url string, body interface{}, result interface{}) error {
	breaker := c.getBreaker(serviceName)

	_, err := breaker.Execute(func() (interface{}, error) {
		var reqBody io.Reader
		if body != nil {
			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			reqBody = bytes.NewReader(jsonBytes)
		}

		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		b, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		fmt.Println("DATA ", string(b))

		if err := json.Unmarshal(b, result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}

		return result, nil
	})

	return err
}

// Exposed HTTP method wrappers
func (c *CircuitBreakerClient) Get(serviceName, url string, result interface{}) error {
	return c.execute(serviceName, http.MethodGet, url, nil, result)
}

func (c *CircuitBreakerClient) Post(serviceName, url string, body interface{}, result interface{}) error {
	return c.execute(serviceName, http.MethodPost, url, body, result)
}

func (c *CircuitBreakerClient) Put(serviceName, url string, body interface{}, result interface{}) error {
	return c.execute(serviceName, http.MethodPut, url, body, result)
}

func (c *CircuitBreakerClient) Patch(serviceName, url string, body interface{}, result interface{}) error {
	return c.execute(serviceName, http.MethodPatch, url, body, result)
}

func (c *CircuitBreakerClient) Delete(serviceName, url string, result interface{}) error {
	return c.execute(serviceName, http.MethodDelete, url, nil, result)
}
