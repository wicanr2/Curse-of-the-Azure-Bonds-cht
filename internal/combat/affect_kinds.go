package combat

// InterpretedAffectKinds 是 remake **真的看得懂**的效果碼。
//
// ★ 為什麼要有這一張表。 `CastEffectSpell` 可以把任何效果碼記到戰鬥員身上，
// 但「記得上去」不等於「戰鬥規則會理它」。少了這張表，覆蓋報告只數得出
// 「宣告了幾支法術」，而那個數字會把「掛上去就沒下文」的法術算成已完成。
//
// ⚠ 這張表由 `TestInterpretedAffectKindsMatchTheSource` 逐條對回原始碼裡真正
// 比對過的常數——新增判讀卻忘了登記，或登記了卻沒有實作，都會當場紅。
var InterpretedAffectKinds = []uint8{
	0x18, // 偵測隱形：MonsterCanDetectInvisible
	0x19, // 隱形：攻擊修正與被攻擊修正
	0x1F, // 定身家族之一：MonsterIsHeld
	0x25, // 使攻擊必失手：MonsterAffectForcesAttackMiss
	0x27, // 加速：MonsterAffectAttacksPerTurn
	0x2A, // 緩速：MonsterAffectAttacksPerTurn
	0x33, // 定身家族：MonsterIsHeld
	0x34, // 定身家族：MonsterIsHeld
	0x35, // 睡眠／定身：MonsterIsHeld 與睡眠判斷
	0x45, // 對動物隱形
	0x47, // 隱形家族之一
}

// AffectKindIsInterpreted 回答「戰鬥規則會不會理這個碼」。
func AffectKindIsInterpreted(kind uint8) bool {
	for _, known := range InterpretedAffectKinds {
		if known == kind {
			return true
		}
	}
	return false
}
