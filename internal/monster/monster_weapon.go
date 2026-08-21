package monster

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"

// 怪物的隨身武器怎麼進到傷害裡（spec 1174）。
//
// ★ 原作對怪物與隊員走**同一支**派生值重算（`overlay-24:0C28h` → `0025h`）：
// 讀 `+151h[0]`（裝備槽 0 的裝備中物品），有就把類別表的傷害三連**無條件**
// 寫進現值；沒有就整支返回，記錄自己的基準骰原封不動。
//
// ⚠ 記錄的骰與武器的骰不一樣**不是矛盾**：記錄那組是**放下武器時**的天生攻擊。
// 六章 43 隻裡有 13 隻是這個形狀（`BUGBEAR` 記錄 `2d4`、武器 `1d6`）。

// ProjectMonsterWeapon 把怪物身上裝備中的槽 0 武器投影進傷害欄位。
// 沒有那樣的武器時原封不動回傳——那正是原作的行為。
func ProjectMonsterWeapon(fighter combat.Fighter, catalog BaseItemCatalog) combat.Fighter {
	weapon, found := readiedSlotZeroWeapon(fighter.MonsterItems, catalog)
	if !found {
		return fighter
	}
	base, ok := catalog.Lookup(weapon.Type)
	if !ok {
		return fighter
	}
	// ⚠ 沒有「骰數 > 0 才換」這個條件：`sub_25` 一律寫進去。
	fighter.DamageDiceCount = int(base.SmallDamageDice)
	fighter.DamageDiceSides = int(base.SmallDamageSides)
	fighter.DamageBonus = int(base.SmallDamageBonus) + weapon.Plus
	fighter.LargeDamageDiceCount = int(base.LargeDamageDice)
	fighter.LargeDamageDiceSides = int(base.LargeDamageSides)
	fighter.LargeDamageBonus = int(base.LargeDamageBonus) + weapon.Plus
	fighter.HasSlotZeroWeapon = true
	fighter.WeaponRange = int(base.Range)
	fighter.MissileWeapon = base.IsMissileWeapon()
	fighter.ThrownWeapon = base.IsThrownWeapon()
	fighter.AmmunitionType = base.AmmunitionType
	return fighter
}

// readiedSlotZeroWeapon 找出裝備中、類別表的槽是 0 的那一件。
//
// ⚠ 判準是**槽**不是「有沒有傷害骰」：弓與彈藥常常排在鏈的前面，
// 用「第一件有骰的」會被它們搶走（spec 1120 量測時多報過兩隻）。
func readiedSlotZeroWeapon(items []combat.MonsterItem, catalog BaseItemCatalog) (combat.MonsterItem, bool) {
	for _, item := range items {
		if !item.Readied {
			continue
		}
		if base, ok := catalog.Lookup(item.Type); ok && base.Slot == 0 {
			return item, true
		}
	}
	return combat.MonsterItem{}, false
}
