package externalapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
)

type RickAndMortyClient struct {
	baseURL string
}

type CharacterData struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Species string `json:"species"`
	Type    string `json:"type"`
	Gender  string `json:"gender"`
	Origin  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"origin"`
	Location struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Image   string   `json:"image"`
	Episode []string `json:"episode"`
	URL     string   `json:"url"`
	Created string   `json:"created"`
}

type apiResponse struct {
	Info struct {
		Next string `json:"next"`
	} `json:"info"`
	Results []CharacterData `json:"results"`
}

func NewRickAndMortyClient() *RickAndMortyClient {
	return &RickAndMortyClient{
		baseURL: "https://rickandmortyapi.com/api",
	}
}

// FetchAllCharacters fetches all characters from the Rick and Morty API
func (c *RickAndMortyClient) FetchAllCharacters() ([]CharacterData, error) {
	var allCharacters []CharacterData
	nextURL := fmt.Sprintf("%s/character", c.baseURL)

	for nextURL != "" {
		resp, err := http.Get(nextURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch characters: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var apiResp apiResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		allCharacters = append(allCharacters, apiResp.Results...)
		nextURL = apiResp.Info.Next
	}

	return allCharacters, nil
}

func (c *RickAndMortyClient) FetchRandomCharacter() (*CharacterData, error) {
	randomID := rand.Intn(826) + 1
	url := fmt.Sprintf("%s/character/%d", c.baseURL, randomID)
	log.Printf("Fetching random character from: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to fetch character: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned status: %d", resp.StatusCode)
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var char CharacterData
	if err := json.NewDecoder(resp.Body).Decode(&char); err != nil {
		log.Printf("JSON decode failed: %v", err)
		return nil, fmt.Errorf("failed to decode character: %w", err)
	}

	log.Printf("Successfully fetched character: %s", char.Name)
	return &char, nil
}

// GetAge returns a human-readable age representation
func (c *CharacterData) GetAge() string {
	// Rick and Morty API doesn't provide exact age, so we'll use species-based logic
	switch c.Species {
	case "Human":
		return "Adult"
	case "Alien":
		return "Unknown"
	case "Robot":
		return "N/A"
	case "Humanoid":
		return "Adult"
	case "Animal":
		return "Adult"
	case "Mythological Creature":
		return "Ancient"
	case "Disease":
		return "N/A"
	case "Cronenberg":
		return "Unknown"
	case "Poopybutthole":
		return "Unknown"
	default:
		return "Unknown"
	}
}
