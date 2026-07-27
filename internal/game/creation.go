package game

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

func starterCharacters() []party.Character {
	// Keep the first three familiar human options stable for existing saves and
	// tests, then expose every currently verified single-class race/class pair.
	combos := []struct {
		race  party.Race
		class party.Class
	}{
		{party.RaceHuman, party.ClassFighter}, {party.RaceHuman, party.ClassCleric},
		{party.RaceHuman, party.ClassMagicUser}, {party.RaceHuman, party.ClassRanger},
		{party.RaceHuman, party.ClassPaladin}, {party.RaceHuman, party.ClassThief},
		{party.RaceDwarf, party.ClassFighter}, {party.RaceDwarf, party.ClassThief},
		{party.RaceElf, party.ClassCleric}, {party.RaceElf, party.ClassFighter},
		{party.RaceElf, party.ClassMagicUser}, {party.RaceElf, party.ClassThief},
		{party.RaceGnome, party.ClassFighter}, {party.RaceGnome, party.ClassThief},
		{party.RaceHalfElf, party.ClassCleric}, {party.RaceHalfElf, party.ClassFighter},
		{party.RaceHalfElf, party.ClassMagicUser}, {party.RaceHalfElf, party.ClassThief},
		{party.RaceHalfling, party.ClassFighter}, {party.RaceHalfling, party.ClassThief},
		{party.RaceHalfOrc, party.ClassCleric}, {party.RaceHalfOrc, party.ClassFighter},
		{party.RaceHalfOrc, party.ClassThief},
	}
	result := make([]party.Character, 0, len(combos))
	for index, combo := range combos {
		character := party.Character{
			ID: fmt.Sprintf("creation-%02d", index), Name: creationClassName(combo.class),
			Race: combo.race, Class: combo.class, Level: 1,
			Abilities: creationBaseAbilities(combo.class),
		}
		if _, err := party.StartingAgeSpecFor(character.Race, character.Class); err != nil {
			continue
		}
		if err := character.Validate(); err != nil {
			continue
		}
		result = append(result, character)
	}
	return result
}

func creationBaseAbilities(class party.Class) party.Abilities {
	switch class {
	case party.ClassCleric:
		return party.Abilities{Strength: 10, Intelligence: 10, Wisdom: 16, Dexterity: 10, Constitution: 14, Charisma: 10}
	case party.ClassFighter:
		return party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10}
	case party.ClassRanger:
		return party.Abilities{Strength: 14, Intelligence: 13, Wisdom: 14, Dexterity: 12, Constitution: 14, Charisma: 10}
	case party.ClassPaladin:
		return party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 14, Dexterity: 12, Constitution: 14, Charisma: 17}
	case party.ClassMagicUser:
		return party.Abilities{Strength: 8, Intelligence: 16, Wisdom: 10, Dexterity: 14, Constitution: 10, Charisma: 12}
	case party.ClassThief:
		return party.Abilities{Strength: 10, Intelligence: 12, Wisdom: 10, Dexterity: 16, Constitution: 12, Charisma: 10}
	default:
		return party.Abilities{}
	}
}

func creationClassName(class party.Class) string {
	return map[party.Class]string{
		party.ClassCleric: "牧師", party.ClassFighter: "戰士", party.ClassRanger: "遊俠",
		party.ClassPaladin: "聖武士", party.ClassMagicUser: "法師", party.ClassThief: "盜賊",
	}[class]
}

func (s *State) OpenCharacterCreation() error {
	if s.Mode == ModeCombat {
		return fmt.Errorf("character creation is unavailable during combat")
	}
	s.creationReturnMode = s.Mode
	s.CreationOptions = starterCharacters()
	s.CreationRoster = nil
	s.CreationCursor = 0
	s.CreationMessage = s.catalog.Text("creation_prompt", "選擇角色模板，Enter 加入隊伍。")
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
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_added", "已加入：%s（目前 %d 人）"), character.Name, len(s.CreationRoster))
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
		s.CreationMessage = s.catalog.Text("creation_ability_prompt", "左右選能力，上下調整數值。")
	}
	return nil
}

func (s *State) RerollCreationAbilities(seed int64) error {
	if s.Mode != ModeCharacterCreation || s.CreationCursor < 0 || s.CreationCursor >= len(s.CreationOptions) {
		return fmt.Errorf("ability reroll is unavailable")
	}
	s.CreationOptions[s.CreationCursor].Abilities = party.RollAbilities(seed)
	s.CreationMessage = s.catalog.Text("creation_rerolled", "能力值已重擲，請確認職業最低值。")
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
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_ability_updated", "能力值已調整為 %d。"), value)
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
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_named", "角色名稱：%s"), s.CreationName)
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
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("party_ready", "隊伍已建立。準備開始冒險。")
	s.Choices = []string{s.catalog.Text("enter_city", "進入城市"), s.catalog.Text("journey_on", "繼續旅程"), s.catalog.Text("camp", "紮營")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	s.CreationMessage = ""
	return nil
}

func (s *State) SavePartyFile(path string) error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("no character-created party is available to save")
	}
	areaState := s.Area
	areaState.GameTime = s.gameClock
	data, err := partySave.EncodeGameWithTime(s.partyRoster, areaState, uint8(s.Mode), uint8(s.Location), s.MapX, s.MapY, s.DungeonX, s.DungeonY, s.DungeonDirection, s.DungeonWallType, s.DungeonWallRoof, s.gameClock, s.gameAgeCycles)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
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
	s.Prompt = s.catalog.Text("party_ready", "隊伍已建立。準備開始冒險。")
	s.Choices = []string{s.catalog.Text("enter_city", "進入城市"), s.catalog.Text("journey_on", "繼續旅程"), s.catalog.Text("camp", "紮營")}
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
	fighters := make([]combat.Fighter, 0, len(file.Characters))
	for _, character := range file.Characters {
		fighter, err := s.fighterForCharacter(character)
		if err != nil {
			return err
		}
		fighters = append(fighters, fighter)
	}
	if err := s.SetParty(fighters); err != nil {
		return err
	}
	s.partyRoster = append(party.Roster(nil), file.Characters...)
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
	if file.Location <= uint8(LocationDaggerFalls) {
		s.Location = Location(file.Location)
	}
	if file.Version == 1 {
		// Legacy party.json had no adventure-state fields.
		s.Mode = ModeWilderness
		s.Location = LocationWilderness
	} else if file.Mode <= uint8(ModeCharacterCreation) {
		s.Mode = Mode(file.Mode)
	} else {
		s.Mode = ModeWilderness
	}
	if s.Mode == ModeMap && s.Location != LocationShadowdale {
		s.Mode = ModeWilderness
	}
	s.Prompt = s.catalog.Text("party_ready", "隊伍已建立。準備開始冒險。")
	s.Choices = []string{s.catalog.Text("enter_city", "進入城市"), s.catalog.Text("journey_on", "繼續旅程"), s.catalog.Text("camp", "紮營")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
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
	s.Prompt = s.catalog.Text("party_ready", "隊伍已建立。準備開始冒險。")
	s.Choices = []string{s.catalog.Text("enter_city", "進入城市"), s.catalog.Text("journey_on", "繼續旅程"), s.catalog.Text("camp", "紮營")}
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
