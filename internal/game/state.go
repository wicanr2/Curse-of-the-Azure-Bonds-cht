// Package game contains platform-neutral remake state. Rendering and input
// adapters (Ebiten or a test harness) call Apply; no DOS assumptions belong
// here.
package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
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
	LocationAshabenford
	LocationDaggerFalls
)

type State struct {
	Mode            Mode
	Title           string
	Prompt          string
	Choices         []string
	Message         string
	Location        Location
	LocationName    string
	MapX            int
	MapY            int
	WildernessFloor mapdata.WildernessFloor
	Area            area.State
	GeoMapSet       uint8
	GeoMapBlock     uint8

	// OriginalOpening records the English sentence found in the ECL payload.
	// It is evidence that the opening state was sourced from the original data,
	// not a replacement for the localized display string.
	OriginalOpening          string
	OriginalChoices          []string
	OriginalEvent            string
	PictureRequested         bool
	PictureBlock             uint16
	BigPictureRequested      bool
	SceneCharacterRequested  bool
	SceneHeadBlock           uint8
	SceneBodyBlock           uint8
	OriginalLocation         string
	JournalTitle             string
	JournalText              string
	JournalCloseText         string
	JournalPages             []string
	JournalPage              int
	CampCount                int
	CreationOptions          []party.Character
	CreationRoster           party.Roster
	CreationCursor           int
	CreationMessage          string
	CreationName             string
	CreationEditing          bool
	CreationAbility          int
	CreationEditingAbilities bool

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
	partyRoster            party.Roster
	battle                 *combat.Battle
	combatTurns            []combat.Turn
	combatTurnIndex        int
	combatTargetIndex      int
	combatMessage          string
	monsterRecords         map[uint8]monster.Record
	itemCatalog            monster.BaseItemCatalog
	itemCatalogReady       bool
	combatSeed             int64
	eclSeed                int64
	mapSeed                int64
	geoMapPending          bool
	shopMenu               bool
	shopOffers             []ShopOffer
	moneyPool              uint32
	appraisalOffers        AppraisalOffers
	shopStockMenu          bool
	shopViewMenu           bool
	shopTakeMenu           bool
	shopTakeAmountMenu     bool
	shopAppraiseMenu       bool
	shopAppraiseConfirm    bool
	shopCharacterIndex     int
	shopTakeCharacter      int
	shopAppraiseCharacter  int
	shopAppraiseKind       TreasureKind
	campMenu               bool
	campViewMenu           bool
	campMagicMenu          bool
	saveRequested          bool
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
		SceneHeadBlock:         0xFF,
		currentOriginalChoices: []string{"ENTER CITY", "JOURNEY ON", "CAMP"},
		JournalTitle:           catalog.Text("journal_title", "冒險手札"),
		JournalText:            journalPages[0],
		JournalCloseText:       catalog.Text("journal_close", "Esc：返回"),
		JournalPages:           journalPages,
		catalog:                catalog,
		combatSeed:             1,
		eclSeed:                1,
		mapSeed:                1,
		GeoMapSet:              2,
		GeoMapBlock:            1,
		Area:                   area.State{GameArea: 2},
	}
}

// SetMonsterRecords installs the decoded MON*CHA table used when an ECL
// COMBAT command reaches the game state. Keeping this as an explicit adapter
// prevents the ECL VM from inventing combat statistics.
func (s *State) SetMonsterRecords(records map[uint8]monster.Record) {
	s.monsterRecords = make(map[uint8]monster.Record, len(records))
	for id, record := range records {
		s.monsterRecords[id] = record
	}
}

// SetItemCatalog installs the decoded original ITEMS table. Until this is
// called, old party/save paths retain their equipment-neutral projection.
func (s *State) SetItemCatalog(catalog monster.BaseItemCatalog) {
	s.itemCatalog = catalog
	s.itemCatalogReady = true
}

func (s *State) fighterForCharacter(character party.Character) (combat.Fighter, error) {
	if s.itemCatalogReady {
		return character.FighterWithEquipment(s.itemCatalog)
	}
	return character.Fighter()
}

// SetCombatSeed makes an ECL-triggered encounter reproducible for tests and
// debug comparisons while leaving the normal startup seed deterministic.
func (s *State) SetCombatSeed(seed int64) { s.combatSeed = seed }

// SetECLSeed controls RANDOM values while replaying an event sequence.
func (s *State) SetECLSeed(seed int64) { s.eclSeed = seed }

// SetMapSeed makes wilderness floor generation reproducible for replay and
// tests. The original engine rolls a fresh floor; this explicit seed keeps
// the remake deterministic until the original save/area seed is decoded.
func (s *State) SetMapSeed(seed int64) { s.mapSeed = seed }

// SetAreaState installs the decoded Area1/Area2 boundary used by ECL file
// loading and GEO selection.
func (s *State) SetAreaState(value area.State) {
	s.Area = value
	s.SceneHeadBlock = value.HeadBlockID
	s.GeoMapSet = value.GameArea
	s.GeoMapBlock = value.Current3DMapBlockID
}

func (s *State) SetInDungeon(value bool) { s.Area.InDungeon = value }

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
		if s.shopMenu {
			return s.selectShop(index, originalChoice)
		}
		return s.selectPlace(index, originalChoice)
	}
	if s.campMenu {
		return s.selectCamp(index, originalChoice)
	}
	if originalChoice == "CAMP" && len(s.eclBlock) == 0 {
		s.enterCampMenu()
		return nil
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
			result, _ = s.session.RunInteractiveSeed(180, s.selectionSequence, s.eclSeed)
			s.eclBlock = s.session.CurrentData()
			if start, err := s.session.InitialEntry(); err == nil {
				s.eclStart = start
			}
		} else {
			result, _ = ecl.RunSubsetInteractiveSeed(s.eclBlock, s.eclStart, 180, s.selectionSequence, s.eclSeed)
		}
		s.applyGeoMapLoad(result)
		s.applyCitySelection()
		if result.PictureRequested {
			s.PictureRequested = true
			s.PictureBlock = result.PictureBlock
			s.BigPictureRequested = result.BigPictureRequested
			s.SceneCharacterRequested = !result.BigPictureRequested && s.SceneHeadBlock != 0xFF
			if s.SceneCharacterRequested {
				s.SceneBodyBlock = uint8(result.PictureBlock)
			}
			s.OriginalEvent = "PICTURE"
			s.Message = "事件畫面"
			return nil
		}
		if result.CombatRequested {
			if len(result.MonsterSpawns) > 0 && len(s.party) > 0 && len(s.monsterRecords) > 0 {
				if err := s.StartEncounter(result, s.monsterRecords, s.party, s.combatSeed); err != nil {
					return err
				}
				return nil
			}
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
		if s.Location != LocationWilderness && originalChoice == "WILDERNESS" {
			s.enterMap()
			return nil
		}
		if s.Location != LocationWilderness && originalChoice == "EXIT" {
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
	if s.Location != LocationWilderness && originalChoice == "WILDERNESS" {
		s.enterMap()
		return nil
	}
	if s.Location != LocationWilderness && originalChoice == "EXIT" {
		s.leaveLocation()
	}
	return nil
}

func (s *State) applyGeoMapLoad(result ecl.RunResult) {
	if !result.LoadFilesRequested {
		return
	}
	lastBlock := uint8(0)
	if s.session != nil {
		lastBlock = s.session.CurrentBlockID()
	}
	effect := s.Area.ApplyLoadFiles(result.LoadFiles, lastBlock)
	if effect.GeoMapBlock == nil {
		return
	}
	s.GeoMapSet = s.Area.GameArea
	s.GeoMapBlock = *effect.GeoMapBlock
	s.geoMapPending = true
}

// ConsumeGeoMapRequest transfers the bounded ECL LOAD FILES effect to the
// renderer's GEO catalog. Keeping this at the state boundary avoids coupling
// the VM to Ebiten or to ZIP asset I/O.
func (s *State) ConsumeGeoMapRequest() (set, block uint8, ok bool) {
	if !s.geoMapPending {
		return 0, 0, false
	}
	s.geoMapPending = false
	return s.GeoMapSet, s.GeoMapBlock, true
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

func (s *State) enterCampMenu() {
	s.campMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_menu_prompt", "紮營選單")
	s.Choices = []string{
		s.catalog.Text("camp_save", "儲存"),
		s.catalog.Text("camp_view", "查看"),
		s.catalog.Text("camp_magic", "法術"),
		s.catalog.Text("camp_rest", "休息"),
		s.catalog.Text("camp_alter", "修改"),
		s.catalog.Text("camp_fix", "修理"),
		s.catalog.Text("camp_exit", "離開"),
	}
	s.currentOriginalChoices = []string{"SAVE", "VIEW", "MAGIC", "REST", "ALTER", "FIX", "EXIT"}
	s.Message = ""
}

func (s *State) selectCamp(index int, originalChoice string) error {
	if s.campMagicMenu {
		if originalChoice == "CAMP_MAGIC_EXIT" {
			s.campMagicMenu = false
			s.enterCampMenu()
			return nil
		}
		if strings.HasPrefix(originalChoice, "CAMP_MAGIC_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_MAGIC_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid camp magic character %q", originalChoice)
			}
			character := s.partyRoster[value]
			slots := make([]string, 0, len(character.SpellSlots))
			for _, spellID := range character.SpellSlots {
				slots = append(slots, fmt.Sprintf("0x%02X", spellID))
			}
			if len(slots) == 0 {
				slots = append(slots, s.catalog.Text("camp_magic_none", "目前沒有已記憶法術"))
			}
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "MAGIC"
			s.Message = fmt.Sprintf(s.catalog.Text("camp_magic_summary", "%s　法術欄位：%s"), character.Name, strings.Join(slots, "、"))
			return nil
		}
	}
	if s.campViewMenu {
		if originalChoice == "CAMP_VIEW_EXIT" {
			s.campViewMenu = false
			s.enterCampMenu()
			return nil
		}
		if strings.HasPrefix(originalChoice, "CAMP_VIEW_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_VIEW_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid camp view character %q", originalChoice)
			}
			character := s.partyRoster[value]
			equipment := make([]string, 0, len(character.Equipment))
			for _, item := range character.Equipment {
				equipment = append(equipment, monster.ChineseName(item))
			}
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "VIEW"
			s.Message = fmt.Sprintf(s.catalog.Text("camp_view_summary", "%s　%s　HP %d/%d　金幣 %d　寶石 %d　珠寶 %d　裝備：%s"), character.Name, characterClassName(character.Class), character.HitPoints, character.MaxHitPoints, character.Gold, character.Gems, character.Jewelry, strings.Join(equipment, "、"))
			return nil
		}
	}
	if originalChoice == "EXIT" {
		s.campMenu = false
		s.Mode = ModeWilderness
		s.Prompt = s.catalog.Text("press_button", "請按任意鍵或 Enter 繼續")
		s.Choices = []string{s.catalog.Text("enter_city", "進入城市"), s.catalog.Text("journey_on", "繼續旅程"), s.catalog.Text("camp", "紮營")}
		s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
		s.Message = ""
		return nil
	}
	if originalChoice == "REST" {
		if err := s.Camp(); err != nil {
			return err
		}
		s.campMenu = true
		return nil
	}
	if originalChoice == "SAVE" {
		s.Mode = ModeEvent
		s.eventReturnMode = ModeWilderness
		s.OriginalEvent = originalChoice
		if len(s.partyRoster) == 0 {
			s.Message = s.catalog.Text("camp_save_unavailable", "目前沒有可儲存的角色隊伍。")
			return nil
		}
		s.saveRequested = true
		s.Message = s.catalog.Text("camp_save_requested", "已要求儲存目前隊伍。")
		return nil
	}
	if originalChoice == "VIEW" {
		if len(s.partyRoster) == 0 {
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = originalChoice
			s.Message = s.catalog.Text("camp_view_unavailable", "目前沒有可查看的角色。")
			return nil
		}
		s.enterCampViewMenu()
		return nil
	}
	if originalChoice == "MAGIC" {
		if len(s.partyRoster) == 0 {
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = originalChoice
			s.Message = s.catalog.Text("camp_magic_unavailable", "目前沒有可查看法術的角色。")
			return nil
		}
		s.enterCampMagicMenu()
		return nil
	}
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.OriginalEvent = originalChoice
	s.Message = s.campActionMessage(originalChoice)
	return nil
}

// ConsumeSaveRequest transfers a CAMP SAVE intent to the platform adapter.
// The state layer never chooses a filesystem path or performs file I/O.
func (s *State) ConsumeSaveRequest() bool {
	requested := s.saveRequested
	s.saveRequested = false
	return requested
}

func (s *State) enterCampViewMenu() {
	s.campMenu = true
	s.campViewMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_view_prompt", "選擇要查看的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（HP %d/%d）", character.Name, character.HitPoints, character.MaxHitPoints))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_VIEW_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_view_exit", "返回紮營選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_VIEW_EXIT")
	s.Message = ""
}

func (s *State) enterCampMagicMenu() {
	s.campMenu = true
	s.campMagicMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_magic_prompt", "選擇要查看法術的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（已記憶 %d 個法術）", character.Name, len(character.SpellSlots)))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_magic_exit", "返回紮營選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_EXIT")
	s.Message = ""
}

func characterClassName(class party.Class) string {
	switch class {
	case party.ClassCleric:
		return "牧師"
	case party.ClassFighter:
		return "戰士"
	case party.ClassRanger:
		return "遊俠"
	case party.ClassPaladin:
		return "聖武士"
	case party.ClassMagicUser:
		return "魔法師"
	case party.ClassThief:
		return "盜賊"
	default:
		return "未知職業"
	}
}

func (s *State) campActionMessage(originalChoice string) string {
	switch originalChoice {
	case "SAVE":
		return s.catalog.Text("camp_save_unavailable", "請使用 F5 儲存目前隊伍；完整原版儲存選單尚待接入。")
	case "VIEW":
		return s.catalog.Text("camp_view_unavailable", "角色查看功能尚待接入。")
	case "MAGIC":
		return s.catalog.Text("camp_magic_unavailable", "法術準備功能尚待接入。")
	case "ALTER":
		return s.catalog.Text("camp_alter_unavailable", "角色修改功能尚待接入。")
	case "FIX":
		return s.catalog.Text("camp_fix_unavailable", "修理功能尚待接入。")
	default:
		return localizeOption(s.catalog, originalChoice)
	}
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
	for index := range s.party {
		if !s.party[index].HasPartyIcon {
			s.party[index].HasPartyIcon = true
			// Original new characters begin with head_icon=0 and
			// weapon_icon=0; do not invent slot-dependent art here.
			s.party[index].PartyHeadBlock = 0
			s.party[index].PartyBodyBlock = 0
			if s.party[index].PartyIconSize == 0 {
				s.party[index].PartyIconSize = 2
			}
		}
	}
	return nil
}

func (s *State) PartyFighters() []combat.Fighter {
	return append([]combat.Fighter(nil), s.party...)
}

// ResolveSpellSearch bridges an ECL SPELL signal to the currently loaded
// remake roster. It returns the first matching character/slot without
// mutating ECL memory; the runtime adapter can perform those writes once its
// resumable memory context is exposed at the right boundary.
func (s *State) ResolveSpellSearch(request ecl.SpellSearch) (party.SpellMatch, bool) {
	return s.partyRoster.FindSpell(request.SpellID)
}

// AdvancePartyEffects consumes imported DOS effect durations for the loaded
// party roster. It deliberately does not apply effect-specific combat
// modifiers; those belong to the rules/action layer that knows the current
// game clock boundary.
func (s *State) AdvancePartyEffects(minutes uint16) int {
	if minutes == 0 {
		return 0
	}
	removed := 0
	for index := range s.partyRoster {
		removed += s.partyRoster[index].AdvanceEffects(minutes)
	}
	return removed
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
	if originalChoice == "STORE" {
		s.enterShopMenu()
		return nil
	}
	s.Mode = ModeEvent
	s.eventReturnMode = ModePlace
	s.OriginalEvent = originalChoice
	s.Message = s.placeEventMessage(originalChoice)
	if originalChoice == "INN" {
		s.restorePartyAtInn()
		s.Message = s.catalog.Text("inn_restored", "你們在客棧安全休息，隊伍恢復體力。")
	}
	if originalChoice == "LEAVE" {
		s.eventReturnMode = ModeMap
	}
	return nil
}

func (s *State) enterShopMenu() {
	s.shopMenu = true
	s.shopStockMenu = false
	s.shopViewMenu = false
	s.shopTakeMenu = false
	s.shopTakeAmountMenu = false
	s.shopAppraiseMenu = false
	s.shopAppraiseConfirm = false
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_menu_prompt", "商店選單")
	s.Choices = []string{
		s.catalog.Text("shop_buy", "購買"),
		s.catalog.Text("shop_view", "查看"),
		s.catalog.Text("shop_take", "取出金幣"),
		s.catalog.Text("shop_pool", "集中金幣"),
		s.catalog.Text("shop_share", "分配金幣"),
		s.catalog.Text("shop_appraise", "估價"),
		s.catalog.Text("shop_exit", "離開商店"),
	}
	s.currentOriginalChoices = []string{"BUY", "VIEW", "TAKE", "POOL", "SHARE", "APPRAISE", "EXIT"}
	s.Message = ""
}

func (s *State) selectShop(index int, originalChoice string) error {
	if originalChoice == "SHOP_APPRAISE_ACCEPT" {
		offer, err := s.AppraiseTreasure(s.shopAppraiseCharacter, s.shopAppraiseKind)
		s.shopAppraiseMenu = false
		s.shopAppraiseConfirm = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "APPRAISE"
		if err != nil {
			s.Message = "估價失敗：" + err.Error()
		} else {
			s.Message = fmt.Sprintf(s.catalog.Text("shop_appraise_done", "店家支付 %d GP。"), offer)
		}
		return nil
	}
	if originalChoice == "SHOP_APPRAISE_REJECT" {
		s.shopAppraiseMenu = false
		s.shopAppraiseConfirm = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "APPRAISE"
		s.Message = s.catalog.Text("shop_appraise_rejected", "你拒絕了報價，財寶仍由隊伍保留。")
		return nil
	}
	if originalChoice == "SHOP_APPRAISE_CANCEL" {
		s.enterShopAppraiseTreasureMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_APPRAISE_CHARACTER_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_APPRAISE_CHARACTER_"))
		if err != nil {
			return fmt.Errorf("invalid shop appraise character command %q", originalChoice)
		}
		s.shopAppraiseCharacter = value
		s.enterShopAppraiseTreasureMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_APPRAISE_TREASURE_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_APPRAISE_TREASURE_"))
		if err != nil {
			return fmt.Errorf("invalid shop appraisal command %q", originalChoice)
		}
		if value != int(TreasureGems) && value != int(TreasureJewelry) {
			return fmt.Errorf("invalid appraisal treasure kind %d", value)
		}
		s.shopAppraiseKind = TreasureKind(value)
		s.enterShopAppraiseConfirmMenu()
		return nil
	}
	if originalChoice == "SHOP_APPRAISE_EXIT" {
		s.enterShopMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_TAKE_CHARACTER_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_TAKE_CHARACTER_"))
		if err != nil {
			return fmt.Errorf("invalid shop take character command %q", originalChoice)
		}
		s.shopTakeCharacter = value
		s.enterShopTakeAmountMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_TAKE_AMOUNT_") {
		value, err := strconv.ParseUint(strings.TrimPrefix(originalChoice, "SHOP_TAKE_AMOUNT_"), 10, 32)
		if err != nil {
			return fmt.Errorf("invalid shop take amount command %q", originalChoice)
		}
		if err := s.TakeGold(s.shopTakeCharacter, uint32(value)); err != nil {
			s.shopTakeMenu = false
			s.shopTakeAmountMenu = false
			s.Mode = ModeEvent
			s.eventReturnMode = ModePlace
			s.OriginalEvent = "TAKE"
			s.Message = "取出金幣失敗：" + err.Error()
			return nil
		}
		s.shopTakeMenu = false
		s.shopTakeAmountMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "TAKE"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_take_done", "已取出 %d GP 給%s。"), value, s.partyRoster[s.shopTakeCharacter].Name)
		return nil
	}
	if originalChoice == "SHOP_TAKE_EXIT" {
		s.enterShopMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_VIEW_") {
		if originalChoice == "SHOP_VIEW_EXIT" {
			s.enterShopMenu()
			return nil
		}
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_VIEW_"))
		if err != nil {
			return fmt.Errorf("invalid shop view command %q", originalChoice)
		}
		if value < 0 || value >= len(s.partyRoster) {
			return fmt.Errorf("shop view character index %d is out of range", value)
		}
		character := s.partyRoster[value]
		equipment := make([]string, 0, len(character.Equipment))
		for _, item := range character.Equipment {
			equipment = append(equipment, monster.ChineseName(item))
		}
		s.shopViewMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "VIEW"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_view_summary", "%s　HP %d/%d　金幣 %d　裝備：%s"), character.Name, character.HitPoints, character.MaxHitPoints, character.Gold, strings.Join(equipment, "、"))
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_OFFER_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_OFFER_"))
		if err != nil {
			return fmt.Errorf("invalid shop offer command %q", originalChoice)
		}
		if err := s.BuyShopOffer(s.shopCharacterIndex, value); err != nil {
			s.shopStockMenu = false
			s.Mode = ModeEvent
			s.eventReturnMode = ModePlace
			s.OriginalEvent = "BUY"
			s.Message = "購買失敗：" + err.Error()
			return nil
		}
		item := s.shopOffers[value].Item
		s.shopOffers = append(s.shopOffers[:value], s.shopOffers[value+1:]...)
		s.shopStockMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "BUY"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_purchase_done", "已購買%s。"), monster.ChineseName(item))
		return nil
	}
	if originalChoice == "SHOP_EXIT" {
		s.enterShopMenu()
		return nil
	}
	if originalChoice == "EXIT" {
		s.shopMenu = false
		return s.EnterPlacesFromEvent()
	}
	message := s.shopActionMessage(originalChoice)
	if s.shopStockMenu || s.shopViewMenu || s.shopTakeMenu || s.shopAppraiseMenu {
		return nil
	}
	s.Mode = ModeEvent
	s.eventReturnMode = ModePlace
	s.OriginalEvent = originalChoice
	s.Message = message
	return nil
}

func (s *State) shopActionMessage(originalChoice string) string {
	switch originalChoice {
	case "BUY":
		if len(s.shopOffers) == 0 {
			return s.catalog.Text("shop_buy_unavailable", "商店庫存尚未載入。")
		}
		s.enterShopStockMenu()
		return ""
	case "VIEW":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_view_unavailable", "目前沒有可查看的角色。")
		}
		s.enterShopViewMenu()
		return ""
	case "TAKE":
		if len(s.partyRoster) == 0 || s.moneyPool == 0 {
			return s.catalog.Text("shop_take_unavailable", "目前沒有可提取的 party 金幣。")
		}
		s.enterShopTakeMenu()
		return ""
	case "POOL":
		if err := s.PoolPartyGold(); err != nil {
			return "集中金幣失敗：" + err.Error()
		}
		return fmt.Sprintf(s.catalog.Text("shop_pool_done", "已集中金幣：%d GP。"), s.moneyPool)
	case "SHARE":
		before := s.moneyPool
		if err := s.ShareGold(); err != nil {
			return "分配金幣失敗：" + err.Error()
		}
		return fmt.Sprintf(s.catalog.Text("shop_share_done", "已分配金幣：%d GP。"), before)
	case "APPRAISE":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_appraise_unavailable", "目前沒有可估價的角色。")
		}
		s.enterShopAppraiseCharacterMenu()
		return ""
	default:
		return localizeOption(s.catalog, originalChoice)
	}
}

func (s *State) enterShopStockMenu() {
	s.shopMenu = true
	s.shopStockMenu = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_stock_prompt", "選擇要購買的物品")
	s.Choices = make([]string, 0, len(s.shopOffers)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.shopOffers)+1)
	for index, offer := range s.shopOffers {
		name := monster.ChineseName(offer.Item)
		s.Choices = append(s.Choices, fmt.Sprintf("%s（%d GP）", name, offer.Price))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_OFFER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_exit", "離開商店商品列表"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_EXIT")
	s.Message = ""
}

func (s *State) enterShopViewMenu() {
	s.shopMenu = true
	s.shopStockMenu = false
	s.shopViewMenu = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_view_prompt", "選擇要查看的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（HP %d/%d，%d GP）", character.Name, character.HitPoints, character.MaxHitPoints, character.Gold))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_VIEW_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_view_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_VIEW_EXIT")
	s.Message = ""
}

func (s *State) enterShopTakeMenu() {
	s.shopMenu = true
	s.shopStockMenu = false
	s.shopViewMenu = false
	s.shopTakeMenu = true
	s.shopTakeAmountMenu = false
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_take_prompt", "選擇要取出金幣的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（目前 %d GP）", character.Name, character.Gold))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_TAKE_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_take_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_TAKE_EXIT")
	s.Message = ""
}

func (s *State) enterShopTakeAmountMenu() {
	s.shopTakeAmountMenu = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_take_amount_prompt", "選擇要取出的金額")
	amounts := make([]uint32, 0, 4)
	for _, amount := range []uint32{1, 10, 100, s.moneyPool} {
		if amount == 0 || amount > s.moneyPool {
			continue
		}
		seen := false
		for _, existing := range amounts {
			if existing == amount {
				seen = true
				break
			}
		}
		if !seen {
			amounts = append(amounts, amount)
		}
	}
	s.Choices = make([]string, 0, len(amounts)+1)
	s.currentOriginalChoices = make([]string, 0, len(amounts)+1)
	for _, amount := range amounts {
		s.Choices = append(s.Choices, fmt.Sprintf("%d GP", amount))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_TAKE_AMOUNT_"+strconv.FormatUint(uint64(amount), 10))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_take_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_TAKE_EXIT")
	s.Message = ""
}

func (s *State) enterShopAppraiseCharacterMenu() {
	s.shopMenu = true
	s.shopStockMenu = false
	s.shopViewMenu = false
	s.shopTakeMenu = false
	s.shopTakeAmountMenu = false
	s.shopAppraiseMenu = true
	s.shopAppraiseConfirm = false
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_appraise_prompt", "選擇要估價的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（寶石 %d、珠寶 %d）", character.Name, character.Gems, character.Jewelry))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_appraise_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_EXIT")
	s.Message = ""
}

func (s *State) enterShopAppraiseTreasureMenu() {
	s.shopAppraiseMenu = true
	s.shopAppraiseConfirm = false
	s.Mode = ModePlace
	character := s.partyRoster[s.shopAppraiseCharacter]
	s.Prompt = s.catalog.Text("shop_appraise_treasure_prompt", "選擇要估價的財寶")
	s.Choices = make([]string, 0, 3)
	s.currentOriginalChoices = make([]string, 0, 3)
	if character.Gems > 0 {
		label := "寶石（報價未載入）"
		if s.appraisalOffers.GemsReady {
			label = fmt.Sprintf("寶石（報價 %d GP）", s.appraisalOffers.Gems)
		}
		s.Choices = append(s.Choices, label)
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_TREASURE_1")
	}
	if character.Jewelry > 0 {
		label := "珠寶（報價未載入）"
		if s.appraisalOffers.JewelryReady {
			label = fmt.Sprintf("珠寶（報價 %d GP）", s.appraisalOffers.Jewelry)
		}
		s.Choices = append(s.Choices, label)
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_TREASURE_2")
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_appraise_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_EXIT")
	s.Message = ""
}

func (s *State) enterShopAppraiseConfirmMenu() {
	s.shopAppraiseMenu = true
	s.shopAppraiseConfirm = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_appraise_confirm_prompt", "接受店家的報價嗎？")
	s.Choices = []string{
		s.catalog.Text("shop_appraise_accept", "接受"),
		s.catalog.Text("shop_appraise_reject", "拒絕"),
		s.catalog.Text("shop_appraise_cancel", "返回"),
	}
	s.currentOriginalChoices = []string{"SHOP_APPRAISE_ACCEPT", "SHOP_APPRAISE_REJECT", "SHOP_APPRAISE_CANCEL"}
	s.Message = ""
}

// SetShopCharacter selects the active character for BUY. The current UI keeps
// this explicit API boundary until the original VIEW/character selector is
// decoded; it defaults to the first roster member.
func (s *State) SetShopCharacter(index int) error {
	if index < 0 || index >= len(s.partyRoster) {
		return fmt.Errorf("shop character index %d is out of range", index)
	}
	s.shopCharacterIndex = index
	return nil
}

// restorePartyAtInn applies the safe-rest boundary described by the original
// manual: an inn is a protected place to rest. The roster is the source of
// truth for save/load, while party mirrors the currently rendered fighters.
func (s *State) restorePartyAtInn() {
	for index := range s.partyRoster {
		s.partyRoster[index].HitPoints = s.partyRoster[index].MaxHitPoints
	}
	for index := range s.party {
		if index < len(s.partyRoster) {
			s.party[index].HitPoints = s.partyRoster[index].HitPoints
			s.party[index].MaxHitPoints = s.partyRoster[index].MaxHitPoints
			continue
		}
		s.party[index].HitPoints = s.party[index].MaxHitPoints
	}
}

// placeEventMessage is the first localized place-event screen. The reference
// engine dispatches these choices into separate routines; keeping the
// dispatch explicit lets those routines replace this bounded screen without
// changing the ECL/menu contract.
func (s *State) placeEventMessage(originalChoice string) string {
	switch originalChoice {
	case "INN":
		return s.catalog.Text("inn_event", "你來到"+s.LocationName+"的客棧。住宿與休息功能尚待接入。")
	case "STORE":
		return s.catalog.Text("store_event", "你來到"+s.LocationName+"的商店。原版商店功能尚待接入。")
	case "BAR":
		return s.catalog.Text("bar_event", "你來到"+s.LocationName+"的酒館。情報與對話功能尚待接入。")
	default:
		return localizeOption(s.catalog, originalChoice)
	}
}

func (s *State) enterMap() {
	s.Mode = ModeMap
	s.MapX, s.MapY = 0, 0
	cityFlags, ok := mapdata.CityInfo(int(s.Area.CurrentCity))
	if !ok {
		cityFlags = 0
	}
	s.WildernessFloor = mapdata.GenerateWilderness(cityFlags, s.mapSeed)
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
	nextX, nextY := s.MapX+dx, s.MapY+dy
	if !s.WildernessFloor.CanEnter(nextX, nextY) {
		return fmt.Errorf("wilderness tile at (%d,%d) is not passable", nextX, nextY)
	}
	s.MapX, s.MapY = nextX, nextY
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
	if s.Mode != ModeMap || s.Location == LocationWilderness {
		return fmt.Errorf("place menu is invalid in mode %d at location %d", s.Mode, s.Location)
	}
	s.Mode = ModePlace
	s.Prompt = s.placePrompt()
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
	if s.PictureRequested {
		s.PictureRequested = false
		s.PictureBlock = 0
		s.BigPictureRequested = false
		s.SceneCharacterRequested = false
		s.SceneBodyBlock = 0
	}
	switch s.eventReturnMode {
	case ModeWilderness:
		if s.campMagicMenu {
			s.enterCampMagicMenu()
			return nil
		}
		if s.campViewMenu {
			s.enterCampViewMenu()
			return nil
		}
		if s.campMenu {
			s.enterCampMenu()
			return nil
		}
		s.Mode = ModeWilderness
		return nil
	case ModePlace:
		if s.shopMenu {
			s.enterShopMenu()
			return nil
		}
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

// SetSceneCharacter selects the reference Area2 HeadBlockId branch for the
// next PICTURE event. 0xFF means no head layer and restores normal PIC mode.
func (s *State) SetSceneCharacter(headBlock, bodyBlock uint8) {
	s.SceneHeadBlock = headBlock
	s.SceneBodyBlock = bodyBlock
}

func (s *State) EnterPlacesFromEvent() error {
	if s.Location == LocationWilderness {
		return fmt.Errorf("place menu is invalid at location %d", s.Location)
	}
	s.Mode = ModePlace
	s.Prompt = s.placePrompt()
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

func (s *State) placePrompt() string {
	return "你在" + s.LocationName + "。要去哪裡？"
}

// applyCitySelection maps the observed ECL city menu order to the three
// named locations. The first three selections are the proven opening path:
// ENTER CITY, CONTINUE, JOURNEY ON; the fourth selects the city.
func (s *State) applyCitySelection() {
	if len(s.selectionSequence) < 4 || s.selectionSequence[0] != 0 || s.selectionSequence[1] != 0 || s.selectionSequence[2] != 1 {
		return
	}
	choice := s.selectionSequence[3]
	if choice > 2 {
		return
	}
	locations := [...]struct {
		location Location
		key      string
		original string
	}{
		{LocationShadowdale, "shadowdale", "SHADOWDALE"},
		{LocationAshabenford, "ashabenford", "ASHABENFORD"},
		{LocationDaggerFalls, "dagger_falls", "DAGGER FALLS"},
	}
	selected := locations[choice]
	s.Location = selected.location
	s.LocationName = s.catalog.Text(selected.key, selected.original)
	s.OriginalLocation = selected.original
	s.Area.CurrentCity = uint8(choice)
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
	case "COMBAT":
		return catalog.Text("encounter_combat", "戰鬥")
	case "WAIT":
		return catalog.Text("encounter_wait", "等待")
	case "FLEE":
		return catalog.Text("encounter_flee", "撤退")
	case "ADVANCE":
		return catalog.Text("encounter_advance", "接近")
	case "PARLAY":
		return catalog.Text("encounter_parlay", "談判")
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
