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
