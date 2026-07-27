package monster

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// BuildEnemies joins ECL spawn descriptors with decoded MON*CHA records. It
// intentionally returns fighters without party or map state; those belong to
// the battle/session adapter.
func BuildEnemies(spawns []ecl.MonsterSpawn, records map[uint8]Record) ([]combat.Fighter, error) {
	return BuildEnemiesWithAffects(spawns, records, nil)
}

// BuildEnemiesWithAffects also attaches chapter-local MON*SPC effects.
func BuildEnemiesWithAffects(spawns []ecl.MonsterSpawn, records map[uint8]Record, affects map[uint8][]AffectRecord) ([]combat.Fighter, error) {
	enemies := make([]combat.Fighter, 0)
	for _, spawn := range spawns {
		record, ok := records[spawn.MonsterID]
		if !ok {
			return nil, fmt.Errorf("monster record 0x%02X is unavailable", spawn.MonsterID)
		}
		count := int(spawn.Count)
		if count == 0 {
			count = 1
		}
		for copyIndex := 0; copyIndex < count; copyIndex++ {
			id := fmt.Sprintf("monster-%02X-%d", spawn.MonsterID, copyIndex+1)
			fighter := record.Fighter(id, combat.SideEnemy)
			for _, affect := range affects[spawn.MonsterID] {
				fighter.MonsterAffects = append(fighter.MonsterAffects, combat.MonsterAffect{
					Kind: affect.Kind, Value: affect.Value, Duration: affect.Duration,
					Strength: affect.Strength, Active: affect.Active, Data: affect.Data,
				})
			}
			fighter.AttacksPerTurn = fighter.MonsterAffectAttacksPerTurn()
			fighter.SpriteBlock = spawn.IconBlock
			enemies = append(enemies, fighter)
		}
	}
	return enemies, nil
}
