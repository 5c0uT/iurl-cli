package repl

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"

	"iurl/internal/cfg"
	fmtc "iurl/internal/fmt"
	httpc "iurl/internal/http"
)

type Session struct {
	baseURL      string
	method       string
	headers      map[string][]string
	body         string
	json         string
	lastResponse *httpc.Result
	rl           *readline.Instance
}

func New(stdout, stderr *os.File) *Session {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("set",
			readline.PcItem("base"),
			readline.PcItem("method",
				readline.PcItem("GET"),
				readline.PcItem("POST"),
				readline.PcItem("PUT"),
				readline.PcItem("PATCH"),
				readline.PcItem("DELETE"),
				readline.PcItem("HEAD"),
				readline.PcItem("OPTIONS"),
			),
		),
		readline.PcItem("header",
			readline.PcItem("add"),
			readline.PcItem("remove"),
		),
		readline.PcItem("body",
			readline.PcItem("set"),
			readline.PcItem("json"),
			readline.PcItem("clear"),
		),
		readline.PcItem("send",
			readline.PcItem("--verbose"),
			readline.PcItem("-v"),
		),
		readline.PcItem("response",
			readline.PcItem("save"),
			readline.PcItem("headers"),
		),
		readline.PcItem("show"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("quit"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "iurl> ",
		HistoryFile:     historyPath(),
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		fmt.Fprintf(stderr, "Error initializing readline: %v\n", err)
		os.Exit(1)
	}

	return &Session{
		baseURL: "",
		method:  "GET",
		headers: make(map[string][]string),
		rl:      rl,
	}
}

func (s *Session) Run() {
	defer s.rl.Close()

	fmt.Fprintln(s.rl, "iurl interactive mode. Type 'help' for commands, 'exit' to quit.")

	for {
		line, err := s.rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		args := strings.Fields(line)
		cmd := args[0]
		cmdArgs := args[1:]

		switch cmd {
		case "exit", "quit":
			return
		case "help":
			s.printHelp()
		case "set":
			s.handleSet(cmdArgs)
		case "header":
			s.handleHeader(cmdArgs)
		case "body":
			s.handleBody(cmdArgs)
		case "send":
			s.handleSend(cmdArgs)
		case "response":
			s.handleResponse(cmdArgs)
		case "show":
			s.showConfig()
		default:
			fmt.Fprintf(s.rl, "Unknown command: %s. Type 'help' for available commands.\n", cmd)
		}
	}
}

func (s *Session) printHelp() {
	fmt.Fprintln(s.rl, `Commands:
  set base <url>           Set base URL
  set method <method>      Set HTTP method (GET, POST, etc.)
  header add <name> <value>  Add header
  header remove <name>     Remove header
  body set <data>          Set request body
  body json <data>         Set JSON body
  body clear               Clear body
  send [path]              Send request (path is appended to base URL)
  send --verbose [path]    Send request with verbose output
  response save <file>     Save last response body to file
  response headers         Show last response headers
  show                     Show current configuration
  help                     Show this help
  exit                     Exit interactive mode

Use Tab for auto-completion.`)
}

func (s *Session) handleSet(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(s.rl, "Usage: set base <url> | set method <method>")
		return
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	switch key {
	case "base":
		s.baseURL = value
		fmt.Fprintf(s.rl, "Base URL set to: %s\n", value)
	case "method":
		s.method = strings.ToUpper(value)
		fmt.Fprintf(s.rl, "Method set to: %s\n", s.method)
	default:
		fmt.Fprintf(s.rl, "Unknown setting: %s\n", key)
	}
}

func (s *Session) handleHeader(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(s.rl, "Usage: header add <name> <value> | header remove <name>")
		return
	}

	action := args[0]
	name := args[1]

	switch action {
	case "add":
		if len(args) < 3 {
			fmt.Fprintln(s.rl, "Usage: header add <name> <value>")
			return
		}
		value := strings.Join(args[2:], " ")
		s.headers[name] = append(s.headers[name], value)
		fmt.Fprintf(s.rl, "Header added: %s: %s\n", name, value)
	case "remove":
		delete(s.headers, name)
		fmt.Fprintf(s.rl, "Header removed: %s\n", name)
	default:
		fmt.Fprintf(s.rl, "Unknown header action: %s\n", action)
	}
}

func (s *Session) handleBody(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(s.rl, "Usage: body set <data> | body json <data> | body clear")
		return
	}

	action := args[0]
	switch action {
	case "set":
		if len(args) < 2 {
			fmt.Fprintln(s.rl, "Usage: body set <data>")
			return
		}
		s.body = strings.Join(args[1:], " ")
		s.json = ""
		fmt.Fprintln(s.rl, "Body set")
	case "json":
		if len(args) < 2 {
			fmt.Fprintln(s.rl, "Usage: body json <data>")
			return
		}
		s.json = strings.Join(args[1:], " ")
		s.body = ""
		fmt.Fprintln(s.rl, "JSON body set")
	case "clear":
		s.body = ""
		s.json = ""
		fmt.Fprintln(s.rl, "Body cleared")
	default:
		fmt.Fprintf(s.rl, "Unknown body action: %s\n", action)
	}
}

func (s *Session) handleSend(args []string) {
	if s.baseURL == "" {
		fmt.Fprintln(s.rl, "Error: base URL not set. Use 'set base <url>' first.")
		return
	}

	verbose := false
	path := ""
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		} else {
			path = arg
		}
	}

	url := s.baseURL
	if path != "" {
		if strings.HasPrefix(path, "/") {
			url = s.baseURL + path
		} else {
			url = s.baseURL + "/" + path
		}
	}

	c := &cfg.Config{
		URLs:    []string{url},
		Method:  s.method,
		Headers: make(map[string][]string),
	}
	for k, v := range s.headers {
		c.Headers[k] = v
	}
	c.Body = s.body
	c.JSON = s.json
	c.Verbose = verbose

	req, err := httpc.NewRequest(c)
	if err != nil {
		fmt.Fprintf(s.rl, "Error creating request: %v\n", err)
		return
	}

	result, err := httpc.Do(req, c)
	if err != nil {
		fmt.Fprintf(s.rl, "Error: %v\n", err)
		return
	}
	resp := result.Response
	defer resp.Body.Close()

	s.lastResponse = result

	if verbose {
		fmt.Fprintf(s.rl, "* %s %s %s\n", c.Method, c.URLs[0], "HTTP/1.1")
		fmt.Fprintf(s.rl, "< %s\n", resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Fprintf(s.rl, "< %s: %s\n", key, value)
			}
		}
		fmt.Fprintf(s.rl, "* Request time: %dms\n", result.Duration.Milliseconds())
	}

	fmt.Fprintf(s.rl, "%s\n", resp.Status)

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		err = fmtc.PrettyPrintJSON(resp.Body, s.rl)
		if err != nil {
			fmt.Fprintf(s.rl, "Warning: could not pretty print JSON: %v\n", err)
			fmtc.CopyRaw(resp.Body, s.rl)
		}
	} else {
		fmtc.CopyRaw(resp.Body, s.rl)
	}
	fmt.Fprintln(s.rl)
}

func (s *Session) handleResponse(args []string) {
	if s.lastResponse == nil {
		fmt.Fprintln(s.rl, "No previous response. Use 'send' first.")
		return
	}

	if len(args) < 1 {
		fmt.Fprintln(s.rl, "Usage: response save <file> | response headers")
		return
	}

	action := args[0]
	switch action {
	case "save":
		if len(args) < 2 {
			fmt.Fprintln(s.rl, "Usage: response save <file>")
			return
		}
		filePath := args[1]
		f, err := os.Create(filePath)
		if err != nil {
			fmt.Fprintf(s.rl, "Error creating file: %v\n", err)
			return
		}
		defer f.Close()

		resp := s.lastResponse.Response
		if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			fmtc.PrettyPrintJSON(resp.Body, f)
		} else {
			io.Copy(f, resp.Body)
		}
		fmt.Fprintf(s.rl, "Response saved to %s\n", filePath)
	case "headers":
		resp := s.lastResponse.Response
		fmt.Fprintf(s.rl, "%s\n", resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Fprintf(s.rl, "%s: %s\n", key, value)
			}
		}
	default:
		fmt.Fprintf(s.rl, "Unknown response action: %s\n", action)
	}
}

func (s *Session) showConfig() {
	fmt.Fprintf(s.rl, "Base URL:  %s\n", s.baseURL)
	fmt.Fprintf(s.rl, "Method:    %s\n", s.method)
	fmt.Fprintln(s.rl, "Headers:")
	for k, v := range s.headers {
		for _, val := range v {
			fmt.Fprintf(s.rl, "  %s: %s\n", k, val)
		}
	}
	if s.body != "" {
		fmt.Fprintln(s.rl, "Body: (set)")
	}
	if s.json != "" {
		fmt.Fprintln(s.rl, "JSON: (set)")
	}
}

func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".iurl_history"
	}
	return home + string(os.PathSeparator) + ".iurl_history"
}