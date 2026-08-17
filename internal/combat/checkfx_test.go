package combat

import (
	"fmt"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
)

func gamepackEffectModifiers() (*gamepack.EffectModifierTable, error) {
	return gamepack.EffectModifiers()
}

// fmtSscanf 把兩位十六進位的鍵轉回整數。
func fmtSscanf(key string, out *int) (int, error) {
	return fmt.Sscanf(key, "%X", out)
}

func affected(kinds ...uint8) Fighter {
	fighter := Fighter{ID: "x", HitPoints: 10, MaxHitPoints: 10}
	for _, kind := range kinds {
		fighter.MonsterAffects = append(fighter.MonsterAffects, MonsterAffect{Kind: kind, Active: true})
	}
	return fighter
}

// 祝福：命中 ＋1、士氣 ＋5。這兩個數字同時出現在一支 handler 裡，
// 由 timing 決定呼叫端讀哪一個。
func TestBlessAddsToBothAttackAndMorale(t *testing.T) {
	attack, err := CheckFXValue(affected(0x01), CheckFXAttackTarget, scratchModifier, 0)
	if err != nil {
		t.Fatal(err)
	}
	if attack != 1 {
		t.Fatalf("命中修正 %d，want 1", attack)
	}
	morale, err := CheckFXValue(affected(0x01), CheckFXMorale, scratchMorale, 50)
	if err != nil {
		t.Fatal(err)
	}
	if morale != 55 {
		t.Fatalf("士氣 %d，want 55（50 ＋ 5）", morale)
	}
}

// 詛咒是祝福的鏡像；兩個一起掛就互相抵銷。
func TestCurseMirrorsBlessAndTheyCancel(t *testing.T) {
	value, err := CheckFXValue(affected(0x02), CheckFXAttackTarget, scratchModifier, 0)
	if err != nil {
		t.Fatal(err)
	}
	if value != -1 {
		t.Fatalf("詛咒的命中修正 %d，want −1", value)
	}
	both, err := CheckFXValue(affected(0x01, 0x02), CheckFXAttackTarget, scratchModifier, 0)
	if err != nil {
		t.Fatal(err)
	}
	if both != 0 {
		t.Fatalf("祝福＋詛咒 %d，want 0", both)
	}
}

// 防護邪惡的 ＋2／−2 **是有條件的**（原作在 handler 裡先比陣營），
// 所以它進不了修正表，只會被回報成「還沒解讀」。
//
// ★ 這條測試存在的理由是**擋住把條件當無條件**：照字面把那兩個數字收進表裡，
// 得到的是「看起來完整但對所有目標都生效」的錯規則。寧可標成不知道。
func TestConditionalProtectionIsReportedAsUnreadNotAsAFlatBonus(t *testing.T) {
	detail, err := CheckFX(affected(0x08), CheckFXSavingThrow, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Applied) != 0 {
		t.Fatalf("有條件的修正被當成無條件套上了：%+v", detail.Applied)
	}
	if len(detail.Unread) == 0 {
		t.Fatal("有條件的 handler 沒有被回報成待解讀")
	}
}

// 致盲在三個時機都是 −4（`0Ah` 共用修正、`0Bh` 護甲、`0Ch` 豁免）。
// 它是 `partial`：數字算數，另外還有動角色記錄的部分沒解讀。
func TestBlindnessTakesFourOffEveryTimingItAppearsIn(t *testing.T) {
	for _, item := range []struct {
		timing  uint8
		scratch string
	}{
		{CheckFXAttackTarget, scratchModifier},
		{checkFXArmourClass, scratchModifier},
		{CheckFXSavingThrow, scratchSavingThrow},
	} {
		value, err := CheckFXValue(affected(0x21), item.timing, item.scratch, 0)
		if err != nil {
			t.Fatal(err)
		}
		if value != -4 {
			t.Fatalf("timing %02Xh 的 %s ＝ %d，want −4", item.timing, item.scratch, value)
		}
	}
}

// 沒有掛效果就不動基準值。
func TestCheckFXLeavesTheBaseAloneWithoutEffects(t *testing.T) {
	value, err := CheckFXValue(Fighter{}, CheckFXAttackTarget, scratchModifier, 7)
	if err != nil {
		t.Fatal(err)
	}
	if value != 7 {
		t.Fatalf("沒有效果卻把 7 改成 %d", value)
	}
}

// 時機決定要問哪些效果：加速（27h）只在移動那一個時機出現。
func TestTimingSelectsWhichEffectsAreAsked(t *testing.T) {
	hasted := affected(0x27)
	if detail, err := CheckFX(hasted, CheckFXAttackTarget, nil); err != nil {
		t.Fatal(err)
	} else if len(detail.Contributed) != 0 || len(detail.Unread) != 0 {
		t.Fatalf("加速不該在命中那一個時機被問到：%+v", detail)
	}
	detail, err := CheckFX(hasted, CheckFXMovement, map[string]int{scratchMovement: 12})
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Contributed) == 0 && len(detail.Unread) == 0 {
		t.Fatal("加速在移動那一個時機一次都沒被問到")
	}
}

// ★ 回傳 0 有兩種意思，測試要能分辨。
func TestUnreadHandlersAreReportedInsteadOfSilentlyCountingAsZero(t *testing.T) {
	// 07h（妖火）在 `0Bh` 時機的清單裡，但 handler 還沒解讀。
	detail, err := CheckFX(affected(0x07), checkFXArmourClass, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Unread) == 0 {
		t.Fatal("還沒解讀的 handler 被當成「沒有修正」——覆蓋報告會虛報")
	}
	if len(detail.Applied) != 0 {
		t.Fatalf("沒有數字卻動了暫存值：%+v", detail.Applied)
	}
}

// 接線：祝福會讓士氣檢定真的好過一點。
func TestBlessRaisesTheMoraleCheck(t *testing.T) {
	build := func(blessed bool) *Battle {
		fighter := Fighter{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 20,
			ControlMorale: 0xA6, MovementAllowance: 2} // 26h × 2 ＝ 76
		if blessed {
			fighter.MonsterAffects = []MonsterAffect{{Kind: 0x01, Active: true}}
		}
		battle, err := NewBattle([]Fighter{fighter,
			{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, MovementAllowance: 12}}, 1)
		if err != nil {
			t.Fatal(err)
		}
		return battle
	}
	plain, err := build(false).CheckMorale("orc")
	if err != nil {
		t.Fatal(err)
	}
	blessed, err := build(true).CheckMorale("orc")
	if err != nil {
		t.Fatal(err)
	}
	if blessed.Morale != plain.Morale+5 {
		t.Fatalf("祝福後士氣 %d，未祝福 %d，差應該是 5", blessed.Morale, plain.Morale)
	}
	// 門檻 60（剩四成血）：76 過、71 也過，所以這裡只比數字不比結果。
	if plain.Threshold != 60 {
		t.Fatalf("門檻 %d，want 60", plain.Threshold)
	}
}

// 接線：緩速把移動預算折半。
func TestSlowHalvesTheApproachBudget(t *testing.T) {
	build := func(slowed bool) *Battle {
		fighter := Fighter{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8,
			MovementAllowance: 8, WeaponRange: 1,
			HasCombatPosition: true, CombatX: 1, CombatY: 1}
		if slowed {
			fighter.MonsterAffects = []MonsterAffect{{Kind: 0x2A, Active: true}}
		}
		battle, err := NewBattle([]Fighter{fighter,
			{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
				HasCombatPosition: true, CombatX: 1, CombatY: 15}}, 1)
		if err != nil {
			t.Fatal(err)
		}
		return battle
	}
	plain, err := build(false).MonsterApproach("orc", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	slowed, err := build(true).MonsterApproach("orc", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(slowed.Steps)*2 != len(plain.Steps) {
		t.Fatalf("緩速走了 %d 步、正常 %d 步，應該剛好一半",
			len(slowed.Steps), len(plain.Steps))
	}
}

// 接線：致盲讓豁免難過 4 點。用致盲而不是防護邪惡，因為防護邪惡的加值在原作
// 是有條件的（比陣營），沒有進修正表。
func TestBlindnessShiftsTheSavingThrow(t *testing.T) {
	build := func(blinded bool) Fighter {
		fighter := Fighter{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			SavingThrows: []uint8{12, 12, 12, 12, 12}}
		if blinded {
			fighter.MonsterAffects = []MonsterAffect{{Kind: 0x21, Active: true}}
		}
		return fighter
	}
	battle, err := NewBattle([]Fighter{build(false),
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8}}, 4)
	if err != nil {
		t.Fatal(err)
	}
	plainSaves, blindSaves := 0, 0
	for attempt := 0; attempt < 400; attempt++ {
		if result, err := battle.RollSavingThrow(build(false), 0, 0); err != nil {
			t.Fatal(err)
		} else if result.Saved {
			plainSaves++
		}
		if result, err := battle.RollSavingThrow(build(true), 0, 0); err != nil {
			t.Fatal(err)
		} else if result.Saved {
			blindSaves++
		}
	}
	if blindSaves >= plainSaves {
		t.Fatalf("致盲 %d 次過、正常 %d 次過——−4 沒有生效", blindSaves, plainSaves)
	}
}

// ★ `inert` 是「原作那一支什麼都沒做」，不是「還沒讀」。
// 兩個獨立的證據：handler 本體只有序幕與 `retf`，而且那個碼**不在任何 timing 的
// 清單裡**——所以連問都不會問到它。宣告這種法術是忠實的：玩家施得出來、
// 法術位會消耗，戰鬥規則不變。
func TestInertEffectCodesAppearInNoTimingList(t *testing.T) {
	table, err := gamepackEffectModifiers()
	if err != nil {
		t.Fatal(err)
	}
	inert := 0
	for key, handler := range table.Effects {
		if handler.Status != "inert" {
			continue
		}
		inert++
		var code int
		if _, err := fmtSscanf(key, &code); err != nil {
			t.Fatal(err)
		}
		for timing, codes := range table.Timings {
			for _, listed := range codes {
				if listed == code {
					t.Errorf("效果 %sh 標成 inert，卻出現在 timing %s 的清單裡", key, timing)
				}
			}
		}
	}
	if inert == 0 {
		t.Fatal("一個 inert 都沒有——這條測試等於沒跑")
	}
	// ★ 反向也要看：**有數字不等於套得上**。`CHECKFX` 是唯一會走 handler 的路，
	// 所以不在任何 timing 清單裡的碼，它的數字永遠套不上。
	// 目前已知兩個（`16h` 緩毒、`93h`）——那兩支大概是施法時直接 `CALLEFFECT`，
	// 不經過 timing 表。清單寫死是為了讓**新增的**這種情況當場現形。
	orphans := map[int]bool{0x16: true, 0x93: true}
	for key, handler := range table.Effects {
		if len(handler.Modifiers) == 0 {
			continue
		}
		var code int
		if _, err := fmtSscanf(key, &code); err != nil {
			t.Fatal(err)
		}
		if table.HasTiming(uint8(code)) {
			if orphans[code] {
				t.Errorf("效果 %sh 被列成「沒有 timing」，但它其實有——清單該更新", key)
			}
			continue
		}
		if !orphans[code] {
			t.Errorf("效果 %sh 有修正卻不在任何 timing 裡，那些數字永遠套不上", key)
		}
	}
}

// ★ 抗寒只減半**冷**傷害。旗標守衛少了的話它會減半所有傷害。
//
// 傷害屬性旗標的三個位元各有兩個獨立證人：抗性 handler 的守衛遮罩
// （抗火看 bit 0、抗寒看 bit 1），與傷害法術推進 `sub_F06` 的那個值
// （火焰打擊 `09h`、冰風暴 `0Ah`、電擊觸手 `0Ch`）。
func TestResistColdOnlyHalvesColdDamage(t *testing.T) {
	const (
		fire = 0x09
		cold = 0x0A
	)
	resistCold := affected(0x0A)
	chilled, err := CheckFX(resistCold, checkFXDamage,
		map[string]int{scratchDamage: 20, scratchDamageElement: cold})
	if err != nil {
		t.Fatal(err)
	}
	if chilled.Applied[scratchDamage] != 10 {
		t.Fatalf("冷傷害 %d，want 10（20 折半）", chilled.Applied[scratchDamage])
	}
	burned, err := CheckFX(resistCold, checkFXDamage,
		map[string]int{scratchDamage: 20, scratchDamageElement: fire})
	if err != nil {
		t.Fatal(err)
	}
	if burned.Applied[scratchDamage] != 20 {
		t.Fatalf("火傷害 %d，want 20：抗寒不該擋火", burned.Applied[scratchDamage])
	}
	// 抗火反過來。
	resistFire := affected(0x14)
	scorched, err := CheckFX(resistFire, checkFXDamage,
		map[string]int{scratchDamage: 20, scratchDamageElement: fire})
	if err != nil {
		t.Fatal(err)
	}
	if scorched.Applied[scratchDamage] != 10 {
		t.Fatalf("抗火對火傷害 %d，want 10", scorched.Applied[scratchDamage])
	}
	frozen, err := CheckFX(resistFire, checkFXDamage,
		map[string]int{scratchDamage: 20, scratchDamageElement: cold})
	if err != nil {
		t.Fatal(err)
	}
	if frozen.Applied[scratchDamage] != 20 {
		t.Fatalf("抗火對冷傷害 %d，want 20", frozen.Applied[scratchDamage])
	}
}

// 抗性同時給豁免 ＋3——同一支 handler 的第二個修正，同一個守衛。
func TestResistanceAlsoGivesThreeOnSaves(t *testing.T) {
	detail, err := CheckFX(affected(0x0A), checkFXDamage,
		map[string]int{scratchDamage: 20, scratchDamageElement: 0x0A,
			scratchSavingThrow: 0})
	if err != nil {
		t.Fatal(err)
	}
	if detail.Applied[scratchSavingThrow] != 3 {
		t.Fatalf("豁免修正 %d，want ＋3", detail.Applied[scratchSavingThrow])
	}
}
