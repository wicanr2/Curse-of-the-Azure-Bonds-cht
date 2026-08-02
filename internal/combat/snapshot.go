package combat

import (
	"fmt"

	engineinitiative "github.com/wicanr2/golden-box-remake-engine/combat/initiative"
	enginerandom "github.com/wicanr2/golden-box-remake-engine/randomstream"
)

const BattleSnapshotVersion = 1

// BattleSnapshot is the renderer-neutral active-combat continuation. It owns
// mutable rules state only; title map callbacks and decoded source assets are
// reconstructed by the game adapter after load.
type BattleSnapshot struct {
	Version             int                        `json:"version"`
	Fighters            []Fighter                  `json:"fighters"`
	PartyAttackModifier int                        `json:"party_attack_modifier"`
	EnemyAttackModifier int                        `json:"enemy_attack_modifier"`
	Random              enginerandom.Snapshot      `json:"random"`
	Round               int                        `json:"round"`
	Status              Status                     `json:"status"`
	Areas               []PersistentArea           `json:"persistent_areas,omitempty"`
	NextArea            uint64                     `json:"next_area"`
	SchedulerEntries    []engineinitiative.Entry   `json:"scheduler_entries,omitempty"`
	SchedulerSelection  engineinitiative.Selection `json:"scheduler_selection,omitempty"`
	SchedulerSelected   bool                       `json:"scheduler_selected"`
	SpellInterruptions  []SpellInterruption        `json:"spell_interruptions,omitempty"`
}

func (b *Battle) Snapshot() (BattleSnapshot, error) {
	if b == nil || b.rngStream == nil {
		return BattleSnapshot{}, fmt.Errorf("battle random stream is unavailable")
	}
	fighters := make([]Fighter, 0, len(b.fighterOrder))
	for _, id := range b.fighterOrder {
		fighter, ok := b.fighters[id]
		if !ok {
			return BattleSnapshot{}, fmt.Errorf("battle order references missing fighter %q", id)
		}
		fighter.MonsterAffects = append([]MonsterAffect(nil), fighter.MonsterAffects...)
		fighter.SavingThrows = append([]uint8(nil), fighter.SavingThrows...)
		fighter.MonsterSpellIDs = append([]uint8(nil), fighter.MonsterSpellIDs...)
		fighters = append(fighters, fighter)
	}
	snapshot := BattleSnapshot{
		Version: BattleSnapshotVersion, Fighters: fighters,
		PartyAttackModifier: b.attackRollModifier[SideParty],
		EnemyAttackModifier: b.attackRollModifier[SideEnemy],
		Random:              b.rngStream.Snapshot(), Round: b.round, Status: b.status,
		Areas: clonePersistentAreas(b.areas), NextArea: b.nextArea,
		SchedulerSelection: b.initiativeSelection,
		SchedulerSelected:  b.initiativeSelected,
		SpellInterruptions: append([]SpellInterruption(nil), b.spellInterruptions...),
	}
	if b.initiativeScheduler != nil {
		snapshot.SchedulerEntries = b.initiativeScheduler.Entries()
	}
	return snapshot, nil
}

func RestoreBattle(snapshot BattleSnapshot) (*Battle, error) {
	if snapshot.Version != BattleSnapshotVersion {
		return nil, fmt.Errorf("unsupported battle snapshot version %d", snapshot.Version)
	}
	if snapshot.Round < 0 {
		return nil, fmt.Errorf("battle round %d is negative", snapshot.Round)
	}
	if snapshot.Status > StatusDraw {
		return nil, fmt.Errorf("unsupported battle status %d", snapshot.Status)
	}
	battle, err := NewBattle(snapshot.Fighters, snapshot.Random.Seed)
	if err != nil {
		return nil, err
	}
	stream, err := enginerandom.Restore(snapshot.Random)
	if err != nil {
		return nil, fmt.Errorf("restore battle random stream: %w", err)
	}
	battle.rngStream, battle.rng = stream, stream.Rand()
	battle.round = snapshot.Round
	battle.updateStatus()
	if battle.status != snapshot.Status {
		return nil, fmt.Errorf("battle status %d contradicts fighter state %d", snapshot.Status, battle.status)
	}
	battle.attackRollModifier[SideParty] = snapshot.PartyAttackModifier
	battle.attackRollModifier[SideEnemy] = snapshot.EnemyAttackModifier
	battle.areas = clonePersistentAreas(snapshot.Areas)
	battle.nextArea = snapshot.NextArea
	areaIDs := make(map[uint64]bool, len(battle.areas))
	for _, area := range battle.areas {
		if area.ID == 0 || area.ID > battle.nextArea {
			return nil, fmt.Errorf("persistent area ID %d outside 1..%d", area.ID, battle.nextArea)
		}
		if _, ok := battle.fighters[area.CasterID]; !ok {
			return nil, fmt.Errorf("persistent area %d references missing caster %q", area.ID, area.CasterID)
		}
		if areaIDs[area.ID] {
			return nil, fmt.Errorf("duplicate persistent area ID %d", area.ID)
		}
		areaIDs[area.ID] = true
	}
	if snapshot.SchedulerSelected && len(snapshot.SchedulerEntries) == 0 {
		return nil, fmt.Errorf("selected initiative action has no scheduler entries")
	}
	if len(snapshot.SchedulerEntries) != 0 {
		orderIndex := 0
		seen := make(map[string]bool, len(snapshot.SchedulerEntries))
		for index, entry := range snapshot.SchedulerEntries {
			if entry.ActionDelay < 0 || entry.ActionDelay > 20 {
				return nil, fmt.Errorf("scheduler entry %d delay %d outside 0..20", index, entry.ActionDelay)
			}
			if seen[entry.ID] {
				return nil, fmt.Errorf("duplicate scheduler fighter %q", entry.ID)
			}
			seen[entry.ID] = true
			for orderIndex < len(battle.fighterOrder) && battle.fighterOrder[orderIndex] != entry.ID {
				orderIndex++
			}
			if orderIndex == len(battle.fighterOrder) {
				return nil, fmt.Errorf("scheduler entry %d ID %q is missing or out of fighter order", index, entry.ID)
			}
			orderIndex++
		}
		battle.initiativeScheduler = engineinitiative.NewScheduler(snapshot.SchedulerEntries)
	}
	if snapshot.SchedulerSelected {
		selection := snapshot.SchedulerSelection
		if selection.Index < 0 || selection.Index >= len(snapshot.SchedulerEntries) ||
			snapshot.SchedulerEntries[selection.Index].ID != selection.ID ||
			snapshot.SchedulerEntries[selection.Index].ActionDelay != selection.ActionDelay ||
			selection.TieRoll < 1 || selection.TieRoll > 100 {
			return nil, fmt.Errorf("invalid selected initiative action %+v", selection)
		}
		battle.initiativeSelection, battle.initiativeSelected = selection, true
	}
	for _, interruption := range snapshot.SpellInterruptions {
		if _, ok := battle.fighters[interruption.FighterID]; !ok || interruption.SpellID == 0 {
			return nil, fmt.Errorf("invalid spell interruption %+v", interruption)
		}
	}
	battle.spellInterruptions = append([]SpellInterruption(nil), snapshot.SpellInterruptions...)
	return battle, nil
}

func clonePersistentAreas(source []PersistentArea) []PersistentArea {
	result := append([]PersistentArea(nil), source...)
	for index := range result {
		result[index].Cells = append([]PersistentAreaCell(nil), result[index].Cells...)
	}
	return result
}
