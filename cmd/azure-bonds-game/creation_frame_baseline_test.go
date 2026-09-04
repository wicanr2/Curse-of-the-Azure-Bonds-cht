package main

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/gfx"
)

// 16x15 點陣 CJK 字在基線以下的下伸部；再加一點餘裕，讓字不會貼著框帶的邊。
const (
	cjkGlyphDescent   = 4
	baselineClearance = 2
)

// creationFrameBands 掃出 gfx.ExtendedCharacterCreationFrame() 在 640x480
// （2×）座標下的不透明橫帶。框自己畫在哪、留哪些空白，由框決定，不寫死。
func creationFrameBands(t *testing.T) (bands [][2]int) {
	t.Helper()
	frame := gfx.ExtendedCharacterCreationFrame()
	bounds := frame.Bounds()
	// 取版面中段的水平區間判斷，避開左右兩側的直框邊。
	left, right := bounds.Min.X+bounds.Dx()/2, bounds.Max.X-8
	start := -1
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		opaque := 0
		for x := left; x < right; x++ {
			if _, _, _, alpha := frame.At(x, y).RGBA(); alpha != 0 {
				opaque++
			}
		}
		solid := opaque*10 > (right-left)*6
		switch {
		case solid && start < 0:
			start = y
		case !solid && start >= 0:
			bands = append(bands, [2]int{start * 2, y * 2})
			start = -1
		}
	}
	if start >= 0 {
		bands = append(bands, [2]int{start * 2, bounds.Max.Y * 2})
	}
	return bands
}

// baselineFits 回報 baseline 那一整行字是否完全落在兩條框帶之間的空白裡。
func baselineFits(bands [][2]int, baseline int) bool {
	bottom := baseline + cjkGlyphDescent + baselineClearance
	for _, band := range bands {
		// 字的下緣或基線本身壓進框帶就不算安全。
		if baseline >= band[0] && baseline < band[1] {
			return false
		}
		if bottom > band[0] && bottom <= band[1] {
			return false
		}
		if baseline < band[0] && bottom > band[1] {
			return false
		}
	}
	return bottom <= 480
}

func TestCreationFrameBaselinesStayInsideTheFramesGaps(t *testing.T) {
	bands := creationFrameBands(t)
	if len(bands) < 2 {
		t.Fatalf("框只掃到 %d 條橫帶，量測方式有問題：%v", len(bands), bands)
	}
	for _, tc := range []struct {
		name     string
		baseline int
	}{
		{"建角與角色 VIEW 的底部提示", creationFrameHelpBaseline},
		{"外層選單的隊伍人數", creationFramePartyCountY},
	} {
		if !baselineFits(bands, tc.baseline) {
			t.Errorf("%s 的基線 y=%d 壓到框帶；框帶為 %v", tc.name, tc.baseline, bands)
		}
	}
}

// 正對照：冒險框那條 y=478 的指令列基線畫在建角框上必定沒入底部框帶。
// 這一則失敗就代表上面那則測不出東西。
func TestAdventureCommandBaselineIsUnsafeOnTheCreationFrame(t *testing.T) {
	bands := creationFrameBands(t)
	if baselineFits(bands, adventureCommandBaseline) {
		t.Fatalf("y=%d 應該落在建角框的底部框帶裡，但被判為安全；框帶為 %v",
			adventureCommandBaseline, bands)
	}
}
