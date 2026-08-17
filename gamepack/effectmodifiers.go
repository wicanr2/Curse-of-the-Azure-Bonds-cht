package gamepack

import (
	"encoding/json"
	"fmt"
	"sync"
)

// 效果 handler 的修正表（spec 1123）。
//
// ★ 為什麼是資料不是程式碼。 原作的效果 handler 是一批極短的函式，把修正寫進
// 一組固定的全域，呼叫端再讀自己要的那一個。`CHECKFX(timing)` 只決定「這個時機
// 有哪些效果要介入」。兩張表湊起來就是完整的規則，remake 這一側不需要一個效果
// 一段程式碼。
//
// ⚠ `status` 分四級。**只有 `decoded` 與 `partial` 有數字**；`unread` 是「有內容
// 但不是加減型」（傷害、狀態、區域那一類），`inert` 是「這一支什麼都沒做」。
// 把 `unread` 當成「沒有修正」會讓覆蓋報告看起來比實際好。

// EffectModifier 是一次對暫存全域的加減。
type EffectModifier struct {
	// Global 是原作的全域位址（小寫十六進位，如 `6f9f`）。
	Global string `json:"global"`
	// Op 是 `add`／`sub`／`set`／`sub_clamped`（夾底減法）／`div`。
	Op    string `json:"op"`
	Value int    `json:"value"`
	// GuardGlobal／GuardMask 是「這個修正有條件」：只有當呼叫端提供的旗標
	// `GuardGlobal` 與 `GuardMask` 有交集時才套。抗寒只減半**冷**傷害，
	// 抗火只減半火傷害——少了這一層，它們會減半所有傷害。
	GuardGlobal string `json:"guard_global,omitempty"`
	GuardMask   int    `json:"guard_mask,omitempty"`
	// Record／Field 是「寫進記錄而不是全域」：`player` 是角色記錄，
	// `combat_state` 是 `+18Dh` 指到的 22 bytes 戰鬥狀態（spec 806）。
	Record string `json:"record,omitempty"`
	Field  int    `json:"field,omitempty"`
	// Cap 與 CapThreshold 是封頂加法：未達門檻就加 Value，否則設成 Cap。
	Cap          int `json:"cap,omitempty"`
	CapThreshold int `json:"cap_threshold,omitempty"`
}

// EffectHandler 是一個效果碼的 handler 摘要。
type EffectHandler struct {
	Overlay      int              `json:"overlay"`
	Offset       int              `json:"offset"`
	Instructions int              `json:"instructions"`
	Modifiers    []EffectModifier `json:"modifiers"`
	Unparsed     int              `json:"unparsed"`
	Status       string           `json:"status"`
}

// ScratchGlobal 是一個規則暫存全域的名字與佐證。
type ScratchGlobal struct {
	Name string `json:"name"`
	Note string `json:"note"`
}

// EffectModifierTable 是整份表。
type EffectModifierTable struct {
	SchemaVersion  int                      `json:"schema_version"`
	Source         string                   `json:"source"`
	Spec           string                   `json:"spec"`
	ScratchGlobals map[string]ScratchGlobal `json:"scratch_globals"`
	// Timings 的鍵是兩位十六進位的 timing，值是該時機要介入的效果碼。
	Timings map[string][]int `json:"timings"`
	// Effects 的鍵是兩位十六進位的效果碼。
	Effects map[string]EffectHandler `json:"effects"`
}

var (
	effectModifiersOnce sync.Once
	effectModifiers     *EffectModifierTable
	effectModifiersErr  error
)

// EffectModifiers 回傳嵌入的效果修正表。
func EffectModifiers() (*EffectModifierTable, error) {
	effectModifiersOnce.Do(func() {
		data, err := ruleFiles.ReadFile("rules/effect-modifiers.json")
		if err != nil {
			effectModifiersErr = fmt.Errorf("read embedded effect modifiers: %w", err)
			return
		}
		parsed := &EffectModifierTable{}
		if err := json.Unmarshal(data, parsed); err != nil {
			effectModifiersErr = fmt.Errorf("parse embedded effect modifiers: %w", err)
			return
		}
		if len(parsed.Effects) == 0 || len(parsed.Timings) == 0 {
			effectModifiersErr = fmt.Errorf("effect modifier table is empty")
			return
		}
		effectModifiers = parsed
	})
	return effectModifiers, effectModifiersErr
}

// TimingEffects 回傳某個時機要問的效果碼。
func (t *EffectModifierTable) TimingEffects(timing uint8) []int {
	if t == nil {
		return nil
	}
	return t.Timings[fmt.Sprintf("%02X", timing)]
}

// Handler 回傳一個效果碼的 handler 摘要。
func (t *EffectModifierTable) Handler(code uint8) (EffectHandler, bool) {
	if t == nil {
		return EffectHandler{}, false
	}
	handler, ok := t.Effects[fmt.Sprintf("%02X", code)]
	return handler, ok
}

// HasTiming 回答「有沒有任何時機會問到這個效果碼」。沒有的話它的 handler
// 永遠不會被 `CHECKFX` 走到——**有數字不等於套得上**。
func (t *EffectModifierTable) HasTiming(code uint8) bool {
	if t == nil {
		return false
	}
	for _, codes := range t.Timings {
		for _, listed := range codes {
			if listed == int(code) {
				return true
			}
		}
	}
	return false
}

// ScratchName 把全域位址翻成名字；沒有名字時回傳原位址。
func (t *EffectModifierTable) ScratchName(global string) string {
	if t == nil {
		return global
	}
	if named, ok := t.ScratchGlobals[global]; ok && named.Name != "" {
		return named.Name
	}
	return global
}
