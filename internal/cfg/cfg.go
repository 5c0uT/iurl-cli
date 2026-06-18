package cfg

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var Version = "v1.0.0"

type Config struct {
	URLs            []string
	Method          string
	Headers         map[string][]string
	Body            string
	BodyStdin       bool
	JSON            string
	JSONFile        string
	FormData        []string
	BasicAuth       string
	ConnectTimeout  float64
	MaxTime         float64
	FollowRedirects bool
	MaxRedirects    int
	IncludeHeaders  bool
	Verbose         bool
	NoPretty        bool
	Trace           string
	TraceASCII      string
	OutputFile      string
	OutputDir       string
	RemoteName      bool
	CreateDirs      bool
	NoClobber       bool
	VarsFile        string
	Vars            []string
	HTTP2           bool
	HTTP10          bool
	HTTP11          bool
	HTTP3           bool
	CookieJarRead   string
	CookieJarWrite  string
	Insecure        bool
	Compressed      bool
	Head            bool
	UploadFile      string
	Resolve         []string
	ConnectTo       []string
	Retry           int
	RetryDelay      float64
	RetryMaxTime    float64
	RetryConnRefused bool
	RetryAllErrors  bool
	ConfigFile      string
	WriteOut        string
	Version         bool
	Fail            bool
	FailWithBody    bool
	ForceIPv4       bool
	ForceIPv6       bool
	NoProgressMeter bool
	ShowError       bool
	Stderr          string
	DataURLEncode   []string
	DataBinary      []string
	DataRaw         []string
	GET             bool
	FormString      []string
	Proxy           string
	ProxyUser       string
	ProxyInsecure   bool
	Post301         bool
	Post302         bool
	Post303         bool
	LocationTrusted bool
	CACert          string
	Cert            string
	Key             string
	CertType        string
	AnyAuth         bool
	Digest          bool
	Negotiate       bool
	NTLM            bool
	MaxFilesize     int64
	LimitRate       string
	Parallel        bool
	ParallelMax     int
	Query           string
	Diff            string
	DiffName        string
	ShowChain       bool
	SaveProfile     string
	LoadProfile     string
	History         bool
	Tag             []string
	Search          string
	Rerun           string
	GenerateCode    string
	Watch           string
	Build           bool
	RawShell        string
	Validate        string
	Force           bool
	PreScript       string
	PostScript      string
	AdaptiveTimeout string
	ShowSecrets     bool
	Help            bool
}

type Variables map[string]string

func Parse(args []string) (*Config, error) {
	cfg := &Config{
		Headers: make(map[string][]string),
	}

	fs := flag.NewFlagSet("iurl", flag.ExitOnError)
	fs.StringVar(&cfg.Method, "X", "GET", "HTTP method")
	fs.StringVar(&cfg.Method, "request", "GET", "HTTP method (long form)")
	fs.Var(&headerValue{cfg.Headers}, "H", "HTTP header (can be specified multiple times)")
	fs.Var(&headerValue{cfg.Headers}, "header", "HTTP header (long form)")
	fs.StringVar(&cfg.Body, "d", "", "HTTP request data")
	fs.StringVar(&cfg.Body, "data", "", "HTTP request data (long form)")
	fs.BoolVar(&cfg.BodyStdin, "data-stdin", false, "Read request body from stdin")
	fs.StringVar(&cfg.JSON, "json", "", "JSON request data")
	fs.StringVar(&cfg.JSONFile, "json-file", "", "JSON file path")
	fs.Var(&formValue{&cfg.FormData}, "F", "Form field (can be specified multiple times)")
	fs.Var(&formValue{&cfg.FormData}, "form", "Form field (long form)")
	fs.Var(&formStringValue{&cfg.FormString}, "form-string", "Form field without parsing (can be specified multiple times)")
	fs.StringVar(&cfg.BasicAuth, "u", "", "Basic authentication (user:password)")
	fs.StringVar(&cfg.BasicAuth, "user", "", "Basic authentication (long form)")
	fs.Float64Var(&cfg.ConnectTimeout, "connect-timeout", 10, "Connection timeout in seconds")
	fs.Float64Var(&cfg.MaxTime, "max-time", 30, "Maximum request time in seconds")
	fs.BoolVar(&cfg.FollowRedirects, "L", false, "Follow redirects")
	fs.BoolVar(&cfg.FollowRedirects, "location", false, "Follow redirects (long form)")
	fs.IntVar(&cfg.MaxRedirects, "max-redirs", 10, "Maximum number of redirects")
	fs.BoolVar(&cfg.IncludeHeaders, "i", false, "Include response headers")
	fs.BoolVar(&cfg.IncludeHeaders, "include", false, "Include response headers (long form)")
	fs.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Verbose output (long form)")
	fs.BoolVar(&cfg.NoPretty, "no-pretty", false, "Disable pretty printing")
	fs.StringVar(&cfg.Trace, "trace", "", "Trace output file")
	fs.StringVar(&cfg.TraceASCII, "trace-ascii", "", "Trace ASCII output file")
	fs.StringVar(&cfg.OutputFile, "o", "", "Output file")
	fs.StringVar(&cfg.OutputFile, "output", "", "Output file (long form)")
	fs.StringVar(&cfg.OutputDir, "output-dir", "", "Output directory")
	fs.BoolVar(&cfg.RemoteName, "O", false, "Save file with name from URL")
	fs.BoolVar(&cfg.RemoteName, "remote-name", false, "Save file with name from URL (long form)")
	fs.BoolVar(&cfg.CreateDirs, "create-dirs", false, "Create output directories")
	fs.BoolVar(&cfg.NoClobber, "no-clobber", false, "Don't overwrite existing files")
	fs.StringVar(&cfg.VarsFile, "vars-file", "", "Variables file (JSON)")
	fs.StringVar(&cfg.VarsFile, "V", "", "Variables file (short form)")
	fs.Var(&varValue{&cfg.Vars}, "var", "Variable in key=value format (can be specified multiple times)")
	fs.BoolVar(&cfg.HTTP2, "http2", false, "Force HTTP/2 usage")
	fs.BoolVar(&cfg.HTTP10, "http1.0", false, "Use HTTP/1.0")
	fs.BoolVar(&cfg.HTTP11, "http1.1", false, "Use HTTP/1.1")
	fs.BoolVar(&cfg.HTTP3, "http3", false, "Force HTTP/3 usage")
	fs.StringVar(&cfg.CookieJarRead, "b", "", "Load cookies from file")
	fs.StringVar(&cfg.CookieJarRead, "cookie", "", "Load cookies from file (long form)")
	fs.StringVar(&cfg.CookieJarWrite, "c", "", "Save cookies to file")
	fs.StringVar(&cfg.CookieJarWrite, "cookie-jar", "", "Save cookies to file (long form)")
	fs.BoolVar(&cfg.Insecure, "k", false, "Skip TLS certificate verification")
	fs.BoolVar(&cfg.Insecure, "insecure", false, "Skip TLS certificate verification (long form)")
	fs.BoolVar(&cfg.Compressed, "compressed", false, "Request compressed response")
	fs.BoolVar(&cfg.Head, "I", false, "Fetch headers only (HEAD method)")
	fs.BoolVar(&cfg.Head, "head", false, "Fetch headers only (long form)")
	fs.StringVar(&cfg.UploadFile, "T", "", "Upload file with PUT")
	fs.StringVar(&cfg.UploadFile, "upload-file", "", "Upload file with PUT (long form)")
	fs.Var(&resolveValue{&cfg.Resolve}, "resolve", "Resolve hostname to IP (host:port:address)")
	fs.Var(&connectToValue{&cfg.ConnectTo}, "connect-to", "Connect to different host (HOST1:PORT1:HOST2:PORT2)")
	fs.IntVar(&cfg.Retry, "retry", 0, "Number of retries on failure")
	fs.Float64Var(&cfg.RetryDelay, "retry-delay", 1.0, "Delay between retries in seconds")
	fs.Float64Var(&cfg.RetryMaxTime, "retry-max-time", 0, "Maximum time for retries in seconds")
	fs.BoolVar(&cfg.RetryConnRefused, "retry-connrefused", false, "Retry on connection refused")
	fs.BoolVar(&cfg.RetryAllErrors, "retry-all-errors", false, "Retry on all errors")
	fs.StringVar(&cfg.ConfigFile, "K", "", "Config file path")
	fs.StringVar(&cfg.ConfigFile, "config", "", "Config file path (long form)")
	fs.StringVar(&cfg.WriteOut, "w", "", "Output format string after request")
	fs.StringVar(&cfg.WriteOut, "write-out", "", "Output format string after request (long form)")
	fs.BoolVar(&cfg.Version, "version", false, "Show version")
	fs.BoolVar(&cfg.Fail, "fail", false, "Fail silently on HTTP errors")
	fs.BoolVar(&cfg.FailWithBody, "fail-with-body", false, "Fail on HTTP errors but show body")
	fs.BoolVar(&cfg.ForceIPv4, "4", false, "Force IPv4")
	fs.BoolVar(&cfg.ForceIPv4, "ipv4", false, "Force IPv4 (long form)")
	fs.BoolVar(&cfg.ForceIPv6, "6", false, "Force IPv6")
	fs.BoolVar(&cfg.ForceIPv6, "ipv6", false, "Force IPv6 (long form)")
	fs.BoolVar(&cfg.NoProgressMeter, "s", false, "Silent mode")
	fs.BoolVar(&cfg.NoProgressMeter, "silent", false, "Silent mode (long form)")
	fs.BoolVar(&cfg.NoProgressMeter, "no-progress-meter", false, "Disable progress meter")
	fs.BoolVar(&cfg.ShowError, "S", false, "Show errors in silent mode")
	fs.BoolVar(&cfg.ShowError, "show-error", false, "Show errors in silent mode (long form)")
	fs.StringVar(&cfg.Stderr, "stderr", "", "Redirect stderr to file")
	fs.Var(&dataURLEncodeValue{&cfg.DataURLEncode}, "data-urlencode", "URL-encode data (can be specified multiple times)")
	fs.Var(&dataBinaryValue{&cfg.DataBinary}, "data-binary", "Send data binary (can be specified multiple times)")
	fs.Var(&dataRawValue{&cfg.DataRaw}, "data-raw", "Send data raw (can be specified multiple times)")
	fs.BoolVar(&cfg.GET, "G", false, "Convert -d to GET parameters")
	fs.BoolVar(&cfg.GET, "get", false, "Convert -d to GET parameters (long form)")
	fs.StringVar(&cfg.Proxy, "x", "", "Proxy URL")
	fs.StringVar(&cfg.Proxy, "proxy", "", "Proxy URL (long form)")
	fs.StringVar(&cfg.ProxyUser, "proxy-user", "", "Proxy authentication (user:password)")
	fs.BoolVar(&cfg.ProxyInsecure, "proxy-insecure", false, "Skip proxy TLS verification")
	fs.BoolVar(&cfg.Post301, "post301", false, "Keep POST on 301 redirect")
	fs.BoolVar(&cfg.Post302, "post302", false, "Keep POST on 302 redirect")
	fs.BoolVar(&cfg.Post303, "post303", false, "Keep POST on 303 redirect")
	fs.BoolVar(&cfg.LocationTrusted, "location-trusted", false, "Follow redirects with credentials")
	fs.StringVar(&cfg.CACert, "cacert", "", "CA certificate file")
	fs.StringVar(&cfg.Cert, "cert", "", "Client certificate file")
	fs.StringVar(&cfg.Key, "key", "", "Private key file")
	fs.StringVar(&cfg.CertType, "cert-type", "", "Certificate type (PEM, DER, ENG)")
	fs.BoolVar(&cfg.AnyAuth, "anyauth", false, "Auto-detect authentication method")
	fs.BoolVar(&cfg.Digest, "digest", false, "Use Digest authentication")
	fs.BoolVar(&cfg.Negotiate, "negotiate", false, "Use Negotiate/Kerberos authentication")
	fs.BoolVar(&cfg.NTLM, "ntlm", false, "Use NTLM authentication")
	fs.Int64Var(&cfg.MaxFilesize, "max-filesize", 0, "Maximum file size to download")
	fs.StringVar(&cfg.LimitRate, "limit-rate", "", "Limit transfer rate (e.g., 100K, 1M)")
	fs.BoolVar(&cfg.Parallel, "parallel", false, "Enable parallel transfers")
	fs.IntVar(&cfg.ParallelMax, "parallel-max", 5, "Maximum parallel transfers")
	fs.StringVar(&cfg.Query, "query", "", "Filter JSON response with jq-like expression")
	fs.StringVar(&cfg.DiffName, "diff", "", "Compare with cached response")
	fs.BoolVar(&cfg.ShowChain, "show-chain", false, "Show redirect chain")
	fs.StringVar(&cfg.SaveProfile, "save", "", "Save request profile to file")
	fs.StringVar(&cfg.LoadProfile, "load", "", "Load request profile from file")
	fs.BoolVar(&cfg.History, "history", false, "Show request history")
	fs.Var(&tagValue{&cfg.Tag}, "tag", "Tag for request (can be specified multiple times)")
	fs.StringVar(&cfg.Search, "search", "", "Search history by tag or text")
	fs.StringVar(&cfg.Rerun, "rerun", "", "Re-execute request from history by ID")
	fs.StringVar(&cfg.GenerateCode, "generate-code", "", "Generate code for request (python/go/js/curl)")
	fs.StringVar(&cfg.Watch, "watch", "", "Monitor URL with interval (e.g., 10s, 2m)")
	fs.BoolVar(&cfg.Build, "build", false, "Interactive request builder")
	fs.StringVar(&cfg.RawShell, "raw-shell", "", "Raw HTTP shell mode")
	fs.StringVar(&cfg.Validate, "validate", "", "Validate request against OpenAPI spec")
	fs.BoolVar(&cfg.Force, "force", false, "Force execution despite validation errors")
	fs.StringVar(&cfg.PreScript, "pre-script", "", "Script to run before request")
	fs.StringVar(&cfg.PostScript, "post-script", "", "Script to run after request")
	fs.StringVar(&cfg.AdaptiveTimeout, "adaptive-timeout", "", "Adaptive timeout (min:max)")
	fs.BoolVar(&cfg.ShowSecrets, "show-secrets", false, "Show resolved secret values")
	fs.BoolVar(&cfg.Help, "help", false, "Show help")
	fs.BoolVar(&cfg.Help, "h", false, "Show help")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if cfg.Version || cfg.Help || cfg.Build || cfg.History {
		return cfg, nil
	}

	if cfg.ConfigFile != "" {
		if err := loadConfigFile(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.LoadProfile != "" {
		// Profile loading handled in main.go
	}

	if fs.NArg() < 1 && cfg.Rerun == "" && cfg.Search == "" {
		return nil, fmt.Errorf("URL is required")
	}

	cfg.URLs = fs.Args()

	if cfg.Method == "" {
		cfg.Method = "GET"
	}
	cfg.Method = strings.ToUpper(cfg.Method)

	if cfg.Head {
		cfg.Method = "HEAD"
	}

	if cfg.GET {
		if cfg.Method == "POST" || cfg.Method == "PUT" || cfg.Method == "PATCH" {
			if cfg.Body != "" {
				parsed, err := url.Parse(cfg.URLs[0])
				if err == nil {
					existing := parsed.Query()
					params := strings.Split(cfg.Body, "&")
					for _, param := range params {
						kv := strings.SplitN(param, "=", 2)
						if len(kv) == 2 {
							existing.Add(kv[0], kv[1])
						}
					}
					parsed.RawQuery = existing.Encode()
					cfg.URLs[0] = parsed.String()
					cfg.Body = ""
					cfg.Method = "GET"
				}
			}
		}
	}

	if cfg.UploadFile != "" {
		if cfg.Method == "GET" {
			cfg.Method = "PUT"
		}
		data, err := os.ReadFile(cfg.UploadFile)
		if err != nil {
			return nil, fmt.Errorf("cannot read upload file: %v", err)
		}
		cfg.Body = string(data)
	}

	for _, d := range cfg.DataURLEncode {
		parts := strings.SplitN(d, "=", 2)
		if len(parts) == 2 {
			encoded := url.QueryEscape(parts[1])
			if cfg.Body != "" {
				cfg.Body += "&"
			}
			cfg.Body += parts[0] + "=" + encoded
		} else {
			encoded := url.QueryEscape(d)
			if cfg.Body != "" {
				cfg.Body += "&"
			}
			cfg.Body += encoded
		}
	}

	for _, d := range cfg.DataBinary {
		if cfg.Body != "" {
			cfg.Body += "&"
		}
		cfg.Body += d
	}

	for _, d := range cfg.DataRaw {
		if cfg.Body != "" {
			cfg.Body += "&"
		}
		cfg.Body += d
	}

	bodyCount := 0
	if cfg.Body != "" || cfg.BodyStdin {
		bodyCount++
	}
	if cfg.JSON != "" || cfg.JSONFile != "" {
		bodyCount++
	}
	if len(cfg.FormData) > 0 || len(cfg.FormString) > 0 {
		bodyCount++
	}
	if bodyCount > 1 {
		return nil, fmt.Errorf("cannot use -d, --json and -F simultaneously")
	}

	for _, fs := range cfg.FormString {
		cfg.FormData = append(cfg.FormData, fs)
	}

	if (cfg.Body != "" || cfg.BodyStdin || cfg.JSON != "" || cfg.JSONFile != "" || len(cfg.FormData) > 0) && cfg.Method == "GET" {
		cfg.Method = "POST"
	}

	if cfg.Body != "" && strings.HasPrefix(cfg.Body, "@-") {
		cfg.BodyStdin = true
		cfg.Body = ""
	}

	if cfg.BasicAuth != "" {
		parts := strings.SplitN(cfg.BasicAuth, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid basic auth format, expected user:password")
		}
	}

	if cfg.JSON != "" && strings.HasPrefix(cfg.JSON, "@") {
		cfg.JSONFile = cfg.JSON[1:]
		cfg.JSON = ""
	}

	for i, u := range cfg.URLs {
		cfg.URLs[i] = NormalizeURL(u)
	}

	return cfg, nil
}

func loadConfigFile(cfg *Config) error {
	data, err := os.ReadFile(cfg.ConfigFile)
	if err != nil {
		return fmt.Errorf("cannot read config file: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")

		switch key {
		case "url":
			cfg.URLs = append(cfg.URLs, value)
		case "method":
			cfg.Method = value
		case "header":
			hParts := strings.SplitN(value, ":", 2)
			if len(hParts) == 2 {
				cfg.Headers[strings.TrimSpace(hParts[0])] = append(
					cfg.Headers[strings.TrimSpace(hParts[0])],
					strings.TrimSpace(hParts[1]),
				)
			}
		case "output":
			cfg.OutputFile = value
		case "user":
			cfg.BasicAuth = value
		}
	}

	return nil
}

func NormalizeURL(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}

	if strings.HasPrefix(raw, "[") {
		return "http://" + raw
	}

	if ip := net.ParseIP(raw); ip != nil {
		if ip.To4() != nil {
			return "http://" + raw
		}
		return "http://[" + raw + "]"
	}

	if strings.Contains(raw, ":") && !strings.Contains(raw, "/") {
		host, port, err := net.SplitHostPort(raw)
		if err == nil {
			if ip := net.ParseIP(host); ip != nil {
				if ip.To4() != nil {
					return "http://" + raw
				}
				return "http://[" + host + "]:" + port
			}
		}
	}

	return "http://" + raw
}

func LoadVars(filePath string) (Variables, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		return loadJSON(filePath)
	case ".yaml", ".yml":
		return loadYAML(filePath)
	case ".env":
		return loadDotenv(filePath)
	default:
		return loadJSON(filePath)
	}
}

func loadJSON(filePath string) (Variables, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %v", err)
	}

	var vars Variables
	if err := json.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("cannot parse JSON: %v", err)
	}

	return vars, nil
}

func loadYAML(filePath string) (Variables, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %v", err)
	}

	var vars Variables
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("cannot parse YAML: %v", err)
	}

	return vars, nil
}

func loadDotenv(filePath string) (Variables, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %v", err)
	}

	vars := make(Variables)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}

		vars[key] = value
	}

	return vars, nil
}

func ParseVars(args []string) Variables {
	vars := make(Variables)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}
	return vars
}

func Substitute(input string, vars Variables) (string, error) {
	result := input
	for {
		start := strings.Index(result, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			return "", fmt.Errorf("unterminated template variable at position %d", start)
		}
		end += start

		varName := strings.TrimSpace(result[start+2 : end])
		value, ok := vars[varName]
		if !ok {
			return "", fmt.Errorf("undefined variable: %s", varName)
		}

		result = result[:start] + value + result[end+2:]
	}
	return result, nil
}

func PrintHelp() {
	fmt.Println(`iurl - A command-line HTTP client

Usage:
  iurl [options] <URL...>

HTTP Options:
  -X, --request <method>       HTTP method (default: GET)
  -H, --header <header>        HTTP header (can be specified multiple times)
  -d, --data <data>            Request body (use @- for stdin)
  --data-stdin                 Read request body from stdin
  --data-urlencode <data>      URL-encode data before sending
  --data-binary <data>         Send data as-is (binary)
  --data-raw <data>            Send data without @-interpretation
  --json <data>                JSON request data (use @file to read from file)
  --json-file <file>           JSON file path
  -F, --form <field=value>     Form field (can be specified multiple times)
  --form-string <field=val>    Form field without parsing
  -G, --get                    Convert -d to URL query parameters
  -T, --upload-file <file>     Upload file with PUT
  --compressed                 Request compressed response

Authentication:
  -u, --user <user:pass>       Basic authentication
  --anyauth                    Auto-detect best auth method
  --digest                     Use Digest authentication
  --negotiate                  Use Negotiate/Kerberos authentication
  --ntlm                       Use NTLM authentication

TLS/SSL:
  -k, --insecure               Skip TLS certificate verification
  --cacert <file>              CA certificate file
  --cert <file>                Client certificate file
  --key <file>                 Private key file
  --cert-type <type>           Certificate type (PEM, DER, ENG)
  --proxy-insecure             Skip proxy TLS verification

Proxy:
  -x, --proxy <url>            Proxy URL
  --proxy-user <user:pass>     Proxy authentication

Network:
  --resolve <host:port:addr>   Resolve hostname to IP
  --connect-to <H1:P1:H2:P2>  Connect to different host
  --connect-timeout <sec>      Connection timeout (default: 10)
  --max-time <sec>             Maximum request time (default: 30)
  -4, --ipv4                   Force IPv4
  -6, --ipv6                   Force IPv6

Redirects:
  -L, --location               Follow redirects
  --max-redirs <num>           Maximum redirects (default: 10)
  --post301                    Keep POST on 301 redirect
  --post302                    Keep POST on 302 redirect
  --post303                    Keep POST on 303 redirect
  --location-trusted           Follow redirects with credentials

HTTP Version:
  --http1.0                    Use HTTP/1.0
  --http1.1                    Use HTTP/1.1
  --http2                      Force HTTP/2 usage
  --http3                      Force HTTP/3 usage

Output:
  -o, --output <file>          Output body to file
  -O, --remote-name            Save file with name from URL
  --output-dir <dir>           Output directory
  --create-dirs                Create output directories
  --no-clobber                 Don't overwrite existing files
  -i, --include                Include response headers
  -w, --write-out <format>     Output format after request
  --no-pretty                  Disable pretty printing
  --trace <file>               Trace raw exchange to file
  --trace-ascii <file>         Trace ASCII output to file

Error Handling:
  --fail                       Fail silently on HTTP errors (4xx/5xx)
  --fail-with-body             Fail on HTTP errors but show body

Retry:
  --retry <num>                Number of retries on failure
  --retry-delay <sec>          Delay between retries (default: 1.0)
  --retry-max-time <sec>       Maximum time for retries
  --retry-connrefused          Retry on connection refused
  --retry-all-errors           Retry on all errors

Limits:
  --max-filesize <bytes>       Maximum file size to download
  --limit-rate <rate>          Limit transfer rate (e.g., 100K, 1M)

Parallel:
  --parallel                   Enable parallel transfers
  --parallel-max <num>         Maximum parallel transfers (default: 5)

Cookies:
  -b, --cookie <file>          Load cookies from file
  -c, --cookie-jar <file>      Save cookies to file

Variables:
  -V, --vars-file <file>       Variables file (JSON/YAML/dotenv)
  --var <key=value>            Set variable (can be specified multiple times)

Debug:
  -v, --verbose                Verbose output
  -I, --head                   Fetch headers only (HEAD method)
  -s, --silent                 Silent mode
  -S, --show-error             Show errors in silent mode
  --stderr <file>              Redirect stderr to file
  -K, --config <file>          Config file path

Misc:
  --version                    Show version
  -h, --help                   Show this help

IP Address Support:
  IPv4:   http://192.168.1.1/path
  IPv6:   http://[::1]/path or http://[2001:db8::1]/path
  IPv8:   http://64496.192.0.2.1/path (ASN-dot notation)

Examples:
  iurl https://api.example.com
  iurl -X POST -d '{"key":"value"}' https://api.example.com
  iurl --compressed https://api.example.com
  iurl -I https://api.example.com
  iurl -k https://self-signed.example.com
  iurl -T file.txt https://api.example.com/upload
  iurl --retry 3 https://unstable.example.com
  iurl -o out.json https://api.example.com
  iurl -O https://example.com/file.zip
  iurl --write-out "%{http_code} %{time_total}s\n" https://api.example.com
  iurl --fail https://api.example.com
  iurl -x http://proxy:8080 https://api.example.com
  iurl -4 https://api.example.com
  iurl --http2 https://api.example.com

Query & Transform:
  --query <expr>                Filter JSON response (jq-like)
  --diff                        Compare with cached response
  --show-chain                  Show redirect chain

Save & Load:
  --save <file>                 Save request profile to file
  --load <file>                 Load request profile from file

History:
  --history                     Show request history
  --tag <tag>                   Tag request (can be specified multiple times)
  --search <query>             Search history by tag or text
  --rerun <id>                 Re-execute request from history

Code Generation:
  --generate-code <lang>        Generate code (python/go/js/curl)

Monitoring:
  --watch <interval>            Monitor URL with interval (e.g., 10s, 2m)

Interactive:
  --build                       Interactive request builder
  --raw-shell <url>             Raw HTTP shell mode

Validation:
  --validate <spec-url>        Validate against OpenAPI spec
  --force                       Force execution despite validation errors

Scripts:
  --pre-script <path>           Script to run before request
  --post-script <path>          Script to run after request

Advanced:
  --adaptive-timeout <min:max>  Adaptive timeout (e.g., 0.5:30)
  --show-secrets                Show resolved secret values`)
}

type headerValue struct {
	headers map[string][]string
}

func (h *headerValue) String() string { return "" }
func (h *headerValue) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header format: %s", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	h.headers[key] = append(h.headers[key], val)
	return nil
}

type formValue struct{ data *[]string }

func (f *formValue) String() string { return "" }
func (f *formValue) Set(value string) error {
	*f.data = append(*f.data, value)
	return nil
}

type tagValue struct{ data *[]string }

func (t *tagValue) String() string { return "" }
func (t *tagValue) Set(value string) error {
	*t.data = append(*t.data, value)
	return nil
}

type formStringValue struct{ data *[]string }

func (f *formStringValue) String() string { return "" }
func (f *formStringValue) Set(value string) error {
	*f.data = append(*f.data, value)
	return nil
}

type varValue struct{ data *[]string }

func (v *varValue) String() string { return "" }
func (v *varValue) Set(value string) error {
	*v.data = append(*v.data, value)
	return nil
}

type resolveValue struct{ data *[]string }

func (r *resolveValue) String() string { return "" }
func (r *resolveValue) Set(value string) error {
	*r.data = append(*r.data, value)
	return nil
}

type connectToValue struct{ data *[]string }

func (c *connectToValue) String() string { return "" }
func (c *connectToValue) Set(value string) error {
	*c.data = append(*c.data, value)
	return nil
}

type dataURLEncodeValue struct{ data *[]string }

func (d *dataURLEncodeValue) String() string { return "" }
func (d *dataURLEncodeValue) Set(value string) error {
	*d.data = append(*d.data, value)
	return nil
}

type dataBinaryValue struct{ data *[]string }

func (d *dataBinaryValue) String() string { return "" }
func (d *dataBinaryValue) Set(value string) error {
	*d.data = append(*d.data, value)
	return nil
}

type dataRawValue struct{ data *[]string }

func (d *dataRawValue) String() string { return "" }
func (d *dataRawValue) Set(value string) error {
	*d.data = append(*d.data, value)
	return nil
}