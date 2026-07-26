// Package game contains platform-neutral remake state. Rendering and input
// adapters (Ebiten or a test harness) call Apply; no DOS assumptions belong
// here.
package game

import (
	"fmt"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
)

type Mode uint8

const (
	ModeTitle Mode = iota
	ModeWilderness
	ModeEvent
	ModeMap
	ModePlace
	ModeCombat
	ModeJournal
	ModeCharacterCreation
)

type Action uint8

const (
	ActionStart Action = iota
	ActionEnterCity
	ActionJourneyOn
)

type Location uint8

const (
	LocationWilderness Location = iota
	LocationShadowdale
)

type State struct {
	Mode         Mode
	Title        string
	Prompt       string
	Choices      []string
	Message      string
	Location     Location
	LocationName string
	MapX         int
	MapY         int

	// OriginalOpening records the English sentence found in the ECL payload.
	// It is evidence that the opening state was sourced from the original data,
	// not a replacement for the localized display string.
	OriginalOpening  string
	OriginalChoices  []string
	OriginalEvent    string
	OriginalLocation string
	JournalTitle     string
	JournalText      string
	JournalCloseText string
	JournalPages     []string
	JournalPage      int
	CampCount        int
	CreationOptions  []party.Character
	CreationRoster   party.Roster
	CreationCursor   int
	CreationMessage  string

	catalog                locale.Catalog
	eclBlock               []byte
	eclStart               int
	selectionSequence      []uint16
	currentOriginalChoices []string
	eventReturnMode        Mode
	journalReturnMode      Mode
	creationReturnMode     Mode
	session                *ecl.BlockSession
	party                  []combat.Fighter
	battle                 *combat.Battle
	combatTurns            []combat.Turn
	combatTurnIndex        int
	combatTargetIndex      int
	combatMessage          string
}

func NewStateFromECL(catalog locale.Catalog, block []byte) State {
	return NewStateFromECLBlocks(catalog, map[uint8][]byte{0: block}, 0)
}

// NewStateFromECLBlocks constructs the opening state over all decoded ECL
// blocks in one DAX member. The session is optional at the API boundary so
// tests and small tools can still provide one block.
func NewStateFromECLBlocks(catalog locale.Catalog, blocks map[uint8][]byte, initial uint8) State {
	state := NewState(catalog)
	session, err := ecl.NewBlockSession(blocks, initial)
	if err == nil {
		state.session = session
		state.eclBlock = session.CurrentData()
	}
	state.initializeECL()
	return state
}

func (s *State) initializeECL() {
	block := s.eclBlock
	if s.session != nil {
		block = s.session.CurrentData()
	}
	for _, candidate := range ecl.FindPackedTextCandidates(block) {
		if strings.Contains(candidate, "YOU ARE AT THE EDGE OF") {
			s.OriginalOpening = "YOU ARE AT THE EDGE OF"
			break
		}
	}
	if points, _, err := ecl.EntryPoints(block, 5); err == nil && len(points) == 5 {
		start := int(points[4]) - ecl.CodeAddressBase
		s.eclStart = start
		if result, runErr := ecl.RunSubset(block, start, 100); runErr == nil || len(result.Menus) > 0 {
			if len(result.Menus) > 0 {
				for _, option := range result.Menus[0].Options {
					s.OriginalChoices = append(s.OriginalChoices, option)
					s.currentOriginalChoices = append(s.currentOriginalChoices, option)
					switch option {
					case "ENTER CITY":
						s.Choices = append(s.Choices, s.catalog.Text("enter_city", "Enter city"))
					case "JOURNEY ON":
						s.Choices = append(s.Choices, s.catalog.Text("journey_on", "Journey on"))
					case "CAMP":
						s.Choices = append(s.Choices, s.catalog.Text("camp", "Camp"))
					default:
						s.Choices = append(s.Choices, option)
					}
				}
			}
		}
	}
}

func NewState(catalog locale.Catalog) State {
	journalPages := []string{
		catalog.Text("journal_page_1", "序章：隊伍醒來後必須查明蔚藍枷的來源。"),
		catalog.Text("journal_page_2", "達倫地區有許多城鎮與荒野等待探索。"),
		catalog.Text("journal_page_3", "五個邪惡勢力各自利用枷印。"),
		catalog.Text("journal_page_4", "解除枷印需要查明並擊破其來源。"),
		catalog.Text("journal_page_5", "三件神器是對抗火焰之主的關鍵。"),
		catalog.Text("journal_page_6", "先整裝，再向城鎮居民查詢線索。"),
		catalog.Text("journal_page_7", "戰鬥時觀察先攻、AC 與攻擊加值。"),
		catalog.Text("journal_page_8", "本 remake 以 Go／Ebiten 重建 Gold Box 冒險。"),
	}
	return State{
		Mode:                   ModeTitle,
		Title:                  catalog.Text("title", "Curse of the Azure Bonds"),
		Prompt:                 catalog.Text("press_enter", "Press Enter to continue"),
		Location:               LocationWilderness,
		LocationName:           catalog.Text("wilderness", "Wilderness"),
		currentOriginalChoices: []string{"ENTER CITY", "JOURNEY ON", "CAMP"},
		JournalTitle:           catalog.Text("journal_title", "冒險手札"),
		JournalText:            journalPages[0],
		JournalCloseText:       catalog.Text("journal_close", "Esc：返回"),
		JournalPages:           journalPages,
		catalog:                catalog,
	}
}

func (s *State) Apply(action Action) error {
	switch {
	case s.Mode == ModeTitle && action == ActionStart:
		s.Mode = ModeWilderness
		s.Prompt = s.catalog.Text("you_are_at_the_edge_of", "You are at the edge of")
		if len(s.Choices) == 0 {
			s.Choices = []string{
				s.catalog.Text("enter_city", "Enter city"),
				s.catalog.Text("journey_on", "Journey on"),
			}
		}
		s.Message = ""
		return nil
	case s.Mode == ModeWilderness && action == ActionEnterCity:
		return s.Select(0)
	case s.Mode == ModeWilderness && action == ActionJourneyOn:
		return s.Select(1)
	default:
		return fmt.Errorf("action %d is invalid in mode %d", action, s.Mode)
	}
}

// Select applies a localized opening choice and, when the state came from an
// ECL block, runs that choice through the bounded ECL subset.
func (s *State) Select(index int) error {
	if s.Mode == ModeMap {
		return fmt.Errorf("choice %d is invalid in map mode", index)
	}
	if s.Mode != ModeWilderness && s.Mode != ModePlace || index < 0 || index >= len(s.Choices) {
		return fmt.Errorf("choice %d is invalid in mode %d", index, s.Mode)
	}
	originalChoice := ""
	if index < len(s.currentOriginalChoices) {
		originalChoice = s.currentOriginalChoices[index]
	}
	if s.Mode == ModePlace {
		return s.selectPlace(index, originalChoice)
	}
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	switch index {
	case 0:
		s.Message = s.catalog.Text("enter_city", "Enter city")
	case 1:
		s.Message = s.catalog.Text("journey_on", "Journey on")
	case 2:
		s.Message = s.catalog.Text("camp", "Camp")
	default:
		s.Message = s.Choices[index]
	}
	if len(s.eclBlock) > 0 {
		s.selectionSequence = append(s.selectionSequence, uint16(index))
		var result ecl.RunResult
		if s.session != nil {
			result, _ = s.session.RunInteractive(180, s.selectionSequence)
			s.eclBlock = s.session.CurrentData()
			if start, err := s.session.InitialEntry(); err == nil {
				s.eclStart = start
			}
		} else {
			result, _ = ecl.RunSubsetInteractive(s.eclBlock, s.eclStart, 180, s.selectionSequence)
		}
		if len(s.selectionSequence) >= 4 && s.selectionSequence[0] == 0 && s.selectionSequence[1] == 0 && s.selectionSequence[2] == 1 && s.selectionSequence[3] == 0 {
			s.Location = LocationShadowdale
			s.LocationName = s.catalog.Text("shadowdale", "Shadowdale")
			s.OriginalLocation = "SHADOWDALE"
		}
		if result.CombatRequested {
			s.OriginalEvent = "COMBAT"
			s.Message = s.catalog.Text("combat_started", "戰鬥開始（戰鬥規則尚未完成）")
			s.eventReturnMode = ModeWilderness
			s.Mode = ModeEvent
			return nil
		}
		if result.ProgramExit && len(result.ProgramIDs) > 0 && result.ProgramIDs[len(result.ProgramIDs)-1] == 9 {
			return s.Camp()
		}
		// WILDERNESS/EXIT is the observed Shadowdale map-entry menu. Handle
		// these semantic transitions before the bounded runner's next-menu
		// result is applied, since the original command may leave another
		// continuation menu in the trace.
		if s.Location == LocationShadowdale && originalChoice == "WILDERNESS" {
			s.enterMap()
			return nil
		}
		if s.Location == LocationShadowdale && originalChoice == "EXIT" {
			s.leaveLocation()
			return nil
		}
		if result.WaitingForMenu && len(result.Menus) > 0 {
			menu := result.Menus[len(result.Menus)-1]
			s.Choices = make([]string, 0, len(menu.Options))
			s.currentOriginalChoices = append([]string(nil), menu.Options...)
			for _, option := range menu.Options {
				s.Choices = append(s.Choices, localizeOption(s.catalog, option))
			}
			if menu.Prompt != "" {
				s.Prompt = localizePrompt(s.catalog, menu.Prompt)
			}
			s.Mode = ModeWilderness
			return nil
		}
		if len(result.Text) > 0 {
			s.OriginalEvent = result.Text[len(result.Text)-1]
		}
	}
	if s.Location == LocationShadowdale && originalChoice == "WILDERNESS" {
		s.enterMap()
		return nil
	}
	if s.Location == LocationShadowdale && originalChoice == "EXIT" {
		s.leaveLocation()
	}
	return nil
}

// Camp applies the observable PROGRAM 9 transition. The full original
// interruption, spell recovery and HP restoration rules remain data-driven
// work; this method keeps the player-visible state transition explicit.
func (s *State) Camp() error {
	if s.Mode != ModeWilderness && s.Mode != ModeEvent {
		return fmt.Errorf("camp is invalid in mode %d", s.Mode)
	}
	s.CampCount++
	s.OriginalEvent = "PROGRAM 9"
	if len(s.party) > 0 {
		for index := range s.party {
			s.party[index].HitPoints = s.party[index].MaxHitPoints
		}
		s.Message = s.catalog.Text("camp_restored", "隊伍休息後恢復體力。")
	} else {
		s.Message = s.catalog.Text("camp_resting", "你們紮營休息。")
	}
	s.eventReturnMode = ModeWilderness
	s.Mode = ModeEvent
	return nil
}

// SetParty stores the current party roster without inventing character
// creation or import fields. It is the state boundary used by combat, CAMP,
// and the future save/import adapter.
func (s *State) SetParty(party []combat.Fighter) error {
	if len(party) == 0 {
		return fmt.Errorf("party cannot be empty")
	}
	seen := make(map[string]struct{}, len(party))
	for _, fighter := range party {
		if fighter.ID == "" {
			return fmt.Errorf("party member has empty ID")
		}
		if fighter.Side != combat.SideParty {
			return fmt.Errorf("fighter %q is not marked as party", fighter.ID)
		}
		if _, exists := seen[fighter.ID]; exists {
			return fmt.Errorf("duplicate party member %q", fighter.ID)
		}
		seen[fighter.ID] = struct{}{}
	}
	s.party = append([]combat.Fighter(nil), party...)
	return nil
}

func (s *State) PartyFighters() []combat.Fighter {
	return append([]combat.Fighter(nil), s.party...)
}

func (s *State) OpenJournal() error {
	if s.Mode == ModeCombat {
		return fmt.Errorf("journal is unavailable during combat")
	}
	s.journalReturnMode = s.Mode
	s.JournalPage = 0
	s.JournalText = s.JournalPages[0]
	s.Mode = ModeJournal
	return nil
}

func (s *State) NextJournalPage() error {
	if s.Mode != ModeJournal {
		return fmt.Errorf("journal is not open")
	}
	if s.JournalPage+1 < len(s.JournalPages) {
		s.JournalPage++
		s.JournalText = s.JournalPages[s.JournalPage]
	}
	return nil
}

func (s *State) PreviousJournalPage() error {
	if s.Mode != ModeJournal {
		return fmt.Errorf("journal is not open")
	}
	if s.JournalPage > 0 {
		s.JournalPage--
		s.JournalText = s.JournalPages[s.JournalPage]
	}
	return nil
}

func (s *State) CloseJournal() error {
	if s.Mode != ModeJournal {
		return fmt.Errorf("journal is not open")
	}
	s.Mode = s.journalReturnMode
	return nil
}

func (s *State) selectPlace(index int, originalChoice string) error {
	s.Mode = ModeEvent
	s.eventReturnMode = ModePlace
	s.OriginalEvent = originalChoice
	s.Message = localizeOption(s.catalog, originalChoice)
	if originalChoice == "LEAVE" {
		s.eventReturnMode = ModeMap
	}
	return nil
}

func (s *State) enterMap() {
	s.Mode = ModeMap
	s.MapX, s.MapY = 0, 0
	s.Choices = nil
	s.Prompt = s.catalog.Text("shadowdale_map_prompt", "暗影谷荒野")
	s.Message = ""
}

// Move changes the data-neutral map cursor used by the first navigable map slice.
// The coordinate system is deliberately data-neutral until the original map
// tile table is decoded; it still gives the renderer and tests a stable input
// contract without inventing tile semantics.
func (s *State) Move(dx, dy int) error {
	if s.Mode != ModeMap {
		return fmt.Errorf("movement is invalid in mode %d", s.Mode)
	}
	s.MapX += dx
	s.MapY += dy
	return nil
}

// LeaveMap returns from the navigable wilderness slice to the known location
// menu. The original ECL's next place menu remains data-driven work.
func (s *State) LeaveMap() error {
	if s.Mode != ModeMap {
		return fmt.Errorf("leave map is invalid in mode %d", s.Mode)
	}
	s.leaveLocation()
	return nil
}

// EnterPlaces exposes the Shadowdale place menu observed in ECL1 block 0x51.
// It is separate from map movement so the future decoded tile/encounter layer
// can be inserted without changing the place-event input contract.
func (s *State) EnterPlaces() error {
	if s.Mode != ModeMap || s.Location != LocationShadowdale {
		return fmt.Errorf("place menu is invalid in mode %d at location %d", s.Mode, s.Location)
	}
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("what_place", "你在暗影谷。要去哪裡？")
	s.Choices = []string{
		s.catalog.Text("inn", "客棧"),
		s.catalog.Text("store", "商店"),
		s.catalog.Text("bar", "酒館"),
		s.catalog.Text("leave", "離開"),
	}
	s.currentOriginalChoices = []string{"INN", "STORE", "BAR", "LEAVE"}
	s.Message = ""
	return nil
}

// Continue advances a localized place event back to its observed parent
// screen. Other event screens remain explicit future work.
func (s *State) Continue() error {
	if s.Mode != ModeEvent {
		return fmt.Errorf("continue is invalid in mode %d", s.Mode)
	}
	switch s.eventReturnMode {
	case ModeWilderness:
		s.Mode = ModeWilderness
		return nil
	case ModePlace:
		s.eventReturnMode = ModeEvent
		return s.EnterPlacesFromEvent()
	case ModeMap:
		s.Mode = ModeMap
		s.Prompt = s.catalog.Text("shadowdale_map_prompt", "暗影谷荒野")
		s.Message = ""
		return nil
	default:
		return fmt.Errorf("event has no continuation")
	}
}

func (s *State) EnterPlacesFromEvent() error {
	if s.Location != LocationShadowdale {
		return fmt.Errorf("place menu is invalid at location %d", s.Location)
	}
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("what_place", "你在暗影谷。要去哪裡？")
	s.Choices = []string{
		s.catalog.Text("inn", "客棧"),
		s.catalog.Text("store", "商店"),
		s.catalog.Text("bar", "酒館"),
		s.catalog.Text("leave", "離開"),
	}
	s.currentOriginalChoices = []string{"INN", "STORE", "BAR", "LEAVE"}
	s.Message = ""
	return nil
}

func (s *State) leaveLocation() {
	s.Mode = ModeWilderness
	s.MapX, s.MapY = 0, 0
	s.Choices = []string{
		s.catalog.Text("enter_city", "Enter city"),
		s.catalog.Text("journey_on", "Journey on"),
		s.catalog.Text("camp", "Camp"),
	}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	s.Prompt = s.catalog.Text("press_button", "Press any button or Enter to continue")
	s.Message = ""
}

func localizeOption(catalog locale.Catalog, option string) string {
	switch option {
	case "ENTER CITY":
		return catalog.Text("enter_city", "Enter city")
	case "JOURNEY ON":
		return catalog.Text("journey_on", "Journey on")
	case "CAMP":
		return catalog.Text("camp", "Camp")
	case "INN":
		return catalog.Text("inn", "Inn")
	case "STORE":
		return catalog.Text("store", "Store")
	case "BAR":
		return catalog.Text("bar", "Bar")
	case "LEAVE":
		return catalog.Text("leave", "Leave")
	case "SHADOWDALE":
		return catalog.Text("shadowdale", "Shadowdale")
	case "ASHABENFORD":
		return catalog.Text("ashabenford", "Ashabenford")
	case "DAGGER FALLS":
		return catalog.Text("dagger_falls", "Dagger Falls")
	case "WILDERNESS":
		return catalog.Text("wilderness", "Wilderness")
	case "EXIT":
		return catalog.Text("exit", "Exit")
	default:
		return option
	}
}

func localizePrompt(catalog locale.Catalog, prompt string) string {
	if prompt == "PRESS BUTTON OR RETURN TO CONTINUE." {
		return catalog.Text("press_button", "Press any button or Enter to continue")
	}
	return prompt
}
