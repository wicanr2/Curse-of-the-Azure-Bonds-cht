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
