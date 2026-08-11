package gfx

import "testing"

func TestExtendedCharacterCreationFrameKeepsSingleDOSPanel(t *testing.T) {
	frame := ExtendedCharacterCreationFrame()
	if frame.Bounds().Dx() != 320 || frame.Bounds().Dy() != 240 {
		t.Fatalf("bounds=%v, want 320x240", frame.Bounds())
	}
	for _, point := range [][2]int{{0, 0}, {7, 40}, {312, 40}, {160, 176}} {
		if _, _, _, alpha := frame.At(point[0], point[1]).RGBA(); alpha == 0 {
			t.Fatalf("single-panel chrome point %v is transparent", point)
		}
	}
	for _, point := range [][2]int{{160, 20}, {128, 80}, {16, 210}} {
		if _, _, _, alpha := frame.At(point[0], point[1]).RGBA(); alpha != 0 {
			t.Fatalf("single-panel interior point %v alpha=%04x, want transparent", point, alpha)
		}
	}
}
