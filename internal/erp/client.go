package erp

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sap_segmentation/internal/config"
)

type Client struct {
	http *http.Client
	cfg  config.Config
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		http: &http.Client{
			Timeout: time.Duration(cfg.ConnTimeout) * time.Second,
		},
		cfg: cfg,
	}
}

func (c *Client) batchOffset(page int) int {
	if page == 0 {
		return 1
	}
	return page * c.cfg.ImportBatchSize
}

func (c *Client) buildURL(offset int) (string, error) {
	u, err := url.Parse(c.cfg.ConnURI)
	if err != nil {
		return "", fmt.Errorf("parse CONN_URI: %w", err)
	}
	q := u.Query()
	q.Set("p_limit", fmt.Sprintf("%d", c.cfg.ImportBatchSize))
	q.Set("p_offset", fmt.Sprintf("%d", offset))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) fetchBatch(ctx context.Context, offset int) (body []byte, status string, requestURL string, err error) {
	requestURL, err = c.buildURL(offset)
	if err != nil {
		return nil, "", "", err
	}

	fmt.Printf("GET %s\n", requestURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, "", requestURL, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", c.cfg.ConnUserAgent)
	auth := base64.StdEncoding.EncodeToString([]byte(c.cfg.ConnAuthLoginPwd))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", requestURL, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Status, requestURL, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return body, resp.Status, requestURL, fmt.Errorf("unexpected status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, resp.Status, requestURL, nil
}

func (c *Client) isEmptyBatch(body []byte) bool {
	trimmed := strings.TrimSpace(string(body))
	return trimmed == "" || trimmed == "null" || trimmed == "[]"
}

// RunAll requests ERP pages until the response body is empty (offset: 1, 50, 100, …).
func (c *Client) RunAll(ctx context.Context) error {
	for page := 0; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		offset := c.batchOffset(page)
		body, status, _, err := c.fetchBatch(ctx, offset)
		if err != nil {
			return err
		}

		if c.isEmptyBatch(body) {
			fmt.Println("empty response, import loop finished")
			return nil
		}

		fmt.Printf("status: %s\n", status)
		fmt.Printf("body:\n%s\n\n", string(body))

		time.Sleep(time.Duration(c.cfg.ConnIntervalMs) * time.Millisecond)
	}
}
