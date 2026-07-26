// Package save defines the remake's versioned, JSON party save format. It is
// intentionally separate from the original DOS save format, which remains
// reverse-engineering work.
package save

import (
	"encoding/json"
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

const CurrentVersion = 1

type PartyFile struct {
	Version    int          `json:"version"`
	Characters party.Roster `json:"characters"`
}

func EncodeParty(roster party.Roster) ([]byte, error) {
	if err := roster.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(PartyFile{Version: CurrentVersion, Characters: roster}, "", "  ")
}

func DecodeParty(data []byte) (party.Roster, error) {
	var file PartyFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("party save JSON: %w", err)
	}
	if file.Version != CurrentVersion {
		return nil, fmt.Errorf("unsupported party save version %d", file.Version)
	}
	if err := file.Characters.Validate(); err != nil {
		return nil, err
	}
	return file.Characters, nil
}
