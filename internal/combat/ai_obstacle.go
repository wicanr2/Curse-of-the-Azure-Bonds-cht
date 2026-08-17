package combat

import "fmt"

// 戰術地圖上的兩種障礙格，以及誰可以照走（spec 1119）。
//
// ★ 為什麼不是「可通行／不可通行」兩態。 原作的地形查詢
// （`overlay-32 entry#19`）把四個腳印格各查一次，回傳三樣東西：擋路的人、
// 地形碼、以及**兩個特別的地形碼有沒有出現**。`1Eh` 與 `1Ch` 不進成本表，
// 它們走的是另一條路——**帶對效果的人可以照走**，其他人這一步就沒了。
// 少了這一層，飛行／自由行動之類的效果在移動上完全沒有差別。
//
// ⚠ 效果碼本身的語意沒有解讀。這裡照碼比對，不替它們取名字：
// 拿法術表的 `+0Ah` 反查會得到「臭雲術」「死者復生」這種對不上的名字，
// 那代表**效果碼不是法術名**（一個效果可以被多支法術套上），不是對照表壞了。

const (
	// ObstacleTerrainSaveable 是「豁免得過就照走」的那一種（原作地形碼 `1Eh`）。
	ObstacleTerrainSaveable uint8 = 0x1E
	// ObstacleTerrainVeteran 是「夠強就照走」的那一種（原作地形碼 `1Ch`）。
	ObstacleTerrainVeteran uint8 = 0x1C
)

// obstacleSaveBypassEffects 是 `1Eh` 的豁免效果清單（overlay-09 sub_735）。
// 順序照原作，因為原作是短路求值——換順序不影響結果，但對照時看得出來。
var obstacleSaveBypassEffects = []uint8{0x20, 0x1E, 0x6F, 0x7D, 0x81, 0x3F}

// obstacleVeteranBypassEffects 是 `1Ch` 的豁免效果清單。與上面**只有部分重疊**：
// `20h`／`3Fh`／`1Eh` 不在裡面，`85h` 只在這裡。
var obstacleVeteranBypassEffects = []uint8{0x81, 0x6F, 0x85, 0x7D}

// obstacleVeteranMinLevel 是 `1Ch` 的等級門檻：`有號(+0E5h) ≥ 7` 就照走。
// **有號**——`0E5h` 這一格對怪物是 HD，對角色是最高職業等級，兩邊共用一個位元組。
const obstacleVeteranMinLevel = 7

// ObstacleTerrainBlocks 回答「這個人走不走得進這一格」。
//
// `combatStateWaiver` 是原作戰鬥狀態記錄的 `+10h ≠ 0`（語意未解讀，兩種障礙
// 都吃這一條）。`save` 只在 `1Eh` 用得到，而且**只有前面每一條都不成立時才會被
// 呼叫**——原作是短路求值，多擲一次骰會讓亂數序列偏掉。
func ObstacleTerrainBlocks(fighter Fighter, code uint8, combatStateWaiver bool, save func() bool) bool {
	switch code {
	case ObstacleTerrainSaveable:
		if fighterHasAnyAffect(fighter, obstacleSaveBypassEffects) || combatStateWaiver {
			return false
		}
		return save == nil || !save()
	case ObstacleTerrainVeteran:
		if int8(fighter.HitDice) >= obstacleVeteranMinLevel {
			return false
		}
		if fighterHasAnyAffect(fighter, obstacleVeteranBypassEffects) || combatStateWaiver {
			return false
		}
		return true
	default:
		return false
	}
}

// fighterHasAnyAffect 重現 `overlay-24 entry#27`：走一遍效果串列，比對每一節的
// 第一個位元組。**不看持續時間也不看 Active**——原作那一支只比對碼本身，
// 到期的節點在到期時就被摘掉了。
func fighterHasAnyAffect(fighter Fighter, kinds []uint8) bool {
	for _, affect := range fighter.MonsterAffects {
		for _, kind := range kinds {
			if affect.Kind == kind {
				return true
			}
		}
	}
	return false
}

// SavingThrowResult 是一次豁免的結果，攤開是為了讓測試能分辨「自動成功」與
// 「加總剛好過門檻」。
type SavingThrowResult struct {
	Roll  int
	Total int
	Saved bool
	// Natural 記錄這一擲是不是天然 1／20——那兩個點數不看門檻。
	Natural bool
}

// RollSavingThrow 重現 `overlay-23 entry#8`：
//
//	1d20 ＝ 1  → 一定失敗
//	1d20 ＝ 20 → 一定成功
//	其餘        → 點數 ＋ 角色豁免加值 ＋ 情境修正 ≥ 門檻 才算過
//
// ⚠ 門檻為 0 時原作會「一定成功」（`0 ≤ 任何總和`）。remake 這裡改判失敗：
// 門檻 0 代表**這筆記錄沒有豁免資料**，而沒有資料的怪物該是弱點不是免疫。
// 這是刻意與原作不同的一處，因為原作的資料保證填滿而 remake 的不保證。
func (b *Battle) RollSavingThrow(target Fighter, category, modifier int) (SavingThrowResult, error) {
	if b == nil || b.rng == nil {
		return SavingThrowResult{}, errNoPRNG
	}
	if category < 0 {
		return SavingThrowResult{}, errSaveCategory(category)
	}
	result := SavingThrowResult{Roll: b.rng.Intn(20) + 1}
	switch result.Roll {
	case 1:
		result.Natural = true
		return result, nil
	case 20:
		result.Natural, result.Saved = true, true
		return result, nil
	}
	result.Total = result.Roll + target.SavingThrowBonus + modifier
	// 表比類別短時當成 0 ＝ 沒有資料 ＝ 失敗，理由同上。**骰還是要擲**，
	// 少擲一次會讓後面每一擲都偏掉。
	threshold := 0
	if category < len(target.SavingThrows) {
		threshold = int(target.SavingThrows[category])
	}
	result.Saved = threshold != 0 && result.Total >= threshold
	return result, nil
}

var errNoPRNG = fmt.Errorf("battle PRNG is unavailable")

func errSaveCategory(category int) error {
	return fmt.Errorf("saving throw category %d is outside this fighter's table", category)
}
