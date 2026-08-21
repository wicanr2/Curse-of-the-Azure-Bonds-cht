package ecl

import (
	"fmt"

	"github.com/wicanr2/golden-box-remake-engine/randomstream"
)

const sessionSnapshotVersion = 1

// SessionSnapshot is the game-neutral mutable ECL continuation stored by the
// remake save. Original code bytes are not copied; CodeMemory contains only
// runtime differences from the player-supplied block image.
type SessionSnapshot struct {
	Version           int                    `json:"version"`
	CurrentBlock      uint8                  `json:"current_block"`
	PC                int                    `json:"pc"`
	Started           bool                   `json:"started"`
	Stack             []int                  `json:"stack,omitempty"`
	Memory            map[uint16]uint16      `json:"memory,omitempty"`
	CodeMemory        map[uint16]uint16      `json:"code_memory_changes,omitempty"`
	Strings           map[uint16]string      `json:"strings,omitempty"`
	Compare           [6]bool                `json:"compare"`
	SelectedPlayer    int                    `json:"selected_player"`
	SelectedPlayerSet bool                   `json:"selected_player_set"`
	MonsterSetup      *MonsterSetup          `json:"monster_setup,omitempty"`
	MonsterSpawns     []MonsterSpawn         `json:"monster_spawns,omitempty"`
	Random            *randomstream.Snapshot `json:"random,omitempty"`
	// BlockScratch 是「停在旁邊」那幾段的暫存（`4C00h`..`4C0Fh`，spec 1162）。
	// 目前這一段的那一份在 Memory 裡，不重複收。舊存檔沒有這一欄，讀回來就是
	// 沒有停放值——與加這一欄之前的行為相同。
	BlockScratch       map[uint8]map[uint16]uint16 `json:"block_scratch,omitempty"`
	SelectionOffset    int                         `json:"selection_offset"`
	WhoSelectionOffset int                         `json:"who_selection_offset"`
	StringInputOffset  int                         `json:"string_input_offset"`
}

// Snapshot returns an owned representation of the shared session runtime.
func (s *BlockSession) Snapshot() (SessionSnapshot, error) {
	if s == nil {
		return SessionSnapshot{}, fmt.Errorf("cannot snapshot a nil ECL session")
	}
	runtime := s.states[s.current]
	if runtime == nil {
		return SessionSnapshot{}, fmt.Errorf("ECL block 0x%02X has no runtime", s.current)
	}
	snapshot := SessionSnapshot{
		Version: sessionSnapshotVersion, CurrentBlock: s.current,
		PC: runtime.PC, Started: runtime.Started,
		Stack:  append([]int(nil), runtime.Stack...),
		Memory: make(map[uint16]uint16), CodeMemory: make(map[uint16]uint16),
		Strings: make(map[uint16]string, len(runtime.Strings)), Compare: runtime.Compare,
		SelectedPlayer: runtime.SelectedPlayerIndex, SelectedPlayerSet: runtime.SelectedPlayerSet,
		MonsterSpawns:   append([]MonsterSpawn(nil), runtime.MonsterSpawns...),
		SelectionOffset: s.selectionOffset, WhoSelectionOffset: s.whoSelectionOffset,
		StringInputOffset: s.stringInputOffset,
	}
	if runtime.MonsterSetup != nil {
		setup := *runtime.MonsterSetup
		snapshot.MonsterSetup = &setup
	}
	for address, value := range runtime.Strings {
		snapshot.Strings[address] = value
	}
	for address, value := range runtime.Memory {
		if address < CodeAddressBase || address > 0x9DFF {
			snapshot.Memory[address] = value
			continue
		}
		index := int(address) - CodeAddressBase
		data := s.blocks[s.current]
		if index < 0 || index+2 >= len(data) || uint16(data[index+2]) != value {
			snapshot.CodeMemory[address] = value
		}
	}
	for blockID, bank := range s.blockScratch {
		if blockID == s.current || len(bank) == 0 {
			continue
		}
		if snapshot.BlockScratch == nil {
			snapshot.BlockScratch = make(map[uint8]map[uint16]uint16, len(s.blockScratch))
		}
		owned := make(map[uint16]uint16, len(bank))
		for address, value := range bank {
			owned[address] = value
		}
		snapshot.BlockScratch[blockID] = owned
	}
	if runtime.Random != nil {
		random := runtime.Random.Snapshot()
		snapshot.Random = &random
	}
	return snapshot, nil
}

// RestoreSnapshot replaces mutable session state while rebuilding original
// code memory from the already loaded, player-supplied ECL blocks.
func (s *BlockSession) RestoreSnapshot(snapshot SessionSnapshot) error {
	if s == nil {
		return fmt.Errorf("cannot restore a nil ECL session")
	}
	if snapshot.Version != sessionSnapshotVersion {
		return fmt.Errorf("unsupported ECL session snapshot version %d", snapshot.Version)
	}
	data, ok := s.blocks[snapshot.CurrentBlock]
	if !ok {
		return fmt.Errorf("ECL session snapshot block 0x%02X is unavailable", snapshot.CurrentBlock)
	}
	payloadLength := len(data) - 2
	if snapshot.Started && (snapshot.PC < 0 || snapshot.PC > payloadLength) {
		return fmt.Errorf("ECL session snapshot PC %d is outside payload length %d", snapshot.PC, payloadLength)
	}
	for _, pc := range snapshot.Stack {
		if pc < 0 || pc > payloadLength {
			return fmt.Errorf("ECL session snapshot stack PC %d is outside payload length %d", pc, payloadLength)
		}
	}
	if snapshot.SelectionOffset < 0 || snapshot.WhoSelectionOffset < 0 || snapshot.StringInputOffset < 0 {
		return fmt.Errorf("ECL session snapshot contains a negative input offset")
	}
	var random *randomstream.Stream
	var err error
	if snapshot.Random != nil {
		random, err = randomstream.Restore(*snapshot.Random)
		if err != nil {
			return fmt.Errorf("restore ECL random stream: %w", err)
		}
	}
	runtime := NewRuntimeState(snapshot.PC)
	runtime.Started = snapshot.Started
	runtime.Stack = append([]int(nil), snapshot.Stack...)
	runtime.Compare = snapshot.Compare
	runtime.SelectedPlayerIndex = snapshot.SelectedPlayer
	runtime.SelectedPlayerSet = snapshot.SelectedPlayerSet
	runtime.MonsterSpawns = append([]MonsterSpawn(nil), snapshot.MonsterSpawns...)
	if snapshot.MonsterSetup != nil {
		setup := *snapshot.MonsterSetup
		runtime.MonsterSetup = &setup
	}
	runtime.Random = random
	for blockID := range s.blocks {
		s.states[blockID] = runtime
	}
	s.current = snapshot.CurrentBlock
	s.blockScratch = make(map[uint8]map[uint16]uint16, len(snapshot.BlockScratch))
	for blockID, bank := range snapshot.BlockScratch {
		if blockID == snapshot.CurrentBlock {
			return fmt.Errorf("ECL session snapshot parks scratch for the current block 0x%02X", blockID)
		}
		owned := make(map[uint16]uint16, len(bank))
		for address, value := range bank {
			if address < blockScratchLow || address > blockScratchHigh {
				return fmt.Errorf("ECL session snapshot parked scratch has non-scratch address 0x%04X", address)
			}
			owned[address] = value
		}
		s.blockScratch[blockID] = owned
	}
	s.selectionOffset = snapshot.SelectionOffset
	s.whoSelectionOffset = snapshot.WhoSelectionOffset
	s.stringInputOffset = snapshot.StringInputOffset
	s.loadCurrentCodeMemory()
	for address, value := range snapshot.Memory {
		if address >= CodeAddressBase && address <= 0x9DFF {
			return fmt.Errorf("ECL session snapshot regular memory contains code address 0x%04X", address)
		}
		runtime.Memory[address] = value
	}
	for address, value := range snapshot.CodeMemory {
		if address < CodeAddressBase || address > 0x9DFF {
			return fmt.Errorf("ECL session snapshot code change has non-code address 0x%04X", address)
		}
		runtime.Memory[address] = value
	}
	for address, value := range snapshot.Strings {
		runtime.Strings[address] = value
	}
	return nil
}
