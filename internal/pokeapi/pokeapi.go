// internal/pokeapi/pokeapi.go
package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/simonlewi/pokedexcli/internal/pokecache"
)

// Client is a PokeAPI client
type Client struct {
	// You could add more fields here like HTTPClient for testing
}

// NewClient creates a new PokeAPI client
func NewClient() *Client {
	return &Client{}
}

// ListLocationAreas fetches a paginated list of location areas
func (c *Client) ListLocationAreas(url string, cache *pokecache.Cache) (LocationAreaResponse, error) {
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area"
	}

	if cachedData, found := cache.Get(url); found {
		fmt.Println("Cache hit!")

		var locationsResp LocationAreaResponse
		err := json.Unmarshal(cachedData, &locationsResp)
		if err != nil {
			return LocationAreaResponse{}, err
		}

		return locationsResp, nil
	}

	fmt.Println("Cache miss! Fetching from API...")
	resp, err := http.Get(url)
	if err != nil {
		return LocationAreaResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LocationAreaResponse{}, err
	}

	cache.Add(url, body)

	var locationsResp LocationAreaResponse
	err = json.Unmarshal(body, &locationsResp)
	if err != nil {
		return LocationAreaResponse{}, err
	}

	return locationsResp, nil
}

// PrintLocationAreas prints the location areas from a response
func PrintLocationAreas(resp LocationAreaResponse) {
	for _, location := range resp.Results {
		fmt.Println(location.Name)
	}
}
