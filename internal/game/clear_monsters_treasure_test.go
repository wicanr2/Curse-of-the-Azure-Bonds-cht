package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// 戰利品堆是**跨執行累積**的（`pendingTreasure`／`pendingTreasureItems` 與寶石、
// 珠寶池），所以 `1Ch` 的清除必須發生在 State 這一側，不只是那一次執行的 result。
//
// ⚠ corpus 唯一走得到「發了又清」那條路的是提爾佛頓火刀首領的重打迴圈
// （`ECL2/0x04`：`27h@0x037b` → `24h` → 判定沒打贏 → `GOTO 0x8359` → `1Ch@0x036c`
// → 重新擺場）。少了這一段，重打一次就多領一次上一輪的戰利品。
func TestClearMonstersDropsAccumulatedTreasurePile(t *testing.T) {
	state := NewState(trainingTestCatalog(t))
	state.pendingTreasure = []ecl.TreasureRequest{{}}
	state.pendingTreasureItems = []monster.ItemRecord{{}}
	state.treasureGems, state.treasureJewelry = 40, 7

	state.applyECLTreasureSignals(ecl.RunResult{ClearMonstersRequested: true})

	if len(state.pendingTreasure) != 0 || len(state.pendingTreasureItems) != 0 {
		t.Errorf("`1Ch` 之後還剩 %d 筆請求、%d 件物品",
			len(state.pendingTreasure), len(state.pendingTreasureItems))
	}
	if state.treasureGems != 0 || state.treasureJewelry != 0 {
		t.Errorf("`1Ch` 之後寶石／珠寶池是 %d／%d", state.treasureGems, state.treasureJewelry)
	}

	// ⚠ 反向：同一次執行裡清完又擺的那一堆必須留著，否則每一場遭遇的戰利品都會
	// 不見——corpus 的慣用法正是「先清再擺」。
	state.applyECLTreasureSignals(ecl.RunResult{
		ClearMonstersRequested: true,
		TreasureRequests:       []ecl.TreasureRequest{{}, {}},
	})
	if len(state.pendingTreasure) != 2 {
		t.Errorf("清完之後擺的那一堆剩 %d 筆，應該是 2 筆", len(state.pendingTreasure))
	}
}
