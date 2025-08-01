package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/keircn/karu/pkg/errors"
)

type Client struct {
	httpClient *http.Client
	userAgents []string
	referer    string
	mu         sync.Mutex
	rand       *rand.Rand
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgents = []string{userAgent}
	}
}

func WithUserAgents(userAgents []string) ClientOption {
	return func(c *Client) {
		c.userAgents = userAgents
	}
}

func WithReferer(referer string) ClientOption {
	return func(c *Client) {
		c.referer = referer
	}
}

func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		},
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) getUserAgent() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.userAgents) == 0 {
		return ""
	}

	return c.userAgents[c.rand.Intn(len(c.userAgents))]
}

func (c *Client) PostJSON(ctx context.Context, url string, payload interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkError, "failed to marshal JSON payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkError, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	if userAgent := c.getUserAgent(); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if c.referer != "" {
		req.Header.Set("Referer", c.referer)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkError, "HTTP request failed")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, errors.New(errors.NetworkError, fmt.Sprintf("HTTP request failed with status %d", resp.StatusCode))
	}

	return resp, nil
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkError, "failed to create HTTP request")
	}

	if userAgent := c.getUserAgent(); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if c.referer != "" {
		req.Header.Set("Referer", c.referer)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkError, "HTTP request failed")
	}

	return resp, nil
}

func (c *Client) DownloadFile(ctx context.Context, url, outputPath string) error {
	resp, err := c.Get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(errors.NetworkError, fmt.Sprintf("failed to download file: status %d", resp.StatusCode))
	}

	return c.writeToFile(resp.Body, outputPath)
}

func (c *Client) writeToFile(src io.Reader, outputPath string) error {
	out, err := createFile(outputPath)
	if err != nil {
		return errors.Wrap(err, errors.ValidationError, "failed to create output file")
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to write file data")
	}

	return nil
}

func createFile(outputPath string) (*os.File, error) {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}
