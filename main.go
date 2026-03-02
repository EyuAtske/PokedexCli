package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"math/rand"
	"github.com/EyuAtske/PokedexCli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

type Location struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

type Pokemon_Area struct{
	PokemonEncounters []struct{
		Pokemon struct{
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name string `json:"name"`
	BaseExperience int `json:"base_experience"`
	Height int `json:"height"`
	Weight int `json:"weight"`
	Stats []struct{
		Basestat int `json:"base_stat"`
		Stat struct{
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct{
		Type struct{
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

type Collection struct{
	Pokemon map[string]Pokemon
}




var supportedCommands map[string]cliCommand = make(map[string]cliCommand)
var offset int = 0
var limit int = 20
var cache *pokecache.Cache = pokecache.NewCache(5 * time.Minute)
var collection Collection = Collection{
	Pokemon: make(map[string]Pokemon),
}

func main(){
	supportedCommands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help":{
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
		"map": {
			name: "map",
			description: "displays a map of the pokemon world",
			callback: commandMap,
		},
		"mapb": {
			name: "mapb",
			description: "displays the previous map of the pokemon world",
			callback: commandMapB,
		},
		"explore": {
			name: "explore <area name>",
			description: "displays the pokemon that can be found in the specified area",
			callback: nil,
		},
		"catch": {
			name: "catch <pokemon name>",
			description: "catches a pokemon and adds it to your collection",
			callback: nil,
		},
		"inspect": {
			name: "inspect <pokemon name>",
			description: "inspects a pokemon in your collection",
			callback: nil,
		},
		"pokedex": {
			name: "pokedex",
			description: "displays all pokemon in your collection",
			callback: commandPokedex,
		},
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		inp := strings.Split(input, " ")
		if len(inp) == 2{
			if inp[0] == "explore"{
				err := commandExplore(inp[1])
				if err != nil {
					fmt.Printf("Error executing command: %s\n", err)
				}
				continue
			}
			if inp[0] == "catch"{
				err := commandCatch(inp[1])
				if err != nil {
					fmt.Printf("Error executing command: %s\n", err)
				}
				continue
			}
			if inp[0] == "inspect"{
				err := commandInspect(inp[1])
				if err != nil {
					fmt.Printf("Error executing command: %s\n", err)
				}
				continue
			}
		}
		switch input {
		case "exit":
			err:= supportedCommands["exit"].callback()
			if err != nil {
				fmt.Printf("Error executing command: %s\n", err)
			}
		case "help":
			err := supportedCommands["help"].callback()
			if err != nil {
				fmt.Printf("Error executing command: %s\n", err)
			}
		case "map":
			err := supportedCommands["map"].callback()
			if err != nil {
				fmt.Printf("Error executing command: %s\n", err)
			}
		case "mapb":
			err := supportedCommands["mapb"].callback()
			if err != nil {
				fmt.Printf("Error executing command: %s\n", err)
			}
		case "pokedex":
			err := supportedCommands["pokedex"].callback()
			if err != nil {
				fmt.Printf("Error executing command: %s\n", err)
			}
		default:
			fmt.Printf("Unknown command: %s\n", input)
			fmt.Println("Type 'help' for a list of available commands.")
			continue
		}
		fmt.Printf("Your command was: %s\n", input)
	}
}

func commandExit() error{
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error{
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")	
	if len(supportedCommands) == 0 {
		return errors.New("no commands available")
	}
	for _, cmd := range supportedCommands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandMap() error{
	key := fmt.Sprintf("%d", offset)
	if val, ok := cache.Get(key); ok {
		var loc Location
		if err := json.Unmarshal(val, &loc); err != nil {
			return fmt.Errorf("error parsing JSON: %w", err)
		}
		for _, location := range loc.Results {
			fmt.Println(location.Name)
		}
		return nil
	}
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area?limit=%d&offset=%d", limit, offset,)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch map data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}
	var loc Location
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	cache.Add(key, data)
	if err := json.Unmarshal(data, &loc); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	if len(loc.Results) == 0 {
		fmt.Println("No more locations found.")
		return nil
	}
	for _, location := range loc.Results {
		fmt.Println(location.Name)
	}
	offset += limit
	return nil
}

func commandMapB() error{
	if offset <= 0 {
		fmt.Println("You are already at the beginning of the map.")
		return nil
	}
	offset -= 2 *limit
	return commandMap()
}

func commandExplore(areaName string) error{
	key := areaName
	if val, ok := cache.Get(key); ok {
		var pokemon Pokemon_Area
		if err := json.Unmarshal(val, &pokemon); err != nil {
			return fmt.Errorf("error parsing JSON: %w", err)
		}
		fmt.Printf("Exploring %s...\n", areaName)
		fmt.Println("Found Pokemon:")
		for _, encounter := range pokemon.PokemonEncounters {
			fmt.Println("- " + encounter.Pokemon.Name)
		}
		return nil
	}
	fullURL := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)
	resp, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("failed to fetch area data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	cache.Add(key, data)
	var pokemon Pokemon_Area
	if err := json.Unmarshal(data, &pokemon); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}
	fmt.Printf("Exploring %s...\n", areaName)
	fmt.Println("Found Pokemon:")
	for _, encounter := range pokemon.PokemonEncounters {
		fmt.Println("- " + encounter.Pokemon.Name)
	}
	return nil
}

func commandCatch(pokemonName string) error{
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)
	for name := range collection.Pokemon {
		if name == pokemonName {
			fmt.Printf("You already have %s in your collection!\n", pokemonName)
			return nil
		}
	}
	key := pokemonName
	if val, ok := cache.Get(key); ok {
		var pokemon Pokemon
		if err := json.Unmarshal(val, &pokemon); err != nil {
			return fmt.Errorf("error parsing JSON: %w", err)
		}
		chance := rand.Intn(pokemon.BaseExperience + 1)
		if chance >= pokemon.BaseExperience/2 {
			fmt.Println(pokemonName + " was caught!")
			collection.Pokemon[pokemonName] = pokemon
		} else {
			fmt.Println(pokemonName + " escaped!")
		}
		return nil
	}
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch pokemon data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	cache.Add(key, data)
	var pokemon Pokemon
	if err := json.Unmarshal(data, &pokemon); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}
	chance := rand.Intn(pokemon.BaseExperience + 1)
	if chance >= pokemon.BaseExperience/2 {
		fmt.Println(pokemonName + " was caught!")
		collection.Pokemon[pokemonName] = pokemon
	} else {
		fmt.Println(pokemonName + " escaped!")
	}
	return nil
}

func commandInspect(pokemonName string) error{
	pokemon, ok := collection.Pokemon[pokemonName]
	if !ok {
		fmt.Printf("You don't have %s in your collection!\n", pokemonName)
		return nil
	}
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("- %s: %d\n", stat.Stat.Name, stat.Basestat)
	}
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("- %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex() error{
	if len(collection.Pokemon) == 0 {
		fmt.Println("Your collection is empty!")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for name := range collection.Pokemon {
		fmt.Println("- " + name)
	}
	return nil
}