package gfx

import "testing"

func TestMergePicturesUsesTransparentAndORLayers(t *testing.T) {
	destination := Picture{WidthUnits: 1, HeightUnits: 2, ItemCount: 1, Pixels: []uint8{1, 16, 4, 16, 16, 2, 8, 16}}
	source := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{16, 2, 8, 16, 16, 16, 16, 16}}
	merged, err := MergePictures(destination, source)
	if err != nil {
		t.Fatal(err)
	}
	want := []uint8{1, 2, 12, 16, 16, 2, 8, 16, 16, 2, 8, 16, 16, 16, 16, 16}
	for i, got := range merged.Pixels {
		if got != want[i] {
			t.Fatalf("pixel %d = %d, want %d", i, got, want[i])
		}
	}
}

func TestMergePicturesRejectsWiderSource(t *testing.T) {
	_, err := MergePictures(Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{16, 16, 16, 16, 16, 16, 16, 16}}, Picture{WidthUnits: 2, HeightUnits: 1, ItemCount: 1, Pixels: make([]uint8, 16)})
	if err == nil {
		t.Fatal("expected dimension error")
	}
}

func TestMergePicturesAtAppliesPixelOffset(t *testing.T) {
	destination := Picture{WidthUnits: 1, HeightUnits: 2, ItemCount: 1, Pixels: make([]uint8, 16)}
	for index := range destination.Pixels {
		destination.Pixels[index] = 16
	}
	source := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{1, 2, 3, 4, 5, 6, 7, 8}}
	merged, err := MergePicturesAt(destination, source, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range source.Pixels {
		if got := merged.Pixels[8+index]; got != want {
			t.Fatalf("offset pixel %d=%d, want %d", index, got, want)
		}
	}
}

func TestFlipHorizontalPreservesIndexedPixels(t *testing.T) {
	picture := Picture{WidthUnits: 1, HeightUnits: 1, ItemCount: 1, Pixels: []uint8{1, 2, 3, 4, 5, 6, 7, 8}}
	flipped := picture.FlipHorizontal()
	want := []uint8{8, 7, 6, 5, 4, 3, 2, 1}
	for i, got := range flipped.Pixels {
		if got != want[i] {
			t.Fatalf("pixel %d = %d, want %d", i, got, want[i])
		}
	}
}
