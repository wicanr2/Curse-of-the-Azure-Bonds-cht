package gamepack

import (
	"reflect"
	"slices"
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
	if len(pack.MusicTracks) != 12 {
		t.Fatalf("music tracks=%d, want 12", len(pack.MusicTracks))
	}
	for index, track := range pack.MusicTracks {
		if track.ReferenceSelector != uint8(index+1) ||
			track.DriverIndex != uint8(index) || track.TitleID == "" {
			t.Fatalf("music track[%d]=%+v", index, track)
		}
		if pack.Locales["en"][track.TitleID] == "" || pack.Locales["zh-TW"][track.TitleID] == "" {
			t.Fatalf("music track[%d] title %q is not localized", index, track.TitleID)
		}
	}
	if len(pack.CombatModifiers) != 2 {
		t.Fatalf("combat modifiers=%+v", pack.CombatModifiers)
	}
	if enemy, party := pack.CombatModifiers[0], pack.CombatModifiers[1]; enemy.SourceAddress != 0x7F70 || enemy.Side != "enemy" ||
		party.SourceAddress != 0x7F71 || party.Side != "party" {
		t.Fatalf("combat modifiers=%+v", pack.CombatModifiers)
	}
	magicMissile, found := pack.FindCombatAISpell(0x0F)
	if !found || magicMissile.Priority != 4 || magicMissile.CastOn != 1 ||
		magicMissile.MinRange != 0 || magicMissile.CastingTime != 1 {
		t.Fatalf("Magic Missile quick AI metadata=%+v found=%v", magicMissile, found)
	}
	blessAI, found := pack.FindCombatAISpell(0x01)
	if !found || blessAI.Priority != 1 || blessAI.CastOn != 0 ||
		blessAI.MinRange != 0 || blessAI.CastingTime != 10 {
		t.Fatalf("Bless quick AI metadata=%+v found=%v", blessAI, found)
	}
	fireballAI, found := pack.FindCombatAISpell(0x2F)
	if !found || fireballAI.Priority != 7 || fireballAI.CastOn != 1 ||
		fireballAI.MinRange != 3 {
		t.Fatalf("Fireball quick AI metadata=%+v found=%v", fireballAI, found)
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
	lightningTravel, found := pack.FindCombatVisual("lightning_bolt", "travel")
	if !found || len(lightningTravel.Frames) != 4 ||
		lightningTravel.Frames[0].Block != 0x05 || lightningTravel.Frames[3].Block != 0x85 {
		t.Fatalf("lightning travel combat visual=%+v found=%v", lightningTravel, found)
	}
	lightningLine, found := pack.FindCombatVisual("lightning_bolt", "line")
	if !found || len(lightningLine.Frames) != 4 ||
		lightningLine.Frames[0].Block != 0x06 || lightningLine.Frames[3].Block != 0x86 {
		t.Fatalf("lightning line combat visual=%+v found=%v", lightningLine, found)
	}
	lightningImpact, found := pack.FindCombatVisual("lightning_bolt", "impact")
	if !found || lightningImpact.ReferenceDelay != 70 || len(lightningImpact.Frames) != 4 ||
		lightningImpact.Frames[0].Block != 0x0A || lightningImpact.Frames[3].Block != 0x8A {
		t.Fatalf("lightning impact combat visual=%+v found=%v", lightningImpact, found)
	}
	stinkingTravel, found := pack.FindCombatVisual("stinking_cloud", "travel")
	if !found || stinkingTravel.ReferenceDelay != 30 || len(stinkingTravel.Frames) != 4 ||
		stinkingTravel.Frames[0].Block != 0x05 || stinkingTravel.Frames[3].Block != 0x85 {
		t.Fatalf("Stinking Cloud travel combat visual=%+v found=%v", stinkingTravel, found)
	}
	stinkingArea, found := pack.FindCombatVisual("stinking_cloud", "area")
	if !found || len(stinkingArea.Frames) != 1 ||
		stinkingArea.Frames[0].SourceFile != "RANDCOM.DAX" || stinkingArea.Frames[0].Block != 4 {
		t.Fatalf("Stinking Cloud area combat visual=%+v found=%v", stinkingArea, found)
	}
	cloudkillArea, found := pack.FindCombatVisual("cloudkill", "area")
	if !found || len(cloudkillArea.Frames) != 1 ||
		cloudkillArea.Frames[0].SourceFile != "RANDCOM.DAX" || cloudkillArea.Frames[0].Block != 2 {
		t.Fatalf("Cloudkill area combat visual=%+v found=%v", cloudkillArea, found)
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

func TestWizardTowerDracandrosStoryAndJournalAreGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
		want      string
	}{
		{"wizard-tower.courtyard.entering", []string{"HEADING UP INTO THE WIZARD'S TOWER"}, "五層高塔"},
		{"wizard-tower.courtyard.description", []string{"COURTYARD OF A FIVE", "SURROUNDING THE TOWER ARE HIGH MOUNTAINS"}, "五層高塔"},
		{"wizard-tower.dracandros.arrival", []string{"AN IMPRESSIVE ROBED FIGURE APPROACHES YOU", "I AM DRACANDROS"}, "德拉坎德羅斯"},
		{"wizard-tower.dracandros.freezes-party", []string{"FREEZE WHERE YOU STAND", "THE BONDS PARALYZE YOU"}, "動彈不得"},
		{"wizard-tower.dragon-roof", []string{"ROOF OF THE TOWER", "HUGE HOST OF BLACK DRAGONS"}, "黑龍"},
		{"wizard-tower.dragon-steps-out", []string{"ONE OF THE DRAGONS DISENGAGES HIMSELF"}, "走上前"},
		{"wizard-tower.dracandros.attack-order", []string{"ATTACK THE DRAGON AS ELMINSTER TOLD YOU"}, "伊爾明斯特"},
		{"wizard-tower.dragon-illusion", []string{"UNDER THE FORCE OF THE BONDS", "DRAGON WAS ONLY AN ILLUSION"}, "幻象"},
		{"wizard-tower.dracandros.bond-fades", []string{"DRACANDROS' MUMBLED PHRASE", "BONDS TO", "FADE"}, "枷印逐漸消退"},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			result := pack.MatchText(test.fragments, "zh-TW")
			if !result.Matched || result.RuleID != test.id || !strings.Contains(result.Message, test.want) {
				t.Fatalf("result=%+v", result)
			}
		})
	}
	journal := pack.MatchText([]string{
		"FREEZE, BASE SLAYERS OF DRAGONKIND",
		"JOURNAL ENTRY 15",
	}, "zh-TW")
	if !journal.Matched || journal.RuleID != "wizard-tower.dracandros.journal-15" ||
		!strings.Contains(journal.Message, "手札條目 15") ||
		len(journal.JournalPages) != 2 ||
		!strings.HasPrefix(journal.JournalPages[0], "手札條目 15（1/2）") ||
		!strings.HasPrefix(journal.JournalPages[1], "手札條目 15（2/2）") {
		t.Fatalf("journal result=%+v", journal)
	}
	if got, want := sortedLocaleKeys(pack.Locales["zh-TW"]), sortedLocaleKeys(pack.Locales["en"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("zh-TW/en stable ID coverage differs:\nzh-TW=%v\nen=%v", got, want)
	}
}

func sortedLocaleKeys(messages map[string]string) []string {
	keys := make([]string, 0, len(messages))
	for key := range messages {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
