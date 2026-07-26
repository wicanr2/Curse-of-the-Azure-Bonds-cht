package ecl

import "fmt"

// MonsterSpawn is the data-bearing part of a LOAD MONSTER command. It keeps
// the ECL identifiers separate from combat stats, which come from MON*CHA
// records and must not be guessed from sprite IDs.
type MonsterSpawn struct {
	MonsterID uint8
	Count     uint8
	IconBlock uint8
}

type MonsterSetup struct {
	SpriteID    uint8
	MaxDistance uint8
	PictureID   uint8
}

func DecodeMonsterSpawn(instruction Instruction) (MonsterSpawn, error) {
	if instruction.Command.Opcode != 0x0B || len(instruction.Operands) != 3 {
		return MonsterSpawn{}, fmt.Errorf("instruction is not LOAD MONSTER")
	}
	values, err := literalOperands(instruction.Operands)
	if err != nil {
		return MonsterSpawn{}, err
	}
	return MonsterSpawn{MonsterID: uint8(values[0]), Count: uint8(values[1]), IconBlock: uint8(values[2])}, nil
}

func DecodeMonsterSetup(instruction Instruction) (MonsterSetup, error) {
	if instruction.Command.Opcode != 0x0C || len(instruction.Operands) != 3 {
		return MonsterSetup{}, fmt.Errorf("instruction is not SETUP MONSTER")
	}
	values, err := literalOperands(instruction.Operands)
	if err != nil {
		return MonsterSetup{}, err
	}
	return MonsterSetup{SpriteID: uint8(values[0]), MaxDistance: uint8(values[1]), PictureID: uint8(values[2])}, nil
}

func literalOperands(operands []Operand) ([]uint16, error) {
	values := make([]uint16, len(operands))
	for index, operand := range operands {
		if operand.Code != 0x00 {
			return nil, fmt.Errorf("operand %d is not a literal: code 0x%02X", index, operand.Code)
		}
		values[index] = uint16(operand.Low)
	}
	return values, nil
}
