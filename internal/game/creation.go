package game

import (
	"fmt"

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
	return nil
}
