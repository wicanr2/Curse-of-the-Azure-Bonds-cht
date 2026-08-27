package game

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/traditionalchinese"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// 原版存檔匯入與 remake 自己的槽用同一個版面，差別只有名字的編碼。
// 這一組測試釘住「兩條路都對」，以及「走錯路一定看得出來」。
//
// ★ 為什麼英文原版看不出差別。 `origtext.Decode` 對純 ASCII 直接回傳原位元組，
// 所以英文資料兩條路結果一樣——中文版才會炸。用英文樣本測這件事等於沒測。
func writeSAVGAMSlot(t *testing.T, directory string, key byte, name, itemName []byte) {
	t.Helper()
	areaState := area.State{GameArea: 3, Current3DMapBlockID: 0x21, InDungeon: true}
	area1, err := area.EncodeArea1(areaState, nil)
	if err != nil {
		t.Fatal(err)
	}
	area2, err := area.EncodeArea2(areaState, nil)
	if err != nil {
		t.Fatal(err)
	}
	base := "CHRDAT" + string(key) + "1"
	prefix, err := partySave.EncodeSAVGAM(partySave.SAVGAMContainer{
		GameArea: 3, Area1: area1, Area2: area2,
		Runtime:    make([]byte, partySave.SAVGAMRuntimeStateSize),
		ECL:        make([]byte, partySave.SAVGAMECLMemorySize),
		PartyCount: 1, CharacterRefs: [8][]byte{[]byte(base)},
	})
	if err != nil {
		t.Fatal(err)
	}
	lower := key + ('a' - 'A')
	if err := os.WriteFile(filepath.Join(directory, "savgam"+string(lower)+".dat"), prefix, 0o600); err != nil {
		t.Fatal(err)
	}
	record := make([]byte, party.DOSPlayerRecordSize)
	record[0] = byte(len(name))
	copy(record[1:], name)
	record[0x10], record[0x11], record[0x12] = 16, 16, 10
	record[0x14], record[0x16], record[0x18], record[0x1A] = 10, 12, 14, 10
	record[0x74], record[0x75] = 7, 2
	record[0x78], record[0x1A4], record[0x10B] = 10, 8, 1
	if err := os.WriteFile(filepath.Join(directory, base+".sav"), record, 0o600); err != nil {
		t.Fatal(err)
	}
	item := make([]byte, monster.ItemRecordSize)
	copy(item, itemName)
	item[0x39] = 1
	if err := os.WriteFile(filepath.Join(directory, base+".swg"), item, 0o600); err != nil {
		t.Fatal(err)
	}
}

func big5(t *testing.T, text string) []byte {
	t.Helper()
	encoded, err := traditionalchinese.Big5.NewEncoder().Bytes([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

// 原版（Big5）的中文名字與物品名，要走 LoadOriginalSAVGAMSlot 才讀得對。
func TestOriginalSAVGAMImportDecodesBig5NamesAndItems(t *testing.T) {
	directory := t.TempDir()
	writeSAVGAMSlot(t, directory, 'C', big5(t, "艾拉"), big5(t, "長劍"))

	state := NewState(testCatalog())
	if err := state.LoadOriginalSAVGAMSlot(directory, 'C'); err != nil {
		t.Fatal(err)
	}
	if got := state.partyRoster[0].Name; got != "艾拉" {
		t.Fatalf("角色名讀成 %q，want 艾拉", got)
	}
	if len(state.partyRoster[0].Equipment) != 1 {
		t.Fatalf("物品沒讀進來：%#v", state.partyRoster[0].Equipment)
	}
	if got := state.partyRoster[0].Equipment[0].Name; got != "長劍" {
		t.Fatalf("物品名讀成 %q，want 長劍", got)
	}
}

// 走錯路的症狀要看得見：同一份原版檔用 remake 那條路讀，名字會是亂碼。
// 這條測試存在的意義是證明「兩條路真的不同」——如果哪天它變成 PASS 得很勉強
// （兩邊讀出同樣的字串），代表分流失效了。
func TestRemakePathMisreadsOriginalBig5Bytes(t *testing.T) {
	directory := t.TempDir()
	writeSAVGAMSlot(t, directory, 'D', big5(t, "艾拉"), big5(t, "長劍"))

	state := NewState(testCatalog())
	if err := state.LoadSAVGAMSlot(directory, 'D'); err != nil {
		t.Fatal(err)
	}
	if got := state.partyRoster[0].Name; got == "艾拉" {
		t.Fatal("remake 那條路把 Big5 讀對了——分流沒有作用，兩支函式其實一樣")
	}
}

// remake 自己寫的槽（UTF-8）用 remake 那條路讀要原樣回來；
// 拿匯入那條路讀會壞——這就是「不能自動判斷、要由呼叫端明講」的理由。
func TestRemakeOwnSlotSurvivesItsOwnPathAndBreaksOnTheImportPath(t *testing.T) {
	directory := t.TempDir()
	writeSAVGAMSlot(t, directory, 'E', []byte("艾拉"), []byte("長劍"))

	state := NewState(testCatalog())
	if err := state.LoadSAVGAMSlot(directory, 'E'); err != nil {
		t.Fatal(err)
	}
	if got := state.partyRoster[0].Name; got != "艾拉" {
		t.Fatalf("remake 自己的存檔讀成 %q，want 艾拉", got)
	}

	broken := NewState(testCatalog())
	if err := broken.LoadOriginalSAVGAMSlot(directory, 'E'); err != nil {
		t.Fatal(err)
	}
	if got := broken.partyRoster[0].Name; got == "艾拉" {
		t.Fatal("匯入那條路把 UTF-8 讀對了——那代表它沒有真的在解 Big5")
	}
}

// 英文名字兩條路必須完全一樣（ASCII 相容）。這是分流的**正對照**：
// 少了它，「兩條路不同」的測試可能是因為匯入那條路整個壞掉。
func TestBothPathsAgreeOnASCIINames(t *testing.T) {
	directory := t.TempDir()
	writeSAVGAMSlot(t, directory, 'F', []byte("ELLA"), []byte("LONG SWORD"))

	remake := NewState(testCatalog())
	if err := remake.LoadSAVGAMSlot(directory, 'F'); err != nil {
		t.Fatal(err)
	}
	imported := NewState(testCatalog())
	if err := imported.LoadOriginalSAVGAMSlot(directory, 'F'); err != nil {
		t.Fatal(err)
	}
	if remake.partyRoster[0].Name != imported.partyRoster[0].Name ||
		remake.partyRoster[0].Name != "ELLA" {
		t.Fatalf("ASCII 名字兩條路不一致：%q vs %q",
			remake.partyRoster[0].Name, imported.partyRoster[0].Name)
	}
	if remake.partyRoster[0].Equipment[0].Name != imported.partyRoster[0].Equipment[0].Name {
		t.Fatalf("ASCII 物品名兩條路不一致：%q vs %q",
			remake.partyRoster[0].Equipment[0].Name, imported.partyRoster[0].Equipment[0].Name)
	}
}
