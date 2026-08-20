package party

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

func dexterityCharacter(t *testing.T, dexterity int) combat.Fighter {
	t.Helper()
	character := Character{
		ID: "pc", Name: "測試", Race: RaceHuman, Class: ClassFighter, Level: 1,
		Abilities: Abilities{Strength: 10, Intelligence: 10, Wisdom: 10,
			Dexterity: dexterity, Constitution: 10, Charisma: 10},
	}
	fighter, err := character.Fighter()
	if err != nil {
		t.Fatal(err)
	}
	return fighter
}

// AC 是畫面刻度：**數字小才難打**。命中判定是
// `d20 ＋ 命中加值 ＋ 目標 AC >= 20`（原作 `d20 ＋ +199h >= 儲存 AC` 換算後的
// 同一式），所以敏捷高的人 AC 低、被打中的次數要**比較少**。
//
// ⚠ 這條擋的是整條刻度反過來的情況：那時候敏捷越高反而越好打，
// 而畫面上的 AC 數字看起來完全正常。
func TestHigherDexterityIsHarderToHit(t *testing.T) {
	hits := func(dexterity int) int {
		defender := dexterityCharacter(t, dexterity)
		defender.ID, defender.Side = "defender", combat.SideEnemy
		defender.HitPoints, defender.MaxHitPoints = 999, 999
		attacker := dexterityCharacter(t, 10)
		attacker.ID, attacker.Side = "attacker", combat.SideParty
		count := 0
		for roll := 2; roll <= 19; roll++ {
			battle, err := combat.NewBattle([]combat.Fighter{attacker, defender}, 1)
			if err != nil {
				t.Fatal(err)
			}
			result, err := battle.ResolveAttack("attacker", "defender", roll, 1)
			if err != nil {
				t.Fatal(err)
			}
			if result.Hit {
				count++
			}
		}
		return count
	}
	clumsy, nimble := hits(6), hits(18)
	if nimble >= clumsy {
		t.Fatalf("敏捷 18 被命中 %d 次、敏捷 6 被命中 %d 次——AC 的刻度反了", nimble, clumsy)
	}
}

// 護甲改善（裝備、法術）讓 AC 變小，變小就是更難打。
func TestArmorClassImprovementMakesTheTargetHarderToHit(t *testing.T) {
	base := dexterityCharacter(t, 10)
	improved := base
	improved.ArmorClass -= 5
	hits := func(defender combat.Fighter) int {
		defender.ID, defender.Side = "defender", combat.SideEnemy
		defender.HitPoints, defender.MaxHitPoints = 999, 999
		attacker := dexterityCharacter(t, 10)
		attacker.ID, attacker.Side = "attacker", combat.SideParty
		count := 0
		for roll := 2; roll <= 19; roll++ {
			battle, err := combat.NewBattle([]combat.Fighter{attacker, defender}, 1)
			if err != nil {
				t.Fatal(err)
			}
			result, err := battle.ResolveAttack("attacker", "defender", roll, 1)
			if err != nil {
				t.Fatal(err)
			}
			if result.Hit {
				count++
			}
		}
		return count
	}
	if hits(improved) >= hits(base) {
		t.Fatalf("AC 改善 5 點之後被命中 %d 次，沒改是 %d 次", hits(improved), hits(base))
	}
}
