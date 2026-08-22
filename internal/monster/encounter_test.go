package monster

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

func TestBuildEnemiesJoinsSpawnAndRecords(t *testing.T) {
	record := Record{Name: "Goblin", HitPoints: 5, MaxHitPoints: 5, ArmorClass: 10, DamageDiceCount: 1, DamageDiceSides: 6}
	enemies, err := BuildEnemies([]ecl.MonsterSpawn{{MonsterID: 7, Count: 2}}, map[uint8]Record{7: record})
	if err != nil {
		t.Fatal(err)
	}
	if len(enemies) != 2 || enemies[0].ID != "monster-07-1" || enemies[1].Name != "Goblin" || enemies[1].Side != 1 {
		t.Fatalf("enemies=%#v", enemies)
	}
}

func TestBuildEnemiesCarriesCPICBlock(t *testing.T) {
	record := Record{Name: "Orc", HitPoints: 5, MaxHitPoints: 5, ArmorClass: 10}
	enemies, err := BuildEnemies([]ecl.MonsterSpawn{{MonsterID: 7, Count: 1, IconBlock: 0x35}}, map[uint8]Record{7: record})
	if err != nil || len(enemies) != 1 || enemies[0].SpriteBlock != 0x35 {
		t.Fatalf("enemies=%#v err=%v", enemies, err)
	}
}

func TestBuildEnemiesRejectsMissingRecord(t *testing.T) {
	if _, err := BuildEnemies([]ecl.MonsterSpawn{{MonsterID: 7, Count: 1}}, nil); err == nil {
		t.Fatal("expected missing record error")
	}
}

func TestBuildEnemiesWithAffectsCopiesSPCRecords(t *testing.T) {
	records := map[uint8]Record{7: {Name: "Goblin", MaxHitPoints: 3, HitPoints: 3}}
	affects := map[uint8][]AffectRecord{7: {{Kind: 0x19, Duration: 4, Active: true}}}
	enemies, err := BuildEnemiesWithAffects([]ecl.MonsterSpawn{{MonsterID: 7, Count: 2}}, records, affects)
	if err != nil || len(enemies) != 2 {
		t.Fatalf("enemies=%#v err=%v", enemies, err)
	}
	if len(enemies[0].MonsterAffects) != 1 || len(enemies[1].MonsterAffects) != 1 {
		t.Fatalf("effects=%#v", enemies)
	}
	if !enemies[0].MonsterAffects[0].Innate {
		t.Fatalf("MON*SPC effect was not marked innate: %#v", enemies[0].MonsterAffects)
	}
	affects[7][0].Duration = 99
	enemies[0].MonsterAffects[0].Duration = 88
	if enemies[1].MonsterAffects[0].Duration != 4 {
		t.Fatalf("effect instances alias each other: %#v", enemies)
	}
}

// 加速／緩速動的是**半次**單位，而且在換算成整數次數**之前**做——先換算再加倍
// 會把 1.5 次的那半次先丟掉（spec 1180）。
func TestBuildEnemiesWithAffectsProjectsHasteAttacks(t *testing.T) {
	records := map[uint8]Record{7: {Name: "Ogre", MaxHitPoints: 8, HitPoints: 8,
		AttackBlows: [2]uint8{2, 0}}}
	affects := map[uint8][]AffectRecord{7: {{Kind: 0x27, Active: true}}}
	enemies, err := BuildEnemiesWithAffects([]ecl.MonsterSpawn{{MonsterID: 7}}, records, affects)
	if err != nil || len(enemies) != 1 || enemies[0].AttackBlows != [2]int{4, 0} {
		t.Fatalf("enemies=%#v err=%v", enemies, err)
	}
	affects[7][0].Active = false
	enemies, err = BuildEnemiesWithAffects([]ecl.MonsterSpawn{{MonsterID: 7}}, records, affects)
	if err != nil || enemies[0].AttackBlows != [2]int{4, 0} {
		t.Fatalf("raw MON*SPC byte-4 zero suppressed innate haste: enemies=%#v err=%v", enemies, err)
	}
	affects[7] = []AffectRecord{{Kind: 0x2A, Active: true}}
	enemies, err = BuildEnemiesWithAffects([]ecl.MonsterSpawn{{MonsterID: 7}}, records, affects)
	if err != nil || enemies[0].AttackBlows != [2]int{1, 0} {
		t.Fatalf("slow attacks=%#v err=%v", enemies, err)
	}
	affects[7] = []AffectRecord{{Kind: 0x27, Active: true}, {Kind: 0x2A, Active: true}}
	enemies, err = BuildEnemiesWithAffects([]ecl.MonsterSpawn{{MonsterID: 7}}, records, affects)
	if err != nil || enemies[0].AttackBlows != [2]int{2, 0} {
		t.Fatalf("haste+slow attacks=%#v err=%v", enemies, err)
	}
}

func TestBuildEnemiesWithAffectsPreservesHeldState(t *testing.T) {
	records := map[uint8]Record{7: {Name: "眠怪", MaxHitPoints: 4, HitPoints: 4}}
	enemies, err := BuildEnemiesWithAffects([]ecl.MonsterSpawn{{MonsterID: 7}}, records, map[uint8][]AffectRecord{7: {{Kind: 0x34, Active: true}}})
	if err != nil || len(enemies) != 1 || !enemies[0].MonsterIsHeld() {
		t.Fatalf("enemies=%#v err=%v", enemies, err)
	}
}
