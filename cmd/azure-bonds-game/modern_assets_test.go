package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModernA6VerticalSlicePNGContracts(t *testing.T) {
	root := filepath.Join("..", "..", "assets", "modern-a6")
	tests := []struct {
		path             string
		width, height    int
		wantTransparency bool
	}{
		{"pictures/character-area-5-head-03-body-03.png", 1024, 1024, false},
		{"pictures/character-area-2-head-04-body-04.png", 1024, 1024, false},
		{"pictures/character-area-2-head-01-body-01.png", 1024, 1024, false},
		{"pictures/character-area-3-head-16-body-16.png", 1024, 1024, false},
		{"pictures/character-area-2-head-02-body-02.png", 1024, 1024, false},
		{"pictures/character-area-2-head-05-body-05.png", 1024, 1024, false},
		{"pictures/character-area-5-head-31-body-31.png", 1024, 1024, false},
		{"pictures/character-area-5-head-33-body-33.png", 1024, 1024, false},
		{"pictures/character-area-5-head-3A-body-3A.png", 1024, 1024, false},
		{"pictures/character-area-3-head-10-body-10.png", 1024, 1024, false},
		{"pictures/character-area-2-head-00-body-00.png", 1024, 1024, false},
		{"pictures/character-area-2-head-41-body-41.png", 1024, 1024, false},
		{"pictures/character-area-3-head-11-body-11.png", 1024, 1024, false},
		{"pictures/character-area-4-head-20-body-20.png", 1024, 1024, false},
		{"pictures/character-area-3-head-12-body-12.png", 1024, 1024, false},
		{"pictures/character-area-3-head-13-body-13.png", 1024, 1024, false},
		{"pictures/character-area-4-head-21-body-21.png", 1024, 1024, false},
		{"pictures/character-area-4-head-22-body-22.png", 1024, 1024, false},
		{"pictures/character-area-2-head-06-body-06.png", 1024, 1024, false},
		{"pictures/character-area-4-head-23-body-23.png", 1024, 1024, false},
		{"pictures/character-area-4-head-2A-body-2A.png", 1024, 1024, false},
		{"pictures/character-area-4-head-31-body-31.png", 1024, 1024, false},
		{"pictures/character-area-4-head-46-body-46.png", 1024, 1024, false},
		{"pictures/character-area-2-head-09-body-06.png", 1024, 1024, false},
		{"pictures/character-area-5-head-32-body-32.png", 1024, 1024, false},
		{"pictures/character-area-5-head-3B-body-3B.png", 1024, 1024, false},
		{"pictures/character-area-6-head-40-body-40.png", 1024, 1024, false},
		{"pictures/bigpic1-block-79-item-00.png", 1216, 480, false},
		{"pictures/bigpic1-block-7B-item-00.png", 1216, 480, false},
		{"pictures/bigpic2-block-78-item-00.png", 1216, 480, false},
		{"pictures/bigpic6-block-7A-item-00.png", 1216, 480, false},
		{"sprites/cpic1-block-01-item-00.png", 48, 48, true},
		{"tiles/tiles-block-01-item-000.png", 96, 96, false},
		{"ui/adventure-frame.png", 640, 480, true},
		{"ui/combat-frame.png", 640, 480, true},
	}
	for _, test := range tests {
		handle, err := os.Open(filepath.Join(root, test.path))
		if err != nil {
			t.Fatalf("open %s: %v", test.path, err)
		}
		decoded, err := png.Decode(handle)
		handle.Close()
		if err != nil {
			t.Fatalf("decode %s: %v", test.path, err)
		}
		bounds := decoded.Bounds()
		if bounds.Dx() != test.width || bounds.Dy() != test.height {
			t.Errorf("%s size=%dx%d, want %dx%d", test.path, bounds.Dx(), bounds.Dy(), test.width, test.height)
		}
		hasTransparency := false
		for y := bounds.Min.Y; y < bounds.Max.Y && !hasTransparency; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, alpha := decoded.At(x, y).RGBA()
				if alpha != 0xffff {
					hasTransparency = true
					break
				}
			}
		}
		if hasTransparency != test.wantTransparency {
			t.Errorf("%s transparency=%v, want %v", test.path, hasTransparency, test.wantTransparency)
		}
	}
}

func TestModernSceneFillsTheGoldFrameOpening(t *testing.T) {
	destination := modernSceneDestination(640, 480)
	want := image.Rect(51, 39, 235, 230)
	if destination != want {
		t.Fatalf("scene destination = %v, want full gold-frame opening %v", destination, want)
	}
	if destination.Dx() != 184 || destination.Dy() != 191 {
		t.Fatalf("scene size = %dx%d, want 184x191", destination.Dx(), destination.Dy())
	}
}

func TestModernA6FramesKeepJointsClosedAndFooterClear(t *testing.T) {
	root := filepath.Join("..", "..", "assets", "modern-a6", "ui")
	decode := func(name string) image.Image {
		handle, err := os.Open(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		defer handle.Close()
		decoded, err := png.Decode(handle)
		if err != nil {
			t.Fatal(err)
		}
		return decoded
	}
	opaque := func(source image.Image, x, y int) bool {
		_, _, _, alpha := source.At(x, y).RGBA()
		return alpha == 0xffff
	}

	adventure := decode("adventure-frame.png")
	if !opaque(adventure, 9, 100) || opaque(adventure, 10, 100) {
		t.Error("adventure outer stone frame must end at 10px")
	}
	for _, point := range [][2]int{{264, 17}, {271, 256}, {17, 256}, {622, 454}, {17, 462}} {
		if !opaque(adventure, point[0], point[1]) {
			t.Errorf("adventure frame joint (%d,%d) is transparent", point[0], point[1])
		}
	}
	combat := decode("combat-frame.png")
	if !opaque(combat, 9, 100) || opaque(combat, 14, 100) {
		t.Error("combat outer stone frame plus inner gold edge exceeds its 10px contract")
	}
	if opaque(combat, 120, 446) {
		t.Error("combat footer y=446 must remain transparent for modern text")
	}
}

func TestModernScenePictureKeyAliasesOnlyByteIdenticalSources(t *testing.T) {
	tests := []struct {
		area       uint8
		head, body byte
		want       string
	}{
		{2, 0x03, 0x03, "character-area-5-head-03-body-03.png"},
		{3, 0x03, 0x03, "character-area-5-head-03-body-03.png"},
		{5, 0x03, 0x03, "character-area-5-head-03-body-03.png"},
		{6, 0x41, 0x41, "character-area-2-head-41-body-41.png"},
		{6, 0x46, 0x46, "character-area-4-head-46-body-46.png"},
		{4, 0x23, 0x23, "character-area-4-head-23-body-23.png"},
	}
	for _, test := range tests {
		if got := modernScenePictureKey(test.area, test.head, test.body); got != test.want {
			t.Errorf("modernScenePictureKey(%d, %02X, %02X)=%q, want %q",
				test.area, test.head, test.body, got, test.want)
		}
	}
}

func TestEverySourceSceneCharacterResolvesToModernA6(t *testing.T) {
	sourceRoot := filepath.Join("..", "..", "assets", "sprites")
	modernRoot := filepath.Join("..", "..", "assets", "modern-a6", "pictures")
	sources, err := filepath.Glob(filepath.Join(sourceRoot, "character-area-*.png"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 31 {
		t.Fatalf("source scene character count=%d, want 31", len(sources))
	}
	for _, source := range sources {
		var area uint8
		var head, body uint8
		if _, err := fmt.Sscanf(filepath.Base(source),
			"character-area-%d-head-%02X-body-%02X.png", &area, &head, &body); err != nil {
			t.Fatalf("parse source scene character %s: %v", source, err)
		}
		key := modernScenePictureKey(area, head, body)
		if _, err := os.Stat(filepath.Join(modernRoot, key)); err != nil {
			t.Errorf("source %s resolves to missing modern A6 picture %s: %v",
				filepath.Base(source), key, err)
		}
	}
}

func TestEveryAreaTileResolvesToModernA6(t *testing.T) {
	root := filepath.Join("..", "..", "assets", "modern-a6", "tiles")
	paths, err := filepath.Glob(filepath.Join(root, "tiles-*.png"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 48 {
		t.Fatalf("modern A6 tile count=%d, want 48", len(paths))
	}
	for index := 0; index < 48; index++ {
		key := modernAreaTileKey(index)
		configFile, err := os.Open(filepath.Join(root, key))
		if err != nil {
			t.Errorf("tile index %d resolves to missing %s: %v", index, key, err)
			continue
		}
		config, err := png.DecodeConfig(configFile)
		configFile.Close()
		if err != nil {
			t.Errorf("decode tile index %d (%s): %v", index, key, err)
			continue
		}
		if config.Width != 96 || config.Height != 96 {
			t.Errorf("tile index %d (%s) size=%dx%d, want 96x96",
				index, key, config.Width, config.Height)
		}
	}
	if got := modernAreaTileKey(-1); got != "" {
		t.Errorf("modernAreaTileKey(-1)=%q, want empty", got)
	}
	if got := modernAreaTileKey(48); got != "" {
		t.Errorf("modernAreaTileKey(48)=%q, want empty", got)
	}
}

func TestEverySourceCPICResolvesToDoubleResolutionModernA6(t *testing.T) {
	sourceRoot := filepath.Join("..", "..", "assets", "sprites")
	modernRoot := filepath.Join("..", "..", "assets", "modern-a6", "sprites")
	sources, err := filepath.Glob(filepath.Join(sourceRoot, "cpic*-item-*.png"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 156 {
		t.Fatalf("source CPIC count=%d, want 156", len(sources))
	}
	for _, source := range sources {
		sourceFile, err := os.Open(source)
		if err != nil {
			t.Fatal(err)
		}
		sourceConfig, err := png.DecodeConfig(sourceFile)
		sourceFile.Close()
		if err != nil {
			t.Fatalf("decode source CPIC %s: %v", source, err)
		}
		modernPath := filepath.Join(modernRoot, filepath.Base(source))
		modernFile, err := os.Open(modernPath)
		if err != nil {
			t.Errorf("source CPIC %s has no modern A6 sprite: %v", filepath.Base(source), err)
			continue
		}
		modernConfig, err := png.DecodeConfig(modernFile)
		modernFile.Close()
		if err != nil {
			t.Errorf("decode modern CPIC %s: %v", modernPath, err)
			continue
		}
		if modernConfig.Width != sourceConfig.Width*2 || modernConfig.Height != sourceConfig.Height*2 {
			t.Errorf("modern CPIC %s size=%dx%d, want source 2x=%dx%d",
				filepath.Base(source), modernConfig.Width, modernConfig.Height,
				sourceConfig.Width*2, sourceConfig.Height*2)
		}
	}
}

func TestEverySourceAnimatedAndPartySpriteResolvesToDoubleResolutionModernA6(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	patterns := []string{
		"pic*-frame-*.png", "sprit*-frame-*.png", "comspr-block-*-item-00.png",
		"chead-block-*-item-00.png", "cbody-block-*-item-00.png", "party*-head-*-body-*.png",
	}
	count := 0
	for _, pattern := range patterns {
		sources, err := filepath.Glob(filepath.Join(root, "sprites", pattern))
		if err != nil {
			t.Fatal(err)
		}
		for _, source := range sources {
			count++
			sourceFile, err := os.Open(source)
			if err != nil {
				t.Fatal(err)
			}
			sourceConfig, err := png.DecodeConfig(sourceFile)
			sourceFile.Close()
			if err != nil {
				t.Fatal(err)
			}
			modernFile, err := os.Open(filepath.Join(root, "modern-a6", "sprites", filepath.Base(source)))
			if err != nil {
				t.Errorf("source %s has no modern A6 sprite: %v", filepath.Base(source), err)
				continue
			}
			modernConfig, err := png.DecodeConfig(modernFile)
			modernFile.Close()
			if err != nil {
				t.Fatal(err)
			}
			paintedStaticPIC := strings.HasPrefix(filepath.Base(source), "pic") && modernConfig.Width == 512 && modernConfig.Height == 512
			if !paintedStaticPIC && (modernConfig.Width != sourceConfig.Width*2 || modernConfig.Height != sourceConfig.Height*2) {
				t.Errorf("modern %s size=%dx%d, want %dx%d", filepath.Base(source),
					modernConfig.Width, modernConfig.Height, sourceConfig.Width*2, sourceConfig.Height*2)
			}
		}
	}
	if count != 512 {
		t.Errorf("covered source count=%d, want 512", count)
	}
}
