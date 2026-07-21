package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const alphaLen = 52

type ShortUrl = string
type LongUrl = string

func short(l LongUrl) ShortUrl {
	var hash uint = 2166136261
	for _, c := range l {
		hash ^= uint(c)
		hash = hash * 16777619
	}
	state := hash
	var sb strings.Builder

	for range 5 {
		state = (state*1664525 + 1013904223)
		ix := state % alphaLen
		sb.WriteByte(alpha[ix])
	}
	return sb.String()
}

type UrlMap = map[ShortUrl]LongUrl

type SafeUrlMap struct {
	mu          sync.RWMutex
	m           UrlMap
	persistPath string
}

func (s *SafeUrlMap) Persist() error {
	var sb strings.Builder
	for key, val := range s.m {
		sb.WriteString(fmt.Sprintf("%s => %s\n", key, val))
	}
	return os.WriteFile(s.persistPath, []byte(sb.String()), 0600)
}

func (s *SafeUrlMap) Set(short ShortUrl, long LongUrl) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[short] = long

	return s.Persist()
}

func (s *SafeUrlMap) Get(short ShortUrl) (LongUrl, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, ok := s.m[short]
	return long, ok
}

func (s *SafeUrlMap) Delete(short ShortUrl) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.m, short)

	return s.Persist()
}

func (s *SafeUrlMap) handleShorten() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		originalUrl := r.FormValue("original-url")
		if originalUrl == "" {
			http.Error(w, "missing original-url", http.StatusBadRequest)
			return
		}

		shortUrl := short(originalUrl)

		if err := s.Set(shortUrl, originalUrl); err != nil {
			http.Error(w, "failed to save URL", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Short string `json:"short"`
			Long  string `json:"long"`
		}{Short: shortUrl, Long: originalUrl})
	}
}

func (s *SafeUrlMap) handleGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortUrl := r.PathValue("short")
		longUrl, ok := s.Get(shortUrl)
		if !ok {
			http.Error(w, "short url not found", http.StatusNotFound)
			return
		}
		w.Header().Add("Location", longUrl)
		http.Redirect(w, r, longUrl, http.StatusMovedPermanently)
	}
}

func (s *SafeUrlMap) handleList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Ret struct {
			Short ShortUrl `json:"short"`
			Long  LongUrl  `json:"long"`
		}
		var rets []Ret
		for key, val := range s.m {
			rets = append(rets, Ret{Short: key, Long: val})
		}
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rets)
	}
}

func (s *SafeUrlMap) handleDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortUrl := r.PathValue("short")
		if shortUrl == "" {
			http.Error(w, "missing 'short' parameter", http.StatusBadRequest)
			return
		}
		if err := s.Delete(shortUrl); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func loadMap(path string) (map[ShortUrl]LongUrl, error) {
	m := make(map[ShortUrl]LongUrl)

	input, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	for line := range strings.Lines(string(input)) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, " => ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("parse error, error in splitting: %s", line)
		}
		key, val := parts[0], strings.TrimSpace(parts[1])
		m[key] = val
	}

	return m, nil
}

// this is some kind of middleware
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hs := map[string]string{
			"Access-Control-Allow-Origin":  "http://localhost:3000",
			"Access-Control-Allow-Methods": "GET, POST, DELETE, PUT, PATCH, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		}
		for key, val := range hs {
			w.Header().Set(key, val)
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	const persistPath = "urlmap.txt"

	rawMap, err := loadMap(persistPath)
	if err != nil {
		log.Fatalf("failed to load map: %v", err)
	}

	urlStorage := &SafeUrlMap{
		m:           rawMap,
		persistPath: persistPath,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST    /shorten", urlStorage.handleShorten())
	mux.HandleFunc("DELETE  /{short}", urlStorage.handleDelete())
	mux.HandleFunc("GET     /{short}", urlStorage.handleGet())
	mux.HandleFunc("GET     /",        urlStorage.handleList())

	handlerWithCORS := withCORS(mux)

	log.Println("Server running on :8080...")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatal(err)
	}
	log.Println("Clossing server...")
}
