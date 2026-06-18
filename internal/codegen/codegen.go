package codegen

import (
	"fmt"
	"strings"

	"iurl/internal/cfg"
)

func Generate(c *cfg.Config, lang string) string {
	switch lang {
	case "curl":
		return genCurl(c)
	case "python":
		return genPython(c)
	case "go":
		return genGo(c)
	case "js", "javascript":
		return genJS(c)
	default:
		return fmt.Sprintf("# Unsupported language: %s\n# Supported: curl, python, go, js", lang)
	}
}

func genCurl(c *cfg.Config) string {
	var parts []string
	parts = append(parts, "curl")

	if c.Method != "GET" {
		parts = append(parts, "-X", c.Method)
	}

	for key, values := range c.Headers {
		for _, v := range values {
			parts = append(parts, "-H", fmt.Sprintf("%s: %s", key, v))
		}
	}

	if c.Body != "" {
		parts = append(parts, "-d", shellQuote(c.Body))
	}

	if c.JSON != "" {
		parts = append(parts, "-H", "Content-Type: application/json", "-d", shellQuote(c.JSON))
	}

	if c.BasicAuth != "" {
		parts = append(parts, "-u", c.BasicAuth)
	}

	parts = append(parts, shellQuote(c.URLs[0]))

	return strings.Join(parts, " ")
}

func genPython(c *cfg.Config) string {
	var lines []string
	lines = append(lines, "import requests")
	lines = append(lines, "")

	headers := "{}"
	if len(c.Headers) > 0 {
		var h []string
		for key, values := range c.Headers {
			for _, v := range values {
				h = append(h, fmt.Sprintf("    \"%s\": \"%s\",", key, v))
			}
		}
		headers = "{\n" + strings.Join(h, "\n") + "\n}"
	}

	data := "None"
	if c.Body != "" {
		data = fmt.Sprintf("\"%s\"", escapeString(c.Body))
	} else if c.JSON != "" {
		data = fmt.Sprintf("%s", c.JSON)
	}

	lines = append(lines, fmt.Sprintf("url = \"%s\"", c.URLs[0]))
	lines = append(lines, fmt.Sprintf("headers = %s", headers))
	lines = append(lines, fmt.Sprintf("data = %s", data))
	lines = append(lines, "")

	method := strings.ToLower(c.Method)
	if method == "" {
		method = "get"
	}

	lines = append(lines, fmt.Sprintf("response = requests.%s(url, headers=headers, data=data)", method))
	lines = append(lines, "print(response.status_code)")
	lines = append(lines, "print(response.text)")

	return strings.Join(lines, "\n")
}

func genGo(c *cfg.Config) string {
	var lines []string
	lines = append(lines, "package main")
	lines = append(lines, "")
	lines = append(lines, "import (")
	lines = append(lines, "\t\"fmt\"")
	lines = append(lines, "\t\"io\"")
	lines = append(lines, "\t\"net/http\"")
	lines = append(lines, "\t\"strings\"")
	lines = append(lines, ")")
	lines = append(lines, "")
	lines = append(lines, "func main() {")

	body := "nil"
	if c.Body != "" || c.JSON != "" {
		b := c.Body
		if c.JSON != "" {
			b = c.JSON
		}
		lines = append(lines, fmt.Sprintf("\tbody := strings.NewReader(`%s`)", b))
		lines = append(lines, "\treq, _ := http.NewRequest(\"%s\", \"%s\", body)", c.Method, c.URLs[0])
	} else {
		lines = append(lines, fmt.Sprintf("\treq, _ := http.NewRequest(\"%s\", \"%s\", nil)", c.Method, c.URLs[0]))
	}

	for key, values := range c.Headers {
		for _, v := range values {
			lines = append(lines, fmt.Sprintf("\treq.Header.Set(\"%s\", \"%s\")", key, v))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "\tresp, _ := http.DefaultClient.Do(req)")
	lines = append(lines, "\tdefer resp.Body.Close()")
	lines = append(lines, "\tdata, _ := io.ReadAll(resp.Body)")
	lines = append(lines, "\tfmt.Println(string(data))")
	lines = append(lines, "}")
	_ = body

	return strings.Join(lines, "\n")
}

func genJS(c *cfg.Config) string {
	var lines []string
	lines = append(lines, "const response = await fetch(\""+c.URLs[0]+"\", {")
	lines = append(lines, fmt.Sprintf("  method: \"%s\",", c.Method))

	if len(c.Headers) > 0 {
		lines = append(lines, "  headers: {")
		for key, values := range c.Headers {
			for _, v := range values {
				lines = append(lines, fmt.Sprintf("    \"%s\": \"%s\",", key, v))
			}
		}
		lines = append(lines, "  },")
	}

	if c.Body != "" || c.JSON != "" {
		b := c.Body
		if c.JSON != "" {
			b = c.JSON
		}
		lines = append(lines, fmt.Sprintf("  body: %s,", escapeJSBody(b)))
	}

	lines = append(lines, "});")
	lines = append(lines, "")
	lines = append(lines, "const data = await response.json();")
	lines = append(lines, "console.log(data);")

	return strings.Join(lines, "\n")
}

func shellQuote(s string) string {
	if !strings.ContainsAny(s, " \t\n\"'") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func escapeJSBody(s string) string {
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return s
	}
	return "\"" + escapeString(s) + "\""
}