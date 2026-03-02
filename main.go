package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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