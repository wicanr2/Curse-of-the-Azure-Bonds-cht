package ecl

import "fmt"

// BlockSession owns decoded ECL blocks and the current block identity. It is
// intentionally separate from RunSubset: VM execution state can be extended
// later without confusing a DAX block switch with a fallthrough branch.
type BlockSession struct {
	blocks          map[uint8][]byte
	current         uint8
	states          map[uint8]*RuntimeState
	selectionOffset int
}

func NewBlockSession(blocks map[uint8][]byte, current uint8) (*BlockSession, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("ECL session has no blocks")
	}
	if _, ok := blocks[current]; !ok {
		return nil, fmt.Errorf("ECL session has no current block 0x%02X", current)
	}
	owned := make(map[uint8][]byte, len(blocks))
	for id, data := range blocks {
		owned[id] = append([]byte(nil), data...)
	}
	states := make(map[uint8]*RuntimeState, len(owned))
	shared := NewRuntimeState(0)
	for id := range owned {
		states[id] = shared
	}
	return &BlockSession{blocks: owned, current: current, states: states}, nil
}

func (s *BlockSession) CurrentBlockID() uint8 { return s.current }

func (s *BlockSession) CurrentData() []byte {
	return append([]byte(nil), s.blocks[s.current]...)
}

func (s *BlockSession) Switch(id uint8) error {
	if _, ok := s.blocks[id]; !ok {
		return fmt.Errorf("ECL session target block 0x%02X is unavailable", id)
	}
	s.current = id
	if _, ok := s.states[id]; !ok {
		s.states[id] = s.states[s.current]
	}
	return nil
}

// InitialEntry returns the fifth vm_init_ecl entry for the current block as a
// decoded payload offset.
func (s *BlockSession) InitialEntry() (int, error) {
	points, _, err := EntryPoints(s.blocks[s.current], 5)
	if err != nil {
		return 0, err
	}
	return int(points[4]) - CodeAddressBase, nil
}

// ApplyResult applies a bounded runner's NEWECL signal to this session.
func (s *BlockSession) ApplyResult(result RunResult) error {
	if result.NewECLBlockID == nil {
		return nil
	}
	return s.Switch(*result.NewECLBlockID)
}

// RunInteractive executes the current block from its original entry and
// automatically follows bounded NEWECL signals. The selection offset lets a
// caller keep one global input sequence while each newly loaded block sees
// only the choices not consumed by earlier blocks.
func (s *BlockSession) RunInteractive(maxSteps int, selections []uint16) (RunResult, error) {
	start, err := s.InitialEntry()
	if err != nil {
		return RunResult{}, err
	}
	return s.RunFrom(start, maxSteps, selections)
}

// RunInteractiveSeed follows NEWECL transitions with a reproducible RANDOM
// stream. The unseeded method remains the compatibility wrapper.
func (s *BlockSession) RunInteractiveSeed(maxSteps int, selections []uint16, seed int64) (RunResult, error) {
	start, err := s.InitialEntry()
	if err != nil {
		return RunResult{}, err
	}
	return s.runFromSeed(start, maxSteps, selections, seed)
}

// RunInteractiveSeedWithPartyContext resolves party-rule commands against a
// caller-owned roster while preserving the session's shared runtime memory.
func (s *BlockSession) RunInteractiveSeedWithPartyContext(maxSteps int, selections []uint16, seed int64, context PartyContext) (RunResult, error) {
	start, err := s.InitialEntry()
	if err != nil {
		return RunResult{}, err
	}
	return s.runFromSeedWithPartyContext(start, maxSteps, selections, seed, &context)
}

// RunFrom executes an explicit event entry in the current block. After a
// NEWECL signal, the target resumes at its own initial entry.
func (s *BlockSession) RunFrom(start, maxSteps int, selections []uint16) (RunResult, error) {
	return s.runFromSeed(start, maxSteps, selections, 1)
}

func (s *BlockSession) runFromSeed(start, maxSteps int, selections []uint16, seed int64) (RunResult, error) {
	return s.runFromSeedWithPartyContext(start, maxSteps, selections, seed, nil)
}

func (s *BlockSession) runFromSeedWithPartyContext(start, maxSteps int, selections []uint16, seed int64, partyContext *PartyContext) (RunResult, error) {
	var aggregate RunResult
	var err error
	selectionOffset := s.selectionOffset
	for transitions := 0; transitions < 8; transitions++ {
		remaining := selections
		if selectionOffset < len(selections) {
			remaining = selections[selectionOffset:]
		} else {
			remaining = nil
		}
		runtime := s.states[s.current]
		if !runtime.Started {
			runtime.PC = start
		}
		result, runErr := runSubsetWithStateContext(s.CurrentData(), start, maxSteps, remaining, true, seed, runtime, partyContext)
		aggregate.Text = append(aggregate.Text, result.Text...)
		aggregate.Menus = append(aggregate.Menus, result.Menus...)
		aggregate.Steps += result.Steps
		aggregate.PC = result.PC
		aggregate.CombatRequested = aggregate.CombatRequested || result.CombatRequested
		if result.MonsterSetup != nil {
			setup := *result.MonsterSetup
			aggregate.MonsterSetup = &setup
		}
		aggregate.MonsterSpawns = append(aggregate.MonsterSpawns, result.MonsterSpawns...)
		aggregate.ProgramIDs = append(aggregate.ProgramIDs, result.ProgramIDs...)
		aggregate.ProgramExit = aggregate.ProgramExit || result.ProgramExit
		aggregate.CallAddresses = append(aggregate.CallAddresses, result.CallAddresses...)
		aggregate.DamageRequests = append(aggregate.DamageRequests, result.DamageRequests...)
		aggregate.PrintReturnCount += result.PrintReturnCount
		aggregate.LoadCharacterAddresses = append(aggregate.LoadCharacterAddresses, result.LoadCharacterAddresses...)
		aggregate.FindItemIDs = append(aggregate.FindItemIDs, result.FindItemIDs...)
		aggregate.DestroyItemIDs = append(aggregate.DestroyItemIDs, result.DestroyItemIDs...)
		aggregate.NPCIDs = append(aggregate.NPCIDs, result.NPCIDs...)
		aggregate.RandomValues = append(aggregate.RandomValues, result.RandomValues...)
		aggregate.EncounterActions = append(aggregate.EncounterActions, result.EncounterActions...)
		aggregate.LoadFilesRequested = aggregate.LoadFilesRequested || result.LoadFilesRequested
		if result.LoadFilesRequested {
			aggregate.LoadFiles = result.LoadFiles
		}
		if result.LoadPiecesRequested {
			aggregate.LoadPiecesRequested = true
			aggregate.LoadPieces = result.LoadPieces
		}
		aggregate.PictureRequested = aggregate.PictureRequested || result.PictureRequested
		if result.PictureRequested {
			aggregate.PictureBlock = result.PictureBlock
			aggregate.BigPictureRequested = result.BigPictureRequested
		}
		aggregate.SpellSearches = append(aggregate.SpellSearches, result.SpellSearches...)
		aggregate.ProtectionRequests = append(aggregate.ProtectionRequests, result.ProtectionRequests...)
		aggregate.ClockRequests = append(aggregate.ClockRequests, result.ClockRequests...)
		aggregate.TreasureRequests = append(aggregate.TreasureRequests, result.TreasureRequests...)
		aggregate.PartyStrengthRequests = append(aggregate.PartyStrengthRequests, result.PartyStrengthRequests...)
		aggregate.PartySurpriseRequests = append(aggregate.PartySurpriseRequests, result.PartySurpriseRequests...)
		aggregate.CheckPartyRequests = append(aggregate.CheckPartyRequests, result.CheckPartyRequests...)
		aggregate.WhoRequests = append(aggregate.WhoRequests, result.WhoRequests...)
		selectionOffset += result.SelectionsConsumed
		s.selectionOffset = selectionOffset
		if runErr != nil {
			return aggregate, runErr
		}
		if result.NewECLBlockID == nil {
			aggregate.WaitingForMenu = result.WaitingForMenu
			return aggregate, nil
		}
		if err := s.ApplyResult(result); err != nil {
			return aggregate, err
		}
		start, err = s.InitialEntry()
		if err != nil {
			return aggregate, err
		}
		runtime = s.states[s.current]
		runtime.PC = start
		runtime.Started = true
	}
	return aggregate, fmt.Errorf("ECL session exceeded block transition limit")
}
