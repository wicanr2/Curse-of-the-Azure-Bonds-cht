package gfx

import "testing"

func TestCharacterStageFrameKeepsInsetAndTransparentCentre(t *testing.T) {
	frame := CharacterStageFrame()
	if frame.Bounds().Dx() != 320 || frame.Bounds().Dy() != 200 {
		t.Fatalf("bounds=%v", frame.Bounds())
	}
	if _, _, _, alpha := frame.At(21, 16).RGBA(); alpha == 0 {
		t.Fatal("outer inset corner is transparent")
	}
	if _, _, _, alpha := frame.At(60, 60).RGBA(); alpha != 0 {
		t.Fatal("portrait centre must remain transparent")
	}
}
