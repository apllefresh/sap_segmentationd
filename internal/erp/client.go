package erp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sap_segmentation/internal/config"
	"sap_segmentation/model"
)

type Client struct {
	http *http.Client
	cfg  config.Config
}

func NewClient(cfg config.Config) *Client {
	return NewClientWithHTTP(cfg, nil)
}

func NewClientWithHTTP(cfg config.Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Duration(cfg.ConnTimeout) * time.Second,
		}
	}
	return &Client{http: httpClient, cfg: cfg}
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

func (c *Client) FetchBatch(ctx context.Context, offset int) ([]model.Segmentation, string, error) {
	requestURL, err := c.buildURL(offset)
	if err != nil {
		return nil, "", err
	}

	fmt.Printf("GET %s\n", requestURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", c.cfg.ConnUserAgent)
	auth := base64.StdEncoding.EncodeToString([]byte(c.cfg.ConnAuthLoginPwd))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, resp.Status, fmt.Errorf("unexpected status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	segments, empty, err := decodeBatch(body)
	if err != nil {
		return nil, resp.Status, err
	}
	if empty {
		return nil, resp.Status, nil
	}

	return segments, resp.Status, nil
}

func decodeBatch(body []byte) (segments []model.Segmentation, empty bool, err error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return nil, true, nil
	}

	if err := json.Unmarshal(body, &segments); err != nil {
		return nil, false, fmt.Errorf("decode json: %w", err)
	}
	if len(segments) == 0 {
		return nil, true, nil
	}
	return segments, false, nil
}

// RunAll requests ERP pages until the response body is empty (offset: 1, 50, 100, …).
func (c *Client) RunAll(ctx context.Context) error {
	for page := 0; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		offset := c.batchOffset(page)
		segments, status, err := c.FetchBatch(ctx, offset)
		if err != nil {
			return err
		}
		if len(segments) == 0 {
			fmt.Println("empty response, import loop finished")
			return nil
		}

		fmt.Printf("status: %s\n", status)
		for _, segment := range segments {
			fmt.Printf("%+v\n", segment)
		}
		fmt.Println()

		time.Sleep(time.Duration(c.cfg.ConnIntervalMs) * time.Millisecond)
	}
}
