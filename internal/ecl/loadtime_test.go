package ecl

import (
	"archive/zip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
)

// eclBlock 取原版 image 裡某一段的位元組碼。image 不在就跳過：這幾條測的是
// 原作資料，不是「有沒有把 image 放進 CI」。
func eclBlock(t *testing.T, member string, id uint8) []byte {
	t.Helper()
	const image = "../../curseoftheazurebonds.zip"
	if _, err := os.Stat(image); err != nil {
		t.Skipf("讀不到遊戲 image：%v", err)
	}
	archive, err := zip.OpenReader(image)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	var payload []byte
	for _, file := range archive.File {
		if !strings.EqualFold(file.Name, member) {
			continue
		}
		handle, openErr := file.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		payload, _ = io.ReadAll(handle)
		handle.Close()
	}
	if payload == nil {
		t.Fatalf("image 裡沒有 %s", member)
	}
	blocks, err := dax.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	for _, block := range blocks {
		if block.Entry.ID == id {
			return block.Data
		}
	}
	t.Fatalf("%s 裡沒有段 0x%02X", member, id)
	return nil
}

// skyWrites 把走訪結果收成「哪一格被寫成什麼」。
func skyWrites(t *testing.T, instructions []Instruction, cell uint16) []uint16 {
	t.Helper()
	values := make([]uint16, 0, 2)
	for _, instruction := range instructions {
		if instruction.Command.Opcode != 0x09 || len(instruction.Operands) != 2 {
			continue
		}
		if !instruction.Operands[1].WordSet || instruction.Operands[1].Word != cell {
			continue
		}
		if value, ok := ConstantOperandValue(instruction.Operands[0]); ok {
			values = append(values, value)
		}
	}
	return values
}

// ★ `ECL2/0x04` 的進入碼在 `LOAD FILES`／`LOAD PIECES` 之後 `GOSUB` 一支子程式，
// 那支拿**當下站的那一格**的地形碼（`C04Fh` ＝ 牆頂）決定天花板顏色：
//
//	16F4  COMPARE C04F 0x95
//	16FA  IF =   → 16FB  SAVE 0A 4BFE   { 紅 }
//	1701  IF <>  → 1702  SAVE 09 4BFE   { 白 }
//
// 靜態走訪（不給記憶體）判不出那個 `IF`，兩條都跳過 ⇒ **一格都收不到**；
// 那正是 remake 先前把火刀據點的紅天花板畫成白色的原因（8 張畫面差 13,748 格，
// 而每一格都是同一個代換 EGA 4 → 15，看起來就是一張正常的畫面）。
func TestLoadTimeWalkNeedsMemoryForTheCeilingSubroutine(t *testing.T) {
	data := eclBlock(t, "ECL2.DAX", 0x04)

	static, err := ReachableOnLoad(data, 0x04)
	if err != nil {
		t.Fatal(err)
	}
	if got := skyWrites(t, static, 0x4BFE); len(got) != 0 {
		t.Fatalf("不給記憶體時不該判得出來，卻收到 %v", got)
	}

	for _, item := range []struct {
		name    string
		terrain uint16
		want    uint16
	}{
		{name: "地形 95h 是紅天花板", terrain: 0x95, want: 0x0A},
		{name: "其餘是白天花板", terrain: 0x80, want: 0x09},
	} {
		t.Run(item.name, func(t *testing.T) {
			memory := func(address uint16) (uint16, bool) {
				if address == 0xC04F {
					return item.terrain, true
				}
				return 0, false
			}
			reached, walkErr := ReachableOnLoadWithMemory(data, 0x04, memory)
			if walkErr != nil {
				t.Fatal(walkErr)
			}
			got := skyWrites(t, reached, 0x4BFE)
			if len(got) != 1 || got[0] != item.want {
				t.Fatalf("室內天空寫入 ＝ %v，want [%d]", got, item.want)
			}
		})
	}
}

// ⚠ 跟進 `GOSUB` 是上面那條成立的前提：那支子程式在 `0x16F4`，而進入碼是
// `0x0022 GOSUB 0x96F4`。不跟的話走訪在 `LOAD PIECES` 之後就停了。
func TestLoadTimeWalkFollowsGosubFromTheEntryPath(t *testing.T) {
	data := eclBlock(t, "ECL2.DAX", 0x04)
	memory := func(address uint16) (uint16, bool) {
		if address == 0xC04F {
			return 0x95, true
		}
		return 0, false
	}
	reached, err := ReachableOnLoadWithMemory(data, 0x04, memory)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, instruction := range reached {
		if instruction.Offset == 0x16FB {
			found = true
		}
	}
	if !found {
		t.Fatal("走訪沒有跟進 `GOSUB`：`0x16FB` 不在結果裡")
	}
}
