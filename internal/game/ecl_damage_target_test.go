package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

// damageTargetAbilities 只是讓角色通得過投影檢查，數值本身與本檔的斷言無關。
var damageTargetAbilities = party.Abilities{
	Strength: 10, StrengthFull: 10, Intelligence: 10, Wisdom: 10,
	Dexterity: 10, Constitution: 10, Charisma: 10,
}

// 三人隊伍，血量分得開，好認出傷害落在誰身上。
func damageTargetState(t *testing.T) State {
	t.Helper()
	state := NewState(testCatalog())
	state.SetECLSeed(1)
	state.partyRoster = party.Roster{
		{ID: "one", Name: "一", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
			HitPoints: 100, MaxHitPoints: 100,
			SavingThrows: []uint8{14, 15, 16, 17, 18}, Abilities: damageTargetAbilities},
		{ID: "two", Name: "二", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
			HitPoints: 100, MaxHitPoints: 100,
			SavingThrows: []uint8{14, 15, 16, 17, 18}, Abilities: damageTargetAbilities},
		{ID: "three", Name: "三", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1,
			HitPoints: 100, MaxHitPoints: 100,
			SavingThrows: []uint8{14, 15, 16, 17, 18}, Abilities: damageTargetAbilities},
	}
	state.party = []combat.Fighter{
		{ID: "one", HitPoints: 100, MaxHitPoints: 100},
		{ID: "two", HitPoints: 100, MaxHitPoints: 100},
		{ID: "three", HitPoints: 100, MaxHitPoints: 100},
	}
	return state
}

func hitPoints(state State) []int {
	points := make([]int, len(state.partyRoster))
	for index, character := range state.partyRoster {
		points[index] = character.HitPoints
	}
	return points
}

// TestAutomaticECLDamageResolvesSelectedCharacterPacket 釘住「目前角色」那一路
// 現在會在正式路徑結算——先前只有全隊那一路會，這 8 處永遠留在 pending。
func TestAutomaticECLDamageResolvesSelectedCharacterPacket(t *testing.T) {
	state := damageTargetState(t)
	state.whoSelectedIndex = 1
	// 旗標 A0h ＝ 單體、不擲豁免；目標旗標 80h ＝ 目前角色（spec 1152）。
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 0xA0, DiceCount: 1, DiceSize: 1, Bonus: 4, SaveFlags: 0x80,
	}}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 1 || outcomes[0].TargetIndex != 1 || outcomes[0].Applied != 5 {
		t.Fatalf("結算結果 ＝ %+v", outcomes)
	}
	if got := hitPoints(state); !reflect.DeepEqual(got, []int{100, 95, 100}) {
		t.Fatalf("血量 ＝ %v，預期只有第二位掉血", got)
	}
	if len(state.ConsumeDamageRequests()) != 0 {
		t.Fatal("目前角色那一路仍留在 pending")
	}
}

// TestAutomaticECLDamageUsesEachPacketsOwnSelectedCharacter 是這一輪的關鍵測試。
// 腳本會把 `LOAD CHARACTER` ＋ `DAMAGE` 包在走過整隊的迴圈裡
// （`ECL5.DAX/0x32:0223h`），一次執行累積好幾組。拿聚合後的「目前角色」去算，
// 整批傷害會落在同一位身上——這個測試就會抓到。
func TestAutomaticECLDamageUsesEachPacketsOwnSelectedCharacter(t *testing.T) {
	state := damageTargetState(t)
	state.whoSelectedIndex = 2
	packet := func(index int, bonus uint16) ecl.DamageRequest {
		return ecl.DamageRequest{
			Flags: 0xA0, DiceCount: 1, DiceSize: 1, Bonus: bonus, SaveFlags: 0x80,
			SelectedPlayerIndex: index, SelectedPlayerSet: true,
		}
	}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{
		packet(0, 9), packet(1, 19),
	}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 2 || outcomes[0].TargetIndex != 0 || outcomes[1].TargetIndex != 1 {
		t.Fatalf("結算結果 ＝ %+v，預期兩個封包各打各的目標", outcomes)
	}
	if got := hitPoints(state); !reflect.DeepEqual(got, []int{90, 80, 100}) {
		t.Fatalf("血量 ＝ %v，預期 [90 80 100]", got)
	}
}

// TestAutomaticECLDamageResolvesRepeatAttackPacket 釘住「旗標 bit 7 清空 ⇒
// 整個 byte 是次數」那一支：不需要選定角色，每一下自己隨機挑人擲命中。
func TestAutomaticECLDamageResolvesRepeatAttackPacket(t *testing.T) {
	state := damageTargetState(t)
	state.whoSelectedIndex = -1
	// 目標旗標 63h 是攻擊值（不是豁免種類），大到必中。
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{{
		Flags: 3, DiceCount: 1, DiceSize: 1, Bonus: 4, SaveFlags: 0x63,
	}}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 3 {
		t.Fatalf("結算結果 ＝ %+v，預期打 3 下", outcomes)
	}
	total := 0
	for _, outcome := range outcomes {
		if !outcome.Hit {
			t.Fatalf("攻擊值 63h 應該必中：%+v", outcomes)
		}
		total += outcome.Applied
	}
	if total != 15 {
		t.Fatalf("三下合計 ＝ %d，預期 15", total)
	}
	if len(state.ConsumeDamageRequests()) != 0 {
		t.Fatal("連打那一路仍留在 pending")
	}
}

// TestAutomaticECLDamageLeavesRandomSingleTargetPending 釘住那條**沒有**接的路。
// 原作裡「單體但隨機挑一名」是活著的程式碼（`overlay-02:2AD6h`），但全 corpus
// 24 處沒有一處走得到，所以留成明確邊界而不是猜一個目標選擇語意。
func TestAutomaticECLDamageLeavesRandomSingleTargetPending(t *testing.T) {
	state := damageTargetState(t)
	state.whoSelectedIndex = 0
	// 旗標 80h ＝ 單體；目標旗標 bit 7 清空 ⇒ 隨機挑一名。
	pending := ecl.DamageRequest{Flags: 0x80, DiceCount: 1, DiceSize: 6, SaveFlags: 0x01}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{pending}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil || outcomes != nil {
		t.Fatalf("結算結果 ＝ %+v err=%v，預期整包留在 pending", outcomes, err)
	}
	if got := hitPoints(state); !reflect.DeepEqual(got, []int{100, 100, 100}) {
		t.Fatalf("血量 ＝ %v，預期沒有人掉血", got)
	}
	if !reflect.DeepEqual(state.pendingDamageRequests, []ecl.DamageRequest{pending}) {
		t.Fatalf("pending ＝ %#v", state.pendingDamageRequests)
	}
}

// TestAutomaticECLDamageLeavesSelectedPacketPendingWithoutACharacter 釘住另一個
// 方向：目前角色那一路要有人被選過才算得出來，沒有就留著，不能拿隊伍第一位頂替。
func TestAutomaticECLDamageLeavesSelectedPacketPendingWithoutACharacter(t *testing.T) {
	state := damageTargetState(t)
	state.whoSelectedIndex = -1
	pending := ecl.DamageRequest{Flags: 0x90, DiceCount: 1, DiceSize: 10, SaveFlags: 0x80}
	state.applyECLDamageSignals(ecl.RunResult{DamageRequests: []ecl.DamageRequest{pending}})
	outcomes, err := state.resolveAutomaticECLDamage()
	if err != nil || outcomes != nil {
		t.Fatalf("結算結果 ＝ %+v err=%v", outcomes, err)
	}
	if got := hitPoints(state); !reflect.DeepEqual(got, []int{100, 100, 100}) {
		t.Fatalf("血量 ＝ %v，預期沒有人掉血", got)
	}
	if !reflect.DeepEqual(state.pendingDamageRequests, []ecl.DamageRequest{pending}) {
		t.Fatalf("pending ＝ %#v", state.pendingDamageRequests)
	}
}
