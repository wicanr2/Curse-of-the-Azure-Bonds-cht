package game

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
)

// 帶效果 `5Ah` 的怪先手：AI 回合走特殊攻擊，訊息是中文的區域吐酸。
// 區域形不擲命中骰，貼身、範圍裡沒有牠的同伴 ⇒ 必定發動。
func TestMonsterSpecialAttackBreathesAcidInTheAITurn(t *testing.T) {
	data, err := os.ReadFile("../../assets/locale/zh-TW.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	value := NewState(catalog)
	state := &value
	if err := state.StartCombat(
		[]combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
			HitPoints: 200, MaxHitPoints: 200, HasCombatPosition: true, CombatX: 6, CombatY: 5}},
		[]combat.Fighter{{ID: "wyrm", Name: "幼龍", Side: combat.SideEnemy,
			HitPoints: 24, MaxHitPoints: 24, InitiativeBonus: 30,
			HasCombatPosition: true, CombatX: 5, CombatY: 5,
			MonsterAffects: []combat.MonsterAffect{{Kind: 0x5A, Innate: true}}}},
		11,
	); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf(catalog.Text("combat_breathes_acid", ""), "幼龍", 1, 0)
	prefix := want[:strings.Index(want, "命中")]
	if !strings.HasPrefix(state.CombatMessage(), prefix) {
		t.Fatalf("AI turn message=%q, want acid breath prefix %q",
			state.CombatMessage(), prefix)
	}
	hero, ok := state.fighter("hero")
	if !ok {
		t.Fatal("hero disappeared")
	}
	if hero.HitPoints != 200-24 && hero.HitPoints != 200-12 {
		t.Fatalf("hero HP=%d, want 176 (full) or 188 (saved half)", hero.HitPoints)
	}
}

// 凝視：直接叫分派器驗訊息與麻痺（不經回合迴圈，避免訊息被後續動作蓋掉）。
// 沒有豁免表 ⇒ 門檻 0 ⇒ 必失敗（RollSavingThrow 的 fail-closed）；
// 門檻 2 ⇒ 幾乎必過 ⇒ 抵抗訊息。
func TestMonsterSpecialAttackGazeUsesGazeMessages(t *testing.T) {
	data, err := os.ReadFile("../../assets/locale/zh-TW.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		saves     []uint8
		paralyzed bool
	}{
		{name: "no-save-table-paralyzes", saves: nil, paralyzed: true},
		{name: "easy-save-resists", saves: []uint8{0, 2}, paralyzed: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := NewState(catalog)
			state := &value
			if err := state.StartCombat(
				[]combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
					HitPoints: 200, MaxHitPoints: 200, InitiativeBonus: 30,
					HasCombatPosition: true, CombatX: 6, CombatY: 5,
					SavingThrows: tc.saves}},
				[]combat.Fighter{{ID: "eye", Name: "眼魔", Side: combat.SideEnemy,
					HitPoints: 24, MaxHitPoints: 24,
					HasCombatPosition: true, CombatX: 5, CombatY: 5,
					MonsterAffects: []combat.MonsterAffect{{Kind: 0x7E, Innate: true}}}},
				9,
			); err != nil {
				t.Fatal(err)
			}
			eye, ok := state.fighter("eye")
			if !ok {
				t.Fatal("eye disappeared")
			}
			handled, queued, err := state.monsterSpecialAttack(eye, combat.SideParty)
			if err != nil {
				t.Fatal(err)
			}
			if !handled || queued {
				t.Fatalf("gaze handled=%v queued=%v, want handled without visual", handled, queued)
			}
			message := state.CombatMessage()
			if !strings.Contains(message, "凝視") {
				t.Fatalf("gaze message=%q", message)
			}
			hero, _ := state.fighter("hero")
			if tc.paralyzed {
				if !strings.Contains(message, "麻痺") || !hero.MonsterIsHeld() {
					t.Fatalf("paralysis missing: message=%q held=%v", message, hero.MonsterIsHeld())
				}
			} else {
				if !strings.Contains(message, "抵抗") || hero.MonsterIsHeld() {
					t.Fatalf("resist missing: message=%q held=%v", message, hero.MonsterIsHeld())
				}
			}
		})
	}
}
