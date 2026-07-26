// Package ecl contains small, evidence-backed pieces of the ECL reader.
// It does not yet execute ECL commands.
package ecl

import "fmt"

type Operand struct {
	Code    byte
	Low     byte
	High    byte
	Word    uint16
	WordSet bool
}

// ParseOperands follows the operand framing used by the Gold Box ECL dump:
// code and low are read at offset+1/offset+2; codes 1, 2 and 3 consume one
// additional high byte. The skipped byte at offset and the final increment
// are part of the original VM's instruction cursor convention.
func ParseOperands(payload []byte, offset, count int) ([]Operand, int, error) {
	if offset < 0 || count < 0 || offset > len(payload) {
		return nil, offset, fmt.Errorf("invalid operand range")
	}
	operands := make([]Operand, 0, count)
	pos := offset
	for i := 0; i < count; i++ {
		if pos+2 >= len(payload) {
			return nil, pos, fmt.Errorf("operand %d is truncated", i)
		}
		operand := Operand{Code: payload[pos+1], Low: payload[pos+2]}
		pos += 2
		if operand.Code == 1 || operand.Code == 2 || operand.Code == 3 {
			pos++
			if pos >= len(payload) {
				return nil, pos, fmt.Errorf("operand %d high byte is truncated", i)
			}
			operand.High = payload[pos]
			operand.Word = uint16(operand.High)<<8 | uint16(operand.Low)
			operand.WordSet = true
		}
		operands = append(operands, operand)
	}
	return operands, pos + 1, nil
}

type Command struct {
	Opcode byte
	Name   string
	Arity  int
}

type Instruction struct {
	Offset   int
	Command  Command
	Operands []Operand
	Next     int
}

const CodeAddressBase = 0x8000

// KnownCommands contains only the command metadata recovered from the public
// ECL dump table. It describes cursor movement, not command execution.
var KnownCommands = map[byte]Command{
	0x00: {0x00, "EXIT", 0}, 0x01: {0x01, "GOTO", 1}, 0x02: {0x02, "GOSUB", 1},
	0x03: {0x03, "COMPARE", 2}, 0x04: {0x04, "ADD", 3}, 0x05: {0x05, "SUBTRACT", 3},
	0x06: {0x06, "DIVIDE", 3}, 0x07: {0x07, "MULTIPLY", 3}, 0x08: {0x08, "RANDOM", 2},
	0x09: {0x09, "SAVE", 2}, 0x0A: {0x0A, "LOAD CHARACTER", 1}, 0x0B: {0x0B, "LOAD MONSTER", 3},
	0x0C: {0x0C, "SETUP MONSTER", 3}, 0x0D: {0x0D, "APPROACH", 0}, 0x0E: {0x0E, "PICTURE", 1},
	0x0F: {0x0F, "INPUT NUMBER", 2}, 0x10: {0x10, "INPUT STRING", 2}, 0x11: {0x11, "PRINT", 1},
	0x12: {0x12, "PRINTCLEAR", 1}, 0x13: {0x13, "RETURN", 0}, 0x14: {0x14, "COMPARE AND", 4},
	0x15: {0x15, "VERTICAL MENU", 0}, 0x16: {0x16, "IF =", 0}, 0x17: {0x17, "IF <>", 0},
	0x18: {0x18, "IF <", 0}, 0x19: {0x19, "IF >", 0}, 0x1A: {0x1A, "IF <=", 0},
	0x1B: {0x1B, "IF >=", 0}, 0x1C: {0x1C, "CLEARMONSTERS", 0}, 0x1D: {0x1D, "PARTYSTRENGTH", 1},
	0x1E: {0x1E, "CHECKPARTY", 6}, 0x20: {0x20, "NEWECL", 1}, 0x21: {0x21, "LOAD FILES", 3},
	0x22: {0x22, "PARTY SURPRISE", 2}, 0x23: {0x23, "SURPRISE", 4}, 0x24: {0x24, "COMBAT", 0},
	0x25: {0x25, "ON GOTO", 0}, 0x26: {0x26, "ON GOSUB", 0}, 0x27: {0x27, "TREASURE", 8},
	0x28: {0x28, "ROB", 3}, 0x29: {0x29, "ENCOUNTER MENU", 14}, 0x2A: {0x2A, "GETTABLE", 3},
	0x2B: {0x2B, "HORIZONTAL MENU", 0}, 0x2C: {0x2C, "PARLAY", 6}, 0x2D: {0x2D, "CALL", 1},
	0x2E: {0x2E, "DAMAGE", 5}, 0x2F: {0x2F, "AND", 3}, 0x30: {0x30, "OR", 3},
	0x31: {0x31, "SPRITE OFF", 0}, 0x32: {0x32, "FIND ITEM", 1}, 0x33: {0x33, "PRINT RETURN", 0},
	0x34: {0x34, "ECL CLOCK", 2}, 0x35: {0x35, "SAVE TABLE", 3}, 0x36: {0x36, "ADD NPC", 2},
	0x37: {0x37, "LOAD PIECES", 3}, 0x38: {0x38, "PROGRAM", 1}, 0x39: {0x39, "WHO", 1},
	0x3A: {0x3A, "DELAY", 0}, 0x3B: {0x3B, "SPELL", 3}, 0x3C: {0x3C, "PROTECTION", 1},
	0x3D: {0x3D, "CLEAR BOX", 0},
}

// Trace decodes cursor movement only. It stops at the first unknown command
// or malformed operand and returns the already decoded prefix plus an error.
func Trace(block []byte, limit int) ([]Instruction, error) {
	if len(block) < 2 {
		return nil, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	if limit <= 0 || limit > len(payload) {
		limit = len(payload)
	}
	trace := make([]Instruction, 0)
	for len(trace) < limit {
		if len(trace) > 0 && trace[len(trace)-1].Next >= len(payload) {
			return trace, nil
		}
		offset := 0
		if len(trace) > 0 {
			offset = trace[len(trace)-1].Next
		}
		instruction, err := decodeInstruction(payload, offset)
		if err != nil {
			return trace, err
		}
		trace = append(trace, instruction)
	}
	return trace, nil
}

func decodeInstruction(payload []byte, offset int) (Instruction, error) {
	if offset < 0 || offset >= len(payload) {
		return Instruction{}, fmt.Errorf("instruction offset %d is outside payload", offset)
	}
	opcode := payload[offset]
	command, ok := KnownCommands[opcode]
	if !ok {
		return Instruction{}, fmt.Errorf("unknown opcode 0x%02X at payload offset %d", opcode, offset)
	}
	next := offset + 1
	var operands []Operand
	if command.Arity > 0 {
		var err error
		operands, next, err = ParseOperands(payload, offset, command.Arity)
		if err != nil {
			return Instruction{}, fmt.Errorf("opcode 0x%02X at %d: %w", opcode, offset, err)
		}
	}
	return Instruction{Offset: offset, Command: command, Operands: operands, Next: next}, nil
}

type Edge struct {
	From int
	To   int
	Kind string
}

type Graph struct {
	Instructions []Instruction
	Edges        []Edge
}

// CodeTarget converts a word operand in the ECL code segment to a decoded
// payload offset. Values outside the code segment are data pointers and are
// deliberately not treated as branch destinations.
func CodeTarget(operand Operand, payloadLength int) (int, bool) {
	if !operand.WordSet || int(operand.Word) < CodeAddressBase {
		return 0, false
	}
	offset := int(operand.Word) - CodeAddressBase
	return offset, offset >= 0 && offset < payloadLength
}

// TraceGraph follows only statically visible GOTO/GOSUB targets and sequential
// fallthrough. It does not evaluate IF conditions or execute side effects.
// This makes it suitable for discovering event entry points without silently
// inventing game state.
func TraceGraph(block []byte, starts []int, limit int) (Graph, error) {
	if len(block) < 2 {
		return Graph{}, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	if limit <= 0 {
		limit = len(payload)
	}
	if len(starts) == 0 {
		starts = []int{0}
	}
	queue := append([]int(nil), starts...)
	seen := make(map[int]bool)
	graph := Graph{}
	for len(queue) > 0 && len(graph.Instructions) < limit {
		offset := queue[0]
		queue = queue[1:]
		if seen[offset] {
			continue
		}
		for offset >= 0 && offset < len(payload) && !seen[offset] && len(graph.Instructions) < limit {
			instruction, err := decodeInstruction(payload, offset)
			if err != nil {
				return graph, err
			}
			seen[offset] = true
			graph.Instructions = append(graph.Instructions, instruction)
			if instruction.Command.Opcode == 0x01 || instruction.Command.Opcode == 0x02 {
				if len(instruction.Operands) == 1 {
					if target, ok := CodeTarget(instruction.Operands[0], len(payload)); ok {
						kind := "GOTO"
						if instruction.Command.Opcode == 0x02 {
							kind = "GOSUB"
						}
						graph.Edges = append(graph.Edges, Edge{From: instruction.Offset, To: target, Kind: kind})
						queue = append(queue, target)
					}
				}
				if instruction.Command.Opcode == 0x01 {
					break
				}
			}
			if instruction.Command.Opcode == 0x00 || instruction.Command.Opcode == 0x13 {
				break
			}
			offset = instruction.Next
		}
	}
	return graph, nil
}
