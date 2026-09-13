package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"sap_segmentation/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("load config: %v\n", err)
		return
	}

	pOffset := 1

	client := &http.Client{
		Timeout: time.Duration(cfg.ConnTimeout) * time.Second,
	}

	u, err := url.Parse(cfg.ConnURI)
	if err != nil {
		fmt.Printf("Invalid CONN_URI: %v\n", err)
		return
	}
	q := u.Query()
	q.Set("p_limit", fmt.Sprintf("%d", cfg.ImportBatchSize))
	q.Set("p_offset", fmt.Sprintf("%d", pOffset))
	u.RawQuery = q.Encode()
	requestURL := u.String()

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		fmt.Printf("create request error: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", cfg.ConnUserAgent)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(cfg.ConnAuthLoginPwd)))

	fmt.Printf("GET %s\n", requestURL)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("http request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read body error: %v\n", err)
		return
	}

	fmt.Printf("status: %s\n", resp.Status)
	fmt.Printf("body:\n%s\n", string(body))
}
