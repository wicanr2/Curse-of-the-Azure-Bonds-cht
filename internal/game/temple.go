package game

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

type templeCure struct {
	Key  string
	Cost uint16
}

var templeCures = []templeCure{
	{Key: "temple_cure_blindness", Cost: 1000},
	{Key: "temple_cure_disease", Cost: 1000},
	{Key: "temple_cure_light_wounds", Cost: 100},
	{Key: "temple_cure_serious_wounds", Cost: 350},
	{Key: "temple_cure_critical_wounds", Cost: 600},
	{Key: "temple_cure_heal", Cost: 5000},
	{Key: "temple_cure_neutralize_poison", Cost: 1000},
	{Key: "temple_cure_raise_dead", Cost: 5500},
	{Key: "temple_cure_remove_curse", Cost: 3500},
	{Key: "temple_cure_stone_to_flesh", Cost: 2000},
}

func (s *State) enterECLTemple() error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("temple service requires a party")
	}
	// 與商店共用原版 MONEY overlay 的公用池；進入服務不會清空玩家尚未
	// SHARE／TAKE 的資金（spec 798／863）。
	s.templeECLService = true
	s.templeCharacterIndex = 0
	s.eclMenuReturnMode = ModeDungeon
	s.eventReturnMode = ModeDungeon
	s.enterTempleMenu()
	return nil
}

func (s *State) enterTempleMenu() {
	s.templeMenu = true
	s.templeHealMenu = false
	s.templeConfirmMenu = false
	s.shopMenu = false
	s.Mode = ModePlace
	character := s.partyRoster[s.templeCharacterIndex]
	s.Prompt = fmt.Sprintf(s.catalog.Text("temple_character_prompt", "temple_character_prompt"), character.Name)
	s.Choices = []string{
		s.catalog.Text("temple_heal", "temple_heal"),
		s.catalog.Text("temple_view", "temple_view"),
		s.catalog.Text("temple_pool", "temple_pool"),
		s.catalog.Text("temple_share", "temple_share"),
		s.catalog.Text("temple_appraise", "temple_appraise"),
		s.catalog.Text("temple_exit", "temple_exit"),
	}
	s.currentOriginalChoices = []string{"TEMPLE_HEAL", "TEMPLE_VIEW", "TEMPLE_POOL", "TEMPLE_SHARE", "TEMPLE_APPRAISE", "TEMPLE_EXIT"}
	s.Message = ""
}

// TempleCycleCharacter mirrors the original GOTEMPLE G/O character keys
// (PC-98: 8/2). Previous wraps from the first member to the last; next stops
// at the last member. The keys are active only on the temple main menu.
func (s *State) TempleCycleCharacter(previous bool) bool {
	if s.Mode != ModePlace || !s.templeMenu || s.templeHealMenu ||
		s.templeConfirmMenu || len(s.partyRoster) == 0 {
		return false
	}
	if previous {
		if s.templeCharacterIndex == 0 {
			s.templeCharacterIndex = len(s.partyRoster) - 1
		} else {
			s.templeCharacterIndex--
		}
	} else if s.templeCharacterIndex+1 < len(s.partyRoster) {
		s.templeCharacterIndex++
	}
	s.enterTempleMenu()
	return true
}

// TempleCurrentCharacter reports the character selected by the original
// temple main-menu G/O controls. It deliberately returns false in submenus so
// callers cannot mistake a stale selection for an active cycling target.
func (s *State) TempleCurrentCharacter() (party.Character, int, bool) {
	if s.Mode != ModePlace || !s.templeMenu || s.templeHealMenu ||
		s.templeConfirmMenu || s.templeCharacterIndex < 0 ||
		s.templeCharacterIndex >= len(s.partyRoster) {
		return party.Character{}, 0, false
	}
	return s.partyRoster[s.templeCharacterIndex], s.templeCharacterIndex, true
}

func (s *State) enterTempleHealMenu() {
	s.templeMenu = true
	s.templeHealMenu = true
	s.templeConfirmMenu = false
	s.Mode = ModePlace
	character := s.partyRoster[s.templeCharacterIndex]
	s.Prompt = fmt.Sprintf(s.catalog.Text("temple_heal_prompt", "temple_heal_prompt"), character.Name)
	s.Choices = make([]string, 0, len(templeCures)+1)
	s.currentOriginalChoices = make([]string, 0, len(templeCures)+1)
	for index, cure := range templeCures {
		s.Choices = append(s.Choices, fmt.Sprintf(
			s.catalog.Text("temple_cure_choice", "temple_cure_choice"),
			s.catalog.Text(cure.Key, cure.Key), cure.Cost))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "TEMPLE_CURE_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("temple_cure_exit", "temple_cure_exit"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "TEMPLE_CURE_EXIT")
	s.Message = ""
}

func (s *State) enterTempleConfirmMenu(cureIndex int) {
	s.templePendingCure = cureIndex
	s.templeConfirmMenu = true
	s.Mode = ModePlace
	cure := templeCures[cureIndex]
	s.Prompt = fmt.Sprintf(s.catalog.Text("temple_confirm_prompt", "temple_confirm_prompt"),
		s.catalog.Text(cure.Key, cure.Key), cure.Cost)
	s.Choices = []string{
		s.catalog.Text("temple_confirm", "temple_confirm"),
		s.catalog.Text("temple_cancel", "temple_cancel"),
	}
	s.currentOriginalChoices = []string{"TEMPLE_CURE_CONFIRM", "TEMPLE_CURE_CANCEL"}
	s.Message = ""
}

func (s *State) selectTemple(originalChoice string) error {
	switch originalChoice {
	case "TEMPLE_HEAL":
		s.enterTempleHealMenu()
	case "TEMPLE_VIEW":
		character := s.partyRoster[s.templeCharacterIndex]
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = fmt.Sprintf(s.catalog.Text("temple_view_summary", "temple_view_summary"),
			character.Name, character.HitPoints, character.MaxHitPoints,
			characterCoinGoldWorth(character), len(character.Effects))
	case "TEMPLE_POOL":
		if err := s.PoolPartyGold(); err != nil {
			return err
		}
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = fmt.Sprintf(s.catalog.Text("temple_pool_done", "temple_pool_done"), s.moneyPool)
	case "TEMPLE_SHARE":
		if err := s.ShareGold(); err != nil {
			return err
		}
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = s.catalog.Text("temple_share_done", "temple_share_done")
	case "TEMPLE_APPRAISE":
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = s.catalog.Text("temple_appraise_empty", "temple_appraise_empty")
	case "TEMPLE_EXIT":
		s.templeMenu = false
		s.templeHealMenu = false
		s.templeConfirmMenu = false
		if s.templeECLService {
			s.templeECLService = false
			continued, err := s.continueECLAfterEngineBoundary()
			if err != nil {
				return err
			}
			if continued {
				return nil
			}
		}
		s.Mode = ModeDungeon
	case "TEMPLE_CURE_EXIT":
		s.enterTempleMenu()
	case "TEMPLE_CURE_CANCEL":
		s.enterTempleHealMenu()
	case "TEMPLE_CURE_CONFIRM":
		return s.applyTempleCure(s.templePendingCure)
	default:
		const prefix = "TEMPLE_CURE_"
		if len(originalChoice) <= len(prefix) || originalChoice[:len(prefix)] != prefix {
			return fmt.Errorf("unknown temple command %q", originalChoice)
		}
		index, err := strconv.Atoi(originalChoice[len(prefix):])
		if err != nil || index < 0 || index >= len(templeCures) {
			return fmt.Errorf("invalid temple cure command %q", originalChoice)
		}
		s.enterTempleConfirmMenu(index)
	}
	return nil
}

func (s *State) applyTempleCure(cureIndex int) error {
	if cureIndex < 0 || cureIndex >= len(templeCures) {
		return fmt.Errorf("temple cure %d is invalid", cureIndex)
	}
	character := &s.partyRoster[s.templeCharacterIndex]
	cure := templeCures[cureIndex]
	if characterCoinGoldWorth(*character) >= uint32(cure.Cost) {
		subtractCharacterGoldWorth(character, uint32(cure.Cost))
	} else if s.moneyPool >= uint32(cure.Cost) {
		s.moneyPool -= uint32(cure.Cost)
	} else {
		s.templeConfirmMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = s.catalog.Text("temple_insufficient_gold", "temple_insufficient_gold")
		return nil
	}

	rng := rand.New(rand.NewSource(s.fixSeed))
	s.fixSeed++
	switch cureIndex {
	case 0:
		removeCharacterAffects(character, 0x21)
	case 1:
		removeCharacterAffects(character, 0x1F, 0x22, 0x2B, 0x2C, 0x20, 0x39)
	case 2:
		healTempleCharacter(character, rng.Intn(8)+1)
	case 3:
		healTempleCharacter(character, rng.Intn(8)+rng.Intn(8)+3)
	case 4:
		healTempleCharacter(character, rng.Intn(8)+rng.Intn(8)+rng.Intn(8)+6)
	case 5:
		amount := character.MaxHitPoints - character.HitPoints - (rng.Intn(4) + 1)
		healTempleCharacter(character, amount)
		removeCharacterAffects(character, 0x21, 0x1F, 0x22, 0x2B, 0x2C, 0x20, 0x39, 0x44)
	case 6:
		removeCharacterAffects(character, 0x0F, 0x16, 0x37)
	case 7:
		if character.HealthStatus == party.HealthStatusDead || character.HealthStatus == party.HealthStatusAnimated {
			removeCharacterAffects(character, 0x20, 0x37)
			character.HitPoints = 1
			character.HealthStatus = party.HealthStatusOK
			character.Bleeding = 0
		}
	case 8:
		removeCharacterAffects(character, 0x24)
		for index := range character.Equipment {
			character.Equipment[index].Cursed = false
		}
	case 9:
		if character.HealthStatus == party.HealthStatusStoned {
			character.HealthStatus = party.HealthStatusOK
			character.HitPoints = 1
		}
	}
	if err := s.syncTempleCharacter(s.templeCharacterIndex); err != nil {
		return err
	}
	s.templeConfirmMenu = false
	s.Mode = ModeEvent
	s.eventReturnMode = ModePlace
	s.Message = fmt.Sprintf(s.catalog.Text("temple_cure_done", "temple_cure_done"),
		character.Name, s.catalog.Text(cure.Key, cure.Key))
	return nil
}

func removeCharacterAffects(character *party.Character, kinds ...uint8) {
	remove := make(map[uint8]struct{}, len(kinds))
	for _, kind := range kinds {
		remove[kind] = struct{}{}
	}
	kept := make([]monster.AffectRecord, 0, len(character.Effects))
	for _, effect := range character.Effects {
		if _, ok := remove[effect.Kind]; !ok {
			kept = append(kept, effect)
		}
	}
	character.Effects = kept
}

func healTempleCharacter(character *party.Character, amount int) {
	if amount <= 0 || character.HealthStatus == party.HealthStatusDead {
		return
	}
	character.HitPoints += amount
	if character.HitPoints > character.MaxHitPoints {
		character.HitPoints = character.MaxHitPoints
	}
	if character.HitPoints > 0 {
		character.HealthStatus = party.HealthStatusOK
		character.Bleeding = 0
	}
}

func (s *State) syncTempleCharacter(index int) error {
	if index >= len(s.party) {
		return nil
	}
	fighter, err := s.fighterForCharacter(s.partyRoster[index])
	if err != nil {
		return err
	}
	s.party[index] = fighter
	return nil
}
