package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"iurl/internal/cfg"
	httpc "iurl/internal/http"
)

func TestIntegrationGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"method":  r.Method,
			"url":     r.URL.String(),
			"headers": r.Header,
		})
	}))
	defer server.Close()

	cfg, err := cfg.Parse([]string{server.URL + "/test"})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if data["method"] != "GET" {
		t.Errorf("method = %v, want GET", data["method"])
	}
}

func TestIntegrationPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Content-Type = %s, want application/json", contentType)
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]string
		if err := json.Unmarshal(body, &data); err != nil {
			t.Fatalf("error decoding body: %v", err)
		}

		if data["key"] != "value" {
			t.Errorf("key = %s, want value", data["key"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	}))
	defer server.Close()

	cfg, err := cfg.Parse([]string{
		"-X", "POST",
		"--json", `{"key":"value"}`,
		server.URL + "/test",
	})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestIntegrationHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"accept":   r.Header.Get("Accept"),
			"x-custom": r.Header.Get("X-Custom"),
		})
	}))
	defer server.Close()

	cfg, err := cfg.Parse([]string{
		"-H", "Accept: text/html",
		"-H", "X-Custom: custom-value",
		server.URL + "/test",
	})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if data["accept"] != "text/html" {
		t.Errorf("accept = %s, want text/html", data["accept"])
	}
	if data["x-custom"] != "custom-value" {
		t.Errorf("x-custom = %s, want custom-value", data["x-custom"])
	}
}

func TestIntegrationBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("basic auth not found")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"user": user,
			"pass": pass,
		})
	}))
	defer server.Close()

	cfg, err := cfg.Parse([]string{
		"-u", "testuser:testpass",
		server.URL + "/test",
	})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if data["user"] != "testuser" {
		t.Errorf("user = %s, want testuser", data["user"])
	}
	if data["pass"] != "testpass" {
		t.Errorf("pass = %s, want testpass", data["pass"])
	}
}

func TestIntegrationRedirect(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"calls": callCount})
	}))
	defer server.Close()

	cfg, err := cfg.Parse([]string{
		"-L",
		server.URL + "/redirect",
	})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestIntegrationNoRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	}))
	defer server.Close()

	cfg, err := cfg.Parse([]string{server.URL + "/redirect"})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusFound)
	}
}

func TestIntegrationHeadMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("method = %s, want HEAD", r.Method)
		}
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg, err := cfg.Parse([]string{
		"-X", "HEAD",
		server.URL + "/test",
	})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestIntegrationOutputFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"output":"file"}`)
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp("", "iurl-test-*.json")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg, err := cfg.Parse([]string{
		"-o", tmpFile.Name(),
		server.URL + "/test",
	})
	if err != nil {
		t.Fatalf("error parsing config: %v", err)
	}

	req, err := httpc.NewRequest(cfg)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}

	result, err := httpc.Do(req, cfg)
	if err != nil {
		t.Fatalf("error executing request: %v", err)
	}
	resp := result.Response
	defer resp.Body.Close()

	f, _ := os.Create(tmpFile.Name())
	f.Close()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	if len(content) > 0 {
		t.Errorf("file should be empty, got: %s", string(content))
	}
}
