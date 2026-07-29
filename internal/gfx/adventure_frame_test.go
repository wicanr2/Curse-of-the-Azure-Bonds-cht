package gfx

import "testing"

func TestAdventureFrameRetainsChromeAndClearsInteriors(t *testing.T) {
	frame := AdventureFrame()
	if frame.Bounds().Dx() != 320 || frame.Bounds().Dy() != 200 {
		t.Fatalf("bounds=%v, want 320x200", frame.Bounds())
	}
	if _, _, _, alpha := frame.At(0, 0).RGBA(); alpha == 0 {
		t.Fatal("outer chrome is transparent")
	}
	for _, point := range [][2]int{{16, 16}, {140, 16}, {16, 140}, {16, 196}} {
		if _, _, _, alpha := frame.At(point[0], point[1]).RGBA(); alpha != 0 {
			t.Fatalf("interior point %v alpha=%04x, want transparent", point, alpha)
		}
	}
}
