package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func main() {
	connURI := "http://bsm.api.iql.ru/ords/bsm/segmentation/get_segmentation"
	connAuthLoginPwd := "4Dfddf5:jKlljHGH"
	connUserAgent := "spacecount-test"
	connTimeoutSec := 5
	importBatchSize := 50
	pOffset := 1

	client := &http.Client{
		Timeout: time.Duration(connTimeoutSec) * time.Second,
	}

	u, err := url.Parse(connURI)
	if err != nil {
		fmt.Printf("Invalid CONN_URI: %v\n", err)
		return
	}
	q := u.Query()
	q.Set("p_limit", fmt.Sprintf("%d", importBatchSize))
	q.Set("p_offset", fmt.Sprintf("%d", pOffset))
	u.RawQuery = q.Encode()
	requestURL := u.String()

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		fmt.Printf("create request error: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", connUserAgent)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(connAuthLoginPwd)))

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
