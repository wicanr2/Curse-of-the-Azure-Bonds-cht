package ecl

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"

// EffectStatus 是「這個 opcode 的副作用在 remake 還原到什麼程度」。
//
// ★ 為什麼要有這張表。 `runtime.go` 對每個 opcode 都有 `case`，所以「有沒有
// handler」問不出東西——真正的差別在**那個 case 做了什麼**：有的完整執行，
// 有的只把運算元讀掉再往下走（`3Bh SPELL`、`3Ch PROTECTION` 就是），
// 兩者在程式碼裡長得很像，
// 在測試裡也一樣綠。表把差別寫死，`cmd/ecl-effect-coverage` 再乘上 corpus 的
// 出現次數，才知道「副作用還差多少」是多大的數字。
type EffectStatus string

const (
	// EffectDone：副作用在 remake 已經產生，且有回歸測試或實機路徑驗過。
	EffectDone EffectStatus = "done"
	// EffectPartial：做了一部分——通常是把請求記進 result 讓上層處理，
	// 或只還原了原作行為的一支。
	EffectPartial EffectStatus = "partial"
	// EffectConsumed：只把運算元讀掉、繼續往下走。**不是 no-op 的委婉說法**：
	// 它明確表示「這條指令在原作有效果，remake 目前沒有」。
	EffectConsumed EffectStatus = "consumed"
)

// OpcodeEffect 描述一個 opcode 的副作用還原狀態。
type OpcodeEffect struct {
	Status EffectStatus
	// Note 寫「還差什麼」或「憑什麼說做完了」，不寫這個 opcode 是什麼意思
	// ——那在 KnownCommands 裡。
	Note string
}

// OpcodeEffects 逐個 opcode 的還原狀態。
//
// ★ 判準（三態的分界，寫死免得下次又憑印象填）：
//
//	done      ECL 看得到的效果（記憶體、比較旗標、控制流）由 VM 產生，
//	          **而且**效果若在 VM 之外（隊伍、物品、畫面、資產），
//	          有**正式路徑**的程式碼把它套用。
//	partial   兩半只做了一半。
//	consumed  兩半都沒有——含「解碼後排進 result，但取用的 API 只有測試呼叫」。
//
// ⚠ **`runtime.go` 有 `case` 不算數。** 本表第一版把 `27h TREASURE` 判成
// `consumed`（其實 `combat_state.go` 有正式消費端）、把 `2Eh DAMAGE` 判成
// `done`（其實只結算全隊封包）。兩個方向都錯過一次，判準才寫成上面這樣：
// 要同時看「VM 產生了什麼」與「誰把它套用」。
//
// ⚠ 新增 opcode 沒有登記會讓 `cmd/ecl-effect-coverage` 的 fail-closed 檢查變紅，
// 不會被靜靜略過。
var OpcodeEffects = map[byte]OpcodeEffect{
	0x00: {EffectDone, tooltext.Text("h.c4a6beef7e5c")},
	0x01: {EffectDone, tooltext.Text("h.d3907402bbb9")},
	0x02: {EffectDone, tooltext.Text("h.a052a8690569")},
	0x03: {EffectDone, tooltext.Text("h.21908d2ddcbf")},
	0x04: {EffectDone, tooltext.Text("h.9255d7d3cbea")},
	0x05: {EffectDone, tooltext.Text("h.9255d7d3cbea")},
	0x06: {EffectDone, tooltext.Text("h.9255d7d3cbea")},
	0x07: {EffectDone, tooltext.Text("h.9255d7d3cbea")},
	0x08: {EffectDone, tooltext.Text("h.3931af437d83")},
	0x09: {EffectDone, tooltext.Text("h.82688b4fd23e")},
	0x0A: {EffectDone, tooltext.Text("h.8367988bb4e0")},
	0x0B: {EffectDone, tooltext.Text("h.107f89522a6d")},
	0x0C: {EffectDone, tooltext.Text("h.8e38d0fa542f")},
	0x0D: {EffectDone, tooltext.Text("h.5c281307c3c7")},
	0x0E: {EffectDone, tooltext.Text("h.48e5c159f777")},
	0x0F: {EffectConsumed, tooltext.Text("h.4bd10192efe1")},
	0x10: {EffectDone, tooltext.Text("h.1d2ca21635d4")},
	0x11: {EffectDone, tooltext.Text("h.8cd4389d36a7")},
	0x12: {EffectDone, tooltext.Text("h.21c65905a289")},
	0x13: {EffectDone, tooltext.Text("h.d3907402bbb9")},
	0x14: {EffectDone, tooltext.Text("h.8b50e41957ed")},
	0x15: {EffectDone, tooltext.Text("h.15d032a528f7")},
	0x16: {EffectDone, tooltext.Text("h.0f973fc6cacc")},
	0x17: {EffectDone, tooltext.Text("h.15552b136dfb")},
	0x18: {EffectDone, tooltext.Text("h.15552b136dfb")},
	0x19: {EffectDone, tooltext.Text("h.15552b136dfb")},
	0x1A: {EffectDone, tooltext.Text("h.15552b136dfb")},
	0x1B: {EffectDone, tooltext.Text("h.15552b136dfb")},
	0x1C: {EffectDone, tooltext.Text("h.608640d02202")},
	0x1D: {EffectDone, tooltext.Text("h.55a7e52ff062")},
	0x1E: {EffectDone, tooltext.Text("h.4542eded7603")},
	0x1F: {EffectConsumed, tooltext.Text("h.4fc0c1270b06")},
	0x20: {EffectDone, tooltext.Text("h.abc7314f1c2f")},
	0x21: {EffectDone, tooltext.Text("h.b28a09b78483")},
	0x22: {EffectDone, tooltext.Text("h.af8825cd37bf")},
	0x23: {EffectConsumed, tooltext.Text("h.a29067b56fba")},
	0x24: {EffectDone, tooltext.Text("h.15af3ce628a3")},
	0x25: {EffectDone, tooltext.Text("h.42cb66584c99")},
	0x26: {EffectDone, tooltext.Text("h.3d89184fd0e0")},
	0x27: {EffectDone, tooltext.Text("h.1a8253772b2f")},
	0x28: {EffectDone, tooltext.Text("h.e09dcc040d10")},
	0x29: {EffectDone, tooltext.Text("h.64cb9b9d91a1")},
	0x2A: {EffectDone, tooltext.Text("h.110d3bbf297c")},
	0x2B: {EffectDone, tooltext.Text("h.91d0da2513ac")},
	0x2C: {EffectDone, "PARLAY"},
	0x2D: {EffectDone, tooltext.Text("h.36fb7b8e9cfb")},
	0x2E: {EffectDone, tooltext.Text("h.0c8efd756060")},
	0x2F: {EffectDone, tooltext.Text("h.3adbb7485471")},
	0x30: {EffectDone, tooltext.Text("h.3adbb7485471")},
	0x31: {EffectDone, tooltext.Text("h.6ce61d6eb667")},
	0x32: {EffectDone, "FIND ITEM"},
	0x33: {EffectDone, tooltext.Text("h.fbcb3b508235")},
	0x34: {EffectDone, "ECL CLOCK"},
	0x35: {EffectDone, "SAVE TABLE"},
	0x36: {EffectDone, tooltext.Text("h.f7ad6af83767")},
	0x37: {EffectDone, tooltext.Text("h.0919d8983f51")},
	0x38: {EffectDone, tooltext.Text("h.428413489e7c")},
	0x39: {EffectDone, tooltext.Text("h.969192ccc3e4")},
	0x3A: {EffectDone, "DELAY"},
	0x3B: {EffectDone, tooltext.Text("h.7f5fae3e885f")},
	0x3C: {EffectDone, tooltext.Text("h.83ae25bb15eb")},
	0x3D: {EffectDone, tooltext.Text("h.eccaf25783a1")},
	0x3E: {EffectDone, "DUMP"},
	0x3F: {EffectDone, "FIND SPECIAL"},
	0x40: {EffectDone, "DESTROY ITEMS"},
}
