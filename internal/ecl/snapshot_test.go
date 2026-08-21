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

// 每一段自己的暫存（`4C00h`..`4C0Fh`，spec 1162）也要進存檔。人在 B 段時
// A 段那一份是「停在旁邊」的，只存目前這一段等於把別段的一次性旗標清掉——
// 讀檔之後走回去，演過的房間會再演一次。
func TestBlockSessionSnapshotKeepsParkedBlockScratch(t *testing.T) {
	blocks := map[uint8][]byte{0x10: {0, 0, 0x00}, 0x20: {0, 0, 0x00}}
	original, err := NewBlockSession(blocks, 0x10)
	if err != nil {
		t.Fatal(err)
	}
	original.SetMemoryValue(0x4C06, 7)
	original.SetMemoryValue(0x7F12, 1)
	if err := original.Switch(0x20); err != nil {
		t.Fatal(err)
	}
	if value, ok := original.MemoryValue(0x4C06); ok && value != 0 {
		t.Fatalf("block 0x20 應該從乾淨的暫存開始，卻讀到 %d", value)
	}
	original.SetMemoryValue(0x4C06, 3)
	snapshot, err := original.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	restored, err := NewBlockSession(blocks, 0x10)
	if err != nil {
		t.Fatal(err)
	}
	if err := restored.RestoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	// 眼前這一段：讀回來就是 B 段自己的值；共用區（`7F12h`）也還在。
	if value, ok := restored.MemoryValue(0x4C06); !ok || value != 3 {
		t.Fatalf("讀檔後目前這一段的 4C06=%d,%v，want 3", value, ok)
	}
	if value, ok := restored.MemoryValue(0x7F12); !ok || value != 1 {
		t.Fatalf("讀檔後共用區的 7F12=%d,%v，want 1", value, ok)
	}
	// 停在旁邊那一段：走回去要還在。
	if err := restored.Switch(0x10); err != nil {
		t.Fatal(err)
	}
	if value, ok := restored.MemoryValue(0x4C06); !ok || value != 7 {
		t.Fatalf("讀檔後走回 0x10 的 4C06=%d,%v，want 7", value, ok)
	}
	// 再換回去，B 段那一份也不能在來回之間掉。
	if err := restored.Switch(0x20); err != nil {
		t.Fatal(err)
	}
	if value, ok := restored.MemoryValue(0x4C06); !ok || value != 3 {
		t.Fatalf("換回 0x20 的 4C06=%d,%v，want 3", value, ok)
	}
}
