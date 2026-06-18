package cookiejar

import (
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestNewJar(t *testing.T) {
	jar := New()
	if jar == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetAndGetCookies(t *testing.T) {
	jar := New()
	u, _ := url.Parse("https://example.com")

	cookie := &http.Cookie{
		Name:  "session",
		Value: "abc123",
		Path:  "/",
	}

	jar.SetCookies(u, []*http.Cookie{cookie})

	cookies := jar.Cookies(u)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "session" {
		t.Errorf("cookie name = %q, want session", cookies[0].Name)
	}
	if cookies[0].Value != "abc123" {
		t.Errorf("cookie value = %q, want abc123", cookies[0].Value)
	}
}

func TestLoadFromFile(t *testing.T) {
	content := `# Netscape HTTP Cookie File
.example.com	TRUE	/	TRUE	2000000000	session	abc123
.example.com	TRUE	/	FALSE	2000000000	user	john`
	tmpFile, err := os.CreateTemp("", "cookies-*.txt")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("error writing temp file: %v", err)
	}
	tmpFile.Close()

	jar := New()
	if err := jar.LoadFromFile(tmpFile.Name()); err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}

	u, _ := url.Parse("https://example.com/path")
	cookies := jar.Cookies(u)
	if len(cookies) < 1 {
		t.Fatalf("expected at least 1 cookie, got %d", len(cookies))
	}

	found := false
	for _, c := range cookies {
		if c.Name == "session" && c.Value == "abc123" {
			found = true
			break
		}
	}
	if !found {
		t.Error("session cookie not found")
	}
}

func TestSaveToFile(t *testing.T) {
	jar := New()
	u, _ := url.Parse("https://example.com")

	cookie := &http.Cookie{
		Name:    "session",
		Value:   "abc123",
		Path:    "/",
		Domain:  "example.com",
		Expires: time.Now().Add(24 * time.Hour),
	}
	jar.SetCookies(u, []*http.Cookie{cookie})

	tmpFile, err := os.CreateTemp("", "cookies-*.txt")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if err := jar.SaveToFile(tmpFile.Name(), u); err != nil {
		t.Fatalf("SaveToFile error: %v", err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	if len(content) == 0 {
		t.Error("file is empty")
	}
}