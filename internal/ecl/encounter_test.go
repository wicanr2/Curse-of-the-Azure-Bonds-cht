package ecl

import "testing"

func TestDecodeMonsterSpawnLiterals(t *testing.T) {
	instruction, err := decodeInstruction([]byte{0x0B, 0x00, 0x59, 0x00, 0x04, 0x00, 0x20}, 0)
	if err != nil {
		t.Fatal(err)
	}
	spawn, err := DecodeMonsterSpawn(instruction)
	if err != nil {
		t.Fatal(err)
	}
	if spawn != (MonsterSpawn{MonsterID: 0x59, Count: 4, IconBlock: 0x20}) {
		t.Fatalf("spawn=%#v", spawn)
	}
}

func TestDecodeMonsterSetupLiterals(t *testing.T) {
	instruction, err := decodeInstruction([]byte{0x0C, 0x00, 0x04, 0x00, 0x02, 0x00, 0x04}, 0)
	if err != nil {
		t.Fatal(err)
	}
	setup, err := DecodeMonsterSetup(instruction)
	if err != nil {
		t.Fatal(err)
	}
	if setup != (MonsterSetup{SpriteID: 4, MaxDistance: 2, PictureID: 4}) {
		t.Fatalf("setup=%#v", setup)
	}
}

func TestDecodeMonsterSpawnRejectsMemoryOperand(t *testing.T) {
	instruction, err := decodeInstruction([]byte{0x0B, 0x01, 0x00, 0x90, 0x00, 0x04, 0x00, 0x20}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeMonsterSpawn(instruction); err == nil {
		t.Fatal("expected non-literal operand error")
	}
}

func TestDecodeMonsterSpawnResolvesMemoryOperands(t *testing.T) {
	instruction := Instruction{Command: Command{Opcode: 0x0B}, Operands: []Operand{
		{Code: 0x01, Word: 0xC04F, WordSet: true},
		{Code: 0x00, Low: 4},
		{Code: 0x01, Word: 0x7F79, WordSet: true},
	}}
	spawn, err := DecodeMonsterSpawnFromMemory(instruction, map[uint16]uint16{
		0xC04F: 0x59,
		0x7F79: 0x20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if spawn != (MonsterSpawn{MonsterID: 0x59, Count: 4, IconBlock: 0x20}) {
		t.Fatalf("spawn=%+v", spawn)
	}
}
