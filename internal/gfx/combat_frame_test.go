package gfx

import "testing"

func TestCombatFrameNativeGeometryAndTransparency(t *testing.T) {
	frame := CombatFrame()
	if frame.Bounds().Dx() != 320 || frame.Bounds().Dy() != 184 {
		t.Fatalf("frame bounds=%v", frame.Bounds())
	}
	for _, point := range [][2]int{{8, 8}, {175, 175}, {184, 8}, {311, 175}} {
		if alpha := frame.RGBAAt(point[0], point[1]).A; alpha != 0 {
			t.Fatalf("panel interior (%d,%d) alpha=%d, want transparent", point[0], point[1], alpha)
		}
	}
	for _, point := range [][2]int{{0, 0}, {319, 0}, {0, 183}, {319, 183}, {180, 80}} {
		if alpha := frame.RGBAAt(point[0], point[1]).A; alpha != 255 {
			t.Fatalf("frame pixel (%d,%d) alpha=%d, want opaque", point[0], point[1], alpha)
		}
	}
}

func TestCombatFrameUsesOracleSampledStonePixels(t *testing.T) {
	frame := CombatFrame()
	oracle := AdventureFrame()
	equal := func(x1, y1, x2, y2 int) bool {
		r1, g1, b1, a1 := frame.At(x1, y1).RGBA()
		r2, g2, b2, a2 := oracle.At(x2, y2).RGBA()
		return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
	}
	for _, point := range [][2]int{{0, 0}, {14, 0}, {319, 7}, {0, 80}, {319, 80}} {
		if !equal(point[0], point[1], point[0], point[1]) {
			t.Fatalf("frame pixel %v does not match oracle", point)
		}
	}
	for y := 0; y < 176; y++ {
		if !equal(176, y, 0, y) {
			t.Fatalf("divider y=%d does not match oracle strip", y)
		}
	}
}
