package ecl

import "fmt"

// BlockSession owns decoded ECL blocks and the current block identity. It is
// intentionally separate from RunSubset: VM execution state can be extended
// later without confusing a DAX block switch with a fallthrough branch.
type BlockSession struct {
	blocks             map[uint8][]byte
	current            uint8
	states             map[uint8]*RuntimeState
	selectionOffset    int
	whoSelectionOffset int
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
	session := &BlockSession{blocks: owned, current: current, states: states}
	session.loadCurrentCodeMemory()
	return session, nil
}

func (s *BlockSession) CurrentBlockID() uint8 { return s.current }

func (s *BlockSession) HasBlock(id uint8) bool {
	_, ok := s.blocks[id]
	return ok
}

func (s *BlockSession) CurrentData() []byte {
	return append([]byte(nil), s.blocks[s.current]...)
}

// MemoryValue exposes one word from the shared ECL VM memory for the work
// adapter. Addresses such as 0xC04B..0xC04D are reference engine registers
// written by scripts before control returns to the world loop.
func (s *BlockSession) MemoryValue(address uint16) (uint16, bool) {
	state := s.states[s.current]
	if state == nil {
		return 0, false
	}
	value, ok := state.Memory[address]
	return value, ok
}

// SetMemoryValue synchronizes a work-owned engine register into shared VM
// memory before a lifecycle entry. The work adapter owns address semantics.
func (s *BlockSession) SetMemoryValue(address, value uint16) {
	state := s.states[s.current]
	if state == nil {
		return
	}
	state.Memory[address] = value
}

func (s *BlockSession) Switch(id uint8) error {
	if _, ok := s.blocks[id]; !ok {
		return fmt.Errorf("ECL session target block 0x%02X is unavailable", id)
	}
	s.current = id
	if _, ok := s.states[id]; !ok {
		s.states[id] = s.states[s.current]
	}
	s.loadCurrentCodeMemory()
	return nil
}

// Reset starts a fresh VM context at one available block. This is distinct
// from NEWECL/Switch, which must preserve shared memory and resumable PC.
func (s *BlockSession) Reset(id uint8) error {
	if _, ok := s.blocks[id]; !ok {
		return fmt.Errorf("ECL session reset block 0x%02X is unavailable", id)
	}
	shared := NewRuntimeState(0)
	for blockID := range s.blocks {
		s.states[blockID] = shared
	}
	s.current = id
	s.selectionOffset = 0
	s.whoSelectionOffset = 0
	s.loadCurrentCodeMemory()
	return nil
}

// loadCurrentCodeMemory mirrors the reference EclBlock mapping: the decoded
// payload is byte-addressable at 0x8000..0x9DFF. ECL GETTABLE routinely reads
// dispatch bytes from this region, so it cannot be represented as an empty
// generic word map. Loading another block replaces this code window while
// preserving Area/player/shared memory outside it.
func (s *BlockSession) loadCurrentCodeMemory() {
	runtime := s.states[s.current]
	if runtime == nil {
		return
	}
	for address := uint16(CodeAddressBase); address <= 0x9DFF; address++ {
		delete(runtime.Memory, address)
	}
	data := s.blocks[s.current]
	if len(data) < 2 {
		return
	}
	for index, value := range data[2:] {
		address := CodeAddressBase + index
		if address > 0x9DFF {
			break
		}
		runtime.Memory[uint16(address)] = uint16(value)
	}
}

// InitialEntry returns the fifth vm_init_ecl entry for the current block as a
// decoded payload offset.
func (s *BlockSession) InitialEntry() (int, error) {
	return s.EntryPoint(4)
}

// EntryPoint returns one of vm_init_ecl's five lifecycle entries:
// per-turn, search-location, pre-camp, camp-interrupted, and initial.
func (s *BlockSession) EntryPoint(index int) (int, error) {
	if index < 0 || index >= 5 {
		return 0, fmt.Errorf("ECL lifecycle entry index %d is outside 0..4", index)
	}
	points, _, err := EntryPoints(s.blocks[s.current], 5)
	if err != nil {
		return 0, err
	}
	return int(points[index]) - CodeAddressBase, nil
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
	owned := context.clone()
	return s.runFromSeedWithPartyContext(start, maxSteps, selections, seed, &owned)
}

func (s *BlockSession) RunInteractiveSeedWithPartyContextAndWhoSelections(maxSteps int, selections, whoSelections []uint16, seed int64, context PartyContext) (RunResult, error) {
	start, err := s.InitialEntry()
	if err != nil {
		return RunResult{}, err
	}
	owned := context.clone()
	return s.runFromSeedWithPartyContextAndWhoSelections(start, maxSteps, selections, whoSelections, seed, &owned)
}

// RunFrom executes an explicit event entry in the current block. After a
// NEWECL signal, the target resumes at its own initial entry.
func (s *BlockSession) RunFrom(start, maxSteps int, selections []uint16) (RunResult, error) {
	return s.runFromSeed(start, maxSteps, selections, 1)
}

// RunEntry starts a fresh invocation at one lifecycle entry while preserving
// shared VM memory. This mirrors separate RunEclVm calls after an earlier
// entry has EXITed; RunFrom remains the resumable transaction primitive.
func (s *BlockSession) RunEntry(index, maxSteps int, selections []uint16) (RunResult, error) {
	return s.RunEntrySeedWithPartyContext(index, maxSteps, selections, nil, 1, PartyContext{})
}

// RunEntrySeedWithPartyContext is the party-aware lifecycle counterpart to
// the resumable interactive APIs.
func (s *BlockSession) RunEntrySeedWithPartyContext(index, maxSteps int, selections, whoSelections []uint16, seed int64, context PartyContext) (RunResult, error) {
	start, err := s.EntryPoint(index)
	if err != nil {
		return RunResult{}, err
	}
	runtime := s.states[s.current]
	runtime.PC = start
	runtime.Started = true
	runtime.Stack = runtime.Stack[:0]
	owned := context.clone()
	return s.runFromSeedWithPartyContextAndWhoSelections(start, maxSteps, selections, whoSelections, seed, &owned)
}

func (s *BlockSession) runFromSeed(start, maxSteps int, selections []uint16, seed int64) (RunResult, error) {
	return s.runFromSeedWithPartyContext(start, maxSteps, selections, seed, nil)
}

func (s *BlockSession) runFromSeedWithPartyContext(start, maxSteps int, selections []uint16, seed int64, partyContext *PartyContext) (RunResult, error) {
	return s.runFromSeedWithPartyContextAndWhoSelections(start, maxSteps, selections, nil, seed, partyContext)
}

func (s *BlockSession) runFromSeedWithPartyContextAndWhoSelections(start, maxSteps int, selections, whoSelections []uint16, seed int64, partyContext *PartyContext) (RunResult, error) {
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
		remainingWho := whoSelections
		if s.whoSelectionOffset < len(whoSelections) {
			remainingWho = whoSelections[s.whoSelectionOffset:]
		} else {
			remainingWho = nil
		}
		runtime := s.states[s.current]
		if !runtime.Started {
			runtime.PC = start
		}
		result, runErr := runSubsetWithStateContextAndWhoSelections(s.CurrentData(), start, maxSteps, remaining, remainingWho, true, seed, runtime, partyContext)
		aggregate.Text = append(aggregate.Text, result.Text...)
		aggregate.Menus = append(aggregate.Menus, result.Menus...)
		aggregate.Steps += result.Steps
		aggregate.PC = result.PC
		aggregate.Exited = aggregate.Exited || result.Exited
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
		aggregate.DelayCount += result.DelayCount
		aggregate.LoadCharacterAddresses = append(aggregate.LoadCharacterAddresses, result.LoadCharacterAddresses...)
		aggregate.LoadCharacterRequests = append(aggregate.LoadCharacterRequests, result.LoadCharacterRequests...)
		aggregate.FindItemIDs = append(aggregate.FindItemIDs, result.FindItemIDs...)
		aggregate.FindItemRequests = append(aggregate.FindItemRequests, result.FindItemRequests...)
		aggregate.FindSpecialRequests = append(aggregate.FindSpecialRequests, result.FindSpecialRequests...)
		aggregate.DumpRequests = append(aggregate.DumpRequests, result.DumpRequests...)
		aggregate.DestroyItemIDs = append(aggregate.DestroyItemIDs, result.DestroyItemIDs...)
		aggregate.NPCIDs = append(aggregate.NPCIDs, result.NPCIDs...)
		aggregate.NPCRequests = append(aggregate.NPCRequests, result.NPCRequests...)
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
			aggregate.PictureHeadBlock = result.PictureHeadBlock
			aggregate.PictureHeadBlockSet = result.PictureHeadBlockSet
		}
		aggregate.SpellSearches = append(aggregate.SpellSearches, result.SpellSearches...)
		aggregate.ProtectionRequests = append(aggregate.ProtectionRequests, result.ProtectionRequests...)
		aggregate.ClockRequests = append(aggregate.ClockRequests, result.ClockRequests...)
		aggregate.TreasureRequests = append(aggregate.TreasureRequests, result.TreasureRequests...)
		aggregate.RobRequests = append(aggregate.RobRequests, result.RobRequests...)
		aggregate.PartyStrengthRequests = append(aggregate.PartyStrengthRequests, result.PartyStrengthRequests...)
		aggregate.PartySurpriseRequests = append(aggregate.PartySurpriseRequests, result.PartySurpriseRequests...)
		aggregate.CheckPartyRequests = append(aggregate.CheckPartyRequests, result.CheckPartyRequests...)
		aggregate.WhoRequests = append(aggregate.WhoRequests, result.WhoRequests...)
		selectionOffset += result.SelectionsConsumed
		s.whoSelectionOffset += result.WhoSelectionsConsumed
		s.selectionOffset = selectionOffset
		if runErr != nil {
			return aggregate, runErr
		}
		if result.NewECLBlockID == nil {
			aggregate.WaitingForMenu = result.WaitingForMenu
			aggregate.WaitingForWho = result.WaitingForWho
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
