package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*
========================
CACHE UNIT TEST
========================
*/

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("India", "TestValue")

	value, found := cache.Get("India")

	if !found {
		t.Fatalf("expected value to be found in cache")
	}

	if value != "TestValue" {
		t.Fatalf("expected TestValue, got %v", value)
	}
}

/*
========================
HTTP HANDLER TEST
========================
*/

func TestCountryHandler_FromCache(t *testing.T) {
	cache := NewMemoryCache()

	// Put fake country data in cache
	cache.Set("India", Country{
		Name:       "India",
		Capital:    "New Delhi",
		Currency:   "₹",
		Population: 123,
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/countries/search?name=India",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler := countryHandler(cache)
	handler.ServeHTTP(recorder, req)

	// Check HTTP status
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()

	// Basic response checks
	if !strings.Contains(body, `"name":"India"`) {
		t.Fatalf("expected response to contain country name, got %s", body)
	}

	if !strings.Contains(body, `"capital":"New Delhi"`) {
		t.Fatalf("expected response to contain capital, got %s", body)
	}
}
