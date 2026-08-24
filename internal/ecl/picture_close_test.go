package ecl

import "testing"

// `0Eh PICTURE` 的運算元 `0FFh` 是**關閉**，不是「什麼都不做」。
//
// 原作 `overlay-02:0841h` 第一件事就是分兩支（`085Ah` 的 `cmp var_1, 0FFh`）：
// `0FFh` 跳到 `08E9h` 把圖關掉（清 `8B62h`／`8B65h`、重繪、`8B48h`／`8B49h` 歸零），
// 其餘才走開啟。
//
// ⚠ 兩個方向都要驗：只驗「`0FFh` 會設關閉旗標」的話，把開啟那一支也一起設成關閉
// 同樣會過，而那會讓每一張圖都畫不出來。
func TestPictureFFRequestsClose(t *testing.T) {
	t.Run("0FFh 是關閉", func(t *testing.T) {
		result, err := RunSubset([]byte{0, 0, 0x0E, 0x00, 0xFF, 0x00}, 0, 8)
		if err != nil {
			t.Fatal(err)
		}
		if !result.PictureCloseRequested {
			t.Error("`PICTURE 0FFh` 沒有要求關閉")
		}
		if result.PictureRequested {
			t.Error("`PICTURE 0FFh` 不該同時要求開啟")
		}
	})

	t.Run("其他值是開啟", func(t *testing.T) {
		result, err := RunSubset([]byte{0, 0, 0x0E, 0x00, 0x7A, 0x00}, 0, 8)
		if err != nil {
			t.Fatal(err)
		}
		if result.PictureCloseRequested {
			t.Error("`PICTURE 7Ah` 不該要求關閉")
		}
		if !result.PictureRequested || result.PictureBlock != 0x7A {
			t.Errorf("開啟的區塊是 %#02x，應該是 0x7a", result.PictureBlock)
		}
		// `n >= 78h` 是大圖（`087Fh` 的 `cmp var_1, 78h / jb`）。
		if !result.BigPictureRequested {
			t.Error("`7Ah` 應該是大圖")
		}
	})
}

// `0FFh` 關閉那一支的不重繪旁路（spec 1148／1150）：
// `not ((4FBBh = 4) and (4FBAh = 4))`——前後都還在第一人稱就跳過重繪，
// 連 `8B62h`／`8B65h` 的清除一起跳過（「圖還開著」這個狀態留著）。
// `4FBAh`／`4FBBh` 由 ECL 格 `4BE6h` 的**換值**輪替（`overlay-07:0DA0h`）。
func TestPictureCloseRedrawBypass(t *testing.T) {
	// 共用的小工具：組一段「先寫 4BE6、開圖、關圖」的 payload。
	save := func(value byte) []byte {
		return []byte{0x09, 0x00, value, 0x01, 0xE6, 0x4B}
	}
	picture := func(value byte) []byte { return []byte{0x0E, 0x00, value} }
	program := func(parts ...[]byte) []byte {
		block := []byte{0, 0}
		for _, part := range parts {
			block = append(block, part...)
		}
		return append(block, 0x00)
	}
	run := func(t *testing.T, block []byte) (RunResult, *RuntimeState) {
		t.Helper()
		state := NewRuntimeState(0)
		result, err := runSubsetWithState(block, 0, 64, nil, false, 1, state)
		if err != nil {
			t.Fatal(err)
		}
		return result, state
	}

	t.Run("模式換過（0→4）就重繪並清旗標", func(t *testing.T) {
		result, state := run(t, program(save(1), picture(5), picture(0xFF)))
		if !result.PictureCloseRequested || !result.PictureCloseRedraw {
			t.Errorf("close=%v redraw=%v，模式 0→4 的關閉應該立即重繪",
				result.PictureCloseRequested, result.PictureCloseRedraw)
		}
		if state.View.Dirty&(ViewDirtyPicture|ViewDirtyWindow) != 0 {
			t.Errorf("重繪那一支應清 8B62h/8B65h，Dirty=%08b", state.View.Dirty)
		}
	})

	t.Run("前後都在第一人稱就不重繪、旗標留著", func(t *testing.T) {
		// 兩次換值都落在非 0：0→1（Prev=0,Mode=4）再 1→2（Prev=4,Mode=4）。
		result, state := run(t, program(save(1), save(2), picture(5), picture(0xFF)))
		if !result.PictureCloseRequested {
			t.Error("旁路仍是關閉：腳本要求的關閉訊號要留著")
		}
		if result.PictureCloseRedraw {
			t.Error("前後都是第一人稱（4FBB=4FBA=4）不該立即重繪")
		}
		if state.View.Dirty&ViewDirtyPicture == 0 {
			t.Errorf("旁路連 8B62h 的清除一起跳過，Dirty=%08b", state.View.Dirty)
		}
	})

	t.Run("同值重寫不輪替模式", func(t *testing.T) {
		// 兩次都寫 1：第二次不是換值，Prev 停在 0 ⇒ 不是旁路。
		result, _ := run(t, program(save(1), save(1), picture(5), picture(0xFF)))
		if !result.PictureCloseRedraw {
			t.Error("同值重寫不該讓 4FBBh 變 4，關閉仍應重繪")
		}
	})

	t.Run("沒圖開著就不重繪", func(t *testing.T) {
		result, _ := run(t, program(save(1), picture(0xFF)))
		if !result.PictureCloseRequested {
			t.Error("關閉訊號仍要發")
		}
		if result.PictureCloseRedraw {
			t.Error("8B62h/8B65h 都是 0 時原作直接跳過重繪")
		}
	})

	t.Run("先關後開，收尾狀態是開", func(t *testing.T) {
		result, _ := run(t, program(save(1), picture(0xFF), picture(5)))
		if result.PictureCloseRequested {
			t.Error("後面又開了圖，關閉訊號要讓給開圖")
		}
		if !result.PictureRequested || result.PictureBlock != 5 {
			t.Errorf("開圖遺失：requested=%v block=%d",
				result.PictureRequested, result.PictureBlock)
		}
	})
}
