package combat

import "testing"

// 解除魔法的兩個方向係數不一樣（spec 1125）：高出去每級 ＋5，低下去每級 −2。
// 寫成對稱的 ±5 會讓低階施法者解得太少、高階解得太多，兩邊都不會有人發現。
func TestDispelChanceIsAsymmetric(t *testing.T) {
	for _, item := range []struct {
		caster, effect, want int
	}{
		{5, 5, 50},
		{9, 5, 70},  // 高 4 級 → 50 ＋ 4×5
		{5, 9, 42},  // 低 4 級 → 50 − 4×2
		{1, 20, 12}, // 低 19 級 → 50 − 38
	} {
		if got := DispelChance(item.caster, item.effect); got != item.want {
			t.Fatalf("施法者 %d 對效果 %d 的機率是 %d，want %d",
				item.caster, item.effect, got, item.want)
		}
	}
}

// 天生能力（`MON*SPC` 帶進來的）解不掉——原作用 `EFFECTREC +3 = 0FFh` 標記，
// remake 用 `Innate`。少了這條界線，一次解除魔法可以把怪物的天生能力清光。
func TestDispelMagicKeepsInnateAffects(t *testing.T) {
	battle := newDispelBattle(t)
	result, err := battle.CastDispelMagic("caster", TilePoint{X: 1, Y: 1}, 3, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, impact := range result.Impacts {
		if impact.Kind == 0x47 {
			t.Fatalf("天生效果 %02Xh 被列進對抗名單", impact.Kind)
		}
	}
	target, _ := battle.Fighter("target")
	innate := false
	for _, affect := range target.MonsterAffects {
		if affect.Kind == 0x47 {
			innate = true
		}
	}
	if !innate {
		t.Fatal("天生效果被解掉了")
	}
}

// 20 級對 1 級的效果是 50 ＋ 19×5 ＝ 145%，`1d100` 一定過——所以那個效果一定
// 會被解掉。這條把「機率算式真的接進擲骰」釘住。
func TestDispelMagicRemovesLowLevelAffect(t *testing.T) {
	battle := newDispelBattle(t)
	if _, err := battle.CastDispelMagic("caster", TilePoint{X: 1, Y: 1}, 3, 20); err != nil {
		t.Fatal(err)
	}
	target, _ := battle.Fighter("target")
	for _, affect := range target.MonsterAffects {
		if affect.Kind == 0x35 {
			t.Fatal("1 級施放的效果在 20 級的解除魔法下還活著")
		}
	}
}

func newDispelBattle(t *testing.T) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "caster", Name: "施法者", Side: SideParty, HitPoints: 20, MaxHitPoints: 20,
			ArmorClass: 10, HasCombatPosition: true, CombatX: 1, CombatY: 1,
			SavingThrows: []uint8{14, 14, 14, 14, 14}},
		{ID: "target", Name: "目標", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20,
			ArmorClass: 10, HasCombatPosition: true, CombatX: 2, CombatY: 1,
			SavingThrows: []uint8{14, 14, 14, 14, 14},
			MonsterAffects: []MonsterAffect{
				{Kind: 0x35, Duration: 10, Raw4: 1, Active: true},
				{Kind: 0x47, Duration: 0, Innate: true},
			}},
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 寒冰錐的半徑是 `(等級 ＋ 1) div 2` 最小 1——**不在**法術屬性表裡（表裡是 0）。
func TestConeOfColdRadiusComesFromCasterLevel(t *testing.T) {
	for _, item := range []struct{ level, want int }{
		{1, 1}, {2, 1}, {5, 3}, {9, 5}, {0, 1},
	} {
		if got := ConeOfColdRadius(item.level); got != item.want {
			t.Fatalf("%d 級的寒冰錐半徑是 %d，want %d", item.level, got, item.want)
		}
	}
}

// 火焰護盾兩種形態各掛兩個效果，第二個是共用的。
func TestFireShieldAffectsDifferOnlyInTheFirstCode(t *testing.T) {
	hot, cold := FireShieldAffects(true), FireShieldAffects(false)
	if len(hot) != 2 || len(cold) != 2 {
		t.Fatalf("熱 %v／冷 %v 應該各兩個效果碼", hot, cold)
	}
	if hot[0] == cold[0] {
		t.Fatal("兩種形態的第一個效果碼應該不同")
	}
	if hot[1] != cold[1] {
		t.Fatalf("第二個效果碼應該共用，得到 %02Xh 與 %02Xh", hot[1], cold[1])
	}
}

// 疾病的 `22h` handler 自己連掛 `2Bh` 與 `2Ch`（spec 1125）。這是**效果碼**
// 的性質，不是法術的——寫在法術那一側會漏掉怪物與物品這兩個來源。
func TestEffectChainAppliesTheHandlerSExtraCodes(t *testing.T) {
	chained := EffectChainCodes(0x22)
	if len(chained) != 2 || chained[0] != 0x2B || chained[1] != 0x2C {
		t.Fatalf("效果 22h 連掛的是 %v，want [2B 2C]", chained)
	}
	if EffectChainCodes(0x01) != nil {
		t.Fatal("沒有連掛的效果碼不該回傳東西")
	}
	battle := newDispelBattle(t)
	if _, err := battle.CastEffectSpell("caster", []string{"target"}, EffectSpellRequest{
		SpellID: 40, EffectKind: 0x22, Duration: 5, CasterLevel: 7,
	}); err != nil {
		t.Fatal(err)
	}
	target, _ := battle.Fighter("target")
	found := map[uint8]bool{}
	for _, affect := range target.MonsterAffects {
		found[affect.Kind] = true
	}
	for _, kind := range []uint8{0x22, 0x2B, 0x2C} {
		if !found[kind] {
			t.Fatalf("效果 %02Xh 沒有掛上去", kind)
		}
	}
}

// 效果記錄要帶施法者等級，否則解除魔法只能拿 0 去對抗——那會讓所有效果
// 都變成「最好解的那一種」。
func TestEffectSpellRecordsTheCasterLevel(t *testing.T) {
	battle := newDispelBattle(t)
	if _, err := battle.CastEffectSpell("caster", []string{"target"}, EffectSpellRequest{
		SpellID: 23, EffectKind: 0x34, Duration: 5, CasterLevel: 9,
	}); err != nil {
		t.Fatal(err)
	}
	target, _ := battle.Fighter("target")
	for _, affect := range target.MonsterAffects {
		if affect.Kind == 0x34 {
			if affect.Raw4 != 9 {
				t.Fatalf("效果記的施法者等級是 %d，want 9", affect.Raw4)
			}
			return
		}
	}
	t.Fatal("效果沒有掛上去")
}

// 次元門把施法者移到指定格。
func TestDimensionDoorMovesTheCaster(t *testing.T) {
	battle := newDispelBattle(t)
	if _, err := battle.CastDimensionDoor("caster", TilePoint{X: 6, Y: 4}); err != nil {
		t.Fatal(err)
	}
	caster, _ := battle.Fighter("caster")
	if caster.CombatX != 6 || caster.CombatY != 4 {
		t.Fatalf("施法者在 (%d,%d)，want (6,4)", caster.CombatX, caster.CombatY)
	}
}

// 移除詛咒的第一步：拿掉效果 `24h`，回傳有沒有拿掉。
func TestRemoveCurseAffectReportsWhetherItFired(t *testing.T) {
	battle := newDispelBattle(t)
	removed, err := battle.RemoveCurseAffect("target")
	if err != nil {
		t.Fatal(err)
	}
	if removed {
		t.Fatal("目標身上沒有 24h，不該回報拿掉了")
	}
	if _, err := battle.ApplyEffectCodes("target", []uint8{0x24}, 10); err != nil {
		t.Fatal(err)
	}
	removed, err = battle.RemoveCurseAffect("target")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("目標身上有 24h，應該拿得掉")
	}
}

// 兩種雲就是 spec 1119 那兩種障礙格的來源（spec 1128）：惡臭之雲寫地形碼
// `1Eh`、致命毒雲寫 `1Ch`。
//
// ★ 這條同時守住「規則有生產者」：先前 `ObstacleTerrainBlocks` 有完整規則，
// 但地形碼查詢**從來沒有人掛上**，所以那段規則實際上一次都沒跑過。
func TestCloudsProduceTheObstacleTerrainCodes(t *testing.T) {
	for _, item := range []struct {
		kind PersistentAreaKind
		want uint8
	}{
		{PersistentAreaStinkingCloud, ObstacleTerrainSaveable},
		{PersistentAreaCloudkill, ObstacleTerrainVeteran},
	} {
		code, ok := item.kind.ObstacleTerrainCode()
		if !ok || code != item.want {
			t.Fatalf("區域種類 %d 的地形碼是 %02Xh（ok=%v），want %02Xh",
				item.kind, code, ok, item.want)
		}
	}
	if _, ok := PersistentAreaKind(0).ObstacleTerrainCode(); ok {
		t.Fatal("未知的區域種類不該回報地形碼")
	}
}

// 沒有外部地形碼來源時，戰鬥自己鋪出來的雲要能被查到——而且**不能**因此
// 把整張圖報成出界。
func TestPersistentAreaTerrainLookupDoesNotClaimOutOfBounds(t *testing.T) {
	battle := newDispelBattle(t)
	if code, ok := battle.PersistentAreaTerrainCode(3, 3); ok || code != 0 {
		t.Fatalf("沒有雲的格子回 %02Xh（ok=%v），want 0／false", code, ok)
	}
	result, err := battle.CastStinkingCloud("caster", TilePoint{X: 4, Y: 1}, 5,
		func(x, y int) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Area.Cells) == 0 {
		t.Fatal("惡臭之雲沒有鋪出任何格子")
	}
	cell := result.Area.Cells[0]
	code, ok := battle.PersistentAreaTerrainCode(cell.X, cell.Y)
	if !ok || code != ObstacleTerrainSaveable {
		t.Fatalf("雲格的地形碼是 %02Xh（ok=%v），want %02Xh",
			code, ok, ObstacleTerrainSaveable)
	}
}
