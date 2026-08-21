package save

import (
	"encoding/binary"
	"testing"
)

func TestECLBankOffsetsMatchTheNamedAreaFields(t *testing.T) {
	// `internal/area` 的具名位移就是 ECL 位址（spec 1163）。這裡用同一條算式
	// 反算回去，任何一邊改了都會紅。
	for _, item := range []struct {
		name    string
		offset  int
		low     uint16
		address uint16
	}{
		{"Area1 InDungeon", 0x1CC, ECLBank0Low, 0x4BE6},
		{"Area1 LastX", 0x1E0, ECLBank0Low, 0x4BF0},
		{"Area1 LastY", 0x1E2, ECLBank0Low, 0x4BF1},
		{"Area1 LastECL", 0x1E4, ECLBank0Low, 0x4BF2},
		{"Area1 OutdoorSky", 0x1FA, ECLBank0Low, 0x4BFD},
		{"Area1 IndoorSky", 0x1FC, ECLBank0Low, 0x4BFE},
		{"Area1 CurrentCity", 0x342, ECLBank0Low, 0x4CA1},
		{"Area2 HeadBlock", 0x5C2, ECLBank1Low, 0x7EE1},
		{"Area2 GameArea", 0x624, ECLBank1Low, 0x7F12},
	} {
		if got := int(item.low) + item.offset/2; got != int(item.address) {
			t.Errorf("%s：位移 %#x 換算成 %#x，want %#x", item.name, item.offset, got, item.address)
		}
	}
}

func TestECLBankRoundTripKeepsUnmodelledBytes(t *testing.T) {
	original := make([]byte, (ECLBank0High-ECLBank0Low+1)*2)
	// 一格 remake 沒有碰過的原版資料。
	binary.LittleEndian.PutUint16(original[0x10:], 0xBEEF)
	memory := map[uint16]uint16{0x4BF2: 0x51, 0x4C06: 0, 0x9000: 1}

	record, err := EncodeECLBank(memory, ECLBank0Low, ECLBank0High, original)
	if err != nil {
		t.Fatal(err)
	}
	if got := binary.LittleEndian.Uint16(record[0x10:]); got != 0xBEEF {
		t.Fatalf("沒建模的位元組被蓋掉了：%#x", got)
	}
	if got := binary.LittleEndian.Uint16(record[(0x4BF2-ECLBank0Low)*2:]); got != 0x51 {
		t.Fatalf("4BF2 寫成 %#x，want 0x51", got)
	}

	decoded, err := DecodeECLBank(record, ECLBank0Low, ECLBank0High)
	if err != nil {
		t.Fatal(err)
	}
	if decoded[0x4BF2] != 0x51 {
		t.Fatalf("讀回來的 4BF2=%#x", decoded[0x4BF2])
	}
	// 值是 0 的格子不收，`MemoryValue` 的第二個回傳值才留得住意義。
	if _, present := decoded[0x4C06]; present {
		t.Fatal("0 值不該被收進來")
	}
	// 範圍外的位址不會寫進這一塊，也讀不回來。
	if _, present := decoded[0x9000]; present {
		t.Fatal("區 0 讀回了範圍外的位址")
	}
}
