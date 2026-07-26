package ecl

import "fmt"

// RunResult is the observable output of the bounded ECL subset runner.
// It deliberately exposes text and stop position, not DOS rendering state.
type RunResult struct {
	Text  []string
	PC    int
	Steps int
}

// RunSubset executes only commands whose semantics are represented here.
// Unsupported commands return an error at their exact payload offset. This is
// useful for proving an event prefix without silently treating the whole ECL
// program as implemented.
func RunSubset(block []byte, start, maxSteps int) (RunResult, error) {
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
	var compare [6]bool
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
		case 0x09: // SAVE
			if !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("save at %d has non-address destination", pc)
			}
			value, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("save at %d: %w", pc, err)
			}
			memory[instruction.Operands[1].Word] = value
		case 0x11, 0x12: // PRINT / PRINTCLEAR
			if len(instruction.Operands) != 1 {
				return result, fmt.Errorf("print at %d has unexpected arity", pc)
			}
			operand := instruction.Operands[0]
			if len(operand.Packed) > 0 {
				result.Text = append(result.Text, DecodePackedText(operand.Packed))
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
		case 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B: // IF comparison
			index := int(instruction.Command.Opcode - 0x16)
			if !compare[index] {
				skipped, err := decodeInstruction(payload, next)
				if err != nil {
					return result, fmt.Errorf("if at %d cannot skip next command: %w", pc, err)
				}
				next = skipped.Next
			}
		case 0x0E, 0x1C, 0x21, 0x31, 0x3D: // stateful in the full engine; safe prefix no-op
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
	default:
		return 0, fmt.Errorf("unsupported value operand code 0x%02X", operand.Code)
	}
}
