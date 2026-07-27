package game

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

func starterCharacters() []party.Character {
	return []party.Character{
		{ID: "fighter", Name: "戰士", Race: party.RaceHuman, Class: party.ClassFighter, Level: 1, Abilities: party.Abilities{Strength: 16, Intelligence: 10, Wisdom: 10, Dexterity: 12, Constitution: 14, Charisma: 10}},
		{ID: "cleric", Name: "牧師", Race: party.RaceHuman, Class: party.ClassCleric, Level: 1, Abilities: party.Abilities{Strength: 10, Intelligence: 10, Wisdom: 16, Dexterity: 10, Constitution: 14, Charisma: 10}},
		{ID: "mage", Name: "法師", Race: party.RaceElf, Class: party.ClassMagicUser, Level: 1, Abilities: party.Abilities{Strength: 8, Intelligence: 16, Wisdom: 10, Dexterity: 14, Constitution: 10, Charisma: 12}},
	}
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
	data, err := partySave.EncodeGameWithDungeonState(s.partyRoster, s.Area, uint8(s.Mode), uint8(s.Location), s.MapX, s.MapY, s.DungeonX, s.DungeonY, s.DungeonDirection, s.DungeonWallType, s.DungeonWallRoof)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// SaveSAVGAMPrefix writes the reference-compatible fixed SAVGAM prefix while
// preserving any raw records loaded previously. It is deliberately separate
// from SavePartyFile: the original player-file side effects are not yet
// decoded, so this method must not be presented as a complete DOS slot save.
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
	s.Area = file.Area
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
// into the running remake. It is intentionally a one-character bridge until
// the SAVGAM party/container format is decoded.
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
