package gamepack

import (
	"strings"
	"testing"
)

func TestEmbeddedPackValidatesAndOwnsZhentilText(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if pack.ID != "curse-of-the-azure-bonds.pit-of-moander" {
		t.Fatalf("pack id=%q", pack.ID)
	}
	firstPerson, found := pack.FindMap(4, 0x20)
	if !found || firstPerson.ID != "zhentil-keep.inner-city" ||
		firstPerson.GeometryFile != "GEO4.DAX" ||
		firstPerson.WallFile != "WALLDEF4.DAX" ||
		firstPerson.SymbolFile != "8X8D4.DAX" ||
		firstPerson.Spawn == nil || firstPerson.Spawn.X != 2 ||
		firstPerson.Spawn.Y != 0 || firstPerson.Spawn.Direction != 4 {
		t.Fatalf("Zhentil Keep map definition=%+v", pack.Maps)
	}
	overland, found := pack.FindMapByKind("overland")
	if !found || overland.ImageFile != "BIGPIC1.DAX" ||
		overland.GeometryBlock != 0x79 || len(overland.Locations) != 14 ||
		overland.Locations[4].ID != "standing-stone" {
		t.Fatalf("overland map definition=%+v found=%v", overland, found)
	}
	areaMap, found := pack.FindMapByKind("area")
	if !found || areaMap.AreaID != 2 || areaMap.GeometryBlock != 1 ||
		areaMap.GeometryFile != "GEO2.DAX" || areaMap.SymbolFile != "8X8D2.DAX" ||
		areaMap.Scale != 2 {
		t.Fatalf("AREA map definition=%+v found=%v", areaMap, found)
	}
	result := pack.MatchText([]string{
		"YOU ARE CONFRONTED BY A PATROL FROM ZHENTIL KEEP.",
		"NOTING THE SIGILS ON YOUR ARMS, THEY LET YOU PASS.",
	}, "zh-TW")
	if !result.Matched || !strings.Contains(result.Message, "手臂上的枷印") {
		t.Fatalf("text result=%+v", result)
	}
}
