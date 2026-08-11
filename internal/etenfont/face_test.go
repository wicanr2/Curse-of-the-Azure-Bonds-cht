package etenfont

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func TestLoadRejectsTruncatedFont(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stdfont.15")
	if err := os.WriteFile(path, []byte{1}, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, "", basicfont.Face7x13, true); err == nil {
		t.Fatal("Load accepted a truncated font")
	}
}

func TestBoldExpandsGlyphRight(t *testing.T) {
	data := make([]byte, 5402*glyphBytes)
	data[0] = 0x80 // Big5 A440, 「一」: test pixel at x=0.
	path := filepath.Join(t.TempDir(), "stdfont.15")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	face, err := Load(path, "", basicfont.Face7x13, true)
	if err != nil {
		t.Fatal(err)
	}
	mask, ok := face.bitmap('一')
	if !ok || mask.AlphaAt(0, 0).A == 0 || mask.AlphaAt(1, 0).A == 0 {
		t.Fatal("bold face did not expand source pixel one column right")
	}
}

func TestASCIICompanionRendersHalfWidthAndFullWidthPunctuation(t *testing.T) {
	directory := t.TempDir()
	standardPath := filepath.Join(directory, "stdfont.15")
	if err := os.WriteFile(standardPath, make([]byte, 5402*glyphBytes), 0o600); err != nil {
		t.Fatal(err)
	}
	ascii := make([]byte, 256*asciiGlyphBytes)
	ascii[int(':')*asciiGlyphBytes] = 0x80
	asciiPath := filepath.Join(directory, "ascfont.15")
	if err := os.WriteFile(asciiPath, ascii, 0o600); err != nil {
		t.Fatal(err)
	}
	face, err := LoadWithASCII(standardPath, "", asciiPath, basicfont.Face7x13, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []rune{':', '：'} {
		advance, ok := face.GlyphAdvance(r)
		if !ok || advance != fixed.I(asciiGlyphWidth) {
			t.Fatalf("glyph %q advance=%v ok=%v, want %v", r, advance, ok, fixed.I(asciiGlyphWidth))
		}
		mask, ok := face.bitmap(r)
		if !ok || mask.AlphaAt(0, 0).A == 0 {
			t.Fatalf("glyph %q did not use the companion ASCII raster", r)
		}
	}
}
