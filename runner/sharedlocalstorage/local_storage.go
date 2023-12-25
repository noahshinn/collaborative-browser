package sharedlocalstorage

import (
	"collaborativebrowser/trajectory"
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
	data   map[string]any
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
		data: make(map[string]any),
	}
}

type BrowserActionType string

const (
	BrowserActionTypeClick    BrowserActionType = "click"
	BrowserActionTypeSendKeys BrowserActionType = "send_keys"
)

type TrajectoryItemPut struct {
	ItemType BrowserActionType `json:"item_type"`
	ID       string            `json:"id"`

	// for send_keys
	Text string `json:"text"`
}

func handleTraj(sls *SharedLocalStorage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("TRAJ: %s %s", r.Method, r.URL.Path)
		if r.Method != "PUT" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var data *TrajectoryItemPut
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var item *trajectory.TrajectoryItem
		switch data.ItemType {
		case BrowserActionTypeClick:
			if data.ID == "" {
				http.Error(w, "id is required for the click action", http.StatusBadRequest)
				return
			}
			item = trajectory.NewBrowserClickAction(data.ID)
		case BrowserActionTypeSendKeys:
			if data.ID == "" {
				http.Error(w, "id is required for the send_keys action", http.StatusBadRequest)
				return
			} else if data.Text == "" {
				http.Error(w, "text is required for the send_keys action", http.StatusBadRequest)
				return
			}
			item = trajectory.NewBrowserSendKeysAction(data.ID, data.Text)
		default:
			http.Error(w, fmt.Sprintf("unknown item type %s", data.ItemType), http.StatusBadRequest)
			return
		}
		if err := sls.AddItemToTrajectory(item); err != nil {
			http.Error(w, fmt.Sprintf("error adding item to trajectory: %s", err.Error()), http.StatusInternalServerError)
		}
	})
}

func (sls *SharedLocalStorage) ReadTrajectory() (*trajectory.Trajectory, error) {
	if items, err := sls.Get("traj"); err != nil || items == nil {
		return &trajectory.Trajectory{
			Items: []*trajectory.TrajectoryItem{},
		}, nil
	} else if traj, ok := items.(*trajectory.Trajectory); !ok {
		return nil, fmt.Errorf("traj is not a trajectory.Trajectory")
	} else {
		return traj, nil
	}
}

func (sls *SharedLocalStorage) AddItemToTrajectory(item *trajectory.TrajectoryItem) error {
	return sls.AddItemsToTrajectory([]*trajectory.TrajectoryItem{item})
}

func (sls *SharedLocalStorage) AddItemsToTrajectory(items []*trajectory.TrajectoryItem) error {
	if traj, err := sls.ReadTrajectory(); err != nil {
		return err
	} else {
		traj.Items = append(traj.Items, items...)
		sls.Set("traj", traj)
	}
	return nil
}

func handleSLS(sls *SharedLocalStorage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("SLS: %s %s", r.Method, r.URL.Path)
		if r.Method != "GET" && r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		} else if r.Method == "GET" {
			data := sls.data
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if r.Method == "POST" {
			var data struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			sls.Set(data.Key, data.Value)
			w.WriteHeader(http.StatusOK)
			return
		}
	})
}

func (sls *SharedLocalStorage) Start() error {
	server := &http.Server{Addr: fmt.Sprintf(":%d", sls.Port)}
	http.Handle("/api-sls/traj", handleTraj(sls))
	http.Handle("/api-sls", handleSLS(sls))

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

func (sls *SharedLocalStorage) Get(key string) (any, error) {
	sls.mu.Lock()
	defer sls.mu.Unlock()
	if value, ok := sls.data[key]; ok {
		return value, nil
	}
	return "", fmt.Errorf("key not found")
}

func (sls *SharedLocalStorage) Set(key string, value any) {
	sls.mu.Lock()
	defer sls.mu.Unlock()
	sls.data[key] = value
}

func (sls *SharedLocalStorage) Clear() {
	sls.mu.Lock()
	defer sls.mu.Unlock()
	sls.data = make(map[string]any)
}
