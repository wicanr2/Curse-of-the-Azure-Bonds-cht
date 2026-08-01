package gfx

import "testing"

func TestFirstPersonStageFrameRetainsInsetAndClearsScene(t *testing.T) {
	frame := FirstPersonStageFrame()
	if frame.Bounds().Dx() != 320 || frame.Bounds().Dy() != 200 {
		t.Fatalf("bounds=%v, want 320x200", frame.Bounds())
	}
	for _, point := range [][2]int{{21, 21}, {120, 21}, {21, 120}, {120, 120}} {
		if _, _, _, alpha := frame.At(point[0], point[1]).RGBA(); alpha == 0 {
			t.Fatalf("inset point %v is transparent", point)
		}
	}
	if _, _, _, alpha := frame.At(24, 24).RGBA(); alpha != 0 {
		t.Fatalf("scene interior alpha=%04x, want transparent", alpha)
	}
}
