// Package audiomap 收原作的換曲點與 remake 這一側的接法。
//
// ★ 存在的理由：「換曲點接上幾個」以前只是敘述，寫在報表的文字裡，而沒有任何
// 東西會在接上一個的時候更新它。這個套件把那張對照表變成**程式碼裡的分母**：
// 原作有幾個換曲點來自 `cmd/pc98-music-triggers` 的掃描結果，remake 接了幾個
// 由 game pack 現場查出來。
package audiomap

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"

// 事件驅動換曲點在 game pack 那一側的 context。原作那幾處都**不看
// `CURRENTECL`**（各自把曲號推給 `MSCPLAY` 或直接寫 `MUSICNO` 之後派曲），
// 所以 pack 那一側是每一段都列。
const (
	// CombatContext ＝ 開戰（`INITCOMBAT`，COMPREP）。
	CombatContext = "pc98-combat"
	// CombatDungeonTwoContext ＝ 開戰且 `LOADMONNUM == 47h`。
	CombatDungeonTwoContext = "pc98-combat-dungeon-two"
	// TitleContext ＝ 開場（`DOINTRO`，overlay-01 `093Ch`）。
	TitleContext = "pc98-title"
	// CreationContext ＝ 角色建立（`GEN`，overlay-17 `0B08h`）。
	CreationContext = "pc98-character-creation"
	// EndingContext ＝ 結局過場（overlay-18 `168Dh`）。
	EndingContext = "pc98-ending"
	// PartyWipeContext ＝ 全滅（POSTCOM，overlay-05 `1955h`）。
	PartyWipeContext = "pc98-party-wipe"
	// TownServicesContext ＝ 派曲表裡走 cue 的那一個（村莊）。
	TownServicesContext = "pc98-town-services-menu"
)

// ChangePoint 是原作的一個換曲點。
type ChangePoint struct {
	// Site 是模組與位移，PC-98 版面。
	Site string
	// Event 是那一刻發生什麼事。
	Event string
	// Selector 是 `MUSICNO` 的值，**1 起算**（pack 的 `reference_selector`）。
	Selector int
	// Context 是 remake 這一側的 binding context；查 `CURRENTECL` 派曲表的那幾個
	// 是空字串（照段綁）。
	Context string
}

// ChangePoints 是全部 13 個換曲點（`docs/audit/pc98-music-triggers.md`）。
//
// ⚠ 位移是 **PC-98** 的。DOS 版沒有音樂——沒有驅動也沒有音樂資料檔
// （`TestDOSImageHasNoMusicData`），這一格整個是 PC-98 專屬。
var ChangePoints = []ChangePoint{
	{tooltext.Text("h.aa93793158b2"), tooltext.Text("h.5bcecbadc726"), 3, ""},
	{tooltext.Text("h.7888920cc340"), tooltext.Text("h.d9d43a1eb5c8"), 4, ""},
	{tooltext.Text("h.5bd04fc436ec"), tooltext.Text("h.f337a03bc2e3"), 6, TownServicesContext},
	{tooltext.Text("h.f4006cf27d32"), tooltext.Text("h.7042af1ec5b6"), 5, ""},
	{tooltext.Text("h.a2b6506d0466"), tooltext.Text("h.83fe968333e2"), 8, ""},
	{tooltext.Text("h.6615f4549cd1"), tooltext.Text("h.01cb6b718229"), 9, ""},
	{tooltext.Text("h.b581670ab870"), tooltext.Text("h.d1b8e0129732"), 12, ""},
	{"overlay-01 093Ch", tooltext.Text("h.28c3432b6257"), 1, TitleContext},
	{"overlay-17 GEN 0B08h", tooltext.Text("h.f7865e52f8b7"), 2, CreationContext},
	{"overlay-10 COMPREP 1DA1h", tooltext.Text("h.b284cba9a979"), 7, CombatContext},
	{"overlay-10 COMPREP 1D97h", tooltext.Text("h.9f4f985c6502"), 11, CombatDungeonTwoContext},
	{"overlay-05 POSTCOM 1955h", tooltext.Text("h.7de3b9fbad0a"), 2, PartyWipeContext},
	{"overlay-18 168Dh", tooltext.Text("h.9a3017542ef8"), 10, EndingContext},
}

// Binding 是 pack 那一側的一條曲目綁定，由呼叫端從 game pack 餵進來。
// 這個套件不 import gamepack——那會讓 `internal/game` 與工具程式繞一大圈。
type Binding struct {
	Context string
	TrackID string
}

// Resolution 是一個換曲點的查詢結果。
type Resolution struct {
	Point ChangePoint
	// TrackID 是 pack 裡對應的曲目；查不到是空字串。
	TrackID string
	// Wired 為真代表 pack 同時有那個曲目、也有指向它的綁定。
	Wired bool
}

// Resolve 把每一個換曲點對到 pack 的綁定。
//
// `trackBySelector` 是 `reference_selector` → 曲目 ID（1 起算）。
//
// ⚠ 這只回答「**有沒有落點**」，不回答「在原作會發的那一刻發」——後者要實機
// 比對。
func Resolve(trackBySelector map[int]string, bindings []Binding) []Resolution {
	out := make([]Resolution, 0, len(ChangePoints))
	for _, point := range ChangePoints {
		result := Resolution{Point: point, TrackID: trackBySelector[point.Selector]}
		if result.TrackID != "" {
			for _, binding := range bindings {
				if binding.Context == point.Context && binding.TrackID == result.TrackID {
					result.Wired = true
					break
				}
			}
		}
		out = append(out, result)
	}
	return out
}

// Wired 數有幾個換曲點接上了。
func Wired(results []Resolution) int {
	count := 0
	for _, result := range results {
		if result.Wired {
			count++
		}
	}
	return count
}
