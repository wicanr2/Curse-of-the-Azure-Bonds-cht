package combat

// 效果掛上效果：有些效果的 handler 自己再呼叫「加效果」把別的碼掛上去。
//
// ★ 這不是法術的性質，是**效果碼的性質**。疾病（法術 40）掛的是效果 `22h`，
// 而 `22h` 的 handler（`overlay-12@09A7h`）連續呼叫兩次 `overlay-23 entry#2`
// 把 `2Bh` 與 `2Ch` 掛上去。所以任何來源（法術、怪物能力、物品）掛上 `22h`
// 時都會連帶掛那兩個，把規則寫在法術那一側會漏掉其他來源。

var effectChainCodes = map[uint8][]uint8{
	// `22h`（疾病）：`overlay-12@09A7h` 連掛 `2Bh` 與 `2Ch`，兩次都用同一個
	// 來源與目標，模式參數原封轉傳。
	0x22: {0x2B, 0x2C},
}

// EffectChainCodes 回傳掛上這個效果碼時會連帶掛上的其他碼。
func EffectChainCodes(kind uint8) []uint8 {
	codes, found := effectChainCodes[kind]
	if !found {
		return nil
	}
	return append([]uint8(nil), codes...)
}
