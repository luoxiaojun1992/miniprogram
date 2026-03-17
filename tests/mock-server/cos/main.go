package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type objectStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func (s *objectStore) put(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *objectStore) get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func main() {
	port := os.Getenv("COS_PORT")
	if port == "" {
		port = "9000"
	}

	store := &objectStore{data: map[string][]byte{}}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.NotFound(w, r)
			return
		}
		objectPath := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.SplitN(objectPath, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			http.Error(w, "invalid cos path", http.StatusBadRequest)
			return
		}
		key := parts[0] + "/" + parts[1]

		switch r.Method {
		case http.MethodPut:
			body, err := io.ReadAll(io.LimitReader(r.Body, 512*1024*1024))
			if err != nil {
				http.Error(w, "read body failed", http.StatusBadRequest)
				return
			}
			store.put(key, body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"code":"Success","key":"%s","size":%d}`, key, len(body))
		case http.MethodGet:
			data, ok := store.get(key)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		case http.MethodHead:
			if _, ok := store.get(key); !ok {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + port
	log.Printf("COS mock server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
