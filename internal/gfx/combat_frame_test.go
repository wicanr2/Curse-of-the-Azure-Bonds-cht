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

func TestCombatFrameUsesNativeDottedInnerEdgeAndFixedCracks(t *testing.T) {
	frame := CombatFrame()
	if frame.RGBAAt(8, 7) == frame.RGBAAt(9, 7) {
		t.Fatal("top inner edge is not the alternating DOS pixel pattern")
	}
	if frame.RGBAAt(14, 0) != frameBlack || frame.RGBAAt(17, 5) != frameDarkGray {
		t.Fatalf("top crack pixels do not match fixed pattern: start=%v end=%v", frame.RGBAAt(14, 0), frame.RGBAAt(17, 5))
	}
}
