package combat

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
)

// `CHECKFX(timing, 對象)`：問「這個時機，這個人身上有哪些效果要介入」（spec 1123）。
//
// ★ 這一層取代的是「一個效果一段 if」。 原作把修正寫進一組固定的暫存全域，
// 呼叫端讀自己要的那一個；timing 表決定要問哪些效果碼，修正表給數字。
// remake 這一側照抄這個形狀，所以新增一個效果只要資料，不必動戰鬥程式碼。
//
// ⚠ **回傳 0 有兩種意思**：沒有任何效果介入，或是介入的效果 handler 還沒解讀。
// 要分辨就看 `CheckFXDetail` 的 `Unread`——覆蓋報告必須把兩者分開，
// 否則「還沒讀」會被算成「沒有影響」。

// CheckFX 常用的幾個時機。名字取自已知的呼叫點（spec 1123 的表）。
const (
	// CheckFXAttackTarget 是 `ATTEMPTTOHIT` 對**目標**問的（`0Ah`）。
	CheckFXAttackTarget uint8 = 0x0A
	// CheckFXAttacker 是 `ATTEMPTTOHIT` 對**攻擊者**問的（`10h`）。
	CheckFXAttacker uint8 = 0x10
	// CheckFXSavingThrow 是 `MAKESAVE`（`0Ch`）。
	CheckFXSavingThrow uint8 = 0x0C
	// CheckFXMorale 是士氣檢定（`11h`，spec 1122）。
	CheckFXMorale uint8 = 0x11
	// CheckFXMovement 是移動率換算（`12h`，spec 1122）。
	CheckFXMovement uint8 = 0x12
	// checkFXDamage 是 `PUTDAMAGE` 進入時（`06h`，spec 581）：抗性在這裡介入。
	checkFXDamage uint8 = 0x06
	// checkFXPutEffect 是 `PUTEFFECT`（`09h`，spec 581）：掛效果之前的閘。
	checkFXPutEffect uint8 = 0x09
	// CheckFXMeleeAttacker／CheckFXMeleeTarget 是近戰傷害算完之後的兩次查詢
	// （`04h` 對攻擊者、`05h` 對目標，呼叫點在 `overlay-13:01F0h`）。
	// 衰弱射線就是掛在攻擊者那一次：傷害減 25%。
	CheckFXMeleeAttacker uint8 = 0x04
	CheckFXMeleeTarget   uint8 = 0x05
	// CheckFXCanAct 是「這個人這一回合動得了嗎」（`07h`）：定身家族、纏繞術
	// 都在這張清單裡。
	CheckFXCanAct uint8 = 0x07
	// checkFXArmourClass 是護甲那一組（`0Bh`）：防護邪惡／善良、護盾、
	// 妖火、致盲都在這張清單裡。呼叫點還沒對回原作，所以不匯出。
	checkFXArmourClass uint8 = 0x0B
)

// 暫存全域的名字，對應 `gamepack/rules/effect-modifiers.json` 的 `scratch_globals`。
const (
	// scratchModifier 是共用的修正暫存（原作 `DS:6F9Fh`）。**它不是「命中修正」**
	// ——同一支 handler 無條件寫它，意思由讀它的 timing 決定：`0Ah`／`10h` 是
	// 命中、`0Bh` 是護甲、`0Dh` 是死亡後。取中性的名字才不會在某一條路上說錯。
	scratchModifier    = "modifier"
	scratchSavingThrow = "saving_throw"
	scratchMorale      = "morale"
	scratchMovement    = "movement"
	scratchDamage      = "damage"
	// scratchDamageElement 是傷害屬性旗標：bit 0 火、bit 1 冷、bit 2 電。
	// 三個位元各有兩個獨立證人——抗性法術的守衛遮罩，與傷害法術推進
	// `sub_F06` 的那個值（spec 1124）。
	scratchDamageElement = "damage_element"
)

// CheckFXDetail 是一次 CHECKFX 的完整結果。
type CheckFXDetail struct {
	// Applied 是各個暫存全域的最終值（只含真的被動到的）。
	Applied map[string]int
	// Contributed 是真的貢獻了修正的效果碼。
	Contributed []uint8
	// Unread 是「這個人有這個效果、這個時機也要問它，但 handler 還沒解讀」。
	// **不是 0 修正**——是不知道。
	Unread []uint8
	// Records 是寫進角色／戰鬥狀態記錄的那一類修正。
	Records []CheckFXRecordWrite
}

// CheckFXRecordWrite 是一次對記錄的寫入。
type CheckFXRecordWrite struct {
	// Record 是 `player`（角色記錄）或 `combat_state`（`+18Dh` 的 22 bytes）。
	Record string
	Field  int
	Op     string
	Value  int
	// Cap／CapThreshold 只有 `add_capped` 用得到。
	Cap          int
	CapThreshold int
}

// CheckFX 對一個戰鬥員跑一次時機查詢。`base` 是各全域的起始值（通常只填一個）。
//
// 有條件的修正（`guard_global`／`guard_mask`）只有在 `base` 裡那個旗標與遮罩
// 有交集時才套用。所以要讓抗寒生效，呼叫端必須把傷害屬性旗標一起放進 `base`。
func CheckFX(fighter Fighter, timing uint8, base map[string]int) (CheckFXDetail, error) {
	table, err := gamepack.EffectModifiers()
	if err != nil {
		return CheckFXDetail{}, err
	}
	detail := CheckFXDetail{Applied: map[string]int{}}
	for name, value := range base {
		detail.Applied[name] = value
	}
	for _, code := range table.TimingEffects(timing) {
		if code < 0 || code > 0xFF {
			return CheckFXDetail{}, fmt.Errorf("timing %02Xh lists effect %d outside a byte", timing, code)
		}
		kind := uint8(code)
		if !fighterHasAnyAffect(fighter, []uint8{kind}) {
			continue
		}
		handler, ok := table.Handler(kind)
		if !ok {
			detail.Unread = append(detail.Unread, kind)
			continue
		}
		if len(handler.Modifiers) == 0 {
			// `inert` 是「原作這一支什麼都沒做」，那是已知結果不是缺口。
			if handler.Status != "inert" {
				detail.Unread = append(detail.Unread, kind)
			}
			continue
		}
		applied := false
		for _, modifier := range handler.Modifiers {
			// ⚠ 守衛要在**記錄寫入之前**檢查。原本記錄那一支排在守衛前面直接
			// `continue`，於是**有條件的記錄寫入完全繞過自己的條件**——效果 `54h`
			// （傷害屬性帶電才回血）不管什麼屬性都會回血。全域修正那一路沒事，
			// 因為它排在守衛後面；這條路只是先前沒有任何有守衛的記錄寫入，
			// 所以一直沒被觸發。
			if modifier.GuardGlobal != "" {
				guard := table.ScratchName(modifier.GuardGlobal)
				if detail.Applied[guard]&modifier.GuardMask == 0 {
					// 旗標不符：這個修正這一次不套。**不是 0，是不適用。**
					continue
				}
			}
			if modifier.Record != "" {
				detail.Records = append(detail.Records, CheckFXRecordWrite{
					Record: modifier.Record, Field: modifier.Field,
					Op: modifier.Op, Value: modifier.Value,
					Cap: modifier.Cap, CapThreshold: modifier.CapThreshold})
				applied = true
				continue
			}
			applied = true
			name := table.ScratchName(modifier.Global)
			switch modifier.Op {
			case "add":
				detail.Applied[name] += modifier.Value
			case "sub":
				detail.Applied[name] -= modifier.Value
			case "set":
				detail.Applied[name] = modifier.Value
			case "sub_clamped":
				// 原作的夾底減法：不足就歸零，不會變成負的。
				if detail.Applied[name] < modifier.Value {
					detail.Applied[name] = 0
				} else {
					detail.Applied[name] -= modifier.Value
				}
			case "div":
				if modifier.Value == 0 {
					return CheckFXDetail{}, fmt.Errorf("effect %02Xh divides by zero", kind)
				}
				detail.Applied[name] /= modifier.Value
			case "sub_fraction":
				// 減掉自己的 1/K（衰弱射線是傷害減 25%）。
				if modifier.Value == 0 {
					return CheckFXDetail{}, fmt.Errorf("effect %02Xh divides by zero", kind)
				}
				detail.Applied[name] -= detail.Applied[name] / modifier.Value
			default:
				return CheckFXDetail{}, fmt.Errorf("effect %02Xh uses unknown operation %q", kind, modifier.Op)
			}
		}
		if !applied {
			continue
		}
		detail.Contributed = append(detail.Contributed, kind)
		if handler.Status == "partial" {
			// 有數字但也有沒解析的指令：數字算數，其餘還不知道。
			detail.Unread = append(detail.Unread, kind)
		}
	}
	return detail, nil
}

// CheckFXValue 是只要一個全域的方便版。
func CheckFXValue(fighter Fighter, timing uint8, scratch string, base int) (int, error) {
	detail, err := CheckFX(fighter, timing, map[string]int{scratch: base})
	if err != nil {
		return 0, err
	}
	return detail.Applied[scratch], nil
}
