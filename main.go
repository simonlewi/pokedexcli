package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/simonlewi/pokedexcli/internal/pokeapi"
	"github.com/simonlewi/pokedexcli/internal/pokecache"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := &Config{
		PokeClient: pokeapi.NewClient(),
		Cache:      pokecache.NewCache(5 * time.Minute),
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

type Config struct {
	NextURL     string
	PreviousURL *string
	PokeClient  *pokeapi.Client
	Cache       *pokecache.Cache
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
}
