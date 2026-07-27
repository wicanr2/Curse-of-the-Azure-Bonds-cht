package ecl

import (
	"fmt"
	"math/rand"
)

// RunResult is the observable output of the bounded ECL subset runner.
// It deliberately exposes text and stop position, not DOS rendering state.
type RunResult struct {
	Text                   []string
	Menus                  []Menu
	PC                     int
	Steps                  int
	WaitingForMenu         bool
	WaitingForWho          bool
	NewECLBlockID          *uint8
	CombatRequested        bool
	MonsterSetup           *MonsterSetup
	MonsterSpawns          []MonsterSpawn
	ProgramIDs             []uint8
	ProgramExit            bool
	CallAddresses          []uint16
	DamageRequests         []DamageRequest
	PrintReturnCount       int
	LoadCharacterAddresses []uint16
	LoadCharacterRequests  []LoadCharacterRequest
	FindItemIDs            []uint16
	FindItemRequests       []FindItemRequest
	FindSpecialRequests    []FindSpecialRequest
	DestroyItemIDs         []uint16
	NPCIDs                 []uint16
	SelectionsConsumed     int
	WhoSelectionsConsumed  int
	RandomValues           []uint16
	EncounterActions       []uint16
	LoadFilesRequested     bool
	LoadFiles              [3]uint16
	LoadPiecesRequested    bool
	LoadPieces             [3]uint16
	PictureRequested       bool
	PictureBlock           uint16
	BigPictureRequested    bool
	SpellSearches          []SpellSearch
	ProtectionRequests     []uint16
	ClockRequests          []ClockRequest
	TreasureRequests       []TreasureRequest
	PartyStrengthRequests  []PartyStrengthRequest
	PartySurpriseRequests  []PartySurpriseRequest
	CheckPartyRequests     []CheckPartyRequest
	WhoRequests            []WhoRequest
}

type PartyStrengthRequest struct {
	Destination uint16
	Value       uint16
	Resolved    bool
}

type CheckPartyRequest struct {
	Query        uint16
	AffectID     uint16
	Destinations [4]uint16
	Minimum      uint16
	Maximum      uint16
	Average      uint16
	AffectFound  bool
	Resolved     bool
}

// LoadCharacterRequest is the decoded boundary for reference LOAD CHARACTER.
// Address preserves the raw word operand; Value is vm_GetCmdValue's result.
// The low 7 bits select the 1-based party member and bit 7 is the reference
// restore/redraw flag.
type LoadCharacterRequest struct {
	Address     uint16
	Value       uint16
	PlayerIndex uint8
	HighBitSet  bool
}

// FindItemRequest records the reference party-wide item query. Resolved is
// false when the bounded VM has no injected party inventory context.
type FindItemRequest struct {
	ItemID   uint16
	Found    bool
	Resolved bool
}

// FindSpecialRequest records FIND SPECIAL against the current selected
// player's active affects.
type FindSpecialRequest struct {
	AffectID            uint16
	SelectedPlayerIndex int
	Found               bool
	Resolved            bool
}

// WhoRequest marks the reference character-selection boundary. WHO consumes
// the current ECL prompt text but its player selection belongs to the UI/state
// adapter rather than a normal HORIZONTAL/VERTICAL MENU.
type WhoRequest struct {
	Prompt            string
	Selected          uint16
	SelectionProvided bool
}

// PartySurpriseRequest preserves the two destination addresses used by the
// reference PARTY SURPRISE command. The party/ranger calculation belongs to
// the game adapter because the bounded VM has no roster context.
type PartySurpriseRequest struct {
	RangerDestination uint16
	OtherDestination  uint16
	RangerValue       uint16
	OtherValue        uint16
	Resolved          bool
}

// PartyMemberContext contains only the values needed by the verified party
// commands. It deliberately avoids importing party/combat packages so the ECL
// VM remains reusable by other Gold Box works.
type PartyMemberContext struct {
	Name              string
	ItemTypes         []uint8
	HitPoints         int
	ArmorClass        int
	AttackBonus       int
	ClericLevel       int
	MagicUserLevel    int
	HasRangerClass    bool
	ThiefSkills       [8]uint8
	MovementAllowance int
	Effects           []uint8
}

type PartyContext struct {
	Members []PartyMemberContext
}

func (c PartyContext) partyStrength() uint16 {
	var value int
	for _, member := range c.Members {
		armorClass := member.ArmorClass
		if armorClass > 60 {
			armorClass -= 60
		} else {
			armorClass = 0
		}
		hitBonus := member.AttackBonus
		if hitBonus > 39 {
			hitBonus -= 39
		} else {
			hitBonus = 0
		}
		value += ((member.ClericLevel * 4) + member.HitPoints + (armorClass * 5) + (hitBonus * 5) + (member.MagicUserLevel * 8)) / 10
	}
	if value < 0 {
		return 0
	}
	if value > 0xFF {
		value = 0xFF
	}
	return uint16(value)
}

func (c PartyContext) hasRanger() bool {
	for _, member := range c.Members {
		if member.HasRangerClass {
			return true
		}
	}
	return false
}

func (c PartyContext) checkParty(query, affectID uint16) (CheckPartyRequest, bool) {
	request := CheckPartyRequest{Query: query, AffectID: affectID}
	normalized := query - 0x7FFF
	if normalized == 8001 {
		for _, member := range c.Members {
			for _, effect := range member.Effects {
				if uint16(effect) == affectID {
					request.AffectFound = true
				}
			}
		}
		request.Resolved = true
		return request, true
	}
	if normalized >= 0xA5 && normalized <= 0xAC {
		index := int(normalized - 0xA5)
		request.Minimum, request.Maximum, request.Average = partyMetric(c.Members, func(member PartyMemberContext) int {
			return int(member.ThiefSkills[index])
		})
		request.Resolved = true
		return request, true
	}
	if normalized == 0x9F {
		request.Minimum, request.Maximum, request.Average = partyMetric(c.Members, func(member PartyMemberContext) int {
			return member.MovementAllowance
		})
		request.Resolved = true
		return request, true
	}
	return request, false
}

func partyMetric(members []PartyMemberContext, metric func(PartyMemberContext) int) (uint16, uint16, uint16) {
	if len(members) == 0 {
		return 0, 0, 0
	}
	minimum, maximum, total := metric(members[0]), metric(members[0]), 0
	for _, member := range members {
		value := metric(member)
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
		total += value
	}
	return uint16(minimum), uint16(maximum), uint16(total / len(members))
}

// TreasureRequest preserves the eight raw TREASURE operands. The first seven
// values are Copper, Silver, Electrum, Gold, Platinum, Gems and Jewelry;
// ItemBlock selects an ITEM{area}.DAX stock block, or a special/random branch.
type TreasureRequest struct {
	Coins     [7]uint16
	ItemBlock uint16
}

// ClockRequest is the raw two-operand signal emitted by ECL CLOCK. The game
// adapter owns the clock and effect expiration; the VM only decodes it.
type ClockRequest struct {
	TimeStep uint16
	TimeSlot uint16
}

// DamageRequest preserves the five numeric operands consumed by the original
// ECL DAMAGE command. Flags encode target count/selection and saving-throw
// behavior in the DOS engine; the bounded VM leaves those rules to a
// party/combat adapter instead of rolling or mutating HP here.
type DamageRequest struct {
	Flags     uint16
	DiceCount uint16
	DiceSize  uint16
	Bonus     uint16
	SaveFlags uint16
}

// SpellSearch is the data-bearing part of ECL SPELL. The bounded runner keeps
// the requested spell and destination addresses; a party spell-slot resolver
// can later fill the reference result values.
type SpellSearch struct {
	SpellID          uint16
	SpellSlotAddress uint16
	CharacterAddress uint16
}

type Menu struct {
	Location uint16
	Options  []string
	Selected uint16
	Vertical bool
	Prompt   string
}

// RuntimeState is the resumable portion of the bounded ECL VM. A menu pause
// must not discard the memory writes, comparison flags, or call stack that led
// to it; otherwise feeding the next choice replays the event from its entry
// with a different machine state.
type RuntimeState struct {
	PC                  int
	Started             bool
	Stack               []int
	Memory              map[uint16]uint16
	Strings             map[uint16]string
	Compare             [6]bool
	SelectedPlayerIndex int
	SelectedPlayerSet   bool
}

func NewRuntimeState(start int) *RuntimeState {
	return &RuntimeState{
		PC:      start,
		Memory:  make(map[uint16]uint16),
		Strings: make(map[uint16]string),
	}
}

// RunSubset executes only commands whose semantics are represented here.
// Unsupported commands return an error at their exact payload offset. This is
// useful for proving an event prefix without silently treating the whole ECL
// program as implemented.
func RunSubset(block []byte, start, maxSteps int) (RunResult, error) {
	return runSubset(block, start, maxSteps, nil, false, 1)
}

// RunSubsetWithSelections is RunSubset with deterministic selections for
// successive HORIZONTAL MENU commands. A missing selection keeps the safe
// default index 0. This is the bridge between ECL menu semantics and a UI;
// it does not pretend to implement the original blocking DOS input routine.
func RunSubsetWithSelections(block []byte, start, maxSteps int, selections []uint16) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, false, 1)
}

// RunSubsetWithSelectionsSeed is the deterministic variant used when an ECL
// event contains RANDOM. It keeps the legacy wrapper behavior stable while
// allowing a game session or regression to choose its own replay seed.
func RunSubsetWithSelectionsSeed(block []byte, start, maxSteps int, selections []uint16, seed int64) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, false, seed)
}

// RunSubsetInteractive pauses at the first menu whose selection is not
// supplied. This allows a UI to feed one choice per frame/event instead of
// silently choosing index zero for all later menus.
func RunSubsetInteractive(block []byte, start, maxSteps int, selections []uint16) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, true, 1)
}

// RunSubsetInteractiveSeed is the seeded interactive runner for ECL RANDOM.
func RunSubsetInteractiveSeed(block []byte, start, maxSteps int, selections []uint16, seed int64) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, true, seed)
}

// RunSubsetInteractiveSeedWithPartyContext resolves verified party commands
// against the supplied roster while preserving the seeded interactive API.
func RunSubsetInteractiveSeedWithPartyContext(block []byte, start, maxSteps int, selections []uint16, seed int64, context PartyContext) (RunResult, error) {
	return runSubsetWithStateContext(block, start, maxSteps, selections, true, seed, NewRuntimeState(start), &context)
}

func RunSubsetInteractiveSeedWithPartyContextAndWhoSelections(block []byte, start, maxSteps int, selections, whoSelections []uint16, seed int64, context PartyContext) (RunResult, error) {
	return runSubsetWithStateContextAndWhoSelections(block, start, maxSteps, selections, whoSelections, true, seed, NewRuntimeState(start), &context)
}

func runSubset(block []byte, start, maxSteps int, selections []uint16, pauseOnMissing bool, seed int64) (RunResult, error) {
	return runSubsetWithState(block, start, maxSteps, selections, pauseOnMissing, seed, NewRuntimeState(start))
}

func runSubsetWithState(block []byte, start, maxSteps int, selections []uint16, pauseOnMissing bool, seed int64, runtime *RuntimeState) (RunResult, error) {
	return runSubsetWithStateContext(block, start, maxSteps, selections, pauseOnMissing, seed, runtime, nil)
}

func runSubsetWithStateContext(block []byte, start, maxSteps int, selections []uint16, pauseOnMissing bool, seed int64, runtime *RuntimeState, partyContext *PartyContext) (RunResult, error) {
	return runSubsetWithStateContextAndWhoSelections(block, start, maxSteps, selections, nil, pauseOnMissing, seed, runtime, partyContext)
}

func runSubsetWithStateContextAndWhoSelections(block []byte, start, maxSteps int, selections, whoSelections []uint16, pauseOnMissing bool, seed int64, runtime *RuntimeState, partyContext *PartyContext) (RunResult, error) {
	if len(block) < 2 {
		return RunResult{}, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	if start < 0 || start >= len(payload) {
		return RunResult{}, fmt.Errorf("runtime start %d is outside payload", start)
	}
	if maxSteps <= 0 {
		return RunResult{}, fmt.Errorf("runtime step limit must be positive")
	}

	pc := start
	stack := make([]int, 0)
	memory := make(map[uint16]uint16)
	stringsMemory := make(map[uint16]string)
	partyItems := make(map[uint16]bool)
	if partyContext != nil {
		for _, member := range partyContext.Members {
			for _, itemType := range member.ItemTypes {
				partyItems[uint16(itemType)] = true
			}
		}
	}
	rng := rand.New(rand.NewSource(seed))
	var compare [6]bool
	selectedPlayerIndex := -1
	selectedPlayerSet := false
	if runtime != nil && runtime.Started {
		pc = runtime.PC
		stack = append(stack, runtime.Stack...)
		for address, value := range runtime.Memory {
			memory[address] = value
		}
		for address, value := range runtime.Strings {
			stringsMemory[address] = value
		}
		compare = runtime.Compare
		selectedPlayerIndex = runtime.SelectedPlayerIndex
		selectedPlayerSet = runtime.SelectedPlayerSet
	}
	saveState := func(nextPC int) {
		if runtime == nil {
			return
		}
		runtime.PC = nextPC
		runtime.Started = true
		runtime.Stack = append(runtime.Stack[:0], stack...)
		runtime.Memory = memory
		runtime.Strings = stringsMemory
		runtime.Compare = compare
		runtime.SelectedPlayerIndex = selectedPlayerIndex
		runtime.SelectedPlayerSet = selectedPlayerSet
	}
	selectionCursor := 0
	whoSelectionCursor := 0
	result := RunResult{PC: pc}
	for result.Steps < maxSteps {
		instruction, err := decodeInstruction(payload, pc)
		if err != nil {
			result.PC = pc
			return result, err
		}
		result.Steps++
		next := instruction.Next
		switch instruction.Command.Opcode {
		case 0x00: // EXIT
			result.PC = next
			return result, nil
		case 0x01, 0x02: // GOTO / GOSUB
			target, ok := CodeTarget(instruction.Operands[0], len(payload))
			if !ok {
				return result, fmt.Errorf("opcode 0x%02X at %d has invalid code target", instruction.Command.Opcode, pc)
			}
			if instruction.Command.Opcode == 0x02 {
				stack = append(stack, next)
			}
			pc = target
			continue
		case 0x03: // COMPARE
			if operandIsText(instruction.Operands[0]) || operandIsText(instruction.Operands[1]) {
				left, err := operandText(instruction.Operands[0], stringsMemory)
				if err != nil {
					return result, fmt.Errorf("string compare at %d: %w", pc, err)
				}
				right, err := operandText(instruction.Operands[1], stringsMemory)
				if err != nil {
					return result, fmt.Errorf("string compare at %d: %w", pc, err)
				}
				compare[0] = left == right
				compare[1] = left != right
				compare[2] = left < right
				compare[3] = left > right
				compare[4] = left <= right
				compare[5] = left >= right
				break
			}
			left, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("compare at %d: %w", pc, err)
			}
			right, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("compare at %d: %w", pc, err)
			}
			compare[0] = left == right
			compare[1] = left != right
			compare[2] = left < right
			compare[3] = left > right
			compare[4] = left <= right
			compare[5] = left >= right
		case 0x04, 0x05, 0x06, 0x07, 0x2F, 0x30: // arithmetic / AND / OR
			if !instruction.Operands[2].WordSet {
				return result, fmt.Errorf("arithmetic at %d has non-address destination", pc)
			}
			left, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("arithmetic at %d: %w", pc, err)
			}
			right, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("arithmetic at %d: %w", pc, err)
			}
			var value uint16
			switch instruction.Command.Opcode {
			case 0x04:
				value = left + right
			case 0x05:
				value = right - left
			case 0x06:
				if right == 0 {
					return result, fmt.Errorf("arithmetic at %d divides by zero", pc)
				}
				value = left / right
			case 0x07:
				value = left * right
			case 0x2F:
				value = left & right
			case 0x30:
				value = left | right
			}
			memory[instruction.Operands[2].Word] = value
		case 0x08: // RANDOM
			if !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("random at %d has non-address destination", pc)
			}
			maximum, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("random at %d: %w", pc, err)
			}
			// The original increments every maximum below 0xFF, making the
			// command an inclusive range [0, maximum].
			bound := int(maximum)
			if bound < 0xFF {
				bound++
			}
			if bound == 0 {
				bound = 1
			}
			value := uint16(rng.Intn(bound))
			memory[instruction.Operands[1].Word] = value
			result.RandomValues = append(result.RandomValues, value)
		case 0x14: // COMPARE AND
			leftA, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			rightA, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			leftB, err := operandValue(instruction.Operands[2], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			rightB, err := operandValue(instruction.Operands[3], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			for i := range compare {
				compare[i] = false
			}
			if leftA == rightA && leftB == rightB {
				compare[0] = true
			} else {
				compare[1] = true
			}
		case 0x2A: // GETTABLE
			if !instruction.Operands[0].WordSet || !instruction.Operands[2].WordSet {
				return result, fmt.Errorf("GETTABLE at %d has non-address operand", pc)
			}
			index, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("GETTABLE at %d: %w", pc, err)
			}
			value := memory[instruction.Operands[0].Word+index]
			memory[instruction.Operands[2].Word] = value
		case 0x29: // ENCOUNTER MENU
			if len(instruction.Operands) != 14 {
				return result, fmt.Errorf("encounter menu at %d has %d operands", pc, len(instruction.Operands))
			}
			if !instruction.Operands[3].WordSet {
				return result, fmt.Errorf("encounter menu at %d has no memory destination", pc)
			}
			maxDistance, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("encounter menu distance at %d: %w", pc, err)
			}
			prompt := ""
			for index := 9; index <= 11; index++ {
				if operandIsText(instruction.Operands[index]) {
					prompt, err = operandText(instruction.Operands[index], stringsMemory)
					if err != nil {
						return result, fmt.Errorf("encounter menu prompt at %d: %w", pc, err)
					}
					if prompt != "" {
						break
					}
				}
			}
			options := []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}
			if maxDistance == 0 {
				options[3] = "PARLAY"
			}
			menu := Menu{Location: instruction.Operands[3].Word, Options: options, Prompt: prompt}
			if pauseOnMissing && selectionCursor >= len(selections) {
				result.Menus = append(result.Menus, menu)
				result.WaitingForMenu = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if selectionCursor < len(selections) && selections[selectionCursor] < uint16(len(options)) {
				menu.Selected = selections[selectionCursor]
			}
			selectionCursor++
			result.SelectionsConsumed = selectionCursor
			mappingIndex := int(menu.Selected)
			mapping, err := operandValue(instruction.Operands[4+mappingIndex], memory)
			if err != nil {
				return result, fmt.Errorf("encounter menu action at %d: %w", pc, err)
			}
			memory[instruction.Operands[3].Word] = mapping
			result.EncounterActions = append(result.EncounterActions, mapping)
			result.Menus = append(result.Menus, menu)
		case 0x2B: // HORIZONTAL MENU
			header, headNext, err := ParseOperands(payload, pc, 2)
			if err != nil {
				return result, fmt.Errorf("HORIZONTAL MENU header at %d: %w", pc, err)
			}
			if !header[0].WordSet {
				return result, fmt.Errorf("HORIZONTAL MENU at %d has non-address destination", pc)
			}
			count, err := operandValue(header[1], memory)
			if err != nil {
				return result, fmt.Errorf("HORIZONTAL MENU count at %d: %w", pc, err)
			}
			if count == 0 || count > 64 {
				return result, fmt.Errorf("HORIZONTAL MENU at %d has invalid option count %d", pc, count)
			}
			stringOperands, stringsEnd, err := ParseOperands(payload, headNext-1, int(count))
			if err != nil {
				return result, fmt.Errorf("HORIZONTAL MENU strings at %d: %w", pc, err)
			}
			menu := Menu{Location: header[0].Word, Options: make([]string, 0, count)}
			for _, operand := range stringOperands {
				message, err := operandText(operand, stringsMemory)
				if err != nil {
					return result, fmt.Errorf("HORIZONTAL MENU option at %d: %w", pc, err)
				}
				menu.Options = append(menu.Options, message)
			}
			if pauseOnMissing && selectionCursor >= len(selections) {
				result.Menus = append(result.Menus, menu)
				result.WaitingForMenu = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if selectionCursor < len(selections) && selections[selectionCursor] < count {
				menu.Selected = selections[selectionCursor]
			}
			selectionCursor++
			result.SelectionsConsumed = selectionCursor
			memory[menu.Location] = menu.Selected
			result.Menus = append(result.Menus, menu)
			next = stringsEnd
		case 0x15: // VERTICAL MENU
			header, headNext, err := ParseOperands(payload, pc, 3)
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU header at %d: %w", pc, err)
			}
			if !header[0].WordSet {
				return result, fmt.Errorf("VERTICAL MENU at %d has non-address destination", pc)
			}
			count, err := operandValue(header[2], memory)
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU count at %d: %w", pc, err)
			}
			if count == 0 || count > 64 {
				return result, fmt.Errorf("VERTICAL MENU at %d has invalid option count %d", pc, count)
			}
			prompt, err := operandText(header[1], stringsMemory)
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU prompt at %d: %w", pc, err)
			}
			stringOperands, stringsEnd, err := ParseOperands(payload, headNext-1, int(count))
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU strings at %d: %w", pc, err)
			}
			menu := Menu{Location: header[0].Word, Options: make([]string, 0, count), Vertical: true, Prompt: prompt}
			for _, operand := range stringOperands {
				message, err := operandText(operand, stringsMemory)
				if err != nil {
					return result, fmt.Errorf("VERTICAL MENU option at %d: %w", pc, err)
				}
				menu.Options = append(menu.Options, message)
			}
			if pauseOnMissing && selectionCursor >= len(selections) {
				result.Menus = append(result.Menus, menu)
				result.WaitingForMenu = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if selectionCursor < len(selections) && selections[selectionCursor] < count {
				menu.Selected = selections[selectionCursor]
			}
			selectionCursor++
			result.SelectionsConsumed = selectionCursor
			memory[menu.Location] = menu.Selected
			result.Menus = append(result.Menus, menu)
			next = stringsEnd
		case 0x09: // SAVE
			if !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("save at %d has non-address destination", pc)
			}
			if operandIsText(instruction.Operands[0]) {
				value, err := operandText(instruction.Operands[0], stringsMemory)
				if err != nil {
					return result, fmt.Errorf("save at %d: %w", pc, err)
				}
				stringsMemory[instruction.Operands[1].Word] = value
			} else {
				value, err := operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("save at %d: %w", pc, err)
				}
				memory[instruction.Operands[1].Word] = value
			}
		case 0x11, 0x12: // PRINT / PRINTCLEAR
			if len(instruction.Operands) != 1 {
				return result, fmt.Errorf("print at %d has unexpected arity", pc)
			}
			operand := instruction.Operands[0]
			if operand.Code == 0x80 {
				result.Text = append(result.Text, DecodePackedText(operand.Packed))
			} else if operand.Code == 0x81 {
				message, err := operandText(operand, stringsMemory)
				if err != nil {
					return result, fmt.Errorf("print at %d: %w", pc, err)
				}
				result.Text = append(result.Text, message)
			} else {
				value, err := operandValue(operand, memory)
				if err != nil {
					return result, fmt.Errorf("print at %d: %w", pc, err)
				}
				result.Text = append(result.Text, fmt.Sprint(value))
			}
		case 0x13: // RETURN
			if len(stack) == 0 {
				result.PC = next
				return result, nil
			}
			pc = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			continue
		case 0x20: // NEWECL
			blockID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("NEWECL at %d: %w", pc, err)
			}
			id := uint8(blockID)
			result.NewECLBlockID = &id
			result.PC = next
			saveState(next)
			return result, nil
		case 0x24: // COMBAT
			// The original engine transfers control to its combat loop here.
			// Expose that control transfer and persist the instruction after it;
			// the game adapter can resume the same ECL event after victory.
			result.CombatRequested = true
			result.PC = next
			saveState(next)
			return result, nil
		case 0x2D: // CALL
			address, err := operandAddress(instruction.Operands[0])
			if err != nil {
				return result, fmt.Errorf("CALL at %d: %w", pc, err)
			}
			// CALL dispatches an engine routine and returns to the next ECL
			// instruction. Keep the address observable while leaving the
			// routine-specific DOS side effect to a later adapter.
			result.CallAddresses = append(result.CallAddresses, address)
		case 0x2E: // DAMAGE
			if len(instruction.Operands) != 5 {
				return result, fmt.Errorf("DAMAGE at %d has unexpected arity", pc)
			}
			values := make([]uint16, len(instruction.Operands))
			for index, operand := range instruction.Operands {
				value, err := operandValue(operand, memory)
				if err != nil {
					return result, fmt.Errorf("DAMAGE operand %d at %d: %w", index+1, pc, err)
				}
				values[index] = value
			}
			// The public CoAB reference confirms the order as flags, dice count,
			// dice size, damage bonus, and saving-throw flags. Keep raw values;
			// signed bonus conversion, target selection, saves, and HP mutation
			// belong to the game adapter.
			result.DamageRequests = append(result.DamageRequests, DamageRequest{
				Flags: values[0], DiceCount: values[1], DiceSize: values[2],
				Bonus: values[3], SaveFlags: values[4],
			})
		case 0x34: // ECL CLOCK
			if len(instruction.Operands) != 2 {
				return result, fmt.Errorf("ECL CLOCK at %d has unexpected arity", pc)
			}
			timeStep, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("ECL CLOCK time step at %d: %w", pc, err)
			}
			timeSlot, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("ECL CLOCK time slot at %d: %w", pc, err)
			}
			result.ClockRequests = append(result.ClockRequests, ClockRequest{TimeStep: timeStep, TimeSlot: timeSlot})
		case 0x35: // SAVE TABLE
			if !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("SAVE TABLE at %d has non-address destination", pc)
			}
			value, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("SAVE TABLE value at %d: %w", pc, err)
			}
			offset, err := operandValue(instruction.Operands[2], memory)
			if err != nil {
				return result, fmt.Errorf("SAVE TABLE offset at %d: %w", pc, err)
			}
			// Reference CMD_SaveTable writes value operand 1 to the address
			// operand 2.Word + value operand 3.
			memory[instruction.Operands[1].Word+offset] = value
		case 0x0A: // LOAD CHARACTER
			address, err := operandAddress(instruction.Operands[0])
			if err != nil {
				return result, fmt.Errorf("LOAD CHARACTER at %d: %w", pc, err)
			}
			value, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("LOAD CHARACTER value at %d: %w", pc, err)
			}
			// Keep both raw address and decoded player selector observable; the
			// State adapter owns roster resolution and renderer side effects.
			result.LoadCharacterAddresses = append(result.LoadCharacterAddresses, address)
			result.LoadCharacterRequests = append(result.LoadCharacterRequests, LoadCharacterRequest{
				Address: address, Value: value, PlayerIndex: uint8(value & 0x7F), HighBitSet: value&0x80 != 0,
			})
			if partyContext != nil {
				playerIndex := int(value&0x7F) - 1
				if value&0x7F > 0 && playerIndex >= 0 && playerIndex < len(partyContext.Members) {
					// Reference vm_CopyStringFromMemory treats 0x7C00 as the
					// selected player's name string. Preserve it in RuntimeState
					// so later COMPARE/PRINT operands see the same selection.
					stringsMemory[0x7C00] = partyContext.Members[playerIndex].Name
					selectedPlayerIndex = playerIndex
					selectedPlayerSet = true
				}
			}
		case 0x32: // FIND ITEM
			itemID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("FIND ITEM at %d: %w", pc, err)
			}
			result.FindItemIDs = append(result.FindItemIDs, itemID)
			request := FindItemRequest{ItemID: itemID}
			if partyContext != nil {
				request.Resolved = true
				request.Found = partyItems[itemID]
				for index := range compare {
					compare[index] = false
				}
				compare[0] = request.Found
				compare[1] = !request.Found
			}
			result.FindItemRequests = append(result.FindItemRequests, request)
		case 0x40: // DESTROY ITEMS
			itemID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("DESTROY ITEMS at %d: %w", pc, err)
			}
			// Keep inventory mutation explicit for the party adapter; the VM
			// itself must not silently delete an item without roster context.
			result.DestroyItemIDs = append(result.DestroyItemIDs, itemID)
			if partyContext != nil {
				delete(partyItems, itemID)
			}
		case 0x1D: // PARTYSTRENGTH
			if !instruction.Operands[0].WordSet {
				return result, fmt.Errorf("PARTYSTRENGTH at %d has non-address destination", pc)
			}
			// Reference CMD_PartyStrength computes a byte from the live party
			// (HP, AC, hit bonus, cleric level and magic-user level). Preserve
			// the destination here; the game adapter supplies those roster stats.
			request := PartyStrengthRequest{Destination: instruction.Operands[0].Word}
			if partyContext != nil {
				request.Value = partyContext.partyStrength()
				request.Resolved = true
				memory[request.Destination] = request.Value
			}
			result.PartyStrengthRequests = append(result.PartyStrengthRequests, request)
		case 0x22: // PARTY SURPRISE
			if !instruction.Operands[0].WordSet || !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("PARTY SURPRISE at %d has non-address destination", pc)
			}
			request := PartySurpriseRequest{
				RangerDestination: instruction.Operands[0].Word,
				OtherDestination:  instruction.Operands[1].Word,
			}
			if partyContext != nil {
				request.RangerValue = 0
				if partyContext.hasRanger() {
					request.RangerValue = 1
				}
				request.Resolved = true
				memory[request.RangerDestination] = request.RangerValue
				memory[request.OtherDestination] = request.OtherValue
			}
			result.PartySurpriseRequests = append(result.PartySurpriseRequests, request)
		case 0x1E: // CHECKPARTY
			query := uint16(0)
			var err error
			if instruction.Operands[0].Code == 1 {
				query = instruction.Operands[0].Word
			} else {
				query, err = operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("CHECKPARTY query at %d: %w", pc, err)
				}
			}
			affectID, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("CHECKPARTY affect at %d: %w", pc, err)
			}
			request := CheckPartyRequest{Query: query, AffectID: affectID}
			for index := range request.Destinations {
				operand := instruction.Operands[index+2]
				if !operand.WordSet {
					return result, fmt.Errorf("CHECKPARTY destination %d at %d is not an address", index+1, pc)
				}
				request.Destinations[index] = operand.Word
			}
			if partyContext != nil {
				resolved, known := partyContext.checkParty(query, affectID)
				resolved.Destinations = request.Destinations
				request = resolved
				if request.Resolved {
					if request.AffectFound {
						memory[request.Destinations[0]] = 0
						memory[request.Destinations[1]] = 0
						memory[request.Destinations[2]] = 0
						memory[request.Destinations[3]] = 1
					} else if known {
						memory[request.Destinations[0]] = request.Minimum
						memory[request.Destinations[1]] = request.Maximum
						memory[request.Destinations[2]] = request.Average
						memory[request.Destinations[3]] = 0
					}
				}
			}
			result.CheckPartyRequests = append(result.CheckPartyRequests, request)
		case 0x39: // WHO
			prompt := ""
			if len(result.Text) > 0 {
				prompt = result.Text[len(result.Text)-1]
			}
			request := WhoRequest{Prompt: prompt}
			if whoSelectionCursor >= len(whoSelections) && pauseOnMissing {
				result.WhoRequests = append(result.WhoRequests, request)
				result.WaitingForWho = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if whoSelectionCursor < len(whoSelections) {
				request.Selected = whoSelections[whoSelectionCursor]
				request.SelectionProvided = true
				whoSelectionCursor++
				result.WhoSelectionsConsumed++
				if partyContext != nil && int(request.Selected) < len(partyContext.Members) {
					selectedPlayerIndex = int(request.Selected)
					selectedPlayerSet = true
					stringsMemory[0x7C00] = partyContext.Members[selectedPlayerIndex].Name
				}
			}
			result.WhoRequests = append(result.WhoRequests, request)
		case 0x3F: // FIND SPECIAL
			affectID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("FIND SPECIAL at %d: %w", pc, err)
			}
			request := FindSpecialRequest{AffectID: affectID, SelectedPlayerIndex: selectedPlayerIndex}
			if partyContext != nil && selectedPlayerSet && selectedPlayerIndex >= 0 && selectedPlayerIndex < len(partyContext.Members) {
				request.Resolved = true
				for _, activeAffect := range partyContext.Members[selectedPlayerIndex].Effects {
					if uint16(activeAffect) == affectID {
						request.Found = true
						break
					}
				}
				for index := range compare {
					compare[index] = false
				}
				compare[0] = request.Found
				compare[1] = !request.Found
			}
			result.FindSpecialRequests = append(result.FindSpecialRequests, request)
		case 0x33: // PRINT RETURN
			// This command changes the original text window/cursor state. Keep
			// its instruction boundary observable while leaving renderer layout
			// to the game UI adapter.
			result.PrintReturnCount++
		case 0x38: // PROGRAM
			// PROGRAM dispatches into an external engine routine. The reference
			// implementation ends the current VM pass for PROGRAM 0/3/8/9;
			// retain the ID and stop at that boundary until the corresponding
			// renderer/game-state routine is implemented.
			program, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("PROGRAM at %d: %w", pc, err)
			}
			programID := uint8(program)
			result.ProgramIDs = append(result.ProgramIDs, programID)
			if programID == 0 || programID == 3 || programID == 8 || programID == 9 {
				result.ProgramExit = true
				result.PC = next
				return result, nil
			}
		case 0x0B: // LOAD MONSTER
			spawn, err := DecodeMonsterSpawnFromMemory(instruction, memory)
			if err != nil {
				return result, fmt.Errorf("LOAD MONSTER at %d: %w", pc, err)
			}
			result.MonsterSpawns = append(result.MonsterSpawns, spawn)
		case 0x0C: // SETUP MONSTER
			setup, err := DecodeMonsterSetupFromMemory(instruction, memory)
			if err != nil {
				return result, fmt.Errorf("SETUP MONSTER at %d: %w", pc, err)
			}
			result.MonsterSetup = &setup
		case 0x36: // ADD NPC
			npcID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("ADD NPC at %d: %w", pc, err)
			}
			// Preserve the observed ID; the NPC table and party insertion side
			// effect belong to the game adapter.
			result.NPCIDs = append(result.NPCIDs, npcID)
		case 0x25, 0x26: // ON GOTO / ON GOSUB
			operands, headNext, err := ParseOperands(payload, pc, 2)
			if err != nil {
				return result, fmt.Errorf("ON branch at %d: %w", pc, err)
			}
			index, err := operandValue(operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("ON branch index at %d: %w", pc, err)
			}
			count, err := operandValue(operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("ON branch count at %d: %w", pc, err)
			}
			if count > 256 {
				return result, fmt.Errorf("ON branch at %d has unreasonable target count %d", pc, count)
			}
			// The original decrements the cursor once before loading the
			// variable target list, so its first skipped byte is headNext-1.
			targets, afterTargets, err := ParseOperands(payload, headNext-1, int(count))
			if err != nil {
				return result, fmt.Errorf("ON branch targets at %d: %w", pc, err)
			}
			if index >= count {
				pc = afterTargets
				result.PC = pc
				continue
			}
			target, ok := CodeTarget(targets[index], len(payload))
			if !ok {
				return result, fmt.Errorf("ON branch at %d has invalid target %d", pc, index)
			}
			if instruction.Command.Opcode == 0x26 {
				stack = append(stack, afterTargets)
			}
			pc = target
			continue
		case 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B: // IF comparison
			index := int(instruction.Command.Opcode - 0x16)
			if !compare[index] {
				skipped, err := decodeInstruction(payload, next)
				if err != nil {
					return result, fmt.Errorf("if at %d cannot skip next command: %w", pc, err)
				}
				next = skipped.Next
			}
		case 0x0E, 0x1C, 0x21, 0x27, 0x31, 0x37, 0x3B, 0x3C, 0x3D:
			// PICTURE, CLEARMONSTERS, LOAD FILES, TREASURE, SPRITE OFF and
			// CLEAR BOX have decoded arity but require the full renderer,
			// party/inventory or asset state. Consuming their operands and
			// continuing is a bounded prefix behavior, not a claim of effects.
			if instruction.Command.Opcode == 0x21 {
				for index, operand := range instruction.Operands {
					value, err := operandValue(operand, memory)
					if err != nil {
						return result, fmt.Errorf("load files at %d: %w", pc, err)
					}
					result.LoadFiles[index] = value
				}
				result.LoadFilesRequested = true
			}
			if instruction.Command.Opcode == 0x37 {
				for index, operand := range instruction.Operands {
					value, err := operandValue(operand, memory)
					if err != nil {
						return result, fmt.Errorf("load pieces at %d: %w", pc, err)
					}
					result.LoadPieces[index] = value
				}
				result.LoadPiecesRequested = true
			}
			if instruction.Command.Opcode == 0x0E {
				value, err := operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("picture at %d: %w", pc, err)
				}
				if value != 0xFF {
					result.PictureRequested = true
					result.PictureBlock = value
					result.BigPictureRequested = value >= 0x78
				}
			}
			if instruction.Command.Opcode == 0x1C {
				result.MonsterSetup = nil
				result.MonsterSpawns = nil
			}
			if instruction.Command.Opcode == 0x27 {
				request := TreasureRequest{}
				for index, operand := range instruction.Operands {
					value, err := operandValue(operand, memory)
					if err != nil {
						return result, fmt.Errorf("treasure at %d: %w", pc, err)
					}
					if index < len(request.Coins) {
						request.Coins[index] = value
					} else {
						request.ItemBlock = value
					}
				}
				result.TreasureRequests = append(result.TreasureRequests, request)
			}
			if instruction.Command.Opcode == 0x3B {
				spellID, err := operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("spell at %d: %w", pc, err)
				}
				slotAddress, err := operandAddress(instruction.Operands[1])
				if err != nil {
					return result, fmt.Errorf("spell slot address at %d: %w", pc, err)
				}
				characterAddress, err := operandAddress(instruction.Operands[2])
				if err != nil {
					return result, fmt.Errorf("spell character address at %d: %w", pc, err)
				}
				result.SpellSearches = append(result.SpellSearches, SpellSearch{
					SpellID: spellID, SpellSlotAddress: slotAddress, CharacterAddress: characterAddress,
				})
			}
			if instruction.Command.Opcode == 0x3C {
				address, err := operandAddress(instruction.Operands[0])
				if err != nil {
					return result, fmt.Errorf("protection address at %d: %w", pc, err)
				}
				result.ProtectionRequests = append(result.ProtectionRequests, address)
			}
		default:
			return result, fmt.Errorf("unsupported opcode 0x%02X at payload offset %d", instruction.Command.Opcode, pc)
		}
		pc = next
		result.PC = pc
	}
	return result, fmt.Errorf("runtime step limit %d reached at payload offset %d", maxSteps, pc)
}

func operandValue(operand Operand, memory map[uint16]uint16) (uint16, error) {
	switch operand.Code {
	case 0x00:
		return uint16(operand.Low), nil
	case 0x01, 0x03:
		if !operand.WordSet {
			return 0, fmt.Errorf("memory operand has no address")
		}
		return memory[operand.Word], nil
	case 0x02:
		if !operand.WordSet {
			return 0, fmt.Errorf("literal operand has no word")
		}
		return operand.Word, nil
	case 0x81:
		return 0, fmt.Errorf("string-memory operand cannot be used as a numeric value")
	default:
		return 0, fmt.Errorf("unsupported value operand code 0x%02X", operand.Code)
	}
}

func operandAddress(operand Operand) (uint16, error) {
	if !operand.WordSet || (operand.Code != 0x01 && operand.Code != 0x02 && operand.Code != 0x03 && operand.Code != 0x81) {
		return 0, fmt.Errorf("operand is not a word address")
	}
	return operand.Word, nil
}

func operandIsText(operand Operand) bool {
	return operand.Code == 0x80 || operand.Code == 0x81
}

func operandText(operand Operand, stringsMemory map[uint16]string) (string, error) {
	if operand.Code == 0x80 {
		return DecodePackedText(operand.Packed), nil
	}
	if operand.Code == 0x81 && operand.WordSet {
		return stringsMemory[operand.Word], nil
	}
	return "", fmt.Errorf("unsupported string operand code 0x%02X", operand.Code)
}
