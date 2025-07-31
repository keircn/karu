package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/keircn/karu/pkg/errors"
)

type Client struct {
	httpClient *http.Client
	userAgent  string
	referer    string
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
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
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
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
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
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

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
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
