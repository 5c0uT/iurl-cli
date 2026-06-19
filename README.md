# iurl

Console HTTP client with pretty JSON formatting, support for IPv4/IPv6/IPv8, and many features from curl.

---

## For Users

### Quick Start

```bash
# Build
go build -o iurl ./cmd/iurl

# Or via make
make build
```

### Basic Usage

```bash
# Simple GET request
iurl https://api.example.com

# POST with JSON
iurl -X POST -d '{"name":"John"}' https://api.example.com/users

# POST via --json (automatically sets Content-Type)
iurl --json '{"name":"John"}' https://api.example.com/users

# Headers only (HEAD)
iurl -I https://api.example.com
```

### Request Body

```bash
# Form-urlencoded (automatically sets Content-Type)
iurl -d "name=John&age=30" https://api.example.com

# JSON (automatically sets Content-Type)
iurl --json '{"name":"John"}' https://api.example.com

# JSON from a file
iurl --json @data.json https://api.example.com

# Data from stdin
echo "data" | iurl -d @- https://api.example.com

# File upload via PUT
iurl -T file.txt https://api.example.com/upload

# URL-encode data
iurl --data-urlencode "name=John Doe" https://api.example.com

# Multipart form-data
iurl -F "name=John" -F "file=@photo.jpg" https://api.example.com/upload
```

### Authentication

```bash
# Basic auth
iurl -u user:password https://api.example.com

# Bearer token (via header)
iurl -H "Authorization: Bearer your-token" https://api.example.com
```

### Headers

```bash
# Single header
iurl -H "Accept: application/json" https://api.example.com

# Multiple headers
iurl -H "Accept: application/json" -H "X-Custom: value" https://api.example.com
```

### Compressed Response

```bash
iurl --compressed https://api.example.com
```

### Timeouts

```bash
# Connection timeout (5 seconds)
iurl --connect-timeout 5 https://api.example.com

# Overall request timeout (30 seconds)
iurl --max-time 30 https://api.example.com
```

### Redirects

```bash
# Follow redirects
iurl -L https://example.com/redirect

# Limit redirects
iurl -L --max-redirs 5 https://example.com/redirect
```

### Output

```bash
# Save body to a file
iurl -o response.json https://api.example.com

# Save with the name from the URL
iurl -O https://example.com/file.zip

# Print response headers
iurl -i https://api.example.com

# Output format
iurl -w "%{http_code} %{time_total}s\n" https://api.example.com
```

### IP Addresses

```bash
# IPv4
iurl http://192.168.1.1/api

# IPv6
iurl http://[::1]:8080/health
iurl http://[2001:db8::1]/api

# IPv8 (ASN-dot notation)
iurl 64496.192.0.2.1
iurl http://64496.192.0.2.1/path
```

### Proxy

```bash
iurl -x http://proxy:8080 https://api.example.com
iurl -x http://proxy:8080 --proxy-user user:pass https://api.example.com
```

### Retries

```bash
# Retry 3 times on failure
iurl --retry 3 https://unstable.example.com

# Delay between retries
iurl --retry 3 --retry-delay 2 https://unstable.example.com
```

### Cookies

```bash
# Load and save cookies
iurl -b cookies.txt -c cookies.txt https://api.example.com/login

# Load only
iurl -b cookies.txt https://api.example.com/protected
```

### Variables and Templates

```bash
# Variables from a file (JSON, YAML, dotenv)
iurl --vars-file vars.json https://{{host}}/api

# Variables from the command line
iurl --var env=prod --var version=1.0 https://api.{{env}}.example.com/{{version}}/data
```

### JSON Filtering (jq-like)

```bash
# Extract a field
iurl https://api.example.com/users --query '.[0].name'

# Filtering
iurl https://api.example.com/logs --query '.[] | select(.level=="error")'

# Aggregation
iurl https://api.example.com/orders --query 'map(.total) | add'
```

### Response Comparison

```bash
# First request – saves a baseline
iurl --diff baseline https://api.example.com/status

# Second request – compares against the baseline
iurl --diff baseline https://api.example.com/status
```

### Monitoring

```bash
# Check every 10 seconds
iurl --watch 10s https://api.example.com/status

# Every 2 minutes
iurl --watch 2m https://api.example.com/status
```

### Code Generation

```bash
# Generate Python code
iurl --generate-code python https://api.example.com

# Go
iurl --generate-code go https://api.example.com

# JavaScript
iurl --generate-code js https://api.example.com

# curl
iurl --generate-code curl https://api.example.com
```

### Request Profiles

```bash
# Save a profile
iurl --save my-request.json -X POST --json '{"key":"value"}' https://api.example.com

# Load and execute
iurl --load my-request.json

# Load and override the URL
iurl --load my-request.json https://other-api.com
```

### History

```bash
# Show history
iurl --history

# Search by tag
iurl --tag "auth-test" https://api.example.com/login
iurl --search auth-test

# Rerun a request from history
iurl --rerun 42
```

### Interactive Mode

```bash
# Request builder
iurl --build

# Raw HTTP dialogue
iurl --raw-shell https://api.example.com
```

### Configuration

```bash
# Configuration file
iurl -K config.txt

# File format:
# url = https://api.example.com
# method = POST
# header = Content-Type: application/json
# output = response.json
# user = admin:secret
```

### Error Handling

```bash
# Silent mode (no output)
iurl -s https://api.example.com

# Silent mode but show errors
iurl -s -S https://api.example.com

# Fail on HTTP 4xx/5xx
iurl --fail https://api.example.com

# Fail with response body
iurl --fail-with-body https://api.example.com
```

### Write-out Format

```bash
# Print response code and time
iurl -w "%{http_code} %{time_total}s\n" https://api.example.com

# Available variables:
# %{http_code} – HTTP status code
# %{http_content_type} – Content-Type of the response
# %{time_total} – total request time
# %{size_download} – body size
# %{url_effective} – final effective URL
# %{remote_ip} – server IP
# %{remote_port} – server port
```

---

## For Developers

### Project Structure

```
iurl/
├── cmd/iurl/
│   ├── main.go           # Entry point, orchestration
│   └── main_test.go      # Integration tests
├── internal/
│   ├── cfg/              # Configuration, CLI, templates, variables
│   │   └── cfg.go
│   ├── http/             # HTTP client, request building
│   │   └── http.go
│   ├── fmt/              # JSON formatting with highlighting
│   │   └── fmt.go
│   ├── cookiejar/        # Cookie jar (Netscape format)
│   │   ├── cookiejar.go
│   │   └── cookiejar_test.go
│   ├── repl/             # Interactive mode (tab-completion)
│   │   └── repl.go
│   ├── query/            # jq-like JSON filtering
│   │   └── query.go
│   ├── codegen/          # Code generation (python/go/js/curl)
│   │   └── codegen.go
│   ├── storage/          # History, profiles, diff cache
│   │   └── storage.go
│   └── interactive/      # Interactive request builder
│       └── interactive.go
├── Makefile
├── build.sh
└── go.mod
```

### Architecture

**Execution flow:**
1. `cfg.Parse()` – parse CLI → `Config`
2. `request.New()` – build `http.Request` from `Config`
3. `http.DoWithResult()` – execute request → `Result`
4. `fmt.PrettyPrintJSON()` / `CopyRaw()` – output the result

**Dependencies:**
- `github.com/chzyer/readline` – tab-completion in REPL
- `golang.org/x/net` – HTTP/2 support
- `gopkg.in/yaml.v3` – YAML variables

### Build

```bash
# Local build
make build

# Build for all platforms
make all

# Tests
make test

# Linter
make lint

# Clean
make clean
```

### Tests

```bash
# All tests
go test ./...

# Verbose output
go test ./... -v

# Specific package
go test ./internal/config/ -v

# Integration tests (with httptest)
go test ./cmd/iurl/ -v
```

### Adding a New Flag

1. Add a field to `Config` in `internal/cfg/cfg.go`
2. Register the flag in `Parse()`
3. Handle the logic in `cmd/iurl/main.go`
4. Add a description in `PrintHelp()`
5. Write a test

### Adding a New Package

1. Create a directory `internal/<name>/`
2. Implement the logic
3. Write tests
4. Import in `cmd/iurl/main.go`

### Cookie File Format

Netscape format (compatible with curl):
```
# Netscape HTTP Cookie File
.domain.com	TRUE	/	FALSE	1735689600	session	abc123
```

Format: `domain \t flag \t path \t secure \t expires \t name \t value`

### Request Profile Format

JSON:
```json
{
  "url": "https://api.example.com",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"key\":\"value\"}",
  "insecure": false,
  "compressed": true
}
```

### History Format

JSON in `~/.iurl_history`:
```json
{
  "entries": [
    {
      "id": 1,
      "timestamp": "2026-01-01T12:00:00Z",
      "method": "GET",
      "url": "https://api.example.com",
      "headers": {},
      "tags": ["test"],
      "status": 200
    }
  ]
}
```

### Completion Scripts

To generate autocompletion scripts (bash/zsh/fish) use:

```bash
iurl --completion bash > /etc/bash_completion.d/iurl
iurl --completion zsh > ~/.zsh/completions/_iurl
iurl --completion fish > ~/.config/fish/completions/iurl.fish
```

### Cross-Platform

The project works on:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

REPL history is stored in `~/.iurl_history` (paths are normalized using `os.PathSeparator`).

---

## License

MIT
