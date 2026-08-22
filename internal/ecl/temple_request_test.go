package ecl

import "testing"

// `24h COMBAT` 的神殿那一支**走得到**：腳本自己會寫 `7EE2h`（spec 1182）。
//
// ★ spec 1095 寫過「這兩格在 1,355 條 ECL 指令裡沒有任何一條寫過（已全掃確認）」
// ——那是**假零**。拿 `cmd/ecl-cell-refs` 掃全 corpus：`7EE2h`（神殿）4 處、
// `7F6Ch`（商店）9 處，而且慣用法一模一樣：
//
//	CLEARMONSTERS          ← 保證第一支（有怪就打）不成立
//	SAVE 01 7EE2           ← 立神殿旗標
//	COMBAT
//
// 這一條直接跑原版資料，把「走得到」變成可執行的斷言。
func TestRealTavernTempleRequestReachesTheTempleBranch(t *testing.T) {
	// `ECL1.DAX` block `0x50` 是世界地圖那一支；`0x0822` 起就是上面那三句。
	block := realBlockData(t, "ECL1.DAX", 0x50)
	runtime := NewRuntimeState(0x0822)
	result, err := runSubsetWithState(block, 0x0822, 8, nil, false, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if !result.TempleRequested {
		t.Fatalf("神殿那一支沒走到：%+v", result)
	}
	if result.CombatRequested || result.PostCombatRequested || result.ShopRequested {
		t.Fatalf("同時走了別支：%+v", result)
	}
	// ⚠ 旗標讀到就清零——exactly-once 靠這個保證，不清的話下一個 `24h` 會再開一次。
	if got := runtime.Memory[addrTempleRequest]; got != 0 {
		t.Fatalf("`7EE2h` 讀完 ＝ %d，want 0（原作 handler 讀到就清零）", got)
	}
}

// 商店那一支同樣走得到，慣用法相同。
func TestRealShopRequestReachesTheShopBranch(t *testing.T) {
	// `ECL4.DAX` block `0x25` 的 `0x009A` 是 `SAVE 01 7F6C`。
	block := realBlockData(t, "ECL4.DAX", 0x25)
	runtime := NewRuntimeState(0x009A)
	result, err := runSubsetWithState(block, 0x009A, 12, nil, false, 1, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ShopRequested {
		t.Fatalf("商店那一支沒走到：%+v", result)
	}
	if got := runtime.Memory[addrShopRequest]; got != 0 {
		t.Fatalf("`7F6Ch` 讀完 ＝ %d，want 0", got)
	}
}
