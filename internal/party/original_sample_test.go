package party

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// 原版自己寫出來的角色檔（`docs/reference/original-dos/save-samples/`）。
//
// ★ 這是第一份**原版產生**的存檔樣本：在 Docker/DOSBox 裡用遊戲自己的
// `START.EXE STING Wooden` 測試模式建一個角色再存檔，遊戲寫出 `BOB.GUY`
// （422 bytes）與 `BOB.FX`（27 bytes）。spec 102 當年寫「那一層必須先取得
// 實際 sample bytes」——這一組就是。
//
// 釘住的是「解碼器對原版位元組成立」，不是對 remake 自己寫的檔成立：
// 兩者版面相同，用 remake 自己的輸出當樣本等於沒測（spec 1121）。
func TestOriginalDOSPlayerSampleDecodes(t *testing.T) {
	base := filepath.Join("..", "..", "docs", "reference", "original-dos", "save-samples")
	record, err := os.ReadFile(filepath.Join(base, "BOB.GUY"))
	if err != nil {
		t.Skipf("原版樣本不存在：%v", err)
	}
	effects, err := os.ReadFile(filepath.Join(base, "BOB.FX"))
	if err != nil {
		t.Fatal(err)
	}
	if len(record) != DOSPlayerRecordSize {
		t.Fatalf("原版角色記錄是 %d bytes，want %d", len(record), DOSPlayerRecordSize)
	}
	// ★ 名字欄是 Turbo Pascal 短字串：第一個位元組是長度。原版寫的就是 03 'B' 'O' 'B'。
	if record[0] != 3 || string(record[1:4]) != "BOB" {
		t.Fatalf("名字欄開頭是 %v，want 03 'BOB'", record[:4])
	}
	if len(effects)%monster.AffectRecordSize != 0 {
		t.Fatalf(".FX 是 %d bytes，不是 %d 的倍數", len(effects), monster.AffectRecordSize)
	}

	character, err := ParseOriginalDOSPlayerFiles("sample", DOSPlayerFiles{Record: record, Effects: effects})
	if err != nil {
		t.Fatal(err)
	}
	if character.Name != "BOB" {
		t.Fatalf("角色名讀成 %q，want BOB", character.Name)
	}
	if character.MaxHitPoints <= 0 {
		t.Fatalf("最大 HP 讀成 %d", character.MaxHitPoints)
	}
	if character.Abilities.Strength <= 0 {
		t.Fatalf("力量讀成 %d", character.Abilities.Strength)
	}
}
