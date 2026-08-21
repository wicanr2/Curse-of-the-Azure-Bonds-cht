package main

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
)

const (
	imageFile = "../../curseoftheazurebonds.zip"
	coreFile  = "../../gamepack/pack/00-core.json"
	zhFile    = "../../gamepack/pack/20-locale.zh-TW.json"
	enFile    = "../../gamepack/pack/20-locale.en.json"
)

// 原版每一種戰鬥員都要有繁中與英文名字。
//
// ★ 這是 fail-closed 閘門：`LocalizeCombatantName` 找不到宣告時**原樣回傳**，
// 不會有錯誤訊息——玩家在戰鬥選單直接看到 `THRI-KREEN`。所以「還有幾種是英文」
// 必須由測試回答，不能靠翻閱（spec 1179）。
//
// ⚠ 走的是 pack 的**執行時查詢**而不是比對兩份 JSON 的鍵：宣告與翻譯各自對得
// 起來、但查詢路徑串不起來的話，比 JSON 一樣會過。
func TestEveryOriginalCombatantNameIsLocalized(t *testing.T) {
	roster, err := CollectRoster(imageFile)
	if err != nil {
		t.Skipf("找不到遊戲 image，跳過：%v", err)
	}
	if len(roster) == 0 {
		t.Fatal("一種戰鬥員都沒讀到——正對照失敗，這個測試等於沒跑")
	}
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	missing := map[string][]string{}
	for source := range roster {
		for _, language := range []string{"zh-TW", "en"} {
			value, found := pack.LocalizeCombatantName(source, language)
			if !found || value == "" {
				missing[source] = append(missing[source], language)
			}
		}
	}
	if len(missing) > 0 {
		t.Fatalf("有 %d 種戰鬥員的名字沒有翻譯：%v", len(missing), missing)
	}
}

// 名字不能兩種生物撞在一起：撞名的話玩家在戰鬥選單分不出誰是誰。
func TestCombatantTranslationsAreDistinct(t *testing.T) {
	rules, err := LoadNameRules(coreFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		path     string
		language string
	}{{zhFile, "zh-TW"}, {enFile, "en"}} {
		table, loadErr := LoadLocale(item.path, item.language)
		if loadErr != nil {
			t.Fatal(loadErr)
		}
		seen := map[string]string{}
		for _, rule := range rules {
			value := table[rule.MessageID]
			if value == "" {
				t.Errorf("%s 缺 %s", item.language, rule.MessageID)
				continue
			}
			if other, clash := seen[value]; clash {
				t.Errorf("%s 的 %q 同時是 %s 與 %s", item.language, value, other, rule.Source)
			}
			seen[value] = rule.Source
		}
	}
}

// 宣告的 `source` 必須真的出現在原版資料裡——寫錯一個字的宣告會靜靜地永遠不生效。
func TestEveryDeclaredSourceExistsInTheOriginalData(t *testing.T) {
	roster, err := CollectRoster(imageFile)
	if err != nil {
		t.Skipf("找不到遊戲 image，跳過：%v", err)
	}
	rules, err := LoadNameRules(coreFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range rules {
		if _, ok := roster[rule.Source]; !ok {
			t.Errorf("宣告了 %q 但原版資料裡沒有這個名字", rule.Source)
		}
	}
}
