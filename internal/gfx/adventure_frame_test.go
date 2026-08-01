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

func TestExtendedAdventureFramePreservesTopAndMovesCommandStrip(t *testing.T) {
	source := AdventureFrame()
	extended := ExtendedAdventureFrame()
	if extended.Bounds().Dx() != 320 || extended.Bounds().Dy() != 240 {
		t.Fatalf("bounds=%v, want 320x240", extended.Bounds())
	}
	for _, point := range [][2]int{{0, 0}, {127, 12}, {135, 127}, {319, 183}} {
		gr, gg, gb, ga := extended.At(point[0], point[1]).RGBA()
		wr, wg, wb, wa := source.At(point[0], point[1]).RGBA()
		if gr != wr || gg != wg || gb != wb || ga != wa {
			t.Fatalf("top pixel %v=%04x/%04x/%04x/%04x, want %04x/%04x/%04x/%04x", point, gr, gg, gb, ga, wr, wg, wb, wa)
		}
	}
	for _, point := range [][2]int{{0, 224}, {16, 236}, {319, 239}} {
		gr, gg, gb, ga := extended.At(point[0], point[1]).RGBA()
		wr, wg, wb, wa := source.At(point[0], point[1]-40).RGBA()
		if gr != wr || gg != wg || gb != wb || ga != wa {
			t.Fatalf("shifted command pixel %v=%04x/%04x/%04x/%04x, want %04x/%04x/%04x/%04x", point, gr, gg, gb, ga, wr, wg, wb, wa)
		}
	}
	if _, _, _, alpha := extended.At(16, 210).RGBA(); alpha != 0 {
		t.Fatalf("extended message interior alpha=%04x, want transparent", alpha)
	}
}
