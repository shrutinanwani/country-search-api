# Golang Country Search API

A simple REST API built in Go that fetches country information using the REST Countries API and caches results in memory.

---

## Features

- Search country by name
- Fetches data from REST Countries API
- Custom in-memory cache (thread-safe)
- Cache-first approach for faster responses
- Unit tests for cache and API handler

---

## API Endpoint

### GET /api/countries/search

**Query Parameters:**
- `name` (string) — country name

**Example Request:**
```bash
curl "http://localhost:8000/api/countries/search?name=India"


Example response :
{
  "name": "India",
  "capital": "New Delhi",
  "currency": "₹",
  "population": 1417492000
}
