package dungeon

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"

// BashResult records one reference bash_door transaction.
type BashResult struct {
	Attempted      bool
	Opened         bool
	CharacterIndex int
}

// BashDoor reproduces the verified strength branches in ovr015.bash_door.
// roll receives the die size and returns one result in 1..sides. Detail 3 is
// the unpickable-door branch; all other details use the ordinary table. The
// original routine does not check health before trying each character.
//
// ⚠ **這張表以原始位元組為準，不以公開 reference 為準。** `dos overlay-14:0302h`
// 與 `pc98 overlay-14:02F5h` 兩份逐條讀過，兩邊完全相同；reference 的 C# 版在
// 一般門那一欄有兩處與位元組不符（力量 16 與 18/≤50），照抄會讓一支力量 16 的
// 隊伍**永遠撞不開任何一扇門**。詳見 spec 1026。
func BashDoor(roster party.Roster, detail uint8, roll func(sides int) int) BashResult {
	result := BashResult{Attempted: true, CharacterIndex: -1}
	if roll == nil {
		return result
	}
	for index, character := range roster {
		strength := character.Abilities.StrengthFull
		if strength == 0 {
			strength = character.Abilities.Strength
		}
		exceptional := character.Abilities.StrengthExceptional
		worked := false
		if detail == 3 {
			switch {
			case strength == 18 && exceptional >= 91 && exceptional <= 99:
				worked = roll(6) == 1
			case strength == 18 && exceptional == 100:
				worked = roll(6) <= 2
			case strength == 19 || strength == 20:
				worked = roll(6) <= 3
			case strength == 21 || strength == 22:
				worked = roll(6) <= 4
			case strength == 23:
				worked = roll(6) <= 5
			case strength == 24:
				worked = roll(8) <= 7
			case strength == 25:
				worked = true
			}
		} else {
			switch {
			case strength >= 3 && strength <= 7:
				worked = roll(6) == 1
			case strength >= 8 && strength <= 15:
				worked = roll(6) <= 2
			case strength == 16 || strength == 17:
				// ⚠ 這一列與公開 reference（`ovr015.cs`）不同：那邊寫的是
				// `str == 15 || str == 17`，而 15 已經被上一列吃掉，於是
				// **力量 16 一列都不符合、永遠撞不開門**。原始位元組兩平台都寫
				// `cmp ax,10h` ＋ `cmp ax,11h`（16 與 17），見 spec 1026。
				worked = roll(6) <= 3
			case strength == 18 && exceptional <= 50:
				// ⚠ 同上：reference 寫成「直接成功、事後補擲一顆 d6」，原始位元組
				// 是**照擲、d6 ≤ 3 才成功**（`cmp al,3` ＋ `ja`）。
				worked = roll(6) <= 3
			case strength == 18 && exceptional >= 51 && exceptional <= 99:
				worked = roll(6) <= 4
			case strength == 18 && exceptional == 100:
				worked = roll(6) <= 5
			case strength == 19 || strength == 20:
				worked = roll(8) <= 7
			case strength == 21:
				worked = roll(10) <= 9
			case strength == 22 || strength == 23:
				worked = roll(12) <= 11
			case strength == 24:
				worked = roll(20) <= 19
			case strength == 25:
				worked = true
			}
		}
		if worked {
			result.Opened = true
			result.CharacterIndex = index
			return result
		}
	}
	return result
}
