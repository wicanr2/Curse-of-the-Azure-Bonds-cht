// Package game contains platform-neutral remake state. Rendering and input
// adapters (Ebiten or a test harness) call Apply; no DOS assumptions belong
// here.
package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dungeon"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// ReportCombatError converts a recoverable combat action failure into a
// Traditional Chinese message without leaving the combat mode. Input
// adapters should call this instead of returning player-caused action errors
// to the Ebiten game loop.
func (s *State) ReportCombatError(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, combat.ErrAdjacentMissileTarget) {
		s.combatMessage = s.catalog.Text("combat_missile_adjacent_error", "飛彈武器不能攻擊相鄰目標。")
		return
	}
	s.combatMessage = fmt.Sprintf(s.catalog.Text("combat_action_error", "無法執行戰鬥行動：%s"), err.Error())
}

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
	Mode             Mode
	Title            string
	Prompt           string
	Choices          []string
	Message          string
	Location         Location
	LocationName     string
	MapX             int
	MapY             int
	DungeonX         int
	DungeonY         int
	DungeonDirection uint8
	DungeonWallType  uint8
	DungeonWallRoof  uint8
	WildernessFloor  mapdata.WildernessFloor
	Area             area.State
	GeoMapSet        uint8
	GeoMapBlock      uint8
	LoadPieces       [3]uint16

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
	savgamPrefix           *partySave.SAVGAMContainer
	savgamPlayers          map[string]party.DOSPlayerFiles
	pendingSoundEvents     []SoundEvent
	battle                 *combat.Battle
	combatTurns            []combat.Turn
	combatTurnIndex        int
	combatTargetIndex      int
	combatCastingSpell     uint8
	combatCastingClass     party.Class
	combatCastingClassSet  bool
	combatSpellTargetIndex int
	combatMoveMode         bool
	combatMoveRemaining    int
	combatView             bool
	combatViewFighterID    string
	combatMessage          string
	monsterRecords         map[uint8]monster.Record
	monsterRecordsByECL    map[uint8]map[uint8]monster.Record
	monsterAffects         map[uint8][]monster.AffectRecord
	monsterAffectsByECL    map[uint8]map[uint8][]monster.AffectRecord
	gameClock              [7]uint16
	gameAgeCycles          uint32
	itemCatalog            monster.BaseItemCatalog
	itemCatalogReady       bool
	ammunitionItemTypes    map[uint8][]uint8
	combatSeed             int64
	eclSeed                int64
	mapSeed                int64
	geoMapPending          bool
	loadPiecesPending      bool
	pendingSpellSearches   []ecl.SpellSearch
	pendingDamageRequests  []ecl.DamageRequest
	pendingProtection      []uint16
	shopMenu               bool
	shopOffers             []ShopOffer
	moneyPool              uint32
	appraisalOffers        AppraisalOffers
	shopStockMenu          bool
	shopViewMenu           bool
	shopTakeMenu           bool
	shopTakeAmountMenu     bool
	shopSellMenu           bool
	shopSellItemMenu       bool
	shopIdentifyMenu       bool
	shopIdentifyItemMenu   bool
	shopAppraiseMenu       bool
	shopAppraiseConfirm    bool
	shopCharacterIndex     int
	shopTakeCharacter      int
	shopSellCharacter      int
	shopIdentifyCharacter  int
	shopAppraiseCharacter  int
	shopAppraiseKind       TreasureKind
	barMenu                bool
	parlayMenu             bool
	barTales               []string
	barTaleIndex           int
	campMenu               bool
	campRestMenu           bool
	restHours              int
	campViewMenu           bool
	campMagicMenu          bool
	campMagicViewMenu      bool
	campMagicMemorizeMenu  bool
	campMagicMemorizeChar  int
	pendingMemorizedSpells map[int][]uint8
	saveRequested          bool
	alterMenu              bool
	alterOrderMenu         bool
	alterOrderSelected     int
	alterDropMenu          bool
	alterDropConfirm       bool
	alterDropSelected      int
	alterPicsMenu          bool
	alterSpeedMenu         bool
	alterIconMenu          bool
	alterIconEdit          bool
	alterIconCharacter     int
	alterIconHeadIndex     int
	alterIconBodyIndex     int
	picturesEnabled        bool
	animationsEnabled      bool
	messageSpeed           int
	fixSeed                int64
	dungeonSeed            int64
	combatMapDirection     uint8
}

// playerIconBlocks are the four verified CHEAD/CBODY block families extracted
// from the original assets. ICON never exposes blocks without a checked
// sprite pair.
var playerIconBlocks = []uint8{
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D,
	0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D,
	0xC0, 0xC1, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD,
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
		barTales: []string{
			catalog.Text("bar_tale_1", "酒客低聲說：公主與國王都喬裝在城中。"),
			catalog.Text("bar_tale_2", "有人說火焰巨人只怕三件古老神器，其中一件可能在北方瀑布下。"),
			catalog.Text("bar_tale_3", "許多士兵覺得深坑不祥，有人寧可逃亡也不願去守衛。"),
			catalog.Text("bar_tale_4", "這座城市的下水道，是達倫地區最危險的地方之一。"),
			catalog.Text("bar_tale_5", "有人看見紅袍刺客在森林小徑巡邏。"),
			catalog.Text("bar_tale_6", "商人冒險者 Akabar 已南下調查 Hap，另有一支女冒險者隊伍同行。"),
		},
		restHours:         24,
		combatSeed:        1,
		eclSeed:           1,
		mapSeed:           1,
		picturesEnabled:   true,
		animationsEnabled: true,
		messageSpeed:      3,
		fixSeed:           1,
		dungeonSeed:       1,
		GeoMapSet:         2,
		GeoMapBlock:       1,
		// Reference seg001.Init: mapPosX=7, mapPosY=0x0D, direction=0.
		DungeonX:         7,
		DungeonY:         13,
		DungeonDirection: 0,
		Area:             area.State{GameArea: 2},
	}
}

// SetBarTales installs the location/script-specific Tavern Tale sequence.
// The default entries come from the supplied Adventure Journal; a later ECL
// decoder can replace the sequence without changing the BAR UI contract.
func (s *State) SetBarTales(tales []string) {
	s.barTales = append([]string(nil), tales...)
	s.barTaleIndex = 0
}

func (s *State) BarTaleIndex() int { return s.barTaleIndex }

// SetRestHours sets the requested rest duration for deterministic tests and
// future memorize/time adapters. REST uses 24-hour units for natural healing.
func (s *State) SetRestHours(hours int) {
	if hours < 0 {
		hours = 0
	}
	s.restHours = hours
}

func (s *State) RestHours() int { return s.restHours }

func (s *State) enterBarMenu() {
	s.barMenu = true
	s.shopMenu = false
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("bar_menu_prompt", "酒館裡，你想做什麼？")
	s.Choices = []string{
		s.catalog.Text("bar_listen", "聽酒館傳聞"),
		s.catalog.Text("bar_exit", "離開酒館"),
	}
	s.currentOriginalChoices = []string{"BAR_LISTEN", "BAR_EXIT"}
	s.Message = ""
}

func (s *State) selectBar(originalChoice string) error {
	s.Mode = ModeEvent
	s.eventReturnMode = ModePlace
	s.OriginalEvent = originalChoice
	switch originalChoice {
	case "BAR_LISTEN":
		if s.barTaleIndex >= len(s.barTales) {
			s.Message = s.catalog.Text("bar_no_tales", "目前沒有新的酒館傳聞。")
			return nil
		}
		taleNumber := s.barTaleIndex + 1
		s.Message = fmt.Sprintf(s.catalog.Text("bar_tale", "酒客傳聞 %d：%s"), taleNumber, s.barTales[s.barTaleIndex])
		s.barTaleIndex++
		return nil
	case "BAR_EXIT":
		s.barMenu = false
		s.Message = s.catalog.Text("bar_exit_message", "你離開酒館，回到城市場所選單。")
		return nil
	default:
		return fmt.Errorf("unknown bar choice %q", originalChoice)
	}
}

func (s *State) enterCampRestMenu() {
	s.campRestMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_rest_menu_prompt", "休息設定")
	s.Choices = []string{
		fmt.Sprintf(s.catalog.Text("camp_rest_start", "開始休息（%d 小時）"), s.restHours),
		s.catalog.Text("camp_rest_add", "增加 24 小時"),
		s.catalog.Text("camp_rest_subtract", "減少 24 小時"),
		s.catalog.Text("camp_rest_exit", "返回紮營選單"),
	}
	s.currentOriginalChoices = []string{"REST_START", "REST_ADD", "REST_SUBTRACT", "REST_EXIT"}
	s.Message = ""
}

func (s *State) applyPendingMemorization() int {
	changed := 0
	for _, characterIndex := range pendingCharacterIndexes(s.pendingMemorizedSpells) {
		spells := s.pendingMemorizedSpells[characterIndex]
		if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
			continue
		}
		character := &s.partyRoster[characterIndex]
		character.SpellSlots = append(character.SpellSlots[:0], spells...)
		changed++
	}
	s.pendingMemorizedSpells = nil
	return changed
}

// restParty applies the manual's natural-healing portion: one HP per 24
// uninterrupted hours. Memorization timing is checked before this service;
// random interruption remains a separate adapter until its original data is
// decoded.
func (s *State) restParty() int {
	healed := s.restHours / 24
	if healed <= 0 {
		return 0
	}
	count := 0
	if len(s.partyRoster) > 0 {
		for index := range s.partyRoster {
			before := s.partyRoster[index].HitPoints
			s.partyRoster[index].HitPoints += healed
			if s.partyRoster[index].HitPoints > s.partyRoster[index].MaxHitPoints {
				s.partyRoster[index].HitPoints = s.partyRoster[index].MaxHitPoints
			}
			count += s.partyRoster[index].HitPoints - before
			id := s.partyRoster[index].ID
			for fighterIndex := range s.party {
				if s.party[fighterIndex].ID == id {
					s.party[fighterIndex].HitPoints = s.partyRoster[index].HitPoints
				}
			}
		}
		return count
	}
	for index := range s.party {
		before := s.party[index].HitPoints
		s.party[index].HitPoints += healed
		if s.party[index].HitPoints > s.party[index].MaxHitPoints {
			s.party[index].HitPoints = s.party[index].MaxHitPoints
		}
		count += s.party[index].HitPoints - before
	}
	return count
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

// SetMonsterRecordsForECL installs the MON*CHA table for one original ECL
// chapter. Monster IDs are chapter-local in the DOS image, so merging all six
// tables into one map would allow a valid ID to resolve to the wrong monster.
func (s *State) SetMonsterRecordsForECL(chapter uint8, records map[uint8]monster.Record) {
	if chapter < 1 || chapter > 6 {
		return
	}
	if s.monsterRecordsByECL == nil {
		s.monsterRecordsByECL = make(map[uint8]map[uint8]monster.Record)
	}
	copyRecords := make(map[uint8]monster.Record, len(records))
	for id, record := range records {
		copyRecords[id] = record
	}
	s.monsterRecordsByECL[chapter] = copyRecords
}

// SetMonsterAffects installs the decoded MON*SPC fallback table.
func (s *State) SetMonsterAffects(affects map[uint8][]monster.AffectRecord) {
	s.monsterAffects = cloneMonsterAffects(affects)
}

// SetMonsterAffectsForECL installs a chapter-local MON*SPC table.
func (s *State) SetMonsterAffectsForECL(chapter uint8, affects map[uint8][]monster.AffectRecord) {
	if chapter < 1 || chapter > 6 {
		return
	}
	if s.monsterAffectsByECL == nil {
		s.monsterAffectsByECL = make(map[uint8]map[uint8][]monster.AffectRecord)
	}
	s.monsterAffectsByECL[chapter] = cloneMonsterAffects(affects)
}

func cloneMonsterAffects(source map[uint8][]monster.AffectRecord) map[uint8][]monster.AffectRecord {
	result := make(map[uint8][]monster.AffectRecord, len(source))
	for id, effects := range source {
		result[id] = append([]monster.AffectRecord(nil), effects...)
	}
	return result
}

// monsterChapterForBlock follows the observed global ECL block namespaces:
// ECL2 uses 0x00..0x0F, ECL3 0x10..0x1F, through ECL6 0x40..0x4F, while
// ECL1 occupies 0x50..0x5F and its additional blocks.
func monsterChapterForBlock(blockID uint8) uint8 {
	switch {
	case blockID >= 0x50:
		return 1
	case blockID >= 0x40:
		return 6
	case blockID >= 0x30:
		return 5
	case blockID >= 0x20:
		return 4
	case blockID >= 0x10:
		return 3
	default:
		return 2
	}
}

func (s *State) monsterRecordsForCurrentECL() map[uint8]monster.Record {
	if s.session != nil && len(s.monsterRecordsByECL) > 0 {
		if records := s.monsterRecordsByECL[monsterChapterForBlock(s.session.CurrentBlockID())]; len(records) > 0 {
			return records
		}
	}
	return s.monsterRecords
}

func (s *State) monsterAffectsForCurrentECL() map[uint8][]monster.AffectRecord {
	if s.session != nil && len(s.monsterAffectsByECL) > 0 {
		if affects := s.monsterAffectsByECL[monsterChapterForBlock(s.session.CurrentBlockID())]; len(affects) > 0 {
			return affects
		}
	}
	return s.monsterAffects
}

// SetItemCatalog installs the decoded original ITEMS table. Until this is
// called, old party/save paths retain their equipment-neutral projection.
func (s *State) SetItemCatalog(catalog monster.BaseItemCatalog) {
	s.itemCatalog = catalog
	s.itemCatalogReady = true
}

// SetAmmunitionItemTypes injects the game's verified mapping from raw ITEMS
// AmmunitionType codes to inventory item types. CoAB stores these in separate
// namespaces, so the state must not invent a default mapping.
func (s *State) SetAmmunitionItemTypes(mapping map[uint8][]uint8) {
	s.ammunitionItemTypes = make(map[uint8][]uint8, len(mapping))
	for ammoType, itemTypes := range mapping {
		s.ammunitionItemTypes[ammoType] = append([]uint8(nil), itemTypes...)
	}
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

// SetDungeonSeed makes the d100 stream used by the dungeon action adapter
// reproducible for tests and replay.
func (s *State) SetDungeonSeed(seed int64) { s.dungeonSeed = seed }

// SetCombatMapDirection supplies the Area/encounter facing used by the
// reference SetupCombatActions icon direction adapter.
func (s *State) SetCombatMapDirection(direction uint8) error {
	if direction >= 8 {
		return fmt.Errorf("combat map direction %d is outside 0..7", direction)
	}
	s.combatMapDirection = direction
	return nil
}

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

// TurnDungeon rotates the dungeon camera in the reference eight-direction
// order (north, northeast, east, ...). The renderer owns movement, while the
// game state owns the persisted facing value.
func (s *State) TurnDungeon(delta int) {
	direction := (int(s.DungeonDirection) + delta) % 8
	if direction < 0 {
		direction += 8
	}
	s.DungeonDirection = uint8(direction)
}

func (s *State) Apply(action Action) error {
	switch {
	case s.Mode == ModeTitle && action == ActionStart:
		s.requestSound(SoundStart)
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
	if s.parlayMenu {
		return s.selectParlay(index, originalChoice)
	}
	if s.Mode == ModePlace {
		if s.shopMenu {
			return s.selectShop(index, originalChoice)
		}
		if s.barMenu {
			return s.selectBar(originalChoice)
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
	if originalChoice == "FLEE" {
		s.OriginalEvent = "FLEE"
		s.Message = s.catalog.Text("encounter_flee_done", "你們成功撤退，返回荒野。")
		return nil
	}
	if originalChoice == "PARLAY" {
		s.enterParlayMenu()
		s.Mode = ModeWilderness
		return nil
	}
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
		s.applyLoadPieces(result)
		s.applySpellSignals(result)
		s.applyECLDamageSignals(result)
		s.applyECLInventorySignals(result)
		s.applyCitySelection()
		if len(result.Text) > 0 {
			s.Message = localizeECLText(s.catalog, result.Text)
		}
		if result.PictureRequested {
			if !s.picturesEnabled {
				s.PictureRequested = false
				s.PictureBlock = result.PictureBlock
				s.OriginalEvent = "PICTURE"
				s.Message = s.catalog.Text("pics_monsters_off_message", "遭遇圖片已關閉。")
				return nil
			}
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
			records := s.monsterRecordsForCurrentECL()
			if len(result.MonsterSpawns) > 0 && len(s.party) > 0 && len(records) > 0 {
				if err := s.StartEncounterWithAffects(result, records, s.monsterAffectsForCurrentECL(), s.party, s.combatSeed); err != nil {
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

func (s *State) enterParlayMenu() {
	s.parlayMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("parlay_menu_prompt", "選擇談判策略")
	s.Choices = []string{
		s.catalog.Text("parlay_haughty", "傲慢"),
		s.catalog.Text("parlay_sly", "狡猾"),
		s.catalog.Text("parlay_meek", "謙卑"),
		s.catalog.Text("parlay_nice", "友善"),
		s.catalog.Text("parlay_abusive", "威嚇"),
	}
	s.currentOriginalChoices = []string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK", "PARLAY_NICE", "PARLAY_ABUSIVE"}
	s.Message = ""
}

func (s *State) selectParlay(index int, originalChoice string) error {
	if index < 0 || index >= len(s.currentOriginalChoices) || originalChoice == "" {
		return fmt.Errorf("parlay choice %d is invalid", index)
	}
	s.parlayMenu = false
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.OriginalEvent = "PARLAY"
	tactic := localizeOption(s.catalog, originalChoice)
	s.Message = fmt.Sprintf(s.catalog.Text("encounter_parlay_done", "你選擇以%s與怪物交涉；對方的反應仍待 encounter script。"), tactic)
	s.Choices = nil
	s.currentOriginalChoices = nil
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

// applyLoadPieces preserves the bounded ECL selector until a map-piece
// adapter (for example DUNGCOM/WALLDEF/TILES) is available. It intentionally
// does not assign file names or mutate the rendered floor.
func (s *State) applyLoadPieces(result ecl.RunResult) {
	if !result.LoadPiecesRequested {
		return
	}
	s.LoadPieces = result.LoadPieces
	s.loadPiecesPending = true
}

// ConsumeLoadPiecesRequest transfers the ECL LOAD PIECES selector exactly
// once to a future map-piece loader.
func (s *State) ConsumeLoadPiecesRequest() (pieces [3]uint16, ok bool) {
	if !s.loadPiecesPending {
		return [3]uint16{}, false
	}
	s.loadPiecesPending = false
	return s.LoadPieces, true
}

// applySpellSignals transfers ECL SPELL/PROTECTION requests to the game
// boundary without guessing their party or combat side effects. The ECL
// runner may reach these signals before a picture/menu pause, so they must
// survive the current State event until a party/rules adapter consumes them.
func (s *State) applySpellSignals(result ecl.RunResult) {
	if len(result.SpellSearches) > 0 {
		s.pendingSpellSearches = append(s.pendingSpellSearches, result.SpellSearches...)
	}
	if len(result.ProtectionRequests) > 0 {
		s.pendingProtection = append(s.pendingProtection, result.ProtectionRequests...)
	}
}

// applyECLDamageSignals preserves ECL environmental／script damage requests
// until a party rules adapter can resolve target selection, saving throws and
// deterministic dice. The bounded VM signal must survive State.Select just as
// SPELL and PROTECTION do; it must not silently become a fixed HP mutation.
func (s *State) applyECLDamageSignals(result ecl.RunResult) {
	if len(result.DamageRequests) > 0 {
		s.pendingDamageRequests = append(s.pendingDamageRequests, result.DamageRequests...)
	}
}

// applyECLInventorySignals is the party adapter for the bounded ECL
// inventory commands. FIND ITEM remains an observable query; DESTROY ITEMS
// is the verified mutation and is applied to the persistent roster.
func (s *State) applyECLInventorySignals(result ecl.RunResult) {
	for _, itemID := range result.DestroyItemIDs {
		for characterIndex := range s.partyRoster {
			s.partyRoster[characterIndex].DestroyItemType(uint8(itemID))
		}
	}
}

// ConsumeSpellSearches transfers pending ECL SPELL requests exactly once.
func (s *State) ConsumeSpellSearches() []ecl.SpellSearch {
	requests := append([]ecl.SpellSearch(nil), s.pendingSpellSearches...)
	s.pendingSpellSearches = nil
	return requests
}

// ConsumeDamageRequests transfers pending ECL DAMAGE requests exactly once.
// A later party/rules adapter owns target selection, saving throws, dice and
// HP writeback; this API intentionally returns the original raw operands.
func (s *State) ConsumeDamageRequests() []ecl.DamageRequest {
	requests := append([]ecl.DamageRequest(nil), s.pendingDamageRequests...)
	s.pendingDamageRequests = nil
	return requests
}

// ResolvePendingECLDamage applies pending requests transactionally through
// the party adapter. The selected index and both dice functions are explicit
// inputs because ECL memory does not yet expose a universal character picker.
// Roster HP is synchronized to any renderer-facing fighter with the same ID.
func (s *State) ResolvePendingECLDamage(selectedIndex int, rollDie, rollSave func(int) int) ([]party.DamageOutcome, error) {
	return s.resolvePendingECLDamage(selectedIndex, rollDie, rollSave, nil)
}

// ResolvePendingECLDamageWithHitResolver additionally enables the original
// random-target DAMAGE branch. The hit resolver is explicit because its AC
// projection and affect checks belong to the party/combat adapter.
func (s *State) ResolvePendingECLDamageWithHitResolver(selectedIndex int, rollDie, rollSave func(int) int, hitTarget party.DamageHitResolver) ([]party.DamageOutcome, error) {
	return s.resolvePendingECLDamage(selectedIndex, rollDie, rollSave, hitTarget)
}

// ResolvePendingECLDamageWithDefaultHitResolver projects the current party
// AC from Character (and decoded equipment when available), then applies the
// verified CanHitTarget natural-roll/invisibility/blink rules. Displace and
// other combat-round effects remain available through the injected variant.
func (s *State) ResolvePendingECLDamageWithDefaultHitResolver(selectedIndex int, rollDie, rollSave func(int) int) ([]party.DamageOutcome, error) {
	return s.ResolvePendingECLDamageWithDefaultHitResolverContext(selectedIndex, party.ECLHitContext{}, rollDie, rollSave)
}

// ResolvePendingECLDamageWithDefaultHitResolverContext is the State adapter
// for callers that have the current combat action delay/round. It preserves
// the transactional pending-request behavior while projecting Type_16 hit
// effects with that explicit context.
func (s *State) ResolvePendingECLDamageWithDefaultHitResolverContext(selectedIndex int, context party.ECLHitContext, rollDie, rollSave func(int) int) ([]party.DamageOutcome, error) {
	hitTarget := func(target party.Character, bonus int, hitRoll func(int) int) (bool, error) {
		fighter, err := target.Fighter()
		if err != nil {
			return false, err
		}
		if s.itemCatalogReady {
			fighter, err = target.FighterWithEquipment(s.itemCatalog)
			if err != nil {
				return false, err
			}
		}
		return party.CanHitECLDamageTargetWithContext(target, fighter.ArmorClass, bonus, context, hitRoll)
	}
	return s.ResolvePendingECLDamageWithHitResolver(selectedIndex, rollDie, rollSave, hitTarget)
}

// ResolveDeathEffects applies explicitly decoded Death-effect context to the
// party transactionally. ECL DAMAGE does not carry damage type, so callers
// must identify flags before enabling troll fire/acid behavior.
func (s *State) ResolveDeathEffects(context party.DeathEffectContext) error {
	wasDowned := make(map[string]bool, len(s.partyRoster))
	working := make(party.Roster, len(s.partyRoster))
	for index, character := range s.partyRoster {
		wasDowned[character.ID] = character.HitPoints == 0
		working[index] = character
		working[index].Effects = append([]monster.AffectRecord(nil), character.Effects...)
		if err := working[index].ApplyDeathEffects(context); err != nil {
			return err
		}
	}
	s.partyRoster = working
	for fighterIndex := range s.party {
		for characterIndex := range s.partyRoster {
			if s.party[fighterIndex].ID == s.partyRoster[characterIndex].ID {
				s.party[fighterIndex].HitPoints = s.partyRoster[characterIndex].HitPoints
				break
			}
		}
	}
	if s.battle != nil && s.Mode == ModeCombat {
		for _, character := range s.partyRoster {
			if _, ok := s.fighter(character.ID); !ok {
				continue
			}
			if err := s.battle.SetHitPoints(character.ID, character.HitPoints); err != nil {
				return err
			}
			if context.CombatHealAllowed && wasDowned[character.ID] && character.HitPoints > 0 && character.HealthStatus == party.HealthStatusOK {
				fighter, ok := s.fighter(character.ID)
				if !ok {
					return fmt.Errorf("combat-heal fighter %q disappeared", character.ID)
				}
				if fighter.DownedCorpse {
					if err := s.battle.RestoreCombatant(character.ID, combat.TilePoint{X: fighter.CombatX, Y: fighter.CombatY}); err != nil {
						return err
					}
				}
			}
			if character.HitPoints == 0 {
				s.clearCombatActionFor(character.ID)
			}
		}
		for index := range s.party {
			if fighter, ok := s.fighter(s.party[index].ID); ok {
				s.party[index] = fighter
			}
		}
		if s.battle.Status() != combat.StatusActive {
			return s.finishCombat()
		}
	}
	return nil
}

// ResolveDragonSlayer exposes the target-aware weapon effect without
// inventing a universal monster-type field in Character or ECL DAMAGE.
func (s *State) ResolveDragonSlayer(characterIndex int, targetMonsterType uint8, strengthDamageBonus int, rollDie func(int) int) (party.DragonSlayerResult, error) {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return party.DragonSlayerResult{}, fmt.Errorf("dragon-slayer character index %d outside party", characterIndex)
	}
	return s.partyRoster[characterIndex].ResolveDragonSlayer(targetMonsterType, strengthDamageBonus, rollDie)
}

func (s *State) resolvePendingECLDamage(selectedIndex int, rollDie, rollSave func(int) int, hitTarget party.DamageHitResolver) ([]party.DamageOutcome, error) {
	if len(s.pendingDamageRequests) == 0 {
		return nil, nil
	}
	working := make(party.Roster, len(s.partyRoster))
	for index, character := range s.partyRoster {
		working[index] = character
		// Type_16 displace consumes a persistent FX data bit. Deep-copy the
		// effect slice so a failed multi-request transaction cannot leak that
		// mutation into the live roster.
		working[index].Effects = append([]monster.AffectRecord(nil), character.Effects...)
	}
	outcomes := make([]party.DamageOutcome, 0)
	for _, request := range s.pendingDamageRequests {
		resolved, err := working.ApplyECLDamageWithHitResolver(request, selectedIndex, rollDie, rollSave, hitTarget)
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, resolved...)
	}
	s.partyRoster = working
	s.pendingDamageRequests = nil
	for fighterIndex := range s.party {
		for characterIndex := range s.partyRoster {
			if s.party[fighterIndex].ID == s.partyRoster[characterIndex].ID {
				s.party[fighterIndex].HitPoints = s.partyRoster[characterIndex].HitPoints
				break
			}
		}
	}
	if s.battle != nil && s.Mode == ModeCombat {
		for index := range s.partyRoster {
			status := s.partyRoster[index].HealthStatus
			if status == party.HealthStatusUnconscious || status == party.HealthStatusDying || status == party.HealthStatusDead {
				s.partyRoster[index].RemoveCombatAffects()
			}
		}
		for _, character := range s.partyRoster {
			if _, ok := s.fighter(character.ID); !ok {
				continue
			}
			if err := s.battle.SetHitPoints(character.ID, character.HitPoints); err != nil {
				return outcomes, err
			}
			if character.HitPoints == 0 {
				s.clearCombatActionFor(character.ID)
			}
		}
		if s.battle.Status() != combat.StatusActive {
			if err := s.finishCombat(); err != nil {
				return outcomes, err
			}
		}
	}
	return outcomes, nil
}

// ConsumeProtectionRequests transfers pending ECL PROTECTION requests
// exactly once. Address resolution remains the responsibility of the
// work-specific party/runtime adapter.
func (s *State) ConsumeProtectionRequests() []uint16 {
	requests := append([]uint16(nil), s.pendingProtection...)
	s.pendingProtection = nil
	return requests
}

// Camp applies the observable PROGRAM 9 transition by opening the CAMP menu.
// Resting is a separate menu service; entering CAMP must not heal the party.
func (s *State) Camp() error {
	if s.Mode != ModeWilderness && s.Mode != ModeEvent {
		return fmt.Errorf("camp is invalid in mode %d", s.Mode)
	}
	s.CampCount++
	s.enterCampMenu()
	return nil
}

// CureLightWoundsSpellID follows the one-based order in the verified
// first-level clerical spell table: Bless=1, Curse=2, Cure Light Wounds=3.
// It is kept explicit because the full DOS spell catalog is not yet decoded.
const CureLightWoundsSpellID uint8 = 3

// SetFixSeed makes the bounded FIX healing rolls reproducible.
func (s *State) SetFixSeed(seed int64) { s.fixSeed = seed }

// Fix applies the currently memorized Cure Light Wounds slots from clerics.
// The spell slots are restored after the operation, matching the manual's
// "rememorize previously memorized spells" behavior; interruption and time
// passage remain outside this deterministic service boundary.
func (s *State) Fix() (healed, casts int, err error) {
	if s.Mode != ModeWilderness && s.Mode != ModeEvent {
		return 0, 0, fmt.Errorf("fix is invalid in mode %d", s.Mode)
	}
	for _, character := range s.partyRoster {
		if character.Class != party.ClassCleric {
			continue
		}
		for _, spellID := range character.SpellSlots {
			if spellID == CureLightWoundsSpellID {
				casts++
			}
		}
	}
	if casts == 0 {
		s.eventReturnMode = ModeWilderness
		s.Mode = ModeEvent
		s.OriginalEvent = "FIX"
		s.Message = s.catalog.Text("fix_no_cure", "沒有已記憶的 Cure Light Wounds，隊伍未改變。")
		return 0, 0, nil
	}
	rng := rand.New(rand.NewSource(s.fixSeed))
	for cast := 0; cast < casts; cast++ {
		target := -1
		for index, character := range s.partyRoster {
			if character.HitPoints < character.MaxHitPoints {
				target = index
				break
			}
		}
		if target < 0 {
			break
		}
		amount := rng.Intn(8) + 1
		before := s.partyRoster[target].HitPoints
		s.partyRoster[target].HitPoints += amount
		if s.partyRoster[target].HitPoints > s.partyRoster[target].MaxHitPoints {
			s.partyRoster[target].HitPoints = s.partyRoster[target].MaxHitPoints
		}
		healed += s.partyRoster[target].HitPoints - before
		id := s.partyRoster[target].ID
		for fighterIndex := range s.party {
			if s.party[fighterIndex].ID == id {
				s.party[fighterIndex].HitPoints = s.partyRoster[target].HitPoints
			}
		}
	}
	s.eventReturnMode = ModeWilderness
	s.Mode = ModeEvent
	s.OriginalEvent = "FIX"
	s.Message = fmt.Sprintf(s.catalog.Text("fix_done", "FIX 完成：施放 %d 次 Cure Light Wounds，共恢復 %d HP。"), casts, healed)
	return healed, casts, nil
}

func (s *State) enterCampMenu() {
	s.campMenu = true
	s.campRestMenu = false
	s.campViewMenu = false
	s.campMagicMenu = false
	s.campMagicViewMenu = false
	s.campMagicMemorizeMenu = false
	s.campMagicMemorizeChar = -1
	s.alterMenu = false
	s.alterOrderMenu = false
	s.alterOrderSelected = -1
	s.alterDropMenu = false
	s.alterDropConfirm = false
	s.alterDropSelected = -1
	s.alterPicsMenu = false
	s.alterSpeedMenu = false
	s.alterIconMenu = false
	s.alterIconEdit = false
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
	if s.campRestMenu {
		switch originalChoice {
		case "REST_ADD":
			s.restHours += 24
			s.enterCampRestMenu()
			return nil
		case "REST_SUBTRACT":
			if s.restHours >= 24 {
				s.restHours -= 24
			}
			s.enterCampRestMenu()
			return nil
		case "REST_EXIT":
			s.enterCampMenu()
			return nil
		case "REST_START":
			requiredHours := firstLevelMemorizationHours(s.pendingMemorizedSpells)
			if requiredHours > s.restHours {
				s.campRestMenu = false
				s.Mode = ModeEvent
				s.eventReturnMode = ModeWilderness
				s.OriginalEvent = "REST"
				s.Message = fmt.Sprintf(s.catalog.Text("camp_rest_insufficient", "休息 %d 小時不足以完成法術記憶（至少需要 %d 小時）；法術選擇仍保留。"), s.restHours, requiredHours)
				return nil
			}
			memorized := s.applyPendingMemorization()
			if err := s.AdvanceGameTimeHours(s.restHours); err != nil {
				return err
			}
			healed := s.restParty()
			s.campRestMenu = false
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "REST"
			s.Message = fmt.Sprintf(s.catalog.Text("camp_rest_done", "休息 %d 小時完成，隊伍自然恢復 %d HP，完成 %d 名角色的法術記憶。"), s.restHours, healed, memorized)
			return nil
		default:
			return fmt.Errorf("unknown camp rest choice %q", originalChoice)
		}
	}
	if s.alterIconMenu {
		if originalChoice == "ALTER_ICON_EXIT" {
			s.alterIconMenu = false
			s.alterIconEdit = false
			s.enterAlterMenu()
			return nil
		}
		if s.alterIconEdit {
			switch originalChoice {
			case "ALTER_ICON_HEAD_STATUS", "ALTER_ICON_BODY_STATUS":
				return nil
			case "ALTER_ICON_HEAD_PREV":
				s.alterIconHeadIndex = (s.alterIconHeadIndex + len(playerIconBlocks) - 1) % len(playerIconBlocks)
				s.applyIconSelection()
				return nil
			case "ALTER_ICON_HEAD_NEXT":
				s.alterIconHeadIndex = (s.alterIconHeadIndex + 1) % len(playerIconBlocks)
				s.applyIconSelection()
				return nil
			case "ALTER_ICON_BODY_PREV":
				s.alterIconBodyIndex = (s.alterIconBodyIndex + len(playerIconBlocks) - 1) % len(playerIconBlocks)
				s.applyIconSelection()
				return nil
			case "ALTER_ICON_BODY_NEXT":
				s.alterIconBodyIndex = (s.alterIconBodyIndex + 1) % len(playerIconBlocks)
				s.applyIconSelection()
				return nil
			case "ALTER_ICON_DONE":
				s.alterIconEdit = false
				s.enterAlterIconMenu()
				return nil
			}
		}
		if strings.HasPrefix(originalChoice, "ALTER_ICON_CHARACTER_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "ALTER_ICON_CHARACTER_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid alter icon character %q", originalChoice)
			}
			s.alterIconCharacter = value
			s.enterAlterIconEditMenu()
			return nil
		}
	}
	if s.alterSpeedMenu {
		switch originalChoice {
		case "ALTER_SPEED_EXIT":
			s.alterSpeedMenu = false
			s.enterAlterMenu()
			return nil
		case "ALTER_SPEED_SLOWER":
			if s.messageSpeed > 1 {
				s.messageSpeed--
			}
			s.enterAlterSpeedMenu()
			return nil
		case "ALTER_SPEED_FASTER":
			if s.messageSpeed < 5 {
				s.messageSpeed++
			}
			s.enterAlterSpeedMenu()
			return nil
		}
	}
	if s.alterPicsMenu {
		switch originalChoice {
		case "ALTER_PICS_EXIT":
			s.alterPicsMenu = false
			s.enterAlterMenu()
			return nil
		case "ALTER_PICS_MONSTERS":
			s.picturesEnabled = !s.picturesEnabled
			s.enterAlterPicsMenu()
			return nil
		case "ALTER_PICS_ANIMATIONS":
			s.animationsEnabled = !s.animationsEnabled
			s.enterAlterPicsMenu()
			return nil
		}
	}
	if s.alterOrderMenu {
		if originalChoice == "ALTER_ORDER_EXIT" {
			s.alterOrderMenu = false
			s.enterAlterMenu()
			return nil
		}
		if strings.HasPrefix(originalChoice, "ALTER_ORDER_CHARACTER_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "ALTER_ORDER_CHARACTER_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid alter order character %q", originalChoice)
			}
			if s.alterOrderSelected < 0 {
				s.alterOrderSelected = value
				s.enterAlterOrderDestinationMenu()
				return nil
			}
			if value >= len(s.partyRoster) {
				return fmt.Errorf("invalid alter order destination %d", value)
			}
			if err := s.movePartyCharacter(s.alterOrderSelected, value); err != nil {
				return err
			}
			s.alterOrderMenu = false
			s.alterOrderSelected = -1
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "ALTER ORDER"
			s.Message = s.catalog.Text("alter_order_done", "隊伍順序已更新。")
			return nil
		}
	}
	if s.alterDropMenu {
		if originalChoice == "ALTER_DROP_EXIT" || originalChoice == "ALTER_DROP_CANCEL" {
			s.alterDropMenu = false
			s.alterDropConfirm = false
			s.alterDropSelected = -1
			s.enterAlterMenu()
			return nil
		}
		if s.alterDropConfirm {
			if originalChoice != "ALTER_DROP_CONFIRM" {
				return fmt.Errorf("invalid alter drop confirmation %q", originalChoice)
			}
			if err := s.dropPartyCharacter(s.alterDropSelected); err != nil {
				s.Mode = ModeEvent
				s.eventReturnMode = ModeWilderness
				s.OriginalEvent = "ALTER DROP"
				s.Message = "移除失敗：" + err.Error()
				return nil
			}
			s.alterDropMenu = false
			s.alterDropConfirm = false
			s.alterDropSelected = -1
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "ALTER DROP"
			s.Message = s.catalog.Text("alter_drop_done", "角色已從隊伍移除。此操作無法復原。")
			return nil
		}
		if strings.HasPrefix(originalChoice, "ALTER_DROP_CHARACTER_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "ALTER_DROP_CHARACTER_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid alter drop character %q", originalChoice)
			}
			s.alterDropSelected = value
			s.enterAlterDropConfirmMenu()
			return nil
		}
	}
	if s.alterMenu {
		if originalChoice == "ALTER_EXIT" {
			s.alterMenu = false
			s.enterCampMenu()
			return nil
		}
		if originalChoice == "ALTER_ORDER" {
			if len(s.partyRoster) < 2 {
				s.Mode = ModeEvent
				s.eventReturnMode = ModeWilderness
				s.OriginalEvent = originalChoice
				s.Message = s.catalog.Text("alter_order_unavailable", "至少需要兩名角色才能調整順序。")
				return nil
			}
			s.enterAlterOrderMenu()
			return nil
		}
		if originalChoice == "ALTER_DROP" {
			if len(s.partyRoster) < 2 {
				s.Mode = ModeEvent
				s.eventReturnMode = ModeWilderness
				s.OriginalEvent = originalChoice
				s.Message = s.catalog.Text("alter_drop_unavailable", "至少需要兩名角色才能移除角色。")
				return nil
			}
			s.enterAlterDropMenu()
			return nil
		}
		if originalChoice == "ALTER_PICS" {
			s.enterAlterPicsMenu()
			return nil
		}
		if originalChoice == "ALTER_SPEED" {
			s.enterAlterSpeedMenu()
			return nil
		}
		if originalChoice == "ALTER_ICON" {
			if len(s.partyRoster) == 0 {
				s.Mode = ModeEvent
				s.eventReturnMode = ModeWilderness
				s.OriginalEvent = originalChoice
				s.Message = s.catalog.Text("alter_icon_unavailable", "目前沒有可設定小人的角色。")
				return nil
			}
			s.enterAlterIconMenu()
			return nil
		}
		s.Mode = ModeEvent
		s.eventReturnMode = ModeWilderness
		s.OriginalEvent = originalChoice
		s.Message = s.alterActionMessage(originalChoice)
		return nil
	}
	if s.campMagicMenu {
		if originalChoice == "CAMP_MAGIC_EXIT" {
			s.campMagicMenu = false
			s.enterCampMenu()
			return nil
		}
		switch originalChoice {
		case "CAMP_MAGIC_DISPLAY":
			s.enterCampMagicViewMenu()
			return nil
		case "CAMP_MAGIC_MEMORIZE":
			s.enterCampMagicMemorizeCharacterMenu()
			return nil
		case "CAMP_MAGIC_REST":
			s.enterCampRestMenu()
			return nil
		case "CAMP_MAGIC_CAST", "CAMP_MAGIC_SCRIBE":
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "MAGIC"
			s.Message = s.catalog.Text("camp_magic_pending", "此法術功能已進入資料邊界，完整規則仍待接入。")
			return nil
		}
	}
	if s.campMagicMemorizeMenu {
		if originalChoice == "CAMP_MAGIC_MEMORIZE_EXIT" {
			s.campMagicMemorizeMenu = false
			s.campMagicMemorizeChar = -1
			s.enterCampMagicMenu()
			return nil
		}
		if s.campMagicMemorizeChar < 0 {
			if strings.HasPrefix(originalChoice, "CAMP_MAGIC_MEM_CHAR_") {
				value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_MAGIC_MEM_CHAR_"))
				if err != nil || value < 0 || value >= len(s.partyRoster) {
					return fmt.Errorf("invalid memorize character %q", originalChoice)
				}
				s.enterCampMagicMemorizeSpellMenu(value)
				return nil
			}
			return fmt.Errorf("unknown memorize character choice %q", originalChoice)
		}
		if originalChoice == "CAMP_MAGIC_MEM_DONE" {
			s.enterCampMagicMemorizeCharacterMenu()
			s.Message = s.catalog.Text("camp_magic_memorize_selected", "法術選擇已暫存，請在 REST 後完成記憶。")
			return nil
		}
		if originalChoice == "CAMP_MAGIC_MEM_CANCEL" {
			delete(s.pendingMemorizedSpells, s.campMagicMemorizeChar)
			s.enterCampMagicMemorizeCharacterMenu()
			return nil
		}
		if strings.HasPrefix(originalChoice, "CAMP_MAGIC_MEM_SPELL_") {
			spellIndex, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_MAGIC_MEM_SPELL_"))
			character := s.partyRoster[s.campMagicMemorizeChar]
			if err != nil || spellIndex < 0 || spellIndex >= len(character.KnownSpells) {
				return fmt.Errorf("invalid memorize spell %q", originalChoice)
			}
			if s.pendingMemorizedSpells == nil {
				s.pendingMemorizedSpells = make(map[int][]uint8)
			}
			selected := s.pendingMemorizedSpells[s.campMagicMemorizeChar]
			spellID := character.KnownSpells[spellIndex]
			removed := false
			for index, selectedID := range selected {
				if selectedID == spellID {
					selected = append(selected[:index], selected[index+1:]...)
					removed = true
					break
				}
			}
			if !removed {
				capacity := firstLevelMemorizedCapacity(character)
				if len(selected) >= capacity {
					s.Message = fmt.Sprintf(s.catalog.Text("camp_magic_memorize_full", "最多可選 %d 個法術。"), capacity)
					return nil
				}
				selected = append(selected, spellID)
			}
			s.pendingMemorizedSpells[s.campMagicMemorizeChar] = selected
			s.enterCampMagicMemorizeSpellMenu(s.campMagicMemorizeChar)
			return nil
		}
	}
	if s.campMagicViewMenu {
		if originalChoice == "CAMP_MAGIC_VIEW_EXIT" {
			s.campMagicViewMenu = false
			s.enterCampMagicMenu()
			return nil
		}
		if strings.HasPrefix(originalChoice, "CAMP_MAGIC_VIEW_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_MAGIC_VIEW_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid camp magic character %q", originalChoice)
			}
			character := s.partyRoster[value]
			slots := make([]string, 0, len(character.SpellSlots))
			for _, spellID := range character.SpellSlots {
				slots = append(slots, campSpellLabel(s.catalog, character.Class, spellID))
			}
			if len(slots) == 0 {
				slots = append(slots, s.catalog.Text("camp_magic_none", "目前沒有已記憶法術"))
			}
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "MAGIC"
			s.Message = fmt.Sprintf(s.catalog.Text("camp_magic_summary", "%s　法術欄位：%s　可用法術：%d 個"), character.Name, strings.Join(slots, "、"), len(character.KnownSpells))
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
		s.enterCampRestMenu()
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
	if originalChoice == "ALTER" {
		s.enterAlterMenu()
		return nil
	}
	if originalChoice == "FIX" {
		_, _, err := s.Fix()
		return err
	}
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.OriginalEvent = originalChoice
	s.Message = s.campActionMessage(originalChoice)
	return nil
}

func (s *State) enterAlterMenu() {
	s.campMenu = true
	s.alterMenu = true
	s.alterOrderMenu = false
	s.alterOrderSelected = -1
	s.alterDropMenu = false
	s.alterDropConfirm = false
	s.alterDropSelected = -1
	s.alterPicsMenu = false
	s.alterSpeedMenu = false
	s.alterIconMenu = false
	s.alterIconEdit = false
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("alter_prompt", "修改隊伍與遊戲設定")
	s.Choices = []string{
		s.catalog.Text("alter_order", "順序"),
		s.catalog.Text("alter_drop", "移除"),
		s.catalog.Text("alter_speed", "速度"),
		s.catalog.Text("alter_icon", "小人"),
		s.catalog.Text("alter_pics", "圖片"),
		s.catalog.Text("alter_exit", "離開"),
	}
	s.currentOriginalChoices = []string{"ALTER_ORDER", "ALTER_DROP", "ALTER_SPEED", "ALTER_ICON", "ALTER_PICS", "ALTER_EXIT"}
	s.Message = ""
}

func (s *State) enterAlterSpeedMenu() {
	s.campMenu = true
	s.alterMenu = true
	s.alterSpeedMenu = true
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf(s.catalog.Text("alter_speed_prompt", "訊息速度：第%d級"), s.messageSpeed)
	s.Choices = []string{
		s.catalog.Text("alter_speed_slower", "較慢"),
		s.catalog.Text("alter_speed_faster", "較快"),
		s.catalog.Text("alter_speed_exit", "返回修改選單"),
	}
	s.currentOriginalChoices = []string{"ALTER_SPEED_SLOWER", "ALTER_SPEED_FASTER", "ALTER_SPEED_EXIT"}
	s.Message = ""
}

// MessageSpeed returns the current 1..5 message reveal speed for renderers.
func (s *State) MessageSpeed() int { return s.messageSpeed }

func (s *State) enterAlterIconMenu() {
	s.campMenu = true
	s.alterMenu = true
	s.alterIconMenu = true
	s.alterIconEdit = false
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("alter_icon_prompt", "選擇要設定小人的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（頭部 %02X／身體 %02X）", character.Name, character.IconHeadBlock, character.IconWeaponBlock))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_ICON_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("alter_icon_exit", "返回修改選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_ICON_EXIT")
	s.Message = ""
}

func (s *State) enterAlterIconEditMenu() {
	character := s.partyRoster[s.alterIconCharacter]
	s.alterIconEdit = true
	s.alterIconHeadIndex = iconBlockIndex(character.IconHeadBlock)
	s.alterIconBodyIndex = iconBlockIndex(character.IconWeaponBlock)
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf(s.catalog.Text("alter_icon_edit_prompt", "設定%s的小人"), character.Name)
	s.Choices = []string{
		fmt.Sprintf(s.catalog.Text("alter_icon_head", "頭部：%02X"), playerIconBlocks[s.alterIconHeadIndex]),
		s.catalog.Text("alter_icon_head_prev", "頭部上一個"),
		s.catalog.Text("alter_icon_head_next", "頭部下一個"),
		fmt.Sprintf(s.catalog.Text("alter_icon_body", "身體：%02X"), playerIconBlocks[s.alterIconBodyIndex]),
		s.catalog.Text("alter_icon_body_prev", "身體上一個"),
		s.catalog.Text("alter_icon_body_next", "身體下一個"),
		s.catalog.Text("alter_icon_done", "完成"),
	}
	s.currentOriginalChoices = []string{"ALTER_ICON_HEAD_STATUS", "ALTER_ICON_HEAD_PREV", "ALTER_ICON_HEAD_NEXT", "ALTER_ICON_BODY_STATUS", "ALTER_ICON_BODY_PREV", "ALTER_ICON_BODY_NEXT", "ALTER_ICON_DONE"}
	s.Message = ""
}

func iconBlockIndex(block uint8) int {
	for index, candidate := range playerIconBlocks {
		if candidate == block {
			return index
		}
	}
	return 0
}

func (s *State) applyIconSelection() {
	if s.alterIconCharacter < 0 || s.alterIconCharacter >= len(s.partyRoster) {
		return
	}
	head := playerIconBlocks[s.alterIconHeadIndex]
	body := playerIconBlocks[s.alterIconBodyIndex]
	s.partyRoster[s.alterIconCharacter].IconHeadBlock = head
	s.partyRoster[s.alterIconCharacter].IconWeaponBlock = body
	id := s.partyRoster[s.alterIconCharacter].ID
	for index := range s.party {
		if s.party[index].ID == id {
			s.party[index].HasPartyIcon = true
			s.party[index].PartyHeadBlock = head
			s.party[index].PartyBodyBlock = body
		}
	}
	s.enterAlterIconEditMenu()
}

func (s *State) enterAlterPicsMenu() {
	s.campMenu = true
	s.alterMenu = true
	s.alterPicsMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("alter_pics_prompt", "遭遇圖片設定")
	monsterState := s.catalog.Text("alter_pics_on", "開啟")
	if !s.picturesEnabled {
		monsterState = s.catalog.Text("alter_pics_off", "關閉")
	}
	animationState := s.catalog.Text("alter_pics_on", "開啟")
	if !s.animationsEnabled {
		animationState = s.catalog.Text("alter_pics_off", "關閉")
	}
	s.Choices = []string{
		fmt.Sprintf(s.catalog.Text("alter_pics_monsters", "怪物圖片：%s"), monsterState),
		fmt.Sprintf(s.catalog.Text("alter_pics_animations", "動畫：%s"), animationState),
		s.catalog.Text("alter_pics_exit", "返回修改選單"),
	}
	s.currentOriginalChoices = []string{"ALTER_PICS_MONSTERS", "ALTER_PICS_ANIMATIONS", "ALTER_PICS_EXIT"}
	s.Message = ""
}

// PicturesEnabled and AnimationsEnabled expose renderer-neutral ALTER PICS
// preferences without coupling the game state to Ebiten.
func (s *State) PicturesEnabled() bool   { return s.picturesEnabled }
func (s *State) AnimationsEnabled() bool { return s.animationsEnabled }

func (s *State) enterAlterDropMenu() {
	s.campMenu = true
	s.alterMenu = true
	s.alterDropMenu = true
	s.alterDropConfirm = false
	s.alterDropSelected = -1
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("alter_drop_prompt", "選擇要移除的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（HP %d/%d）", character.Name, character.HitPoints, character.MaxHitPoints))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_DROP_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("alter_drop_exit", "返回修改選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_DROP_EXIT")
	s.Message = ""
}

func (s *State) enterAlterDropConfirmMenu() {
	character := s.partyRoster[s.alterDropSelected]
	s.alterDropConfirm = true
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf(s.catalog.Text("alter_drop_confirm_prompt", "確定要永久移除%s？"), character.Name)
	s.Choices = []string{s.catalog.Text("alter_drop_confirm", "確認移除"), s.catalog.Text("alter_drop_cancel", "取消")}
	s.currentOriginalChoices = []string{"ALTER_DROP_CONFIRM", "ALTER_DROP_CANCEL"}
	s.Message = s.catalog.Text("alter_drop_warning", "移除後角色將從隊伍與存檔中刪除，無法復原。")
}

func (s *State) dropPartyCharacter(index int) error {
	if index < 0 || index >= len(s.partyRoster) {
		return fmt.Errorf("party character index %d is out of range", index)
	}
	if len(s.partyRoster) <= 1 {
		return fmt.Errorf("cannot remove the last party character")
	}
	id := s.partyRoster[index].ID
	s.partyRoster = append(s.partyRoster[:index], s.partyRoster[index+1:]...)
	for fighterIndex, fighter := range s.party {
		if fighter.ID == id {
			s.party = append(s.party[:fighterIndex], s.party[fighterIndex+1:]...)
			break
		}
	}
	return nil
}

func (s *State) enterAlterOrderMenu() {
	s.campMenu = true
	s.alterMenu = true
	s.alterOrderMenu = true
	s.alterOrderSelected = -1
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("alter_order_prompt", "選擇要移動的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%d. %s", index+1, character.Name))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_ORDER_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("alter_order_exit", "返回修改選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_ORDER_EXIT")
	s.Message = ""
}

func (s *State) enterAlterOrderDestinationMenu() {
	s.Prompt = s.catalog.Text("alter_order_destination_prompt", "選擇新的位置")
	for index, character := range s.partyRoster {
		s.Choices[index] = fmt.Sprintf("第%d位：%s", index+1, character.Name)
		s.currentOriginalChoices[index] = "ALTER_ORDER_CHARACTER_" + strconv.Itoa(index)
	}
	s.Choices[len(s.Choices)-1] = s.catalog.Text("alter_order_cancel", "取消")
	s.currentOriginalChoices[len(s.currentOriginalChoices)-1] = "ALTER_ORDER_EXIT"
	s.Message = s.catalog.Text("alter_order_selected", "已選取角色，請選擇新的位置。")
}

func (s *State) movePartyCharacter(from, to int) error {
	if from < 0 || from >= len(s.partyRoster) || to < 0 || to >= len(s.partyRoster) {
		return fmt.Errorf("party order index out of range: %d -> %d", from, to)
	}
	if from == to {
		return nil
	}
	selected := s.partyRoster[from]
	s.partyRoster = append(s.partyRoster[:from], s.partyRoster[from+1:]...)
	s.partyRoster = append(s.partyRoster, party.Character{})
	copy(s.partyRoster[to+1:], s.partyRoster[to:])
	s.partyRoster[to] = selected
	if len(s.party) == 0 {
		return nil
	}
	byID := make(map[string]combat.Fighter, len(s.party))
	for _, fighter := range s.party {
		byID[fighter.ID] = fighter
	}
	reordered := make([]combat.Fighter, 0, len(s.party))
	used := make(map[string]bool, len(s.party))
	for _, character := range s.partyRoster {
		if fighter, ok := byID[character.ID]; ok && !used[character.ID] {
			reordered = append(reordered, fighter)
			used[character.ID] = true
		}
	}
	for _, fighter := range s.party {
		if !used[fighter.ID] {
			reordered = append(reordered, fighter)
		}
	}
	s.party = reordered
	return nil
}

func (s *State) alterActionMessage(originalChoice string) string {
	switch originalChoice {
	case "ALTER_DROP":
		return s.catalog.Text("alter_drop_unavailable", "角色移除功能尚待接入。")
	case "ALTER_SPEED":
		return s.catalog.Text("alter_speed_unavailable", "遊戲速度設定功能尚待接入。")
	case "ALTER_ICON":
		return s.catalog.Text("alter_icon_unavailable", "戰鬥小人設定功能尚待接入。")
	case "ALTER_PICS":
		return s.catalog.Text("alter_pics_unavailable", "遭遇圖片設定功能尚待接入。")
	default:
		return localizeOption(s.catalog, originalChoice)
	}
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
	s.campMagicViewMenu = false
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_magic_menu_prompt", "法術選單")
	s.Choices = []string{
		s.catalog.Text("camp_magic_cast", "施法"),
		s.catalog.Text("camp_magic_memorize", "記憶法術"),
		s.catalog.Text("camp_magic_scribe", "抄錄"),
		s.catalog.Text("camp_magic_display", "查看已記憶法術"),
		s.catalog.Text("camp_magic_rest", "休息"),
		s.catalog.Text("camp_magic_exit", "返回紮營選單"),
	}
	s.currentOriginalChoices = []string{"CAMP_MAGIC_CAST", "CAMP_MAGIC_MEMORIZE", "CAMP_MAGIC_SCRIBE", "CAMP_MAGIC_DISPLAY", "CAMP_MAGIC_REST", "CAMP_MAGIC_EXIT"}
	s.Message = ""
}

func (s *State) enterCampMagicViewMenu() {
	s.campMenu = true
	s.campMagicMenu = false
	s.campMagicViewMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_magic_prompt", "選擇要查看法術的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("camp_magic_character", "%s（已記憶 %d 個法術／可用 %d 個）"), character.Name, len(character.SpellSlots), len(character.KnownSpells)))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_VIEW_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_magic_view_exit", "返回法術選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_VIEW_EXIT")
	s.Message = ""
}

func (s *State) enterCampMagicMemorizeCharacterMenu() {
	s.campMenu = true
	s.campMagicMenu = false
	s.campMagicViewMenu = false
	s.campMagicMemorizeMenu = true
	s.campMagicMemorizeChar = -1
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_magic_memorize_prompt", "選擇要準備法術的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		capacity := firstLevelMemorizedCapacity(character)
		selected := len(s.pendingMemorizedSpells[index])
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("camp_magic_memorize_character", "%s（已選 %d/%d）"), character.Name, selected, capacity))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_MEM_CHAR_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_magic_memorize_exit", "返回法術選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_MEMORIZE_EXIT")
	s.Message = ""
}

func (s *State) enterCampMagicMemorizeSpellMenu(characterIndex int) {
	character := s.partyRoster[characterIndex]
	s.campMagicMemorizeChar = characterIndex
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf(s.catalog.Text("camp_magic_memorize_spell_prompt", "%s 的可用法術"), character.Name)
	s.Choices = make([]string, 0, len(character.KnownSpells)+2)
	s.currentOriginalChoices = make([]string, 0, len(character.KnownSpells)+2)
	selected := s.pendingMemorizedSpells[characterIndex]
	for index, spellID := range character.KnownSpells {
		mark := " "
		for _, selectedID := range selected {
			if selectedID == spellID {
				mark = "*"
				break
			}
		}
		s.Choices = append(s.Choices, fmt.Sprintf("%s %s", mark, campSpellLabel(s.catalog, character.Class, spellID)))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_MEM_SPELL_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_magic_mem_done", "完成選擇"), s.catalog.Text("camp_magic_mem_cancel", "取消此角色"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_MEM_DONE", "CAMP_MAGIC_MEM_CANCEL")
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
		return s.catalog.Text("fix_no_cure", "沒有已記憶的 Cure Light Wounds，隊伍未改變。")
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

// PickDungeonLock resolves one original pick-lock attempt against the loaded
// roster. The map mutation belongs to the GEO adapter because State does not
// own a particular area's wrapped grid.
func (s *State) PickDungeonLock() dungeon.PickLockResult {
	seed := s.dungeonSeed
	s.dungeonSeed++
	rng := rand.New(rand.NewSource(seed))
	return dungeon.PickLock(s.partyRoster, func() uint8 { return uint8(rng.Intn(100) + 1) })
}

// BashDungeonDoor resolves the reference strength/dice table against the
// loaded roster. GEO unlock mutation remains with the map adapter.
func (s *State) BashDungeonDoor(detail uint8) dungeon.BashResult {
	seed := s.dungeonSeed
	s.dungeonSeed++
	rng := rand.New(rand.NewSource(seed))
	return dungeon.BashDoor(s.partyRoster, detail, func(sides int) int { return rng.Intn(sides) + 1 })
}

// DungeonDoorMenuOptions exposes only the actions available to the loaded
// party for the current raw WallDoorFlags detail.
func (s *State) DungeonDoorMenuOptions(detail uint8) dungeon.DoorMenuOptions {
	return dungeon.DoorMenuOptionsFor(s.partyRoster, detail)
}

// ConsumeDungeonKnockSpell removes the first memorized Knock slot from the
// loaded roster, preserving the reference party-order transaction.
func (s *State) ConsumeDungeonKnockSpell() bool {
	updated, ok := dungeon.ConsumeSpell(s.partyRoster, dungeon.KnockSpellID)
	if !ok {
		return false
	}
	s.partyRoster = updated
	return true
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
	if originalChoice == "BAR" {
		s.enterBarMenu()
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
	s.barMenu = false
	s.shopStockMenu = false
	s.shopViewMenu = false
	s.shopTakeMenu = false
	s.shopTakeAmountMenu = false
	s.shopSellMenu = false
	s.shopSellItemMenu = false
	s.shopIdentifyMenu = false
	s.shopIdentifyItemMenu = false
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
		s.catalog.Text("shop_sell", "販售"),
		s.catalog.Text("shop_identify", "鑑定"),
		s.catalog.Text("shop_exit", "離開商店"),
	}
	s.currentOriginalChoices = []string{"BUY", "VIEW", "TAKE", "POOL", "SHARE", "APPRAISE", "SELL", "ID", "EXIT"}
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
	if strings.HasPrefix(originalChoice, "SHOP_SELL_CHARACTER_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_SELL_CHARACTER_"))
		if err != nil {
			return fmt.Errorf("invalid shop sell character command %q", originalChoice)
		}
		if value < 0 || value >= len(s.partyRoster) {
			return fmt.Errorf("shop sell character index %d is out of range", value)
		}
		s.shopSellCharacter = value
		s.enterShopSellItemMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_SELL_ITEM_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_SELL_ITEM_"))
		if err != nil {
			return fmt.Errorf("invalid shop sell item command %q", originalChoice)
		}
		if value < 0 || value >= len(s.partyRoster[s.shopSellCharacter].Equipment) {
			return fmt.Errorf("shop sell item index %d is out of range", value)
		}
		item := s.partyRoster[s.shopSellCharacter].Equipment[value]
		if err := s.SellShopItem(s.shopSellCharacter, value); err != nil {
			s.shopSellMenu = false
			s.shopSellItemMenu = false
			s.Mode = ModeEvent
			s.eventReturnMode = ModePlace
			s.OriginalEvent = "SELL"
			s.Message = "販售失敗：" + err.Error()
			return nil
		}
		s.shopSellMenu = false
		s.shopSellItemMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "SELL"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_sale_done", "已販售%s，取得 %d GP。"), monster.ChineseName(item), item.Value)
		return nil
	}
	if originalChoice == "SHOP_SELL_EXIT" {
		s.enterShopMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_IDENTIFY_CHARACTER_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_IDENTIFY_CHARACTER_"))
		if err != nil || value < 0 || value >= len(s.partyRoster) {
			return fmt.Errorf("invalid shop identify character command %q", originalChoice)
		}
		s.shopIdentifyCharacter = value
		s.enterShopIdentifyItemMenu()
		return nil
	}
	if strings.HasPrefix(originalChoice, "SHOP_IDENTIFY_ITEM_") {
		value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "SHOP_IDENTIFY_ITEM_"))
		if err != nil || value < 0 || value >= len(s.partyRoster[s.shopIdentifyCharacter].Equipment) {
			return fmt.Errorf("invalid shop identify item command %q", originalChoice)
		}
		item, identifyErr := s.IdentifyShopItem(s.shopIdentifyCharacter, value)
		s.shopIdentifyMenu = false
		s.shopIdentifyItemMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "ID"
		if identifyErr != nil {
			s.Message = "鑑定失敗：" + identifyErr.Error()
		} else {
			s.Message = fmt.Sprintf(s.catalog.Text("shop_identify_done", "已支付 %d GP 鑑定%s；完整辨識資料仍待載入。"), party.ShopIdentifyFee, monster.ChineseName(item))
		}
		return nil
	}
	if originalChoice == "SHOP_IDENTIFY_EXIT" {
		s.enterShopMenu()
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
	if s.shopStockMenu || s.shopViewMenu || s.shopTakeMenu || s.shopSellMenu || s.shopIdentifyMenu || s.shopAppraiseMenu {
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
	case "SELL":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_sell_unavailable", "目前沒有可販售物品的角色。")
		}
		s.enterShopSellCharacterMenu()
		return ""
	case "ID":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_identify_unavailable", "目前沒有可鑑定物品的角色。")
		}
		s.enterShopIdentifyCharacterMenu()
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

func (s *State) enterShopSellCharacterMenu() {
	s.shopMenu = true
	s.shopStockMenu = false
	s.shopViewMenu = false
	s.shopTakeMenu = false
	s.shopTakeAmountMenu = false
	s.shopSellMenu = true
	s.shopSellItemMenu = false
	s.shopAppraiseMenu = false
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_sell_prompt", "選擇要販售物品的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（%d 件物品）", character.Name, len(character.Equipment)))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_SELL_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_sell_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_SELL_EXIT")
	s.Message = ""
}

func (s *State) enterShopSellItemMenu() {
	s.shopSellMenu = true
	s.shopSellItemMenu = true
	s.Mode = ModePlace
	character := s.partyRoster[s.shopSellCharacter]
	s.Prompt = fmt.Sprintf(s.catalog.Text("shop_sell_item_prompt", "選擇%s要販售的物品"), character.Name)
	s.Choices = make([]string, 0, len(character.Equipment)+1)
	s.currentOriginalChoices = make([]string, 0, len(character.Equipment)+1)
	for index, item := range character.Equipment {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（%d GP）", monster.ChineseName(item), item.Value))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_SELL_ITEM_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_sell_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_SELL_EXIT")
	s.Message = ""
}

func (s *State) enterShopIdentifyCharacterMenu() {
	s.shopMenu = true
	s.shopStockMenu = false
	s.shopViewMenu = false
	s.shopTakeMenu = false
	s.shopTakeAmountMenu = false
	s.shopSellMenu = false
	s.shopSellItemMenu = false
	s.shopIdentifyMenu = true
	s.shopIdentifyItemMenu = false
	s.shopAppraiseMenu = false
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_identify_prompt", "選擇要鑑定物品的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（%d 件物品，%d GP）", character.Name, len(character.Equipment), character.Gold))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_IDENTIFY_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_identify_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_IDENTIFY_EXIT")
	s.Message = ""
}

func (s *State) enterShopIdentifyItemMenu() {
	s.shopIdentifyMenu = true
	s.shopIdentifyItemMenu = true
	s.Mode = ModePlace
	character := s.partyRoster[s.shopIdentifyCharacter]
	s.Prompt = fmt.Sprintf(s.catalog.Text("shop_identify_item_prompt", "選擇%s要鑑定的物品"), character.Name)
	s.Choices = make([]string, 0, len(character.Equipment)+1)
	s.currentOriginalChoices = make([]string, 0, len(character.Equipment)+1)
	for index, item := range character.Equipment {
		s.Choices = append(s.Choices, monster.ChineseName(item))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_IDENTIFY_ITEM_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_identify_exit", "返回商店"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_IDENTIFY_EXIT")
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
		return s.catalog.Text("bar_event", "你來到"+s.LocationName+"的酒館。")
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
	s.requestSound(SoundStep)
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
		if s.alterDropMenu {
			if s.alterDropConfirm {
				s.enterAlterDropConfirmMenu()
			} else {
				s.enterAlterDropMenu()
			}
			return nil
		}
		if s.alterMenu {
			s.enterAlterMenu()
			return nil
		}
		if s.campMagicMenu {
			s.enterCampMagicMenu()
			return nil
		}
		if s.campMagicViewMenu {
			s.enterCampMagicViewMenu()
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
		s.restoreWildernessMenu()
		return nil
	case ModePlace:
		if s.shopMenu {
			s.enterShopMenu()
			return nil
		}
		if s.barMenu {
			s.enterBarMenu()
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
	s.restoreWildernessMenu()
	s.MapX, s.MapY = 0, 0
}

func (s *State) restoreWildernessMenu() {
	s.Mode = ModeWilderness
	s.Choices = []string{
		s.catalog.Text("enter_city", "進入城市"),
		s.catalog.Text("journey_on", "繼續旅程"),
		s.catalog.Text("camp", "紮營"),
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
	case "PARLAY_HAUGHTY":
		return catalog.Text("parlay_haughty", "傲慢")
	case "PARLAY_SLY":
		return catalog.Text("parlay_sly", "狡猾")
	case "PARLAY_MEEK":
		return catalog.Text("parlay_meek", "謙卑")
	case "PARLAY_NICE":
		return catalog.Text("parlay_nice", "友善")
	case "PARLAY_ABUSIVE":
		return catalog.Text("parlay_abusive", "威嚇")
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

func localizeECLText(catalog locale.Catalog, texts []string) string {
	localized := make([]string, 0, len(texts))
	for _, text := range texts {
		localized = append(localized, localizeECLLine(catalog, text))
	}
	return strings.Join(localized, " ")
}

func localizeECLLine(catalog locale.Catalog, line string) string {
	switch line {
	case "SMOKE RISES FROM BEHIND THE RUINED WALLS":
		return catalog.Text("ecl_smoke_rises", "煙霧從殘破的牆後升起")
	case "OF YULASH. THE SOUND":
		return catalog.Text("ecl_yulash_sound", "尤拉什。聲音")
	case "OF BATTLE RINGS OUT FROM INSIDE":
		return catalog.Text("ecl_battle_rings", "從裡面傳來戰鬥聲")
	case "YOU SEE THREE CULTISTS LYING DEAD ON THE FLOOR.":
		return catalog.Text("ecl_cultists_dead", "你們看見三名邪教徒倒臥在地板上。")
	case "JUST AHEAD OF YOU, ANOTHER CLERIC GASPS FOR BREATH.":
		return catalog.Text("ecl_wounded_cleric", "就在前方，另一名牧師喘著氣。")
	case "THE WOUNDED CLERIC'S EYES WIDEN IN FANATIC":
		return catalog.Text("ecl_cleric_fanatic", "受傷牧師的雙眼因狂熱而睜大。")
	case "TRIUMPH. HE HOWLS,":
		return catalog.Text("ecl_cleric_howl", "勝利。他嚎叫著：")
	case "YOU FIND A WAR BLASTED SECTION OF THE CITY.":
		return catalog.Text("ecl_war_blasted_city", "你們找到城市中一片遭戰火摧毀的區域。")
	case "YOU DISCOVER A SMALL MAGIC SHOP.":
		return catalog.Text("ecl_small_magic_shop", "你們發現一間小型魔法商店。")
	default:
		return line
	}
}
