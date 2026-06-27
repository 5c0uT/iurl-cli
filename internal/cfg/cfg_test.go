package cfg

import "testing"

func TestParseDiffFlag(t *testing.T) {
	c, err := Parse([]string{"--diff", "baseline", "https://example.com/status"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if c.Diff != "baseline" {
		t.Fatalf("Diff = %q, want %q", c.Diff, "baseline")
	}
}
