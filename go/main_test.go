package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestShort(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"simple url", "https://example.com"},
		{"longer url", "https://example.com/very/long/path?query=value&another=123"},
		{"empty string", ""},
		{"unicode", "https://example.com/路径"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := short(tt.url)

			if len(result) != 5 {
				t.Errorf("short() returned %d chars, want 5", len(result))
			}

			for i, c := range result {
				if !strings.ContainsRune(alpha, c) {
					t.Errorf("short() returned invalid char %q at index %d", c, i)
				}
			}

			result2 := short(tt.url)
			if result != result2 {
				t.Errorf("short() not deterministic: %q != %q", result, result2)
			}
		})
	}
}

func TestShortDifferentInputs(t *testing.T) {
	urls := []string{
		"https://example.com",
		"https://google.com",
		"https://github.com",
		"https://stackoverflow.com",
	}

	seen := make(map[string]bool)
	for _, url := range urls {
		result := short(url)
		if seen[result] {
			t.Logf("collision for %q -> %s (not necessarily a bug, but noted)", url, result)
		}
		seen[result] = true
	}
}

func TestSafeUrlMap_SetGet(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/urlmap.txt"

	sm := &SafeUrlMap{m: make(UrlMap), persistPath: path}

	sm.Set("abc", "https://example.com")

	val, ok := sm.Get("abc")
	if !ok {
		t.Fatal("Get() returned false for existing key")
	}
	if val != "https://example.com" {
		t.Errorf("Get() = %q, want %q", val, "https://example.com")
	}

	_, ok = sm.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for nonexistent key")
	}

	if _, err := os.ReadFile(path); err != nil {
		t.Fatalf("persis file not created: %v", err)
	}
}

func TestLoadMap(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/urlmap.txt"
		os.WriteFile(path, []byte("abc => https://example.com\ndef => https://google.com\n"), 0600)

		m, err := loadMap(path)
		if err != nil {
			t.Fatalf("loadMap() error: %v", err)
		}
		if len(m) != 2 {
			t.Errorf("loadMap() returned %d entries, want 2", len(m))
		}
		if m["abc"] != "https://example.com" {
			t.Errorf("m[abc] = %q, want %q", m["abc"], "https://example.com")
		}
		if m["def"] != "https://google.com" {
			t.Errorf("m[def] = %q, want %q", m["def"], "https://google.com")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		m, err := loadMap("/nonexistent/urlmap.txt")
		if err != nil {
			t.Fatalf("loadMap() error for missing file: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("loadMap() returned %d entries for missing file, want 0", len(m))
		}
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/urlmap.txt"
		os.WriteFile(path, []byte(""), 0600)

		m, err := loadMap(path)
		if err != nil {
			t.Fatalf("loadMap() error for empty file: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("loadMap() returned %d entries for empty file, want 0", len(m))
		}
	})

	t.Run("malformed line", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/urlmap.txt"
		os.WriteFile(path, []byte("bad line without arrow\n"), 0600)

		_, err := loadMap(path)
		if err == nil {
			t.Error("loadMap() should error on malformed line")
		}
	})

	t.Run("ignores blank lines", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/urlmap.txt"
		os.WriteFile(path, []byte("\nabc => https://example.com\n\n"), 0600)

		m, err := loadMap(path)
		if err != nil {
			t.Fatalf("loadMap() error: %v", err)
		}
		if len(m) != 1 {
			t.Errorf("loadMap() returned %d entries, want 1", len(m))
		}
	})
}

func TestHandleShorten(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/urlmap.txt"
	sm := &SafeUrlMap{m: make(UrlMap), persistPath: path}

	t.Run("success", func(t *testing.T) {
		body := "original-url=https://example.com"
		req := httptest.NewRequest("POST", "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		sm.handleShorten()(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		ct := w.Header().Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type = %q, want JSON", ct)
		}

		var resp struct {
			Short string `json:"short"`
			Long  string `json:"long"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Long != "https://example.com" {
			t.Errorf("response Long = %q, want %q", resp.Long, "https://example.com")
		}
		if resp.Short == "" {
			t.Error("response Short is empty")
		}
	})

	t.Run("missing param", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/shorten", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		sm.handleShorten()(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleGet(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/urlmap.txt"
	sm := &SafeUrlMap{m: UrlMap{"abc": "https://example.com"}, persistPath: path}

	t.Run("found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/abc", nil)
		req.SetPathValue("short", "abc")
		w := httptest.NewRecorder()

		sm.handleGet()(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("status = %d, want %d", w.Code, http.StatusMovedPermanently)
		}
		loc := w.Header().Get("Location")
		if loc != "https://example.com" {
			t.Errorf("Location = %q, want %q", loc, "https://example.com")
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		req.SetPathValue("short", "nonexistent")
		w := httptest.NewRecorder()

		sm.handleGet()(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestHandleList(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/urlmap.txt"

	t.Run("with entries", func(t *testing.T) {
		sm := &SafeUrlMap{m: UrlMap{
			"abc": "https://example.com",
			"def": "https://google.com",
		}, persistPath: path}

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		sm.handleList()(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var entries []struct {
			Short string `json:"short"`
			Long  string `json:"long"`
		}
		if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(entries) != 2 {
			t.Errorf("returned %d entries, want 2", len(entries))
		}
	})

	t.Run("empty", func(t *testing.T) {
		sm := &SafeUrlMap{m: make(UrlMap), persistPath: path}

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		sm.handleList()(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var entries []struct {
			Short string `json:"short"`
			Long  string `json:"long"`
		}
		if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("returned %d entries, want 0", len(entries))
		}
	})
}
