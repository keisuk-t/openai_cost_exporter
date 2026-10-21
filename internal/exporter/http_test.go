package exporter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestMetrics(t *testing.T) {
	handler, err := NewHandler("test-version")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, err := server.Client().Get(server.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/plain") {
		t.Fatalf("content type = %q", resp.Header.Get("Content-Type"))
	}
	for _, want := range []string{
		"# TYPE openai_cost_exporter_build_info gauge",
		`openai_cost_exporter_build_info{version="test-version"} 1`,
		"go_goroutines ",
		"process_start_time_seconds ",
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("missing %q in metrics", want)
		}
	}
	if strings.Contains(string(body), "openai_cost_amount") {
		t.Fatal("unimplemented providers must not emit cost amounts")
	}
}

func TestRoutes(t *testing.T) {
	handler, err := NewHandler("dev")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method string
		path   string
		status int
		body   string
	}{
		{http.MethodGet, "/healthz", http.StatusOK, "ok\n"},
		{http.MethodGet, "/missing", http.StatusNotFound, "404 page not found\n"},
		{http.MethodPost, "/metrics", http.StatusMethodNotAllowed, "Method Not Allowed\n"},
		{http.MethodPost, "/healthz", http.StatusMethodNotAllowed, "Method Not Allowed\n"},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, httptest.NewRequest(tc.method, tc.path, nil))
			if resp.Code != tc.status || resp.Body.String() != tc.body {
				t.Fatalf("response = %d %q, want %d %q", resp.Code, resp.Body.String(), tc.status, tc.body)
			}
		})
	}
}

func TestIndependentRegistriesAndConcurrentScrapes(t *testing.T) {
	var wg sync.WaitGroup
	for range 3 {
		handler, err := NewHandler("dev")
		if err != nil {
			t.Fatal(err)
		}
		for range 4 {
			wg.Go(func() {
				resp := httptest.NewRecorder()
				handler.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/metrics", nil))
				if resp.Code != http.StatusOK {
					t.Errorf("scrape status = %d", resp.Code)
				}
			})
		}
	}
	wg.Wait()
}
