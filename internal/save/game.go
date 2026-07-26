package save

import (
	"encoding/json"
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// CurrentGameVersion is the version of the remake's resumable game save.
const CurrentGameVersion = 2

// GameFile contains the party plus the platform-neutral adventure state that
// the remake can currently restore. Numeric mode/location values are kept
// here to avoid coupling the save package to the game UI package.
type GameFile struct {
	Version    int          `json:"version"`
	Characters party.Roster `json:"characters"`
	Area       area.State   `json:"area"`
	Mode       uint8        `json:"mode"`
	Location   uint8        `json:"location"`
	MapX       int          `json:"map_x"`
	MapY       int          `json:"map_y"`
}

func EncodeGame(roster party.Roster, areaState area.State, mode, location uint8, mapX, mapY int) ([]byte, error) {
	if err := roster.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(GameFile{
		Version: CurrentGameVersion, Characters: roster, Area: areaState,
		Mode: mode, Location: location, MapX: mapX, MapY: mapY,
	}, "", "  ")
}

func DecodeGame(data []byte) (GameFile, error) {
	var file GameFile
	if err := json.Unmarshal(data, &file); err != nil {
		return GameFile{}, fmt.Errorf("game save JSON: %w", err)
	}
	// Version 1 was a party-only save. Accept it and use safe defaults for
	// fields introduced by the resumable game format.
	if file.Version != 1 && file.Version != CurrentGameVersion {
		return GameFile{}, fmt.Errorf("unsupported game save version %d", file.Version)
	}
	if err := file.Characters.Validate(); err != nil {
		return GameFile{}, err
	}
	return file, nil
}
