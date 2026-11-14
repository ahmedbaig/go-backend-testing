package testkit

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func ConfigDir(t testing.TB, files map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	for name, v := range files {
		var b []byte
		switch s := v.(type) {
		case string:
			b = []byte(s)
		case []byte:
			b = s
		default:
			var err error
			if b, err = json.MarshalIndent(v, "", "  "); err != nil {
				t.Fatalf("config %s: %v", name, err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, name+".json"), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("CONFIG_DIR", dir)
	return dir
}

func Fixture[T any](t testing.TB, name string) T {
	t.Helper()
	var v T
	b, err := os.ReadFile(filepath.Join("testdata", "fixtures", name+".json"))
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return v
}

var update = flag.Bool("update", false, "rewrite golden files")

func Golden(t testing.TB, name string, got any) {
	t.Helper()
	var b []byte
	switch v := got.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		var err error
		if b, err = json.MarshalIndent(v, "", "  "); err != nil {
			t.Fatal(err)
		}
		b = append(b, '\n')
	}
	path := filepath.Join("testdata", "golden", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s missing (run go test -update): %v", name, err)
	}
	if string(want) != string(b) {
		t.Fatalf("golden %s mismatch\n--- want\n%s\n--- got\n%s", name, want, b)
	}
}
