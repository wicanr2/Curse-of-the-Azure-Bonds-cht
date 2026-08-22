package game

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
)

// pc98BGMSelector 是 spec 355 §「精確 selector 表」逐列抄下來的：PC-98 的
// `BGMPLAY` 不吃參數，直接讀 `CURRENTECL`（＝目前的 ECL block），依 block 決定
// 傳給 `MSCPLAY` 的 selector。
//
// ⚠ `0x30` 是「不變」、`0x52` 是「無分支」，兩者都**不該有 binding**——宣告了
// 就會在那兩段強行換曲。
var pc98BGMSelector = map[uint8]string{
	0x01: "pc98-bgm-selector-03", 0x31: "pc98-bgm-selector-03",

	0x11: "pc98-bgm-selector-04", 0x12: "pc98-bgm-selector-04",
	0x21: "pc98-bgm-selector-04", 0x22: "pc98-bgm-selector-04",
	0x23: "pc98-bgm-selector-04", 0x15: "pc98-bgm-selector-04",
	0x43: "pc98-bgm-selector-04", 0x45: "pc98-bgm-selector-04",

	// ⚠ `0x50`／`0x51` 依 `WLDTWN` 分兩支：0 是戶外導航（5），非零是城鎮設施
	// 選單（6）。context-free 的 fallback 是 5。
	0x50: "pc98-bgm-selector-05", 0x51: "pc98-bgm-selector-05",

	0x20: "pc98-bgm-selector-08", 0x40: "pc98-bgm-selector-08",
	0x42: "pc98-bgm-selector-08",

	0x02: "pc98-bgm-selector-09", 0x10: "pc98-bgm-selector-09",
	0x05: "pc98-bgm-selector-09", 0x35: "pc98-bgm-selector-09",

	0x03: "pc98-bgm-selector-0c", 0x04: "pc98-bgm-selector-0c",
	0x25: "pc98-bgm-selector-0c", 0x32: "pc98-bgm-selector-0c",
	0x33: "pc98-bgm-selector-0c",
}

// pc98TownServicesTrack 是 `WLDTWN != 0` 那一支。
const pc98TownServicesTrack = "pc98-bgm-selector-06"

// game pack 宣告的選曲要跟 PC-98 原作逐格對得上。`SEG-33` 前半只驗了
// 「有一首在曲目表裡的曲子」——那擋不住「每一段都播同一首」。
func TestMusicBindingsMatchThePC98Selector(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	for block, want := range pc98BGMSelector {
		binding, found := pack.FindMusicBinding(block, "")
		if !found {
			t.Errorf("block %#02X 沒有選曲宣告，spec 355 說是 %s", block, want)
			continue
		}
		if binding.TrackID != want {
			t.Errorf("block %#02X 宣告的是 %s，spec 355 說是 %s",
				block, binding.TrackID, want)
		}
	}
	// `0x30` 不換曲、`0x52` 沒有分支。
	for _, block := range []uint8{0x30, 0x52} {
		if binding, found := pack.FindMusicBinding(block, ""); found {
			t.Errorf("block %#02X 不該有選曲宣告，卻宣告了 %s", block, binding.TrackID)
		}
	}
	// 城鎮設施選單那一支要靠 context 才取得到。
	for _, block := range []uint8{0x50, 0x51} {
		binding, found := pack.FindMusicBinding(block, "pc98-town-services-menu")
		if !found || binding.TrackID != pc98TownServicesTrack {
			t.Errorf("block %#02X 的城鎮設施選單取到 %+v，應該是 %s",
				block, binding, pc98TownServicesTrack)
		}
	}
}

// expectedMusicForBlock 回傳一段結束時應該正在播的曲子。`0x50`／`0x51` 兩支都算
// 對，因為同一個 block 內會依 `PICTURE` cue 在戶外導航與城鎮設施選單之間切換。
func expectedMusicForBlock(block uint8) ([]string, bool) {
	want, ok := pc98BGMSelector[block]
	if !ok {
		return nil, false
	}
	if block == 0x50 || block == 0x51 {
		return []string{want, pc98TownServicesTrack}, true
	}
	return []string{want}, true
}

// ★ 酒館傳聞的規則是**子字串比對**，而 `MatchText` 先命中先贏——
// `TAVERN TALE 3` 是 `TAVERN TALE 31` 的子字串。兩位數的規則要是排在一位數後面，
// 第 31 則就會被第 3 則的規則攔走，**而且兩邊都是合法的傳聞、畫面上看不出錯**。
//
// ⚠ 這條測試不是在驗「有沒有譯文」，是在驗**順序**：每一則傳聞餵自己的原文，
// 都要命中自己那一條。
func TestTavernTaleRulesDoNotStealEachOther(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	numbers := []int{3, 9, 13, 15, 17, 20, 24, 28, 31, 43, 44, 45, 52, 56, 59, 60}
	for _, number := range numbers {
		text := fmt.Sprintf(
			"AS YOU CONSUME THE LOCAL EXCUSE FOR FOOD AND DRINK, YOU OVERHEAR TAVERN TALE %d", number)
		result := pack.MatchText([]string{text}, pack.DefaultLocale)
		if !result.Matched {
			t.Errorf("第 %d 則傳聞沒有規則命中", number)
			continue
		}
		want := fmt.Sprintf("tavern-tale-%d", number)
		if !strings.HasSuffix(result.RuleID, want) {
			t.Errorf("第 %d 則傳聞命中的是 %q，被別條攔走了（要的是 …%s）",
				number, result.RuleID, want)
		}
	}
}
