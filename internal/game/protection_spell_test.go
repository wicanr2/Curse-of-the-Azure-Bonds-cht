package game

import (
	"fmt"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// 法術主表裡「防護邪惡」有兩支：6 是牧師、16 是法師，效果碼同樣是 `08h`；
// 「防護善良」是 7 與 17。這條測試施的是**法師版**。
//
// ★ 為什麼要單獨釘。 這兩支的實作原本寫死法術編號與職業，所以宣告法師版之後
// game pack 看起來已經接好，一施卻報「a different spell target is being selected」。
// 覆蓋報告只看得到 behavior 的分派 case 存在，看不出這種寫死。
func TestMagicUserProtectionSpellsCastLikeTheClericVersions(t *testing.T) {
	for _, item := range []struct {
		name     string
		spellID  uint8
		enemy    combat.Fighter
		affected func(combat.Fighter) bool
	}{
		{
			name:    "防護邪惡（法師版，法術 16）",
			spellID: 16,
			enemy: combat.Fighter{ID: "orc", Name: "獸人", Side: combat.SideEnemy, Evil: true,
				HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 4, CombatY: 4},
			affected: func(f combat.Fighter) bool { return f.ProtectedFromEvil && f.ProtectionEvilRounds == 5 },
		},
		{
			name:    "防護善良（法師版，法術 17）",
			spellID: 17,
			enemy: combat.Fighter{ID: "paladin", Name: "聖騎士", Side: combat.SideEnemy, Good: true,
				HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 4, CombatY: 4},
			affected: func(f combat.Fighter) bool { return f.ProtectedFromGood && f.ProtectionGoodRounds == 5 },
		},
	} {
		t.Run(item.name, func(t *testing.T) {
			state := NewState(testCatalog())
			state.partyRoster = party.Roster{{ID: "mage", Name: "法師",
				Class: party.ClassMagicUser, Level: 2, SpellSlots: []uint8{item.spellID}}}
			partyFighters := []combat.Fighter{{ID: "mage", Name: "法師", Side: combat.SideParty,
				HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20,
				HasCombatPosition: true, CombatX: 1, CombatY: 1}}
			if err := state.StartCombat(partyFighters, []combat.Fighter{item.enemy}, 7); err != nil {
				t.Fatal(err)
			}
			if err := state.BeginCombatCast(item.spellID); err != nil {
				t.Fatal(err)
			}
			if err := state.CombatCast(item.spellID); err != nil {
				t.Fatal(err)
			}
			if len(state.partyRoster[0].SpellSlots) != 0 {
				t.Fatalf("法術位沒被消耗：%#v", state.partyRoster[0].SpellSlots)
			}
			for _, fighter := range state.CombatFighters() {
				if fighter.ID == "mage" && !item.affected(fighter) {
					t.Fatalf("保護沒生效：%+v", fighter)
				}
			}
		})
	}
}

// 牧師版仍然擋法師、法師版仍然擋牧師——參數化不能把職業限制一起弄丟。
func TestProtectionSpellsStillRejectTheWrongClass(t *testing.T) {
	for _, item := range []struct {
		name    string
		spellID uint8
		class   party.Class
	}{
		{"牧師施法師版 16", 16, party.ClassCleric},
		{"牧師施法師版 17", 17, party.ClassCleric},
		{"法師施牧師版 6", ProtectionFromEvilSpellID, party.ClassMagicUser},
		{"法師施牧師版 7", ProtectionFromGoodSpellID, party.ClassMagicUser},
	} {
		t.Run(item.name, func(t *testing.T) {
			state := NewState(testCatalog())
			state.partyRoster = party.Roster{{ID: "caster", Name: "施法者",
				Class: item.class, Level: 2, SpellSlots: []uint8{item.spellID}}}
			partyFighters := []combat.Fighter{{ID: "caster", Name: "施法者", Side: combat.SideParty,
				HitPoints: 10, MaxHitPoints: 10, ArmorClass: 10, InitiativeBonus: 20,
				HasCombatPosition: true, CombatX: 1, CombatY: 1}}
			enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy, Evil: true,
				HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, HasCombatPosition: true, CombatX: 4, CombatY: 4}}
			if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
				t.Fatal(err)
			}
			_ = state.BeginCombatCast(item.spellID)
			if err := state.CombatCast(item.spellID); err == nil {
				t.Fatal("職業不符卻施成功了")
			}
		})
	}
}

// 士氣崩了而且跑得掉的怪物，在自己的回合會印「驚慌逃竄」而不是攻擊。
//
// ★ 這條測試釘的是**接線**，不是規則本身（規則在 `internal/combat` 的單元測試）。
// 少了接線，士氣那一整套會躺在那裡沒有任何呼叫點，而覆蓋報告看不出來。
func TestBrokenMoraleMonsterFleesInsteadOfAttacking(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{{ID: "hero", Name: "英雄",
		Class: party.ClassFighter, Level: 3, HitPoints: 20, MaxHitPoints: 20}}
	partyFighters := []combat.Fighter{{ID: "hero", Name: "英雄", Side: combat.SideParty,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10, MovementAllowance: 2,
		HasCombatPosition: true, CombatX: 1, CombatY: 1}}
	// 士氣 32（90h）、剩一成血 ⇒ 門檻 90，過不了；移動率 12 對上 2 ⇒ 跑得掉。
	enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy,
		HitPoints: 2, MaxHitPoints: 20, ArmorClass: 10, ControlMorale: 0x90,
		MovementAllowance: 12, InitiativeBonus: 20,
		HasCombatPosition: true, CombatX: 2, CombatY: 1}}
	if err := state.StartCombat(partyFighters, enemies, 5); err != nil {
		t.Fatal(err)
	}
	// 測試用的 catalog 只有幾個鍵，查不到就退回鍵名——鍵名不含 `%s`，
	// 所以格式化之後會多出 `%!(EXTRA...)`。這裡直接比對格式化後的結果，
	// 不論 catalog 有沒有翻譯都測得到「有沒有走這條路」。
	want := fmt.Sprintf(state.catalog.Text("combat_flees_in_panic", "combat_flees_in_panic"), "獸人")
	if got := state.CombatMessage(); got != want {
		t.Fatalf("訊息 %q，want %q（士氣崩了的怪物該印驚慌逃竄，不是攻擊）", got, want)
	}
}

// 十呎半徑版套給**整支隊伍**，單目標版只套一個人。
func TestTenFootProtectionCoversTheWholeParty(t *testing.T) {
	state := NewState(testCatalog())
	state.partyRoster = party.Roster{
		{ID: "cleric", Name: "牧師", Class: party.ClassCleric, Level: 4,
			HitPoints: 20, MaxHitPoints: 20, SpellSlots: []uint8{69}},
		{ID: "fighter", Name: "戰士", Class: party.ClassFighter, Level: 4,
			HitPoints: 25, MaxHitPoints: 25},
	}
	partyFighters := []combat.Fighter{
		{ID: "cleric", Name: "牧師", Side: combat.SideParty, HitPoints: 20,
			MaxHitPoints: 20, ArmorClass: 10, InitiativeBonus: 20,
			HasCombatPosition: true, CombatX: 1, CombatY: 1},
		{ID: "fighter", Name: "戰士", Side: combat.SideParty, HitPoints: 25,
			MaxHitPoints: 25, ArmorClass: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 5},
	}
	enemies := []combat.Fighter{{ID: "orc", Name: "獸人", Side: combat.SideEnemy, Evil: true,
		HitPoints: 20, MaxHitPoints: 20, ArmorClass: 10,
		HasCombatPosition: true, CombatX: 9, CombatY: 9}}
	if err := state.StartCombat(partyFighters, enemies, 7); err != nil {
		t.Fatal(err)
	}
	if err := state.BeginCombatCast(69); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatCast(69); err != nil {
		t.Fatal(err)
	}
	protected := 0
	for _, fighter := range state.CombatFighters() {
		if fighter.Side == combat.SideParty && fighter.ProtectedFromEvil {
			protected++
		}
	}
	if protected != 2 {
		t.Fatalf("只有 %d 個人受到防護，十呎半徑該蓋整隊", protected)
	}
	if len(state.partyRoster[0].SpellSlots) != 0 {
		t.Fatalf("法術位沒被消耗：%#v", state.partyRoster[0].SpellSlots)
	}
}

// 笨拙術（`1Bh`）與定身家族共用同一支 handler，所以中了就不能行動。
func TestFumbleSharesTheHoldHandler(t *testing.T) {
	fumbled := combat.Fighter{MonsterAffects: []combat.MonsterAffect{
		{Kind: 0x1B, Value: 3, Duration: 3, Active: true}}}
	if !fumbled.MonsterIsHeld() {
		t.Fatal("笨拙術中了卻還能行動——`1Bh` 與 1Fh／33h／34h／35h 是同一支 handler")
	}
	clean := combat.Fighter{}
	if clean.MonsterIsHeld() {
		t.Fatal("沒有效果卻被當成定身")
	}
}
