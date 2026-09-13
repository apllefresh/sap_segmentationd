package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sap_segmentation/internal/config"
)

func testConfig(baseURL string) config.Config {
	return config.Config{
		ConnURI:          baseURL,
		ConnAuthLoginPwd: "user:pass",
		ConnUserAgent:    "test-agent",
		ConnTimeout:      5,
		ConnIntervalMs:   0,
		ImportBatchSize:  50,
	}
}

func TestBatchOffset(t *testing.T) {
	c := NewClient(testConfig("http://example.com"))

	tests := []struct {
		page   int
		offset int
	}{
		{0, 1},
		{1, 50},
		{2, 100},
		{3, 150},
	}

	for _, tt := range tests {
		if got := c.batchOffset(tt.page); got != tt.offset {
			t.Errorf("page %d: offset = %d, want %d", tt.page, got, tt.offset)
		}
	}
}

func TestBuildURL(t *testing.T) {
	c := NewClient(testConfig("http://example.com/api"))

	url, err := c.buildURL(1)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(url, "p_limit=50") {
		t.Errorf("url %q missing p_limit=50", url)
	}
	if !strings.Contains(url, "p_offset=1") {
		t.Errorf("url %q missing p_offset=1", url)
	}
}

func TestDecodeBatch(t *testing.T) {
	t.Run("empty body", func(t *testing.T) {
		_, empty, err := decodeBatch([]byte("  "))
		if err != nil || !empty {
			t.Fatalf("empty=%v err=%v", empty, err)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		_, empty, err := decodeBatch([]byte("[]"))
		if err != nil || !empty {
			t.Fatalf("empty=%v err=%v", empty, err)
		}
	})

	t.Run("valid segments", func(t *testing.T) {
		body := `[{"address_sap_id":"A1","adr_segment":"S1","segment_id":10}]`
		segments, empty, err := decodeBatch([]byte(body))
		if err != nil || empty {
			t.Fatalf("empty=%v err=%v", empty, err)
		}
		if len(segments) != 1 || segments[0].AddressSapID != "A1" || segments[0].SegmentID != 10 {
			t.Fatalf("unexpected segments: %+v", segments)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, _, err := decodeBatch([]byte(`{`))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestFetchBatch_headersAndDecode(t *testing.T) {
	var gotUserAgent, gotAuth string
	var gotLimit, gotOffset string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		gotAuth = r.Header.Get("Authorization")
		gotLimit = r.URL.Query().Get("p_limit")
		gotOffset = r.URL.Query().Get("p_offset")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"address_sap_id":"X","adr_segment":"AB","segment_id":1}]`))
	}))
	defer srv.Close()

	c := NewClient(testConfig(srv.URL))
	segments, status, err := c.FetchBatch(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if status != "200 OK" {
		t.Fatalf("status: %s", status)
	}
	if len(segments) != 1 || segments[0].AddressSapID != "X" {
		t.Fatalf("segments: %+v", segments)
	}
	if gotUserAgent != "test-agent" {
		t.Errorf("User-Agent = %q", gotUserAgent)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if gotAuth != wantAuth {
		t.Errorf("Authorization = %q, want %q", gotAuth, wantAuth)
	}
	if gotLimit != "50" || gotOffset != "1" {
		t.Errorf("query limit=%q offset=%q", gotLimit, gotOffset)
	}
}

func TestFetchBatch_emptyStopsPagination(t *testing.T) {
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		offset := r.URL.Query().Get("p_offset")
		if offset == "1" {
			_, _ = w.Write([]byte(`[{"address_sap_id":"A","adr_segment":"S","segment_id":1}]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := NewClient(testConfig(srv.URL))
	if err := c.RunAll(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if call != 2 {
		t.Fatalf("expected 2 HTTP calls, got %d", call)
	}
}

func TestFetchBatch_httpError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(testConfig(srv.URL))
	_, _, err := c.FetchBatch(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}
