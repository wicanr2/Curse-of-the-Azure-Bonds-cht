package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

// canActTiming 是 `CHECKFX(07h)`：「這個人這一回合動得了嗎」。
const canActTiming = "07"

// 原作 `overlay-08 entry#4`（COMBAT 單元的回合開始重設，spec 804）在分派到
// AI 或玩家選單**之前**就呼叫 `CHECKFX(7)`。玩家操作的隊員因此一樣要套到
// 這個時機的記錄寫入——效果 88h 把戰鬥狀態的移動率設成 0（纏繞術那一族）。
//
// ⚠ 這條測試看的是**玩家控制**的隊員：AI 那一側本來就會走到，
// 所以拿敵人來測會過，什麼都證明不了。
func TestCanActRecordWritesApplyToPlayerControlledTurn(t *testing.T) {
	state := NewState(testCatalog())
	party := []combat.Fighter{{
		ID: "pc", Side: combat.SideParty, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, MovementAllowance: 12, InitiativeBonus: 20,
		MonsterAffects: []combat.MonsterAffect{{Kind: 0x88, Active: true}},
	}}
	enemies := []combat.Fighter{{
		ID: "enemy", Side: combat.SideEnemy, HitPoints: 10, MaxHitPoints: 10,
		ArmorClass: 10, MovementAllowance: 12, InitiativeBonus: 1,
	}}
	if err := state.StartCombat(party, enemies, 4711); err != nil {
		t.Fatal(err)
	}
	active, ok := state.CombatActiveFighter()
	if !ok || active.ID != "pc" {
		t.Fatalf("玩家隊員沒有拿到第一個回合：%+v ok=%v", active, ok)
	}
	if active.MovementAllowance != 0 {
		t.Fatalf("效果 88h 應該把移動率設成 0，實際是 %d", active.MovementAllowance)
	}
}

// 回合開頭的記錄寫入在玩家還在選的期間會被重新走到（移動、選目標都會回到
// `advanceCombatToParty`），所以這個時機的修正必須是冪等的。
// 之後若有人在 `07h` 加上 `add`／`sub` 這一類修正，這條會紅，
// 提醒改成「每個回合只套一次」而不是直接放行。
func TestCanActTimingModifiersAreIdempotent(t *testing.T) {
	table, err := gamepack.EffectModifiers()
	if err != nil {
		t.Fatal(err)
	}
	codes, ok := table.Timings[canActTiming]
	if !ok || len(codes) == 0 {
		t.Fatalf("effect-modifiers.json 沒有 timing %s 的效果清單", canActTiming)
	}
	for _, code := range codes {
		key := hexEffectKey(code)
		handler, found := table.Effects[key]
		if !found {
			t.Fatalf("timing %s 列了效果 %s，但表裡沒有這個 handler", canActTiming, key)
		}
		for index, modifier := range handler.Modifiers {
			if modifier.Op != "set" {
				t.Fatalf("效果 %s 的第 %d 個修正是 %q；`CHECKFX(07h)` 的修正會被重複套用，只能是冪等的 `set`",
					key, index, modifier.Op)
			}
		}
	}
}

func hexEffectKey(code int) string {
	const digits = "0123456789ABCDEF"
	return string([]byte{digits[(code>>4)&0x0F], digits[code&0x0F]})
}
