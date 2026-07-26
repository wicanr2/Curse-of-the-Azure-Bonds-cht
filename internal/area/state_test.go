package area

import "testing"

func TestLoadFilesSelectsDungeonGeoOnlyInDungeon(t *testing.T) {
	values := [3]uint16{0xFF, 0xFF, 0x10}
	outdoor := State{GameArea: 2}
	if effect := outdoor.ApplyLoadFiles(values, 0); effect.GeoMapBlock != nil || effect.BigPicture {
		t.Fatalf("outdoor effect=%+v, want no GEO or picture for sentinel first operand", effect)
	}
	indoor := State{GameArea: 2, InDungeon: true}
	effect := indoor.ApplyLoadFiles(values, 0)
	if effect.GeoMapBlock == nil || *effect.GeoMapBlock != 0x10 || indoor.Current3DMapBlockID != 0x10 {
		t.Fatalf("indoor effect=%+v state=%+v, want GEO block 0x10", effect, indoor)
	}
}

func TestLoadFilesOutdoorBigPictureRule(t *testing.T) {
	state := State{}
	effect := state.ApplyLoadFiles([3]uint16{0x12, 0x34, 0xFF}, 0x40)
	if !effect.BigPicture {
		t.Fatal("non-dungeon LOAD FILES should request big picture when first operand is valid")
	}
	if state.ApplyLoadFiles([3]uint16{0x12, 0x34, 0xFF}, 0x50).BigPicture {
		t.Fatal("last DAX block 0x50 suppresses original big-picture reload")
	}
}
