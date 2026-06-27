package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/http2"

	"iurl/internal/cfg"
)

type Result struct {
	Response *http.Response
	Duration time.Duration
}

func NewRequest(c *cfg.Config) (*http.Request, error) {
	var body io.Reader
	var contentType string

	if c.BodyStdin {
		body = os.Stdin
	} else if c.Body != "" {
		body = strings.NewReader(c.Body)
	} else if c.JSON != "" {
		body = strings.NewReader(c.JSON)
	} else if c.JSONFile != "" {
		f, err := os.Open(c.JSONFile)
		if err != nil {
			return nil, fmt.Errorf("cannot open JSON file %s: %v", c.JSONFile, err)
		}
		body = f
	} else if len(c.FormData) > 0 {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		for _, field := range c.FormData {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid form field: %s", field)
			}
			name := parts[0]
			value := parts[1]

			if strings.HasPrefix(value, "@") {
				filePath := value[1:]
				file, err := os.Open(filePath)
				if err != nil {
					return nil, fmt.Errorf("cannot open file %s: %v", filePath, err)
				}
				defer file.Close()

				part, err := writer.CreateFormFile(name, filePath)
				if err != nil {
					return nil, fmt.Errorf("error creating form field: %v", err)
				}
				if _, err := io.Copy(part, file); err != nil {
					return nil, fmt.Errorf("error reading file: %v", err)
				}
			} else {
				if err := writer.WriteField(name, value); err != nil {
					return nil, fmt.Errorf("error writing form field: %v", err)
				}
			}
		}

		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("error closing form: %v", err)
		}

		body = &buf
		contentType = writer.FormDataContentType()
	}

	req, err := http.NewRequest(c.Method, c.URLs[0], body)
	if err != nil {
		return nil, err
	}
	if c.JSONFile != "" {
		jsonFile := c.JSONFile
		req.GetBody = func() (io.ReadCloser, error) {
			return os.Open(jsonFile)
		}
	}

	for key, values := range c.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if c.BasicAuth != "" {
		parts := strings.SplitN(c.BasicAuth, ":", 2)
		if len(parts) == 2 {
			req.SetBasicAuth(parts[0], parts[1])
		}
	}

	if contentType != "" && !hasHeader(c.Headers, "Content-Type") {
		req.Header.Set("Content-Type", contentType)
	} else if (c.Body != "" || c.BodyStdin) && !hasHeader(c.Headers, "Content-Type") {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if (c.JSON != "" || c.JSONFile != "") && !hasHeader(c.Headers, "Content-Type") {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func Do(req *http.Request, c *cfg.Config) (*Result, error) {
	var lastErr error

	for attempt := 0; attempt <= c.Retry; attempt++ {
		if attempt > 0 {
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("cannot reset request body for retry: %v", err)
				}
				req.Body = body
			} else if req.Body != nil && req.Body != http.NoBody {
				return nil, fmt.Errorf("cannot retry request with non-rewindable body")
			}
			time.Sleep(time.Duration(c.RetryDelay * float64(time.Second)))
		}

		result, err := doRequest(req, c)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

func doRequest(req *http.Request, c *cfg.Config) (*Result, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.Insecure,
	}

	if c.CACert != "" {
		caCert, err := os.ReadFile(c.CACert)
		if err != nil {
			return nil, fmt.Errorf("cannot read CA cert: %v", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	if c.Cert != "" && c.Key != "" {
		cert, err := tls.LoadX509KeyPair(c.Cert, c.Key)
		if err != nil {
			return nil, fmt.Errorf("cannot load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 10,
		TLSClientConfig:     tlsConfig,
	}

	if c.ConnectTimeout > 0 {
		dialer := &net.Dialer{
			Timeout:   time.Duration(c.ConnectTimeout * float64(time.Second)),
			KeepAlive: 30 * time.Second,
		}
		transport.DialContext = dialer.DialContext
	}

	if c.Proxy != "" {
		proxyURL, err := url.Parse(c.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	if c.HTTP2 {
		transport.TLSClientConfig = &tls.Config{
			NextProtos:         []string{"h2"},
			InsecureSkipVerify: c.Insecure,
		}
		t := &http2.Transport{
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return tls.Dial(network, addr, cfg)
			},
		}
		client := &http.Client{
			Transport: t,
			Timeout:   time.Duration(c.MaxTime * float64(time.Second)),
		}
		return doHTTP(client, req)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(c.MaxTime * float64(time.Second)),
	}

	if !c.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if c.MaxRedirects > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= c.MaxRedirects {
				return fmt.Errorf("redirect limit exceeded: %d", c.MaxRedirects)
			}
			return nil
		}
	}

	return doHTTP(client, req)
}

func doHTTP(client *http.Client, req *http.Request) (*Result, error) {
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "no such host"):
			return nil, fmt.Errorf("failed to resolve hostname: %v", err)
		case strings.Contains(errStr, "connection refused"):
			return nil, fmt.Errorf("connection refused: %v", err)
		case strings.Contains(errStr, "i/o timeout"):
			return nil, fmt.Errorf("connection timeout: %v", err)
		case strings.Contains(errStr, "context deadline exceeded"):
			return nil, fmt.Errorf("request timeout: %v", err)
		case strings.Contains(errStr, "certificate"):
			return nil, fmt.Errorf("TLS error (use --insecure): %v", err)
		default:
			return nil, fmt.Errorf("request failed: %v", err)
		}
	}

	return &Result{Response: resp, Duration: duration}, nil
}

func hasHeader(headers map[string][]string, key string) bool {
	for k := range headers {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}
