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
	if len(pack.Maps) != 1 || pack.Maps[0].ID != "zhentil-keep.inner-city" ||
		pack.Maps[0].AreaID != 4 || pack.Maps[0].GeometryBlock != 0x20 ||
		pack.Maps[0].Spawn == nil || pack.Maps[0].Spawn.X != 2 ||
		pack.Maps[0].Spawn.Y != 0 || pack.Maps[0].Spawn.Direction != 4 {
		t.Fatalf("Zhentil Keep map definition=%+v", pack.Maps)
	}
	result := pack.MatchText([]string{
		"YOU ARE CONFRONTED BY A PATROL FROM ZHENTIL KEEP.",
		"NOTING THE SIGILS ON YOUR ARMS, THEY LET YOU PASS.",
	}, "zh-TW")
	if !result.Matched || !strings.Contains(result.Message, "手臂上的枷印") {
		t.Fatalf("text result=%+v", result)
	}
}
