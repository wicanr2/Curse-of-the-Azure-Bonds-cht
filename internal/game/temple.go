package game

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

type templeCure struct {
	Name string
	Cost uint16
}

var templeCures = []templeCure{
	{Name: "治療目盲", Cost: 1000},
	{Name: "治療疾病", Cost: 1000},
	{Name: "治療輕傷", Cost: 100},
	{Name: "治療重傷", Cost: 350},
	{Name: "治療致命傷", Cost: 600},
	{Name: "完全治療", Cost: 5000},
	{Name: "中和毒素", Cost: 1000},
	{Name: "死者復生", Cost: 5500},
	{Name: "解除詛咒", Cost: 3500},
	{Name: "石化復原", Cost: 2000},
}

func (s *State) enterECLTemple() error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("temple service requires a party")
	}
	s.moneyPool = 0
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
	s.Prompt = s.catalog.Text("temple_prompt", "剛德神殿")
	s.Choices = []string{"治療", "查看", "集中金幣", "分配金幣", "估價", "離開神殿"}
	s.currentOriginalChoices = []string{"TEMPLE_HEAL", "TEMPLE_VIEW", "TEMPLE_POOL", "TEMPLE_SHARE", "TEMPLE_APPRAISE", "TEMPLE_EXIT"}
	s.Message = ""
}

func (s *State) enterTempleHealMenu() {
	s.templeMenu = true
	s.templeHealMenu = true
	s.templeConfirmMenu = false
	s.Mode = ModePlace
	character := s.partyRoster[s.templeCharacterIndex]
	s.Prompt = fmt.Sprintf("%s，需要什麼幫助？", character.Name)
	s.Choices = make([]string, 0, len(templeCures)+1)
	s.currentOriginalChoices = make([]string, 0, len(templeCures)+1)
	for index, cure := range templeCures {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（%d GP）", cure.Name, cure.Cost))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "TEMPLE_CURE_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, "返回神殿")
	s.currentOriginalChoices = append(s.currentOriginalChoices, "TEMPLE_CURE_EXIT")
	s.Message = ""
}

func (s *State) enterTempleConfirmMenu(cureIndex int) {
	s.templePendingCure = cureIndex
	s.templeConfirmMenu = true
	s.Mode = ModePlace
	cure := templeCures[cureIndex]
	s.Prompt = fmt.Sprintf("%s需要 %d GP。確定施術？", cure.Name, cure.Cost)
	s.Choices = []string{"確定", "取消"}
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
		s.Message = fmt.Sprintf("%s　HP %d/%d　金幣價值 %d GP　狀態效果 %d",
			character.Name, character.HitPoints, character.MaxHitPoints,
			characterCoinGoldWorth(character), len(character.Effects))
	case "TEMPLE_POOL":
		if err := s.PoolPartyGold(); err != nil {
			return err
		}
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = fmt.Sprintf("已集中隊伍金幣；目前共有 %d GP。", s.moneyPool)
	case "TEMPLE_SHARE":
		if err := s.ShareGold(); err != nil {
			return err
		}
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = "已將集中金幣平均分配給隊伍。"
	case "TEMPLE_APPRAISE":
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.Message = "神殿目前沒有可估價的寶石或珠寶。"
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
		s.Message = "金幣不足。"
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
	s.Message = fmt.Sprintf("%s已接受%s。", character.Name, cure.Name)
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
