package ecl

import "fmt"

// RunResult is the observable output of the bounded ECL subset runner.
// It deliberately exposes text and stop position, not DOS rendering state.
type RunResult struct {
	Text               []string
	Menus              []Menu
	PC                 int
	Steps              int
	WaitingForMenu     bool
	NewECLBlockID      *uint8
	CombatRequested    bool
	MonsterSetup       *MonsterSetup
	MonsterSpawns      []MonsterSpawn
	ProgramIDs         []uint8
	ProgramExit        bool
	SelectionsConsumed int
}

type Menu struct {
	Location uint16
	Options  []string
	Selected uint16
	Vertical bool
	Prompt   string
}

// RunSubset executes only commands whose semantics are represented here.
// Unsupported commands return an error at their exact payload offset. This is
// useful for proving an event prefix without silently treating the whole ECL
// program as implemented.
func RunSubset(block []byte, start, maxSteps int) (RunResult, error) {
	return runSubset(block, start, maxSteps, nil, false)
}

// RunSubsetWithSelections is RunSubset with deterministic selections for
// successive HORIZONTAL MENU commands. A missing selection keeps the safe
// default index 0. This is the bridge between ECL menu semantics and a UI;
// it does not pretend to implement the original blocking DOS input routine.
func RunSubsetWithSelections(block []byte, start, maxSteps int, selections []uint16) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, false)
}

// RunSubsetInteractive pauses at the first menu whose selection is not
// supplied. This allows a UI to feed one choice per frame/event instead of
// silently choosing index zero for all later menus.
func RunSubsetInteractive(block []byte, start, maxSteps int, selections []uint16) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, true)
}

func runSubset(block []byte, start, maxSteps int, selections []uint16, pauseOnMissing bool) (RunResult, error) {
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
	var compare [6]bool
	selectionCursor := 0
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
		case 0x04, 0x05, 0x06, 0x07: // ADD / SUBTRACT / DIVIDE / MULTIPLY
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
			}
			memory[instruction.Operands[2].Word] = value
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
			return result, nil
		case 0x24: // COMBAT
			// The original engine transfers control to its combat loop here.
			// Expose that control transfer; do not silently fall through or
			// claim that combat rules have been recreated.
			result.CombatRequested = true
			result.PC = next
			return result, nil
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
			spawn, err := DecodeMonsterSpawn(instruction)
			if err != nil {
				return result, fmt.Errorf("LOAD MONSTER at %d: %w", pc, err)
			}
			result.MonsterSpawns = append(result.MonsterSpawns, spawn)
		case 0x0C: // SETUP MONSTER
			setup, err := DecodeMonsterSetup(instruction)
			if err != nil {
				return result, fmt.Errorf("SETUP MONSTER at %d: %w", pc, err)
			}
			result.MonsterSetup = &setup
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
		case 0x0E, 0x1C, 0x21, 0x27, 0x31, 0x3D:
			// PICTURE, CLEARMONSTERS, LOAD FILES, TREASURE, SPRITE OFF and
			// CLEAR BOX have decoded arity but require the full renderer,
			// party/inventory or asset state. Consuming their operands and
			// continuing is a bounded prefix behavior, not a claim of effects.
			if instruction.Command.Opcode == 0x1C {
				result.MonsterSetup = nil
				result.MonsterSpawns = nil
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
