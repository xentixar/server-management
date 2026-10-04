package helpers

import "testing"

func TestIsValidHostHostname(t *testing.T) {
	ok := []string{"edge-01", "edge01.example.com", "a", "A1"}
	for _, name := range ok {
		if !IsValidHostHostname(name) {
			t.Fatalf("expected valid hostname %q", name)
		}
	}
	bad := []string{"", "bad host", "under_score", "-leading", "trailing-", stringsRepeat("a", 254)}
	for _, name := range bad {
		if IsValidHostHostname(name) {
			t.Fatalf("expected invalid hostname %q", name)
		}
	}
}

func TestIsValidTimezone(t *testing.T) {
	if !IsValidTimezone("UTC") || !IsValidTimezone("Asia/Kathmandu") {
		t.Fatal("expected valid timezones")
	}
	if IsValidTimezone("Not/AZone") {
		t.Fatal("expected invalid timezone")
	}
}

func stringsRepeat(s string, n int) string {
	b := make([]byte, 0, n*len(s))
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}
