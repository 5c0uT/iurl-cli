package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"iurl/internal/cfg"
)

type HistoryEntry struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	Headers   map[string][]string `json:"headers"`
	Body      string    `json:"body,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Status    int       `json:"status,omitempty"`
}

type History struct {
	Entries []HistoryEntry `json:"entries"`
}

func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".iurl_history"
	}
	return filepath.Join(home, ".iurl_history")
}

func LoadHistory() (*History, error) {
	path := historyPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return &History{}, nil
	}

	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return &History{}, nil
	}

	return &h, nil
}

func SaveHistory(h *History) error {
	path := historyPath()
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func AddToHistory(c *cfg.Config, status int, tags []string) error {
	h, err := LoadHistory()
	if err != nil {
		return err
	}

	id := 1
	if len(h.Entries) > 0 {
		id = h.Entries[len(h.Entries)-1].ID + 1
	}

	entry := HistoryEntry{
		ID:        id,
		Timestamp: time.Now(),
		Method:    c.Method,
		URL:       c.URLs[0],
		Headers:   c.Headers,
		Body:      c.Body,
		Tags:      tags,
		Status:    status,
	}

	h.Entries = append(h.Entries, entry)

	if len(h.Entries) > 1000 {
		h.Entries = h.Entries[len(h.Entries)-1000:]
	}

	return SaveHistory(h)
}

func SearchHistory(query string, tags []string) ([]HistoryEntry, error) {
	h, err := LoadHistory()
	if err != nil {
		return nil, err
	}

	var results []HistoryEntry
	for _, e := range h.Entries {
		if query != "" {
			if !strings.Contains(strings.ToLower(e.URL), strings.ToLower(query)) &&
				!strings.Contains(strings.ToLower(e.Method), strings.ToLower(query)) {
				continue
			}
		}

		if len(tags) > 0 {
			matched := false
			for _, t := range tags {
				for _, et := range e.Tags {
					if strings.EqualFold(t, et) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}

		results = append(results, e)
	}

	return results, nil
}

func GetHistoryEntry(id int) (*HistoryEntry, error) {
	h, err := LoadHistory()
	if err != nil {
		return nil, err
	}

	for _, e := range h.Entries {
		if e.ID == id {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("history entry %d not found", id)
}

func FormatHistoryTable(entries []HistoryEntry) string {
	if len(entries) == 0 {
		return "No history entries found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-5s %-8s %-25s %-20s %s\n", "ID", "Method", "URL", "Time", "Tags"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	for _, e := range entries {
		ts := e.Timestamp.Format("2006-01-02 15:04:05")
		url := e.URL
		if len(url) > 25 {
			url = url[:22] + "..."
		}
		tags := strings.Join(e.Tags, ",")
		sb.WriteString(fmt.Sprintf("%-5d %-8s %-25s %-20s %s\n", e.ID, e.Method, url, ts, tags))
	}

	return sb.String()
}

type Profile struct {
	URL             string              `json:"url"`
	Method          string              `json:"method"`
	Headers         map[string][]string `json:"headers,omitempty"`
	Body            string              `json:"body,omitempty"`
	JSON            string              `json:"json,omitempty"`
	BasicAuth       string              `json:"basic_auth,omitempty"`
	ConnectTimeout  float64             `json:"connect_timeout,omitempty"`
	MaxTime         float64             `json:"max_time,omitempty"`
	FollowRedirects bool                `json:"follow_redirects,omitempty"`
	MaxRedirects    int                 `json:"max_redirs,omitempty"`
	Insecure        bool                `json:"insecure,omitempty"`
	Compressed      bool                `json:"compressed,omitempty"`
}

func SaveProfile(c *cfg.Config, path string) error {
	p := Profile{
		URL:             c.URLs[0],
		Method:          c.Method,
		Headers:         c.Headers,
		Body:            c.Body,
		JSON:            c.JSON,
		BasicAuth:       c.BasicAuth,
		ConnectTimeout:  c.ConnectTimeout,
		MaxTime:         c.MaxTime,
		FollowRedirects: c.FollowRedirects,
		MaxRedirects:    c.MaxRedirects,
		Insecure:        c.Insecure,
		Compressed:      c.Compressed,
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadProfile(c *cfg.Config) error {
	data, err := os.ReadFile(c.LoadProfile)
	if err != nil {
		return fmt.Errorf("cannot read profile: %v", err)
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("cannot parse profile: %v", err)
	}

	if p.URL != "" && len(c.URLs) == 0 {
		c.URLs = []string{p.URL}
	}
	if p.Method != "" {
		c.Method = p.Method
	}
	if p.Headers != nil {
		c.Headers = p.Headers
	}
	if p.Body != "" {
		c.Body = p.Body
	}
	if p.JSON != "" {
		c.JSON = p.JSON
	}
	if p.BasicAuth != "" {
		c.BasicAuth = p.BasicAuth
	}
	if p.ConnectTimeout > 0 {
		c.ConnectTimeout = p.ConnectTimeout
	}
	if p.MaxTime > 0 {
		c.MaxTime = p.MaxTime
	}
	c.FollowRedirects = p.FollowRedirects
	if p.MaxRedirects > 0 {
		c.MaxRedirects = p.MaxRedirects
	}
	c.Insecure = p.Insecure
	c.Compressed = p.Compressed

	return nil
}

func GetDiffCache(name string) ([]byte, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".iurl_diff", name+".json")
	return os.ReadFile(path)
}

func SaveDiffCache(name string, data []byte) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".iurl_diff")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, name+".json")
	return os.WriteFile(path, data, 0644)
}
