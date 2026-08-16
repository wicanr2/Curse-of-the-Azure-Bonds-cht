// Package ecl contains evidence-backed ECL parsing and a deliberately bounded
// command subset runner. It is not a complete recreation of the DOS engine.
package ecl

import "fmt"

type Operand struct {
	Code    byte
	Low     byte
	High    byte
	Word    uint16
	WordSet bool
	// Packed is present for code 0x80 operands. It contains exactly the
	// length-prefixed compressed bytes consumed by vm_LoadCmdSets.
	Packed []byte
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
		if operand.Code == 0x80 {
			length := int(operand.Low)
			if pos+1+length > len(payload) {
				return nil, pos, fmt.Errorf("operand %d packed string is truncated", i)
			}
			operand.Packed = append([]byte(nil), payload[pos+1:pos+1+length]...)
			pos += length
		} else if operand.Code == 1 || operand.Code == 2 || operand.Code == 3 || operand.Code == 0x81 {
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
// ECL dump table and the public CoAB reimplementation. It describes cursor
// movement, not command execution.
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
	0x1E: {0x1E, "CHECKPARTY", 6}, 0x1F: {0x1F, "UNKNOWN_1F", 2}, 0x20: {0x20, "NEWECL", 1}, 0x21: {0x21, "LOAD FILES", 3},
	0x22: {0x22, "PARTY SURPRISE", 2}, 0x23: {0x23, "SURPRISE", 4}, 0x24: {0x24, "COMBAT", 0},
	0x25: {0x25, "ON GOTO", 0}, 0x26: {0x26, "ON GOSUB", 0}, 0x27: {0x27, "TREASURE", 8},
	0x28: {0x28, "ROB", 3}, 0x29: {0x29, "ENCOUNTER MENU", 14}, 0x2A: {0x2A, "GETTABLE", 3},
	0x2B: {0x2B, "HORIZONTAL MENU", 0}, 0x2C: {0x2C, "PARLAY", 6}, 0x2D: {0x2D, "CALL", 1},
	0x2E: {0x2E, "DAMAGE", 5}, 0x2F: {0x2F, "AND", 3}, 0x30: {0x30, "OR", 3},
	0x31: {0x31, "SPRITE OFF", 0}, 0x32: {0x32, "FIND ITEM", 1}, 0x33: {0x33, "PRINT RETURN", 0},
	0x34: {0x34, "ECL CLOCK", 2}, 0x35: {0x35, "SAVE TABLE", 3}, 0x36: {0x36, "ADD NPC", 2},
	0x37: {0x37, "LOAD PIECES", 3}, 0x38: {0x38, "PROGRAM", 1}, 0x39: {0x39, "WHO", 1},
	0x3A: {0x3A, "DELAY", 0}, 0x3B: {0x3B, "SPELL", 3}, 0x3C: {0x3C, "PROTECTION", 1},
	0x3D: {0x3D, "CLEAR BOX", 0}, 0x3E: {0x3E, "DUMP", 0}, 0x3F: {0x3F, "FIND SPECIAL", 1},
	0x40: {0x40, "DESTROY ITEMS", 1},
}

// EntryPoints reads the five command-set words loaded by vm_init_ecl in the
// original engine. The returned values are code-segment addresses (normally
// 0x8000+), not decoded payload offsets. It intentionally supports only the
// word-valued command-set forms used by these five headers.
func EntryPoints(block []byte, count int) ([]uint16, int, error) {
	if len(block) < 2 {
		return nil, 0, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	if count < 0 {
		return nil, 0, fmt.Errorf("negative entry-point count")
	}
	payload := block[2:]
	points := make([]uint16, 0, count)
	pos := 0
	for i := 0; i < count; i++ {
		if pos+2 >= len(payload) {
			return points, pos, fmt.Errorf("entry point %d header is truncated at payload offset %d", i, pos)
		}
		code, low := payload[pos+1], payload[pos+2]
		if code != 1 && code != 2 && code != 3 {
			return points, pos, fmt.Errorf("entry point %d uses non-word code 0x%02X at payload offset %d", i, code, pos)
		}
		if pos+3 >= len(payload) {
			return points, pos, fmt.Errorf("entry point %d word is truncated", i)
		}
		points = append(points, uint16(payload[pos+3])<<8|uint16(low))
		pos += 4
	}
	return points, pos, nil
}

// Trace decodes cursor movement only. It stops at the first unknown command
// or malformed operand and returns the already decoded prefix plus an error.
func Trace(block []byte, limit int) ([]Instruction, error) {
	return TraceAt(block, 0, limit)
}

// TraceAt is Trace with an explicit decoded payload offset. This is needed
// for ECL blocks whose executable entry is stored in the initialization
// command sets rather than at payload offset zero.
func TraceAt(block []byte, start, limit int) ([]Instruction, error) {
	if len(block) < 2 {
		return nil, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	if start < 0 || start >= len(payload) {
		return nil, fmt.Errorf("trace start %d is outside payload", start)
	}
	if limit <= 0 || limit > len(payload) {
		limit = len(payload)
	}
	trace := make([]Instruction, 0)
	for len(trace) < limit {
		if len(trace) > 0 && trace[len(trace)-1].Next >= len(payload) {
			return trace, nil
		}
		offset := start
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

// ScanKnownInstructions linearly scans a decoded payload, resynchronizing at
// bytes that do not form a known instruction. It is an analysis aid for
// locating command records outside the currently reachable control-flow
// graph; callers must treat candidates as evidence to verify, not execution.
func ScanKnownInstructions(block []byte) ([]Instruction, error) {
	if len(block) < 2 {
		return nil, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	output := make([]Instruction, 0)
	for offset := 0; offset < len(payload); {
		instruction, err := decodeInstruction(payload, offset)
		if err != nil || instruction.Next <= offset || instruction.Next > len(payload) {
			offset++
			continue
		}
		output = append(output, instruction)
		offset = instruction.Next
	}
	return output, nil
}

// FindSaveDestinationCandidates checks every payload byte as a possible SAVE
// instruction and returns records whose second operand is the requested
// destination. Unlike ScanKnownInstructions it never skips over a candidate
// because unrelated bytes happened to decode as a longer instruction first.
// Results are still linear-scan evidence and require control-flow validation.
func FindSaveDestinationCandidates(block []byte, destination uint16) ([]Instruction, error) {
	if len(block) < 2 {
		return nil, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	var output []Instruction
	for offset := range payload {
		instruction, err := decodeInstruction(payload, offset)
		if err != nil || instruction.Command.Opcode != 0x09 ||
			len(instruction.Operands) != 2 ||
			!instruction.Operands[1].WordSet ||
			instruction.Operands[1].Word != destination {
			continue
		}
		output = append(output, instruction)
	}
	return output, nil
}

// VariableLengthCommands lists the opcodes whose record length is only known
// after an operand is read, and whose KnownCommands arity is therefore 0.
//
// ⚠ 這四個 opcode 的 `Instruction.Next` **指向自己的第一個運算元**，不是下一條
// 指令。任何拿 `Next` 往下走的程式都會把目標表／選項字串當成程式碼解，
// 而且**不會報錯**——只會安靜地少走一塊（spec 1110 §一）。走控制流一律用
// `RecordEnd`，不要用 `Next`。
var VariableLengthCommands = map[byte]string{
	0x15: "VERTICAL MENU",
	0x25: "ON GOTO",
	0x26: "ON GOSUB",
	0x2B: "HORIZONTAL MENU",
}

// RecordEnd returns the offset just past the instruction at offset — the real
// one, including the variable-length tail of the four commands above. For every
// other opcode it is Instruction.Next.
//
// 邊界檢查在這裡集中做：算出來的結尾必須**往前走**且落在 payload 內，
// 個數運算元必須是常數（`ON` 最多 256 個目的地、選單 1..64 個選項）。
// 任何一條不成立就回錯誤，讓呼叫端停下來，而不是拿一個看起來合理的數字繼續走。
func RecordEnd(block []byte, offset int) (int, error) {
	if len(block) < 2 {
		return 0, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	if offset < 0 || offset >= len(payload) {
		return 0, fmt.Errorf("record offset %d is outside payload", offset)
	}
	var end int
	switch opcode := payload[offset]; opcode {
	case 0x25, 0x26:
		_, after, err := branchTargets(payload, offset)
		if err != nil {
			return 0, err
		}
		end = after
	case 0x15, 0x2B:
		after, err := menuEnd(payload, offset)
		if err != nil {
			return 0, err
		}
		end = after
	default:
		instruction, err := decodeInstruction(payload, offset)
		if err != nil {
			return 0, err
		}
		end = instruction.Next
	}
	if end <= offset || end > len(payload) {
		return 0, fmt.Errorf("record at %d ends at %d, outside (%d, %d]", offset, end, offset, len(payload))
	}
	return end, nil
}

// BranchTargets decodes the variable-length 25h ON GOTO / 26h ON GOSUB record
// at a payload offset: the branch index, the target count and the target list,
// followed by the offset of the next instruction. The command table gives both
// opcodes arity 0 because their length is only known after the count operand is
// read, so Instruction.Next points one byte past the opcode — into the operand
// bytes. Any caller that walks control flow must use this instead of Next, or
// it decodes the target list as if it were code.
func BranchTargets(block []byte, offset int) (targets []int, after int, err error) {
	if len(block) < 2 {
		return nil, 0, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	return branchTargets(block[2:], offset)
}

func branchTargets(payload []byte, offset int) ([]int, int, error) {
	if offset < 0 || offset >= len(payload) {
		return nil, 0, fmt.Errorf("ON branch offset %d is outside payload", offset)
	}
	opcode := payload[offset]
	if opcode != 0x25 && opcode != 0x26 {
		return nil, 0, fmt.Errorf("opcode 0x%02X at %d is not an ON branch", opcode, offset)
	}
	head, headNext, err := ParseOperands(payload, offset, 2)
	if err != nil {
		return nil, 0, fmt.Errorf("ON branch at %d: %w", offset, err)
	}
	// 目標個數必須是常數，否則表的長度是執行期才知道的，靜態走訪無從跟起。
	// 實測 corpus 裡每一處都是 code 0x00 的立即值（`00 0E` ＝ 14）。
	count, err := literalOperand(head[1])
	if err != nil {
		return nil, 0, fmt.Errorf("ON branch at %d: %w", offset, err)
	}
	if count < 0 || count > 256 {
		return nil, 0, fmt.Errorf("ON branch at %d has unreasonable target count %d", offset, count)
	}
	// The original decrements the cursor once before loading the target list,
	// so the list starts at headNext (see the runtime's 0x25/0x26 handler).
	list, after, err := ParseOperands(payload, headNext-1, count)
	if err != nil {
		return nil, 0, fmt.Errorf("ON branch targets at %d: %w", offset, err)
	}
	targets := make([]int, 0, count)
	for _, operand := range list {
		if target, ok := CodeTarget(operand, len(payload)); ok {
			targets = append(targets, target)
		}
	}
	return targets, after, nil
}

// MenuEnd returns the offset just past a 15h VERTICAL MENU or 2Bh HORIZONTAL
// MENU record. Both carry a variable-length option-string list, so — like the
// ON branches above — their command-table arity is 0 and Instruction.Next lands
// inside the operand bytes. The layout matches the runtime's 0x15/0x2B cases:
// a two- (horizontal) or three-operand (vertical) header, then `count` option
// strings starting at headNext.
func MenuEnd(block []byte, offset int) (int, error) {
	if len(block) < 2 {
		return 0, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	return menuEnd(block[2:], offset)
}

func menuEnd(payload []byte, offset int) (int, error) {
	if offset < 0 || offset >= len(payload) {
		return 0, fmt.Errorf("menu offset %d is outside payload", offset)
	}
	opcode := payload[offset]
	headerArity, countIndex := 2, 1
	switch opcode {
	case 0x2B:
	case 0x15:
		headerArity, countIndex = 3, 2
	default:
		return 0, fmt.Errorf("opcode 0x%02X at %d is not a menu", opcode, offset)
	}
	header, headNext, err := ParseOperands(payload, offset, headerArity)
	if err != nil {
		return 0, fmt.Errorf("menu header at %d: %w", offset, err)
	}
	count, err := literalOperand(header[countIndex])
	if err != nil {
		return 0, fmt.Errorf("menu at %d: %w", offset, err)
	}
	if count == 0 || count > 64 {
		return 0, fmt.Errorf("menu at %d has invalid option count %d", offset, count)
	}
	_, stringsEnd, err := ParseOperands(payload, headNext-1, count)
	if err != nil {
		return 0, fmt.Errorf("menu strings at %d: %w", offset, err)
	}
	return stringsEnd, nil
}

// literalOperand reads a count that must be knowable without running the game:
// code 0x00 carries the value in the low byte, code 0x02 in the word. A memory
// reference (0x01/0x03) is only known at runtime, so static walkers must stop
// rather than guess a record length.
func literalOperand(operand Operand) (int, error) {
	switch {
	case operand.Code == 0x00:
		return int(operand.Low), nil
	case operand.Code == 0x02 && operand.WordSet:
		return int(operand.Word), nil
	default:
		return 0, fmt.Errorf("operand code 0x%02X is not a literal count", operand.Code)
	}
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
		guarded := false
		for offset >= 0 && offset < len(payload) && !seen[offset] && len(graph.Instructions) < limit {
			instruction, err := decodeInstruction(payload, offset)
			if err != nil {
				return graph, err
			}
			seen[offset] = true
			graph.Instructions = append(graph.Instructions, instruction)
			// `IF`（16h..1Bh）條件不成立時會**跳過下一條指令**：handler 讀六個
			// 比較旗標之一，為 0 就呼叫 overlay-07 entry#29，那一支照 opcode 的
			// 操作元個數把 PC 推過整條指令（spec 1106）。所以被 IF 守衛的那條
			// 指令之後的位址也是可達的——即使那條是 GOTO 而線性走訪會在此中斷。
			// 漏掉這條路等於漏掉所有 else 分支，而那是大部分的劇情文字。
			if guarded {
				queue = append(queue, instruction.Next)
			}
			guarded = instruction.Command.Opcode >= 0x16 && instruction.Command.Opcode <= 0x1B
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
			// `15h`／`2Bh` 選單的選項字串接在後面，長度同樣要自己算。
			if instruction.Command.Opcode == 0x15 || instruction.Command.Opcode == 0x2B {
				end, err := menuEnd(payload, instruction.Offset)
				if err != nil {
					break
				}
				offset = end
				continue
			}
			// `ON GOTO`／`ON GOSUB` 的目的地是靜態的字面位址，只有選哪一個才是動態
			// 的——所以每一個目的地都要進佇列。同時這條指令的長度必須自己算：
			// 相信 `Next` 會走進目標表的位元組裡，把資料當成程式解。
			if instruction.Command.Opcode == 0x25 || instruction.Command.Opcode == 0x26 {
				targets, after, err := branchTargets(payload, instruction.Offset)
				if err != nil {
					return graph, err
				}
				kind := "ON GOTO"
				if instruction.Command.Opcode == 0x26 {
					kind = "ON GOSUB"
				}
				for _, target := range targets {
					graph.Edges = append(graph.Edges, Edge{From: instruction.Offset, To: target, Kind: kind})
					queue = append(queue, target)
				}
				// index 超出 count 時原作直接落到表後面，`ON GOSUB` 回來也在那裡。
				offset = after
				continue
			}
			offset = instruction.Next
		}
	}
	return graph, nil
}
