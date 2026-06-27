package http

import (
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"iurl/internal/cfg"
)

func TestDoRetriesResetRequestBody(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}
		if attempt == 1 {
			hj, ok := w.(nethttp.Hijacker)
			if !ok {
				t.Fatal("ResponseWriter does not support hijacking")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("Hijack failed: %v", err)
			}
			conn.Close()
			return
		}
		if string(body) != "payload" {
			t.Fatalf("retry body = %q, want %q", string(body), "payload")
		}
		w.WriteHeader(nethttp.StatusOK)
	}))
	defer server.Close()

	c := &cfg.Config{
		URLs:       []string{server.URL},
		Method:     "POST",
		Headers:    make(map[string][]string),
		Body:       "payload",
		Retry:      1,
		RetryDelay: 0,
		MaxTime:    5,
	}
	req, err := NewRequest(c)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	result, err := Do(req, c)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	result.Response.Body.Close()
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}
