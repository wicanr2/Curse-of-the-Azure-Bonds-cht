// Package game contains platform-neutral remake state. Rendering and input
// adapters (Ebiten or a test harness) call Apply; no DOS assumptions belong
// here.
package game

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dungeon"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/locale"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/mapdata"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
	engineaction "github.com/wicanr2/golden-box-remake-engine/combat/action"
	enginescan "github.com/wicanr2/golden-box-remake-engine/combat/scan"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
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
	ModeDungeon
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
	LocationTilverton
	LocationStandingStone
	LocationEssembra
	LocationHap
	LocationVoonlar
	LocationPhlan
	LocationTeshwave
	LocationYulash
	LocationHillsfar
	LocationZhentilKeep
	LocationMythDrannor
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

	catalog                 locale.Catalog
	eclBlock                []byte
	eclStart                int
	selectionSequence       []uint16
	whoSelectionSequence    []uint16
	whoMenu                 bool
	whoSelectedIndex        int
	loadCharacterNotFound   bool
	loadCharacterHighBit    bool
	currentOriginalChoices  []string
	eventReturnMode         Mode
	journalReturnMode       Mode
	creationReturnMode      Mode
	session                 *ecl.BlockSession
	pendingPictureResult    *ecl.RunResult
	pendingECLMenu          *ecl.Menu
	pendingECLMenuMessage   string
	eclStringEditing        bool
	eclStringValue          string
	eclStringMaxLength      int
	pendingDungeonEntry     bool
	dungeonBoundaryAttempt  bool
	pendingWorldDestination uint8
	pendingWorldTravel      bool
	dataPack                *goldenbox.Pack
	dataPackError           error
	appliedDataPackEvents   map[string]bool
	newGameEntryActive      bool
	eclMenuReturnMode       Mode
	party                   []combat.Fighter
	partyRoster             party.Roster
	savgamPrefix            *partySave.SAVGAMContainer
	savgamPlayers           map[string]party.DOSPlayerFiles
	pendingSoundEvents      []SoundEvent
	pendingMusicEvents      []MusicEvent
	activeMusicTrackID      string
	musicPlaybackSnapshot   *pc98music.TrackPCMStreamSnapshot
	oneShotPlaybackSnapshot *audiostate.Snapshot
	pendingECLCalls         []uint16
	battle                  *combat.Battle
	combatTurns             []combat.Turn
	combatTurnIndex         int
	combatDelayedTurns      map[int]bool
	combatTargetIndex       int
	combatCastingSpell      uint8
	combatCastingClass      party.Class
	combatCastingClassSet   bool
	combatSpellTargetIndex  int
	combatSpellTargetPoint  combat.TilePoint
	combatSpellTargetsPoint bool
	combatMoveMode          bool
	combatMoveRemaining     int
	combatSpeed             engineaction.Speed
	combatQuickMagic        bool
	combatReferenceCoords   bool
	combatLineTerrain       combat.LineTerrain
	combatScanMapProvider   func() (enginescan.TacticalMap, error)
	combatView              bool
	combatViewFighterID     string
	combatMessage           string
	combatReturnMode        Mode
	combatVisualEnabled     bool
	combatVisualSerial      uint64
	combatVisual            *combat.VisualEvent
	combatVisualElapsed     time.Duration
	combatVisualTravelSent  bool
	combatVisualImpactSent  int
	combatVisualDeathSent   int
	combatVisualAdvanceTurn bool
	monsterRecords          map[uint8]monster.Record
	monsterRecordsByECL     map[uint8]map[uint8]monster.Record
	monsterAffects          map[uint8][]monster.AffectRecord
	monsterAffectsByECL     map[uint8]map[uint8][]monster.AffectRecord
	monsterItemsByECL       map[uint8]map[uint8][]monster.ItemRecord
	gameClock               [7]uint16
	gameAgeCycles           uint32
	itemCatalog             monster.BaseItemCatalog
	itemCatalogReady        bool
	ammunitionItemTypes     map[uint8][]uint8
	combatSeed              int64
	eclSeed                 int64
	mapSeed                 int64
	geoMapPending           bool
	loadPiecesPending       bool
	pendingSpellSearches    []ecl.SpellSearch
	pendingDamageRequests   []ecl.DamageRequest
	pendingProtection       []uint16
	pendingTreasure         []ecl.TreasureRequest
	treasureItemBlocks      map[uint16][]monster.ItemRecord
	pendingTreasureItems    []monster.ItemRecord
	pendingTreasureMessage  string
	treasureMenu            bool
	treasureTakeMenu        bool
	treasureItemIndex       int
	treasureReturnMode      Mode
	treasureResumeECL       bool
	shopMenu                bool
	shopECLService          bool
	templeMenu              bool
	templeHealMenu          bool
	templeConfirmMenu       bool
	templeECLService        bool
	templeCharacterIndex    int
	templePendingCure       int
	trainingMenu            bool
	trainingConfirmMenu     bool
	trainingSpellMenu       bool
	trainingCharacterIndex  int
	trainingSpellChoices    []uint8
	trainingResult          string
	shopOffers              []ShopOffer
	moneyPool               uint32
	moneyCopperRemainder    uint16
	treasureGems            uint32
	treasureJewelry         uint32
	appraisalOffers         AppraisalOffers
	shopStockMenu           bool
	shopViewMenu            bool
	shopTakeMenu            bool
	shopTakeAmountMenu      bool
	shopSellMenu            bool
	shopSellItemMenu        bool
	shopIdentifyMenu        bool
	shopIdentifyItemMenu    bool
	shopAppraiseMenu        bool
	shopAppraiseConfirm     bool
	shopCharacterIndex      int
	shopTakeCharacter       int
	shopSellCharacter       int
	shopIdentifyCharacter   int
	shopAppraiseCharacter   int
	shopAppraiseKind        TreasureKind
	barMenu                 bool
	parlayMenu              bool
	barTales                []string
	barTaleIndex            int
	campMenu                bool
	campECLService          bool
	campReturnMode          Mode
	campRestMenu            bool
	restHours               int
	restEncounterPeriod     uint16
	restEncounterPercent    uint16
	campViewMenu            bool
	campMagicMenu           bool
	campMagicViewMenu       bool
	campMagicMemorizeMenu   bool
	campMagicMemorizeChar   int
	campMagicCastMenu       bool
	campMagicCastChar       int
	campMagicCastSpell      uint8
	pendingMemorizedSpells  map[int][]uint8
	saveRequested           bool
	programEndMenu          bool
	gameWon                 bool
	partyKilled             bool
	alterMenu               bool
	alterOrderMenu          bool
	alterOrderSelected      int
	alterDropMenu           bool
	alterDropConfirm        bool
	alterDropSelected       int
	alterRenameMenu         bool
	alterRenameChar         int
	renameEditing           bool
	renameCharacter         int
	renameName              string
	alterPicsMenu           bool
	alterSpeedMenu          bool
	alterIconMenu           bool
	alterIconEdit           bool
	alterIconCharacter      int
	alterIconHeadIndex      int
	alterIconBodyIndex      int
	picturesEnabled         bool
	animationsEnabled       bool
	messageSpeed            int
	fixSeed                 int64
	dungeonSeed             int64
	combatMapDirection      uint8
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
	dataPack, dataPackErr := gamepack.Default()
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
		dataPack:               dataPack,
		dataPackError:          dataPackErr,
		appliedDataPackEvents:  make(map[string]bool),
		barTales: []string{
			catalog.Text("bar_tale_1", "bar_tale_1"),
			catalog.Text("bar_tale_2", "bar_tale_2"),
			catalog.Text("bar_tale_3", "bar_tale_3"),
			catalog.Text("bar_tale_4", "bar_tale_4"),
			catalog.Text("bar_tale_5", "bar_tale_5"),
			catalog.Text("bar_tale_6", "bar_tale_6"),
		},
		restHours:         24,
		combatSeed:        1,
		eclSeed:           1,
		mapSeed:           1,
		picturesEnabled:   true,
		animationsEnabled: true,
		messageSpeed:      3,
		combatSpeed:       engineaction.DefaultSpeed,
		fixSeed:           1,
		dungeonSeed:       1,
		GeoMapSet:         2,
		GeoMapBlock:       1,
		// Reference seg001.Init: mapPosX=7, mapPosY=0x0D, direction=0.
		DungeonX:         7,
		DungeonY:         13,
		DungeonDirection: 0,
		whoSelectedIndex: -1,
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

// SelectedPlayerID exposes the last character selected by ECL WHO or LOAD
// CHARACTER. An empty string means no valid character has been committed.
func (s *State) SelectedPlayerID() string {
	if s.whoSelectedIndex < 0 || s.whoSelectedIndex >= len(s.partyRoster) {
		return ""
	}
	return s.partyRoster[s.whoSelectedIndex].ID
}

// LoadCharacterNotFound reports the last LOAD CHARACTER lookup result.
func (s *State) LoadCharacterNotFound() bool { return s.loadCharacterNotFound }

// LoadCharacterHighBit reports the reference restore/redraw flag on the last
// LOAD CHARACTER value.
func (s *State) LoadCharacterHighBit() bool { return s.loadCharacterHighBit }

// eclPartyContext projects the currently loaded roster into the small,
// renderer-neutral context required by reference PARTY commands. Combat
// projection is preferred when available because it includes readied effects;
// the character projection remains the deterministic fallback outside combat.
func (s *State) eclPartyContext() ecl.PartyContext {
	context := ecl.PartyContext{Members: make([]ecl.PartyMemberContext, 0, len(s.partyRoster))}
	for _, character := range s.partyRoster {
		scriptName := character.ScriptName
		if scriptName == "" {
			scriptName = character.Name
		}
		member := ecl.PartyMemberContext{
			Name:           scriptName,
			ControlMorale:  character.ControlMorale,
			HitPoints:      character.HitPoints,
			ClericLevel:    character.ClassLevel(party.ClassCleric),
			MagicUserLevel: character.ClassLevel(party.ClassMagicUser),
			HasRangerClass: character.HasClass(party.ClassRanger),
		}
		for _, item := range character.Equipment {
			member.ItemTypes = append(member.ItemTypes, item.Type)
		}
		copy(member.ThiefSkills[:], character.ThiefSkills)
		for _, fighter := range s.party {
			if fighter.ID == character.ID {
				member.HitPoints = fighter.HitPoints
				member.ArmorClass = fighter.ArmorClass
				member.AttackBonus = fighter.AttackBonus
				member.MovementAllowance = fighter.MovementAllowance
				break
			}
		}
		for _, effect := range character.Effects {
			if effect.Active {
				member.Effects = append(member.Effects, effect.Kind)
			}
		}
		context.Members = append(context.Members, member)
	}
	return context
}

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
	s.Prompt = s.catalog.Text("bar_menu_prompt", "bar_menu_prompt")
	s.Choices = []string{
		s.catalog.Text("bar_listen", "bar_listen"),
		s.catalog.Text("bar_exit", "bar_exit"),
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
			s.Message = s.catalog.Text("bar_no_tales", "bar_no_tales")
			return nil
		}
		taleNumber := s.barTaleIndex + 1
		s.Message = fmt.Sprintf(s.catalog.Text("bar_tale", "bar_tale"), taleNumber, s.barTales[s.barTaleIndex])
		s.barTaleIndex++
		return nil
	case "BAR_EXIT":
		s.barMenu = false
		s.Message = s.catalog.Text("bar_exit_message", "bar_exit_message")
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
	return s.restPartyHours(s.restHours)
}

func (s *State) restPartyHours(hours int) int {
	healed := hours / 24
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

func (s *State) restInterruption(requestedHours int) (completedHours int, interrupted bool) {
	if requestedHours <= 0 || s.restEncounterPeriod == 0 || s.restEncounterPercent == 0 {
		return requestedHours, false
	}
	rng := rand.New(rand.NewSource(s.eclSeed))
	period := int(s.restEncounterPeriod)
	for completed := period; completed <= requestedHours; completed += period {
		if rng.Intn(100)+1 <= int(s.restEncounterPercent) {
			return completed, true
		}
	}
	return requestedHours, false
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

// SetMonsterItemsForECL installs the chapter-local MON*ITM sidecars used by
// both encounter equipment and CMD_AddNPC's load_mob transaction.
func (s *State) SetMonsterItemsForECL(chapter uint8, items map[uint8][]monster.ItemRecord) {
	if chapter < 1 || chapter > 6 {
		return
	}
	if s.monsterItemsByECL == nil {
		s.monsterItemsByECL = make(map[uint8]map[uint8][]monster.ItemRecord)
	}
	copyItems := make(map[uint8][]monster.ItemRecord, len(items))
	for id, records := range items {
		copyItems[id] = append([]monster.ItemRecord(nil), records...)
	}
	s.monsterItemsByECL[chapter] = copyItems
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

func (s *State) monsterItemsForCurrentECL() map[uint8][]monster.ItemRecord {
	if s.session == nil {
		return nil
	}
	return s.monsterItemsByECL[monsterChapterForBlock(s.session.CurrentBlockID())]
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

// SetECLSeed controls the deterministic RANDOM stream while replaying an
// event path. BlockSession retains the generator across ECL invocations, so
// revisiting a random terrain consumes the next roll instead of restarting.
func (s *State) SetECLSeed(seed int64) {
	s.eclSeed = seed
	if s.session != nil {
		s.session.ResetRandomSeed(seed)
	}
}

// SetECLMemoryValue seeds one verified engine/script work word for a
// reproducible story preview. Normal gameplay obtains these values from ECL.
func (s *State) SetECLMemoryValue(address, value uint16) {
	if s.session != nil {
		s.session.SetMemoryValue(address, value)
	}
}

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
	s.gameClock = value.GameTime
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
		s.requestSound(SoundOverture)
		if s.session != nil && s.session.HasBlock(0x01) && len(s.party) == 0 {
			return s.OpenCharacterCreation()
		}
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

// BeginAdventure follows the non-demo new-game dispatch in sub_29758:
// LastEclBlockId==0 selects global ECL block 0x01, resets VM state, runs its
// fifth vm_init_ecl entry, and pauses at the first data-driven menu.
func (s *State) BeginAdventure() error {
	if len(s.partyRoster) == 0 || len(s.party) == 0 {
		return fmt.Errorf("adventure requires a created or loaded party")
	}
	if s.session == nil {
		return fmt.Errorf("adventure requires a global ECL session")
	}
	if err := s.session.Reset(0x01); err != nil {
		return err
	}
	s.requestMusicForCurrentBlock("")
	s.eclBlock = s.session.CurrentData()
	start, err := s.session.InitialEntry()
	if err != nil {
		return err
	}
	s.eclStart = start
	s.selectionSequence = nil
	s.whoSelectionSequence = nil
	s.whoMenu = false
	s.currentOriginalChoices = nil
	s.Choices = nil
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.eclMenuReturnMode = ModeTitle
	s.newGameEntryActive = true
	// seg001.Init/InitAgain establish a fresh campaign as an indoor map before
	// sub_29758 selects and runs global block 0x01.
	s.Area.InDungeon = true

	result, err := s.session.RunInteractiveSeedWithPartyContextAndWhoSelections(
		180, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
	s.eclBlock = s.session.CurrentData()
	s.applyGeoMapLoad(result)
	s.applyLoadPieces(result)
	s.applyECLCallSignals(result)
	s.applySpellSignals(result)
	s.applyECLDamageSignals(result)
	s.applyECLLoadCharacterSignals(result)
	if err := s.applyECLNPCSignals(result); err != nil {
		return err
	}
	if err := s.applyECLDumpSignals(result); err != nil {
		return err
	}
	if err := s.applyECLClockSignals(result); err != nil {
		return err
	}
	s.applyECLInventorySignals(result)
	s.applyECLTreasureSignals(result)
	s.applyECLRobSignals(result)
	if len(result.Text) > 0 {
		s.unlockJournalEntries(result.Text)
		s.Message = s.localizeECLText(result.Text)
	}
	if result.WaitingForString && len(result.StringInputRequests) > 0 {
		s.beginECLStringInput(result.StringInputRequests[len(result.StringInputRequests)-1])
		return nil
	}
	if result.PictureRequested {
		s.requestMusicForSignal("picture", result.PictureBlock)
		s.PictureRequested = true
		s.PictureBlock = result.PictureBlock
		s.BigPictureRequested = result.BigPictureRequested
		if result.PictureHeadBlockSet {
			s.SceneHeadBlock = uint8(result.PictureHeadBlock)
		}
		s.SceneCharacterRequested = !result.BigPictureRequested && s.SceneHeadBlock != 0xFF
		if s.SceneCharacterRequested {
			s.SceneBodyBlock = uint8(result.PictureBlock)
		}
		s.OriginalEvent = "PICTURE"
		if result.CombatRequested || result.WaitingForMenu || result.WaitingForString {
			pending := result
			pending.PictureRequested = false
			s.pendingPictureResult = &pending
		}
		return nil
	}
	if result.CombatRequested {
		records := s.monsterRecordsForCurrentECL()
		if len(result.MonsterSpawns) > 0 && len(records) > 0 {
			return s.StartEncounterWithAffects(result, records, s.monsterAffectsForCurrentECL(), s.party, s.combatSeed)
		}
		s.OriginalEvent = "COMBAT"
		return nil
	}
	if result.WaitingForMenu && len(result.Menus) > 0 {
		s.enterECLMenu(result.Menus[len(result.Menus)-1])
		return nil
	}
	s.OriginalEvent = "NEW GAME"
	return nil
}

func (s *State) enterECLMenu(menu ecl.Menu) {
	if slices.Equal(menu.Options, []string{"PATROL", "FOREST", "JOURNEY ON", "CAMP"}) {
		menu.Options = []string{"PATROL FOREST", "JOURNEY ON", "CAMP"}
	}
	if slices.Contains(menu.Options, "ENTER CITY") {
		s.pendingWorldTravel = false
	}
	s.syncWorldDestinationSelectors(menu.Options)
	s.Choices = make([]string, 0, len(menu.Options))
	s.currentOriginalChoices = append([]string(nil), menu.Options...)
	for _, option := range menu.Options {
		s.Choices = append(s.Choices, s.localizeOption(option))
	}
	if menu.Prompt != "" {
		s.Prompt = s.localizeECLPrompt(menu.Prompt)
	} else {
		s.Prompt = s.catalog.Text("press_button", "請按任意鍵或 Enter 繼續")
	}
	s.Mode = ModeWilderness
}

func (s *State) localizeECLPrompt(prompt string) string {
	if s.dataPack != nil {
		if result := s.dataPack.MatchText([]string{prompt}, s.catalog.Language); result.Matched {
			return result.Message
		}
	}
	return s.localizePrompt(prompt)
}

// syncWorldDestinationSelectors projects the title-declared route graph into
// the four legacy work cells consumed by ECL1's route dispatcher. The UI
// keeps zero-based menu indices; each work value is the original world
// location byte copied from ECL1's adjacency row.
func (s *State) syncWorldDestinationSelectors(options []string) {
	if s.session == nil || s.dataPack == nil {
		return
	}
	definition, found := s.dataPack.FindMapByKind("overland")
	if !found {
		return
	}
	var current goldenbox.MapPoint
	for _, point := range definition.Locations {
		if point.Value == s.Area.CurrentCity {
			current = point
			break
		}
	}
	// Conditional ECL branches may deliberately hide destinations. Their
	// compact selector table is already supplied by the original program;
	// only project JSON when the complete declared adjacency row is visible.
	if len(options) != len(current.Destinations) {
		return
	}
	for index, option := range options {
		point, ok := s.worldPointForOriginalOption(definition, option)
		if !ok {
			return
		}
		s.session.SetMemoryValue(0x4C02+uint16(index), uint16(point.Value))
	}
}

func (s *State) worldPointForOriginalOption(definition goldenbox.MapDefinition, option string) (goldenbox.MapPoint, bool) {
	english := s.dataPack.Locales["en"]
	for _, point := range definition.Locations {
		if strings.EqualFold(strings.TrimSpace(english[point.MessageID]), strings.TrimSpace(option)) {
			return point, true
		}
	}
	return goldenbox.MapPoint{}, false
}

// Select applies a localized opening choice and, when the state came from an
// ECL block, runs that choice through the bounded ECL subset.
func (s *State) Select(index int) error {
	if s.Mode == ModeMap {
		return fmt.Errorf("choice %d is invalid in map mode", index)
	}
	whoSelecting := s.whoMenu
	if (s.Mode != ModeWilderness && s.Mode != ModePlace) || index < 0 || index >= len(s.Choices) {
		return fmt.Errorf("choice %d is invalid in mode %d", index, s.Mode)
	}
	originalChoice := ""
	if index < len(s.currentOriginalChoices) {
		originalChoice = s.currentOriginalChoices[index]
	}
	messageBeforeSelection := s.Message
	if s.dataPack != nil {
		if definition, found := s.dataPack.FindMapByKind("overland"); found {
			if point, ok := s.worldPointForOriginalOption(definition, originalChoice); ok {
				s.pendingWorldDestination = point.Value
				s.pendingWorldTravel = true
			}
		}
	}
	if s.session != nil && s.pendingWorldTravel && isWorldTravelRouteChoice(originalChoice) {
		// AREA publishes the selected destination through this separate
		// arrival cell. ECL may overwrite 4C9B while dispatching the route.
		s.session.SetMemoryValue(0x4C9C, uint16(s.pendingWorldDestination))
	}
	if s.treasureMenu {
		return s.selectTreasure(index, originalChoice)
	}
	if s.parlayMenu {
		return s.selectParlay(index, originalChoice)
	}
	if s.Mode == ModePlace {
		if s.templeMenu {
			return s.selectTemple(originalChoice)
		}
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
	if s.trainingMenu {
		return s.selectTraining(originalChoice)
	}
	if s.programEndMenu {
		return s.selectProgramEnd(originalChoice)
	}
	if whoSelecting {
		s.whoMenu = false
		s.whoSelectionSequence = append(s.whoSelectionSequence, uint16(index))
		s.Message = fmt.Sprintf("已選擇角色：%s", s.Choices[index])
	}
	if originalChoice == "CAMP" && len(s.eclBlock) == 0 {
		s.enterCampMenu()
		return nil
	}
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	if s.eclMenuReturnMode != ModeTitle {
		s.eventReturnMode = s.eclMenuReturnMode
	}
	if len(s.eclBlock) == 0 && originalChoice == "FLEE" {
		s.OriginalEvent = "FLEE"
		s.Message = s.catalog.Text("encounter_flee_done", "你們成功撤退，返回荒野。")
		return nil
	}
	if len(s.eclBlock) == 0 && originalChoice == "PARLAY" {
		s.enterParlayMenu()
		s.Mode = ModeWilderness
		return nil
	}
	if len(s.eclBlock) > 0 {
		s.Message = s.Choices[index]
	} else {
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
	}
	if len(s.eclBlock) > 0 {
		if !whoSelecting {
			s.selectionSequence = append(s.selectionSequence, uint16(index))
		}
		var result ecl.RunResult
		var runErr error
		var blockBefore uint8
		if s.session != nil {
			blockBefore = s.session.CurrentBlockID()
			if blockBefore == 0x10 && originalChoice == "GO WITH GUARDS" {
				// Yulash's outer dispatcher teleports the party to the Red
				// Plume commander's waiting room at (0,3), facing east.
				// The ECL branch owns the dialogue but the DOS area loop owns
				// these map registers, so reproduce that engine transaction
				// before resuming the shared block.
				s.session.SetMemoryValue(0xC04B, 0)
				s.session.SetMemoryValue(0xC04C, 3)
				s.session.SetMemoryValue(0xC04D, 1)
			}
			dungeonRegistersBefore := [3]uint16{}
			dungeonRegistersKnown := [3]bool{}
			for i, address := range [...]uint16{0xC04B, 0xC04C, 0xC04D} {
				dungeonRegistersBefore[i], dungeonRegistersKnown[i] =
					s.session.MemoryValue(address)
			}
			selection := uint16(index)
			var menuSelection, whoSelection *uint16
			if whoSelecting {
				whoSelection = &selection
			} else {
				menuSelection = &selection
			}
			result, runErr = s.session.ResumeInteractiveSelectionSeed(180, menuSelection, whoSelection, s.eclSeed, s.eclPartyContext())
			s.eclBlock = s.session.CurrentData()
			if start, err := s.session.InitialEntry(); err == nil {
				s.eclStart = start
			}
			currentBlock := s.session.CurrentBlockID()
			s.requestMusicIfBlockChanged(blockBefore)
			if blockBefore != currentBlock && s.Area.InDungeon {
				// NEWECL destinations may assign a new C04B/C04C/C04D
				// starting position. Only project registers actually changed
				// by this resumed branch: some destinations intentionally
				// inherit a location selected by the outer area loop, while
				// stale work-register values remain visible to the VM.
				dungeonRegistersChanged := false
				for i, address := range [...]uint16{0xC04B, 0xC04C, 0xC04D} {
					value, known := s.session.MemoryValue(address)
					if known != dungeonRegistersKnown[i] ||
						(known && value != dungeonRegistersBefore[i]) {
						dungeonRegistersChanged = true
						break
					}
				}
				var definition goldenbox.MapDefinition
				destinationDeclared := false
				if s.dataPack != nil {
					definition, destinationDeclared = s.dataPack.FindMapByKindScript(
						"first_person", s.Area.GameArea, currentBlock,
					)
				}
				registerOwnedDestination := destinationDeclared && definition.Spawn == nil
				hapRegisterOwnedDestination := s.Area.GameArea == 5 &&
					(currentBlock == 0x31 || currentBlock == 0x32 || currentBlock == 0x33)
				if dungeonRegistersChanged &&
					(registerOwnedDestination || hapRegisterOwnedDestination) {
					s.syncDungeonStateFromECLRegisters()
				}
			}
			if blockBefore == 0x31 && s.session.CurrentBlockID() == 0x32 {
				// The verified Hap CAVES transition starts a fresh map
				// movement cycle. Do not leak the village exit-attempt and
				// forced-move work bytes into the lava-tube post-combat path.
				s.session.SetMemoryValue(0x7ED5, 0)
				s.session.SetMemoryValue(0x7EC9, 0)
			}
			if blockBefore == 0x11 && s.session.CurrentBlockID() == 0x12 &&
				s.Area.InDungeon && s.Area.GameArea == 3 {
				// Pit of Moander levels share GEO3 block 0x11. The
				// destination ECL initial entry places the party on the
				// lower-level landing at (15,14), facing south.
				s.DungeonX, s.DungeonY, s.DungeonDirection = 15, 14, 4
				s.MapX, s.MapY = 15, 14
				s.session.SetMemoryValue(0xC04B, 15)
				s.session.SetMemoryValue(0xC04C, 14)
				s.session.SetMemoryValue(0xC04D, 2)
			}
			if s.dungeonBoundaryAttempt &&
				(blockBefore != currentBlock || result.Exited) {
				// The resumed source-map branch has consumed its boundary
				// attempt. Never expose that work signal to the destination
				// block's next per-turn entry or to ordinary movement after
				// TURN BACK.
				s.session.SetMemoryValue(0x7ED5, 0)
				s.dungeonBoundaryAttempt = false
			}
			// Global ENTER CITY is an engine-level area transition wrapped
			// around ECL's NEWECL. The destination namespace selects the
			// chapter's ECL/GEO resources; its story may pause several times
			// before an empty EXIT finally hands control to dungeon movement.
			if originalChoice == "ENTER CITY" && blockBefore >= 0x50 && currentBlock < 0x50 {
				gameArea := monsterChapterForBlock(currentBlock)
				s.Area.GameArea = gameArea
				s.Area.InDungeon = true
				s.GeoMapSet = gameArea
				s.GeoMapBlock = currentBlock
				s.geoMapPending = true
				s.pendingDungeonEntry = true
				s.eclMenuReturnMode = ModeDungeon
				s.eventReturnMode = ModeDungeon
			}
		} else {
			result, runErr = ecl.RunSubsetInteractiveSeedWithPartyContextAndWhoSelections(s.eclBlock, s.eclStart, 180, s.selectionSequence, s.whoSelectionSequence, s.eclSeed, s.eclPartyContext())
		}
		if runErr != nil {
			return runErr
		}
		s.applyGeoMapLoad(result)
		if s.session != nil && blockBefore != s.session.CurrentBlockID() &&
			s.Area.InDungeon && s.dataPack != nil {
			currentBlock := s.session.CurrentBlockID()
			// A title pack can prove that a NEWECL destination owns a
			// matching first-person geometry block. Prefer that declaration
			// over stale LOAD FILES aggregation from the source ECL; areas
			// whose ECL blocks intentionally share one GEO map simply omit
			// the destination definition.
			if definition, found := s.dataPack.FindMapByKindScript(
				"first_person", s.Area.GameArea, currentBlock,
			); found {
				s.GeoMapSet = definition.AreaID
				s.GeoMapBlock = definition.GeometryBlock
				s.geoMapPending = true
			}
		}
		if s.session != nil && blockBefore != s.session.CurrentBlockID() &&
			s.Area.InDungeon && s.Area.GameArea == 5 {
			currentBlock := s.session.CurrentBlockID()
			if currentBlock == 0x31 || currentBlock == 0x32 || currentBlock == 0x33 {
				// Hap village, caves, and wizard-tower roof use matching
				// ECL/GEO block IDs. Resolve the destination after LOAD FILES
				// aggregation, which can still contain the source block.
				s.GeoMapSet = s.Area.GameArea
				s.GeoMapBlock = currentBlock
				s.geoMapPending = true
			}
		}
		s.applyLoadPieces(result)
		s.applyECLCallSignals(result)
		s.applySpellSignals(result)
		s.applyECLDamageSignals(result)
		if _, err := s.resolveAutomaticWholePartyECLDamage(); err != nil {
			return err
		}
		s.applyECLLoadCharacterSignals(result)
		if err := s.applyECLNPCSignals(result); err != nil {
			return err
		}
		if err := s.applyECLDumpSignals(result); err != nil {
			return err
		}
		if err := s.applyECLClockSignals(result); err != nil {
			return err
		}
		s.applyECLInventorySignals(result)
		s.applyECLTreasureSignals(result)
		s.applyECLRobSignals(result)
		treasureReady := false
		if len(result.TreasureRequests) > 0 {
			// Some encounter scripts queue their reward immediately before
			// COMBAT. Keep that request raw until the party actually wins.
			// COMBAT without monster spawns is the separate treasure-service
			// boundary and must still open its loot menu immediately.
			deferUntilVictory := result.CombatRequested && len(result.MonsterSpawns) > 0
			if deferUntilVictory {
				treasureReady = false
			} else {
				beforeMoney := s.moneyPool
				beforeGems, beforeJewelry := s.treasureGems, s.treasureJewelry
				beforeItems := len(s.pendingTreasureItems)
				if err := s.ResolveTreasureRequests(); err != nil {
					// A headless/test adapter may not have loaded ITEM*.DAX yet.
					// Keep the raw request pending and let the ECL control flow reach
					// its next command (including COMBAT) instead of aborting it.
					s.Message = "財寶等待素材載入：" + err.Error()
				} else {
					// TREASURE may contain only coins, gems, or jewelry.
					// Open the service when actual pooled content changed,
					// but do not insert an empty page for a zero request.
					treasureReady = result.CombatRequested &&
						(s.moneyPool != beforeMoney ||
							s.treasureGems != beforeGems ||
							s.treasureJewelry != beforeJewelry ||
							len(s.pendingTreasureItems) != beforeItems)
				}
			}
		}
		if s.session != nil && s.pendingWorldTravel && isWorldTravelRouteChoice(originalChoice) {
			// AREA owns the actual movement. ECL1 may reuse its route-selector
			// work value while dispatching a trail encounter, so project the
			// chosen location only after that ECL dispatch returns.
			s.session.SetMemoryValue(0x4C9B, uint16(s.pendingWorldDestination))
		}
		s.applyCitySelection()
		if len(result.Text) > 0 {
			s.unlockJournalEntries(result.Text)
			s.Message = s.localizeECLText(result.Text)
		}
		if len(result.WhoRequests) > 0 {
			request := result.WhoRequests[len(result.WhoRequests)-1]
			if request.SelectionProvided {
				selected := int(request.Selected)
				if selected < 0 || selected >= len(s.partyRoster) {
					return fmt.Errorf("WHO selected character %d is outside party roster", selected)
				}
				s.whoSelectedIndex = selected
			}
		}
		if result.WaitingForWho {
			s.whoMenu = true
			s.Choices = make([]string, 0, len(s.partyRoster))
			s.currentOriginalChoices = make([]string, 0, len(s.partyRoster))
			for _, character := range s.partyRoster {
				s.Choices = append(s.Choices, character.Name)
				s.currentOriginalChoices = append(s.currentOriginalChoices, character.ID)
			}
			if len(result.WhoRequests) > 0 && result.WhoRequests[len(result.WhoRequests)-1].Prompt != "" {
				s.Prompt = s.localizeECLText([]string{result.WhoRequests[len(result.WhoRequests)-1].Prompt})
			} else {
				s.Prompt = "請選擇角色"
			}
			s.Mode = ModeWilderness
			return nil
		}
		if result.WaitingForString && len(result.StringInputRequests) > 0 {
			s.beginECLStringInput(result.StringInputRequests[len(result.StringInputRequests)-1])
			return nil
		}
		if result.PictureRequested {
			s.requestMusicForSignal("picture", result.PictureBlock)
			if !s.picturesEnabled {
				s.PictureRequested = false
				s.PictureBlock = result.PictureBlock
				s.OriginalEvent = "PICTURE"
				s.Message = s.catalog.Text("pics_monsters_off_message", "遭遇圖片已關閉。")
				if handled, err := s.continueAfterSuppressedPicture(result); handled || err != nil {
					return err
				}
				return nil
			}
			s.PictureRequested = true
			s.PictureBlock = result.PictureBlock
			s.BigPictureRequested = result.BigPictureRequested
			if result.PictureHeadBlockSet {
				s.SceneHeadBlock = uint8(result.PictureHeadBlock)
			}
			s.SceneCharacterRequested = !result.BigPictureRequested && s.SceneHeadBlock != 0xFF
			if s.SceneCharacterRequested {
				s.SceneBodyBlock = uint8(result.PictureBlock)
			}
			s.OriginalEvent = "PICTURE"
			if s.Message == "" {
				s.Message = "事件畫面"
			}
			if result.CombatRequested || result.ShopRequested || result.TempleRequested || result.WaitingForMenu {
				pending := result
				pending.PictureRequested = false
				s.pendingPictureResult = &pending
			}
			return nil
		}
		if result.ShopRequested {
			return s.enterECLShop(result)
		}
		if result.TempleRequested {
			return s.enterECLTemple()
		}
		// TREASURE followed by COMBAT is the reference treasure-service
		// dispatch, not a monster battle. Resolve and present loot first.
		if treasureReady {
			s.treasureResumeECL = s.session != nil && len(s.eclBlock) > 0
			if s.eclMenuReturnMode == ModeDungeon {
				s.enterTreasureMenuFor(ModeDungeon)
			} else {
				s.enterTreasureMenu()
			}
			if hasMeaningfulECLText(result.Text) {
				s.Message = s.localizeECLText(result.Text)
			} else if strings.TrimSpace(messageBeforeSelection) != "" {
				s.Message = messageBeforeSelection
			}
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
			if len(result.MonsterSpawns) == 0 && s.session != nil &&
				s.eclMenuReturnMode == ModeDungeon {
				// A real dungeon script may deliberately reduce every encounter
				// group count to zero before COMBAT (for example after allies
				// remove the opposition). The DOS scheduler wins that empty
				// battle immediately; resume the saved ECL PC instead of showing
				// the unsupported-combat fallback. Synthetic/no-session combat
				// remains fail-closed below.
				s.combatReturnMode = ModeDungeon
				continued, continueErr := s.continueECLAfterEngineBoundary()
				if continueErr != nil {
					return continueErr
				}
				if !continued {
					s.Mode = ModeDungeon
					s.eventReturnMode = ModeDungeon
					s.Message = ""
				}
				return nil
			}
			s.OriginalEvent = "COMBAT"
			s.Message = s.catalog.Text("combat_started", "戰鬥開始（戰鬥規則尚未完成）")
			s.eventReturnMode = ModeWilderness
			s.Mode = ModeEvent
			return nil
		}
		if isTrainingProgramChoice(originalChoice, result) {
			s.enterTrainingMenu()
			return nil
		}
		if handled, err := s.applyECLProgram(result); handled || err != nil {
			return err
		}
		// WILDERNESS/EXIT is the observed Shadowdale map-entry menu. Handle
		// these semantic transitions before the bounded runner's next-menu
		// result is applied, since the original command may leave another
		// continuation menu in the trace.
		if s.Location != LocationWilderness && originalChoice == "WILDERNESS" &&
			!(s.session != nil && s.session.CurrentBlockID() == 0x33) {
			s.enterMap()
			return nil
		}
		if s.Location != LocationWilderness && originalChoice == "EXIT" {
			s.leaveLocation()
			return nil
		}
		if result.WaitingForMenu && len(result.Menus) > 0 {
			s.enterECLMenu(result.Menus[len(result.Menus)-1])
			return nil
		}
		if result.Exited && s.newGameEntryActive && s.session != nil && s.session.CurrentBlockID() == 0x01 {
			s.finishNewGameEntry()
			return nil
		}
		if result.Exited && s.eclMenuReturnMode == ModeDungeon &&
			strings.TrimSpace(strings.Join(result.Text, "")) == "" {
			s.Mode = ModeDungeon
			s.syncCurrentECLDungeonArea()
			if s.pendingDungeonEntry {
				s.syncDungeonStateFromECLRegisters()
				s.pendingDungeonEntry = false
			}
			s.eclMenuReturnMode = ModeTitle
			s.Message = ""
			if s.session != nil && s.session.CurrentBlockID() == 0x31 {
				s.Prompt = s.catalog.Text("hap_dungeon_prompt", "哈普村　↑：前進　K／M：轉向　S：搜索　E：紮營")
			} else {
				s.Prompt = ""
			}
			s.Choices = nil
			s.currentOriginalChoices = nil
			return nil
		}
		if len(result.Text) > 0 {
			s.OriginalEvent = result.Text[len(result.Text)-1]
		}
	}
	if s.Location != LocationWilderness && originalChoice == "WILDERNESS" &&
		!(s.session != nil && s.session.CurrentBlockID() == 0x33) {
		s.enterMap()
		return nil
	}
	if s.Location != LocationWilderness && originalChoice == "EXIT" {
		s.leaveLocation()
	}
	return nil
}

func isWorldTravelRouteChoice(choice string) bool {
	switch choice {
	case "TRAIL", "ROAD", "RIVER", "WILDERNESS":
		return true
	default:
		return false
	}
}

func (s *State) continueAfterSuppressedPicture(result ecl.RunResult) (bool, error) {
	if result.WaitingForString && len(result.StringInputRequests) > 0 {
		s.beginECLStringInput(result.StringInputRequests[len(result.StringInputRequests)-1])
		return true, nil
	}
	if result.ShopRequested {
		return true, s.enterECLShop(result)
	}
	if result.TempleRequested {
		return true, s.enterECLTemple()
	}
	if result.CombatRequested {
		records := s.monsterRecordsForCurrentECL()
		if len(result.MonsterSpawns) > 0 && len(s.party) > 0 && len(records) > 0 {
			return true, s.StartEncounterWithAffects(result, records, s.monsterAffectsForCurrentECL(), s.party, s.combatSeed)
		}
		s.Mode = ModeEvent
		s.OriginalEvent = "COMBAT"
		return true, nil
	}
	if result.WaitingForMenu && len(result.Menus) > 0 {
		s.enterECLMenu(result.Menus[len(result.Menus)-1])
		return true, nil
	}
	return false, nil
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

func (s *State) enterTreasureMenu() {
	s.enterTreasureMenuFor(ModeEvent)
}

func (s *State) enterTreasureMenuFor(returnMode Mode) {
	eventMessage := s.pendingTreasureMessage
	s.pendingTreasureMessage = ""
	s.treasureMenu = true
	s.treasureTakeMenu = false
	s.treasureReturnMode = returnMode
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("treasure_prompt", "選擇要收下的財寶")
	s.Choices = make([]string, 0, len(s.pendingTreasureItems)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.pendingTreasureItems)+1)
	for index, item := range s.pendingTreasureItems {
		s.Choices = append(s.Choices, monster.LocalizedItemName(item, s.catalog))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "TREASURE_ITEM_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("treasure_exit", "暫不收下／繼續"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "TREASURE_EXIT")
	s.Message = s.catalog.Text("treasure_ready", "發現財寶。")
	if strings.TrimSpace(eventMessage) != "" {
		s.Message = eventMessage
	}
}

func (s *State) enterTreasureTakeMenu() {
	s.treasureTakeMenu = true
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("treasure_take_prompt", "選擇由哪位角色收下")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, character.Name)
		s.currentOriginalChoices = append(s.currentOriginalChoices, "TREASURE_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("treasure_cancel", "返回財寶列表"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "TREASURE_CANCEL")
}

func (s *State) selectTreasure(index int, originalChoice string) error {
	if s.treasureTakeMenu {
		if originalChoice == "TREASURE_CANCEL" {
			s.enterTreasureMenuFor(s.treasureReturnMode)
			return nil
		}
		if !strings.HasPrefix(originalChoice, "TREASURE_CHARACTER_") {
			return fmt.Errorf("invalid treasure character command %q", originalChoice)
		}
		characterIndex, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "TREASURE_CHARACTER_"))
		if err != nil {
			return fmt.Errorf("invalid treasure character command %q", originalChoice)
		}
		if err := s.TakeTreasureItem(characterIndex, s.treasureItemIndex); err != nil {
			return err
		}
		if len(s.pendingTreasureItems) > 0 {
			s.enterTreasureMenuFor(s.treasureReturnMode)
			return nil
		}
		return s.leaveTreasureMenu(s.catalog.Text("treasure_taken", "財寶已加入隊伍裝備。"))
	}
	if originalChoice == "TREASURE_EXIT" {
		s.pendingTreasureItems = nil
		return s.leaveTreasureMenu(s.catalog.Text("treasure_skipped", "隊伍繼續前進，未收下剩餘財寶。"))
	}
	if !strings.HasPrefix(originalChoice, "TREASURE_ITEM_") {
		return fmt.Errorf("invalid treasure item command %q", originalChoice)
	}
	itemIndex, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "TREASURE_ITEM_"))
	if err != nil || itemIndex < 0 || itemIndex >= len(s.pendingTreasureItems) {
		return fmt.Errorf("invalid treasure item index %q", originalChoice)
	}
	s.treasureItemIndex = itemIndex
	s.enterTreasureTakeMenu()
	return nil
}

func (s *State) leaveTreasureMenu(message string) error {
	s.treasureMenu = false
	s.treasureTakeMenu = false
	s.Mode = s.treasureReturnMode
	s.OriginalEvent = "TREASURE"
	s.Message = message
	if !s.treasureResumeECL {
		return nil
	}
	s.treasureResumeECL = false
	continued, err := s.continueECLAfterEngineBoundary()
	if err != nil {
		return err
	}
	if !continued {
		s.Mode = s.treasureReturnMode
	}
	return nil
}

func (s *State) selectParlay(index int, originalChoice string) error {
	if index < 0 || index >= len(s.currentOriginalChoices) || originalChoice == "" {
		return fmt.Errorf("parlay choice %d is invalid", index)
	}
	s.parlayMenu = false
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.OriginalEvent = "PARLAY"
	tactic := s.localizeOption(originalChoice)
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

// applyECLLoadCharacterSignals resolves the reference's 1-based player index
// against the persistent roster. The high bit is retained for the later
// redraw/free-current-player path, but does not invent that renderer side
// effect at this State boundary.
func (s *State) applyECLLoadCharacterSignals(result ecl.RunResult) {
	if len(result.LoadCharacterRequests) == 0 {
		return
	}
	request := result.LoadCharacterRequests[len(result.LoadCharacterRequests)-1]
	s.loadCharacterHighBit = request.HighBitSet
	index := int(request.PlayerIndex)
	if index < 0 || index >= len(s.partyRoster) {
		s.loadCharacterNotFound = true
		return
	}
	s.loadCharacterNotFound = false
	s.whoSelectedIndex = index
}

// applyECLDumpSignals applies reference DUMP removal to the persistent roster
// and its combat projection. Unlike player-facing ALTER DROP, an ECL script
// may remove the last member; the VM request already carries the reference
// predecessor-selection result for each ordered removal.
func (s *State) applyECLDumpSignals(result ecl.RunResult) error {
	for _, request := range result.DumpRequests {
		if !request.Resolved {
			continue
		}
		index := request.SelectedPlayerIndex
		if index < 0 || index >= len(s.partyRoster) {
			return fmt.Errorf("DUMP selected character %d is outside party roster", index)
		}
		id := s.partyRoster[index].ID
		s.partyRoster = append(s.partyRoster[:index], s.partyRoster[index+1:]...)
		for fighterIndex := range s.party {
			if s.party[fighterIndex].ID == id {
				s.party = append(s.party[:fighterIndex], s.party[fighterIndex+1:]...)
				break
			}
		}
		if request.NextSelectedPlayerSet {
			if request.NextSelectedPlayerIndex < 0 || request.NextSelectedPlayerIndex >= len(s.partyRoster) {
				return fmt.Errorf("DUMP next selected character %d is outside party roster", request.NextSelectedPlayerIndex)
			}
			s.whoSelectedIndex = request.NextSelectedPlayerIndex
		} else {
			s.whoSelectedIndex = -1
		}
	}
	return nil
}

// applyECLClockSignals bridges the reference ECL CLOCK command to the shared
// game-time adapter. Invalid slots are returned to the caller instead of
// silently mutating the clock with an invented interpretation.
func (s *State) applyECLClockSignals(result ecl.RunResult) error {
	for _, request := range result.ClockRequests {
		if err := s.AdvanceGameTime(int(request.TimeSlot), request.TimeStep); err != nil {
			return err
		}
	}
	return nil
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

// applyECLTreasureSignals keeps the raw loot request available to the item
// and money adapters. It intentionally does not load ITEM*.DAX or generate
// random treasure here: those operations need the active area and item data.
func (s *State) applyECLTreasureSignals(result ecl.RunResult) {
	s.pendingTreasure = append(s.pendingTreasure, result.TreasureRequests...)
}

func (s *State) enterECLShop(result ecl.RunResult) error {
	if !result.ShopRequested {
		return fmt.Errorf("ECL shop service was not requested")
	}
	if err := s.ResolveTreasureRequests(); err != nil {
		return err
	}
	offers := make([]ShopOffer, 0, len(s.pendingTreasureItems))
	for _, item := range s.pendingTreasureItems {
		price := int(item.Value)
		if price <= 0 {
			price = 1
		}
		switch result.ShopPriceScale {
		case 0x01:
			price >>= 4
		case 0x02:
			price >>= 3
		case 0x04:
			price >>= 2
		case 0x08:
			price >>= 1
		case 0x20:
			price <<= 1
		case 0x40:
			price <<= 2
		case 0x80:
			price <<= 3
		}
		if price < 1 {
			price = 1
		}
		if price > 0xFFFF {
			price = 0xFFFF
		}
		offers = append(offers, ShopOffer{Item: item, Price: uint16(price)})
	}
	s.pendingTreasureItems = nil
	s.SetShopOffers(offers)
	s.moneyPool = 0
	s.shopECLService = true
	s.eclMenuReturnMode = ModeDungeon
	s.eventReturnMode = ModeDungeon
	s.enterShopMenu()
	return nil
}

// applyECLRobSignals mirrors CMD_Rob: money is scaled by the retained
// percentage, then inventory entries are tested in order. Five typed coin
// fields mirror the DOS MoneySet; gems and jewelry are deliberately excluded.
func (s *State) applyECLRobSignals(result ecl.RunResult) {
	for requestIndex, request := range result.RobRequests {
		targets := make([]int, 0, len(s.partyRoster))
		if request.AllParty {
			for index := range s.partyRoster {
				targets = append(targets, index)
			}
		} else if request.SelectedPlayerSet &&
			request.SelectedPlayerIndex >= 0 && request.SelectedPlayerIndex < len(s.partyRoster) {
			targets = append(targets, request.SelectedPlayerIndex)
		}
		retained := 100
		if request.LossPercent >= 100 {
			retained = 0
		} else {
			retained -= int(request.LossPercent)
		}
		rng := rand.New(rand.NewSource(s.eclSeed + int64(requestIndex)))
		for _, characterIndex := range targets {
			character := &s.partyRoster[characterIndex]
			scaleCoin := func(value uint16) uint16 {
				return uint16((uint32(value) * uint32(retained)) / 100)
			}
			character.Copper = scaleCoin(character.Copper)
			character.Silver = scaleCoin(character.Silver)
			character.Electrum = scaleCoin(character.Electrum)
			character.Gold = scaleCoin(character.Gold)
			character.Platinum = scaleCoin(character.Platinum)
			chance := int(request.ItemChance)
			kept := character.Equipment[:0]
			for _, item := range character.Equipment {
				if item.Weight > 255 {
					if chance > 90 {
						chance -= 90
					} else {
						chance = 0
					}
				} else if item.Weight > 24 {
					if chance > 50 {
						chance -= 50
					} else {
						chance = 0
					}
				}
				if rng.Intn(100)+1 > chance {
					kept = append(kept, item)
				}
			}
			character.Equipment = kept
			if characterIndex < len(s.party) {
				if fighter, err := s.fighterForCharacter(*character); err == nil {
					s.party[characterIndex] = fighter
				}
			}
		}
	}
}

// ConsumeTreasureRequests transfers pending ECL TREASURE requests exactly once.
func (s *State) ConsumeTreasureRequests() []ecl.TreasureRequest {
	requests := append([]ecl.TreasureRequest(nil), s.pendingTreasure...)
	s.pendingTreasure = nil
	return requests
}

// SetTreasureItemBlocks installs decoded ITEM{area}.DAX blocks. The map key is
// normally (area << 8) | raw TREASURE item-block operand; a raw block key is
// also accepted for focused tests and callers with one active area.
func (s *State) SetTreasureItemBlocks(blocks map[uint16][]monster.ItemRecord) {
	s.treasureItemBlocks = make(map[uint16][]monster.ItemRecord, len(blocks))
	for block, items := range blocks {
		s.treasureItemBlocks[block] = append([]monster.ItemRecord(nil), items...)
	}
}

// PendingTreasureItems returns loot made available by resolved TREASURE
// requests. Items are not silently assigned to the first character; the UI or
// caller must explicitly choose a recipient with TakeTreasureItem.
func (s *State) PendingTreasureItems() []monster.ItemRecord {
	return append([]monster.ItemRecord(nil), s.pendingTreasureItems...)
}

// TreasurePool returns the non-coin pooled treasure that the remake state
// keeps separately from its gold-only shop pool.
func (s *State) TreasurePool() (gems, jewelry uint32) {
	return s.treasureGems, s.treasureJewelry
}

// ResolveTreasureRequests applies deterministic TREASURE effects and queues
// the item records for explicit pickup. It supports area item blocks and the
// 0xFF no-item branch; >=0x80 random generation remains an explicit boundary.
func (s *State) ResolveTreasureRequests() error {
	if len(s.pendingTreasure) == 0 {
		return nil
	}
	type resolved struct {
		copper        uint64
		gems, jewelry uint32
		items         []monster.ItemRecord
	}
	var total resolved
	rng := rand.New(rand.NewSource(s.eclSeed))
	for _, request := range s.pendingTreasure {
		total.copper += uint64(request.Coins[0]) +
			uint64(request.Coins[1])*10 +
			uint64(request.Coins[2])*100 +
			uint64(request.Coins[3])*200 +
			uint64(request.Coins[4])*1000
		total.gems += uint32(request.Coins[5])
		total.jewelry += uint32(request.Coins[6])
		if request.ItemBlock == 0xFF {
			continue
		}
		if request.ItemBlock >= 0x80 {
			if request.ItemBlock == 0xFF {
				continue
			}
			for count := 0; count < int(request.ItemBlock-0x80); count++ {
				total.items = append(total.items, monster.ItemRecord{Type: randomTreasureItemType(rng), Count: 1})
			}
			continue
		}
		key := uint16(s.Area.GameArea)<<8 | request.ItemBlock
		items, ok := s.treasureItemBlocks[key]
		if !ok {
			items, ok = s.treasureItemBlocks[request.ItemBlock]
		}
		if !ok {
			return fmt.Errorf("TREASURE item block 0x%02X for area %d is not loaded", request.ItemBlock, s.Area.GameArea)
		}
		total.items = append(total.items, items...)
	}
	total.copper += uint64(s.moneyCopperRemainder)
	s.moneyPool += uint32(total.copper / 200)
	s.moneyCopperRemainder = uint16(total.copper % 200)
	s.treasureGems += total.gems
	s.treasureJewelry += total.jewelry
	s.pendingTreasureItems = append(s.pendingTreasureItems, total.items...)
	s.pendingTreasure = nil
	return nil
}

// randomTreasureItemType mirrors the reference CMD_Treasure d100 table. The
// caller owns the seeded stream; this helper only returns the raw item type.
func randomTreasureItemType(rng *rand.Rand) uint8 {
	roll := rng.Intn(100) + 1
	if roll <= 60 {
		itemRoll := rng.Intn(100) + 1
		if itemRoll <= 47 || (itemRoll >= 50 && itemRoll <= 59) {
			if itemRoll == 45 {
				return 59
			}
			return uint8(itemRoll)
		}
		if itemRoll <= 90 {
			swordRoll := rng.Intn(10) + 1
			switch {
			case swordRoll <= 4:
				return 36
			case swordRoll <= 7:
				return 35
			case swordRoll == 8:
				return 34
			case swordRoll == 9:
				return 37
			default:
				return 38
			}
		}
		if itemRoll <= 94 {
			return 73
		}
		if itemRoll <= 97 {
			return 93
		}
		return 77
	}
	if roll <= 0x55 {
		return 61
	}
	if roll <= 0x5C {
		return 62
	}
	if roll <= 0x62 {
		potionRoll := rng.Intn(15) + 1
		if potionRoll <= 9 {
			return 71
		}
		if potionRoll == 10 {
			return 84
		}
		return 79
	}
	return 59
}

// TakeTreasureItem assigns one queued loot record to a selected character.
func (s *State) TakeTreasureItem(characterIndex, itemIndex int) error {
	if characterIndex < 0 || characterIndex >= len(s.partyRoster) {
		return fmt.Errorf("character index %d is out of range", characterIndex)
	}
	if itemIndex < 0 || itemIndex >= len(s.pendingTreasureItems) {
		return fmt.Errorf("treasure item index %d is out of range", itemIndex)
	}
	item := s.pendingTreasureItems[itemIndex]
	if item.Count == 0 {
		item.Count = 1
	}
	s.partyRoster[characterIndex].Equipment = append(s.partyRoster[characterIndex].Equipment, item)
	s.pendingTreasureItems = append(s.pendingTreasureItems[:itemIndex], s.pendingTreasureItems[itemIndex+1:]...)
	if characterIndex < len(s.party) {
		fighter, err := s.fighterForCharacter(s.partyRoster[characterIndex])
		if err != nil {
			return err
		}
		s.party[characterIndex] = fighter
	}
	return nil
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

// resolveAutomaticWholePartyECLDamage commits environmental DAMAGE packets
// that explicitly target the whole party (flags 0x80|0x40). The 0x20 bit
// bypasses saving throws; packets without it still resolve each member's
// encoded save type. Other DAMAGE forms remain pending because they need a
// selected character or the reference CanHitTarget adapter.
func (s *State) resolveAutomaticWholePartyECLDamage() ([]party.DamageOutcome, error) {
	if len(s.pendingDamageRequests) == 0 {
		return nil, nil
	}
	automatic := make([]ecl.DamageRequest, 0, len(s.pendingDamageRequests))
	remaining := make([]ecl.DamageRequest, 0, len(s.pendingDamageRequests))
	for _, request := range s.pendingDamageRequests {
		if request.Flags&0xC0 == 0xC0 {
			automatic = append(automatic, request)
		} else {
			remaining = append(remaining, request)
		}
	}
	if len(automatic) == 0 {
		return nil, nil
	}
	rng := rand.New(rand.NewSource(s.eclSeed))
	roll := func(sides int) int {
		if sides <= 0 {
			return 0
		}
		return rng.Intn(sides) + 1
	}
	original := s.pendingDamageRequests
	s.pendingDamageRequests = automatic
	outcomes, err := s.ResolvePendingECLDamage(-1, roll, roll)
	if err != nil {
		s.pendingDamageRequests = original
		return nil, err
	}
	s.pendingDamageRequests = remaining
	return outcomes, nil
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
	if s.campReturnMode == ModeTitle {
		s.campReturnMode = ModeWilderness
	}
	s.enterCampMenu()
	return nil
}

// EnterDungeonCamp mirrors TryEncamp: run PreCampCheck before opening CAMP.
// The script writes encounter period/percentage to 0x7ED2/0x7ED3.
func (s *State) EnterDungeonCamp() error {
	if s.Mode != ModeDungeon {
		return fmt.Errorf("dungeon camp is invalid in mode %d", s.Mode)
	}
	if s.session == nil {
		return fmt.Errorf("dungeon camp requires an ECL session")
	}
	s.syncDungeonECLRegisters()
	blockBefore := s.session.CurrentBlockID()
	result, err := s.session.RunEntrySeedWithPartyContext(
		2, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
	s.requestMusicIfBlockChanged(blockBefore)
	if period, ok := s.session.MemoryValue(0x7ED2); ok {
		s.restEncounterPeriod = period
	}
	if percent, ok := s.session.MemoryValue(0x7ED3); ok {
		s.restEncounterPercent = percent
	}
	if handled, err := s.applyDungeonLifecycleResult(result); handled || err != nil {
		return err
	}
	s.CampCount++
	s.campReturnMode = ModeDungeon
	s.enterCampMenu()
	return nil
}

// RunCampInterrupted invokes the fourth vm_init_ecl entry after a rest
// adapter has actually reported an interruption.
func (s *State) RunCampInterrupted() error {
	if s.session == nil {
		return fmt.Errorf("camp interruption requires an ECL session")
	}
	blockBefore := s.session.CurrentBlockID()
	result, err := s.session.RunEntrySeedWithPartyContext(
		3, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
	s.requestMusicIfBlockChanged(blockBefore)
	_, err = s.applyDungeonLifecycleResult(result)
	return err
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
		if !character.HasClass(party.ClassCleric) {
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
	s.campMagicCastMenu = false
	s.campMagicCastChar = -1
	s.campMagicCastSpell = 0
	s.alterMenu = false
	s.alterOrderMenu = false
	s.alterOrderSelected = -1
	s.alterDropMenu = false
	s.alterDropConfirm = false
	s.alterDropSelected = -1
	s.alterRenameMenu = false
	s.alterRenameChar = -1
	s.renameEditing = false
	s.renameCharacter = -1
	s.renameName = ""
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
			completedHours, interrupted := s.restInterruption(s.restHours)
			if err := s.AdvanceGameTimeHours(completedHours); err != nil {
				return err
			}
			healed := s.restPartyHours(completedHours)
			s.campRestMenu = false
			if interrupted {
				s.campMenu = false
				s.Mode = ModeDungeon
				s.eventReturnMode = ModeDungeon
				if err := s.RunCampInterrupted(); err != nil {
					return err
				}
				prefix := s.catalog.Text("camp_rest_interrupted", "你們的休息突然中斷！")
				if s.Message == "" {
					s.Mode = ModeEvent
					s.Message = prefix
				} else {
					s.Message = prefix + "\n" + s.Message
				}
				s.OriginalEvent = "CAMP INTERRUPTED"
				return nil
			}
			memorized := s.applyPendingMemorization()
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
	if s.alterRenameMenu {
		if originalChoice == "ALTER_RENAME_EXIT" {
			s.alterRenameMenu = false
			s.alterRenameChar = -1
			s.enterAlterMenu()
			return nil
		}
		if strings.HasPrefix(originalChoice, "ALTER_RENAME_CHARACTER_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "ALTER_RENAME_CHARACTER_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid alter rename character %q", originalChoice)
			}
			s.alterRenameChar = value
			return s.BeginRenameCharacter(value)
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
		if originalChoice == "ALTER_RENAME" {
			s.enterAlterRenameMenu()
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
		case "CAMP_MAGIC_CAST":
			s.enterCampMagicCastCharacterMenu()
			return nil
		case "CAMP_MAGIC_MEMORIZE":
			s.enterCampMagicMemorizeCharacterMenu()
			return nil
		case "CAMP_MAGIC_REST":
			s.enterCampRestMenu()
			return nil
		case "CAMP_MAGIC_SCRIBE":
			s.Mode = ModeEvent
			s.eventReturnMode = ModeWilderness
			s.OriginalEvent = "MAGIC"
			s.Message = s.catalog.Text("camp_magic_pending", "此法術功能已進入資料邊界，完整規則仍待接入。")
			return nil
		}
	}
	if s.campMagicCastMenu {
		if originalChoice == "CAMP_MAGIC_CAST_EXIT" {
			s.enterCampMagicMenu()
			return nil
		}
		if s.campMagicCastChar < 0 {
			if strings.HasPrefix(originalChoice, "CAMP_MAGIC_CAST_CHAR_") {
				value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_MAGIC_CAST_CHAR_"))
				if err != nil || value < 0 || value >= len(s.partyRoster) {
					return fmt.Errorf("invalid cast character %q", originalChoice)
				}
				s.enterCampMagicCastSpellMenu(value)
				return nil
			}
			return fmt.Errorf("unknown cast character choice %q", originalChoice)
		}
		if strings.HasPrefix(originalChoice, "CAMP_MAGIC_CAST_SPELL_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_MAGIC_CAST_SPELL_"))
			character := s.partyRoster[s.campMagicCastChar]
			if err != nil || value < 0 || value >= len(character.SpellSlots) {
				return fmt.Errorf("invalid cast spell %q", originalChoice)
			}
			s.campMagicCastSpell = character.SpellSlots[value]
			if s.campMagicCastSpell != CureLightWoundsSpellID {
				s.campMagicCastMenu = false
				s.campMagicMenu = true
				s.Mode = ModeEvent
				s.eventReturnMode = ModeWilderness
				s.OriginalEvent = "MAGIC CAST"
				s.Message = fmt.Sprintf(s.catalog.Text("camp_magic_cast_unknown", "目前只能在紮營施放 Cure Light Wounds；法術 0x%02X 尚待接入。"), s.campMagicCastSpell)
				return nil
			}
			s.enterCampMagicCastTargetMenu()
			return nil
		}
		if strings.HasPrefix(originalChoice, "CAMP_MAGIC_CAST_TARGET_") {
			value, err := strconv.Atoi(strings.TrimPrefix(originalChoice, "CAMP_MAGIC_CAST_TARGET_"))
			if err != nil || value < 0 || value >= len(s.partyRoster) {
				return fmt.Errorf("invalid cast target %q", originalChoice)
			}
			return s.castCampCureLightWounds(value)
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
				equipment = append(equipment, monster.LocalizedItemName(item, s.catalog))
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
		if s.campECLService {
			s.campECLService = false
			continued, err := s.continueECLAfterEngineBoundary()
			if err != nil {
				return err
			}
			if continued {
				return nil
			}
		}
		if s.campReturnMode == ModeDungeon {
			s.Mode = ModeDungeon
			s.Prompt = s.catalog.Text("dungeon_prompt", "↑：前進　K／M：轉向　S：搜索　E：紮營")
			s.Choices = nil
			s.currentOriginalChoices = nil
			s.Message = ""
			s.campReturnMode = ModeTitle
			return nil
		}
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
	s.alterRenameMenu = false
	s.alterRenameChar = -1
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
		s.catalog.Text("alter_rename", "改名"),
	}
	s.currentOriginalChoices = []string{"ALTER_ORDER", "ALTER_DROP", "ALTER_SPEED", "ALTER_ICON", "ALTER_PICS", "ALTER_EXIT", "ALTER_RENAME"}
	s.Message = ""
}

func (s *State) enterAlterRenameMenu() {
	s.campMenu = true
	s.alterMenu = false
	s.alterRenameMenu = true
	s.alterRenameChar = -1
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("alter_rename_prompt", "選擇要改名的角色")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf("%s（%s）", character.Name, character.ID))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_RENAME_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("alter_rename_exit", "返回修改選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "ALTER_RENAME_EXIT")
	s.Message = ""
}

// RenameEditing reports whether the Ebiten input adapter should route text
// input to the ALTER rename editor instead of the normal menu.
func (s *State) RenameEditing() bool { return s.renameEditing }

func (s *State) RenameText() string { return s.renameName }

func (s *State) BeginRenameCharacter(index int) error {
	if index < 0 || index >= len(s.partyRoster) {
		return fmt.Errorf("rename character index %d is out of range", index)
	}
	s.renameEditing = true
	s.renameCharacter = index
	s.renameName = s.partyRoster[index].Name
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf(s.catalog.Text("alter_rename_edit_prompt", "輸入 %s 的新名稱"), s.partyRoster[index].Name)
	s.Message = ""
	return nil
}

func (s *State) AppendRenameName(chars []rune) error {
	if !s.renameEditing {
		return fmt.Errorf("rename editor is not active")
	}
	if len([]byte(s.renameName+string(chars))) > 15 {
		return fmt.Errorf("DOS 角色名稱最多 15 bytes")
	}
	s.renameName += string(chars)
	return nil
}

func (s *State) BackspaceRenameName() error {
	if !s.renameEditing {
		return fmt.Errorf("rename editor is not active")
	}
	name := []rune(s.renameName)
	if len(name) > 0 {
		s.renameName = string(name[:len(name)-1])
	}
	return nil
}

func (s *State) CancelRename() error {
	if !s.renameEditing {
		return fmt.Errorf("rename editor is not active")
	}
	s.renameEditing = false
	s.renameCharacter = -1
	s.renameName = ""
	s.enterAlterRenameMenu()
	return nil
}

func (s *State) CommitRename() error {
	if !s.renameEditing || s.renameCharacter < 0 || s.renameCharacter >= len(s.partyRoster) {
		return fmt.Errorf("rename editor is not active")
	}
	if s.renameName == "" {
		return fmt.Errorf("character name cannot be empty")
	}
	index := s.renameCharacter
	oldName := s.partyRoster[index].Name
	s.partyRoster[index].Name = s.renameName
	id := s.partyRoster[index].ID
	for fighterIndex := range s.party {
		if s.party[fighterIndex].ID == id {
			s.party[fighterIndex].Name = s.renameName
		}
	}
	s.renameEditing = false
	s.renameCharacter = -1
	s.renameName = ""
	s.alterRenameMenu = false
	s.alterRenameChar = -1
	s.alterMenu = true
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.OriginalEvent = "ALTER RENAME"
	s.Message = fmt.Sprintf(s.catalog.Text("alter_rename_done", "%s 已改名為 %s。"), oldName, s.partyRoster[index].Name)
	return nil
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
		return s.localizeOption(originalChoice)
	}
}

// ConsumeSaveRequest transfers a CAMP SAVE intent to the platform adapter.
// The state layer never chooses a filesystem path or performs file I/O.
func (s *State) ConsumeSaveRequest() bool {
	requested := s.saveRequested
	s.saveRequested = false
	return requested
}

// GameWon and PartyKilled expose terminal PROGRAM side effects without
// coupling frontends to the ECL VM's numeric routine IDs.
func (s *State) GameWon() bool {
	return s.gameWon
}

func (s *State) PartyKilled() bool {
	return s.partyKilled
}

// applyECLProgram translates the external routines observed in CMD_Program
// into frontend-neutral State transactions. PROGRAM 8 asks before saving in
// the reference; the remake keeps that choice but returns to its title screen
// instead of terminating the host process.
func (s *State) applyECLProgram(result ecl.RunResult) (bool, error) {
	if !result.ProgramExit || len(result.ProgramIDs) == 0 {
		return false, nil
	}
	id := result.ProgramIDs[len(result.ProgramIDs)-1]
	switch id {
	case 0:
		if s.eventReturnMode == ModeDungeon && s.DungeonWallRoof == 0x8C {
			s.enterTrainingMenu()
			return true, nil
		}
		s.enterProgramTitle("主選單")
		return true, nil
	case 3:
		s.partyKilled = true
		s.programEndMenu = true
		s.Mode = ModeWilderness
		s.OriginalEvent = "PROGRAM 3"
		s.Prompt = "隊伍全滅"
		s.Choices = []string{"返回標題"}
		s.currentOriginalChoices = []string{"PROGRAM_END"}
		s.Message = "隊伍已全滅。"
		return true, nil
	case 8:
		s.gameWon = true
		for index := range s.partyRoster {
			s.partyRoster[index].HitPoints = s.partyRoster[index].MaxHitPoints
			s.partyRoster[index].HealthStatus = party.HealthStatusOK
			s.partyRoster[index].Bleeding = 0
		}
		for index := range s.party {
			s.party[index].HitPoints = s.party[index].MaxHitPoints
			s.party[index].DeathOverlay = false
			s.party[index].DownedCorpse = false
		}
		s.programEndMenu = true
		s.Mode = ModeWilderness
		s.OriginalEvent = "PROGRAM 8"
		s.Prompt = "你們解除了青色枷鎖的詛咒！"
		s.Choices = []string{"保存勝利進度並結束", "不保存並結束"}
		s.currentOriginalChoices = []string{"PROGRAM_WIN_SAVE", "PROGRAM_END"}
		s.Message = "全隊已恢復，是否在結束前保存？"
		return true, nil
	case 9:
		s.campECLService = s.session != nil && len(s.eclBlock) > 0
		return true, s.Camp()
	default:
		return false, nil
	}
}

func (s *State) selectProgramEnd(originalChoice string) error {
	switch originalChoice {
	case "PROGRAM_WIN_SAVE":
		if len(s.partyRoster) == 0 {
			return fmt.Errorf("cannot save victory without a party roster")
		}
		s.saveRequested = true
		s.enterProgramTitle("勝利進度已要求保存。")
		return nil
	case "PROGRAM_END":
		s.enterProgramTitle("冒險告一段落。")
		return nil
	default:
		return fmt.Errorf("invalid PROGRAM end choice %q", originalChoice)
	}
}

func isTrainingProgramChoice(originalChoice string, result ecl.RunResult) bool {
	return originalChoice == "HALL" && result.ProgramExit &&
		len(result.ProgramIDs) > 0 && result.ProgramIDs[len(result.ProgramIDs)-1] == 0
}

func (s *State) enterProgramTitle(message string) {
	s.programEndMenu = false
	s.Mode = ModeTitle
	s.OriginalEvent = "PROGRAM 0"
	s.Prompt = s.Title
	s.Choices = nil
	s.currentOriginalChoices = nil
	s.Message = message
	s.eclBlock = nil
	s.session = nil
}

// applyECLCallSignals translates the three raw CALL operands observed in the
// CoAB ECL image. CALL 0xC01E maps to reference MovePositionForward and is a
// forced wrapped move (the routine itself does not perform collision checks).
// The renderer still consumes the ordered calls so redraw can use its loaded
// GEO/piece assets.
func (s *State) applyECLCallSignals(result ecl.RunResult) {
	for index, address := range result.CallAddresses {
		s.pendingECLCalls = append(s.pendingECLCalls, address)
		switch address {
		case 0x2E10:
			// The reference redraw routine consumes the current dungeon
			// registers. Some same-block ECL events assign a new position
			// immediately before this CALL, so project those registers before
			// the renderer rebuilds the view.
			if s.session != nil && s.Area.InDungeon &&
				index < len(result.CallRequests) &&
				result.SessionBlockRangeSet &&
				result.SessionStartBlockID == result.CallRequests[index].BlockID &&
				result.SessionEndBlockID == result.CallRequests[index].BlockID &&
				result.CallRequests[index].BlockID == s.session.CurrentBlockID() {
				s.projectFreshDungeonCoordinatesBeforeCall(
					result, result.CallRequests[index],
				)
			}
		case 0xC01E:
			switch s.DungeonDirection {
			case 0:
				s.DungeonY = (s.DungeonY + 15) % 16
			case 2:
				s.DungeonX = (s.DungeonX + 1) % 16
			case 4:
				s.DungeonY = (s.DungeonY + 1) % 16
			case 6:
				s.DungeonX = (s.DungeonX + 15) % 16
			}
		case 0xB200:
			// Reference word_1EE76 selects sound_a for the default/8 branch
			// and sound_b for 10. That transient is not yet projected into
			// State, so preserve the verified default sound_a behavior.
			s.requestSound(SoundStep)
		}
	}
}

func (s *State) projectFreshDungeonCoordinatesBeforeCall(
	result ecl.RunResult,
	call ecl.CallRequest,
) {
	var mask uint8
	var x, y uint16
	var direction uint16
	for _, write := range result.SaveWrites {
		if write.BlockID != call.BlockID || write.PC >= call.PC {
			continue
		}
		switch write.Address {
		case 0xC04B:
			x = write.Value
			mask |= 1
		case 0xC04C:
			y = write.Value
			mask |= 2
		case 0xC04D:
			direction = write.Value
			mask |= 4
		}
	}
	// Verified party-position transactions always commit a fresh facing.
	// Some dialogue events write C04B/C04C as scratch coordinates before the
	// same redraw CALL without moving the party; do not project those values.
	if mask&4 == 0 {
		return
	}
	if mask&1 != 0 {
		s.DungeonX = int(int16(x))
	}
	if mask&2 != 0 {
		s.DungeonY = int(int16(y))
	}
	if mask&4 != 0 {
		s.DungeonDirection = uint8(direction&3) * 2
	}
	if mask != 0 {
		s.MapX, s.MapY = s.DungeonX, s.DungeonY
	}
}

// applyECLNPCSignals mirrors load_npc: resolve the current chapter's shared
// MON*CHA Player record plus SPC/ITM sidecars, insert it into the party,
// assign the lowest free icon slot, select it, then apply CMD_AddNPC morale.
func (s *State) applyECLNPCSignals(result ecl.RunResult) error {
	if len(result.NPCRequests) == 0 {
		return nil
	}
	chapter := uint8(1)
	if s.session != nil {
		chapter = monsterChapterForBlock(s.session.CurrentBlockID())
	}
	records := s.monsterRecordsForCurrentECL()
	affects := s.monsterAffectsForCurrentECL()
	items := s.monsterItemsForCurrentECL()
	for _, request := range result.NPCRequests {
		if resultUsesTemporaryMonsterAlly(result, uint8(request.ID)) {
			// Some scripts ADD NPC only to assign morale after loading the
			// same MON*CHA record into reserved combatant slot 8 and changing
			// its team. StartEncounter owns that encounter-scoped monster
			// fighter; it must not be parsed as a persistent player record.
			continue
		}
		if len(s.partyRoster) > 7 {
			continue
		}
		npcID := uint8(request.ID)
		record, ok := records[npcID]
		if !ok || len(record.Raw) < party.DOSPlayerRecordSize {
			return fmt.Errorf("ADD NPC 0x%02X has no MON%dCHA Player record", npcID, chapter)
		}
		id := fmt.Sprintf("npc-%d-%02x-%d", chapter, npcID, len(s.partyRoster))
		parsed, err := party.ParseDOSNPCRecord(record.Raw, id)
		if err != nil {
			return fmt.Errorf("ADD NPC 0x%02X: %w", npcID, err)
		}
		parsed.Inventory = append([]monster.ItemRecord(nil), items[npcID]...)
		parsed.Effects = append([]monster.AffectRecord(nil), affects[npcID]...)
		parsed.ControlMorale = uint8(request.Morale>>1) + 0x80
		character, err := parsed.Character()
		if err != nil {
			return fmt.Errorf("ADD NPC 0x%02X character: %w", npcID, err)
		}
		if chapter == 5 && npcID == 0x3B {
			character.Name = s.catalog.Text("npc_akabar", "阿卡巴・貝爾・阿卡什")
		}
		if chapter == 3 && npcID == 0x16 {
			character.Name = s.catalog.Text("npc_alias", "愛麗雅絲")
		}
		if chapter == 3 && npcID == 0x17 {
			character.Name = s.catalog.Text("npc_dragonbait", "龍餌")
		}
		usedIcons := [8]bool{}
		for _, member := range s.partyRoster {
			if member.IconID < 8 {
				usedIcons[member.IconID] = true
			}
		}
		character.IconID = 0
		for character.IconID < 8 && usedIcons[character.IconID] {
			character.IconID++
		}
		fighter, err := s.fighterForCharacter(character)
		if err != nil {
			return fmt.Errorf("ADD NPC 0x%02X fighter: %w", npcID, err)
		}
		s.partyRoster = append(s.partyRoster, character)
		s.party = append(s.party, fighter)
		s.whoSelectedIndex = len(s.partyRoster) - 1
	}
	return nil
}

func resultUsesTemporaryMonsterAlly(result ecl.RunResult, npcID uint8) bool {
	spawns := append([]ecl.MonsterSpawn(nil), result.MonsterSpawns...)
	ecl.ApplyCombatTeamWrites(spawns, result.CombatTeamWrites)
	for _, spawn := range spawns {
		if spawn.MonsterID == npcID && spawn.PartyMask != 0 {
			return true
		}
	}
	return false
}

// ConsumeECLCallRequests transfers ordered redraw/external-call intents to the
// frontend. State-owned effects (position and default sound) are already
// applied before this transaction is exposed.
func (s *State) ConsumeECLCallRequests() []uint16 {
	requests := append([]uint16(nil), s.pendingECLCalls...)
	s.pendingECLCalls = nil
	return requests
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
	s.campMagicCastMenu = false
	s.campMagicCastChar = -1
	s.campMagicCastSpell = 0
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

func (s *State) enterCampMagicCastCharacterMenu() {
	s.campMenu = true
	s.campMagicMenu = false
	s.campMagicViewMenu = false
	s.campMagicMemorizeMenu = false
	s.campMagicCastMenu = true
	s.campMagicCastChar = -1
	s.campMagicCastSpell = 0
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_magic_cast_character_prompt", "選擇施法者")
	s.Choices = nil
	s.currentOriginalChoices = nil
	for index, character := range s.partyRoster {
		if (!character.HasClass(party.ClassCleric) && !character.HasClass(party.ClassMagicUser)) || len(character.SpellSlots) == 0 {
			continue
		}
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("camp_magic_cast_character", "%s（已記憶 %d 個法術）"), character.Name, len(character.SpellSlots)))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_CAST_CHAR_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_magic_cast_exit", "返回法術選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_CAST_EXIT")
	s.Message = ""
}

func (s *State) enterCampMagicCastSpellMenu(characterIndex int) {
	character := s.partyRoster[characterIndex]
	s.campMagicCastChar = characterIndex
	s.Mode = ModeWilderness
	s.Prompt = fmt.Sprintf(s.catalog.Text("camp_magic_cast_spell_prompt", "%s 要施放哪個法術？"), character.Name)
	s.Choices = make([]string, 0, len(character.SpellSlots)+1)
	s.currentOriginalChoices = make([]string, 0, len(character.SpellSlots)+1)
	for index, spellID := range character.SpellSlots {
		s.Choices = append(s.Choices, campSpellLabel(s.catalog, character.Class, spellID))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_CAST_SPELL_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_magic_cast_exit", "返回法術選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_CAST_EXIT")
	s.Message = ""
}

func (s *State) enterCampMagicCastTargetMenu() {
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("camp_magic_cast_target_prompt", "選擇 Cure Light Wounds 目標")
	s.Choices = nil
	s.currentOriginalChoices = nil
	for index, character := range s.partyRoster {
		if character.HealthStatus == party.HealthStatusDead || character.HitPoints >= character.MaxHitPoints {
			continue
		}
		s.Choices = append(s.Choices, fmt.Sprintf("%s（HP %d/%d）", character.Name, character.HitPoints, character.MaxHitPoints))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_CAST_TARGET_"+strconv.Itoa(index))
	}
	if len(s.Choices) == 0 {
		s.campMagicCastMenu = false
		s.campMagicMenu = true
		s.Mode = ModeEvent
		s.eventReturnMode = ModeWilderness
		s.OriginalEvent = "MAGIC CAST"
		s.Message = s.catalog.Text("camp_magic_cast_no_target", "沒有需要治療的隊員，法術未消耗。")
		return
	}
	s.Choices = append(s.Choices, s.catalog.Text("camp_magic_cast_exit", "返回法術選單"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "CAMP_MAGIC_CAST_EXIT")
	s.Message = ""
}

func (s *State) castCampCureLightWounds(targetIndex int) error {
	if s.campMagicCastChar < 0 || s.campMagicCastChar >= len(s.partyRoster) {
		return fmt.Errorf("Cure Light Wounds caster is not selected")
	}
	caster := &s.partyRoster[s.campMagicCastChar]
	if !caster.HasClass(party.ClassCleric) && !caster.HasClass(party.ClassMagicUser) {
		return fmt.Errorf("character %q cannot cast Cure Light Wounds", caster.Name)
	}
	spellIndex := -1
	for index, spellID := range caster.SpellSlots {
		if spellID == CureLightWoundsSpellID {
			spellIndex = index
			break
		}
	}
	if spellIndex < 0 {
		return fmt.Errorf("caster %q has no Cure Light Wounds", caster.Name)
	}
	if targetIndex < 0 || targetIndex >= len(s.partyRoster) {
		return fmt.Errorf("Cure Light Wounds target %d is outside party", targetIndex)
	}
	target := &s.partyRoster[targetIndex]
	if target.HealthStatus == party.HealthStatusDead || target.HitPoints >= target.MaxHitPoints {
		return fmt.Errorf("Cure Light Wounds target %q is not wounded", target.Name)
	}
	casterName, targetName, targetID := caster.Name, target.Name, target.ID
	caster.SpellSlots = append(caster.SpellSlots[:spellIndex], caster.SpellSlots[spellIndex+1:]...)
	rng := rand.New(rand.NewSource(s.fixSeed))
	s.fixSeed++
	healed := rng.Intn(8) + 1
	before := target.HitPoints
	target.HitPoints += healed
	if target.HitPoints > target.MaxHitPoints {
		target.HitPoints = target.MaxHitPoints
	}
	actual := target.HitPoints - before
	for index := range s.party {
		if s.party[index].ID == targetID {
			s.party[index].HitPoints = target.HitPoints
		}
	}
	s.campMagicCastMenu = false
	s.campMagicMenu = true
	s.campMagicCastChar = -1
	s.campMagicCastSpell = 0
	s.Mode = ModeEvent
	s.eventReturnMode = ModeWilderness
	s.OriginalEvent = "MAGIC CAST"
	s.Message = fmt.Sprintf(s.catalog.Text("camp_magic_cast_done", "%s 對 %s 施放 Cure Light Wounds，恢復 %d HP。"), casterName, targetName, actual)
	return nil
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
		return s.localizeOption(originalChoice)
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

// SetPartyRoster installs typed characters through the same conversion used
// by save loading. Frontends and deterministic integration oracles can use
// this boundary without reaching into State's roster internals.
func (s *State) SetPartyRoster(roster party.Roster) error {
	if len(roster) == 0 {
		return fmt.Errorf("party roster cannot be empty")
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
	s.partyRoster = append(party.Roster(nil), roster...)
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
func (s *State) AdvancePartyEffects(ticks uint16) int {
	if ticks == 0 {
		return 0
	}
	removed := 0
	for index := range s.partyRoster {
		removed += s.partyRoster[index].AdvanceEffects(ticks)
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
	s.Prompt = s.catalog.Text("shop_menu_prompt", "shop_menu_prompt")
	s.Choices = []string{
		s.catalog.Text("shop_buy", "shop_buy"),
		s.catalog.Text("shop_view", "shop_view"),
		s.catalog.Text("shop_take", "shop_take"),
		s.catalog.Text("shop_pool", "shop_pool"),
		s.catalog.Text("shop_share", "shop_share"),
		s.catalog.Text("shop_appraise", "shop_appraise"),
		s.catalog.Text("shop_sell", "shop_sell"),
		s.catalog.Text("shop_identify", "shop_identify"),
		s.catalog.Text("shop_exit", "shop_exit"),
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
			s.Message = fmt.Sprintf(s.catalog.Text("shop_appraise_failed", "shop_appraise_failed: %s"), err)
		} else {
			s.Message = fmt.Sprintf(s.catalog.Text("shop_appraise_done", "shop_appraise_done"), offer)
		}
		return nil
	}
	if originalChoice == "SHOP_APPRAISE_REJECT" {
		s.shopAppraiseMenu = false
		s.shopAppraiseConfirm = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "APPRAISE"
		s.Message = s.catalog.Text("shop_appraise_rejected", "shop_appraise_rejected")
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
			s.Message = fmt.Sprintf(s.catalog.Text("shop_sell_failed", "shop_sell_failed: %s"), err)
			return nil
		}
		s.shopSellMenu = false
		s.shopSellItemMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "SELL"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_sale_done", "shop_sale_done"), monster.LocalizedItemName(item, s.catalog), item.Value)
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
			s.Message = fmt.Sprintf(s.catalog.Text("shop_identify_failed", "shop_identify_failed: %s"), identifyErr)
		} else {
			s.Message = fmt.Sprintf(s.catalog.Text("shop_identify_done", "shop_identify_done"), party.ShopIdentifyFee, monster.LocalizedItemName(item, s.catalog))
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
			s.Message = fmt.Sprintf(s.catalog.Text("shop_take_failed", "shop_take_failed: %s"), err)
			return nil
		}
		s.shopTakeMenu = false
		s.shopTakeAmountMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "TAKE"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_take_done", "shop_take_done"), value, s.partyRoster[s.shopTakeCharacter].Name)
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
			equipment = append(equipment, monster.LocalizedItemName(item, s.catalog))
		}
		s.shopViewMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "VIEW"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_view_summary", "shop_view_summary"), character.Name, character.HitPoints, character.MaxHitPoints, character.Gold, strings.Join(equipment, s.catalog.Text("list_separator", ", ")))
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
			s.Message = fmt.Sprintf(s.catalog.Text("shop_purchase_failed", "shop_purchase_failed: %s"), err)
			return nil
		}
		item := s.shopOffers[value].Item
		s.shopStockMenu = false
		s.Mode = ModeEvent
		s.eventReturnMode = ModePlace
		s.OriginalEvent = "BUY"
		s.Message = fmt.Sprintf(s.catalog.Text("shop_purchase_done", "shop_purchase_done"), monster.LocalizedItemName(item, s.catalog))
		return nil
	}
	if originalChoice == "SHOP_EXIT" {
		s.enterShopMenu()
		return nil
	}
	if originalChoice == "EXIT" {
		s.shopMenu = false
		if s.shopECLService {
			s.shopECLService = false
			continued, err := s.continueECLAfterEngineBoundary()
			if err != nil {
				return err
			}
			if continued {
				return nil
			}
			s.Mode = ModeDungeon
			return nil
		}
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
			return s.catalog.Text("shop_buy_unavailable", "shop_buy_unavailable")
		}
		s.enterShopStockMenu()
		return ""
	case "VIEW":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_view_unavailable", "shop_view_unavailable")
		}
		s.enterShopViewMenu()
		return ""
	case "TAKE":
		if len(s.partyRoster) == 0 || s.moneyPool == 0 {
			return s.catalog.Text("shop_take_unavailable", "shop_take_unavailable")
		}
		s.enterShopTakeMenu()
		return ""
	case "POOL":
		if err := s.PoolPartyGold(); err != nil {
			return fmt.Sprintf(s.catalog.Text("shop_pool_failed", "shop_pool_failed: %s"), err)
		}
		return fmt.Sprintf(s.catalog.Text("shop_pool_done", "shop_pool_done"), s.moneyPool)
	case "SHARE":
		before := s.moneyPool
		if err := s.ShareGold(); err != nil {
			return fmt.Sprintf(s.catalog.Text("shop_share_failed", "shop_share_failed: %s"), err)
		}
		return fmt.Sprintf(s.catalog.Text("shop_share_done", "shop_share_done"), before)
	case "APPRAISE":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_appraise_unavailable", "shop_appraise_unavailable")
		}
		s.enterShopAppraiseCharacterMenu()
		return ""
	case "SELL":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_sell_unavailable", "shop_sell_unavailable")
		}
		s.enterShopSellCharacterMenu()
		return ""
	case "ID":
		if len(s.partyRoster) == 0 {
			return s.catalog.Text("shop_identify_unavailable", "shop_identify_unavailable")
		}
		s.enterShopIdentifyCharacterMenu()
		return ""
	default:
		return s.localizeOption(originalChoice)
	}
}

func (s *State) enterShopStockMenu() {
	s.shopMenu = true
	s.shopStockMenu = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_stock_prompt", "shop_stock_prompt")
	s.Choices = make([]string, 0, len(s.shopOffers)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.shopOffers)+1)
	for index, offer := range s.shopOffers {
		name := monster.LocalizedItemName(offer.Item, s.catalog)
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_item_price", "shop_item_price"), name, offer.Price))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_OFFER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_exit", "shop_exit"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_EXIT")
	s.Message = ""
}

func (s *State) enterShopViewMenu() {
	s.shopMenu = true
	s.shopStockMenu = false
	s.shopViewMenu = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_view_prompt", "shop_view_prompt")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_view_character", "shop_view_character"), character.Name, character.HitPoints, character.MaxHitPoints, character.Gold))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_VIEW_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_view_exit", "shop_view_exit"))
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
	s.Prompt = s.catalog.Text("shop_take_prompt", "shop_take_prompt")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_take_character", "shop_take_character"), character.Name, character.Gold))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_TAKE_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_take_exit", "shop_take_exit"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_TAKE_EXIT")
	s.Message = ""
}

func (s *State) enterShopTakeAmountMenu() {
	s.shopTakeAmountMenu = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_take_amount_prompt", "shop_take_amount_prompt")
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
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_gold_amount", "shop_gold_amount"), amount))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_TAKE_AMOUNT_"+strconv.FormatUint(uint64(amount), 10))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_take_exit", "shop_take_exit"))
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
	s.Prompt = s.catalog.Text("shop_sell_prompt", "shop_sell_prompt")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_character_items", "shop_character_items"), character.Name, len(character.Equipment)))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_SELL_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_sell_exit", "shop_sell_exit"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_SELL_EXIT")
	s.Message = ""
}

func (s *State) enterShopSellItemMenu() {
	s.shopSellMenu = true
	s.shopSellItemMenu = true
	s.Mode = ModePlace
	character := s.partyRoster[s.shopSellCharacter]
	s.Prompt = fmt.Sprintf(s.catalog.Text("shop_sell_item_prompt", "shop_sell_item_prompt"), character.Name)
	s.Choices = make([]string, 0, len(character.Equipment)+1)
	s.currentOriginalChoices = make([]string, 0, len(character.Equipment)+1)
	for index, item := range character.Equipment {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_item_price", "shop_item_price"), monster.LocalizedItemName(item, s.catalog), item.Value))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_SELL_ITEM_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_sell_exit", "shop_sell_exit"))
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
	s.Prompt = s.catalog.Text("shop_identify_prompt", "shop_identify_prompt")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_identify_character", "shop_identify_character"), character.Name, len(character.Equipment), character.Gold))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_IDENTIFY_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_identify_exit", "shop_identify_exit"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_IDENTIFY_EXIT")
	s.Message = ""
}

func (s *State) enterShopIdentifyItemMenu() {
	s.shopIdentifyMenu = true
	s.shopIdentifyItemMenu = true
	s.Mode = ModePlace
	character := s.partyRoster[s.shopIdentifyCharacter]
	s.Prompt = fmt.Sprintf(s.catalog.Text("shop_identify_item_prompt", "shop_identify_item_prompt"), character.Name)
	s.Choices = make([]string, 0, len(character.Equipment)+1)
	s.currentOriginalChoices = make([]string, 0, len(character.Equipment)+1)
	for index, item := range character.Equipment {
		s.Choices = append(s.Choices, monster.LocalizedItemName(item, s.catalog))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_IDENTIFY_ITEM_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_identify_exit", "shop_identify_exit"))
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
	s.Prompt = s.catalog.Text("shop_appraise_prompt", "shop_appraise_prompt")
	s.Choices = make([]string, 0, len(s.partyRoster)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.partyRoster)+1)
	for index, character := range s.partyRoster {
		s.Choices = append(s.Choices, fmt.Sprintf(s.catalog.Text("shop_appraise_character", "shop_appraise_character"), character.Name, character.Gems, character.Jewelry))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_CHARACTER_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_appraise_exit", "shop_appraise_exit"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_EXIT")
	s.Message = ""
}

func (s *State) enterShopAppraiseTreasureMenu() {
	s.shopAppraiseMenu = true
	s.shopAppraiseConfirm = false
	s.Mode = ModePlace
	character := s.partyRoster[s.shopAppraiseCharacter]
	s.Prompt = s.catalog.Text("shop_appraise_treasure_prompt", "shop_appraise_treasure_prompt")
	s.Choices = make([]string, 0, 3)
	s.currentOriginalChoices = make([]string, 0, 3)
	if character.Gems > 0 {
		label := s.catalog.Text("shop_gems_offer_unavailable", "shop_gems_offer_unavailable")
		if s.appraisalOffers.GemsReady {
			label = fmt.Sprintf(s.catalog.Text("shop_gems_offer", "shop_gems_offer"), s.appraisalOffers.Gems)
		}
		s.Choices = append(s.Choices, label)
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_TREASURE_1")
	}
	if character.Jewelry > 0 {
		label := s.catalog.Text("shop_jewelry_offer_unavailable", "shop_jewelry_offer_unavailable")
		if s.appraisalOffers.JewelryReady {
			label = fmt.Sprintf(s.catalog.Text("shop_jewelry_offer", "shop_jewelry_offer"), s.appraisalOffers.Jewelry)
		}
		s.Choices = append(s.Choices, label)
		s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_TREASURE_2")
	}
	s.Choices = append(s.Choices, s.catalog.Text("shop_appraise_exit", "shop_appraise_exit"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "SHOP_APPRAISE_EXIT")
	s.Message = ""
}

func (s *State) enterShopAppraiseConfirmMenu() {
	s.shopAppraiseMenu = true
	s.shopAppraiseConfirm = true
	s.Mode = ModePlace
	s.Prompt = s.catalog.Text("shop_appraise_confirm_prompt", "shop_appraise_confirm_prompt")
	s.Choices = []string{
		s.catalog.Text("shop_appraise_accept", "shop_appraise_accept"),
		s.catalog.Text("shop_appraise_reject", "shop_appraise_reject"),
		s.catalog.Text("shop_appraise_cancel", "shop_appraise_cancel"),
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
		return s.catalog.Text("inn_event", "inn_event")
	case "STORE":
		return s.catalog.Text("store_event", "store_event")
	case "BAR":
		return s.catalog.Text("bar_event", "bar_event")
	default:
		return s.localizeOption(originalChoice)
	}
}

func (s *State) enterMap() {
	s.enterMapAt(0, 0)
}

func (s *State) enterMapAt(x, y int) {
	s.Mode = ModeMap
	s.MapX, s.MapY = x, y
	cityFlags, ok := mapdata.CityInfo(int(s.Area.CurrentCity))
	if !ok {
		cityFlags = 0
	}
	s.WildernessFloor = mapdata.GenerateWilderness(cityFlags, s.mapSeed)
	s.Choices = nil
	if s.pendingWorldTravel {
		s.Prompt = s.catalog.Text("world_travel_map_prompt", "荒野旅程　方向鍵：移動　Enter：抵達目的地　Esc：返回")
	} else {
		s.Prompt = s.catalog.Text("shadowdale_map_prompt", "暗影谷荒野")
	}
	s.Message = ""
}

// finishNewGameEntry mirrors sub_29758 after the initial ECL entry returns:
// a fresh campaign remains in the indoor DungeonMap state established by
// seg001 and uses the script-written map position and half-direction.
func (s *State) finishNewGameEntry() {
	s.newGameEntryActive = false
	x, y := uint16(7), uint16(13)
	if value, ok := s.session.MemoryValue(0xC04B); ok {
		x = value
	}
	if value, ok := s.session.MemoryValue(0xC04C); ok {
		y = value
	}
	if value, ok := s.session.MemoryValue(0xC04D); ok {
		s.DungeonDirection = uint8(value&3) * 2
	}
	s.DungeonX, s.DungeonY = int(x), int(y)
	s.MapX, s.MapY = int(x), int(y)
	s.Location = LocationTilverton
	s.LocationName = s.catalog.Text("tilverton", "提爾佛頓")
	s.OriginalLocation = "TILVERTON"
	s.Mode = ModeDungeon
	s.Choices = nil
	s.Prompt = s.catalog.Text("dungeon_prompt", "↑：前進　K／M：轉向　S：搜索　E：紮營")
	s.Message = ""
}

// RunDungeonLifecycle synchronizes reference map registers and invokes the
// per-turn then search-location ECL entries used by sub_29758 after a
// successful forward step.
func (s *State) RunDungeonLifecycle() error {
	return s.runDungeonLifecycle(false)
}

// SearchDungeonLocation mirrors the explicit dungeon SEARCH command. The
// reference engine exposes that action through work flag 0x7ECA while running
// SearchLocation; ordinary movement keeps the flag clear.
func (s *State) SearchDungeonLocation() error {
	if s.Mode != ModeDungeon {
		return fmt.Errorf("dungeon search is invalid in mode %d", s.Mode)
	}
	if s.session == nil {
		return fmt.Errorf("dungeon search requires an ECL session")
	}
	s.syncDungeonECLRegisters()
	s.session.SetMemoryValue(0x7ECA, 1)
	defer s.session.SetMemoryValue(0x7ECA, 0)
	blockBefore := s.session.CurrentBlockID()
	result, err := s.session.RunEntrySeedWithPartyContext(
		1, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
	s.requestMusicIfBlockChanged(blockBefore)
	_, err = s.applyDungeonLifecycleResult(result)
	return err
}

func (s *State) runDungeonLifecycle(exitAttempt bool) error {
	if s.Mode != ModeDungeon {
		return fmt.Errorf("dungeon lifecycle is invalid in mode %d", s.Mode)
	}
	if s.session == nil {
		return fmt.Errorf("dungeon lifecycle requires an ECL session")
	}
	// A new movement transaction must not inherit text or menu labels from the
	// previous terrain boundary. Some original events deliberately emit an
	// empty PRESS pause; leaving Message untouched would display stale combat
	// or story text as though that blank pause repeated it.
	s.Message = ""
	s.Choices = nil
	s.currentOriginalChoices = nil
	s.syncDungeonECLRegisters()
	if exitAttempt {
		s.dungeonBoundaryAttempt = true
		x, y, direction := s.DungeonGeometryView()
		s.session.SetMemoryValue(0xC04B, uint16(x))
		s.session.SetMemoryValue(0xC04C, uint16(y))
		s.session.SetMemoryValue(0xC04D, uint16(direction/2))
		// The reference movement helper clears the stale forced-move sentinel
		// before per-turn ECL decides whether this boundary attempt is real.
		s.session.SetMemoryValue(0x7EC9, 0)
		s.session.SetMemoryValue(0x7ED5, 1)
	}

	for _, entry := range []int{0, 1} {
		blockBefore := s.session.CurrentBlockID()
		result, err := s.session.RunEntrySeedWithPartyContext(
			entry, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
		)
		if err != nil {
			return err
		}
		if s.session.CurrentBlockID() != blockBefore {
			s.syncDungeonStateFromECLRegisters()
			if s.dungeonBoundaryAttempt {
				s.session.SetMemoryValue(0x7ED5, 0)
				s.dungeonBoundaryAttempt = false
			}
		}
		s.requestMusicIfBlockChanged(blockBefore)
		handled, err := s.applyDungeonLifecycleResult(result)
		if err != nil {
			return err
		}
		// The reference loop always runs SearchLocation after a quiet
		// per-turn invocation. Some CALL-only entries retain an empty packed
		// text element; that is not a player-visible boundary and must not
		// suppress the terrain dispatcher.
		if entry == 0 {
			if dungeonLifecycleResultBlocksSearch(result) {
				return nil
			}
			continue
		}
		if handled {
			return nil
		}
	}
	return nil
}

func dungeonLifecycleResultBlocksSearch(result ecl.RunResult) bool {
	return result.PictureRequested ||
		result.ShopRequested ||
		result.TempleRequested ||
		len(result.TreasureRequests) > 0 ||
		result.CombatRequested ||
		(result.WaitingForMenu && len(result.Menus) > 0) ||
		result.WaitingForString ||
		hasMeaningfulECLText(result.Text)
}

func hasMeaningfulECLText(texts []string) bool {
	for _, text := range texts {
		if strings.TrimSpace(text) != "" {
			return true
		}
	}
	return false
}

// RunDungeonExitLifecycle supplies the boundary-crossing signal before running
// the normal per-step entries. ECL2 block 2 gates both Thieves' Guild sewer
// exits on 0x7ED5; the original 3-D movement loop sets this state only when a
// passable step attempts to leave the 16x16 geometry.
func (s *State) RunDungeonExitLifecycle() error {
	if s.session == nil {
		return fmt.Errorf("dungeon exit lifecycle requires an ECL session")
	}
	return s.runDungeonLifecycle(true)
}

// StartDungeonStoryPreview enters a verified ECL dungeon block through its
// normal initialization entry. It is used by reproducible frontend previews;
// previousBlockID supplies the same source-block context that NEWECL would
// leave in the DOS dispatcher.
func (s *State) StartDungeonStoryPreview(blockID, previousBlockID, gameArea uint8) error {
	if s.session == nil {
		return fmt.Errorf("dungeon story preview requires an ECL session")
	}
	blockBefore := s.session.CurrentBlockID()
	if err := s.session.Switch(blockID); err != nil {
		return err
	}
	s.requestMusicIfBlockChanged(blockBefore)
	s.eclBlock = s.session.CurrentData()
	start, err := s.session.InitialEntry()
	if err != nil {
		return err
	}
	s.eclStart = start
	s.session.SetMemoryValue(0x4BF2, uint16(previousBlockID))
	s.session.SetMemoryValue(0x7ED5, 0)
	s.session.SetMemoryValue(0x7EC9, 0)
	s.Area.GameArea = gameArea
	s.Area.InDungeon = true
	s.GeoMapSet = gameArea
	s.Mode = ModeDungeon
	blockBefore = s.session.CurrentBlockID()
	result, err := s.session.RunEntrySeedWithPartyContext(
		4, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
	s.requestMusicIfBlockChanged(blockBefore)
	s.applyGeoMapLoad(result)
	s.applyLoadPieces(result)
	s.syncDungeonStateFromECLRegisters()
	_, err = s.applyDungeonLifecycleResult(result)
	return err
}

func (s *State) syncDungeonStateFromECLRegisters() {
	if value, ok := s.session.MemoryValue(0xC04B); ok {
		s.DungeonX = int(int16(value))
	}
	if value, ok := s.session.MemoryValue(0xC04C); ok {
		s.DungeonY = int(int16(value))
	}
	if value, ok := s.session.MemoryValue(0xC04D); ok {
		s.DungeonDirection = uint8(value&3) * 2
	}
	s.MapX, s.MapY = s.DungeonX, s.DungeonY
}

func (s *State) syncDungeonECLRegisters() {
	s.session.SetMemoryValue(0xC04B, uint16(s.DungeonX))
	s.session.SetMemoryValue(0xC04C, uint16(s.DungeonY))
	s.session.SetMemoryValue(0xC04D, uint16(s.DungeonDirection/2))
	s.session.SetMemoryValue(0xC04E, uint16(s.DungeonWallType))
	s.session.SetMemoryValue(0xC04F, uint16(s.DungeonWallRoof))
}

// DungeonGeometryView maps ECL-facing map registers to the cell and facing
// used by the combined GEO geometry. Tilverton's ECL block 2 stores the
// Thieves' Guild as a local 8x16 map mirrored into the right half of GEO2
// block 1: script (1,12,N) corresponds to geometry (9,3,S).
func (s *State) DungeonGeometryView() (x, y int, direction uint8) {
	x, y, direction = s.DungeonX, s.DungeonY, s.DungeonDirection
	if s.session != nil && s.session.CurrentBlockID() == 0x02 &&
		s.GeoMapSet == 2 && s.GeoMapBlock == 1 {
		x = (x + 8) % geo.Width
		y = geo.Height - 1 - y
		direction = (4 - direction + 8) % 8
	}
	return x, y, direction
}

// SetDungeonGeometryView is the inverse adapter used by renderers after a
// movement in combined GEO coordinates. ECL continues to see its local map.
func (s *State) SetDungeonGeometryView(x, y int, direction uint8) {
	oldX, oldY, _ := s.DungeonGeometryView()
	moved := oldX != x || oldY != y
	if s.session != nil && s.session.CurrentBlockID() == 0x02 &&
		s.GeoMapSet == 2 && s.GeoMapBlock == 1 {
		s.DungeonX = (x + 8) % geo.Width
		s.DungeonY = geo.Height - 1 - y
		s.DungeonDirection = (4 - direction + 8) % 8
		if moved {
			s.session.SetMemoryValue(0x7F81, 0)
		}
		return
	}
	s.DungeonX, s.DungeonY, s.DungeonDirection = x, y, direction
	if moved && s.session != nil {
		// SearchLocation uses 7F81h as a per-successful-step event guard:
		// event branches set it after firing, and later terrain dispatchers
		// exit while it remains one. The reference movement helper clears
		// that transient before the next search-location invocation.
		s.session.SetMemoryValue(0x7F81, 0)
	}
}

func (s *State) applyDungeonLifecycleResult(result ecl.RunResult) (bool, error) {
	s.applyGeoMapLoad(result)
	s.applyLoadPieces(result)
	s.applyECLCallSignals(result)
	s.applySpellSignals(result)
	s.applyECLDamageSignals(result)
	s.applyECLLoadCharacterSignals(result)
	if err := s.applyECLNPCSignals(result); err != nil {
		return true, err
	}
	if err := s.applyECLDumpSignals(result); err != nil {
		return true, err
	}
	if err := s.applyECLClockSignals(result); err != nil {
		return true, err
	}
	s.applyECLInventorySignals(result)
	s.applyECLTreasureSignals(result)
	s.applyECLRobSignals(result)
	if handled, err := s.applyDataPackEvent(result); handled || err != nil {
		return handled, err
	}
	if hasMeaningfulECLText(result.Text) {
		s.unlockJournalEntries(result.Text)
		s.Message = s.localizeECLText(result.Text)
		if result.WaitingForMenu && len(result.Menus) > 0 &&
			slices.Equal(result.Menus[len(result.Menus)-1].Options,
				[]string{"PRESS BUTTON OR RETURN TO CONTINUE."}) {
			s.pendingTreasureMessage = s.Message
		}
	}
	s.eclMenuReturnMode = ModeDungeon
	s.eventReturnMode = ModeDungeon
	if result.WaitingForString && len(result.StringInputRequests) > 0 {
		s.beginECLStringInput(result.StringInputRequests[len(result.StringInputRequests)-1])
		return true, nil
	}
	if result.PictureRequested {
		s.requestMusicForSignal("picture", result.PictureBlock)
		s.Mode = ModeEvent
		s.PictureRequested = true
		s.PictureBlock = result.PictureBlock
		s.BigPictureRequested = result.BigPictureRequested
		if result.PictureHeadBlockSet {
			s.SceneHeadBlock = uint8(result.PictureHeadBlock)
		}
		s.SceneCharacterRequested = !result.BigPictureRequested && s.SceneHeadBlock != 0xFF
		if s.SceneCharacterRequested {
			s.SceneBodyBlock = uint8(result.PictureBlock)
		}
		s.OriginalEvent = "PICTURE"
		if result.CombatRequested || result.ShopRequested || result.TempleRequested ||
			result.WaitingForMenu || result.WaitingForString {
			pending := result
			pending.PictureRequested = false
			s.pendingPictureResult = &pending
		}
		return true, nil
	}
	if result.ShopRequested {
		return true, s.enterECLShop(result)
	}
	if result.TempleRequested {
		return true, s.enterECLTemple()
	}
	if len(result.TreasureRequests) > 0 {
		beforeMoney := s.moneyPool
		beforeGems, beforeJewelry := s.treasureGems, s.treasureJewelry
		beforeItems := len(s.pendingTreasureItems)
		if err := s.ResolveTreasureRequests(); err != nil {
			return true, err
		}
		if s.moneyPool != beforeMoney ||
			s.treasureGems != beforeGems ||
			s.treasureJewelry != beforeJewelry ||
			len(s.pendingTreasureItems) != beforeItems {
			s.treasureResumeECL = s.session != nil && len(s.eclBlock) > 0
			s.enterTreasureMenuFor(ModeDungeon)
			if hasMeaningfulECLText(result.Text) {
				s.Message = s.localizeECLText(result.Text)
			}
			return true, nil
		}
	}
	if result.CombatRequested {
		records := s.monsterRecordsForCurrentECL()
		if len(result.MonsterSpawns) > 0 && len(records) > 0 {
			return true, s.StartEncounterWithAffects(result, records, s.monsterAffectsForCurrentECL(), s.party, s.combatSeed)
		}
		s.Mode = ModeEvent
		s.OriginalEvent = "COMBAT"
		return true, nil
	}
	if result.WaitingForMenu && len(result.Menus) > 0 {
		s.enterECLMenu(result.Menus[len(result.Menus)-1])
		return true, nil
	}
	if hasMeaningfulECLText(result.Text) {
		s.Mode = ModeEvent
		return true, nil
	}
	return false, nil
}

// applyDataPackEvent is the title adapter for the reusable declarative engine.
// It projects only typed runtime inputs requested by the pack, then commits
// generic outputs back to the persistent roster and UI. Plot flags, character
// names, messages and destination block IDs remain in JSON.
func (s *State) applyDataPackEvent(result ecl.RunResult) (bool, error) {
	if s.dataPackError != nil {
		return true, fmt.Errorf("load CoAB game pack: %w", s.dataPackError)
	}
	if s.dataPack == nil || s.session == nil {
		return false, nil
	}
	memory := make(map[uint16]uint16)
	for _, address := range s.dataPack.MemoryAddresses() {
		if value, ok := s.session.MemoryValue(address); ok {
			memory[address] = value
		}
	}
	roster := make([]goldenbox.Member, 0, len(s.partyRoster))
	for _, character := range s.partyRoster {
		roster = append(roster, goldenbox.Member{
			ID: character.ID, ScriptName: character.ScriptName, HitPoints: character.HitPoints,
		})
	}
	runtime := &goldenbox.Runtime{
		ECLBlock:      s.session.CurrentBlockID(),
		Memory:        memory,
		Roster:        roster,
		Locale:        s.catalog.Language,
		PendingMenu:   result.WaitingForMenu && len(result.Menus) > 0,
		AppliedEvents: s.appliedDataPackEvents,
	}
	applied, err := s.dataPack.ApplyFirst(runtime)
	if err != nil {
		return true, err
	}
	if !applied.Applied {
		return false, nil
	}

	removedIDs := make(map[string]bool, len(runtime.RemovedMembers))
	for _, id := range runtime.RemovedMembers {
		removedIDs[id] = true
	}
	keptRoster := s.partyRoster[:0]
	for _, character := range s.partyRoster {
		if !removedIDs[character.ID] {
			keptRoster = append(keptRoster, character)
		}
	}
	s.partyRoster = keptRoster
	keptFighters := s.party[:0]
	for _, fighter := range s.party {
		if !removedIDs[fighter.ID] {
			keptFighters = append(keptFighters, fighter)
		}
	}
	s.party = keptFighters
	s.whoSelectedIndex = -1
	s.Message = runtime.Message
	s.Mode = ModeEvent
	switch runtime.Mode {
	case "world_menu":
		s.eventReturnMode = ModeWilderness
		s.eclMenuReturnMode = ModeWilderness
	default:
		return true, fmt.Errorf("data-pack event %q returned unsupported mode %q", applied.EventID, runtime.Mode)
	}
	s.currentOriginalChoices = []string{"PRESS BUTTON OR RETURN TO CONTINUE."}
	s.Choices = []string{s.localizeOption(s.currentOriginalChoices[0])}
	if result.WaitingForMenu && len(result.Menus) > 0 {
		menu := result.Menus[len(result.Menus)-1]
		s.pendingECLMenu = &menu
		if hasMeaningfulECLText(result.Text) {
			s.pendingECLMenuMessage = s.localizeECLText(result.Text)
		}
	}
	return true, nil
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
	if s.pendingWorldTravel {
		return s.completeWildernessTravel()
	}
	s.Mode = ModePlace
	s.Prompt = s.placePrompt()
	s.Choices = []string{
		s.catalog.Text("inn", "客棧"),
		s.catalog.Text("store", "store"),
		s.catalog.Text("bar", "bar"),
		s.catalog.Text("leave", "離開"),
	}
	s.currentOriginalChoices = []string{"INN", "STORE", "BAR", "LEAVE"}
	s.Message = ""
	return nil
}

func (s *State) completeWildernessTravel() error {
	if s.session == nil {
		return fmt.Errorf("wilderness travel requires an ECL session")
	}
	destination := s.pendingWorldDestination
	s.pendingWorldTravel = false
	return s.arriveAtWorldLocation(destination)
}

func (s *State) arriveAtWorldLocation(destination uint8) error {
	if s.session == nil {
		return fmt.Errorf("world arrival requires an ECL session")
	}
	s.session.SetMemoryValue(0x4C9C, uint16(destination))
	blockBefore := s.session.CurrentBlockID()
	result, err := s.session.RunEntrySeedWithPartyContext(
		1, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
	s.requestMusicIfBlockChanged(blockBefore)
	s.eclBlock = s.session.CurrentData()
	s.setWorldLocation(uint16(destination))
	s.Area.GameArea = 1
	s.Area.InDungeon = false
	s.applyGeoMapLoad(result)
	s.applyLoadPieces(result)
	s.applyECLCallSignals(result)
	if hasMeaningfulECLText(result.Text) {
		s.Message = s.localizeECLText(result.Text)
	}
	s.Mode = ModeWilderness
	s.eventReturnMode = ModeWilderness
	s.eclMenuReturnMode = ModeWilderness
	if result.WaitingForMenu && len(result.Menus) > 0 {
		s.enterECLMenu(result.Menus[len(result.Menus)-1])
		return nil
	}
	return fmt.Errorf("wilderness arrival at %d did not produce a world menu", destination)
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
		if s.pendingPictureResult != nil {
			result := *s.pendingPictureResult
			s.pendingPictureResult = nil
			records := s.monsterRecordsForCurrentECL()
			if result.ShopRequested {
				return s.enterECLShop(result)
			}
			if result.TempleRequested {
				return s.enterECLTemple()
			}
			if result.CombatRequested {
				if len(result.MonsterSpawns) > 0 && len(s.party) > 0 && len(records) > 0 {
					return s.StartEncounterWithAffects(result, records, s.monsterAffectsForCurrentECL(), s.party, s.combatSeed)
				}
				s.OriginalEvent = "COMBAT"
				s.Message = s.catalog.Text("combat_started", "戰鬥開始（戰鬥資料尚未完成）")
				s.eventReturnMode = ModeWilderness
				return nil
			}
			if result.WaitingForString && len(result.StringInputRequests) > 0 {
				s.beginECLStringInput(result.StringInputRequests[len(result.StringInputRequests)-1])
				return nil
			}
			if result.WaitingForMenu && len(result.Menus) > 0 {
				s.enterECLMenu(result.Menus[len(result.Menus)-1])
				return nil
			}
		}
	}
	if s.pendingECLMenu != nil {
		menu := *s.pendingECLMenu
		s.pendingECLMenu = nil
		s.enterECLMenu(menu)
		s.Message = s.pendingECLMenuMessage
		s.pendingECLMenuMessage = ""
		return nil
	}
	switch s.eventReturnMode {
	case ModeWilderness:
		if s.trainingMenu {
			s.enterTrainingMenu()
			return nil
		}
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
		if s.templeMenu {
			s.enterTempleMenu()
			return nil
		}
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
	case ModeDungeon:
		s.Mode = ModeDungeon
		s.Message = ""
		s.eclMenuReturnMode = ModeTitle
		s.syncCurrentECLDungeonArea()
		if s.pendingDungeonEntry {
			s.syncDungeonStateFromECLRegisters()
			s.pendingDungeonEntry = false
		}
		return nil
	default:
		return fmt.Errorf("event has no continuation")
	}
}

func (s *State) syncCurrentECLDungeonArea() {
	if s.session == nil || s.Area.InDungeon {
		return
	}
	blockID := s.session.CurrentBlockID()
	if blockID >= 0x50 {
		return
	}
	gameArea := monsterChapterForBlock(blockID)
	s.Area.GameArea = gameArea
	s.Area.InDungeon = true
	s.GeoMapSet = gameArea
	s.GeoMapBlock = blockID
	s.geoMapPending = true
	s.syncDungeonStateFromECLRegisters()
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
		s.catalog.Text("store", "store"),
		s.catalog.Text("bar", "bar"),
		s.catalog.Text("leave", "離開"),
	}
	s.currentOriginalChoices = []string{"INN", "STORE", "BAR", "LEAVE"}
	s.Message = ""
	return nil
}

func (s *State) placePrompt() string {
	return "你在" + s.LocationName + "。要去哪裡？"
}

// applyCitySelection prefers the ECL world dispatcher's current/destination
// bytes, then retains the observed opening sequence as a compatibility
// fallback for synthetic sessions.
func (s *State) applyCitySelection() {
	if s.session != nil &&
		(s.session.CurrentBlockID() == 0x50 || s.session.CurrentBlockID() == 0x51) {
		if s.pendingWorldTravel {
			s.setWorldLocation(uint16(s.pendingWorldDestination))
			s.Area.GameArea = 1
			s.Area.InDungeon = false
			return
		}
		for _, address := range []uint16{0x4C9B, 0x4C9C} {
			if value, ok := s.session.MemoryValue(address); ok && value <= 13 {
				s.setWorldLocation(value)
				// The global ECL1 dispatcher owns city edges, city services,
				// and wilderness travel. A party may arrive here through an
				// Area 5 DEPART continuation, so clear the old dungeon
				// namespace instead of leaking its GEO/CPIC state into town.
				s.Area.GameArea = 1
				s.Area.InDungeon = false
				return
			}
		}
	}
	if len(s.selectionSequence) < 4 || s.selectionSequence[0] != 0 || s.selectionSequence[1] != 0 || s.selectionSequence[2] != 1 {
		return
	}
	choice := s.selectionSequence[3]
	if choice > 2 {
		return
	}
	s.setNamedLocation(int(choice))
}

func (s *State) setWorldLocation(value uint16) {
	s.Area.CurrentCity = uint8(value)
	if value == 0 {
		s.Location = LocationTilverton
		s.LocationName = s.catalog.Text("tilverton", "提爾佛頓")
		s.OriginalLocation = "TILVERTON"
		return
	}
	if value >= 1 && value <= 3 {
		s.setNamedLocation(int(value - 1))
		s.Area.CurrentCity = uint8(value)
		return
	}
	if value == 4 {
		s.Location = LocationStandingStone
		s.LocationName = s.catalog.Text("standing_stone", "立石群")
		s.OriginalLocation = "THE STANDING STONE"
		return
	}
	if value == 8 {
		s.Location = LocationEssembra
		s.LocationName = s.catalog.Text("essembra", "艾森布拉")
		s.OriginalLocation = "ESSEMBRA"
		return
	}
	if value == 9 {
		s.Location = LocationHap
		s.LocationName = s.catalog.Text("hap", "哈普")
		s.OriginalLocation = "HAP"
		return
	}
	worldLocations := map[uint16]struct {
		location Location
		key      string
		original string
	}{
		5:  {LocationVoonlar, "voonlar", "VOONLAR"},
		6:  {LocationPhlan, "phlan", "PHLAN"},
		7:  {LocationTeshwave, "teshwave", "TESHWAVE"},
		10: {LocationYulash, "yulash", "YULASH"},
		11: {LocationHillsfar, "hillsfar", "HILLSFAR"},
		12: {LocationZhentilKeep, "zhentil_keep", "ZHENTIL KEEP"},
		13: {LocationMythDrannor, "myth_drannor", "MYTH DRANNOR"},
	}
	if selected, ok := worldLocations[value]; ok {
		s.Location = selected.location
		s.LocationName = s.catalog.Text(selected.key, selected.original)
		s.OriginalLocation = selected.original
	}
}

func (s *State) setNamedLocation(choice int) {
	locations := [...]struct {
		location Location
		key      string
		original string
	}{
		{LocationShadowdale, "shadowdale", "SHADOWDALE"},
		{LocationAshabenford, "ashabenford", "ASHABENFORD"},
		{LocationDaggerFalls, "dagger_falls", "DAGGER FALLS"},
	}
	if choice < 0 || choice >= len(locations) {
		return
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
	s.Prompt = s.catalog.Text("press_button", "press_button")
	s.Message = ""
}

func (s *State) localizeOption(option string) string {
	if s != nil && s.dataPack != nil {
		if value, ok := s.dataPack.LocalizeOption(option, s.catalog.Language); ok {
			return value
		}
	}
	return option
}

func (s *State) localizePrompt(prompt string) string {
	if prompt == "PRESS BUTTON OR RETURN TO CONTINUE." {
		return s.catalog.Text("press_button", "press_button")
	}
	if strings.HasPrefix(prompt, "HOW WILL YOU GET TO ") {
		destination := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(prompt, "HOW WILL YOU GET TO "), "?"))
		return fmt.Sprintf(s.catalog.Text("route_prompt", "route_prompt"), s.localizeOption(destination))
	}
	if prompt == "FROM HERE YOU MAY JOURNEY TO" {
		return s.catalog.Text("journey_destination_prompt", "journey_destination_prompt")
	}
	if prompt == "WHAT WILL YOU DRINK?" {
		return s.catalog.Text("tavern_drink_prompt", "tavern_drink_prompt")
	}
	if prompt == "A DARK ELF PATROL ARRIVES" {
		return s.catalog.Text("ecl_hap_dark_elf_patrol", "ecl_hap_dark_elf_patrol")
	}
	return prompt
}

func (s *State) localizeECLText(texts []string) string {
	if s.dataPack != nil {
		if result := s.dataPack.MatchText(texts, s.catalog.Language); result.Matched {
			return result.Message
		}
	}
	return localizeECLText(s.catalog, texts)
}

func localizeECLText(catalog locale.Catalog, texts []string) string {
	joined := strings.Join(texts, " ")
	switch {
	case strings.Contains(joined, "WHAT WILL YOU DRINK"):
		return catalog.Text("tavern_drink_prompt", "tavern_drink_prompt")
	}
	localized := make([]string, 0, len(texts))
	for _, text := range texts {
		localized = append(localized, localizeECLLine(catalog, text))
	}
	return strings.Join(localized, " ")
}

// unlockJournalEntries bridges original "record it in Journal Entry N"
// script text to the remake's in-game journal. Entries are appended only
// after their event fires, so the bundled manual does not reveal future
// clues prematurely.
func (s *State) unlockJournalEntries(texts []string) {
	if s.dataPack == nil {
		return
	}
	match := s.dataPack.MatchText(texts, s.catalog.Language)
	for _, page := range match.JournalPages {
		s.appendJournalPages(page, []string{page})
	}
}

func (s *State) appendJournalPages(marker string, pages []string) {
	for _, page := range s.JournalPages {
		if strings.HasPrefix(page, marker) {
			return
		}
	}
	s.JournalPages = append(s.JournalPages, pages...)
}

func localizeECLLine(catalog locale.Catalog, line string) string {
	switch strings.TrimSpace(line) {
	case "DO YOU WANT TO TRAIN?", "'DO YOU WANT TO TRAIN?'":
		return catalog.Text("ecl_training_prompt", "ecl_training_prompt")
	case "YOU'RE SHOWING GREAT PROGRESS. RETURN AGAIN WHEN":
		return catalog.Text("ecl_training_progress", "ecl_training_progress")
	case "YOU ARE READY.' YOU EXIT THE HALL.":
		return catalog.Text("ecl_training_exit", "ecl_training_exit")
	case "'WHAT'S YOUR PLEASURE?'":
		return catalog.Text("ecl_tavern_pleasure", "ecl_tavern_pleasure")
	case "'A SPECIAL CUSTOMER'S ARRIVED. YOU HAVE TO SLIP":
		return catalog.Text("ecl_tavern_special_1", "ecl_tavern_special_1")
	case "OUTSIDE FOR A MOMENT.' DO YOU GO?":
		return catalog.Text("ecl_tavern_special_2", "ecl_tavern_special_2")
	case "AS YOU BEGIN TO WALK OUT THE DOOR, YOU SEE A":
		return catalog.Text("ecl_tavern_purple_1", "ecl_tavern_purple_1")
	case "YOUNG WOMAN WITH A PURPLE SASH SLIP IN THE SIDE DOOR.":
		return catalog.Text("ecl_tavern_purple_2", "ecl_tavern_purple_2")
	case "A FEW OF THE OTHER PATRONS HANG BACK, AS IF TO MEET HER.":
		return catalog.Text("ecl_tavern_purple_3", "ecl_tavern_purple_3")
	case "AS YOU CONSIDER YOUR NEXT MOVE, YOU HEAR A":
		return catalog.Text("ecl_tavern_commotion_1", "ecl_tavern_commotion_1")
	case "COMMOTION AROUND THE SIDE OF THE BUILDING. DO YOU GO":
		return catalog.Text("ecl_tavern_commotion_2", "ecl_tavern_commotion_2")
	case "TO INVESTIGATE?":
		return catalog.Text("ecl_tavern_commotion_3", "ecl_tavern_commotion_3")
	case "A PATROL ARRIVES.":
		return catalog.Text("ecl_tilverton_patrol_arrives", "ecl_tilverton_patrol_arrives")
	case "ROYAL GUARDS TELL YOU TO MOVE ALONG.":
		return catalog.Text("ecl_tilverton_guards_move", "ecl_tilverton_guards_move")
	case "'WELCOME TO THE FAIR CITY OF TILVERTON,' BEAMS THE":
		return catalog.Text("ecl_tilverton_inn_welcome", "ecl_tilverton_inn_welcome")
	case "INNKEEPER. THEN SHE NOTICES YOUR COLLECTIVE SCOWLS.":
		return catalog.Text("ecl_tilverton_inn_scowls", "ecl_tilverton_inn_scowls")
	case "'PLEASE CALM DOWN WHILE I EXPLAIN.'":
		return catalog.Text("ecl_tilverton_inn_calm", "ecl_tilverton_inn_calm")
	case "YOU LISTEN,":
		return catalog.Text("ecl_tilverton_inn_listen", "ecl_tilverton_inn_listen")
	case "'I AM THE SAGE FILANI. YOU ARE HERE ABOUT THE SIGILS,":
		return catalog.Text("ecl_filani_intro", "ecl_filani_intro")
	case "CORRECT?'":
		return catalog.Text("ecl_filani_correct", "ecl_filani_correct")
	case "'THIS IS AN INTERESTING CASE. I'LL DO IT FOR HALF YOUR":
		return catalog.Text("ecl_filani_price", "ecl_filani_price")
	case "FUNDS. HOW MUCH DO YOU HAVE?'":
		return catalog.Text("ecl_filani_funds", "ecl_filani_funds")
	case "'DO NOT THINK SAGES ARE FOOLS.' SHE SENDS YOU OUT.":
		return catalog.Text("ecl_filani_lie", "ecl_filani_lie")
	case "'THEN WE HAVE NOTHING TO DISCUSS.'":
		return catalog.Text("ecl_filani_no", "ecl_filani_no")
	case "'WE HAVE A SELECTION OF THE FINEST CORMYR STEEL.":
		return catalog.Text("ecl_weaponers_intro", "ecl_weaponers_intro")
	case "INTERESTED?":
		return catalog.Text("ecl_weaponers_interested", "ecl_weaponers_interested")
	case "'MAY YOU ALWAYS STRIKE TRUE.'":
		return catalog.Text("ecl_weaponers_farewell", "ecl_weaponers_farewell")
	case "'GOOD DAY THEN.'":
		return catalog.Text("ecl_weaponers_decline", "ecl_weaponers_decline")
	case "'GOOD DAY TO YOU, GENTLE PERSONS. DO YOU WISH":
		return catalog.Text("ecl_general_store_intro", "ecl_general_store_intro")
	case "TO MAKE A PURCHASE?'":
		return catalog.Text("ecl_general_store_purchase", "ecl_general_store_purchase")
	case "'THANK YOU. RETURN SOON.'":
		return catalog.Text("ecl_general_store_farewell", "ecl_general_store_farewell")
	case "YOU MOVE AWAY.":
		return catalog.Text("ecl_move_away", "ecl_move_away")
	case "SMOKE RISES FROM BEHIND THE RUINED WALLS":
		return catalog.Text("ecl_smoke_rises", "ecl_smoke_rises")
	case "OF YULASH. THE SOUND":
		return catalog.Text("ecl_yulash_sound", "ecl_yulash_sound")
	case "OF BATTLE RINGS OUT FROM INSIDE":
		return catalog.Text("ecl_battle_rings", "ecl_battle_rings")
	case "YOU FIND A WAR BLASTED SECTION OF THE CITY.":
		return catalog.Text("ecl_war_blasted_city", "ecl_war_blasted_city")
	case "YOU DISCOVER A SMALL MAGIC SHOP.":
		return catalog.Text("ecl_small_magic_shop", "ecl_small_magic_shop")
	default:
		return line
	}
}
