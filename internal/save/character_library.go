package save

import (
	"encoding/json"
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

const CharacterLibraryVersion = 1

type CharacterLibrary struct {
	Version    int          `json:"version"`
	Characters party.Roster `json:"characters"`
}

func EncodeCharacterLibrary(roster party.Roster) ([]byte, error) {
	for index, character := range roster {
		if err := character.Validate(); err != nil {
			return nil, fmt.Errorf("character %d: %w", index, err)
		}
	}
	return json.MarshalIndent(CharacterLibrary{Version: CharacterLibraryVersion, Characters: roster}, "", "  ")
}

func DecodeCharacterLibrary(data []byte) (party.Roster, error) {
	var file CharacterLibrary
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("character library JSON: %w", err)
	}
	if file.Version != CharacterLibraryVersion {
		return nil, fmt.Errorf("unsupported character library version %d", file.Version)
	}
	for index, character := range file.Characters {
		if err := character.Validate(); err != nil {
			return nil, fmt.Errorf("character %d: %w", index, err)
		}
	}
	return file.Characters, nil
}
