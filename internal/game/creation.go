package game

import (
	"fmt"
	"unicode/utf8"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
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
		fighter, err := character.Fighter()
		if err != nil {
			return err
		}
		fighters = append(fighters, fighter)
	}
	if err := s.SetParty(fighters); err != nil {
		return err
	}
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("party_ready", "隊伍已建立。準備開始冒險。")
	s.Choices = []string{s.catalog.Text("enter_city", "進入城市"), s.catalog.Text("journey_on", "繼續旅程"), s.catalog.Text("camp", "紮營")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	s.CreationMessage = ""
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
