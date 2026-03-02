package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"math/rand"
)

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