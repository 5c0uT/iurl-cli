package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"iurl/internal/cfg"
)

func NormalizeURL(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return "http://" + raw
}

func Build() (*cfg.Config, error) {
	reader := bufio.NewReader(os.Stdin)
	cfg := &cfg.Config{
		Headers: make(map[string][]string),
	}

	fmt.Println("=== iurl Interactive Request Builder ===")
	fmt.Println()

	fmt.Print("URL: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}
		cfg.URLs = []string{NormalizeURL(url)}

	fmt.Print("Method (GET): ")
	method, _ := reader.ReadString('\n')
	method = strings.TrimSpace(method)
	if method == "" {
		method = "GET"
	}
	cfg.Method = strings.ToUpper(method)

	fmt.Println()
	fmt.Println("Headers (enter empty line to finish, format: Name: Value):")
	for {
		fmt.Print("  Header: ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			cfg.Headers[key] = append(cfg.Headers[key], val)
		}
	}

	fmt.Println()
	fmt.Println("Body type:")
	fmt.Println("  1. None")
	fmt.Println("  2. Form data")
	fmt.Println("  3. JSON")
	fmt.Println("  4. Raw text")
	fmt.Print("Choice (1): ")
	bodyType, _ := reader.ReadString('\n')
	bodyType = strings.TrimSpace(bodyType)
	if bodyType == "" {
		bodyType = "1"
	}

	switch bodyType {
	case "2":
		fmt.Print("Form data (key=value&key2=value2): ")
		body, _ := reader.ReadString('\n')
		cfg.Body = strings.TrimSpace(body)
		if cfg.Method == "GET" {
			cfg.Method = "POST"
		}
	case "3":
		fmt.Print("JSON data: ")
		body, _ := reader.ReadString('\n')
		cfg.JSON = strings.TrimSpace(body)
		if cfg.Method == "GET" {
			cfg.Method = "POST"
		}
	case "4":
		fmt.Print("Raw data: ")
		body, _ := reader.ReadString('\n')
		cfg.Body = strings.TrimSpace(body)
		if cfg.Method == "GET" {
			cfg.Method = "POST"
		}
	}

	fmt.Println()
	fmt.Print("Follow redirects? (y/N): ")
	follow, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(follow)) == "y" {
		cfg.FollowRedirects = true
	}

	fmt.Print("Include response headers? (y/N): ")
	include, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(include)) == "y" {
		cfg.IncludeHeaders = true
	}

	fmt.Print("Compressed response? (y/N): ")
	compressed, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(compressed)) == "y" {
		cfg.Compressed = true
	}

	fmt.Println()
	fmt.Println("--- Generated command ---")
	fmt.Println(generateCommandLine(cfg))
	fmt.Println()

	fmt.Print("Send request? (Y/n): ")
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) == "n" {
		return nil, fmt.Errorf("cancelled by user")
	}

	return cfg, nil
}

func generateCommandLine(c *cfg.Config) string {
	var parts []string
	parts = append(parts, "iurl")

	if c.Method != "GET" {
		parts = append(parts, "-X", c.Method)
	}

	for key, values := range c.Headers {
		for _, v := range values {
			parts = append(parts, "-H", fmt.Sprintf("%s: %s", key, v))
		}
	}

	if c.Body != "" {
		parts = append(parts, "-d", fmt.Sprintf("%q", c.Body))
	}

	if c.JSON != "" {
		parts = append(parts, "--json", fmt.Sprintf("%q", c.JSON))
	}

	if c.FollowRedirects {
		parts = append(parts, "-L")
	}

	if c.IncludeHeaders {
		parts = append(parts, "-i")
	}

	if c.Compressed {
		parts = append(parts, "--compressed")
	}

	parts = append(parts, c.URLs[0])

	return strings.Join(parts, " ")
}