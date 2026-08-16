package main

import "testing"

// 彈窗的兩個純函式：倍率與平移夾限。兩者都不碰 Ebiten，所以能在沒有顯示的
// 情況下釘住行為。
func TestJournalImageScale(t *testing.T) {
	const viewWidth, viewHeight = 576, 336
	tests := []struct {
		name          string
		width, height int
		zoom          bool
		want          float64
	}{
		{"高圖以高度為準", 338, 1000, false, float64(viewHeight) / 1000},
		{"寬圖以寬度為準", 883, 400, false, float64(viewWidth) / 883},
		{"小圖不放大", 216, 194, false, 1},
		{"原尺寸一律 1:1", 883, 1000, true, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := journalImageScale(test.width, test.height, viewWidth, viewHeight, test.zoom)
			if got != test.want {
				t.Fatalf("scale=%v, want %v", got, test.want)
			}
		})
	}
}

// 平移量必須夾在「圖還看得到」的範圍。圖比視窗小的那一軸要歸零，否則玩家會把
// 圖推出畫面外再也找不回來。
func TestClampPan(t *testing.T) {
	tests := []struct {
		name                  string
		offset, content, view int
		want                  int
	}{
		{"圖比視窗小就歸零", 120, 300, 576, 0},
		{"圖剛好一樣大也歸零", -40, 576, 576, 0},
		{"在範圍內不動", 100, 1000, 576, 100},
		{"超過右界夾住", 900, 1000, 576, 212},
		{"超過左界夾住", -900, 1000, 576, -212},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := clampPan(test.offset, test.content, test.view); got != test.want {
				t.Fatalf("clampPan(%d,%d,%d)=%d, want %d",
					test.offset, test.content, test.view, got, test.want)
			}
		})
	}
}
