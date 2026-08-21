package ecl

import (
	"fmt"

	"github.com/wicanr2/golden-box-remake-engine/randomstream"
)

// BlockSession owns decoded ECL blocks and the current block identity. It is
// intentionally separate from RunSubset: VM execution state can be extended
// later without confusing a DAX block switch with a fallthrough branch.
type BlockSession struct {
	blocks             map[uint8][]byte
	current            uint8
	states             map[uint8]*RuntimeState
	selectionOffset    int
	whoSelectionOffset int
	stringInputOffset  int
	// blockScratch 依 block 保存 `4C00h`..`4C0Fh` 這 16 格。整個 session 共用
	// 一份會讓世界地圖的旅行時鐘把地城的一次性旗標覆蓋掉（spec 1162）。
	blockScratch map[uint8]map[uint16]uint16
}

// per-block scratch 的位址範圍。整個區 0 是 `4B00h`..`4EFFh`（spec 1096），
// 但只有開頭這 16 格是每一段自己的暫存：`cmd/ecl-cell-refs -range 4C00-4C0F`
// 顯示這 16 格**沒有任何唯讀的消費者**——會讀的段一定自己也寫。`4C10h` 以上
// 反過來滿是跨段交接（`4CE3h` 由 `ECL4/0x20` 寫、`ECL4/0x21`／`0x22` 讀；
// `4C59h`／`4C5Ah` 由巫師塔與眼魔洞穴寫、世界地圖讀），必須留在共用區。
// 推導見 spec 1162。
const (
	blockScratchLow  = 0x4C00
	blockScratchHigh = 0x4C0F
)

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
	session := &BlockSession{
		blocks:       owned,
		current:      current,
		states:       states,
		blockScratch: map[uint8]map[uint16]uint16{},
	}
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
// MemorySnapshot 回傳目前這一段看得到的整份記憶體（含共用區與這一段的暫存），
// 不含停在旁邊那幾段的暫存。給存檔編碼器用。
func (s *BlockSession) MemorySnapshot() map[uint16]uint16 {
	runtime := s.states[s.current]
	if runtime == nil {
		return nil
	}
	memory := make(map[uint16]uint16, len(runtime.Memory))
	for address, value := range runtime.Memory {
		memory[address] = value
	}
	return memory
}

// BlockMemoryValue 讀「某一段自己那份」暫存格。隊伍已經換到別段時，
// `MemoryValue` 看不到停在那邊的值——那正是 per-block scratch 的本意
// （spec 1162）。位址不在暫存區內就退回 `MemoryValue`。
func (s *BlockSession) BlockMemoryValue(blockID uint8, address uint16) (uint16, bool) {
	if address < blockScratchLow || address > blockScratchHigh || blockID == s.current {
		return s.MemoryValue(address)
	}
	bank, ok := s.blockScratch[blockID]
	if !ok {
		return 0, false
	}
	value, ok := bank[address]
	return value, ok
}

func (s *BlockSession) SetMemoryValue(address, value uint16) {
	state := s.states[s.current]
	if state == nil {
		return
	}
	state.Memory[address] = value
}

// ResetRandomSeed starts a new deterministic PRNG stream without resetting
// ECL memory or the resumable PC. Normal gameplay does not call this between
// entries; it exists for an explicit replay/test seed change.
func (s *BlockSession) ResetRandomSeed(seed int64) {
	runtime := s.states[s.current]
	if runtime == nil {
		return
	}
	runtime.Random = randomstream.New(seed)
}

func (s *BlockSession) Switch(id uint8) error {
	if _, ok := s.blocks[id]; !ok {
		return fmt.Errorf("ECL session target block 0x%02X is unavailable", id)
	}
	if id != s.current {
		s.swapBlockScratch(id)
	}
	s.current = id
	if _, ok := s.states[id]; !ok {
		s.states[id] = s.states[s.current]
	}
	s.loadCurrentCodeMemory()
	return nil
}

// swapBlockScratch 把目前這一段的暫存格收起來，換上目標那一段自己那份。沒跑過
// 的段從全 0 開始。`4C10h` 以上不動，跨段交接（`4CE3h`、`4C59h`、`4C5Ah`）才
// 留得住。
func (s *BlockSession) swapBlockScratch(next uint8) {
	runtime := s.states[s.current]
	if runtime == nil {
		return
	}
	if s.blockScratch == nil {
		s.blockScratch = map[uint8]map[uint16]uint16{}
	}
	saved := make(map[uint16]uint16)
	for address := uint16(blockScratchLow); address <= blockScratchHigh; address++ {
		if value, ok := runtime.Memory[address]; ok {
			saved[address] = value
		}
		delete(runtime.Memory, address)
	}
	s.blockScratch[s.current] = saved
	for address, value := range s.blockScratch[next] {
		runtime.Memory[address] = value
	}
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
	s.blockScratch = map[uint8]map[uint16]uint16{}
	s.current = id
	s.selectionOffset = 0
	s.whoSelectionOffset = 0
	s.stringInputOffset = 0
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
	return s.RunInteractiveSeedWithPartyContextAndInputs(maxSteps, selections, whoSelections, nil, seed, context)
}

func (s *BlockSession) RunInteractiveSeedWithPartyContextAndInputs(maxSteps int, selections, whoSelections []uint16, stringInputs []string, seed int64, context PartyContext) (RunResult, error) {
	start, err := s.InitialEntry()
	if err != nil {
		return RunResult{}, err
	}
	owned := context.clone()
	return s.runFromSeedWithPartyContextAndInputs(start, maxSteps, selections, whoSelections, stringInputs, seed, &owned)
}

// ResumeInteractiveSelectionSeed supplies only the choices made at the
// currently paused UI boundary. BlockSession owns the cumulative offsets;
// callers must not reconstruct a growing history of synthetic Continue,
// menu, and WHO inputs themselves.
func (s *BlockSession) ResumeInteractiveSelectionSeed(maxSteps int, selection, whoSelection *uint16, seed int64, context PartyContext) (RunResult, error) {
	return s.ResumeInteractiveInputSeed(maxSteps, selection, whoSelection, nil, seed, context)
}

// ResumeInteractiveInputSeed supplies at most one value for each currently
// paused UI kind. The session retains cumulative offsets so a resumed opcode
// consumes the new value exactly once without replaying prior menu, WHO, or
// INPUT STRING transactions.
func (s *BlockSession) ResumeInteractiveInputSeed(maxSteps int, selection, whoSelection *uint16, stringInput *string, seed int64, context PartyContext) (RunResult, error) {
	selections := make([]uint16, s.selectionOffset)
	if selection != nil {
		selections = append(selections, *selection)
	}
	whoSelections := make([]uint16, s.whoSelectionOffset)
	if whoSelection != nil {
		whoSelections = append(whoSelections, *whoSelection)
	}
	stringInputs := make([]string, s.stringInputOffset)
	if stringInput != nil {
		stringInputs = append(stringInputs, *stringInput)
	}
	return s.RunInteractiveSeedWithPartyContextAndInputs(maxSteps, selections, whoSelections, stringInputs, seed, context)
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
	runtime.Compare = [6]bool{}
	s.selectionOffset = 0
	s.whoSelectionOffset = 0
	s.stringInputOffset = 0
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
	return s.runFromSeedWithPartyContextAndInputs(start, maxSteps, selections, whoSelections, nil, seed, partyContext)
}

func (s *BlockSession) runFromSeedWithPartyContextAndInputs(start, maxSteps int, selections, whoSelections []uint16, stringInputs []string, seed int64, partyContext *PartyContext) (RunResult, error) {
	aggregate := RunResult{
		SessionStartBlockID:  s.current,
		SessionEndBlockID:    s.current,
		SessionBlockRangeSet: true,
	}
	var err error
	selectionOffset := s.selectionOffset
	// sequenceBase 把每一段 sub-run 各自從 1 起算的執行序接成一條。
	sequenceBase := 0
	for transitions := 0; transitions < 8; transitions++ {
		aggregate.SessionEndBlockID = s.current
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
		remainingStrings := stringInputs
		if s.stringInputOffset < len(stringInputs) {
			remainingStrings = stringInputs[s.stringInputOffset:]
		} else {
			remainingStrings = nil
		}
		runtime := s.states[s.current]
		if !runtime.Started {
			runtime.PC = start
			// SetMemoryValue may seed Area/player/work memory before the first
			// explicit RunFrom invocation. Marking the runtime started here
			// makes the bounded runner import that shared memory instead of
			// silently replacing it with a fresh empty map.
			runtime.Started = true
		}
		result, runErr := runSubsetWithStateContextAndInputs(s.CurrentData(), start, maxSteps, remaining, remainingWho, remainingStrings, true, seed, runtime, partyContext)
		aggregate.Text = append(aggregate.Text, result.Text...)
		aggregate.Menus = append(aggregate.Menus, result.Menus...)
		aggregate.Steps += result.Steps
		aggregate.PC = result.PC
		aggregate.Exited = aggregate.Exited || result.Exited
		aggregate.CombatRequested = aggregate.CombatRequested || result.CombatRequested
		aggregate.ShopRequested = aggregate.ShopRequested || result.ShopRequested
		if result.ShopRequested {
			aggregate.ShopPriceScale = result.ShopPriceScale
		}
		aggregate.TempleRequested = aggregate.TempleRequested || result.TempleRequested
		if result.MonsterSetup != nil {
			setup := *result.MonsterSetup
			aggregate.MonsterSetup = &setup
		}
		aggregate.MonsterSpawns = append(aggregate.MonsterSpawns, result.MonsterSpawns...)
		aggregate.ProgramIDs = append(aggregate.ProgramIDs, result.ProgramIDs...)
		aggregate.ProgramExit = aggregate.ProgramExit || result.ProgramExit
		aggregate.CallAddresses = append(aggregate.CallAddresses, result.CallAddresses...)
		// 每一段 sub-run 的執行序都從 1 重新算，聚合時要加上前面幾段的總量，
		// 否則跨段之後「誰先發生」會亂掉。
		for _, request := range result.CallRequests {
			request.BlockID = s.current
			request.Sequence += sequenceBase
			aggregate.CallRequests = append(aggregate.CallRequests, request)
		}
		for _, write := range result.SaveWrites {
			write.BlockID = s.current
			write.Sequence += sequenceBase
			aggregate.SaveWrites = append(aggregate.SaveWrites, write)
		}
		sequenceBase += len(result.CallRequests) + len(result.SaveWrites)
		aggregate.DamageRequests = append(aggregate.DamageRequests, result.DamageRequests...)
		aggregate.PrintReturnCount += result.PrintReturnCount
		aggregate.ApproachCount += result.ApproachCount
		aggregate.DelayCount += result.DelayCount
		aggregate.LoadCharacterAddresses = append(aggregate.LoadCharacterAddresses, result.LoadCharacterAddresses...)
		aggregate.LoadCharacterRequests = append(aggregate.LoadCharacterRequests, result.LoadCharacterRequests...)
		aggregate.CombatTeamWrites = append(aggregate.CombatTeamWrites, result.CombatTeamWrites...)
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
			// 這一段自己換過圖，游標已經被 LOADSEQUENCE 設回第 1 格，
			// 前面幾段推的格數不再算數（spec 1150）。
			aggregate.PictureFrameAdvances = result.PictureFrameAdvances
		} else {
			aggregate.PictureFrameAdvances += result.PictureFrameAdvances
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
		aggregate.StringInputRequests = append(aggregate.StringInputRequests, result.StringInputRequests...)
		selectionOffset += result.SelectionsConsumed
		s.whoSelectionOffset += result.WhoSelectionsConsumed
		s.stringInputOffset += result.StringInputsConsumed
		s.selectionOffset = selectionOffset
		if runErr != nil {
			return aggregate, runErr
		}
		if result.NewECLBlockID == nil {
			ApplyCombatTeamWrites(aggregate.MonsterSpawns, aggregate.CombatTeamWrites)
			aggregate.WaitingForMenu = result.WaitingForMenu
			aggregate.WaitingForWho = result.WaitingForWho
			aggregate.WaitingForString = result.WaitingForString
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
