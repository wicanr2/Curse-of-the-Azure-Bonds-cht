package ecl

import "fmt"

// MonsterSpawn is the data-bearing part of a LOAD MONSTER command. It keeps
// the ECL identifiers separate from combat stats, which come from MON*CHA
// records and must not be guessed from sprite IDs.
type MonsterSpawn struct {
	MonsterID uint8
	Count     uint8
	IconBlock uint8
	// PartyMask records copies reassigned to CombatTeam.Ours by a later
	// LOAD CHARACTER + SAVE to the selected combat-team field (0x7D0C).
	// Bit zero is the first copy produced by this LOAD MONSTER.
	PartyMask uint64
}

type MonsterSetup struct {
	SpriteID    uint8
	MaxDistance uint8
	PictureID   uint8
}

func DecodeMonsterSpawn(instruction Instruction) (MonsterSpawn, error) {
	return DecodeMonsterSpawnFromMemory(instruction, nil)
}

// DecodeMonsterSpawnFromMemory resolves numeric LOAD MONSTER operands using
// the VM memory produced by earlier SAVE/bitwise commands. A nil memory map
// keeps the compatibility API literal-only and therefore preserves its old
// strict behavior.
func DecodeMonsterSpawnFromMemory(instruction Instruction, memory map[uint16]uint16) (MonsterSpawn, error) {
	if instruction.Command.Opcode != 0x0B || len(instruction.Operands) != 3 {
		return MonsterSpawn{}, fmt.Errorf("instruction is not LOAD MONSTER")
	}
	values, err := numericOperands(instruction.Operands, memory)
	if err != nil {
		return MonsterSpawn{}, err
	}
	return MonsterSpawn{MonsterID: uint8(values[0]), Count: uint8(values[1]), IconBlock: uint8(values[2])}, nil
}

func DecodeMonsterSetup(instruction Instruction) (MonsterSetup, error) {
	return DecodeMonsterSetupFromMemory(instruction, nil)
}

// DecodeMonsterSetupFromMemory is the bounded-memory counterpart for real
// ECL entries whose setup fields are variables rather than literals.
func DecodeMonsterSetupFromMemory(instruction Instruction, memory map[uint16]uint16) (MonsterSetup, error) {
	if instruction.Command.Opcode != 0x0C || len(instruction.Operands) != 3 {
		return MonsterSetup{}, fmt.Errorf("instruction is not SETUP MONSTER")
	}
	values, err := numericOperands(instruction.Operands, memory)
	if err != nil {
		return MonsterSetup{}, err
	}
	return MonsterSetup{SpriteID: uint8(values[0]), MaxDistance: uint8(values[1]), PictureID: uint8(values[2])}, nil
}

func literalOperands(operands []Operand) ([]uint16, error) {
	return numericOperands(operands, nil)
}

func numericOperands(operands []Operand, memory map[uint16]uint16) ([]uint16, error) {
	values := make([]uint16, len(operands))
	for index, operand := range operands {
		var value uint16
		var err error
		if memory == nil {
			if operand.Code != 0x00 {
				return nil, fmt.Errorf("operand %d is not a literal: code 0x%02X", index, operand.Code)
			}
			value = uint16(operand.Low)
		} else {
			value, err = operandValue(operand, memory)
			if err != nil {
				return nil, fmt.Errorf("operand %d: %w", index, err)
			}
		}
		if value > 0xFF {
			return nil, fmt.Errorf("operand %d value 0x%04X does not fit byte descriptor", index, value)
		}
		values[index] = value
	}
	return values, nil
}
