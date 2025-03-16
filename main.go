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
			err := cmd.callback(config)
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

func commandExit(config *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(commands map[string]cliCommand) func(*Config) error {
	return func(config *Config) error {
		fmt.Println("Welcome to the Pokedex!")
		fmt.Println("Usage:")
		fmt.Println()
		for _, cmd := range commands {
			fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
		}
		return nil
	}
}

func commandMap(config *Config) error {
	resp, err := config.PokeClient.ListLocationAreas(config.NextURL, config.Cache)
	if err != nil {
		return err
	}

	config.NextURL = resp.Next
	config.PreviousURL = resp.Previous

	pokeapi.PrintLocationAreas(resp)

	return nil
}

func commandMapBack(config *Config) error {
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

type Config struct {
	NextURL     string
	PreviousURL *string
	PokeClient  *pokeapi.Client
	Cache       *pokecache.Cache
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}
