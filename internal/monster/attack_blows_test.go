package monster

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

// 原版六章 `MON*CHA` 的 `BASEATTBLOWS[0]` 分布。
//
// ★ 這條同時是一道**反迴歸**：`+0A1h`（remake 先前拿來當攻擊次數的那一格）在
// 68 種怪物身上**全部是 0**，而且整個 overlay 反組譯裡**沒有任何一處讀它**——
// 它落在 `SPELLSKNOWN`（`+079h`，100 bytes）裡面。照那一格算，每一隻都只打一下
// （spec 1180）。
func TestMonsterRecordsCarryBaseAttackBlows(t *testing.T) {
	records := originalMonsterRecords(t)
	if len(records) == 0 {
		t.Fatal("一筆怪物記錄都沒讀到——正對照失敗，這個測試等於沒跑")
	}
	distribution := map[uint8]int{}
	names := map[uint8][]string{}
	for _, data := range records {
		record, err := Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		if data[0xA1] != 0 {
			t.Fatalf("%s 的 +0A1h ＝ %d：那一格在原作沒有任何讀取端，"+
				"不可以拿來當攻擊次數", record.Name, data[0xA1])
		}
		if seen := names[record.AttackBlows[0]]; !contains(seen, record.Name) {
			distribution[record.AttackBlows[0]]++
			names[record.AttackBlows[0]] = append(seen, record.Name)
		}
	}
	// 半次單位：2 ＝ 一回合一次、4 ＝ 兩次、8 ＝ 四次。
	for _, blows := range []uint8{0, 2, 4, 8} {
		if distribution[blows] == 0 {
			t.Errorf("半次值 %d 一筆都沒有——原版資料四種值都該出現", blows)
		}
	}
	for blows := range distribution {
		switch blows {
		case 0, 2, 4, 8:
		default:
			t.Errorf("出現沒見過的半次值 %d（%v）", blows, names[blows])
		}
	}
	// ⚠ 一次都不能少：`BASEATTBLOWS[0]` 是 0 的兩隻蜘蛛靠槽 1 咬人（spec 1010），
	// 把牠們當成「不會攻擊」是錯的。
	if got := len(names[0]); got != 2 {
		t.Errorf("槽 0 為 0 的有 %d 種（%v），want 2", got, names[0])
	}
	// 一回合不只一次的怪物確實存在——這就是先前整批被壓成一次的那一批。
	multiple := distribution[4] + distribution[8]
	if multiple < 10 {
		t.Errorf("一回合多於一次的怪物只有 %d 種，量測顯示應該有十幾種", multiple)
	}
}

// 投影到 `Fighter` 的路上不能掉。
func TestMonsterFighterCarriesBothBlowSlots(t *testing.T) {
	data := make([]byte, RecordSize)
	data[0x11C] = 8
	data[0x11D] = 2
	record, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	fighter := record.Fighter("beast-1", combat.SideEnemy)
	if fighter.AttackBlows != [2]int{8, 2} {
		t.Fatalf("Fighter.AttackBlows ＝ %v，want {8, 2}", fighter.AttackBlows)
	}
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
