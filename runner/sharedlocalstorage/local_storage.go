package sharedlocalstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type SharedLocalStorage struct {
	ctx    context.Context
	Port   int
	mu     *sync.Mutex
	data   map[string]string
	server *http.Server
}

type Options struct {
	Port int
}

const DefaultPort = 2334

func NewSharedLocalStorage(ctx context.Context, options *Options) *SharedLocalStorage {
	port := DefaultPort
	if options != nil {
		if options.Port > 0 {
			port = options.Port
		}
	}
	return &SharedLocalStorage{
		ctx:  ctx,
		Port: port,
		mu:   &sync.Mutex{},
		data: make(map[string]string),
	}
}

func (sls *SharedLocalStorage) Start() error {
	server := &http.Server{Addr: fmt.Sprintf(":%d", sls.Port)}
	http.HandleFunc("/api-sls", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("SLS: %s %s", r.Method, r.URL.Path)
		if r.Method != "GET" && r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		} else if r.Method == "GET" {
			key := r.URL.Query().Get("key")
			if key == "" {
				http.Error(w, "key not specified", http.StatusBadRequest)
			}
			if value, err := sls.Get(key); err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(fmt.Sprintf("key not found: %s", key)))
			} else {
				data := fmt.Sprintf(`{
	"data": {
		"key": "%s",
		"value": "%s"
	}
}`, key, value)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(data))
			}
		} else if r.Method == "POST" {
			var data struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				sls.Set(data.Key, data.Value)
				w.WriteHeader(http.StatusOK)
			}
		}
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Shared local storage HTTP server ListenAndServe: %v", err)
		}
	}()
	sls.server = server
	log.Printf("Shared local storage HTTP server listening on port %d", sls.Port)
	return nil
}

func (sls *SharedLocalStorage) Stop() error {
	log.Printf("Stopping shared local storage HTTP server...")
	ctx, cancel := context.WithTimeout(sls.ctx, 5*time.Second)
	defer cancel()
	if err := sls.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down shared local storage HTTP server: %w", err)
	}
	log.Printf("Shared local storage HTTP server stopped")
	return nil
}

func (sls *SharedLocalStorage) Get(key string) (string, error) {
	sls.mu.Lock()
	defer sls.mu.Unlock()
	if value, ok := sls.data[key]; ok {
		return value, nil
	}
	return "", fmt.Errorf("key not found")
}

func (sls *SharedLocalStorage) Set(key string, value string) {
	sls.mu.Lock()
	defer sls.mu.Unlock()
	sls.data[key] = value
}

func (sls *SharedLocalStorage) Clear() {
	sls.mu.Lock()
	defer sls.mu.Unlock()
	sls.data = make(map[string]string)
}
