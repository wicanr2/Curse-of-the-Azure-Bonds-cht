package etenfont

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/basicfont"
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
