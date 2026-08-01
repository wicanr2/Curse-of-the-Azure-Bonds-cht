package ecl

import (
	"reflect"
	"testing"
)

func TestBlockSessionSnapshotRestoresMemoryContinuationAndRandomStream(t *testing.T) {
	// RANDOM 0..0x7FFE -> [7000], SAVE 1234 -> [7001], PROGRAM 9, then
	// another RANDOM -> [7002], EXIT. The external PROGRAM boundary separates
	// the two draws; the second one must continue rather than restart.
	payload := []byte{
		0x08, 0x02, 0xFF, 0x7F, 0x01, 0x00, 0x70,
		0x09, 0x02, 0x34, 0x12, 0x01, 0x01, 0x70,
		0x38, 0x00, 0x09,
		0x08, 0x02, 0xFF, 0x7F, 0x01, 0x02, 0x70,
		0x00,
	}
	block := append([]byte{0, 0}, payload...)
	original, err := NewBlockSession(map[uint8][]byte{1: block}, 1)
	if err != nil {
		t.Fatal(err)
	}
	first, err := original.runFromSeed(0, 100, nil, 77)
	if err != nil {
		t.Fatal(err)
	}
	if !first.ProgramExit || !reflect.DeepEqual(first.ProgramIDs, []uint8{9}) {
		t.Fatalf("first boundary=%+v", first)
	}
	snapshot, err := original.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Memory[0x7001] != 0x1234 || snapshot.Random == nil || snapshot.Random.Draws == 0 {
		t.Fatalf("snapshot=%+v", snapshot)
	}

	want, err := original.runFromSeed(first.PC, 100, nil, 77)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := NewBlockSession(map[uint8][]byte{1: block}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := restored.RestoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	got, err := restored.runFromSeed(first.PC, 100, nil, 77)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.RandomValues, want.RandomValues) {
		t.Fatalf("restored random=%v, want %v", got.RandomValues, want.RandomValues)
	}
	if value, ok := restored.MemoryValue(0x7001); !ok || value != 0x1234 {
		t.Fatalf("restored memory[7001]=%04x,%v", value, ok)
	}
}

func TestBlockSessionSnapshotStoresOnlyCodeDifferences(t *testing.T) {
	block := append([]byte{0, 0}, make([]byte, 32)...)
	session, err := NewBlockSession(map[uint8][]byte{1: block}, 1)
	if err != nil {
		t.Fatal(err)
	}
	session.SetMemoryValue(CodeAddressBase+4, 0xAA)
	snapshot, err := session.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.CodeMemory) != 1 || snapshot.CodeMemory[CodeAddressBase+4] != 0xAA {
		t.Fatalf("code changes=%v", snapshot.CodeMemory)
	}
	if len(snapshot.Memory) != 0 {
		t.Fatalf("regular memory leaked code bytes: %v", snapshot.Memory)
	}
}
