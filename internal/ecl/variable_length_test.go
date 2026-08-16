package ecl

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
)

// 四個變長指令的守衛（spec 1110 §一）。
//
// 這一組測試存在的理由是：拿 `Instruction.Next` 去走這四個 opcode **不會報錯**，
// 只會安靜地把目標表／選項字串當成程式碼解，讓覆蓋率少一塊而沒有任何跡象。
// 所以要有一條測試把「陷阱確實存在」與「`RecordEnd` 確實避開它」同時釘住——
// 只釘後者的話，哪天 `KnownCommands` 給了非零 arity，測試會照樣綠，
// 而陷阱已經換了形狀。
func TestVariableLengthCommandsAreNotWalkableByNext(t *testing.T) {
	for opcode, name := range VariableLengthCommands {
		command, ok := KnownCommands[opcode]
		if !ok {
			t.Fatalf("opcode 0x%02X (%s) is not in KnownCommands", opcode, name)
		}
		if command.Arity != 0 {
			t.Fatalf("opcode 0x%02X (%s) arity=%d，變長指令的 arity 必須是 0；"+
				"若原作真的改成定長，RecordEnd 的特例要一起拿掉", opcode, name, command.Arity)
		}
		if command.Name != name {
			t.Fatalf("opcode 0x%02X name=%q, VariableLengthCommands 說是 %q",
				opcode, command.Name, name)
		}
	}
}

// 對真實 corpus 逐筆檢查：RecordEnd 一定往前走、落在 payload 內，而且對這四個
// opcode **一定大於** Instruction.Next（Next 指向第一個運算元）。
func TestRecordEndCoversEveryVariableLengthRecordInTheCorpus(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()

	seen := map[byte]int{}
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
		blocks, err := dax.Parse(realZipMember(t, archive, member))
		if err != nil {
			t.Fatalf("%s: %v", member, err)
		}
		for _, block := range blocks {
			points, _, err := EntryPoints(block.Data, 5)
			if err != nil {
				t.Fatalf("%s block 0x%02X entry points: %v", member, block.Entry.ID, err)
			}
			starts := make([]int, 0, len(points))
			for _, point := range points {
				starts = append(starts, int(point)-CodeAddressBase)
			}
			graph, err := TraceGraph(block.Data, starts, len(block.Data)*8)
			if err != nil {
				t.Fatalf("%s block 0x%02X graph: %v", member, block.Entry.ID, err)
			}
			for _, instruction := range graph.Instructions {
				end, err := RecordEnd(block.Data, instruction.Offset)
				if err != nil {
					t.Fatalf("%s block 0x%02X +%04Xh RecordEnd: %v",
						member, block.Entry.ID, instruction.Offset, err)
				}
				if _, variable := VariableLengthCommands[instruction.Command.Opcode]; !variable {
					if end != instruction.Next {
						t.Fatalf("%s block 0x%02X +%04Xh opcode 0x%02X RecordEnd=%d, Next=%d，"+
							"定長指令兩者必須相同",
							member, block.Entry.ID, instruction.Offset,
							instruction.Command.Opcode, end, instruction.Next)
					}
					continue
				}
				seen[instruction.Command.Opcode]++
				if instruction.Next != instruction.Offset+1 {
					t.Fatalf("%s block 0x%02X +%04Xh opcode 0x%02X Next=%d，"+
						"arity 0 的指令 Next 應為 offset+1",
						member, block.Entry.ID, instruction.Offset,
						instruction.Command.Opcode, instruction.Next)
				}
				if end <= instruction.Next {
					t.Fatalf("%s block 0x%02X +%04Xh opcode 0x%02X RecordEnd=%d <= Next=%d，"+
						"變長指令的真正結尾必須在運算元之後",
						member, block.Entry.ID, instruction.Offset,
						instruction.Command.Opcode, end, instruction.Next)
				}
			}
		}
	}
	// corpus 裡四個 opcode 都要出現過，否則這條測試沒有真的驗到東西。
	for opcode, name := range VariableLengthCommands {
		if seen[opcode] == 0 {
			t.Fatalf("corpus 裡沒有出現 opcode 0x%02X (%s)，這條測試等於沒驗", opcode, name)
		}
	}
	t.Logf("變長指令出現次數：%v", seen)
}

// 上限本身也要驗：個數運算元若是記憶體參照（執行期才知道），或算出來的結尾
// 越界，RecordEnd 必須回錯誤而不是回一個看起來合理的數字。
func TestRecordEndRejectsUnknowableAndOutOfRangeLengths(t *testing.T) {
	// `25h ON GOTO` 的個數運算元用 code 0x01（記憶體參照）⇒ 靜態無從得知。
	dynamicCount := []byte{0x00, 0x00, 0x25, 0x01, 0x7A, 0x7F, 0x01, 0x40, 0x4C, 0x00}
	if _, err := RecordEnd(dynamicCount, 0); err == nil {
		t.Fatal("個數是記憶體參照時 RecordEnd 應該回錯誤")
	}
	// 個數是常數但目標表被截斷 ⇒ 結尾會越過 payload。
	truncated := []byte{0x00, 0x00, 0x25, 0x00, 0x01, 0x00, 0x08, 0x01}
	if _, err := RecordEnd(truncated, 0); err == nil {
		t.Fatal("目標表截斷時 RecordEnd 應該回錯誤")
	}
	// `2Bh HORIZONTAL MENU` 的選項個數 0 不合法（原作 handler 也擋）。
	zeroOptions := []byte{0x00, 0x00, 0x2B, 0x01, 0x79, 0x7F, 0x00, 0x00, 0x80, 0x00}
	if _, err := RecordEnd(zeroOptions, 0); err == nil {
		t.Fatal("選單選項個數 0 時 RecordEnd 應該回錯誤")
	}
}

// 副作用還原狀態的表必須蓋住 corpus 裡每一個可達 opcode（`RE-04` 的分母）。
// 少一列就會讓 `cmd/ecl-effect-coverage` 的統計默默漏掉那一類指令。
func TestOpcodeEffectsCoversEveryReachableOpcode(t *testing.T) {
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()

	reachable := map[byte]int{}
	for _, member := range []string{"ECL1.DAX", "ECL2.DAX", "ECL3.DAX", "ECL4.DAX", "ECL5.DAX", "ECL6.DAX"} {
		blocks, err := dax.Parse(realZipMember(t, archive, member))
		if err != nil {
			t.Fatalf("%s: %v", member, err)
		}
		for _, block := range blocks {
			points, _, err := EntryPoints(block.Data, 5)
			if err != nil {
				t.Fatalf("%s block 0x%02X: %v", member, block.Entry.ID, err)
			}
			starts := make([]int, 0, len(points))
			for _, point := range points {
				starts = append(starts, int(point)-CodeAddressBase)
			}
			graph, err := TraceGraph(block.Data, starts, len(block.Data)*8)
			if err != nil {
				t.Fatalf("%s block 0x%02X: %v", member, block.Entry.ID, err)
			}
			for _, instruction := range graph.Instructions {
				reachable[instruction.Command.Opcode]++
			}
		}
	}
	for opcode := range reachable {
		effect, ok := OpcodeEffects[opcode]
		if !ok {
			t.Fatalf("opcode 0x%02X 可達卻沒有登記在 OpcodeEffects", opcode)
		}
		switch effect.Status {
		case EffectDone, EffectPartial, EffectConsumed:
		default:
			t.Fatalf("opcode 0x%02X 的狀態 %q 不是三種之一", opcode, effect.Status)
		}
		if effect.Note == "" {
			t.Fatalf("opcode 0x%02X 沒寫 Note；狀態不附理由等於沒有判定", opcode)
		}
	}
	// 表裡不該有 corpus 沒出現的 opcode 以外的東西——命令表以外的鍵是打錯字。
	for opcode := range OpcodeEffects {
		if _, ok := KnownCommands[opcode]; !ok {
			t.Fatalf("OpcodeEffects 有 0x%02X，但 KnownCommands 沒有", opcode)
		}
	}
}
