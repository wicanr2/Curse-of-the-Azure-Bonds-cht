package ecl

import (
	"archive/zip"
	"io"
	"testing"

	"github.com/wicanr2/golden-box-remake-engine/dax"
)

// `2Dh CALL 6803h` 是「圖片序列往前推一格」。
//
// 原作 `overlay-02:3035h`：用游標 `DS:722Dh` 取序列 `DS:722Ch` 的第 n 格遠指標，
// 叫 `SHOWPORTRAIT(3, 3, 1, 那一格)`，游標加一、超過張數（`DS:722Ch`）回到 1，
// 最後叫一次 `GAMEDELAY`（與 `3Ah DELAY` 同一支）。
//
// ⚠ 兩個方向都要驗：只驗「`6803h` 會加一」的話，把每一個 CALL 都算成推格同樣
// 會過，而 corpus 裡 `2E10h` 有 125 處、`B200h` 19 處、`C01Eh` 13 處，那會讓
// 每一次重畫與每一聲音效都把圖跳掉一格。
func TestCallAdvancesPictureFrameCursor(t *testing.T) {
	t.Run("6803h 每次推一格", func(t *testing.T) {
		result, err := RunSubset([]byte{0, 0,
			0x2D, 0x02, 0x03, 0x68,
			0x2D, 0x02, 0x03, 0x68,
			0x2D, 0x02, 0x03, 0x68,
			0x00,
		}, 0, 16)
		if err != nil {
			t.Fatal(err)
		}
		if result.PictureFrameAdvances != 3 {
			t.Errorf("推了 %d 格，應該是 3", result.PictureFrameAdvances)
		}
	})

	t.Run("其他 operand 不推格", func(t *testing.T) {
		result, err := RunSubset([]byte{0, 0,
			0x2D, 0x02, 0x10, 0x2E,
			0x2D, 0x02, 0x00, 0xB2,
			0x2D, 0x02, 0x1E, 0xC0,
			0x00,
		}, 0, 16)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.CallAddresses) != 3 {
			t.Fatalf("CallAddresses=%v，應該有三筆", result.CallAddresses)
		}
		if result.PictureFrameAdvances != 0 {
			t.Errorf("推了 %d 格，重畫／音效／前進都不該推格", result.PictureFrameAdvances)
		}
	})

	t.Run("換圖把游標設回第 1 格", func(t *testing.T) {
		// 原作換圖走 `LOADSEQUENCE`，它在 `overlay-29:0270h` 寫 `p^[1] := 1`。
		result, err := RunSubset([]byte{0, 0,
			0x2D, 0x02, 0x03, 0x68,
			0x2D, 0x02, 0x03, 0x68,
			0x0E, 0x00, 0x1D,
			0x2D, 0x02, 0x03, 0x68,
			0x00,
		}, 0, 16)
		if err != nil {
			t.Fatal(err)
		}
		if !result.PictureRequested || result.PictureBlock != 0x1D {
			t.Fatalf("result=%+v，應該要求 PIC 0x1D", result)
		}
		if result.PictureFrameAdvances != 1 {
			t.Errorf("換圖之後推了 %d 格，應該是 1", result.PictureFrameAdvances)
		}
	})
}

// 開場那一段是 corpus 裡唯一用到 `6803h` 的地方：`PICTURE 1Dh` 之後連著十一條
// `CALL 6803h`（`ECL1/0x52` 位移 `01E2h`..`020Ah`，每條四個位元組）。
func TestRealOpeningSequenceAdvancesElevenFrames(t *testing.T) {
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("original image unavailable: %v", err)
	}
	defer archive.Close()
	var eclData []byte
	for _, entry := range archive.File {
		if entry.Name != "ECL1.DAX" {
			continue
		}
		reader, openErr := entry.Open()
		if openErr != nil {
			t.Fatal(openErr)
		}
		eclData, err = io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if eclData == nil {
		t.Skip("ECL1.DAX is absent")
	}
	blocks, err := dax.Parse(eclData)
	if err != nil {
		t.Fatal(err)
	}
	var block dax.Block
	found := false
	for _, candidate := range blocks {
		if candidate.Entry.ID == 0x52 {
			block, found = candidate, true
			break
		}
	}
	if !found {
		t.Fatal("ECL1 block 0x52 is absent")
	}
	result, err := RunSubset(block.Data, 0x14, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !result.PictureRequested || result.PictureBlock != 0x1D {
		t.Fatalf("最後一張圖是 %#02x（requested=%v），應該是 0x1D",
			result.PictureBlock, result.PictureRequested)
	}
	if result.PictureFrameAdvances != 11 {
		t.Fatalf("開場推了 %d 格，應該是 11", result.PictureFrameAdvances)
	}
}
