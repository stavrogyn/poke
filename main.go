package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/stavrogyn/pokedexcli/internals"
)

type cliCommand struct {
	name        string
	description string
	callback    func([]string) error
}

var limit, offset = 20, 0

var cache = internals.NewCache(time.Minute * 10)

var commands = []cliCommand{
	{
		name:        "exit",
		description: "Exit the program",
		callback:    commandExit,
	},
	{
		name:        "help",
		description: "Show available commands",
		callback:    commandHelp,
	},
	{
		name:        "map",
		description: "Show the map",
		callback:    commandMap,
	},
	{
		name:        "mapb",
		description: "Show the map backwards",
		callback:    commandMapBackwards,
	},
	{
		name:        "explore",
		description: "Explore the map",
		callback:    commandExplore,
	},
	{
		name:        "catch",
		description: "Catch a pokemon",
		callback:    commandCatch,
	},
	{
		name:        "inspect",
		description: "Inspect a pokemon",
		callback:    commandInspect,
	},
	{
		name:        "pokedex",
		description: "Show the pokedex",
		callback:    commandPokedex,
	},
}

type Pokemon struct {
	Name   string
	Height int
	Weight int
	Types  []string
	Stats  map[string]int
}

var caughtPokemon = make(map[string]Pokemon)

func commandPokedex(args []string) error {
	fmt.Printf("Your Pokedex:\n")

	for pokemon := range caughtPokemon {
		fmt.Printf("  - %s\n", pokemon)
	}

	return nil
}

func commandInspect(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("inspect command requires a pokemon name")
	}

	pokemon := args[0]

	p, exists := caughtPokemon[pokemon]

	if !exists {
		return fmt.Errorf("you have not caught %s", pokemon)
	}

	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Height: %d\n", p.Height)
	fmt.Printf("Weight: %d\n", p.Weight)
	fmt.Printf("Stats:\n")
	for name, value := range p.Stats {
		fmt.Printf("  - %s: %d\n", name, value)
	}
	fmt.Printf("Types:\n")
	for _, t := range p.Types {
		fmt.Printf("  - %s\n", t)
	}

	return nil
}

func commandCatch(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("catch command requires a pokemon name")
	}

	pokemon := args[0]

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon)

	req, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemon))

	if err != nil {
		return err
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)

	if err != nil {
		return err
	}

	var pokemonResponse struct {
		BaseExperience int `json:"base_experience"`
		Height         int `json:"height"`
		Weight         int `json:"weight"`
		Types          []struct {
			Slot int `json:"slot"`
			Type struct {
				Name string `json:"name"`
				Url  string `json:"url"`
			} `json:"type"`
		} `json:"types"`
		Stats []struct {
			BaseStat int `json:"base_stat"`
			Effort   int `json:"effort"`
			Stat     struct {
				Name string `json:"name"`
				Url  string `json:"url"`
			} `json:"stat"`
		} `json:"stats"`
	}

	err = json.Unmarshal(body, &pokemonResponse)

	if err != nil {
		return err
	}

	chanceReduction := float64(pokemonResponse.BaseExperience) / 1000.0
	chance := rand.Float64() - chanceReduction

	if rand.Float64() < chance {
		fmt.Printf("%s was caught!\n", pokemon)

		types := make([]string, len(pokemonResponse.Types))
		for i, t := range pokemonResponse.Types {
			types[i] = t.Type.Name
		}

		stats := make(map[string]int)
		for _, s := range pokemonResponse.Stats {
			stats[s.Stat.Name] = s.BaseStat
		}

		caughtPokemon[pokemon] = Pokemon{
			Name:   pokemon,
			Height: pokemonResponse.Height,
			Weight: pokemonResponse.Weight,
			Types:  types,
			Stats:  stats,
		}
	} else {
		fmt.Printf("%s escaped!\n", pokemon)
	}

	return nil
}
func commandExplore(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("explore command requires a location name")
	}

	location := args[0]

	req, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", location))

	if err != nil {
		return err
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)

	if err != nil {
		return err
	}

	var locationResponse struct {
		PokemonEncounters []struct {
			Pokemon struct {
				Name string `json:"name"`
				Url  string `json:"url"`
			} `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}

	err = json.Unmarshal(body, &locationResponse)

	if err != nil {
		return err
	}

	for _, pokemon := range locationResponse.PokemonEncounters {
		fmt.Println(pokemon.Pokemon.Name)
	}

	return nil
}

func getLocations() ([]string, error) {
	locations := []string{}

	cacheKey := fmt.Sprintf("locations-%d-%d", limit, offset)

	if val, exists := cache.Get(cacheKey); exists {
		err := json.Unmarshal(val, &locations)
		if err == nil {
			return locations, nil
		}
	}

	req, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/location-area/?limit=%d&offset=%d", limit, offset))

	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	var locationsResponse struct {
		Count    int     `json:"count"`
		Next     string  `json:"next"`
		Previous *string `json:"previous"`
		Results  []struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"results"`
	}

	err = json.Unmarshal(body, &locationsResponse)

	if err != nil {
		return nil, err
	}

	for _, result := range locationsResponse.Results {
		locations = append(locations, result.Name)
	}

	cache.Add(cacheKey, body)

	return locations, nil
}

func commandMapBackwards(args []string) error {
	offset -= limit
	if offset < 0 {
		offset = 0
	}
	fmt.Printf("Offset: %d\n", offset)
	fmt.Printf("Limit: %d\n", limit)

	err := commandMap(args)

	offset -= limit
	if offset < 0 {
		offset = 0
	}

	return err
}

func commandMap(_ []string) error {
	res, err := getLocations()

	if err != nil {
		return err
	}

	for _, location := range res {
		fmt.Println(location)
	}

	offset += limit

	return nil
}

func commandExit(_ []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!!")
	os.Exit(0)
	return nil
}

func commandHelp(_ []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("\nhelp: Displays a help message")
	fmt.Println("exit: Exit the Pokedex")
	return nil
}

func main() {
	cache.StartReapLoop(time.Minute * 10)

	scanner := bufio.NewScanner(os.Stdin)

	commandMap := make(map[string]cliCommand)
	for _, command := range commands {
		commandMap[command.name] = command
	}

	for scanner.Scan() {
		text := scanner.Text()
		words := CleanInput(text)
		if len(words) > 0 {
			command := words[0]
			args := words[1:]
			if cmd, exists := commandMap[command]; exists {
				err := cmd.callback(args)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}

func CleanInput(text string) []string {
	words := strings.Fields(strings.TrimSpace(strings.ToLower(text)))
	return words
}
