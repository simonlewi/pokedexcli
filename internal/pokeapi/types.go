package pokeapi

// LocationAreaResponse represents the response from the location-area endpoint
type LocationAreaResponse struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous *string `json:"previous"` // Note the *string for nullable field
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationArea struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name           string
	BaseExperience int
}

type PokemonResponse struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}
