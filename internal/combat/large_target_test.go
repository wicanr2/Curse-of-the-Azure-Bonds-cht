package combat

import "testing"

// ★★ 大／小體型在**攻擊結算當下**才換，而且整段包在「攻擊者有槽 0 的武器」裡
// ——天生攻擊不分大小（spec 1175）。
func TestWeaponDamageSwitchesOnTargetSize(t *testing.T) {
	armed := Fighter{
		ID: "a", HasSlotZeroWeapon: true,
		DamageDiceCount: 1, DamageDiceSides: 6, DamageBonus: 1,
		LargeDamageDiceCount: 1, LargeDamageDiceSides: 12, LargeDamageBonus: 3,
	}
	natural := Fighter{
		ID: "n",
		// 沒有槽 0 的武器：兩組都是記錄自己的骰，換也沒有意義。
		DamageDiceCount: 2, DamageDiceSides: 4,
		LargeDamageDiceCount: 9, LargeDamageDiceSides: 9, LargeDamageBonus: 9,
	}
	small := Fighter{ID: "s"}
	large := Fighter{ID: "l", LargeTarget: true}

	for _, probe := range []struct {
		name                string
		attacker, target    Fighter
		count, sides, bonus int
	}{
		{name: "武器打小型", attacker: armed, target: small, count: 1, sides: 6, bonus: 1},
		{name: "武器打大型", attacker: armed, target: large, count: 1, sides: 12, bonus: 3},
		{name: "天生攻擊打大型", attacker: natural, target: large, count: 2, sides: 4, bonus: 0},
	} {
		count, sides, bonus := probe.attacker.WeaponDamageAgainst(probe.target)
		if count != probe.count || sides != probe.sides || bonus != probe.bonus {
			t.Errorf("%s：%dd%d＋%d，want %dd%d＋%d",
				probe.name, count, sides, bonus, probe.count, probe.sides, probe.bonus)
		}
	}
}

// `+0DEh` 的判準：**bit 7 單獨成立**。只看 `and 7` 會把 `81h` 判成小型，
// 而那一格正是眼魔／熊地精／兩種巨蛛（spec 1175）。
func TestLargeTargetFlagIsNotJustTheLowBits(t *testing.T) {
	for raw, want := range map[uint8]bool{
		0x01: false, 0x02: true, 0x03: true,
		0x81: true, 0x82: true, 0x83: true, 0x84: true,
		0x80: false, // 邊界：原作寫的是 `> 80h`
	} {
		if got := largeDamageTargetForTest(raw); got != want {
			t.Errorf("`%02Xh` 判成大型 ＝ %v，want %v", raw, got, want)
		}
	}
}

// largeDamageTargetForTest 是 `monster.LargeDamageTarget` 的獨立謄本。
// 故意不 import `internal/monster`（那會讓 combat 依賴 legacy codec 層），
// 兩份各自出錯的機率才低。
func largeDamageTargetForTest(rawSize uint8) bool {
	return rawSize > 0x80 || rawSize&7 > 1
}
