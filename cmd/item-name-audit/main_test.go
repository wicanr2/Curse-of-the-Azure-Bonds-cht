package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

const (
	imageFile  = "../../curseoftheazurebonds.zip"
	dsegFile   = "../../workplace/re-sweep/dos/dseg/dos-dseg-dseg.bin"
	localeFile = "../../assets/locale/zh-TW.json"
)

func loadCatalog(t *testing.T) locale.Catalog {
	t.Helper()
	payload, err := os.ReadFile(localeFile)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(payload)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

// 詞表的基底與間距是**斷言**，不是註解：拿兩端與中間三個已知詞去釘。
// 抓錯基底的話這三個一定同時錯位（spec 1178）。
func TestWordTableBaseAndStrideAreAnchored(t *testing.T) {
	words, err := loadWordTable(dsegFile)
	if err != nil {
		t.Skipf("找不到 DOS 資料段 dump，跳過：%v", err)
	}
	if len(words) != wordTableCount {
		t.Fatalf("詞表 %d 筆，預期 %d 筆", len(words), wordTableCount)
	}
	for number, want := range map[uint8]string{
		0x01: "Battle Axe", 0x24: "Long Sword", 0xA2: "+1",
		0xA7: "of", 0xD4: "With 3 Spells", 0xFF: "Cursed",
	} {
		if got := words[number]; got != want {
			t.Errorf("第 %02Xh 筆 ＝ %q，預期 %q", number, got, want)
		}
	}
	// 表的**尾巴**要釘住：`FFh` 之後是另一張表（方位名），不是名稱成分。
	// 沒有這一條的話把 count 往上調也不會有人發現。
	if words[0xFE] != "Pass" {
		t.Errorf("第 FEh 筆 ＝ %q，預期 %q", words[0xFE], "Pass")
	}
}

// 缺一個名稱成分的翻譯，玩家看到的是**少一截的名字**而且沒有任何錯誤訊息。
// 這條就是那個 fail-closed 閘門：原版資料用到的編號，一個都不能沒譯。
func TestEveryUsedNameNumberIsTranslated(t *testing.T) {
	items, err := collectOriginalItems(imageFile)
	if err != nil {
		t.Skipf("找不到遊戲 image，跳過：%v", err)
	}
	if len(items) == 0 {
		t.Fatal("一件物品都沒讀到——正對照失敗，這個測試等於沒跑")
	}
	catalog := loadCatalog(t)
	missing := map[uint8]bool{}
	for _, item := range items {
		for _, number := range item.NameNumbers {
			if number == 0 {
				continue
			}
			if catalog.Text(fmt.Sprintf("item_name_%02X", number), "") == "" {
				missing[number] = true
			}
		}
	}
	if len(missing) > 0 {
		t.Fatalf("有 %d 個名稱編號沒譯：%v", len(missing), missing)
	}
}

// 整條鏈（DAX 欄位 → 名稱編號 → 詞表 → 語系檔 → 組名規則）的端對端樣本。
// 挑的是四種不同形狀：單成分、加值後綴、連接詞對調、附加成分。
func TestComposedNamesMatchTheGoldenSample(t *testing.T) {
	items, err := collectOriginalItems(imageFile)
	if err != nil {
		t.Skipf("找不到遊戲 image，跳過：%v", err)
	}
	catalog := loadCatalog(t)
	want := map[[3]uint8]string{
		{0x00, 0x00, 0x24}: "長劍",         // Long Sword
		{0x00, 0xA2, 0x24}: "長劍+1",       // Long Sword +1
		{0xF2, 0xA7, 0x45}: "火球之魔杖",      // Wand of Fireballs
		{0x00, 0x74, 0x6C}: "伊奧恩石‧深紅",    // Ioun Stone Deep Red
		{0xF1, 0xA2, 0x24}: "長劍+1‧火焰舌",   // Long Sword +1 Flame Tongue
		{0x00, 0xD4, 0xD1}: "法師卷軸（三個法術）", // MU Scroll With 3 Spells
		{0x00, 0x30, 0x37}: "鏈甲",         // Chain Mail
	}
	seen := map[[3]uint8]bool{}
	for _, item := range items {
		expected, ok := want[item.NameNumbers]
		if !ok {
			continue
		}
		seen[item.NameNumbers] = true
		item.HiddenNameFlags = 0
		item.Count = 0
		if got := monster.LocalizedItemName(item, catalog); got != expected {
			t.Errorf("%v 組出 %q，預期 %q", item.NameNumbers, got, expected)
		}
	}
	for numbers := range want {
		if !seen[numbers] {
			t.Errorf("原版資料裡找不到名稱編號 %v——樣本挑錯了，這一條沒驗到東西", numbers)
		}
	}
}
