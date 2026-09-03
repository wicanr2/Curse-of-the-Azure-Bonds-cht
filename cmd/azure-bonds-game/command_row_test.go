package main

import (
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// 原版營地主選單的七項（`docs/reference/original-dos/camp/camp-menu.png`）。
var originalCampRow = []string{"SAVE", "VIEW", "MAGIC", "REST", "ALTER", "FIX", "EXIT"}

// 指令列排完之後不准比可用寬度寬——除非它自己標了要縮放。
//
// 這是這條修正的實質：營地選單有七項，縱排時第四項起就落在畫布外面，沒有任何
// 錯誤也沒有任何日誌，只是看不見（原版路徑上有四項是玩家點不到的）。橫排必須
// 自己保證放得下：間隔先壓縮，壓到底仍太寬就整行縮放。
func TestCommandRowLayoutNeverExceedsItsWidth(t *testing.T) {
	face := basicfont.Face7x13
	for _, maxWidth := range []int{40, 80, 120, 200, 320, 592, 1200} {
		layout := commandRowLayout("CAMP:", originalCampRow, face, maxWidth)
		if !layout.scaled && layout.total > maxWidth {
			t.Fatalf("maxWidth=%d：排出 %dpx 卻沒有標要縮放", maxWidth, layout.total)
		}
		if layout.gap < 1 {
			t.Fatalf("maxWidth=%d：間隔壓到 %d，選項會黏成一團", maxWidth, layout.gap)
		}
		if len(layout.widths) != len(originalCampRow) {
			t.Fatalf("maxWidth=%d：量到 %d 個寬度，選項有 %d 個",
				maxWidth, len(layout.widths), len(originalCampRow))
		}
	}
}

// 正對照：remake 底部的可用寬度放得下原版那七項，不該退到縮放。
// 少了這一條，「不超出寬度」可以靠「一律縮放」蒙過去。
func TestCommandRowLayoutFitsTheOriginalCampMenuWithoutScaling(t *testing.T) {
	face := basicfont.Face7x13
	const bottomWidth = 592 // drawOverlandMap 傳進去的可用寬度
	layout := commandRowLayout("CAMP:", originalCampRow, face, bottomWidth)
	if layout.scaled {
		t.Fatalf("原版七項在 %dpx 內就要放得下，卻退到縮放（排出 %dpx）",
			bottomWidth, layout.total)
	}
	// 間隔也不該被壓縮——那表示排版還在勉強。
	if space := font.MeasureString(face, " ").Ceil(); layout.gap != space {
		t.Fatalf("間隔被壓成 %d，原本的空白是 %d", layout.gap, space)
	}
}
