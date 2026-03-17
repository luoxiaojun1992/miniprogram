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

type objectMeta struct {
	path string
	size int64
}

type objectStore struct {
	mu   sync.RWMutex
	data map[string]objectMeta
}

func (s *objectStore) put(key string, value objectMeta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.data[key]; ok && old.path != "" {
		_ = os.Remove(old.path)
	}
	s.data[key] = value
}

func (s *objectStore) get(key string) (objectMeta, bool) {
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

	store := &objectStore{data: map[string]objectMeta{}}
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
			tmpFile, err := os.CreateTemp("", "cos-mock-*")
			if err != nil {
				http.Error(w, "create temp file failed", http.StatusInternalServerError)
				return
			}
			defer tmpFile.Close()

			size, err := io.Copy(tmpFile, io.LimitReader(r.Body, 512*1024*1024))
			if err != nil {
				_ = os.Remove(tmpFile.Name())
				http.Error(w, "read body failed", http.StatusBadRequest)
				return
			}
			store.put(key, objectMeta{path: tmpFile.Name(), size: size})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"code":"Success","key":"%s","size":%d}`, key, size)
		case http.MethodGet:
			meta, ok := store.get(key)
			if !ok {
				http.NotFound(w, r)
				return
			}
			f, err := os.Open(meta.path)
			if err != nil {
				http.Error(w, "open object failed", http.StatusInternalServerError)
				return
			}
			defer f.Close()
			w.WriteHeader(http.StatusOK)
			_, _ = io.Copy(w, f)
		case http.MethodHead:
			meta, ok := store.get(key)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.size))
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + port
	log.Printf("COS mock server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
