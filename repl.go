package main

import (
	"strings"
)

func cleanInput(text string) []string{
	var clean []string
	trimmed := strings.TrimSpace(text)
	lowered := strings.ToLower(trimmed)
	clean = strings.Fields(lowered)
	return clean
}