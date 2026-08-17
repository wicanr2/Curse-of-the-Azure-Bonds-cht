package combat

// 等級吸取（spec 1127／1129）。
//
// ★★ **這是刻意偏離原作的規則，不是還原。** spec 1127 用位元組層面掃過整份
// 執行檔證明 CoAB **沒有**等級吸取：`+0E7h`／`+0E8h` 只有兩個消費者（訓練升級、
// 復原術）、零個生產者，職業等級陣列連一個 `dec` 都沒有；68 種怪物裡也沒有
// 屍妖／幽魂／幽靈／吸血鬼／暗影那一族。
//
// 使用者 2026-08-17 指定要給龍巫妖這個能力。規則本身沿用原作復原術的資料形狀
// （被吸幾級、連帶少多少 HP），所以**復原術與訓練升級不必改一行**就能把它還回去
// ——原作那兩支消費者現在有東西可消費了。
//
// ⚠ 偏離的部分一律走 game pack 宣告，不寫死在引擎裡：哪一隻怪有這個能力、
// 一次吸幾級，都由 `combat_level_drain_rules` 決定。

// LevelDrainRule 是「帶這個效果碼的怪物打中人時吸幾級」。
type LevelDrainRule struct {
	// EffectKind 是怪物 `MON*SPC` 帶的效果碼。
	EffectKind uint8
	// Levels 是一次吸幾級。
	Levels int
}

// LevelDrainResult 是一次吸取的結果，交給上層套到角色記錄。
type LevelDrainResult struct {
	TargetID string `json:"target_id"`
	Levels   int    `json:"levels"`
}

// SetLevelDrainRules 掛上 game pack 宣告的吸取規則。
func (b *Battle) SetLevelDrainRules(rules []LevelDrainRule) {
	if b == nil {
		return
	}
	b.levelDrainRules = append([]LevelDrainRule(nil), rules...)
}

// LevelDrainOnHit 回答「這個攻擊者打中人時吸幾級」。0 代表不吸。
//
// 只看**天生**效果（`MON*SPC` 帶進來的）：吸等級是怪物的能力，不是誰施在它
// 身上的法術效果。
func (b *Battle) LevelDrainOnHit(attackerID string) int {
	if b == nil || len(b.levelDrainRules) == 0 {
		return 0
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return 0
	}
	total := 0
	for _, affect := range attacker.MonsterAffects {
		if !affect.Innate {
			continue
		}
		for _, rule := range b.levelDrainRules {
			if rule.EffectKind == affect.Kind && rule.Levels > 0 {
				total += rule.Levels
			}
		}
	}
	return total
}
