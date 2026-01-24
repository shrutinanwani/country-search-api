package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

/*
========================
CACHE IMPLEMENTATION
========================
*/

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type MemoryCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]interface{}),
	}
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *MemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

/*
========================
DATA STRUCTS
========================
*/

// Final response returned to client
type Country struct {
	Name       string `json:"name"`
	Capital    string `json:"capital"`
	Currency   string `json:"currency"`
	Population int64  `json:"population"`
}

// Partial REST Countries API response
type restCountry struct {
	Name struct {
		Common string `json:"common"`
	} `json:"name"`
	Capital    []string `json:"capital"`
	Population int64    `json:"population"`
	Currencies map[string]struct {
		Symbol string `json:"symbol"`
	} `json:"currencies"`
}

/*
========================
HTTP HANDLER (TESTABLE)
========================
*/

func countryHandler(cache *MemoryCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name query param is required", http.StatusBadRequest)
			return
		}

		// 1️⃣ Check cache first
		if cached, found := cache.Get(name); found {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cached)
			return
		}

		// 2️⃣ Call REST Countries API
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		url := fmt.Sprintf("https://restcountries.com/v3.1/name/%s", name)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var apiResp []restCountry
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(apiResp) == 0 {
			http.Error(w, "country not found", http.StatusNotFound)
			return
		}

		// 3️⃣ Prefer exact name match
		var rc restCountry
		found := false
		for _, c := range apiResp {
			if c.Name.Common == name {
				rc = c
				found = true
				break
			}
		}
		if !found {
			rc = apiResp[0]
		}

		// Extract currency symbol
		var currency string
		for _, c := range rc.Currencies {
			currency = c.Symbol
			break
		}

		country := Country{
			Name:       rc.Name.Common,
			Capital:    rc.Capital[0],
			Currency:   currency,
			Population: rc.Population,
		}

		// 4️⃣ Save to cache
		cache.Set(name, country)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(country)
	}
}

/*
========================
MAIN FUNCTION
========================
*/

func main() {
	cache := NewMemoryCache()

	http.HandleFunc("/api/countries/search", countryHandler(cache))

	fmt.Println("✅ Server running at http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}
