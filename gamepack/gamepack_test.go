package gamepack

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	enginedamage "github.com/wicanr2/golden-box-remake-engine/combat/damage"
	enginemodifier "github.com/wicanr2/golden-box-remake-engine/combat/modifier"
	enginespell "github.com/wicanr2/golden-box-remake-engine/combat/monsterspell"
	engineposthit "github.com/wicanr2/golden-box-remake-engine/combat/posthit"
	engineresistance "github.com/wicanr2/golden-box-remake-engine/combat/resistance"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

func TestPackDeclaresEveryOriginalGEOBlock(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer archive.Close()

	declared := make(map[[2]uint8]bool)
	for _, definition := range pack.Maps {
		if definition.Kind != "first_person" {
			continue
		}
		declared[[2]uint8{definition.AreaID, definition.GeometryBlock}] = true
	}
	for index, member := range []string{"GEO2.DAX", "GEO3.DAX", "GEO4.DAX", "GEO5.DAX", "GEO6.DAX"} {
		data := originalMapMember(t, archive, member)
		blocks, err := dax.Parse(data)
		if err != nil {
			t.Fatalf("%s: %v", member, err)
		}
		set := uint8(index + 2)
		for _, block := range blocks {
			if !declared[[2]uint8{set, block.Entry.ID}] {
				t.Fatalf("GEO%d block 0x%02X has no first-person game-pack declaration", set, block.Entry.ID)
			}
		}
	}
	if len(declared) < 16 {
		t.Fatalf("declared first-person geometry blocks=%d, want at least 16 original blocks", len(declared))
	}
}

func originalMapMember(t *testing.T, archive *zip.ReadCloser, name string) []byte {
	t.Helper()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read %s: read=%v close=%v", name, readErr, closeErr)
		}
		return data
	}
	t.Fatalf("archive member %q not found", name)
	return nil
}

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
	if len(pack.CombatAffectRules) != 3 {
		t.Fatalf("combat affect rules=%+v", pack.CombatAffectRules)
	}
	if len(pack.CombatConditionalModifiers) != 2 {
		t.Fatalf("combat conditional modifiers=%+v", pack.CombatConditionalModifiers)
	}
	if len(pack.CombatMagicResistanceRules) != 1 {
		t.Fatalf("combat magic resistance rules=%+v", pack.CombatMagicResistanceRules)
	}
	magicResistanceRules, err := pack.ResolveCombatMagicResistanceRules()
	if err != nil {
		t.Fatal(err)
	}
	wantMagicResistanceRules := []engineresistance.Rule{{
		ID: "coab.monster_affect_6a.magic-resistance-15", EffectKind: 0x6A,
		Formula: engineresistance.FormulaLevelAdjustedD100, Base: 15,
	}}
	if !reflect.DeepEqual(magicResistanceRules, wantMagicResistanceRules) {
		t.Fatalf("combat magic resistance rules=%+v want=%+v", magicResistanceRules, wantMagicResistanceRules)
	}
	if len(pack.CombatPostHitRules) != 1 {
		t.Fatalf("combat post-hit rules=%+v", pack.CombatPostHitRules)
	}
	postHitRules, err := pack.ResolveCombatPostHitRules()
	if err != nil {
		t.Fatal(err)
	}
	wantPostHitRules := []engineposthit.Rule{{
		ID: "coab.monster_affect_4f.fire-magic-2d10-slots-1-2", EffectKind: 0x4F,
		MinAttackSlot: 1, MaxAttackSlot: 2, DamageDiceCount: 2, DamageDiceSides: 10,
		DamageMask: 0x09,
	}}
	if !reflect.DeepEqual(postHitRules, wantPostHitRules) {
		t.Fatalf("combat post-hit rules=%+v want=%+v", postHitRules, wantPostHitRules)
	}
	if len(pack.CombatMonsterSpellRules) != 1 {
		t.Fatalf("combat monster spell rules=%+v", pack.CombatMonsterSpellRules)
	}
	monsterSpellRules, err := pack.ResolveCombatMonsterSpellRules()
	if err != nil {
		t.Fatal(err)
	}
	wantMonsterSpellRules := []enginespell.Rule{{
		ID: "coab.monster_affect_84.lightning-bolt-rounds-1-3", EffectKind: 0x84, SpellID: 0x33,
		MaxRound: 3, CasterLevel: 1, TargetRange: 10, LineBudget: 10,
		InitialDamageDice: 16, PathDamageDice: 16, DamageDiceSides: 6, DamageMask: 0x0C,
		FirstReflectionOriginThreshold: 8, FirstReflectionPenalty: 8,
	}}
	if !reflect.DeepEqual(monsterSpellRules, wantMonsterSpellRules) {
		t.Fatalf("combat monster spell rules=%+v want=%+v", monsterSpellRules, wantMonsterSpellRules)
	}
	conditionalRules, err := pack.ResolveCombatConditionalModifiers()
	if err != nil {
		t.Fatal(err)
	}
	wantConditionalRules := []enginemodifier.Rule{
		{ID: "coab.monster_affect_08.protection-from-good", EffectKind: 0x08, Predicate: enginemodifier.PredicateValueIn, Values: []uint8{2, 5, 8}, AttackRollDelta: -2, SavingThrowDelta: 2},
		{ID: "coab.monster_affect_09.protection-from-evil", EffectKind: 0x09, Predicate: enginemodifier.PredicateValueIn, Values: []uint8{0, 3, 6}, AttackRollDelta: -2, SavingThrowDelta: 2},
	}
	if !reflect.DeepEqual(conditionalRules, wantConditionalRules) {
		t.Fatalf("combat conditional rules=%+v want=%+v", conditionalRules, wantConditionalRules)
	}
	affectRules, err := pack.ResolveCombatAffectRules()
	if err != nil {
		t.Fatal(err)
	}
	wantAffectRules := []enginedamage.Rule{
		{ID: "coab.monster_affect_0a.resist_cold", EffectKind: 0x0A, DamageMask: 0x02, Mode: enginedamage.ModeHalf},
		{ID: "coab.monster_affect_70.fire_immunity", EffectKind: 0x70, DamageMask: 0x01, Mode: enginedamage.ModeImmune},
		{ID: "coab.monster_affect_87.electricity_immunity", EffectKind: 0x87, DamageMask: 0x04, Mode: enginedamage.ModeImmune},
	}
	if !reflect.DeepEqual(affectRules, wantAffectRules) {
		t.Fatalf("combat affect rules=%+v want=%+v", affectRules, wantAffectRules)
	}
	if pack.CharacterCreation == nil || len(pack.CharacterCreation.Templates) != 40 {
		t.Fatalf("character creation templates=%+v", pack.CharacterCreation)
	}
	first := pack.CharacterCreation.Templates[0]
	last := pack.CharacterCreation.Templates[len(pack.CharacterCreation.Templates)-1]
	if first.ID != "human.fighter" || first.DisplayID != "creation.class.fighter" ||
		first.RaceID != 5 || first.PrimaryClassID != 1 || first.RawClassID != 2 ||
		len(first.ClassLevels) != 8 || first.ClassLevels[2] != 1 ||
		len(first.BaseAbilities) != 6 || first.BaseAbilities[0] != 16 {
		t.Fatalf("first character creation template=%+v", first)
	}
	if last.ID != "half-orc.fighter-thief" || last.RawClassID != 14 ||
		last.ClassLevels[2] != 1 || last.ClassLevels[6] != 1 {
		t.Fatalf("last character creation template=%+v", last)
	}
	for _, language := range []string{"en", "zh-TW"} {
		for _, template := range pack.CharacterCreation.Templates {
			if text, ok := pack.Text(template.DisplayID, language); !ok || text == "" {
				t.Fatalf("template %q display %q missing from %s", template.ID, template.DisplayID, language)
			}
		}
	}
	if len(pack.CombatantNameRules) != 23 {
		t.Fatalf("combatant name rules=%d, want 23", len(pack.CombatantNameRules))
	}
	for _, rule := range pack.CombatantNameRules {
		for _, language := range []string{"en", "zh-TW"} {
			value, found := pack.LocalizeCombatantName(rule.Source, language)
			if !found || value == "" {
				t.Fatalf("combatant %q message %q missing from %s", rule.Source, rule.MessageID, language)
			}
		}
	}
	if value, found := pack.LocalizeCombatantName("HIPPO", "zh-TW"); found || value != "" {
		t.Fatalf("partial combatant source unexpectedly matched=%q,%v", value, found)
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
	if len(pack.CombatPlayerSpells) != 12 {
		t.Fatalf("combat player spells=%d, want 12", len(pack.CombatPlayerSpells))
	}
	playerSpellIDs := make(map[string]bool, len(pack.CombatPlayerSpells))
	for _, playerSpell := range pack.CombatPlayerSpells {
		if playerSpell.ID == "" || playerSpell.SpellID == 0 || playerSpell.CasterClass == "" ||
			playerSpell.Behavior == "" || playerSpell.MessageID == "" || playerSpellIDs[playerSpell.ID] {
			t.Fatalf("invalid or duplicate combat player spell=%+v", playerSpell)
		}
		playerSpellIDs[playerSpell.ID] = true
		for _, language := range []string{"en", "zh-TW"} {
			if text, ok := pack.Text(playerSpell.MessageID, language); !ok || text == "" {
				t.Fatalf("player spell %q message %q missing from %s", playerSpell.ID, playerSpell.MessageID, language)
			}
		}
	}
	magicMissilePlayer, found := pack.FindCombatPlayerSpell(0x0F, "magic_user")
	if !found || magicMissilePlayer.ID != "coab.spell.magic-missile" ||
		magicMissilePlayer.TargetMode != "enemy" || magicMissilePlayer.Behavior != "magic_missile" ||
		magicMissilePlayer.MessageID != "spell_magic_user_7" || magicMissilePlayer.CastingTime != 1 {
		t.Fatalf("Magic Missile player contract=%+v found=%v", magicMissilePlayer, found)
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
		cave.Spawn == nil || cave.Spawn.X != 5 || cave.Spawn.Y != 7 ||
		cave.Spawn.Direction != 6 || len(cave.SearchEdges) != 2 ||
		cave.SearchEdges[0].ID != "zhentil-keep.beholder-cave.dexam-east" ||
		cave.SearchEdges[0].X != 14 || cave.SearchEdges[0].Y != 1 ||
		cave.SearchEdges[0].Direction != 2 || cave.SearchEdges[0].WallType != 9 ||
		cave.SearchEdges[0].Discovery != "search_or_look" ||
		cave.SearchEdges[0].Confidence != "strong inference" ||
		cave.SearchEdges[1].ID != "zhentil-keep.beholder-cave.dexam-shrine-east" ||
		cave.SearchEdges[1].X != 15 || cave.SearchEdges[1].Y != 1 ||
		cave.SearchEdges[1].Direction != 2 || cave.SearchEdges[1].WallType != 9 ||
		cave.SearchEdges[1].Discovery != "search_or_look" ||
		cave.SearchEdges[1].Confidence != "strong inference" {
		t.Fatalf("Zhentil beholder cave map definition=%+v found=%v", cave, found)
	}
	var caveTeleportFound bool
	for _, event := range pack.Events {
		if event.ID != "zhentil-keep.beholder-cave.same-block-launch" {
			continue
		}
		caveTeleportFound = true
		if !event.Once || event.When.ECLBlock == nil || *event.When.ECLBlock != 0x22 ||
			event.When.Memory["0xC04B"] != 13 || event.When.Memory["0xC04C"] != 1 ||
			event.When.Memory["0xC04D"] != 3 || len(event.Actions) != 2 ||
			event.Actions[0].Type != "set_memory" || event.Actions[0].Address != "0x4C03" ||
			event.Actions[0].Value == nil || *event.Actions[0].Value != 0 ||
			event.Actions[1].Type != "set_map_position" || event.Actions[1].Position == nil {
			t.Fatalf("Beholder Cave teleporter event=%+v", event)
		}
		position := event.Actions[1].Position
		if position.AreaID != 4 || position.GeometryBlock != 0x25 || position.X != 13 ||
			position.Y != 1 || position.Direction != 6 || position.WallType == nil ||
			*position.WallType != 8 || position.WallRoof == nil || *position.WallRoof != 0xC0 {
			t.Fatalf("Beholder Cave teleporter position=%+v", position)
		}
	}
	if !caveTeleportFound {
		t.Fatal("missing Beholder Cave teleporter event")
	}
	hap, found := pack.FindMapByKindScript("first_person", 5, 0x31)
	if !found || hap.ID != "original.geo5.block-31" || hap.GeometryFile != "GEO5.DAX" ||
		hap.GeometryBlock != 0x32 || hap.ScriptBlock == nil || *hap.ScriptBlock != 0x31 {
		t.Fatalf("Hap script-to-geometry map definition=%+v found=%v", hap, found)
	}
	if len(hap.ExternalExits) != 1 || hap.ExternalExits[0].ID != "hap.village.east-to-dracolich-cave" ||
		hap.ExternalExits[0].RoofType == nil || *hap.ExternalExits[0].RoofType != 2 {
		t.Fatalf("Hap external exit definition=%+v", hap.ExternalExits)
	}
	tower, found := pack.FindMapByKindScript("first_person", 5, 0x33)
	if !found || tower.Spawn == nil || tower.Spawn.X != 7 || tower.Spawn.Y != 15 || tower.Spawn.Direction != 6 {
		t.Fatalf("Hap wizard-tower spawn definition=%+v found=%v", tower, found)
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
	wantPatrol, _ := pack.Text("zhentil.patrol_pass", "zh-TW")
	if !result.Matched || result.RuleID != "zhentil.patrol_pass" || result.Message != wantPatrol {
		t.Fatalf("text result=%+v", result)
	}
	greenRobes := pack.MatchText([]string{
		"AS YOU ARE PASSING THROUGH THE CROWDS, YOU HEAR,",
		"THE WOMAN IN THE GREEN ROBES -- EYES OF A FANATIC.",
	}, "zh-TW")
	wantGreenRobes, _ := pack.Text("tilverton.green-robes-rumor", "zh-TW")
	if !greenRobes.Matched || greenRobes.RuleID != "tilverton.green-robes-rumor" || greenRobes.Message != wantGreenRobes {
		t.Fatalf("green-robes text result=%+v", greenRobes)
	}
	dimswart := pack.MatchText([]string{
		"YOU SEE AN OLD MAN IN THE CELL.",
		"HE INTRODUCES HIMSELF AND YOU RECORD HIS REMARKS AS JOURNAL ENTRY 12.",
	}, "zh-TW")
	wantDimswart, _ := pack.Text("zhentil.dimswart_appears", "zh-TW")
	wantJournal121, _ := pack.Text("journal.12.1", "zh-TW")
	wantJournal126, _ := pack.Text("journal.12.6", "zh-TW")
	if !dimswart.Matched || dimswart.RuleID != "zhentil.dimswart_appears" ||
		dimswart.Message != wantDimswart || len(dimswart.JournalPages) != 6 ||
		dimswart.JournalPages[0] != wantJournal121 || dimswart.JournalPages[5] != wantJournal126 {
		t.Fatalf("Dimswart text result=%+v", dimswart)
	}
	dexam := pack.MatchText([]string{
		"DEXAM SPEAKS.",
		"YOU RECORD HIS SPEECH AS JOURNAL ENTRY 30.",
	}, "zh-TW")
	wantDexam, _ := pack.Text("dexam.journal_30", "zh-TW")
	wantJournal302, _ := pack.Text("journal.30.2", "zh-TW")
	if !dexam.Matched || dexam.RuleID != "dexam.journal_30" ||
		dexam.Message != wantDexam || len(dexam.JournalPages) != 2 ||
		dexam.JournalPages[1] != wantJournal302 {
		t.Fatalf("Dexam text result=%+v", dexam)
	}
	deadElfMap := pack.MatchText([]string{
		"YOU DISCOVER A MAP.  ON IT, YOU SEE DEXAMS ALTAR",
		"INDICATED AND A PATH THAT SEEMS TO LEAD OUTSIDE.",
		"YOU PLACE IT IN YOUR JOURNAL AS ENTRY 59.",
	}, "zh-TW")
	wantDeadElfMap, _ := pack.Text("dexam.dead-elf.map", "zh-TW")
	wantJournal59, _ := pack.Text("journal.59", "zh-TW")
	if !deadElfMap.Matched || deadElfMap.RuleID != "dexam.dead-elf.map" ||
		deadElfMap.Message != wantDeadElfMap || len(deadElfMap.JournalPages) != 1 ||
		deadElfMap.JournalPages[0] != wantJournal59 {
		t.Fatalf("dead-elf Journal 59 result=%+v", deadElfMap)
	}
	gasTrap := pack.MatchText([]string{"A GAS TRAP GOES OFF!"}, "zh-TW")
	wantGasTrap, _ := pack.Text("dexam.dead-elf.gas-trap", "zh-TW")
	if !gasTrap.Matched || gasTrap.RuleID != "dexam.dead-elf.gas-trap" || gasTrap.Message != wantGasTrap {
		t.Fatalf("dead-elf gas trap result=%+v", gasTrap)
	}
}

func TestBeholderCaveMapHandoffContinuesSameECLResult(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	runtime := &goldenbox.Runtime{
		ECLBlock: 0x22,
		Memory: map[uint16]uint16{
			0xC04B: 13,
			0xC04C: 1,
			0xC04D: 3,
			0x4C06: 1,
		},
		Locale: "zh-TW",
	}
	result, err := pack.ApplyFirst(runtime)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied || result.EventID != "zhentil-keep.beholder-cave.same-block-launch" ||
		!runtime.ContinueResult || runtime.Mode != "dungeon" || len(runtime.MapPositions) != 1 {
		t.Fatalf("beholder-cave handoff result=%+v runtime=%+v", result, runtime)
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
		{"wizard-tower.dragons-depart", []string{"THIS IS A MATTER BETWEEN MEN", "WE LEAVE YOU TO YOUR SQUABBLES"}, "振翅飛離"},
		{"wizard-tower.dracandros.calls-troops", []string{"TROOPS DEFEND ME", "DRACANDROS FLEES DOWN THE STAIRS"}, "逃下樓梯"},
		{"wizard-tower.safe-roof", []string{"HOLD THE ROOF WELL ENOUGH", "REST SAFELY"}, "安全休息"},
		{"wizard-tower.dragons-convinced", []string{"YOU HAVE CONVINCED US", "NO PLOT AGAINST", "DISPUTE WITH DRACANDROS"}, "沒有對付龍族的陰謀"},
		{"wizard-tower.dragons-condemn", []string{"YOU ARE RIGHT DRACANDROS", "THEY CONDEMN THEMSELVES"}, "自行定罪"},
		{"wizard-tower.take-dragon-heart", []string{"DRACANDROS ESCAPED DOWNSTAIRS", "DRAGON BODIES LIE STREWN ABOUT", "DO YOU TAKE ONE OF THEIR HEARTS"}, "取走其中一顆龍心"},
		{"wizard-tower.dragon-bodies", []string{"DRACANDROS ESCAPED DOWNSTAIRS", "DRAGON BODIES LIE STREWN ABOUT"}, "屍體散落"},
		{"wizard-tower.dragon-heart-acid", []string{"CUT INTO THE DRAGON", "SPRAY OF ACID", "EXTRACT THE HEART"}, "成功取出了龍心"},
		{"wizard-tower.wilderness-exit", []string{"STOP BY HAPTOOTH VILLAGE", "DEPART THE AREA"}, "哈普圖斯村"},
		{"wizard-tower.roof-exit", []string{"WAY DOWN TO THE CAVES", "SECRET PASSAGE", "DIRECTLY TO THE WILDERNESS"}, "直達荒野"},
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
	for source, want := range map[string]string{
		"ATTACK DRAGONS":          "攻擊龍群",
		"ATTACK WIZARD":           "攻擊法師",
		"PARLAY WITH THE DRAGONS": "與龍群交涉",
	} {
		if got, ok := pack.LocalizeOption(source, "zh-TW"); !ok || got != want {
			t.Fatalf("option %q=%q,%v want %q", source, got, ok, want)
		}
	}
	if got, want := sortedLocaleKeys(pack.Locales["zh-TW"]), sortedLocaleKeys(pack.Locales["en"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("zh-TW/en stable ID coverage differs:\nzh-TW=%v\nen=%v", got, want)
	}
}

func TestLegacyJournalTriggersAndPagesAreGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
		pages     int
	}{
		{"journal-trigger.pit-temple-map-20", []string{"YOU HAVE ALSO FOUND A MAP OF THE TEMPLE", "JOURNAL ENTRY 20"}, 1},
		{"journal-trigger.alias-story-3", []string{"SHE TELLS HER STORY", "JOURNAL ENTRY 3"}, 3},
		{"journal-trigger.yulash-commander-22", []string{"YOU HAVE PLEASED THE COMMANDER", "JOURNAL ENTRY 22"}, 3},
		{"journal-trigger.guildmaster-map-4", []string{"THE GUILDMASTER GASPS", "JOURNAL ENTRY 4"}, 1},
		{"journal-trigger.fire-knives-leader-11", []string{"LEADER OF THE FIRE KNIVES", "JOURNAL ENTRY", "11"}, 2},
		{"journal-trigger.fire-knives-victory-54", []string{"FIRE KNIVES HAVE BEEN DEFEATED", "JOURNAL ENTRY 54"}, 1},
		{"journal-trigger.fire-knives-royal-arrival-53", []string{"FREEING GIOGI", "JOURNAL ENTRY 53"}, 2},
		{"journal-trigger.tilverton-inn-31", []string{"JOURNAL ENTRY 31."}, 1},
		{"journal-trigger.filani-38", []string{"SHE TALKS", "38."}, 3},
		{"journal-trigger.tavern-knife-17", []string{"ORNATE KNIFE", "17."}, 1},
		{"journal-trigger.shadowdale-warning-18", []string{"A HOODED, GREY ROBED MAN SITS IN A DARK CORNER", "18."}, 1},
		{"journal-trigger.high-priest-19", []string{"REMOVE CURSE SPELL", "19."}, 1},
		{"journal-trigger.frozen-room-26", []string{"YOU DISARM THE FIRE KNIVES", "JOURNAL ENTRY 26"}, 1},
		{"journal-trigger.fire-knives-office-9", []string{"DRAWERS OF A ROSEWOOD DESK", "9. OTHER ITEMS"}, 1},
		{"journal-trigger.fire-knives-paper-29", []string{"HAND KEPT THE PAPER", "JOURNAL ENTRY 29"}, 1},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != test.pages {
					t.Fatalf("result=%+v, want rule %q with %d pages", result, test.id, test.pages)
				}
				for index, page := range result.JournalPages {
					if page == "" {
						t.Fatalf("journal page %d is empty: %+v", index, result)
					}
				}
			})
		}
	}
}

func TestOpeningNarrativesAreGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"opening.curse-summary", []string{"ON YOUR WAY TO THE TOWN OF TILVERTON YOU ARE", "THE SYMBOLS ENSNARE YOUR WILL LIKE METAL BONDS", "AND REGAIN CONTROL OF YOUR OWN DESTINY", "THE MOST PEACEFUL SCENE CAN HIDE A DEADLY FOE"}},
		{"opening.new-game-awakening", []string{"YOU AWAKEN IN A SMALL ROOM", "ALL YOUR GEAR IS GONE"}},
		{"opening.new-game-marks", []string{"ADDING TO YOUR DISQUIET", "IMPRINTED WITH STRANGE PATTERNS", "IDENTICALLY MARKED"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestHapVillageStoryIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"hap.black-dragons", []string{"SAILING ACROSS THE SKY ARE GREAT BLACK SHAPES", "FEARSOME BLACK DRAGONS"}},
		{"hap.edge", []string{"YOU ARE AT THE EDGE OF HAP"}},
		{"hap.abandoned-village", []string{"THIS RUN DOWN VILLAGE IS STRANGELY QUIET", "NO ONE IS ABOUT"}},
		{"hap.hiding-peasants", []string{"YOU BURST IN ON SOME PEASANTS WHO SCUTTLE BACK", "LEAVE BEFORE THE HORDE FINDS YOU WITH US"}},
		{"hap.peasants-flee", []string{"THE CRINGING PEASANTS FLEE OUT INTO THE STREET"}},
		{"hap.dark-elf-attack", []string{"YOU UNGRATEFUL SLIME", "BE HAPPY FOR A QUICK DEATH"}},
		{"hap.akabar-join", []string{"I AM AKABAR BEL AKASH", "WILL YOU LET HIM JOIN YOUR PARTY"}},
		{"hap.inn-before-liberation", []string{"A SURLY INNKEEPER COMES UP", "DO YOU STAY"}},
		{"hap.efreet-barn", []string{"THIS BARN IS EMPTY", "EFREET AND HIS DARK ELFIN COHORTS"}},
		{"hap.efreet-threat", []string{"THE EFREET VOICE BOOMS OUT", "DOOM ON YOUR VILLAGE"}},
		{"hap.efreet-map", []string{"ON THE BODY OF THE EFREET IS A MAP", "THE TOWN AND A CAVE"}},
		{"hap.liberated-crowd", []string{"A SHORT TIME AFTER THE SOUNDS OF BATTLE FADE", "LOUD CHEERS AND LAUGHTER"}},
		{"hap.elder-thanks", []string{"AN ELDER OF THE VILLAGE COMES FORWARD", "ALWAYS BE WELCOME IN HAPTOOTH"}},
		{"hap.elder-wizard-tower", []string{"THE ELDER LOWERS HIS VOICE", "CONTROLLED FROM THE WIZARD'S TOWER NEARBY"}},
		{"hap.akabar-secret-routes", []string{"AKABAR MENTIONS THAT HE HAS HEARD OF SECRET TRADE ROUTES", "HAPPY TO GUIDE THE PARTY THERE"}},
		{"hap.leave", []string{"YOU ARE HEADING BACK TO THE WILDERNESS", "WANT TO CONTINUE"}},
		{"hap.map-route", []string{"FOLLOW THE MAP TO THE CAVES", "GO INTO THE WILDERNESS"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestLavaTubeStoryIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"lava-tube.entry", []string{"YOU HAVE ENTERED AN ANCIENT LAVA TUBE", "ASH COVERS THE FLOOR"}},
		{"lava-tube.ambush", []string{"FROM HIDDEN ALCOVES COMES A WAVE OF HEAT", "SALAMANDERS AND DARK ELVES"}},
		{"lava-tube.guarded-door", []string{"THE DOOR IS GUARDED BY A SALAMANDER LED PATROL"}},
		{"lava-tube.dream-warning", []string{"A DREAM-LIKE VOICE IN YOUR HEAD SAYS", "BE FULLY PREPARED"}},
		{"lava-tube.salamander-pools", []string{"THE ROOM IS FILLED WITH ACTIVE GEYSERS AND LAVA PITS", "SALAMANDERS ARE SPORTING IN THE POOLS"}},
		{"lava-tube.intense-heat", []string{"INTENSE HEAT WASHES OVER YOU"}},
		{"lava-tube.sly-parlay", []string{"WE HAVE NO LOVE FOR DARK ELVES", "TAKE ANY TREASURE"}},
		{"lava-tube.nice-parlay", []string{"YOU COLD THINGS SHOULD LEAVE", "CRIMDRAC FINDS YOU"}},
		{"lava-tube.fireproof-casks", []string{"AMONGST THE POOLS OF LAVA", "SIX FIREPROOF CASKS", "OPEN ONE"}},
		{"lava-tube.cask-heat-retreat", []string{"THE HEAT IS TOO INTENSE", "DOES ANYONE WANT TO TRY AGAIN"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestAreaFiveDepartureIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"area5.depart-akabar", []string{"YOUR HELP WAS INVALUABLE TO ME", "BUSINESS TO ATTEND TO"}},
		{"area5.depart-akabar-reluctant", []string{"AKABAR FROWNS AT YOU", "MUST FREE THIS TOWN FROM THE WIZARD'S TYRANY", "HE STAMPS OFF"}},
		{"area5.dark-elf-gear-decays", []string{"DARK ELF", "DECAY TO USELESSNESS"}},
		{"post-wizard.dracolich", []string{"OUT OF A COPSE OF TREES COMES A SKELETAL", "YOU HAVE DEPRIVED ME OF MY TUTOR", "I CAN AVENGE MYSELF"}},
		{"essembra.edge", []string{"YOU ARE AT THE EDGE OF ESSEMBRA"}},
		{"essembra.places", []string{"YOU ARE IN ESSEMBRA", "WHAT PLACE WILL YOU VISIT"}},
		{"essembra.branching-oak", []string{"WELCOME TO THE BRANCHING OAK"}},
		{"essembra.outdoor-bar", []string{"YOU ARE IN A N OUTDOOR BAR", "OVERLOOKING THE WOODS"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestHillsfarStoryIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"hillsfar.fire-knives-ambush", []string{"AMBUSHED BY FIRE KNIVES DISGUISED AS FIGHTERS"}},
		{"hillsfar.edge", []string{"YOU ARE AT THE EDGE OF HILLSFAR"}},
		{"hillsfar.places", []string{"YOU ARE IN HILLSFAR", "WHAT PLACE WILL YOU VISIT"}},
		{"hillsfar.dockside-bar", []string{"YOU ARE IN A DOCKSIDE BAR"}},
		{"hillsfar.red-plumes-spill-drinks", []string{"SOME RED PLUMES COME OVER", "ORDER YOU TO CLEAN UP THE MESS"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestYulashStoryIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"yulash.red-plume-patrol", []string{"YOU ARE APPROACHED BY A RED PLUME PATROL", "TATOO BETRAYS YOU AS A ZHENTRIM SPY"}},
		{"yulash.edge", []string{"YOU ARE AT THE EDGE OF YULASH"}},
		{"yulash.entry", []string{"SMOKE RISES FROM BEHIND THE RUINED WALLS", "OF YULASH", "HOW DO YOU ENTER"}},
		{"yulash.riders-burst-out", []string{"JUST BEFORE YOU ENTER A MAN MOUNTED ON A LARGE HORSE", "A WOMAN DRESSED IN PURPLE", "SORRY"}},
		{"yulash.checkpoint-halt", []string{"HALT! A GUARD WARILY COMES OUT OF A CHECKPOINT", "OTHER GUARDS GATHER BEHIND HIM"}},
		{"yulash.see-commander", []string{"YOU MUST COME WITH US TO SEE THE COMMANDER"}},
		{"yulash.waiting-room", []string{"THIS IS THE COMMANDER'S WAITING ROOM", "REMAIN HERE UNTIL YOU ARE CALLED"}},
		{"yulash.zhentarim-spies", []string{"TROOPS COME BURSTING OUT OF THE COMMANDER'S OFFICE", "THEY'RE SPIES FOR ZHENTIL KEEP"}},
		{"yulash.led-to-commander", []string{"YOU HAVE BEEN LED IN TO SEE THE RED PLUME COMMANDER"}},
		{"yulash.commander-business", []string{"THE COMMANDER DEMANDS TO KNOW YOUR BUSINESS IN YULASH", "HOW DO YOU RESPOND"}},
		{"yulash.commander-side-door", []string{"THE COMMANDER SHOWS YOU OUT THE SIDE DOOR"}},
		{"yulash.pit-entrance", []string{"THE PIT CREATED BY MOANDER", "STEP FORWARD TO ENTER THE DARK DEMESNE"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestPitOpeningAndCompanionsAreGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"pit.opening-dead-cultists", []string{"YOU SEE THREE CULTISTS LYING DEAD ON THE FLOOR", "JUST AHEAD OF YOU, ANOTHER CLERIC GASPS FOR BREATH"}},
		{"pit.opening-chosen", []string{"THE WOUNDED CLERIC'S EYES WIDEN IN FANATIC", "TRIUMPH. HE HOWLS", "THE CHOSEN ONES"}},
		{"pit.trapped", []string{"THE CLERIC SLAMS HIS FIST AGAINST A PROTRUDING ROCK", "YOU ARE TRAPPED IN THE PIT OF MOANDER"}},
		{"pit.cleric-dies", []string{"THE CLERIC GIVES YOU ONE LAST TRIUMPHANT GLARE", "COUGHS BLOOD AND DIES AT YOUR FEET"}},
		{"pit.ambience", []string{"YOU HEAR THE SOUNDS OF BATTLE IN THE DISTANCE", "SMELL OF BAKED BREAD"}},
		{"pit.alias-dragonbait-meet", []string{"YOU SEE A FEMALE FIGHTER AND A STRANGE-LOOKING LIZARD MAN", "VIOLETS, BRIMSTONE AND HONEYSUCKLE"}},
		{"pit.alias-bonded-reaction", []string{"THE FEMALE FIGHTER GASPS", "THEY'RE BONDED", "WHAT DO YOU DO"}},
		{"pit.alias-dragonbait-introduction", []string{"THE FIGHTER INTRODUCES HERSELF AS ALIAS", "HER COMPANION AS DRAGONBAIT", "SHE ASKS YOU TO TELL YOUR STORY"}},
		{"pit.alias-dragonbait-join", []string{"DO YOU WANT THEM TO JOIN YOU"}},
		{"pit.alias-dragonbait-joined", []string{"ALIAS AND DRAGONBAIT JOIN YOUR PARTY", "TREASURE THAT MOGION", "KEEPS BEHIND HER ALTAR"}},
		{"pit.stairs-down", []string{"YOU SEE STAIRS LEADING DOWN TO THE SOUTH", "DO YOU WISH TO GO DOWN"}},
		{"pit.stairs-up", []string{"YOU SEE STAIRS GOING UP IN THE NORTH WALL", "DO YOU WISH TO GO UP"}},
		{"pit.dead-zhentrim", []string{"MANGLED REMAINS OF A DEAD ZHENTRIM FIGHTER", "WHAT DO YOU DO"}},
		{"pit.zhentrim-scroll", []string{"GRASPED IN THE FIGHTER'S FIST", "SEAL OF ZHENTIL", "JOURNAL ENTRY 46"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestPitMogionFinaleIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"pit.mogion-altar", []string{"YOU SEE A PRIESTESS TURN AND SMILE WICKEDLY", "CULTISTS CHANTING IN A LOW DRONE"}},
		{"pit.mogion-self-identifies", []string{"I AM MOGION"}},
		{"pit.alias-identifies-mogion", []string{"ALIAS MUTTERS", "PRIESTESS OF MOANDER", "SPITS ON THE GROUND"}},
		{"pit.mogion-greeting", []string{"MOGION SAYS", "PROPER TOOLS", "WHAT DO YOU DO"}},
		{"pit.bond-paralysis", []string{"BEFORE YOU CAN ACT", "BLUE FLASH", "YOU CANNOT MOVE"}},
		{"pit.alias-dragonbait-tendrils", []string{"TENDRILS COME UP FROM THE FLOOR", "ALIAS AND DRAGONBAIT"}},
		{"pit.mogion-ritual", []string{"MOGION TURNS TO THE ALTAR", "CHANTING RISES"}},
		{"pit.dimensional-window", []string{"BLUE LIGHT THAT SURROUNDS YOU", "DIMENSIONAL WINDOW ABOVE THE ALTAR"}},
		{"pit.moander-returns", []string{"MOGION SHRIEKS", "MOANDER RETURNS", "DIMENSIONAL RIFT"}},
		{"pit.bond-fades", []string{"ENERGY IN THE DIMENSIONAL RIFT INCREASES", "BOND OF MOANDER BEGIN TO FADE"}},
		{"pit.bond-broken", []string{"THE SIGIL DISAPPEARS", "PARALYSIS THAT GRIPPED YOU IS NOW GONE"}},
		{"pit.alias-attack-mogion", []string{"ALIAS AND DRAGONBAIT HAVE HACKED THEIR WAY FREE", "UNLESS YOU WISH TO FIGHT A GOD"}},
		{"pit.rift-closes", []string{"THE DIMENSIONAL RIFT SNAPS SHUT"}},
		{"pit.remnants-scream", []string{"THREE PSUEDOPODS OF MOANDER", "HUNDREDS OF MOUTHS", "YOU HAVE KILLED ME"}},
		{"pit.remnants-attack", []string{"THE OOZING MOUNDS TURN AND ATTACK YOU"}},
		{"pit.gauntlet", []string{"YOU FIND THE GAUNTLET OF MOANDER", "SLIMY REMAINS"}},
		{"pit.priest-flees", []string{"A PRIEST BURSTS INTO THE ROOM", "THEY HAVE KILLED THE GOD"}},
		{"pit.altar-treasure", []string{"YOU HAVE FOUND A CACHE OF JEWELS AND GEMS"}},
		{"pit.exit-last-stand", []string{"YOU ARE ATTACKED BY A LARGE FORCE OF", "CULTISTS IN A LAST-DITCH EFFORT TO STOP YOU"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestAshabenfordAndStandingStoneStoryIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"ashabenford.tilvers-gap", []string{"MOUNTAINS RISE INTO AN IMPASSABLE WALL", "TILVER'S GAP", "FLYING SHAPES"}},
		{"ashabenford.edge", []string{"YOU ARE AT THE EDGE OF ASHABENFORD"}},
		{"ashabenford.places", []string{"YOU ARE IN ASHABENFORD", "WHAT PLACE WILL YOU VISIT"}},
		{"ashabenford.ale-house", []string{"YOU ARE IN A RIVERSIDE ALE HOUSE", "WHAT WILL YOU DO"}},
		{"ashabenford.tavern-tale-28", []string{"YOU OVERHEAR TAVERN TALE 28"}},
		{"shadow-gap.fire-knives-patrol", []string{"AMBUSHED BY FIRE KNIVES DISGUISED AS A PATROL"}},
		{"standing-stone.grey-man", []string{"YOU ARE AT THE STANDING STONES", "GREY ROBED MAN"}},
		{"standing-stone.four-masters", []string{"YOU PRESENTLY SERVE FOUR MASTER", "RETURN TO ME WHEN YOU HAVE SLAIN THREE MORE"}},
		{"standing-stone.seek-red", []string{"SEEK RED TO THE SOUTH"}},
		{"world-route.essembra", []string{"HOW WILL YOU GET TO ESSEMBRA"}},
		{"world-route.hap", []string{"HOW WILL YOU GET TO HAP"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestTilvertonGuildAndHideoutTransitionIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"tilverton.guildmaster-greeting", []string{"BEFORE YOU STANDS A BURLY MAN", "CARE TO REST"}},
		{"tilverton.running-thieves", []string{"YOU COME UPON SOME RUNNING THIEVES", "WHAT DO YOU DO"}},
		{"tilverton.running-thieves-warning", []string{"THEY YELL", "FIRE KNIVES ARE PUSHING UP FROM THE SOUTH", "BOILING OUT OF THE SEWERS"}},
		{"tilverton.fire-knives-spot-you", []string{"A PARTY OF FIRE KNIVES SPOTS YOU"}},
		{"tilverton.guild-assassins-attack", []string{"ASSASSINS LEAP ON YOU"}},
		{"tilverton.guild-metal-and-animals", []string{"CLANG OF METAL ON METAL", "GROWLS OF ANIMALS", "MORTAL COMBAT"}},
		{"tilverton.guild-bodies-after-battle", []string{"BODIES LIE TWISTED ONE ABOUT ANOTHER", "LOCKED IN COMBAT UNTIL DEATH"}},
		{"tilverton.guildmaster-briefing", []string{"THE FIRE KNIVES HAVE THE KING'S DAUGHTER", "I CAN OFFER INFORMATION"}},
		{"tilverton.guild-breach", []string{"SIDE DOOR EXPLODES INWARD", "DEAFENING CRASH"}},
		{"tilverton.guild-fire-knife-command", []string{"TRAITOROUS SCUM", "SEIZE THEM ALL"}},
		{"tilverton.guild-poisoned-dagger", []string{"GUILDMASTER HURLS A POISONED DAGGER", "TWITCHING VIOLENTLY"}},
		{"tilverton.guild-battle-joined", []string{"ARROW HIT THE GUILDMASTER IN THE CHEST", "THE BATTLE IS JOINED"}},
		{"tilverton.guild-halfling", []string{"HALFLING WITH A HARP", "DISAPPEAR"}},
		{"tilverton.guild-kennel-intro", []string{"HUNGRY SNARLS", "RELEASES THE PACK"}},
		{"tilverton.guild-kennel-aftermath", []string{"GNAWED BONES", "LEASHES"}},
		{"tilverton.guild-monkey-cages", []string{"CAGES THAT ONCE HELD MONKEYS"}},
		{"tilverton.guild-guest-book", []string{"OPEN GUEST BOOK", "O.RUSKETTLE"}},
		{"tilverton.guild-sewer-traces", []string{"GREEN SLIMY MARKS", "MORE DISTINCT NEAR THE DOOR"}},
		{"tilverton.sewers-entry", []string{"FOUL SMELLING, SLIME COVERED", "FIGHTING WILL BE", "AWKWARD"}},
		{"tilverton.sewers.guild-battle-echoes", []string{"YOU STILL HEAR THE OCCASIONAL SOUNDS OF BATTLE", "ECHOING FROM THE GUILD HALL"}},
		{"fire-knife.hideout-entry", []string{"YOU ARE ENTERING THE HIDEOUT"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestFireKnifeHideoutRoomsAreGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"fire-knife.blade-barrier-fades", []string{"BLADES SLOW DOWN", "FADE AWAY"}},
		{"fire-knife.blade-barrier", []string{"CLOUD OF BLADES WHIRLING", "METALLIC WHINE"}},
		{"fire-knife.blade-barrier-damage", []string{"THE BLADES TEAR INTO YOU"}},
		{"fire-knife.frozen-kill", []string{"YOU SLAUGHTER THEM", "BEING HELD"}},
		{"fire-knife.frozen-room", []string{"PEOPLE FROZEN IN", "BEGINNING TO MOVE"}},
		{"fire-knife.office", []string{"ORNATE ROOM", "HIGH UP IN THE FIRE KNIVES"}},
		{"fire-knife.smoky-hall", []string{"STRANGE SMOKY SCENT"}},
		{"fire-knife.ordered-bedroom", []string{"EXTREMELY WELL ORDERED BEDROOM", "UNSEEN SERVANTS"}},
		{"fire-knife.burned-library", []string{"ROOM WAS ONCE A LIBRARY", "CHARRED BODY"}},
		{"fire-knife.burned-lab", []string{"ONCE A LAB", "NOTHING ESCAPED DESTRUCTION"}},
		{"fire-knife.shrouded-bodies", []string{"TWO ROWS OF SHROUDED BODIES", "TO BE RAISED"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestTilvertonCarriageAndSewersStoryIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"tilverton.carriage-gate-closed", []string{"THIS WAY IS CLOSED", "ROYAL CARRIAGE IS COMING SOON"}},
		{"tilverton.carriage-make-way", []string{"MAKE WAY FOR THE ROYAL CARRIAGE"}},
		{"tilverton.carriage-bond-compulsion", []string{"KING'S VOICE COMING FROM THE CARRIAGE", "COMPULSION TO ATTACK"}},
		{"tilverton.carriage-false-king", []string{"I'M NOT REALLY THE KING", "OH NO! NOT AGAIN"}},
		{"tilverton.carriage-alarm", []string{"LOUD BELL STARTS RINGING", "SWORDS DRAWN"}},
		{"tilverton.carriage-abduction", []string{"TWO RED ROBED MEN JUMP THE CARRIAGE", "DRAG HIM INTO AN ALLEYWAY"}},
		{"tilverton.carriage-surrender", []string{"DO YOU SURRENDER"}},
		{"tilverton.carriage-jailed", []string{"YOU ARE THROWN IN JAIL"}},
		{"tilverton.carriage-thief-rescue", []string{"ONE WALL SLIDES OPEN AND A THIEF APPEARS", "SIGNALS YOU TO FOLLOW HIM"}},
		{"tilverton.carriage-guild-arrival", []string{"LEADS YOU THROUGH HIDDEN PASSAGES", "THE THIEVES' GUILD"}},
		{"tilverton.sewers-checkpoint", []string{"FIRE KNIVES DEMAND YOUR IMMEDIATE SURRENDER", "DO YOU SURRENDER"}},
		{"tilverton.sewers-hide-bodies", []string{"YOU QUICKLY HIDE THEIR BODIES"}},
		{"tilverton.sewers-knight-appears", []string{"SLAUGHTERED REMAINS OF A FIRE KNIFE", "KNIGHTS OF MYTH DRANNOR"}},
		{"tilverton.sewers-knight-allegiance", []string{"BLUE TATTOO MARKINGS OF THE FIRE KNIVES", "TO WHOM DO YOU OWE ALLEGIANCE"}},
		{"tilverton.sewers-knight-princess-friend", []string{"THAT PRINCESS IS A POPULAR GIRL", "DON'T KILL THE CLERIC WITH A HAMMER"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestTilvertonBondDreamAndReturnBoundaryIsGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"tilverton.high-priest-intro", []string{"I AM THE HIGH PRIEST", "TELL ME YOUR STORY"}},
		{"tilverton.edge", []string{"YOU ARE AT THE EDGE OF TILVERTON"}},
		{"tilverton.entry-barred", []string{"GUARDS BAR YOUR WAY"}},
		{"bond-dream.first-night", []string{"FIRST NIGHT OUTSIDE THE CITY", "VIVID DREAM"}},
		{"bond-dream.masters-taunt", []string{"FOUR FACES LEER DOWN", "WEAKEST OF YOUR MASTERS"}},
		{"bond-dream.masters-prophecy", []string{"WIZARD IN RED", "PAWNS OF THE FLAMED ONE"}},
		{"bond-dream.ends", []string{"FACES LAUGH WITH EVIL JOY", "AWAKE IN A COLD SWEAT"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
	}
}

func TestEssembraTavernTalesAreGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id        string
		fragments []string
	}{
		{"essembra.tavern-tale-60", []string{"YOU OVERHEAR TAVERN TALE 60"}},
		{"essembra.tavern-tale-44", []string{"YOU OVERHEAR TAVERN TALE 44"}},
	}
	for _, test := range tests {
		for _, language := range []string{"en", "zh-TW"} {
			t.Run(test.id+"/"+language, func(t *testing.T) {
				result := pack.MatchText(test.fragments, language)
				if !result.Matched || result.RuleID != test.id || result.Message == "" || len(result.JournalPages) != 0 {
					t.Fatalf("result=%+v", result)
				}
			})
		}
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

func TestAllLegacyECLMenuTokensAreGamePackDriven(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	legacySources := []string{
		"PRESS BUTTON OR RETURN TO CONTINUE.", "YES", "NO", "TELL THE TRUTH",
		"PUNCH BARKEEP", "HAVE A DRINK", "DRAGON'S BREATH", "BASILISK",
		"LEMONADE", "WHISKEY", "BEER", "ALE", "PORT", "MEAD", "LIE",
		"ENTER CITY", "JOURNEY ON", "CAMP", "SEARCH AREA", "INN", "STORE",
		"BAR", "HALL", "TEMPLE", "RELAX", "PATROL FOREST", "THANK HIM",
		"ATTACK", "REMAIN CALM", "ENTER IT", "SPEAK", "HACK IT", "GREET", "LEAVE", "Leave",
		"EXAMINE CORPSE", "SHADOWDALE", "ASHABENFORD", "DAGGER FALLS",
		"TILVERTON", "THE STANDING STONE", "ESSEMBRA", "HAP", "HILLSFAR",
		"VOONLAR", "PHLAN", "TESHWAVE", "YULASH", "ZHENTIL KEEP", "MYTH DRANNOR",
		"SNEAK IN", "ASK PERMISSION", "RUN AWAY", "FIGHT", "GO WITH GUARDS",
		"FIGHT THE MEN", "LET THEM GO", "TELL HER YOUR STORY",
		"TELL HER YOU'RE HUNTING CULTISTS", "TELL HER IT'S NONE OF HER AFFAIR",
		"TRY TO TALK FURTHER", "WILDERNESS", "CAVES", "STAY HERE", "VILLAGE",
		"DEPART", "TRAIL", "COMBAT", "WAIT", "ENTER THE BLADES", "RETREAT",
		"INTERROGATE", "KILL", "FLEE", "ADVANCE", "PARLAY", "PARLAY_HAUGHTY",
		"PARLAY_SLY", "PARLAY_MEEK", "PARLAY_NICE", "PARLAY_ABUSIVE",
		"FIRE KNIVES", "PRINCESS NACACIA", "NO ONE", "EXIT",
	}
	if len(legacySources) != 85 {
		t.Fatalf("legacy source oracle count=%d", len(legacySources))
	}
	for _, source := range legacySources {
		for _, language := range []string{"en", "zh-TW"} {
			if value, ok := pack.LocalizeOption(source, language); !ok || value == "" {
				t.Fatalf("option %q missing from %s", source, language)
			}
		}
	}
	// The title pack may legitimately gain options as another ECL menu becomes
	// data-driven. Verify the actual stable-ID binding instead of freezing a
	// historical rule count that would reject a valid new menu.
	for _, rule := range pack.OptionRules {
		for _, language := range []string{"en", "zh-TW"} {
			want, present := pack.Locales[language][rule.MessageID]
			got, found := pack.LocalizeOption(rule.Source, language)
			if !present || !found || got != want || got == "" {
				t.Fatalf("option rule %q source=%q language=%s resolved=%q found=%v want=%q present=%v", rule.ID, rule.Source, language, got, found, want, present)
			}
		}
	}
}

func TestPackAndUILocaleSharedStableIDsDoNotDrift(t *testing.T) {
	pack, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "assets", "locale", "zh-TW.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := locale.Load(data)
	if err != nil {
		t.Fatal(err)
	}
	compared := 0
	for messageID, packValue := range pack.Locales["zh-TW"] {
		uiValue, ok := catalog.Strings[messageID]
		if !ok {
			continue
		}
		compared++
		if uiValue != packValue {
			t.Fatalf("shared stable ID %q drifted: pack=%q ui=%q", messageID, packValue, uiValue)
		}
	}
	if compared != 77 {
		t.Fatalf("shared stable IDs compared=%d, want the 77 current pack/UI overlaps", compared)
	}
}
