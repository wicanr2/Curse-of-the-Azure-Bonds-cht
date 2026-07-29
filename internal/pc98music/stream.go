package pc98music

import (
	"errors"
	"fmt"
	"io"
)

type ChannelFamily string

const (
	ChannelFM     ChannelFamily = "fm"
	ChannelPSG    ChannelFamily = "psg"
	ChannelTiming ChannelFamily = "timing"
)

// StreamCommand preserves the structural boundary of one sequence command.
// Semantic names are intentionally limited to behavior proven in sub_10410.
type StreamCommand struct {
	Offset   int    `json:"offset"`
	Opcode   byte   `json:"opcode"`
	Name     string `json:"name"`
	Operands []byte `json:"operands,omitempty"`
}

type sequenceLoop struct {
	Target    int
	Remaining byte
}

// SequenceMachine executes only the control-flow and command-width semantics
// proven in sub_10410. PC/Start/End are data-segment offsets.
type SequenceMachine struct {
	Family    ChannelFamily
	PC        int
	Start     int
	End       int
	callStack []int
	loopStack []sequenceLoop
}

func channelFamily(channel int) (ChannelFamily, error) {
	switch {
	case channel >= 0 && channel < 3:
		return ChannelFM, nil
	case channel >= 3 && channel < 6:
		return ChannelPSG, nil
	case channel == 6:
		return ChannelTiming, nil
	default:
		return "", fmt.Errorf("channel %d is outside 0..6", channel)
	}
}

func commandShape(family ChannelFamily, opcode byte) (string, int, bool) {
	if opcode <= 0x60 {
		return "note", 1, true
	}
	if opcode == 0x80 {
		return "rest", 1, true
	}
	// The timing channel's sub_10410 branch only consumes operands for 85/8A.
	// Every other byte above 60h (except 80h) is skipped as one byte; it does
	// not share the FM/PSG control-flow grammar.
	if family == ChannelTiming {
		switch opcode {
		case 0x85:
			return "parameter_85", 1, true
		case 0x8A:
			return "parameter_8a", 1, true
		default:
			return "ignored", 0, true
		}
	}
	switch opcode {
	case 0x85:
		return "parameter_85", 1, true
	case 0x8A:
		return "parameter_8a", 1, true
	case 0x90:
		if family == ChannelFM {
			return "tempo_step", 0, true
		}
	case 0x91:
		if family == ChannelPSG {
			return "mode_91", 0, true
		}
	case 0x92:
		if family == ChannelPSG {
			return "mode_92", 0, true
		}
	case 0xA0:
		return "jump", 2, true
	case 0xA1:
		return "call", 2, true
	case 0xA2:
		return "return", 0, true
	case 0xA3:
		return "loop_begin", 1, true
	case 0xA4:
		return "loop_end", 0, true
	case 0xB0:
		return "extension_b0", 2, true
	}
	return "ignored", 0, false
}

func NewSequenceMachine(channel, start, end int) (*SequenceMachine, error) {
	family, err := channelFamily(channel)
	if err != nil {
		return nil, err
	}
	if start < 0 || end <= start {
		return nil, fmt.Errorf("invalid sequence range 0x%X..0x%X", start, end)
	}
	return &SequenceMachine{Family: family, PC: start, Start: start, End: end}, nil
}

func (machine *SequenceMachine) target(word []byte) int {
	return int(word[0]) | int(word[1])<<8
}

func (machine *SequenceMachine) checkTarget(target int, name string) error {
	if target < machine.Start || target >= machine.End {
		return fmt.Errorf(
			"%s target 0x%X is outside sequence 0x%X..0x%X",
			name, target, machine.Start, machine.End,
		)
	}
	return nil
}

// NextTimed executes control/parameter commands until one note/rest command
// establishes a duration. The returned slice includes that timed command.
func (machine *SequenceMachine) NextTimed(data []byte, maxCommands int) ([]StreamCommand, error) {
	if maxCommands <= 0 {
		return nil, fmt.Errorf("maxCommands must be positive")
	}
	commands := make([]StreamCommand, 0, 8)
	for len(commands) < maxCommands {
		if machine.PC == machine.End {
			return commands, io.EOF
		}
		if machine.PC < machine.Start || machine.PC >= machine.End || machine.PC >= len(data) {
			return nil, fmt.Errorf(
				"sequence PC 0x%X is outside 0x%X..0x%X",
				machine.PC, machine.Start, machine.End,
			)
		}
		offset := machine.PC
		opcode := data[offset]
		name, operandCount, known := commandShape(machine.Family, opcode)
		if !known {
			return nil, fmt.Errorf("unknown opcode 0x%02X at data offset 0x%X", opcode, offset)
		}
		end := offset + 1 + operandCount
		if end > machine.End || end > len(data) {
			return nil, fmt.Errorf("%s at data offset 0x%X crosses sequence end", name, offset)
		}
		operands := append([]byte(nil), data[offset+1:end]...)
		command := StreamCommand{Offset: offset, Opcode: opcode, Name: name, Operands: operands}
		commands = append(commands, command)
		machine.PC = end

		switch name {
		case "jump":
			target := machine.target(operands)
			if err := machine.checkTarget(target, "jump"); err != nil {
				return nil, err
			}
			machine.PC = target
		case "call":
			target := machine.target(operands)
			if err := machine.checkTarget(target, "call"); err != nil {
				return nil, err
			}
			if len(machine.callStack) < 16 {
				machine.callStack = append(machine.callStack, machine.PC)
				machine.PC = target
			}
		case "return":
			if len(machine.callStack) != 0 {
				last := len(machine.callStack) - 1
				machine.PC = machine.callStack[last]
				machine.callStack = machine.callStack[:last]
			}
		case "loop_begin":
			if len(machine.loopStack) < 16 {
				machine.loopStack = append(machine.loopStack, sequenceLoop{
					Target: machine.PC, Remaining: operands[0],
				})
			}
		case "loop_end":
			if len(machine.loopStack) != 0 {
				last := len(machine.loopStack) - 1
				machine.loopStack[last].Remaining--
				if machine.loopStack[last].Remaining != 0 {
					machine.PC = machine.loopStack[last].Target
				} else {
					machine.loopStack = machine.loopStack[:last]
				}
			}
		}
		if name == "note" || name == "rest" {
			return commands, nil
		}
	}
	return nil, fmt.Errorf("sequence exceeded %d commands without a timed event", maxCommands)
}

func IsSequenceEnd(err error) bool {
	return errors.Is(err, io.EOF)
}

// DecodeStreamStructure walks the stored bytes linearly. It validates command
// widths but does not execute jump/call/loop control flow; SequenceMachine is
// the independently verified runtime layer.
func DecodeStreamStructure(channel int, stream []byte) ([]StreamCommand, error) {
	family, err := channelFamily(channel)
	if err != nil {
		return nil, err
	}
	commands := make([]StreamCommand, 0, len(stream)/2)
	for offset := 0; offset < len(stream); {
		opcode := stream[offset]
		name, operandCount, known := commandShape(family, opcode)
		if !known {
			return nil, fmt.Errorf(
				"channel %d unknown opcode 0x%02X at stream offset 0x%X",
				channel, opcode, offset,
			)
		}
		end := offset + 1 + operandCount
		if end > len(stream) {
			return nil, fmt.Errorf(
				"channel %d %s opcode 0x%02X at stream offset 0x%X needs %d operand bytes",
				channel, name, opcode, offset, operandCount,
			)
		}
		commands = append(commands, StreamCommand{
			Offset:   offset,
			Opcode:   opcode,
			Name:     name,
			Operands: append([]byte(nil), stream[offset+1:end]...),
		})
		offset = end
	}
	return commands, nil
}
