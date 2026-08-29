package main

import (
	"os"
	"testing"

	"golang.org/x/image/font/sfnt"
)

func TestBundledFontCoversShippedLocaleScripts(t *testing.T) {
	cases := map[string]string{
		"../../assets/fonts/NotoSansTC-Regular.ttf":    "繁體中文English青色枷",
		"../../assets/fonts/NotoSansCJKsc-Regular.otf": "简体中文English青色枷",
		"../../assets/fonts/NotoSansCJKjp-Regular.otf": "日本語かなカナEnglishアズール攻略マップ",
	}
	for path, sample := range cases {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		font, err := sfnt.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		var buffer sfnt.Buffer
		for _, character := range sample {
			glyph, err := font.GlyphIndex(&buffer, character)
			if err != nil {
				t.Fatalf("%s glyph lookup %q: %v", path, character, err)
			}
			if glyph == 0 {
				t.Errorf("%s lacks %q (U+%04X)", path, character, character)
			}
		}
	}
}
