package dungeon

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

func TestBashDoorUsesDetailThreeStrengthTable(t *testing.T) {
	roster := party.Roster{
		{Abilities: party.Abilities{StrengthFull: 18, StrengthExceptional: 90}},
		{Abilities: party.Abilities{StrengthFull: 23}},
	}
	seenSides := []int{}
	result := BashDoor(roster, 3, func(sides int) int {
		seenSides = append(seenSides, sides)
		if sides == 6 {
			return 5
		}
		return sides
	})
	if !result.Attempted || !result.Opened || result.CharacterIndex != 1 {
		t.Fatalf("result=%#v sides=%v", result, seenSides)
	}
	if len(seenSides) != 1 || seenSides[0] != 6 {
		t.Fatalf("die calls=%v", seenSides)
	}
}

// TestBashDoorStrength18NormalBranchRollsD6 釘住「18／百分位 ≤ 50 要照擲」。
//
// ⚠ 公開 reference 的 C# 版把這一列寫成「直接成功、事後補擲一顆 d6」，原始位元組
// 是 `cmp al,3` ＋ `ja`——**擲出 4 以上就是失敗**（spec 1026）。
func TestBashDoorStrength18NormalBranchRollsD6(t *testing.T) {
	roster := party.Roster{{Abilities: party.Abilities{StrengthFull: 18, StrengthExceptional: 50}}}
	calls := 0
	failed := BashDoor(roster, 2, func(int) int { calls++; return 4 })
	if failed.Opened || calls != 1 {
		t.Fatalf("擲到 4 應該撞不開：result=%#v calls=%d", failed, calls)
	}
	calls = 0
	opened := BashDoor(roster, 2, func(int) int { calls++; return 3 })
	if !opened.Opened || calls != 1 {
		t.Fatalf("擲到 3 應該撞得開：result=%#v calls=%d", opened, calls)
	}
}

// TestBashDoorStrength16IsNotADeadRow 是這一條表最貴的一格。
//
// ★ 為什麼單獨釘住 16。 公開 reference 寫的是 `str == 15 || str == 17`，而 15
// 已經被上一列（8..15）吃掉 ⇒ **力量 16 一列都不符合，永遠撞不開門**。遊戲內建的
// 六個預設角色力量正好是 16，照抄的結果是按鍵玩到第一扇上鎖的門就永遠過不去。
// 原始位元組兩平台都寫 `cmp ax,10h`／`cmp ax,11h`。
func TestBashDoorStrength16IsNotADeadRow(t *testing.T) {
	roster := party.Roster{{Abilities: party.Abilities{StrengthFull: 16}}}
	calls := 0
	opened := BashDoor(roster, 2, func(sides int) int {
		calls++
		if sides != 6 {
			t.Fatalf("力量 16 要擲 d6，實際 d%d", sides)
		}
		return 3
	})
	if !opened.Opened || calls != 1 {
		t.Fatalf("力量 16 擲到 3 應該撞得開：result=%#v calls=%d", opened, calls)
	}
	failed := BashDoor(roster, 2, func(int) int { return 4 })
	if failed.Opened {
		t.Fatalf("力量 16 擲到 4 應該撞不開：result=%#v", failed)
	}
}

// TestBashDoorStrength15UsesTheEightToFifteenRow 把邊界另一側也釘住：15 屬於
// 8..15 那一列（d6 ≤ 2），不是 16–17 那一列。
func TestBashDoorStrength15UsesTheEightToFifteenRow(t *testing.T) {
	roster := party.Roster{{Abilities: party.Abilities{StrengthFull: 15}}}
	if BashDoor(roster, 2, func(int) int { return 3 }).Opened {
		t.Fatal("力量 15 擲到 3 應該撞不開（門檻是 ≤ 2）")
	}
	if !BashDoor(roster, 2, func(int) int { return 2 }).Opened {
		t.Fatal("力量 15 擲到 2 應該撞得開")
	}
}

func TestBashDoorStrength25NeedsNoDie(t *testing.T) {
	calls := 0
	result := BashDoor(party.Roster{{Abilities: party.Abilities{StrengthFull: 25}}}, 2, func(int) int { calls++; return 20 })
	if !result.Opened || calls != 0 {
		t.Fatalf("result=%#v calls=%d", result, calls)
	}
}
