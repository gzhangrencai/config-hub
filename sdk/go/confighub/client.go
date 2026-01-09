// Package confighub provides a Go SDK for ConfigHub configuration management.
package confighub

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var (
	ErrNotFound       = errors.New("config not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidConfig  = errors.New("invalid configuration")
	ErrWatchTimeout   = errors.New("watch timeout")
	ErrClientClosed   = errors.New("client closed")
)

// Config represents a configuration item
type Config struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Environment string `json:"environment"`
	Version     int    `json:"version"`
	Content     string `json:"content"`
}

// ClientOptions configures the ConfigHub client
type ClientOptions struct {
	// ServerURL is the ConfigHub server URL (required)
	ServerURL string

	// AccessKey is the API access key (required)
	AccessKey string

	// SecretKey is the API secret key (required)
	SecretKey string

	// Namespace is the default namespace (default: "application")
	Namespace string

	// Environment is the default environment (default: "default")
	Environment string

	// WatchTimeout is the long-polling timeout in seconds (default: 30)
	WatchTimeout int

	// HTTPClient is a custom HTTP client (optional)
	HTTPClient *http.Client

	// OnChange is called when configuration changes
	OnChange func(config *Config)

	// OnError is called when an error occurs during watch
	OnError func(err error)
}

// Client is the ConfigHub SDK client
type Client struct {
	opts       *ClientOptions
	httpClient *http.Client
	cache      map[string]*Config
	cacheMu    sync.RWMutex
	watching   bool
	watchMu    sync.Mutex
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// NewClient creates a new ConfigHub client
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts.ServerURL == "" {
		return nil, errors.New("server URL is required")
	}
	if opts.AccessKey == "" || opts.SecretKey == "" {
		return nil, errors.New("access key and secret key are required")
	}

	if opts.Namespace == "" {
		opts.Namespace = "application"
	}
	if opts.Environment == "" {
		opts.Environment = "default"
	}
	if opts.WatchTimeout <= 0 {
		opts.WatchTimeout = 30
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Duration(opts.WatchTimeout+10) * time.Second,
		}
	}

	return &Client{
		opts:       opts,
		httpClient: httpClient,
		cache:      make(map[string]*Config),
		stopCh:     make(chan struct{}),
	}, nil
}


// Get fetches a configuration by name
func (c *Client) Get(ctx context.Context, name string) (*Config, error) {
	return c.GetWithOptions(ctx, name, c.opts.Namespace, c.opts.Environment)
}

// GetWithOptions fetches a configuration with custom namespace and environment
func (c *Client) GetWithOptions(ctx context.Context, name, namespace, env string) (*Config, error) {
	// Check cache first
	cacheKey := c.cacheKey(name, namespace, env)
	c.cacheMu.RLock()
	if cached, ok := c.cache[cacheKey]; ok {
		c.cacheMu.RUnlock()
		return cached, nil
	}
	c.cacheMu.RUnlock()

	// Fetch from server
	config, err := c.fetchConfig(ctx, name, namespace, env, 0)
	if err != nil {
		return nil, err
	}

	// Update cache
	c.cacheMu.Lock()
	c.cache[cacheKey] = config
	c.cacheMu.Unlock()

	return config, nil
}

// GetString returns the configuration content as string
func (c *Client) GetString(ctx context.Context, name string) (string, error) {
	config, err := c.Get(ctx, name)
	if err != nil {
		return "", err
	}
	return config.Content, nil
}

// GetJSON unmarshals the configuration content into the provided interface
func (c *Client) GetJSON(ctx context.Context, name string, v interface{}) error {
	config, err := c.Get(ctx, name)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(config.Content), v)
}

// Refresh forces a refresh of the cached configuration
func (c *Client) Refresh(ctx context.Context, name string) (*Config, error) {
	cacheKey := c.cacheKey(name, c.opts.Namespace, c.opts.Environment)
	
	config, err := c.fetchConfig(ctx, name, c.opts.Namespace, c.opts.Environment, 0)
	if err != nil {
		return nil, err
	}

	c.cacheMu.Lock()
	c.cache[cacheKey] = config
	c.cacheMu.Unlock()

	return config, nil
}

// fetchConfig fetches configuration from the server
func (c *Client) fetchConfig(ctx context.Context, name, namespace, env string, currentVersion int) (*Config, error) {
	u, err := url.Parse(c.opts.ServerURL)
	if err != nil {
		return nil, err
	}
	u.Path = "/api/v1/config"

	q := u.Query()
	q.Set("name", name)
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	if env != "" {
		q.Set("env", env)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	c.signRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var config Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// signRequest adds authentication headers to the request
func (c *Client) signRequest(req *http.Request) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	
	// Create signature: HMAC-SHA256(timestamp + method + path, secretKey)
	message := timestamp + req.Method + req.URL.Path
	if req.URL.RawQuery != "" {
		message += "?" + req.URL.RawQuery
	}
	
	h := hmac.New(sha256.New, []byte(c.opts.SecretKey))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("X-Access-Key", c.opts.AccessKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)
}

// cacheKey generates a cache key for the configuration
func (c *Client) cacheKey(name, namespace, env string) string {
	return fmt.Sprintf("%s:%s:%s", namespace, env, name)
}


// Watch starts watching for configuration changes
func (c *Client) Watch(ctx context.Context, names ...string) error {
	c.watchMu.Lock()
	if c.watching {
		c.watchMu.Unlock()
		return errors.New("already watching")
	}
	c.watching = true
	c.watchMu.Unlock()

	// Initial fetch for all configs
	for _, name := range names {
		if _, err := c.Get(ctx, name); err != nil {
			return fmt.Errorf("failed to fetch initial config %s: %w", name, err)
		}
	}

	// Start watch goroutines
	for _, name := range names {
		c.wg.Add(1)
		go c.watchConfig(name)
	}

	return nil
}

// watchConfig watches a single configuration for changes
func (c *Client) watchConfig(name string) {
	defer c.wg.Done()

	namespace := c.opts.Namespace
	env := c.opts.Environment
	cacheKey := c.cacheKey(name, namespace, env)

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		// Get current version from cache
		c.cacheMu.RLock()
		currentVersion := 0
		if cached, ok := c.cache[cacheKey]; ok {
			currentVersion = cached.Version
		}
		c.cacheMu.RUnlock()

		// Long-poll for changes
		config, err := c.watchOnce(name, namespace, env, currentVersion)
		if err != nil {
			if err == ErrWatchTimeout {
				continue // Normal timeout, retry
			}
			if c.opts.OnError != nil {
				c.opts.OnError(err)
			}
			time.Sleep(5 * time.Second) // Backoff on error
			continue
		}

		if config != nil && config.Version > currentVersion {
			// Update cache
			c.cacheMu.Lock()
			c.cache[cacheKey] = config
			c.cacheMu.Unlock()

			// Notify callback
			if c.opts.OnChange != nil {
				c.opts.OnChange(config)
			}
		}
	}
}

// watchOnce performs a single long-poll request
func (c *Client) watchOnce(name, namespace, env string, currentVersion int) (*Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.opts.WatchTimeout+5)*time.Second)
	defer cancel()

	u, err := url.Parse(c.opts.ServerURL)
	if err != nil {
		return nil, err
	}
	u.Path = "/api/v1/config/watch"

	q := u.Query()
	q.Set("name", name)
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	if env != "" {
		q.Set("env", env)
	}
	q.Set("version", strconv.Itoa(currentVersion))
	q.Set("timeout", strconv.Itoa(c.opts.WatchTimeout))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	c.signRequest(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, ErrWatchTimeout
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("watch error: %s", string(body))
	}

	var result struct {
		Changed     bool   `json:"changed"`
		Name        string `json:"name"`
		Namespace   string `json:"namespace"`
		Environment string `json:"environment"`
		Version     int    `json:"version"`
		Content     string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Changed {
		return nil, ErrWatchTimeout
	}

	return &Config{
		Name:        result.Name,
		Namespace:   result.Namespace,
		Environment: result.Environment,
		Version:     result.Version,
		Content:     result.Content,
	}, nil
}

// StopWatch stops watching for configuration changes
func (c *Client) StopWatch() {
	c.watchMu.Lock()
	if !c.watching {
		c.watchMu.Unlock()
		return
	}
	c.watching = false
	c.watchMu.Unlock()

	close(c.stopCh)
	c.wg.Wait()
	c.stopCh = make(chan struct{})
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	c.StopWatch()
	return nil
}

// ClearCache clears the configuration cache
func (c *Client) ClearCache() {
	c.cacheMu.Lock()
	c.cache = make(map[string]*Config)
	c.cacheMu.Unlock()
}

// GetCachedVersion returns the cached version of a configuration
func (c *Client) GetCachedVersion(name string) int {
	cacheKey := c.cacheKey(name, c.opts.Namespace, c.opts.Environment)
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	if cached, ok := c.cache[cacheKey]; ok {
		return cached.Version
	}
	return 0
}
