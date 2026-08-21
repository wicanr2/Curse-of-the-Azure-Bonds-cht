package ecl

import "testing"

// `24h COMBAT` 的三選一順序：**場上有怪就直接打**，根本不看商店旗標
// （原作 `overlay-02:179Ah`：`8B69h`／`8B56h` 非 0 就 `jmp sub_1956`，
// `6D8h` 的商店判斷排在它們後面）。
//
// ⚠ 兩個方向都要驗：只驗「有怪就打」的話，把商店那一支整個刪掉也會過，
// 而那會讓商店永遠開不起來。
func TestCombatDispatchChecksMonstersBeforeShop(t *testing.T) {
	// `0Bh LOAD MONSTER` 擺一隻，再 `24h`。
	loadMonster := []byte{0x0B, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01}

	t.Run("有怪就打，不理商店旗標", func(t *testing.T) {
		block := append([]byte{0, 0}, loadMonster...)
		block = append(block, 0x24, 0x00)
		runtime := NewRuntimeState(0)
		// ⚠ `runtime.Memory` 只有在 `Started` 為真時才會被帶進這一次執行。
		runtime.Started = true
		runtime.Memory[addrShopRequest] = 1
		result, err := runSubsetWithState(block, 0, 12, nil, false, 1, runtime)
		if err != nil {
			t.Fatal(err)
		}
		if !result.CombatRequested {
			t.Error("場上有怪卻沒有進戰鬥")
		}
		if result.ShopRequested {
			t.Error("場上有怪時不該走商店")
		}
		if result.ShopPriceScale != 0 {
			t.Error("沒走商店就不該帶出價目倍率")
		}
	})

	t.Run("沒怪才輪到商店", func(t *testing.T) {
		runtime := NewRuntimeState(0)
		runtime.Started = true
		runtime.Memory[addrShopRequest] = 1
		result, err := runSubsetWithState([]byte{0, 0, 0x24, 0x00}, 0, 8, nil, false, 1, runtime)
		if err != nil {
			t.Fatal(err)
		}
		if !result.ShopRequested {
			t.Error("場上沒怪、旗標是 1，應該走商店")
		}
		if result.CombatRequested {
			t.Error("走商店就不該同時進戰鬥")
		}
	})
}
