package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
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



var supportedCommands map[string]cliCommand = make(map[string]cliCommand)
var offset int = 0
var limit int = 20
var cache *pokecache.Cache = pokecache.NewCache(5 * time.Minute)

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
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
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
