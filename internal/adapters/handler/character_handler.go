package handler

import (
	"encoding/json"
	"net/http"
	"1337b04rd/internal/adapters/externalapi"
)

type CharacterHandler struct {
	rickAndMortyClient *externalapi.RickAndMortyClient
}

func NewCharacterHandler(rickAndMortyClient *externalapi.RickAndMortyClient) *CharacterHandler {
	return &CharacterHandler{
		rickAndMortyClient: rickAndMortyClient,
	}
}

// GetRandomCharacter returns a random character from the Rick and Morty API
func (h *CharacterHandler) GetRandomCharacter(w http.ResponseWriter, r *http.Request) {
	character, err := h.rickAndMortyClient.FetchRandomCharacter()
	if err != nil {
		http.Error(w, "Failed to fetch random character: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(character)
}

// GetAllCharacters returns all characters from the Rick and Morty API
func (h *CharacterHandler) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
	characters, err := h.rickAndMortyClient.FetchAllCharacters()
	if err != nil {
		http.Error(w, "Failed to fetch characters: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(characters)
}
