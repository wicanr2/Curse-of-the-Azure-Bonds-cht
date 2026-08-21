package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// chargedItem 造一件充能物品。判準是三個欄位（spec 921）：`+3Ch` 充能、
// `+3Dh` 效果、`+3Eh` 最高位沒立起來。**類別不參與判斷**。
func chargedItem(name string, itemType uint8, charges, effect, count uint8) monster.ItemRecord {
	return monster.ItemRecord{
		Name: name, Type: itemType, Count: count,
		Affects: [3]uint8{charges, effect, 0},
	}
}

// useItemState 起一場「一個隊員對上一隻怪」，並讓隊員先動。
func useItemState(t *testing.T, items []monster.ItemRecord, raceType uint8) *State {
	t.Helper()
	state := NewState(trainingTestCatalog(t))
	abilities := party.Abilities{Strength: 12, Intelligence: 10, Wisdom: 12,
		Dexterity: 12, Constitution: 14, Charisma: 10}
	state.partyRoster = party.Roster{
		{ID: "p1", Name: "使者", Race: party.RaceHuman, Class: party.ClassFighter,
			Level: 5, Abilities: abilities, Equipment: items},
	}
	raw := make([]byte, monster.RecordSize)
	raw[0x11A] = raceType
	result := ecl.RunResult{CombatRequested: true,
		MonsterSpawns: []ecl.MonsterSpawn{{MonsterID: 7, Count: 1, IconBlock: 81}}}
	records := map[uint8]monster.Record{
		7: {Name: "SHAMBLING MOUND", HitPoints: 40, MaxHitPoints: 40,
			AttacksPerTurn: 1, RaceType: raceType, Raw: raw},
	}
	user, err := state.partyRoster[0].Fighter()
	if err != nil {
		t.Fatal(err)
	}
	user.HitPoints, user.MaxHitPoints, user.InitiativeBonus = 30, 40, 99
	if err := state.StartEncounter(result, records, []combat.Fighter{user}, 11); err != nil {
		t.Fatal(err)
	}
	if _, ok := state.combatPartyTurn(); !ok {
		t.Skip("這一局的先攻沒有讓隊員先動")
	}
	return &state
}

// ★ 判準是欄位不是類別：粉塵與項鍊的物品類別是 `46h`，用類別去判會整批漏掉。
func TestChargedItemClassificationUsesTheFieldsNotTheType(t *testing.T) {
	catalog := scrollCatalog()
	dust := chargedItem("Dust", 0x46, 1, 0x3F, 0)
	effect, ok := chargedItemEffect(dust, catalog)
	if !ok || effect != 0x3F {
		t.Fatalf("粉塵應該被判成充能物品：effect=%02X ok=%v", effect, ok)
	}
	// ★ `3Ch` 名字叫 `Scroll`，但槽是 `0Ah` ⇒ 原作把它當**充能物品**（spec 1171）。
	if _, ok := chargedItemEffect(
		monster.ItemRecord{Type: 0x3C, Affects: [3]uint8{1, 0x5F, 0}}, catalog); !ok {
		t.Fatal("`3Ch` 應該走充能物品那條路")
	}
	// 真正的卷軸（槽 `0Bh`）走另一條路。
	if _, ok := chargedItemEffect(
		monster.ItemRecord{Type: 0x3D, Affects: [3]uint8{1, 0x0F, 0}}, catalog); ok {
		t.Fatal("卷軸不該被判成充能物品")
	}
	// `+3Eh` 最高位立起來就不是充能物品。
	if _, ok := chargedItemEffect(
		monster.ItemRecord{Type: 0x4F, Affects: [3]uint8{5, 0x41, 0x80}}, catalog); ok {
		t.Fatal("`+3Eh` ≥ 80h 不該被判成充能物品")
	}
}

// 魔法飛彈魔杖：傷害 `2d4＋2`（spec 1169），而且扣掉一次充能。
func TestUseWandOfMagicMissilesDamagesTheAimedEnemy(t *testing.T) {
	state := useItemState(t, []monster.ItemRecord{
		chargedItem("Wand of Magic Missiles", 0x4F, 88, 0x41, 0)}, 0x11)
	before := state.battle.Fighters()
	enemyHP := 0
	enemyID := ""
	for _, fighter := range before {
		if fighter.Side == combat.SideEnemy {
			enemyID, enemyHP = fighter.ID, fighter.HitPoints
		}
	}
	if err := state.CombatUseItem(); err != nil {
		t.Fatalf("使用魔杖：%v", err)
	}
	after, ok := state.fighter(enemyID)
	if !ok {
		t.Fatal("敵人不見了")
	}
	damage := enemyHP - after.HitPoints
	if damage < 4 || damage > 10 {
		t.Fatalf("`2d4＋2` 應該落在 4..10，實際 %d", damage)
	}
	if got := state.partyRoster[0].Equipment[0].Affects[0]; got != 87 {
		t.Fatalf("充能應該從 88 變成 87，實際 %d", got)
	}
}

// ⚠ spec 921 的扣除順序：**數量 > 1 就只扣數量，最後一個才會消耗充能**。
func TestUseItemSpendsTheStackBeforeTheCharge(t *testing.T) {
	state := useItemState(t, []monster.ItemRecord{
		chargedItem("Potion Extra Healing", 0x47, 3, 0x63, 2)}, 0x11)
	if err := state.CombatUseItem(); err != nil {
		t.Fatalf("使用藥水：%v", err)
	}
	item := state.partyRoster[0].Equipment[0]
	if item.Count != 1 || item.Affects[0] != 3 {
		t.Fatalf("第一次該只扣數量：count=%d charges=%d", item.Count, item.Affects[0])
	}
}

// 治療藥水治的是使用者自己（主表 `+6 = 0`、`+7 = 1`）。
func TestUseHealingPotionHealsTheUser(t *testing.T) {
	state := useItemState(t, []monster.ItemRecord{
		chargedItem("Potion", 0x47, 1, 0x63, 0)}, 0x11)
	user, _ := state.fighter("p1")
	before := user.HitPoints
	if err := state.CombatUseItem(); err != nil {
		t.Fatalf("使用藥水：%v", err)
	}
	user, _ = state.fighter("p1")
	healed := user.HitPoints - before
	if healed < 4 || healed > 10 {
		t.Fatalf("`2d4＋2` 應該落在 4..10，實際 %d", healed)
	}
	if len(state.partyRoster[0].Equipment) != 0 {
		t.Fatal("充能歸零的物品應該從背包消失")
	}
}

// ★ 比克魔杖只打 `RACETYPE = 12h`（Moander 的植物眷屬）。
func TestWandOfBeakerOnlyDamagesPlants(t *testing.T) {
	for _, item := range []struct {
		raceType uint8
		hurt     bool
	}{{raceType: 0x12, hurt: true}, {raceType: 0x11, hurt: false}} {
		state := useItemState(t, []monster.ItemRecord{
			chargedItem("Wand of Beaker", 0x4F, 89, 0x62, 0)}, item.raceType)
		// 圓心取使用者自己（`CenterOnUser`），所以要先走進半徑 3 之內。
		if _, err := state.battle.Move("p1", 1, 0); err != nil {
			t.Fatal(err)
		}
		enemyID, before := "", 0
		for _, fighter := range state.battle.Fighters() {
			if fighter.Side == combat.SideEnemy {
				enemyID, before = fighter.ID, fighter.HitPoints
			}
		}
		if err := state.CombatUseItem(); err != nil {
			t.Fatalf("RACETYPE %02Xh：%v", item.raceType, err)
		}
		after, _ := state.fighter(enemyID)
		damaged := after.HitPoints < before
		if damaged != item.hurt {
			t.Fatalf("RACETYPE %02Xh：受傷 ＝ %v，want %v（%d → %d）",
				item.raceType, damaged, item.hurt, before, after.HitPoints)
		}
	}
}

// 速度藥水對已經被緩速的目標什麼都不做，而且**充能不扣**（spec 1169）。
func TestPotionOfSpeedIsBlockedBySlowAndKeepsItsCharge(t *testing.T) {
	state := useItemState(t, []monster.ItemRecord{
		chargedItem("Potion of Speed", 0x47, 1, 0x39, 0)}, 0x11)
	if _, err := state.battle.CastEffectSpell("p1", []string{"p1"}, combat.EffectSpellRequest{
		SpellID: 55, EffectKind: 0x2A, Duration: 9, CasterLevel: 6,
	}); err != nil {
		t.Fatal(err)
	}
	if err := state.CombatUseItem(); err != nil {
		t.Fatalf("使用藥水：%v", err)
	}
	if len(state.partyRoster[0].Equipment) != 1 {
		t.Fatal("被擋下來的那一次不該把物品用掉")
	}
	if got := state.partyRoster[0].Equipment[0].Affects[0]; got != 1 {
		t.Fatalf("充能應該還是 1，實際 %d", got)
	}
}

// 巨人力量藥水把力量提到 21，並掛一筆 `92h` 記錄留住舊值（spec 1169）。
func TestPotionOfGiantStrengthRaisesStrengthToTwentyOne(t *testing.T) {
	state := useItemState(t, []monster.ItemRecord{
		chargedItem("Potion", 0x54, 1, 0x3B, 0)}, 0x11)
	if err := state.CombatUseItem(); err != nil {
		t.Fatalf("使用藥水：%v", err)
	}
	abilities := state.partyRoster[0].Abilities
	if abilities.StrengthFull != 21 {
		t.Fatalf("力量應該變成 21，實際 %d", abilities.StrengthFull)
	}
	found := false
	for _, effect := range state.partyRoster[0].Effects {
		if effect.Kind != giantStrengthAffectKind {
			continue
		}
		found = true
		if effect.Strength != 12 {
			t.Fatalf("舊力量應該存在記錄裡：%d", effect.Strength)
		}
		if effect.Duration < 50 || effect.Duration > 80 {
			t.Fatalf("`(1d4 ＋ 4) × 10` 應該落在 50..80，實際 %d", effect.Duration)
		}
	}
	if !found {
		t.Fatal("沒有掛上 `92h` 的效果記錄")
	}
}

// 卷軸那一條還沒接：回錯誤，而且**不可以**默默把卷軸消耗掉。
func TestUseScrollReportsTheMissingReadingWithoutConsumingIt(t *testing.T) {
	scroll := monster.ItemRecord{Name: "Scroll", Type: 0x3D, Affects: [3]uint8{1, 0x0F, 0}}
	state := useItemState(t, []monster.ItemRecord{scroll}, 0x11)
	if err := state.CombatUseItemAt(0); err == nil {
		t.Fatal("卷軸應該回錯誤")
	}
	if len(state.partyRoster[0].Equipment) != 1 {
		t.Fatal("卷軸不該被消耗")
	}
}
