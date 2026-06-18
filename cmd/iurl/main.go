package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"iurl/internal/cfg"
	"iurl/internal/codegen"
	"iurl/internal/cookiejar"
	fmtc "iurl/internal/fmt"
	httpc "iurl/internal/http"
	"iurl/internal/interactive"
	"iurl/internal/query"
	"iurl/internal/repl"
	"iurl/internal/storage"
)

func main() {
	c, err := cfg.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if c.Version {
		fmt.Printf("iurl %s\n", cfg.Version)
		os.Exit(0)
	}

	if c.Help {
		cfg.PrintHelp()
		os.Exit(0)
	}

	if c.Build {
		newCfg, err := interactive.Build()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		c = newCfg
	}

	if c.History {
		h, err := storage.LoadHistory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		entries := h.Entries
		if len(entries) > 50 {
			entries = entries[len(entries)-50:]
		}
		fmt.Print(storage.FormatHistoryTable(entries))
		os.Exit(0)
	}

	if c.Search != "" || len(c.Tag) > 0 {
		entries, err := storage.SearchHistory(c.Search, c.Tag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(storage.FormatHistoryTable(entries))
		os.Exit(0)
	}

	if c.Rerun != "" {
		id := 0
		fmt.Sscanf(c.Rerun, "%d", &id)
		entry, err := storage.GetHistoryEntry(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		c.URLs = []string{entry.URL}
		c.Method = entry.Method
		c.Headers = entry.Headers
		c.Body = entry.Body
	}

	if c.LoadProfile != "" {
		if err := loadProfileFromMain(c); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if len(c.URLs) == 0 && c.RawShell == "" {
		session := repl.New(os.Stdout, os.Stderr)
		session.Run()
		return
	}

	if c.RawShell != "" {
		runRawShell(c)
		return
	}

	vars := make(cfg.Variables)
	if c.VarsFile != "" {
		fileVars, err := cfg.LoadVars(c.VarsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading variables: %v\n", err)
			os.Exit(1)
		}
		for k, v := range fileVars {
			vars[k] = v
		}
	}
	cliVars := cfg.ParseVars(c.Vars)
	for k, v := range cliVars {
		vars[k] = v
	}

	var jar *cookiejar.Jar
	if c.CookieJarRead != "" || c.CookieJarWrite != "" {
		jar = cookiejar.New()
		if c.CookieJarRead != "" {
			if err := jar.LoadFromFile(c.CookieJarRead); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading cookies: %v\n", err)
				os.Exit(1)
			}
		}
	}

	for i, rawURL := range c.URLs {
		url := rawURL
		if len(vars) > 0 {
			var err error
			url, err = cfg.Substitute(url, vars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error substituting variables in URL: %v\n", err)
				os.Exit(1)
			}
		}

		singleCfg := *c
		singleCfg.URLs = []string{url}

		if len(c.URLs) > 1 {
			fmt.Fprintf(os.Stdout, "\n[%d/%d] %s\n", i+1, len(c.URLs), url)
		}

		exitCode := executeRequest(&singleCfg, jar, vars)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}
}

func executeRequest(c *cfg.Config, jar *cookiejar.Jar, vars cfg.Variables) int {
	rawURL := c.URLs[0]

	req, err := httpc.NewRequest(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return 1
	}

	if len(vars) > 0 {
		for key, values := range req.Header {
			for i, value := range values {
				substituted, err := cfg.Substitute(value, vars)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error substituting variables in header %s: %v\n", key, err)
					return 1
				}
				values[i] = substituted
			}
		}
	}

	if c.Compressed {
		// Go net/http automatically adds Accept-Encoding and decompresses
		// when DisableCompression=false and no Accept-Encoding is set
	}

	if jar != nil {
		for _, cookie := range jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	result, err := httpc.Do(req, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	resp := result.Response
	defer resp.Body.Close()

	if jar != nil && c.CookieJarWrite != "" {
		for _, setCookie := range resp.Cookies() {
			jar.SetCookies(req.URL, []*http.Cookie{setCookie})
		}
		if err := jar.SaveToFile(c.CookieJarWrite, req.URL); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving cookies: %v\n", err)
		}
	}

	if c.Fail && resp.StatusCode >= 400 {
		return 1
	}

	var bodyData []byte
	if c.WriteOut != "" || c.RemoteName || c.MaxFilesize > 0 || c.LimitRate != "" || c.FailWithBody || c.Query != "" || c.Diff != "" {
		bodyData, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	} else {
		bodyData = nil
	}

	if c.FailWithBody && resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "%s\n", string(bodyData))
		return 1
	}

	if c.Watch != "" {
		return runWatch(c, jar, vars)
	}

	if c.SaveProfile != "" {
		if err := storage.SaveProfile(c, c.SaveProfile); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving profile: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "Profile saved to %s\n", c.SaveProfile)
		return 0
	}

	tags := c.Tag
	storage.AddToHistory(c, resp.StatusCode, tags)

	if c.Query != "" {
		result, err := query.Apply(bodyData, c.Query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
			return 1
		}
		fmt.Println(string(result))
		return 0
	}

	if c.Diff != "" {
		cached, err := storage.GetDiffCache(c.Diff)
		if err != nil {
			storage.SaveDiffCache(c.Diff, bodyData)
			fmt.Fprintf(os.Stderr, "Cached response as baseline for '%s'\n", c.Diff)
			return 0
		}
		printDiff(cached, bodyData)
		storage.SaveDiffCache(c.Diff, bodyData)
		return 0
	}

	var traceData bytes.Buffer
	if c.Verbose || c.Trace != "" || c.TraceASCII != "" {
		traceData.WriteString(fmt.Sprintf("* %s %s %s\n", c.Method, rawURL, "HTTP/1.1"))
		for key, values := range req.Header {
			for _, value := range values {
				traceData.WriteString(fmt.Sprintf("> %s: %s\n", key, value))
			}
		}
		traceData.WriteString(">\n")
		if c.Body != "" {
			traceData.WriteString(fmt.Sprintf("> %s\n", c.Body))
		}
		if c.JSON != "" {
			traceData.WriteString(fmt.Sprintf("> %s\n", c.JSON))
		}

		traceData.WriteString(fmt.Sprintf("< %s\n", resp.Status))
		for key, values := range resp.Header {
			for _, value := range values {
				traceData.WriteString(fmt.Sprintf("< %s: %s\n", key, value))
			}
		}
		traceData.WriteString("< \n")
		traceData.WriteString(fmt.Sprintf("* Request time: %dms\n", result.Duration.Milliseconds()))
	}

	if c.Verbose {
		os.Stderr.Write(traceData.Bytes())
	}

	if c.Trace != "" {
		f, err := os.Create(c.Trace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating trace file: %v\n", err)
			return 1
		}
		defer f.Close()
		f.Write(traceData.Bytes())
	}

	if c.TraceASCII != "" {
		f, err := os.Create(c.TraceASCII)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating trace-ascii file: %v\n", err)
			return 1
		}
		defer f.Close()
		f.Write(traceData.Bytes())
	}

	var bodyOut io.Writer = os.Stdout
	if c.IncludeHeaders {
		fmt.Printf("%s\n", resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("%s\n", resp.Status)
	}

	if c.OutputFile != "" || c.RemoteName {
		outputPath := c.OutputFile
		if c.RemoteName {
			outputPath = getFilenameFromURL(rawURL, resp)
		}
		if c.OutputDir != "" {
			outputPath = filepath.Join(c.OutputDir, outputPath)
		}
		if c.CreateDirs {
			dir := filepath.Dir(outputPath)
			if dir != "" {
				os.MkdirAll(dir, 0755)
			}
		}
		if c.NoClobber {
			if _, err := os.Stat(outputPath); err == nil {
				fmt.Fprintf(os.Stderr, "File already exists: %s\n", outputPath)
				return 1
			}
		}
		f, err := os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			return 1
		}
		defer f.Close()
		bodyOut = f
	}

	if c.Method == "HEAD" {
		return 0
	}

	contentType := resp.Header.Get("Content-Type")
	if len(bodyData) > 0 {
		if strings.Contains(contentType, "application/json") && !c.NoPretty {
			var bodyBuf bytes.Buffer
			bodyBuf.Write(bodyData)
			err = fmtc.PrettyPrintJSON(&bodyBuf, bodyOut)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not pretty print JSON: %v\n", err)
				bodyOut.Write(bodyData)
			}
		} else {
			bodyOut.Write(bodyData)
		}
	} else {
		if strings.Contains(contentType, "application/json") && !c.NoPretty {
			var bodyBuf bytes.Buffer
			tee := io.TeeReader(resp.Body, &bodyBuf)
			err = fmtc.PrettyPrintJSON(tee, bodyOut)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not pretty print JSON: %v\n", err)
				bodyBuf.WriteTo(bodyOut)
			}
		} else {
			fmtc.CopyRaw(resp.Body, bodyOut)
		}
	}

	if c.WriteOut != "" {
		fmt.Print(formatWriteOut(c.WriteOut, resp, result.Duration))
	}

	if c.GenerateCode != "" {
		fmt.Fprintf(os.Stderr, "\n--- Generated code (%s) ---\n", c.GenerateCode)
		fmt.Println(codegen.Generate(c, c.GenerateCode))
	}

	return 0
}

func runWatch(c *cfg.Config, jar *cookiejar.Jar, vars cfg.Variables) int {
	interval, err := time.ParseDuration(c.Watch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid interval: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Watching %s every %s (Ctrl+C to stop)\n", c.URLs[0], interval)

	var prevBody []byte
	first := true

	for {
		req, err := httpc.NewRequest(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			return 1
		}

		if jar != nil {
			for _, cookie := range jar.Cookies(req.URL) {
				req.AddCookie(cookie)
			}
		}

		result, err := httpc.Do(req, c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] Error: %v\n", time.Now().Format("15:04:05"), err)
			time.Sleep(interval)
			continue
		}

		body, _ := io.ReadAll(result.Response.Body)
		result.Response.Body.Close()

		if !first && prevBody != nil {
			if string(body) != string(prevBody) {
				fmt.Fprintf(os.Stderr, "\n[%s] Change detected! (status: %d)\n", time.Now().Format("15:04:05"), result.Response.StatusCode)
				fmt.Println(string(body))
				notify()
			}
		} else {
			fmt.Fprintf(os.Stderr, "[%s] Initial response (status: %d)\n", time.Now().Format("15:04:05"), result.Response.StatusCode)
		}

		prevBody = body
		first = false
		time.Sleep(interval)
	}
}

func notify() {
	fmt.Fprintf(os.Stderr, "\a")
}

func loadProfileFromMain(c *cfg.Config) error {
	return storage.LoadProfile(c)
}

func runRawShell(c *cfg.Config) {
	fmt.Printf("Connected to %s (type 'exit' to quit)\n", c.RawShell)

	buf := make([]byte, 4096)
	for {
		fmt.Print("> ")
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			break
		}
		line := strings.TrimSpace(string(buf[:n]))
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		fmt.Printf("HTTP> %s\n", line)
	}
}

func printDiff(old, new []byte) {
	oldStr := strings.TrimSpace(string(old))
	newStr := strings.TrimSpace(string(new))

	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	changes := 0
	for i := 0; i < maxLen; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			changes++
			if i < len(oldLines) {
				fmt.Printf("\033[31m- %s\033[0m\n", oldLine)
			}
			if i < len(newLines) {
				fmt.Printf("\033[32m+ %s\033[0m\n", newLine)
			}
		}
	}

	if changes == 0 {
		fmt.Println("No differences found.")
	} else {
		fmt.Printf("\n%d difference(s) found.\n", changes)
	}
}

func getFilenameFromURL(rawURL string, resp *http.Response) string {
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if idx := strings.Index(cd, "filename="); idx != -1 {
			filename := cd[idx+9:]
			filename = strings.Trim(filename, "\"")
			if filename != "" {
				return filename
			}
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "output"
	}

	path := parsed.Path
	if path == "" || path == "/" {
		return "index.html"
	}

	basename := filepath.Base(path)
	if idx := strings.Index(basename, "?"); idx != -1 {
		basename = basename[:idx]
	}

	return basename
}

func formatWriteOut(format string, resp *http.Response, duration time.Duration) string {
	result := format
	result = strings.ReplaceAll(result, "%{http_code}", fmt.Sprintf("%d", resp.StatusCode))
	result = strings.ReplaceAll(result, "%{http_content_type}", resp.Header.Get("Content-Type"))
	result = strings.ReplaceAll(result, "%{time_total}", fmt.Sprintf("%.3f", duration.Seconds()))
	result = strings.ReplaceAll(result, "%{time_connect}", "0.000")
	result = strings.ReplaceAll(result, "%{size_download}", fmt.Sprintf("%d", resp.ContentLength))
	result = strings.ReplaceAll(result, "%{url_effective}", resp.Request.URL.String())
	result = strings.ReplaceAll(result, "%{remote_ip}", resp.Request.URL.Hostname())
	result = strings.ReplaceAll(result, "%{remote_port}", resp.Request.URL.Port())
	return result
}

func hasHeader(headers http.Header, key string) bool {
	for k := range headers {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}

type bufioReader = io.Reader