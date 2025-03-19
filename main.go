package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/simonlewi/pokedexcli/internal/pokeapi"
	"github.com/simonlewi/pokedexcli/internal/pokecache"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	scanner := bufio.NewScanner(os.Stdin)
	config := &Config{
		PokeClient:    pokeapi.NewClient(),
		Cache:         pokecache.NewCache(5 * time.Minute),
		CaughtPokemon: make(map[string]Pokemon),
	}

	commands := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Display the map",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the previous map",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Explore the current area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon",
			callback:    commandInspect,
		},
	}

	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp(commands),
	}

	// REPL loop
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()

		words := cleanInput(input)
		if len(words) == 0 {
			continue
		}

		command := words[0]
		if cmd, exists := commands[command]; exists {
			args := []string{}
			if len(words) > 1 {
				args = words[1:]
			}
			err := cmd.callback(config, args)
			if err != nil {
				fmt.Println("Error: ", err)
			}
		} else {
			fmt.Println("Unknown command")
		}

	}

}

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	lowercased := strings.ToLower(trimmed)
	words := strings.Fields(lowercased)

	return words

}

func commandExit(config *Config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(commands map[string]cliCommand) func(*Config, []string) error {
	return func(config *Config, args []string) error {
		fmt.Println("Welcome to the Pokedex!")
		fmt.Println("Usage:")
		fmt.Println()
		for _, cmd := range commands {
			fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
		}
		return nil
	}
}

func commandMap(config *Config, args []string) error {
	resp, err := config.PokeClient.ListLocationAreas(config.NextURL, config.Cache)
	if err != nil {
		return err
	}

	config.NextURL = resp.Next
	config.PreviousURL = resp.Previous

	pokeapi.PrintLocationAreas(resp)

	return nil
}

func commandMapBack(config *Config, args []string) error {
	if config.PreviousURL == nil {
		fmt.Println("You're on the first page")
		return nil
	}

	resp, err := config.PokeClient.ListLocationAreas(*config.PreviousURL, config.Cache)
	if err != nil {
		return err
	}

	// Update config with new URLs
	config.NextURL = resp.Next
	config.PreviousURL = resp.Previous

	// Print the locations
	pokeapi.PrintLocationAreas(resp)

	return nil
}

func commandExplore(config *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("you must provide a location area name")
	}
	locationArea := args[0]

	// Call the API to get location area details
	locationInfo, err := config.PokeClient.GetLocationArea(locationArea, config.Cache)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", locationArea)
	fmt.Println("Found Pokemon")
	for _, encounter := range locationInfo.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}

func commandCatch(config *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("you must provide a pokemon name")
	}
	pokemonName := args[0]

	// Get Pokemon info from API
	pokemon, err := config.PokeClient.GetPokemon(pokemonName, config.Cache)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// Calculate catch chance - higher base experience = lower chance
	catchChance := 0.5 - float64(pokemon.BaseExperience)/1000.0 // Adjust to make catching easier or harder
	if catchChance < 0.1 {
		catchChance = 0.1 // Minimum 10% chance
	}

	// Random number between 0 - 1
	r := rand.Float64()

	// If random number is less than catch chance, catch succeeded
	if r < catchChance {
		fmt.Printf("%s was caught!\n", pokemonName)

		// Convert API stats to our internal format
		stats := make([]Stat, len(pokemon.Stats))
		for i, stat := range pokemon.Stats {
			stats[i] = Stat{
				Name:  stat.Stat.Name,
				Value: stat.BaseStat,
			}
		}

		// Convert API types to our internal format
		types := make([]Type, len(pokemon.Types))
		for i, t := range pokemon.Types {
			types[i] = Type{
				Name: t.Type.Name,
			}
		}

		config.CaughtPokemon[pokemonName] = Pokemon{
			Name:           pokemonName,
			BaseExperience: pokemon.BaseExperience,
			Height:         pokemon.Height,
			Weight:         pokemon.Weight,
			Stats:          stats,
			Types:          types,
		}
		return nil
	}

	fmt.Printf("%s escaped!\n", pokemonName)
	return nil
}

func commandInspect(config *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("you must provide a pokemon name")
	}
	pokemonName := args[0]

	pokemon, ok := config.CaughtPokemon[pokemonName]
	if !ok {
		return fmt.Errorf("you have not caught that pokemon yet")
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Base Experience: %d\n", pokemon.BaseExperience)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Printf("Stats:\n")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Name, stat.Value)
	}
	fmt.Printf("Types:\n")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Name)
	}

	return nil
}

type Pokemon struct {
	Name           string
	BaseExperience int
	Height         int
	Weight         int
	Stats          []Stat
	Types          []Type
}

type Stat struct {
	Name  string
	Value int
}

type Type struct {
	Name string
}

type Config struct {
	NextURL       string
	PreviousURL   *string
	PokeClient    *pokeapi.Client
	Cache         *pokecache.Cache
	CaughtPokemon map[string]Pokemon
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
}
