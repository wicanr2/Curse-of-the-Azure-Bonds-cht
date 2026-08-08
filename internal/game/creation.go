package game

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
	engineaction "github.com/wicanr2/golden-box-remake-engine/combat/action"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

func starterCharacters(pack *goldenbox.Pack, language string) ([]party.Character, error) {
	if pack == nil || pack.CharacterCreation == nil {
		return nil, fmt.Errorf("character creation templates are missing from game pack")
	}
	result := make([]party.Character, 0, len(pack.CharacterCreation.Templates))
	for index, template := range pack.CharacterCreation.Templates {
		if template.RaceID > uint8(party.RaceHalfOrc) || template.PrimaryClassID > uint8(party.ClassThief) {
			return nil, fmt.Errorf("character creation template %q has unsupported race/class IDs %d/%d", template.ID, template.RaceID, template.PrimaryClassID)
		}
		name, ok := pack.Text(template.DisplayID, language)
		if !ok || name == "" {
			return nil, fmt.Errorf("character creation template %q display %q is unavailable", template.ID, template.DisplayID)
		}
		var levels [8]uint8
		copy(levels[:], template.ClassLevels)
		abilities := template.BaseAbilities
		character := party.Character{
			ID: "creation." + template.ID, Name: name,
			Race: party.Race(template.RaceID), Class: party.Class(template.PrimaryClassID),
			RawClassID: template.RawClassID, Level: int(template.Level), ClassLevels: levels,
			Abilities: party.Abilities{
				Strength: abilities[0], Intelligence: abilities[1], Wisdom: abilities[2],
				Dexterity: abilities[3], Constitution: abilities[4], Charisma: abilities[5],
			},
		}
		if _, err := party.StartingAgeSpecFor(character.Race, character.Class); err != nil {
			return nil, fmt.Errorf("character creation template %q age: %w", template.ID, err)
		}
		if err := character.Validate(); err != nil {
			return nil, fmt.Errorf("character creation template %d %q: %w", index, template.ID, err)
		}
		result = append(result, character)
	}
	return result, nil
}

func (s *State) OpenCharacterCreation() error {
	if s.Mode == ModeCombat {
		return fmt.Errorf("character creation is unavailable during combat")
	}
	s.creationReturnMode = s.Mode
	options, err := starterCharacters(s.dataPack, s.catalog.Language)
	if err != nil {
		return err
	}
	s.CreationOptions = options
	s.CreationRoster = nil
	s.CreationCursor = 0
	s.CreationMessage = s.catalog.Text("creation_prompt", "creation_prompt")
	s.Mode = ModeCharacterCreation
	return nil
}

func (s *State) AddCreationCharacter(index int) error {
	if s.Mode != ModeCharacterCreation {
		return fmt.Errorf("character creation is not open")
	}
	if index < 0 || index >= len(s.CreationOptions) {
		return fmt.Errorf("character option %d is out of range", index)
	}
	if len(s.CreationRoster) >= 6 {
		return fmt.Errorf("party already has six characters")
	}
	character := s.CreationOptions[index]
	// The reference assigns age after race/class selection. Keep the template
	// values editable, then apply the generated age only to the copied party
	// character so an option can be previewed and reused without double-aging.
	seed := int64(0xC0AB) + int64(index)*97 + int64(len(s.CreationRoster))*13
	age, err := party.RollStartingAge(character.Race, character.Class, seed)
	if err != nil {
		return err
	}
	character.Age = age
	character.Abilities, err = character.Abilities.WithAgeEffects(character.Race, int(age))
	if err != nil {
		return err
	}
	character.ID = fmt.Sprintf("%s-%d", character.ID, len(s.CreationRoster)+1)
	s.CreationRoster = append(s.CreationRoster, character)
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_added", "creation_added"), character.Name, len(s.CreationRoster))
	return nil
}

func (s *State) BeginCreationName() error {
	if s.Mode != ModeCharacterCreation || s.CreationEditingAbilities || s.CreationCursor < 0 || s.CreationCursor >= len(s.CreationOptions) {
		return fmt.Errorf("character name editor is unavailable")
	}
	s.CreationName = ""
	s.CreationEditing = true
	return nil
}

func (s *State) ToggleCreationAbilities() error {
	if s.Mode != ModeCharacterCreation || s.CreationCursor < 0 || s.CreationCursor >= len(s.CreationOptions) {
		return fmt.Errorf("ability editor is unavailable")
	}
	s.CreationEditingAbilities = !s.CreationEditingAbilities
	s.CreationAbility = 0
	if s.CreationEditingAbilities {
		s.CreationMessage = s.catalog.Text("creation_ability_prompt", "creation_ability_prompt")
	}
	return nil
}

func (s *State) RerollCreationAbilities(seed int64) error {
	if s.Mode != ModeCharacterCreation || s.CreationCursor < 0 || s.CreationCursor >= len(s.CreationOptions) {
		return fmt.Errorf("ability reroll is unavailable")
	}
	s.CreationOptions[s.CreationCursor].Abilities = party.RollAbilities(seed)
	s.CreationMessage = s.catalog.Text("creation_rerolled", "creation_rerolled")
	return nil
}

func (s *State) MoveCreationAbility(delta int) error {
	if !s.CreationEditingAbilities {
		return fmt.Errorf("ability editor is not active")
	}
	s.CreationAbility += delta
	if s.CreationAbility < 0 {
		s.CreationAbility = 5
	}
	if s.CreationAbility > 5 {
		s.CreationAbility = 0
	}
	return nil
}

func (s *State) AdjustCreationAbility(delta int) error {
	if !s.CreationEditingAbilities || len(s.CreationOptions) == 0 {
		return fmt.Errorf("ability editor is not active")
	}
	if err := s.CreationOptions[s.CreationCursor].Abilities.Adjust(s.CreationAbility, delta); err != nil {
		s.CreationMessage = err.Error()
		return err
	}
	value, _ := s.CreationOptions[s.CreationCursor].Abilities.Value(s.CreationAbility)
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_ability_updated", "creation_ability_updated"), value)
	return nil
}

func (s *State) AppendCreationName(chars []rune) error {
	if !s.CreationEditing {
		return fmt.Errorf("character name editor is not active")
	}
	if utf8.RuneCountInString(s.CreationName)+len(chars) > 20 {
		return fmt.Errorf("character name is limited to 20 characters")
	}
	s.CreationName += string(chars)
	return nil
}

func (s *State) BackspaceCreationName() error {
	if !s.CreationEditing {
		return fmt.Errorf("character name editor is not active")
	}
	name := []rune(s.CreationName)
	if len(name) > 0 {
		s.CreationName = string(name[:len(name)-1])
	}
	return nil
}

func (s *State) CommitCreationName() error {
	if !s.CreationEditing {
		return fmt.Errorf("character name editor is not active")
	}
	if s.CreationName == "" {
		return fmt.Errorf("character name cannot be empty")
	}
	s.CreationOptions[s.CreationCursor].Name = s.CreationName
	s.CreationEditing = false
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_named", "creation_named"), s.CreationName)
	return nil
}

func (s *State) CancelCreationName() error {
	if !s.CreationEditing {
		return fmt.Errorf("character name editor is not active")
	}
	s.CreationEditing = false
	s.CreationName = ""
	return nil
}

func (s *State) FinishCharacterCreation() error {
	if s.Mode != ModeCharacterCreation {
		return fmt.Errorf("character creation is not open")
	}
	if err := s.CreationRoster.Validate(); err != nil {
		return err
	}
	fighters := make([]combat.Fighter, 0, len(s.CreationRoster))
	for _, character := range s.CreationRoster {
		fighter, err := s.fighterForCharacter(character)
		if err != nil {
			return err
		}
		fighters = append(fighters, fighter)
	}
	if err := s.SetParty(fighters); err != nil {
		return err
	}
	s.partyRoster = append(party.Roster(nil), s.CreationRoster...)
	s.CreationMessage = ""
	if s.session == nil {
		// Data-neutral tests/embedders may construct State without the
		// original image. Production NewStateFromECLBlocks always takes the
		// verified block-0x01 path below.
		s.Mode = ModeWilderness
		s.Prompt = s.catalog.Text("party_ready", "party_ready")
		s.Choices = []string{s.catalog.Text("enter_city", "enter_city"), s.catalog.Text("journey_on", "journey_on"), s.catalog.Text("camp", "camp")}
		s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
		return nil
	}
	return s.BeginAdventure()
}

// LocaleText exposes the State-owned catalog to the frontend without letting
// renderer code maintain a second translation table.
func (s *State) LocaleText(key string) string { return s.catalog.Text(key, key) }

func (s *State) CharacterRaceName(race party.Race) string {
	key := "race_unknown"
	switch race {
	case party.RaceDwarf:
		key = "race_dwarf"
	case party.RaceElf:
		key = "race_elf"
	case party.RaceGnome:
		key = "race_gnome"
	case party.RaceHalfElf:
		key = "race_half_elf"
	case party.RaceHalfling:
		key = "race_halfling"
	case party.RaceHuman:
		key = "race_human"
	case party.RaceHalfOrc:
		key = "race_half_orc"
	}
	return s.LocaleText(key)
}

func (s *State) CharacterClassName(class party.Class) string {
	return s.localizedCharacterClassName(class)
}

func (s *State) SavePartyFile(path string) error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("no character-created party is available to save")
	}
	areaState := s.Area
	areaState.GameTime = s.gameClock
	var sessionSnapshot *ecl.SessionSnapshot
	if s.session != nil {
		snapshot, snapshotErr := s.session.Snapshot()
		if snapshotErr != nil {
			return snapshotErr
		}
		sessionSnapshot = &snapshot
	}
	combatSnapshot, err := s.activeCombatSnapshot()
	if err != nil {
		return err
	}
	data, err := partySave.EncodeGameWithJournalState(s.partyRoster, areaState, uint8(s.Mode), uint8(s.Location), s.MapX, s.MapY, s.DungeonX, s.DungeonY, s.DungeonDirection, s.DungeonWallType, s.DungeonWallRoof, s.gameClock, s.gameAgeCycles, sessionSnapshot, combatSnapshot, s.musicSnapshot(), s.oneShotSnapshot(), s.journalMessageIDs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (s *State) activeCombatSnapshot() (*partySave.CombatSnapshot, error) {
	if s.battle == nil {
		if s.Mode == ModeCombat {
			return nil, fmt.Errorf("combat mode has no active battle to save")
		}
		return nil, nil
	}
	if s.Mode != ModeCombat {
		return nil, fmt.Errorf("active battle cannot be saved from mode %d", s.Mode)
	}
	battleSnapshot, err := s.battle.Snapshot()
	if err != nil {
		return nil, err
	}
	var visual *combat.VisualEvent
	if s.combatVisual != nil {
		copy := *s.combatVisual
		copy.Impacts = append([]combat.VisualImpactTarget(nil), copy.Impacts...)
		copy.Segments = append([]combat.VisualPathSegment(nil), copy.Segments...)
		visual = &copy
	}
	return &partySave.CombatSnapshot{
		Battle: battleSnapshot, Turns: append([]combat.Turn(nil), s.combatTurns...),
		TurnIndex: s.combatTurnIndex, DelayedTurns: cloneDelayedTurns(s.combatDelayedTurns),
		TargetIndex: s.combatTargetIndex, CastingSpell: s.combatCastingSpell,
		CastingClass: uint8(s.combatCastingClass), CastingClassSet: s.combatCastingClassSet,
		SpellTargetIndex: s.combatSpellTargetIndex, SpellTargetPoint: s.combatSpellTargetPoint,
		SpellTargetsPoint: s.combatSpellTargetsPoint, MoveMode: s.combatMoveMode,
		MoveRemaining: s.combatMoveRemaining, Speed: uint8(s.combatSpeed),
		QuickMagic: s.combatQuickMagic, ReferenceCoords: s.combatReferenceCoords,
		View: s.combatView, ViewFighterID: s.combatViewFighterID,
		Message: s.combatMessage, ReturnMode: uint8(s.combatReturnMode),
		VisualSerial: s.combatVisualSerial, VisualEnabled: s.combatVisualEnabled,
		Visual: visual, VisualElapsedNanos: int64(s.combatVisualElapsed),
		VisualTravelSent: s.combatVisualTravelSent,
		VisualImpactSent: s.combatVisualImpactSent, VisualDeathSent: s.combatVisualDeathSent,
		VisualAdvanceTurn: s.combatVisualAdvanceTurn,
	}, nil
}

func cloneDelayedTurns(source map[int]bool) map[int]bool {
	if source == nil {
		return nil
	}
	result := make(map[int]bool, len(source))
	for index, delayed := range source {
		result[index] = delayed
	}
	return result
}

// SaveSAVGAMPrefix writes only the reference-compatible fixed SAVGAM prefix
// while preserving raw records loaded previously. SaveSAVGAMSlot is the
// higher-level API for a loaded prefix plus CHRDAT player sidecars.
func (s *State) SaveSAVGAMPrefix(path string) error {
	container, err := s.savgamContainerForSave()
	if err != nil {
		return err
	}
	data, err := partySave.EncodeSAVGAM(container)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadSAVGAMPrefix imports the fixed binary prefix and applies only fields
// whose Area/map meaning is already proven. Character CHRDAT player files
// remain a separate import step and are not synthesized from name records.
func (s *State) LoadSAVGAMPrefix(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	container, err := partySave.DecodeSAVGAM(data)
	if err != nil {
		return err
	}
	area1, err := area.DecodeArea1(container.Area1)
	if err != nil {
		return err
	}
	area2, err := area.DecodeArea2(container.Area2)
	if err != nil {
		return err
	}
	area1.GameArea = area2.GameArea
	area1.HeadBlockID = area2.HeadBlockID
	s.Area = area1
	s.gameClock = area1.GameTime
	s.GeoMapSet = area1.GameArea
	s.GeoMapBlock = area1.Current3DMapBlockID
	s.MapX = int(container.MapPosX)
	s.MapY = int(container.MapPosY)
	s.DungeonDirection = container.MapDirection
	s.DungeonWallType = container.MapWallType
	s.DungeonWallRoof = container.MapWallRoof
	s.savgamPrefix = &container
	return nil
}

// LoadSAVGAMSlot loads the reference save-game prefix plus its CHRDAT player
// bundles from a directory. SaveGame writes the prefix as savgamA..J.dat and
// names party records CHRDAT{key}{1..6}.sav; .swg/.fx are optional sidecars.
// This loader intentionally does not infer unsupported multi-class fields or
// rewrite the original files.
func (s *State) LoadSAVGAMSlot(directory string, key byte) error {
	if key < 'A' || key > 'J' {
		return fmt.Errorf("SAVGAM slot key %q is outside A..J", key)
	}
	prefixPath := filepath.Join(directory, fmt.Sprintf("savgam%c.dat", key+('a'-'A')))
	if err := s.LoadSAVGAMPrefix(prefixPath); err != nil {
		return err
	}
	container := s.savgamPrefix
	if container == nil {
		return fmt.Errorf("SAVGAM prefix was not retained")
	}
	roster := make(party.Roster, 0, container.PartyCount)
	for index := 0; index < int(container.PartyCount); index++ {
		base := fmt.Sprintf("CHRDAT%c%d", key, index+1)
		recordPath := filepath.Join(directory, base+".sav")
		record, err := os.ReadFile(recordPath)
		if err != nil {
			return fmt.Errorf("SAVGAM player %s: %w", base, err)
		}
		effects, err := readOptionalSAVGAMSidecar(filepath.Join(directory, base+".fx"))
		if err != nil {
			return err
		}
		inventory, err := readOptionalSAVGAMSidecar(filepath.Join(directory, base+".swg"))
		if err != nil {
			return err
		}
		character, err := party.ParseDOSPlayerFiles(base, party.DOSPlayerFiles{Record: record, Effects: effects, Inventory: inventory})
		if err != nil {
			return fmt.Errorf("SAVGAM player %s: %w", base, err)
		}
		if s.savgamPlayers == nil {
			s.savgamPlayers = make(map[string]party.DOSPlayerFiles)
		}
		s.savgamPlayers[base] = party.DOSPlayerFiles{
			Record: append([]byte(nil), record...), Effects: append([]byte(nil), effects...), Inventory: append([]byte(nil), inventory...),
		}
		roster = append(roster, character)
	}
	if len(roster) == 0 {
		return fmt.Errorf("SAVGAM slot %c has no player records", key)
	}
	fighters := make([]combat.Fighter, 0, len(roster))
	for _, character := range roster {
		fighter, err := s.fighterForCharacter(character)
		if err != nil {
			return err
		}
		fighters = append(fighters, fighter)
	}
	if err := s.SetParty(fighters); err != nil {
		return err
	}
	s.partyRoster = roster
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("party_ready", "party_ready")
	s.Choices = []string{s.catalog.Text("enter_city", "enter_city"), s.catalog.Text("journey_on", "journey_on"), s.catalog.Text("camp", "camp")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	return nil
}

// SaveSAVGAMSlot writes the current party back to a reference-style slot.
// Only documented player offsets and decoded SWG/FX records are serialized;
// every unknown byte in each original .sav record is preserved. Files are
// prepared in a sibling staging directory before replacement, so a failed
// encode never leaves a partially written staged bundle behind.
func (s *State) SaveSAVGAMSlot(directory string, key byte) error {
	if key >= 'a' && key <= 'j' {
		key -= 'a' - 'A'
	}
	if key < 'A' || key > 'J' {
		return fmt.Errorf("SAVGAM slot key %q is outside A..J", key)
	}
	if s.savgamPrefix == nil {
		return fmt.Errorf("no SAVGAM prefix is loaded")
	}
	if len(s.partyRoster) == 0 || len(s.partyRoster) > 6 {
		return fmt.Errorf("SAVGAM party size %d is outside 1..6", len(s.partyRoster))
	}
	if s.savgamPlayers == nil {
		return fmt.Errorf("no raw SAVGAM player records are retained")
	}

	container, err := s.savgamContainerForSave()
	if err != nil {
		return err
	}
	for index := range container.CharacterRefs {
		container.CharacterRefs[index] = nil
	}
	for index := range s.partyRoster {
		container.CharacterRefs[index] = []byte(fmt.Sprintf("CHRDAT%c%d", key, index+1))
	}
	prefix, err := partySave.EncodeSAVGAM(container)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(directory, ".savgam-save-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	write := func(name string, data []byte) error {
		return os.WriteFile(filepath.Join(stage, name), data, 0o600)
	}
	prefixName := fmt.Sprintf("savgam%c.dat", key+('a'-'A'))
	if err := write(prefixName, prefix); err != nil {
		return err
	}
	for index, character := range s.partyRoster {
		base := fmt.Sprintf("CHRDAT%c%d", key, index+1)
		raw, ok := s.savgamPlayers[character.ID]
		if !ok {
			return fmt.Errorf("no raw SAVGAM player record retained for %q", character.ID)
		}
		record, err := party.PatchDOSPlayerRecord(raw.Record, character)
		if err != nil {
			return fmt.Errorf("patch SAVGAM player %s: %w", base, err)
		}
		inventory, err := monster.EncodeItems(character.Equipment)
		if err != nil {
			return fmt.Errorf("encode SAVGAM inventory %s: %w", base, err)
		}
		if err := write(base+".sav", record); err != nil {
			return err
		}
		if err := write(base+".swg", inventory); err != nil {
			return err
		}
		if err := write(base+".fx", monster.EncodeAffects(character.Effects)); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(stage)
	if err != nil {
		return err
	}
	backup, err := os.MkdirTemp(directory, ".savgam-backup-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(backup)
	fileNames := []string{prefixName}
	for index := 1; index <= 6; index++ {
		base := fmt.Sprintf("CHRDAT%c%d", key, index)
		fileNames = append(fileNames, base+".sav", base+".swg", base+".fx")
	}
	backedUp := make([]string, 0, len(fileNames))
	for _, name := range fileNames {
		target := filepath.Join(directory, name)
		if _, statErr := os.Stat(target); os.IsNotExist(statErr) {
			continue
		} else if statErr != nil {
			return statErr
		}
		if err := os.Rename(target, filepath.Join(backup, name)); err != nil {
			for _, restored := range backedUp {
				_ = os.Rename(filepath.Join(backup, restored), filepath.Join(directory, restored))
			}
			return fmt.Errorf("backup SAVGAM file %s: %w", name, err)
		}
		backedUp = append(backedUp, name)
	}
	installed := make([]string, 0, len(entries))
	rollback := func() {
		for _, name := range installed {
			_ = os.Remove(filepath.Join(directory, name))
		}
		for _, name := range backedUp {
			_ = os.Rename(filepath.Join(backup, name), filepath.Join(directory, name))
		}
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(stage, entry.Name()), filepath.Join(directory, entry.Name())); err != nil {
			rollback()
			return fmt.Errorf("replace SAVGAM file %s: %w", entry.Name(), err)
		}
		installed = append(installed, entry.Name())
	}
	s.savgamPrefix = &container
	return nil
}

func readOptionalSAVGAMSidecar(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *State) savgamContainerForSave() (partySave.SAVGAMContainer, error) {
	var container partySave.SAVGAMContainer
	var err error
	if s.savgamPrefix != nil {
		container = *s.savgamPrefix
		container.Area1 = append([]byte(nil), s.savgamPrefix.Area1...)
		container.Area2 = append([]byte(nil), s.savgamPrefix.Area2...)
		container.Runtime = append([]byte(nil), s.savgamPrefix.Runtime...)
		container.ECL = append([]byte(nil), s.savgamPrefix.ECL...)
	}
	container.GameArea = s.Area.GameArea
	container.Area1, err = area.EncodeArea1(s.Area, container.Area1)
	if err != nil {
		return partySave.SAVGAMContainer{}, err
	}
	container.Area2, err = area.EncodeArea2(s.Area, container.Area2)
	if err != nil {
		return partySave.SAVGAMContainer{}, err
	}
	if container.Runtime == nil {
		container.Runtime = make([]byte, partySave.SAVGAMRuntimeStateSize)
	}
	if container.ECL == nil {
		container.ECL = make([]byte, partySave.SAVGAMECLMemorySize)
	}
	if s.MapX < -128 || s.MapX > 127 || s.MapY < -128 || s.MapY > 127 {
		return partySave.SAVGAMContainer{}, fmt.Errorf("SAVGAM map position out of signed-byte range: (%d,%d)", s.MapX, s.MapY)
	}
	container.MapPosX = int8(s.MapX)
	container.MapPosY = int8(s.MapY)
	container.MapDirection = s.DungeonDirection
	container.MapWallType = s.DungeonWallType
	container.MapWallRoof = s.DungeonWallRoof
	container.PartyCount = uint8(len(s.partyRoster))
	for index := range container.CharacterRefs {
		container.CharacterRefs[index] = nil
	}
	for index, character := range s.partyRoster {
		if index >= len(container.CharacterRefs) {
			break
		}
		container.CharacterRefs[index] = []byte(character.Name)
	}
	return container, nil
}

func (s *State) LoadPartyFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	file, err := partySave.DecodeGame(data)
	if err != nil {
		return err
	}
	if err := s.SetPartyRoster(file.Characters); err != nil {
		return err
	}
	if err := s.restoreJournalMessageIDs(file.JournalMessageIDs); err != nil {
		return err
	}
	if file.Version >= 5 {
		s.gameClock = file.GameTime
		s.gameAgeCycles = file.GameAgeCycles
	} else {
		s.gameClock = [7]uint16{}
		s.gameAgeCycles = 0
	}
	s.Area = file.Area
	s.gameClock = file.GameTime
	if file.GameTime == [7]uint16{} {
		s.gameClock = file.Area.GameTime
	}
	s.Area.GameTime = s.gameClock
	s.GeoMapSet = file.Area.GameArea
	s.GeoMapBlock = file.Area.Current3DMapBlockID
	s.MapX, s.MapY = file.MapX, file.MapY
	if file.Version >= 3 {
		s.DungeonX, s.DungeonY, s.DungeonDirection = file.DungeonX, file.DungeonY, file.DungeonDir
		if file.Version >= 4 {
			s.DungeonWallType, s.DungeonWallRoof = file.DungeonWallType, file.DungeonWallRoof
		} else {
			s.DungeonWallType, s.DungeonWallRoof = 0, 0
		}
		if s.DungeonX < 0 || s.DungeonX >= 16 || s.DungeonY < 0 || s.DungeonY >= 16 || s.DungeonDirection >= 8 {
			s.DungeonX, s.DungeonY, s.DungeonDirection = 7, 13, 0
			s.DungeonWallType, s.DungeonWallRoof = 0, 0
		}
	} else {
		s.DungeonX, s.DungeonY, s.DungeonDirection = 7, 13, 0
		s.DungeonWallType, s.DungeonWallRoof = 0, 0
	}
	if file.Location <= uint8(LocationMythDrannor) {
		s.Location = Location(file.Location)
	}
	if file.Version == 1 {
		// Legacy party.json had no adventure-state fields.
		s.Mode = ModeWilderness
		s.Location = LocationWilderness
	} else if file.Mode <= uint8(ModeDungeon) {
		s.Mode = Mode(file.Mode)
	} else {
		s.Mode = ModeWilderness
	}
	if s.Mode == ModeMap && s.Location != LocationShadowdale {
		s.Mode = ModeWilderness
	}
	if file.Version >= 6 && file.ECLSession != nil {
		if s.session == nil {
			return fmt.Errorf("game save contains an ECL session but no original ECL blocks are loaded")
		}
		if err := s.session.RestoreSnapshot(*file.ECLSession); err != nil {
			return fmt.Errorf("restore game ECL session: %w", err)
		}
		s.eclBlock = s.session.CurrentData()
		s.eclStart, err = s.session.InitialEntry()
		if err != nil {
			return err
		}
	}
	s.activeMusicTrackID = ""
	s.musicPlaybackSnapshot = nil
	s.oneShotPlaybackSnapshot = nil
	s.pendingMusicEvents = nil
	if file.OneShotAudio != nil {
		copy := audiostate.Clone(*file.OneShotAudio)
		s.oneShotPlaybackSnapshot = &copy
	}
	if file.Music != nil {
		if s.dataPack != nil {
			track, found := s.dataPack.FindMusicTrack(file.Music.TrackID)
			if !found {
				return fmt.Errorf("game save music track %q is not in the data pack", file.Music.TrackID)
			}
			if file.Music.Stream != nil && file.Music.Stream.Selector != int(track.ReferenceSelector) {
				return fmt.Errorf("game save music track %q selector %d does not match data-pack selector %d", file.Music.TrackID, file.Music.Stream.Selector, track.ReferenceSelector)
			}
		}
		s.activeMusicTrackID = file.Music.TrackID
		if file.Music.Stream != nil {
			copy := cloneTrackPCMStreamSnapshot(*file.Music.Stream)
			s.musicPlaybackSnapshot = &copy
		}
	} else if file.Version < 8 {
		// Legacy saves never recorded an active track. Restart the verified
		// binding for the restored ECL block instead of preserving a stale
		// pre-load frontend event or pretending an unknown sample position.
		if s.dataPack != nil && s.session != nil {
			if binding, found := s.dataPack.FindMusicBinding(s.session.CurrentBlockID(), ""); found {
				s.activeMusicTrackID = binding.TrackID
			}
		}
	}
	if file.Combat != nil {
		if s.Mode != ModeCombat {
			return fmt.Errorf("game save contains active combat while mode is %d", s.Mode)
		}
		if err := s.restoreActiveCombat(*file.Combat); err != nil {
			return fmt.Errorf("restore active combat: %w", err)
		}
		return nil
	}
	if s.Mode == ModeCombat {
		return fmt.Errorf("game save mode is combat but no active-combat snapshot is present")
	}
	s.Prompt = s.catalog.Text("party_ready", "party_ready")
	s.Choices = []string{s.catalog.Text("enter_city", "enter_city"), s.catalog.Text("journey_on", "journey_on"), s.catalog.Text("camp", "camp")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	return nil
}

func (s *State) restoreJournalMessageIDs(messageIDs []string) error {
	s.journalMessageIDs = nil
	s.JournalPages = nil
	s.JournalPage = 0
	if len(messageIDs) == 0 {
		s.JournalText = s.catalog.Text("journal_empty", "journal_empty")
		return nil
	}
	if s.dataPack == nil {
		return fmt.Errorf("game save contains journal entries but no game pack is loaded")
	}
	for _, messageID := range messageIDs {
		page, found := s.dataPack.Text(messageID, s.catalog.Language)
		if !found {
			return fmt.Errorf("game save journal message ID %q is not in the data pack", messageID)
		}
		s.appendJournalPage(messageID, page)
	}
	s.JournalText = s.JournalPages[0]
	return nil
}

func (s *State) restoreActiveCombat(snapshot partySave.CombatSnapshot) error {
	for index := range snapshot.Battle.Fighters {
		fighter := &snapshot.Battle.Fighters[index]
		if fighter.SourceName == "" || s.dataPack == nil {
			continue
		}
		if localized, found := s.dataPack.LocalizeCombatantName(fighter.SourceName, s.catalog.Language); found {
			fighter.Name = localized
		} else {
			fighter.Name = fighter.SourceName
		}
	}
	battle, err := combat.RestoreBattle(snapshot.Battle)
	if err != nil {
		return err
	}
	if s.dataPack != nil {
		rules, err := s.dataPack.ResolveCombatAffectRules()
		if err != nil {
			return fmt.Errorf("resolve combat affect rules after load: %w", err)
		}
		battle.SetDamageRules(rules)
		conditionalRules, err := s.dataPack.ResolveCombatConditionalModifiers()
		if err != nil {
			return fmt.Errorf("resolve combat conditional modifiers after load: %w", err)
		}
		battle.SetConditionalModifierRules(conditionalRules)
		magicResistanceRules, err := s.dataPack.ResolveCombatMagicResistanceRules()
		if err != nil {
			return fmt.Errorf("resolve combat magic resistance rules after load: %w", err)
		}
		battle.SetMagicResistanceRules(magicResistanceRules)
		postHitRules, err := s.dataPack.ResolveCombatPostHitRules()
		if err != nil {
			return fmt.Errorf("resolve combat post-hit rules after load: %w", err)
		}
		battle.SetPostHitRules(postHitRules)
		monsterSpellRules, err := s.dataPack.ResolveCombatMonsterSpellRules()
		if err != nil {
			return fmt.Errorf("resolve combat monster spell rules after load: %w", err)
		}
		battle.SetMonsterSpellRules(monsterSpellRules)
	}
	if snapshot.TurnIndex < 0 || snapshot.TurnIndex > len(snapshot.Turns) {
		return fmt.Errorf("combat turn index %d outside 0..%d", snapshot.TurnIndex, len(snapshot.Turns))
	}
	for index, turn := range snapshot.Turns {
		if _, ok := battle.Fighter(turn.FighterID); !ok {
			return fmt.Errorf("combat turn %d references missing fighter %q", index, turn.FighterID)
		}
	}
	for index, delayed := range snapshot.DelayedTurns {
		if index < 0 || index >= len(snapshot.Turns) || !delayed {
			return fmt.Errorf("invalid delayed-turn entry %d=%v", index, delayed)
		}
	}
	if snapshot.TargetIndex < 0 || snapshot.SpellTargetIndex < 0 || snapshot.MoveRemaining < 0 {
		return fmt.Errorf("negative combat cursor or movement state")
	}
	if snapshot.CastingClass > uint8(party.ClassThief) || snapshot.Speed > uint8(engineaction.SlowestSpeed) {
		return fmt.Errorf("invalid combat class %d or speed %d", snapshot.CastingClass, snapshot.Speed)
	}
	if snapshot.ReturnMode > uint8(ModeDungeon) || snapshot.ReturnMode == uint8(ModeCombat) {
		return fmt.Errorf("invalid combat return mode %d", snapshot.ReturnMode)
	}
	if snapshot.ViewFighterID != "" {
		if _, ok := battle.Fighter(snapshot.ViewFighterID); !ok {
			return fmt.Errorf("combat view references missing fighter %q", snapshot.ViewFighterID)
		}
	}
	if snapshot.Visual != nil {
		elapsed := time.Duration(snapshot.VisualElapsedNanos)
		if snapshot.VisualElapsedNanos < 0 || elapsed > snapshot.Visual.Duration() {
			return fmt.Errorf("combat visual elapsed time %d outside 0..%d", snapshot.VisualElapsedNanos, snapshot.Visual.Duration())
		}
		if snapshot.VisualImpactSent < -1 || snapshot.VisualImpactSent >= snapshot.Visual.ImpactCount() {
			return fmt.Errorf("combat visual impact marker %d outside -1..%d", snapshot.VisualImpactSent, snapshot.Visual.ImpactCount()-1)
		}
		if snapshot.VisualDeathSent < -1 || snapshot.VisualDeathSent >= snapshot.Visual.ImpactCount() {
			return fmt.Errorf("combat visual death marker %d outside -1..%d", snapshot.VisualDeathSent, snapshot.Visual.ImpactCount()-1)
		}
		if snapshot.Visual.ActorID != "" {
			if _, ok := battle.Fighter(snapshot.Visual.ActorID); !ok {
				return fmt.Errorf("combat visual references missing actor %q", snapshot.Visual.ActorID)
			}
		}
		for index := 0; index < snapshot.Visual.ImpactCount(); index++ {
			impact, _ := snapshot.Visual.ImpactAt(index)
			if impact.TargetID == "" {
				continue
			}
			if _, ok := battle.Fighter(impact.TargetID); !ok {
				return fmt.Errorf("combat visual impact %d references missing target %q", index, impact.TargetID)
			}
		}
	} else if snapshot.VisualElapsedNanos != 0 {
		return fmt.Errorf("combat visual elapsed time %d has no visual event", snapshot.VisualElapsedNanos)
	}
	s.battle = battle
	s.combatTurns = append([]combat.Turn(nil), snapshot.Turns...)
	s.combatTurnIndex = snapshot.TurnIndex
	s.combatDelayedTurns = cloneDelayedTurns(snapshot.DelayedTurns)
	s.combatTargetIndex = snapshot.TargetIndex
	s.combatCastingSpell = snapshot.CastingSpell
	s.combatCastingClass = party.Class(snapshot.CastingClass)
	s.combatCastingClassSet = snapshot.CastingClassSet
	s.combatSpellTargetIndex = snapshot.SpellTargetIndex
	s.combatSpellTargetPoint = snapshot.SpellTargetPoint
	s.combatSpellTargetsPoint = snapshot.SpellTargetsPoint
	s.combatMoveMode, s.combatMoveRemaining = snapshot.MoveMode, snapshot.MoveRemaining
	s.combatSpeed = engineaction.Speed(snapshot.Speed)
	s.combatQuickMagic, s.combatReferenceCoords = snapshot.QuickMagic, snapshot.ReferenceCoords
	s.combatView, s.combatViewFighterID = snapshot.View, snapshot.ViewFighterID
	s.combatMessage, s.combatReturnMode = snapshot.Message, Mode(snapshot.ReturnMode)
	s.combatVisualSerial, s.combatVisualEnabled = snapshot.VisualSerial, snapshot.VisualEnabled
	if snapshot.Visual != nil {
		copy := *snapshot.Visual
		copy.Impacts = append([]combat.VisualImpactTarget(nil), copy.Impacts...)
		copy.Segments = append([]combat.VisualPathSegment(nil), copy.Segments...)
		s.combatVisual = &copy
	} else {
		s.combatVisual = nil
	}
	s.combatVisualElapsed = time.Duration(snapshot.VisualElapsedNanos)
	s.combatVisualTravelSent = snapshot.VisualTravelSent
	s.combatVisualImpactSent = snapshot.VisualImpactSent
	s.combatVisualDeathSent = snapshot.VisualDeathSent
	s.combatVisualAdvanceTurn = snapshot.VisualAdvanceTurn
	return nil
}

// LoadDOSCharacterFiles imports one original DOS character bundle directly
// into the running remake. It remains useful for inspecting an individual
// record without loading a complete SAVGAM slot.
func (s *State) LoadDOSCharacterFiles(id string, files party.DOSPlayerFiles) error {
	character, err := party.ParseDOSPlayerFiles(id, files)
	if err != nil {
		return err
	}
	fighter, err := s.fighterForCharacter(character)
	if err != nil {
		return err
	}
	if err := s.SetParty([]combat.Fighter{fighter}); err != nil {
		return err
	}
	s.partyRoster = party.Roster{character}
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("party_ready", "party_ready")
	s.Choices = []string{s.catalog.Text("enter_city", "enter_city"), s.catalog.Text("journey_on", "journey_on"), s.catalog.Text("camp", "camp")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	return nil
}

func (s *State) CancelCharacterCreation() error {
	if s.Mode != ModeCharacterCreation {
		return fmt.Errorf("character creation is not open")
	}
	s.Mode = s.creationReturnMode
	s.CreationEditing = false
	s.CreationEditingAbilities = false
	return nil
}
