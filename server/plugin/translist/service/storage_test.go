package service

import "testing"

func TestBuildTranslistKey(t *testing.T) {
	key := buildTranslistKey("demo.xlsx", "src")
	if !hasPrefix(key, "translist/") {
		t.Fatalf("key should be under translist/: %s", key)
	}
	if !contains(key, "_src_") {
		t.Fatalf("key should contain kind: %s", key)
	}
	if !hasSuffix(key, ".xlsx") {
		t.Fatalf("key should keep ext: %s", key)
	}
}

func TestResolveObjectKey(t *testing.T) {
	cases := []struct {
		key, url, want string
	}{
		{"translist/a.xlsx", "", "translist/a.xlsx"},
		{"", "http://cdn.example.com/translist/a.xlsx", "translist/a.xlsx"},
		{"/translist/a.xlsx", "", "translist/a.xlsx"},
	}
	for _, c := range cases {
		got := resolveObjectKey(c.key, c.url)
		if got != c.want {
			t.Fatalf("resolveObjectKey(%q,%q)=%q want %q", c.key, c.url, got, c.want)
		}
	}
}

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }
func hasSuffix(s, p string) bool {
	return len(s) >= len(p) && s[len(s)-len(p):] == p
}
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
