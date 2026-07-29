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
	arrow, found := pack.FindCombatVisual("missile", "travel")
	if !found || arrow.ID != "coab.arrow" || arrow.Scale != 2 ||
		arrow.ReferenceDelay != 10 || len(arrow.Frames) != 8 {
		t.Fatalf("arrow combat visual=%+v found=%v", arrow, found)
	}
	west, found := arrow.FrameForDirection(6)
	if !found || west.SourceFile != "COMSPR.DAX" || west.Block != 0x82 || west.FlipX {
		t.Fatalf("west arrow frame=%+v found=%v", west, found)
	}
	magicTravel, found := pack.FindCombatVisual("magic_missile", "travel")
	if !found || magicTravel.ReferenceDelay != 30 || len(magicTravel.Frames) != 4 ||
		magicTravel.Frames[2].Block != 0x85 || !magicTravel.Frames[2].FlipX {
		t.Fatalf("magic travel combat visual=%+v found=%v", magicTravel, found)
	}
	magicImpact, found := pack.FindCombatVisual("magic_missile", "impact")
	if !found || magicImpact.ReferenceDelay != 70 || len(magicImpact.Frames) != 4 ||
		magicImpact.Frames[3].Block != 0x8A || magicImpact.Frames[3].FlipX {
		t.Fatalf("magic impact combat visual=%+v found=%v", magicImpact, found)
	}
	fireballTravel, found := pack.FindCombatVisual("fireball", "travel")
	if !found || fireballTravel.ReferenceDelay != 30 || len(fireballTravel.Frames) != 4 ||
		fireballTravel.Frames[0].Block != 0x05 || fireballTravel.Frames[3].Block != 0x85 {
		t.Fatalf("fireball travel combat visual=%+v found=%v", fireballTravel, found)
	}
	fireballImpact, found := pack.FindCombatVisual("fireball", "impact")
	if !found || fireballImpact.ReferenceDelay != 70 || len(fireballImpact.Frames) != 4 ||
		fireballImpact.Frames[0].Block != 0x0A || fireballImpact.Frames[3].Block != 0x8A {
		t.Fatalf("fireball impact combat visual=%+v found=%v", fireballImpact, found)
	}
	tilverton, found := pack.FindMapByKindLocation("first_person", 2, 1)
	if !found || tilverton.ID != "tilverton.first-person" ||
		tilverton.GeometryFile != "GEO2.DAX" ||
		tilverton.OutdoorSkyColor == nil || *tilverton.OutdoorSkyColor != 3 ||
		tilverton.Spawn == nil || tilverton.Spawn.X != 7 ||
		tilverton.Spawn.Y != 13 || tilverton.Spawn.Direction != 0 {
		t.Fatalf("Tilverton first-person definition=%+v found=%v", tilverton, found)
	}
	firstPerson, found := pack.FindMap(4, 0x20)
	if !found || firstPerson.ID != "zhentil-keep.inner-city" ||
		firstPerson.GeometryFile != "GEO4.DAX" ||
		firstPerson.WallFile != "WALLDEF4.DAX" ||
		firstPerson.SymbolFile != "8X8D4.DAX" ||
		firstPerson.SkyFile != "SKY.DAX" || firstPerson.SkyBlocks != [3]uint8{250, 251, 252} ||
		firstPerson.Spawn == nil || firstPerson.Spawn.X != 2 ||
		firstPerson.Spawn.Y != 0 || firstPerson.Spawn.Direction != 4 {
		t.Fatalf("Zhentil Keep map definition=%+v", pack.Maps)
	}
	shrine, found := pack.FindMap(4, 0x21)
	if !found || shrine.ID != "zhentil-keep.dark-shrine" ||
		shrine.GeometryFile != "GEO4.DAX" || shrine.GeometryBlock != 0x21 ||
		shrine.Spawn == nil || shrine.Spawn.X != 10 || shrine.Spawn.Y != 6 ||
		shrine.Spawn.Direction != 0 {
		t.Fatalf("Zhentil Dark Shrine map definition=%+v found=%v", shrine, found)
	}
	cave, found := pack.FindMapByKindScript("first_person", 4, 0x22)
	if !found || cave.ID != "zhentil-keep.beholder-cave" ||
		cave.GeometryFile != "GEO4.DAX" || cave.ScriptBlock == nil ||
		*cave.ScriptBlock != 0x22 || cave.GeometryBlock != 0x25 ||
		cave.Spawn == nil || cave.Spawn.X != 4 || cave.Spawn.Y != 5 ||
		cave.Spawn.Direction != 0 {
		t.Fatalf("Zhentil beholder cave map definition=%+v found=%v", cave, found)
	}
	overland, found := pack.FindMapByKind("overland")
	if !found || overland.ImageFile != "BIGPIC1.DAX" ||
		overland.GeometryBlock != 0x79 || len(overland.Locations) != 14 ||
		overland.Locations[4].ID != "standing-stone" {
		t.Fatalf("overland map definition=%+v found=%v", overland, found)
	}
	areaMap, found := pack.FindMapByKind("area")
	if !found || areaMap.AreaID != 2 || areaMap.GeometryBlock != 1 ||
		areaMap.GeometryFile != "GEO2.DAX" || areaMap.SymbolFile != "8X8D1.DAX" ||
		areaMap.SymbolBlock != 0xCA || areaMap.Scale != 2 {
		t.Fatalf("AREA map definition=%+v found=%v", areaMap, found)
	}
	result := pack.MatchText([]string{
		"YOU ARE CONFRONTED BY A PATROL FROM ZHENTIL KEEP.",
		"NOTING THE SIGILS ON YOUR ARMS, THEY LET YOU PASS.",
	}, "zh-TW")
	if !result.Matched || !strings.Contains(result.Message, "手臂上的枷印") {
		t.Fatalf("text result=%+v", result)
	}
	dimswart := pack.MatchText([]string{
		"YOU SEE AN OLD MAN IN THE CELL.",
		"HE INTRODUCES HIMSELF AND YOU RECORD HIS REMARKS AS JOURNAL ENTRY 12.",
	}, "zh-TW")
	if !dimswart.Matched || !strings.Contains(dimswart.Message, "牢房裡有一位老人") ||
		len(dimswart.JournalPages) != 6 ||
		!strings.Contains(dimswart.JournalPages[0], "手札條目 12（1/6）") ||
		!strings.Contains(dimswart.JournalPages[5], "摩安德護手") {
		t.Fatalf("Dimswart text result=%+v", dimswart)
	}
	dexam := pack.MatchText([]string{
		"DEXAM SPEAKS.",
		"YOU RECORD HIS SPEECH AS JOURNAL ENTRY 30.",
	}, "zh-TW")
	if !dexam.Matched || !strings.Contains(dexam.Message, "手札第 30 條") ||
		len(dexam.JournalPages) != 2 ||
		!strings.Contains(dexam.JournalPages[1], "兩三個星期") {
		t.Fatalf("Dexam text result=%+v", dexam)
	}
}
