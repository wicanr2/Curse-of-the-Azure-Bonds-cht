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

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dungeon"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
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

	catalog                locale.Catalog
	eclBlock               []byte
	eclStart               int
	selectionSequence      []uint16
	whoSelectionSequence   []uint16
	whoMenu                bool
	whoSelectedIndex       int
	loadCharacterNotFound  bool
	loadCharacterHighBit   bool
	currentOriginalChoices []string
	eventReturnMode        Mode
	journalReturnMode      Mode
	creationReturnMode     Mode
	session                *ecl.BlockSession
	pendingPictureResult   *ecl.RunResult
	newGameEntryActive     bool
	eclMenuReturnMode      Mode
	party                  []combat.Fighter
	partyRoster            party.Roster
	savgamPrefix           *partySave.SAVGAMContainer
	savgamPlayers          map[string]party.DOSPlayerFiles
	pendingSoundEvents     []SoundEvent
	pendingECLCalls        []uint16
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
	combatReferenceCoords  bool
	combatView             bool
	combatViewFighterID    string
	combatMessage          string
	combatReturnMode       Mode
	monsterRecords         map[uint8]monster.Record
	monsterRecordsByECL    map[uint8]map[uint8]monster.Record
	monsterAffects         map[uint8][]monster.AffectRecord
	monsterAffectsByECL    map[uint8]map[uint8][]monster.AffectRecord
	monsterItemsByECL      map[uint8]map[uint8][]monster.ItemRecord
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
	pendingTreasure        []ecl.TreasureRequest
	treasureItemBlocks     map[uint16][]monster.ItemRecord
	pendingTreasureItems   []monster.ItemRecord
	treasureMenu           bool
	treasureTakeMenu       bool
	treasureItemIndex      int
	treasureReturnMode     Mode
	treasureResumeECL      bool
	shopMenu               bool
	shopECLService         bool
	templeMenu             bool
	templeHealMenu         bool
	templeConfirmMenu      bool
	templeECLService       bool
	templeCharacterIndex   int
	templePendingCure      int
	trainingMenu           bool
	trainingConfirmMenu    bool
	trainingSpellMenu      bool
	trainingCharacterIndex int
	trainingSpellChoices   []uint8
	trainingResult         string
	shopOffers             []ShopOffer
	moneyPool              uint32
	treasureGems           uint32
	treasureJewelry        uint32
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
	campECLService         bool
	campReturnMode         Mode
	campRestMenu           bool
	restHours              int
	restEncounterPeriod    uint16
	restEncounterPercent   uint16
	campViewMenu           bool
	campMagicMenu          bool
	campMagicViewMenu      bool
	campMagicMemorizeMenu  bool
	campMagicMemorizeChar  int
	campMagicCastMenu      bool
	campMagicCastChar      int
	campMagicCastSpell     uint8
	pendingMemorizedSpells map[int][]uint8
	saveRequested          bool
	programEndMenu         bool
	gameWon                bool
	partyKilled            bool
	alterMenu              bool
	alterOrderMenu         bool
	alterOrderSelected     int
	alterDropMenu          bool
	alterDropConfirm       bool
	alterDropSelected      int
	alterRenameMenu        bool
	alterRenameChar        int
	renameEditing          bool
	renameCharacter        int
	renameName             string
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

// SetECLSeed controls RANDOM values while replaying an event sequence.
func (s *State) SetECLSeed(seed int64) { s.eclSeed = seed }

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
		s.requestSound(SoundStart)
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
		s.Message = localizeECLText(s.catalog, result.Text)
	}
	if result.PictureRequested {
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
		if result.CombatRequested || result.WaitingForMenu {
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
	s.Choices = make([]string, 0, len(menu.Options))
	s.currentOriginalChoices = append([]string(nil), menu.Options...)
	for _, option := range menu.Options {
		s.Choices = append(s.Choices, localizeOption(s.catalog, option))
	}
	if menu.Prompt != "" {
		s.Prompt = localizePrompt(s.catalog, menu.Prompt)
	} else {
		s.Prompt = s.catalog.Text("press_button", "請按任意鍵或 Enter 繼續")
	}
	s.Mode = ModeWilderness
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
			if blockBefore != currentBlock && s.Area.InDungeon && s.Area.GameArea == 5 &&
				(currentBlock == 0x31 || currentBlock == 0x32 || currentBlock == 0x33) {
				s.syncDungeonStateFromECLRegisters()
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
			// Hap ENTER CITY is an engine-level area transition wrapped
			// around ECL's NEWECL 0x31. The script loads its map pieces, while
			// the DOS dispatcher selects Area 5 and dungeon exploration.
			if originalChoice == "ENTER CITY" && s.session.CurrentBlockID() == 0x31 {
				s.Area.GameArea = 5
				s.Area.InDungeon = true
				s.GeoMapSet = 5
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
			} else if err := s.ResolveTreasureRequests(); err != nil {
				// A headless/test adapter may not have loaded ITEM*.DAX yet.
				// Keep the raw request pending and let the ECL control flow reach
				// its next command (including COMBAT) instead of aborting it.
				s.Message = "財寶等待素材載入：" + err.Error()
			} else if len(s.pendingTreasureItems) > 0 {
				treasureReady = true
			}
		}
		s.applyCitySelection()
		if len(result.Text) > 0 {
			s.unlockJournalEntries(result.Text)
			s.Message = localizeECLText(s.catalog, result.Text)
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
				s.Prompt = localizeECLText(s.catalog, []string{result.WhoRequests[len(result.WhoRequests)-1].Prompt})
			} else {
				s.Prompt = "請選擇角色"
			}
			s.Mode = ModeWilderness
			return nil
		}
		if result.PictureRequested {
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

func (s *State) continueAfterSuppressedPicture(result ecl.RunResult) (bool, error) {
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
	s.treasureMenu = true
	s.treasureTakeMenu = false
	s.treasureReturnMode = returnMode
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("treasure_prompt", "選擇要收下的財寶")
	s.Choices = make([]string, 0, len(s.pendingTreasureItems)+1)
	s.currentOriginalChoices = make([]string, 0, len(s.pendingTreasureItems)+1)
	for index, item := range s.pendingTreasureItems {
		s.Choices = append(s.Choices, monster.ChineseName(item))
		s.currentOriginalChoices = append(s.currentOriginalChoices, "TREASURE_ITEM_"+strconv.Itoa(index))
	}
	s.Choices = append(s.Choices, s.catalog.Text("treasure_exit", "暫不收下／繼續"))
	s.currentOriginalChoices = append(s.currentOriginalChoices, "TREASURE_EXIT")
	s.Message = s.catalog.Text("treasure_ready", "發現財寶。")
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
		gold          uint32
		gems, jewelry uint32
		items         []monster.ItemRecord
	}
	var total resolved
	rng := rand.New(rand.NewSource(s.eclSeed))
	for _, request := range s.pendingTreasure {
		copper := uint64(request.Coins[0]) + uint64(request.Coins[1])*10 + uint64(request.Coins[2])*100 + uint64(request.Coins[3])*200 + uint64(request.Coins[4])*1000
		total.gold += uint32(copper / 200)
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
	s.moneyPool += total.gold
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
	result, err := s.session.RunEntrySeedWithPartyContext(
		2, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
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
	result, err := s.session.RunEntrySeedWithPartyContext(
		3, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
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
	for _, address := range result.CallAddresses {
		s.pendingECLCalls = append(s.pendingECLCalls, address)
		switch address {
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
	s.Prompt = s.catalog.Text("shadowdale_map_prompt", "暗影谷荒野")
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
	result, err := s.session.RunEntrySeedWithPartyContext(
		1, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
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
	s.syncDungeonECLRegisters()
	if exitAttempt {
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
		}
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
	if err := s.session.Switch(blockID); err != nil {
		return err
	}
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
	result, err := s.session.RunEntrySeedWithPartyContext(
		4, 500, nil, nil, s.eclSeed, s.eclPartyContext(),
	)
	if err != nil {
		return err
	}
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
	if s.session != nil && s.session.CurrentBlockID() == 0x02 &&
		s.GeoMapSet == 2 && s.GeoMapBlock == 1 {
		s.DungeonX = (x + 8) % geo.Width
		s.DungeonY = geo.Height - 1 - y
		s.DungeonDirection = (4 - direction + 8) % 8
		return
	}
	s.DungeonX, s.DungeonY, s.DungeonDirection = x, y, direction
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
	if hasMeaningfulECLText(result.Text) {
		s.unlockJournalEntries(result.Text)
		s.Message = localizeECLText(s.catalog, result.Text)
	}
	s.eclMenuReturnMode = ModeDungeon
	s.eventReturnMode = ModeDungeon
	if result.PictureRequested {
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
		if result.CombatRequested || result.ShopRequested || result.TempleRequested || result.WaitingForMenu {
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
		if err := s.ResolveTreasureRequests(); err != nil {
			return true, err
		}
		s.treasureResumeECL = s.session != nil && len(s.eclBlock) > 0
		s.enterTreasureMenuFor(ModeDungeon)
		return true, nil
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
			if result.WaitingForMenu && len(result.Menus) > 0 {
				s.enterECLMenu(result.Menus[len(result.Menus)-1])
				return nil
			}
		}
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

// applyCitySelection prefers the ECL world dispatcher's current/destination
// bytes, then retains the observed opening sequence as a compatibility
// fallback for synthetic sessions.
func (s *State) applyCitySelection() {
	if s.session != nil &&
		(s.session.CurrentBlockID() == 0x50 || s.session.CurrentBlockID() == 0x51) {
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
	s.Prompt = s.catalog.Text("press_button", "Press any button or Enter to continue")
	s.Message = ""
}

func localizeOption(catalog locale.Catalog, option string) string {
	switch option {
	case "PRESS BUTTON OR RETURN TO CONTINUE.":
		return catalog.Text("press_button", "請按任意鍵或 Enter 繼續")
	case "YES":
		return catalog.Text("yes", "是")
	case "NO":
		return catalog.Text("no", "否")
	case "TELL THE TRUTH":
		return catalog.Text("tell_truth", "如實相告")
	case "PUNCH BARKEEP":
		return catalog.Text("tavern_punch", "揍酒保")
	case "HAVE A DRINK":
		return catalog.Text("tavern_drink", "喝一杯")
	case "DRAGON'S BREATH":
		return catalog.Text("tavern_dragon_breath", "龍息酒")
	case "BASILISK":
		return catalog.Text("tavern_basilisk", "石化蜥蜴酒")
	case "LEMONADE":
		return catalog.Text("tavern_lemonade", "檸檬水")
	case "WHISKEY":
		return catalog.Text("tavern_whiskey", "威士忌")
	case "BEER":
		return catalog.Text("tavern_beer", "啤酒")
	case "ALE":
		return catalog.Text("tavern_ale", "愛爾啤酒")
	case "PORT":
		return catalog.Text("tavern_port", "波特酒")
	case "MEAD":
		return catalog.Text("tavern_mead", "蜂蜜酒")
	case "LIE":
		return catalog.Text("lie", "說謊")
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
	case "HALL":
		return catalog.Text("training_hall", "訓練場")
	case "TEMPLE":
		return catalog.Text("temple", "神殿")
	case "RELAX":
		return catalog.Text("relax", "休息並聽取傳聞")
	case "PATROL FOREST":
		return catalog.Text("patrol_forest", "巡查森林")
	case "THANK HIM":
		return catalog.Text("thank_him", "向他道謝")
	case "ATTACK":
		return catalog.Text("attack", "攻擊")
	case "LEAVE", "Leave":
		return catalog.Text("leave", "離開")
	case "EXAMINE CORPSE":
		return catalog.Text("examine_corpse", "檢查屍體")
	case "SHADOWDALE":
		return catalog.Text("shadowdale", "Shadowdale")
	case "ASHABENFORD":
		return catalog.Text("ashabenford", "Ashabenford")
	case "DAGGER FALLS":
		return catalog.Text("dagger_falls", "Dagger Falls")
	case "TILVERTON":
		return catalog.Text("tilverton", "提爾佛頓")
	case "THE STANDING STONE":
		return catalog.Text("standing_stone", "立石群")
	case "ESSEMBRA":
		return catalog.Text("essembra", "艾森布拉")
	case "HAP":
		return catalog.Text("hap", "哈普")
	case "HILLSFAR":
		return catalog.Text("hillsfar", "希爾斯法")
	case "VOONLAR":
		return catalog.Text("voonlar", "沃恩拉")
	case "PHLAN":
		return catalog.Text("phlan", "弗蘭")
	case "TESHWAVE":
		return catalog.Text("teshwave", "泰什浪")
	case "YULASH":
		return catalog.Text("yulash", "尤拉什")
	case "ZHENTIL KEEP":
		return catalog.Text("zhentil_keep", "散提爾堡")
	case "MYTH DRANNOR":
		return catalog.Text("myth_drannor", "迷斯卓諾")
	case "SNEAK IN":
		return catalog.Text("sneak_in", "潛入")
	case "ASK PERMISSION":
		return catalog.Text("ask_permission", "請求許可")
	case "RUN AWAY":
		return catalog.Text("run_away", "逃走")
	case "FIGHT":
		return catalog.Text("fight", "戰鬥")
	case "GO WITH GUARDS":
		return catalog.Text("go_with_guards", "跟衛兵走")
	case "FIGHT THE MEN":
		return catalog.Text("fight_the_men", "攔下他們戰鬥")
	case "LET THEM GO":
		return catalog.Text("let_them_go", "放他們離開")
	case "TELL HER YOUR STORY":
		return catalog.Text("tell_her_your_story", "告訴她你們的經歷")
	case "TELL HER YOU'RE HUNTING CULTISTS":
		return catalog.Text("tell_her_hunting_cultists", "告訴她你們正在追捕邪教徒")
	case "TELL HER IT'S NONE OF HER AFFAIR":
		return catalog.Text("tell_her_none_of_affair", "告訴她這不關她的事")
	case "TRY TO TALK FURTHER":
		return catalog.Text("try_talk_further", "繼續交談")
	case "WILDERNESS":
		return catalog.Text("wilderness", "Wilderness")
	case "CAVES":
		return catalog.Text("caves", "洞穴")
	case "STAY HERE":
		return catalog.Text("stay_here", "留在這裡")
	case "VILLAGE":
		return catalog.Text("village", "村莊")
	case "DEPART":
		return catalog.Text("depart", "離開此區")
	case "TRAIL":
		return catalog.Text("trail", "小徑")
	case "COMBAT":
		return catalog.Text("encounter_combat", "戰鬥")
	case "WAIT":
		return catalog.Text("encounter_wait", "等待")
	case "ENTER THE BLADES":
		return catalog.Text("enter_blades", "闖入刀刃")
	case "RETREAT":
		return catalog.Text("retreat", "撤退")
	case "INTERROGATE":
		return catalog.Text("interrogate", "審問")
	case "KILL":
		return catalog.Text("kill", "殺死")
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
	case "ATTACK DRAGONS":
		return catalog.Text("attack_dragons", "攻擊龍群")
	case "ATTACK WIZARD":
		return catalog.Text("attack_wizard", "攻擊法師")
	case "PARLAY WITH THE DRAGONS":
		return catalog.Text("parlay_with_dragons", "與龍群交涉")
	case "FIRE KNIVES":
		return catalog.Text("fire_knives", "火刀")
	case "PRINCESS NACACIA":
		return catalog.Text("princess_nacacia", "娜卡西亞公主")
	case "NO ONE":
		return catalog.Text("no_one", "不效忠任何人")
	case "EXIT":
		return catalog.Text("exit", "離開")
	default:
		return option
	}
}

func localizePrompt(catalog locale.Catalog, prompt string) string {
	if prompt == "PRESS BUTTON OR RETURN TO CONTINUE." {
		return catalog.Text("press_button", "Press any button or Enter to continue")
	}
	if strings.HasPrefix(prompt, "HOW WILL YOU GET TO ") {
		destination := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(prompt, "HOW WILL YOU GET TO "), "?"))
		return fmt.Sprintf(catalog.Text("route_prompt", "要如何前往%s？"), localizeOption(catalog, destination))
	}
	if prompt == "FROM HERE YOU MAY JOURNEY TO" {
		return catalog.Text("journey_destination_prompt", "從這裡可以前往")
	}
	if prompt == "WHAT WILL YOU DRINK?" {
		return catalog.Text("tavern_drink_prompt", "要喝什麼？")
	}
	if prompt == "A DARK ELF PATROL ARRIVES" {
		return catalog.Text("ecl_hap_dark_elf_patrol", "一隊黑暗精靈巡邏兵出現了")
	}
	return prompt
}

func localizeECLText(catalog locale.Catalog, texts []string) string {
	joined := strings.Join(texts, " ")
	if strings.Contains(joined, "I AM THE HIGH PRIEST") &&
		strings.Contains(joined, "TELL ME YOUR STORY") {
		return catalog.Text(
			"ecl_high_priest_intro",
			"「我是這裡的高階祭司。孩子們，你們看來憂心忡忡，願意告訴我發生了什麼事嗎？」",
		)
	}
	if strings.Contains(joined, "REMOVE CURSE SPELL") &&
		strings.Contains(joined, "JOURNAL ENTRY") && strings.Contains(joined, "19.") {
		return catalog.Text(
			"ecl_high_priest_spell",
			"他同情地聽完你們的遭遇，接著施展移除詛咒法術；你們將結果記入冒險手札第 19 條。",
		)
	}
	switch {
	case strings.Contains(joined, "THIS RUN DOWN VILLAGE IS STRANGELY QUIET") &&
		strings.Contains(joined, "NO ONE IS ABOUT"):
		return catalog.Text(
			"ecl_hap_abandoned_village",
			"這座破敗的村莊異常寂靜。風沿著空蕩街道呼嘯，掠過一扇扇緊閉的窗戶；四下不見人影。",
		)
	case strings.Contains(joined, "YOU BURST IN ON SOME PEASANTS WHO SCUTTLE BACK") &&
		strings.Contains(joined, "LEAVE BEFORE THE HORDE FINDS YOU WITH US"):
		return catalog.Text(
			"ecl_hap_hiding_peasants",
			"你們闖進屋裡，幾名村民驚慌退縮，喊道：「快走！別讓那群怪物發現你們和我們在一起！」你們要怎麼做？",
		)
	case strings.Contains(joined, "THE CRINGING PEASANTS FLEE OUT INTO THE STREET"):
		return catalog.Text("ecl_hap_peasants_flee", "畏縮的村民奪門而出，逃進街道。")
	case strings.Contains(joined, "YOU UNGRATEFUL SLIME") &&
		strings.Contains(joined, "BE HAPPY FOR A QUICK DEATH"):
		return catalog.Text(
			"ecl_hap_dark_elf_attack",
			"「忘恩負義的渣滓！你們一再挑戰我們的耐性，就慶幸自己能死得痛快吧！」",
		)
	case strings.Contains(joined, "THIS BARN IS EMPTY") &&
		strings.Contains(joined, "EFREET AND HIS DARK ELFIN COHORTS"):
		return catalog.Text(
			"ecl_hap_efreet_barn",
			"穀倉裡空無一物，只剩伊弗利特與他的黑暗精靈黨羽。",
		)
	case strings.Contains(joined, "THE EFREET VOICE BOOMS OUT") &&
		strings.Contains(joined, "DOOM ON YOUR VILLAGE"):
		return catalog.Text(
			"ecl_hap_efreet_threat",
			"伊弗利特隆隆吼道：「可悲的蟲子竟敢反抗！我們會先殺了你們，再燒毀這片破爛村落。是你們把毀滅帶給了自己的村莊！」",
		)
	case strings.Contains(joined, "ON THE BODY OF THE EFREET IS A MAP") &&
		strings.Contains(joined, "THE TOWN AND A CAVE"):
		return catalog.Text(
			"ecl_hap_efreet_map",
			"你們在伊弗利特的屍體上找到一張地圖，標出了村莊與一處洞穴。",
		)
	case strings.Contains(joined, "A SHORT TIME AFTER THE SOUNDS OF BATTLE FADE") &&
		strings.Contains(joined, "LOUD CHEERS AND LAUGHTER"):
		return catalog.Text(
			"ecl_hap_liberated_crowd",
			"戰聲平息不久，幾顆膽怯的腦袋探進穀倉。人群很快聚集起來，整座村莊隨即充滿歡呼與笑聲。",
		)
	case strings.Contains(joined, "AN ELDER OF THE VILLAGE COMES FORWARD") &&
		strings.Contains(joined, "ALWAYS BE WELCOME IN HAPTOOTH"):
		return catalog.Text(
			"ecl_hap_elder_thanks",
			"一位村中長老上前說：「我們永遠感激你們；哈普圖斯永遠歡迎各位。」",
		)
	case strings.Contains(joined, "THE ELDER LOWERS HIS VOICE") &&
		strings.Contains(joined, "CONTROLLED FROM THE WIZARD'S TOWER NEARBY"):
		return catalog.Text(
			"ecl_hap_elder_wizard_tower",
			"長老壓低聲音：「我不願顯得忘恩負義，但這些精靈受附近法師塔控制。只有摧毀那個巢穴，我們才真正安全。」",
		)
	case strings.Contains(joined, "AKABAR MENTIONS THAT HE HAS HEARD OF SECRET TRADE ROUTES") &&
		strings.Contains(joined, "HAPPY TO GUIDE THE PARTY THERE"):
		return catalog.Text(
			"ecl_hap_akabar_secret_routes",
			"阿卡巴提到，他聽說有祕密商路可以繞過法師塔；他很樂意帶領隊伍前往。",
		)
	case strings.Contains(joined, "YOU ARE HEADING BACK TO THE WILDERNESS") &&
		strings.Contains(joined, "WANT TO CONTINUE"):
		return catalog.Text("ecl_hap_leave", "你們正準備返回荒野。要繼續嗎？")
	case strings.Contains(joined, "FOLLOW THE MAP TO THE CAVES") &&
		strings.Contains(joined, "GO INTO THE WILDERNESS"):
		return catalog.Text("ecl_hap_map_route", "要循著地圖前往洞穴，還是回到荒野？")
	case strings.Contains(joined, "YOU HAVE ENTERED AN ANCIENT LAVA TUBE") &&
		strings.Contains(joined, "ASH COVERS THE FLOOR"):
		return catalog.Text("ecl_lava_tube_entry", "你們進入一條古老的熔岩隧道，地面覆滿火山灰。")
	case strings.Contains(joined, "FROM HIDDEN ALCOVES COMES A WAVE OF HEAT") &&
		strings.Contains(joined, "SALAMANDERS AND DARK ELVES"):
		return catalog.Text(
			"ecl_lava_tube_ambush",
			"熱浪從隱蔽凹室中湧出，緊接著火蜥蜴與黑暗精靈一擁而上。",
		)
	case strings.Contains(joined, "THE DOOR IS GUARDED BY A SALAMANDER LED PATROL"):
		return catalog.Text("ecl_lava_tube_guarded_door", "門前由一支火蜥蜴率領的巡邏隊把守。")
	case strings.Contains(joined, "A DREAM-LIKE VOICE IN YOUR HEAD SAYS") &&
		strings.Contains(joined, "BE FULLY PREPARED"):
		return catalog.Text(
			"ecl_lava_tube_dream_warning",
			"一個夢境般的聲音在腦中響起：「前方危機重重，務必做好萬全準備！」",
		)
	case strings.Contains(joined, "THE ROOM IS FILLED WITH ACTIVE GEYSERS AND LAVA PITS") &&
		strings.Contains(joined, "SALAMANDERS ARE SPORTING IN THE POOLS"):
		return catalog.Text(
			"ecl_lava_tube_salamander_pools",
			"房內遍布活躍的間歇泉與熔岩池，火蜥蜴正在池中嬉戲。",
		)
	case strings.Contains(joined, "INTENSE HEAT WASHES OVER YOU"):
		return catalog.Text("ecl_lava_tube_intense_heat", "灼熱的熱浪席捲全隊。")
	case strings.Contains(joined, "WE HAVE NO LOVE FOR DARK ELVES") &&
		strings.Contains(joined, "TAKE ANY TREASURE"):
		return catalog.Text(
			"ecl_lava_tube_sly_parlay",
			"「我們對黑暗精靈毫無好感。你們想拿多少財寶，就拿多少吧。」",
		)
	case strings.Contains(joined, "YOU COLD THINGS SHOULD LEAVE") &&
		strings.Contains(joined, "CRIMDRAC FINDS YOU"):
		return catalog.Text(
			"ecl_lava_tube_nice_parlay",
			"「你們這些冰冷生物，最好趁克林德拉克還沒發現前離開。」",
		)
	case strings.Contains(joined, "AMONGST THE POOLS OF LAVA") &&
		strings.Contains(joined, "SIX FIREPROOF CASKS") &&
		strings.Contains(joined, "OPEN ONE"):
		return catalog.Text(
			"ecl_lava_tube_fireproof_casks",
			"熔岩池之間放著六只防火桶。要派人前去打開其中一只嗎？",
		)
	case strings.Contains(joined, "THE HEAT IS TOO INTENSE") &&
		strings.Contains(joined, "DOES ANYONE WANT TO TRY AGAIN"):
		return catalog.Text(
			"ecl_lava_tube_cask_heat_retreat",
			"熱度過於猛烈，他只得退回來。要換另一個人再試一次嗎？",
		)
	case strings.Contains(joined, "HEADING UP INTO THE WIZARD'S TOWER") ||
		(strings.Contains(joined, "COURTYARD OF A FIVE") &&
			strings.Contains(joined, "SURROUNDING THE TOWER ARE HIGH MOUNTAINS")):
		return catalog.Text(
			"ecl_wizard_tower_courtyard",
			"你們走入法師塔，來到一座五層高塔的庭院。塔身受魔法保護，石工無瑕而美麗，四周盡是高山。",
		)
	case strings.Contains(joined, "AN IMPRESSIVE ROBED FIGURE APPROACHES YOU") &&
		strings.Contains(joined, "I AM DRACANDROS"):
		return catalog.Text(
			"ecl_dracandros_arrival",
			"一名氣勢逼人的長袍人物走近：「我是德拉坎德羅斯。很高興你們終於到了。時間緊迫，你們必須扮演好自己的角色。」",
		)
	case strings.Contains(joined, "FREEZE WHERE YOU STAND") &&
		strings.Contains(joined, "THE BONDS PARALYZE YOU"):
		return catalog.Text(
			"ecl_dracandros_freezes_party",
			"「站住！我沒時間陪你們胡鬧！」枷印發出力量，令你們全身動彈不得。",
		)
	case strings.Contains(joined, "ROOF OF THE TOWER") &&
		strings.Contains(joined, "HUGE HOST OF BLACK DRAGONS"):
		return catalog.Text("ecl_wizard_tower_dragon_roof", "轉瞬之間，你們已站在塔頂，四周聚集著一大群黑龍。")
	case strings.Contains(joined, "ONE OF THE DRAGONS DISENGAGES HIMSELF"):
		return catalog.Text("ecl_wizard_tower_dragon_steps_out", "其中一隻黑龍離開龍群，獨自走上前來。")
	case strings.Contains(joined, "ATTACK THE DRAGON AS ELMINSTER TOLD YOU"):
		return catalog.Text("ecl_dracandros_attack_order", "德拉坎德羅斯喝令：「照伊爾明斯特的命令，攻擊那條龍！」")
	case strings.Contains(joined, "UNDER THE FORCE OF THE BONDS") &&
		strings.Contains(joined, "DRAGON WAS ONLY AN ILLUSION"):
		return catalog.Text(
			"ecl_dracandros_dragon_illusion",
			"枷印強迫你們衝上前攻擊；那條龍挨了一擊便化為煙霧——它只是一道幻象！",
		)
	case strings.Contains(joined, "FREEZE, BASE SLAYERS OF DRAGONKIND") &&
		strings.Contains(joined, "JOURNAL ENTRY 15"):
		return catalog.Text(
			"ecl_dracandros_journal_15",
			"「站住，你們這些卑劣的屠龍者！」德拉坎德羅斯再次令你們動彈不得，轉身向龍群發表一番演說；內容已記入手札條目 15。",
		)
	case strings.Contains(joined, "DRACANDROS' MUMBLED PHRASE") &&
		strings.Contains(joined, "BONDS TO") &&
		strings.Contains(joined, "FADE"):
		return catalog.Text(
			"ecl_dracandros_bond_fades",
			"德拉坎德羅斯含糊念出一段咒語，你們身上的一枚枷印逐漸消退。",
		)
	case strings.Contains(joined, "THIS IS A MATTER BETWEEN MEN") &&
		strings.Contains(joined, "WE LEAVE YOU TO YOUR SQUABBLES"):
		return catalog.Text(
			"ecl_wizard_tower_dragons_depart",
			"黑龍說：「這是人類之間的爭端，我們把你們留給自己的內鬥。」龍群隨即振翅飛離。",
		)
	case strings.Contains(joined, "TROOPS DEFEND ME") &&
		strings.Contains(joined, "DRACANDROS FLEES DOWN THE STAIRS"):
		return catalog.Text(
			"ecl_dracandros_calls_troops",
			"德拉坎德羅斯高喊：「部隊，保護我！」一支巡邏隊衝上前來，他則趁機逃下樓梯。",
		)
	case strings.Contains(joined, "HOLD THE ROOF WELL ENOUGH") &&
		strings.Contains(joined, "REST SAFELY"):
		return catalog.Text(
			"ecl_wizard_tower_safe_roof",
			"看來你們足以守住屋頂，可以在這裡安全休息。",
		)
	case strings.Contains(joined, "YOU HAVE CONVINCED US") &&
		strings.Contains(joined, "NO PLOT AGAINST") &&
		strings.Contains(joined, "DISPUTE WITH DRACANDROS"):
		return catalog.Text(
			"ecl_wizard_tower_dragons_convinced",
			"黑龍說：「你們已使我們相信，這裡並沒有對付龍族的陰謀。我們現在離開，讓你們自行解決與德拉坎德羅斯的爭端。」",
		)
	case strings.Contains(joined, "YOU ARE RIGHT DRACANDROS") &&
		strings.Contains(joined, "THEY CONDEMN THEMSELVES"):
		return catalog.Text(
			"ecl_wizard_tower_dragons_condemn",
			"黑龍說：「德拉坎德羅斯，你說得對。他們已自行定罪。」",
		)
	case strings.Contains(joined, "DRACANDROS ESCAPED DOWNSTAIRS") &&
		strings.Contains(joined, "DRAGON BODIES LIE STREWN ABOUT") &&
		strings.Contains(joined, "DO YOU TAKE ONE OF THEIR HEARTS"):
		return catalog.Text(
			"ecl_wizard_tower_take_dragon_heart",
			"戰鬥期間，德拉坎德羅斯逃下樓梯；黑龍的屍體散落在屋頂四周。要取走其中一顆龍心嗎？",
		)
	case strings.Contains(joined, "DRACANDROS ESCAPED DOWNSTAIRS") &&
		strings.Contains(joined, "DRAGON BODIES LIE STREWN ABOUT"):
		return catalog.Text(
			"ecl_wizard_tower_dragon_bodies",
			"戰鬥期間，德拉坎德羅斯逃下樓梯；黑龍的屍體散落在屋頂四周。",
		)
	case strings.Contains(joined, "CUT INTO THE DRAGON") &&
		strings.Contains(joined, "SPRAY OF ACID") &&
		strings.Contains(joined, "EXTRACT THE HEART"):
		return catalog.Text(
			"ecl_wizard_tower_dragon_heart_acid",
			"你們剖開黑龍、取出內臟時，被噴濺的酸液淋得滿身，但仍成功取出了龍心。",
		)
	case strings.Contains(joined, "STOP BY HAPTOOTH VILLAGE") &&
		strings.Contains(joined, "DEPART THE AREA"):
		return catalog.Text(
			"ecl_wizard_tower_wilderness_exit",
			"要先繞到哈普圖斯村，還是直接離開這一帶？",
		)
	case strings.Contains(joined, "YOUR HELP WAS INVALUABLE TO ME") &&
		strings.Contains(joined, "BUSINESS TO ATTEND TO"):
		return catalog.Text(
			"ecl_area5_depart_akabar",
			"阿卡巴向你們道別，隨後離隊處理自己的事務。",
		)
	case strings.Contains(joined, "DARK ELF") &&
		strings.Contains(joined, "DECAY TO USELESSNESS"):
		return catalog.Text(
			"ecl_area5_depart_dark_elf_decay",
			"日光使黑暗精靈的武器與護甲腐朽失效。",
		)
	case strings.Contains(joined, "OUT OF A COPSE OF TREES COMES A SKELETAL") &&
		strings.Contains(joined, "YOU HAVE DEPRIVED ME OF MY TUTOR") &&
		strings.Contains(joined, "I CAN AVENGE MYSELF"):
		return catalog.Text(
			"ecl_post_wizard_dracolich",
			"一具骷髏身影從樹叢現身，誓言為導師復仇。",
		)
	case strings.Contains(joined, "WAY DOWN TO THE CAVES") &&
		strings.Contains(joined, "SECRET PASSAGE") &&
		strings.Contains(joined, "DIRECTLY TO THE WILDERNESS"):
		return catalog.Text(
			"ecl_wizard_tower_roof_exit",
			"這條路可下到洞穴；你們也發現一條能直達荒野的祕道。要走哪一條路？",
		)
	case strings.Contains(joined, "I AM AKABAR BEL AKASH") &&
		strings.Contains(joined, "WILL YOU LET HIM JOIN YOUR PARTY"):
		return catalog.Text(
			"ecl_hap_akabar_join",
			"「你們終於來了。我是阿卡巴・貝爾・阿卡什。只要聯手，我們就能粉碎這股黑暗浪潮。」要讓他加入隊伍嗎？",
		)
	case strings.Contains(joined, "A SURLY INNKEEPER COMES UP") &&
		strings.Contains(joined, "DO YOU STAY"):
		return catalog.Text(
			"ecl_hap_inn_before_liberation",
			"一名板著臉的旅店老闆走來：「把門關上！怪物正在外頭。你們想住就住，只要保持低調。」要留下嗎？",
		)
	case strings.Contains(joined, "SAILING ACROSS THE SKY ARE GREAT BLACK SHAPES") &&
		strings.Contains(joined, "FEARSOME BLACK DRAGONS"):
		return catalog.Text(
			"ecl_hap_black_dragons",
			"巨大的黑影在天際翱翔，突然俯衝而下——竟是三隻駭人的黑龍！",
		)
	case strings.Contains(joined, "YOU ARE AT THE EDGE OF ESSEMBRA"):
		return catalog.Text(
			"ecl_essembra_edge",
			"你們來到艾森布拉城外。要進城，還是繼續旅程？",
		)
	case strings.Contains(joined, "YOU ARE IN ESSEMBRA") &&
		strings.Contains(joined, "WHAT PLACE WILL YOU VISIT"):
		return catalog.Text(
			"ecl_essembra_places",
			"你們身在艾森布拉。要前往哪個場所？",
		)
	case strings.Contains(joined, "YOU ARE IN HILLSFAR") &&
		strings.Contains(joined, "WHAT PLACE WILL YOU VISIT"):
		return catalog.Text(
			"ecl_hillsfar_places",
			"你們身在希爾斯法。要前往哪個場所？",
		)
	case strings.Contains(joined, "WELCOME TO THE BRANCHING OAK"):
		return catalog.Text(
			"ecl_essembra_branching_oak",
			"「歡迎光臨枝椏橡樹客棧。」",
		)
	case strings.Contains(joined, "YOU ARE IN A N OUTDOOR BAR") &&
		strings.Contains(joined, "OVERLOOKING THE WOODS"):
		return catalog.Text(
			"ecl_essembra_outdoor_bar",
			"你們來到一座俯瞰林地的露天酒館。要做什麼？",
		)
	case strings.Contains(joined, "YOU ARE IN A DOCKSIDE BAR"):
		return catalog.Text(
			"ecl_hillsfar_dockside_bar",
			"你們來到碼頭邊的酒館。要做什麼？",
		)
	case strings.Contains(joined, "SOME RED PLUMES COME OVER") &&
		strings.Contains(joined, "ORDER YOU TO CLEAN UP THE MESS"):
		return catalog.Text(
			"ecl_hillsfar_red_plumes_spill_drinks",
			"幾名紅羽衛走過來，故意打翻你們的酒，命令你們把髒亂清乾淨。要照辦嗎？",
		)
	case strings.Contains(joined, "YOU ARE APPROACHED BY A RED PLUME PATROL") &&
		strings.Contains(joined, "TATOO BETRAYS YOU AS A ZHENTRIM SPY"):
		return catalog.Text(
			"ecl_yulash_red_plume_patrol",
			"一隊紅羽衛巡邏兵走近。其中一人咆哮：「你們身上的刺青暴露了散塔林間諜的身分！」",
		)
	case strings.Contains(joined, "SMOKE RISES FROM BEHIND THE RUINED WALLS") &&
		strings.Contains(joined, "OF YULASH") &&
		strings.Contains(joined, "HOW DO YOU ENTER"):
		return catalog.Text(
			"ecl_yulash_entry",
			"煙霧從尤拉什殘破的城牆後升起，裡面不斷傳來戰鬥聲。你們要如何進入？",
		)
	case strings.Contains(joined, "JUST BEFORE YOU ENTER A MAN MOUNTED ON A LARGE HORSE") &&
		strings.Contains(joined, "A WOMAN DRESSED IN PURPLE") &&
		strings.Contains(joined, "SORRY"):
		return catalog.Text(
			"ecl_yulash_riders_burst_out",
			"正要進城時，一名騎著高大駿馬的男子突然衝出尤拉什，撞倒隊伍成員。"+
				"駿馬疾馳而過，你們看見一名紫衣女子緊抱在男子背後；兩人迅速遠去，只聽見她高喊：「抱歉！」",
		)
	case strings.Contains(joined, "HALT! A GUARD WARILY COMES OUT OF A CHECKPOINT") &&
		strings.Contains(joined, "OTHER GUARDS GATHER BEHIND HIM"):
		return catalog.Text(
			"ecl_yulash_checkpoint_halt",
			"「站住！」一名衛兵警戒地走出檢查哨，其他紅羽衛也在他身後集結。你們要怎麼做？",
		)
	case strings.Contains(joined, "YOU MUST COME WITH US TO SEE THE COMMANDER"):
		return catalog.Text(
			"ecl_yulash_see_commander",
			"「你們必須跟我們去見指揮官。」要怎麼做？",
		)
	case strings.Contains(joined, "THIS IS THE COMMANDER'S WAITING ROOM") &&
		strings.Contains(joined, "REMAIN HERE UNTIL YOU ARE CALLED"):
		return catalog.Text(
			"ecl_yulash_waiting_room",
			"這裡是指揮官的等候室。衛兵命令你們留在此處，等待傳喚。",
		)
	case strings.Contains(joined, "TROOPS COME BURSTING OUT OF THE COMMANDER'S OFFICE") &&
		strings.Contains(joined, "THEY'RE SPIES FOR ZHENTIL KEEP"):
		return catalog.Text(
			"ecl_yulash_zhentarim_spies",
			"一群人突然從指揮官辦公室衝出來。有人大喊：「攔住他們！他們是散提爾堡的間諜！」你們要怎麼做？",
		)
	case strings.Contains(joined, "YOU HAVE BEEN LED IN TO SEE THE RED PLUME COMMANDER"):
		return catalog.Text(
			"ecl_yulash_led_to_commander",
			"衛兵領你們進去晉見紅羽衛指揮官。",
		)
	case strings.Contains(joined, "THE COMMANDER DEMANDS TO KNOW YOUR BUSINESS IN YULASH") &&
		strings.Contains(joined, "HOW DO YOU RESPOND"):
		return catalog.Text(
			"ecl_yulash_commander_business",
			"指揮官厲聲質問你們來尤拉什有何目的。你們要用什麼態度回答？",
		)
	case strings.Contains(joined, "YOU HAVE PLEASED THE COMMANDER") &&
		strings.Contains(joined, "JOURNAL ENTRY 22"):
		return catalog.Text(
			"ecl_yulash_commander_pleased",
			"你們的表現令指揮官滿意。他的說明已記入冒險手札第 22 條。",
		)
	case strings.Contains(joined, "THE COMMANDER SHOWS YOU OUT THE SIDE DOOR"):
		return catalog.Text(
			"ecl_yulash_commander_side_door",
			"指揮官親自帶你們從側門離開。",
		)
	case strings.Contains(joined, "THE PIT CREATED BY MOANDER") &&
		strings.Contains(joined, "STEP FORWARD TO ENTER THE DARK DEMESNE"):
		return catalog.Text(
			"ecl_yulash_pit_entrance",
			"眼前就是摩安德上次降臨時留下的巨坑。再向前一步，便會進入那片黑暗領域。",
		)
	case strings.Contains(joined, "THE CLERIC SLAMS HIS FIST AGAINST A PROTRUDING ROCK") &&
		strings.Contains(joined, "YOU ARE TRAPPED IN THE PIT OF MOANDER"):
		return catalog.Text(
			"ecl_pit_of_moander_trapped",
			"牧師猛力擊打一塊突出的岩石；你們身後的洞頂隨即崩塌。隊伍已被困在摩安德之坑裡。",
		)
	case strings.Contains(joined, "THE CLERIC GIVES YOU ONE LAST TRIUMPHANT GLARE") &&
		strings.Contains(joined, "COUGHS BLOOD AND DIES AT YOUR FEET"):
		return catalog.Text(
			"ecl_pit_of_moander_cleric_dies",
			"牧師最後得意地瞪了你們一眼，隨即咳出鮮血，倒斃在眾人腳邊。",
		)
	case strings.Contains(joined, "YOU HEAR THE SOUNDS OF BATTLE IN THE DISTANCE") &&
		strings.Contains(joined, "SMELL OF BAKED BREAD"):
		return catalog.Text(
			"ecl_pit_of_moander_ambience",
			"遠處傳來戰鬥聲，空氣中隱約飄著烤麵包的氣味。",
		)
	case strings.Contains(joined, "YOU SEE A FEMALE FIGHTER AND A STRANGE-LOOKING LIZARD MAN") &&
		strings.Contains(joined, "VIOLETS, BRIMSTONE AND HONEYSUCKLE"):
		return catalog.Text(
			"ecl_pit_alias_dragonbait_meet",
			"你們看見一名女戰士與一名外貌奇特的蜥蜴人；紫羅蘭、硫磺與忍冬的強烈氣味接連飄來。",
		)
	case strings.Contains(joined, "THE FEMALE FIGHTER GASPS") &&
		strings.Contains(joined, "THEY'RE BONDED") &&
		strings.Contains(joined, "WHAT DO YOU DO"):
		return catalog.Text(
			"ecl_pit_alias_bonded_reaction",
			"女戰士倒抽一口氣：「他們也被枷印控制了！」你們要怎麼做？",
		)
	case strings.Contains(joined, "THE FIGHTER INTRODUCES HERSELF AS ALIAS") &&
		strings.Contains(joined, "HER COMPANION AS DRAGONBAIT") &&
		strings.Contains(joined, "SHE ASKS YOU TO TELL YOUR STORY"):
		return catalog.Text(
			"ecl_pit_alias_dragonbait_introduction",
			"女戰士自稱愛麗雅絲，並介紹她的同伴龍餌。她說自己過去也有與你們相似的刺青，請眾人說明來歷。",
		)
	case strings.Contains(joined, "SHE TELLS HER STORY") &&
		strings.Contains(joined, "JOURNAL ENTRY 3"):
		return catalog.Text(
			"ecl_pit_alias_story",
			"愛麗雅絲說出自己的過往；你們將故事記入冒險手札第 3 條。",
		)
	case strings.Contains(joined, "DO YOU WANT THEM TO JOIN YOU"):
		return catalog.Text(
			"ecl_pit_alias_dragonbait_join",
			"要讓愛麗雅絲與龍餌加入隊伍嗎？",
		)
	case strings.Contains(joined, "ALIAS AND DRAGONBAIT JOIN YOUR PARTY") &&
		strings.Contains(joined, "TREASURE THAT MOGION") &&
		strings.Contains(joined, "KEEPS BEHIND HER ALTAR"):
		return catalog.Text(
			"ecl_pit_alias_dragonbait_joined",
			"愛麗雅絲與龍餌加入隊伍。愛麗雅絲挖苦地補充：「另外，摩貢大祭司藏在祭壇後方的財寶也值得處理。」",
		)
	case strings.Contains(joined, "YOU SEE STAIRS LEADING DOWN TO THE SOUTH") &&
		strings.Contains(joined, "DO YOU WISH TO GO DOWN"):
		return catalog.Text(
			"ecl_pit_stairs_down",
			"你們看見一道通往南方下層的階梯。要往下走嗎？",
		)
	case strings.Contains(joined, "YOU SEE STAIRS GOING UP IN THE NORTH WALL") &&
		strings.Contains(joined, "DO YOU WISH TO GO UP"):
		return catalog.Text(
			"ecl_pit_stairs_up",
			"北側牆邊有一道向上的階梯。要回到上層嗎？",
		)
	case strings.Contains(joined, "MANGLED REMAINS OF A DEAD ZHENTRIM FIGHTER") &&
		strings.Contains(joined, "WHAT DO YOU DO"):
		return catalog.Text(
			"ecl_pit_dead_zhentrim",
			"你們看見一名死去的散塔林戰士，遺體已殘破不堪。要怎麼做？",
		)
	case strings.Contains(joined, "GRASPED IN THE FIGHTER'S FIST") &&
		strings.Contains(joined, "SEAL OF ZHENTIL") &&
		strings.Contains(joined, "JOURNAL ENTRY 46"):
		return catalog.Text(
			"ecl_pit_zhentrim_scroll",
			"戰士緊握的拳中有一卷正式文書，上面蓋著散提爾堡的印璽。你們將內容抄錄為冒險手札第 46 條。",
		)
	case strings.Contains(joined, "YOU SEE A PRIESTESS TURN AND SMILE WICKEDLY") &&
		strings.Contains(joined, "CULTISTS CHANTING IN A LOW DRONE"):
		return catalog.Text(
			"ecl_pit_mogion_altar",
			"你們看見一名女祭司轉身陰險地微笑。她站在祭壇前，四周的摩安德教徒正低聲吟唱。",
		)
	case strings.Contains(joined, "ALIAS MUTTERS") &&
		strings.Contains(joined, "PRIESTESS OF MOANDER") &&
		strings.Contains(joined, "SPITS ON THE GROUND"):
		return catalog.Text(
			"ecl_pit_alias_identifies_mogion",
			"愛麗雅絲低聲說：「她就是摩安德的大祭司。」女祭司轉頭朝地上啐了一口。",
		)
	case strings.Contains(joined, "MOGION SAYS") &&
		strings.Contains(joined, "PROPER TOOLS") &&
		strings.Contains(joined, "WHAT DO YOU DO"):
		return catalog.Text(
			"ecl_pit_mogion_greeting",
			"摩貢說：「真高興你們終於到了。沒有合適的工具，想完成任何大事都很困難，不是嗎？」你們要怎麼做？",
		)
	case strings.Contains(joined, "BEFORE YOU CAN ACT") &&
		strings.Contains(joined, "BLUE FLASH") &&
		strings.Contains(joined, "YOU CANNOT MOVE"):
		return catalog.Text(
			"ecl_pit_moander_bond_paralysis",
			"你們尚未來得及行動，手臂上的枷印便迸出藍光，將全隊籠罩其中；所有人頓時動彈不得。",
		)
	case strings.Contains(joined, "TENDRILS COME UP FROM THE FLOOR") &&
		strings.Contains(joined, "ALIAS AND DRAGONBAIT"):
		return catalog.Text(
			"ecl_pit_alias_dragonbait_tendrils",
			"藤蔓從地面竄出，緊緊纏住愛麗雅絲與龍餌。",
		)
	case strings.Contains(joined, "MOGION TURNS TO THE ALTAR") &&
		strings.Contains(joined, "CHANTING RISES"):
		return catalog.Text(
			"ecl_pit_mogion_ritual",
			"摩貢轉向祭壇，四周吟唱聲逐漸高昂。",
		)
	case strings.Contains(joined, "BLUE LIGHT THAT SURROUNDS YOU") &&
		strings.Contains(joined, "DIMENSIONAL WINDOW ABOVE THE ALTAR"):
		return catalog.Text(
			"ecl_pit_moander_dimensional_window",
			"籠罩你們的藍光流向摩貢；枷印抽出的力量在祭壇上方撕開一道異次元窗口。",
		)
	case strings.Contains(joined, "MOGION SHRIEKS") &&
		strings.Contains(joined, "MOANDER RETURNS") &&
		strings.Contains(joined, "DIMENSIONAL RIFT"):
		return catalog.Text(
			"ecl_pit_moander_returns",
			"摩貢尖叫：「摩安德回來了！」一團由黏液、黴菌與腐敗穢物構成的噁心物質開始從異次元裂隙滲出。",
		)
	case strings.Contains(joined, "ENERGY IN THE DIMENSIONAL RIFT INCREASES") &&
		strings.Contains(joined, "BOND OF MOANDER BEGIN TO FADE"):
		return catalog.Text(
			"ecl_pit_moander_bond_fades",
			"裂隙中的能量持續增強，摩安德枷印開始灼燒；隨著開口擴大，枷印也逐漸褪去。",
		)
	case strings.Contains(joined, "THE SIGIL DISAPPEARS") &&
		strings.Contains(joined, "PARALYSIS THAT GRIPPED YOU IS NOW GONE"):
		return catalog.Text(
			"ecl_pit_moander_bond_broken",
			"摩安德枷印消失了！束縛全隊的麻痺也隨之解除。",
		)
	case strings.Contains(joined, "ALIAS AND DRAGONBAIT HAVE HACKED THEIR WAY FREE") &&
		strings.Contains(joined, "UNLESS YOU WISH TO FIGHT A GOD"):
		return catalog.Text(
			"ecl_pit_alias_attack_mogion",
			"愛麗雅絲與龍餌已砍斷藤蔓脫困。她嘶聲喊道：「現在就攻擊他們，除非你們想直接面對一位神！」你們要怎麼做？",
		)
	case strings.Contains(joined, "THE DIMENSIONAL RIFT SNAPS SHUT"):
		return catalog.Text(
			"ecl_pit_moander_rift_closes",
			"異次元裂隙猛然閉合。",
		)
	case strings.Contains(joined, "THREE PSUEDOPODS OF MOANDER") &&
		strings.Contains(joined, "HUNDREDS OF MOUTHS") &&
		strings.Contains(joined, "YOU HAVE KILLED ME"):
		return catalog.Text(
			"ecl_pit_moander_remnants_scream",
			"三塊已穿過裂隙的摩安德殘軀突然長出數百張嘴，齊聲尖叫：「你們殺了我！」",
		)
	case strings.Contains(joined, "THE OOZING MOUNDS TURN AND ATTACK YOU"):
		return catalog.Text(
			"ecl_pit_moander_remnants_attack",
			"滲著黏液的殘軀轉身向你們撲來！",
		)
	case strings.Contains(joined, "YOU FIND THE GAUNTLET OF MOANDER") &&
		strings.Contains(joined, "SLIMY REMAINS"):
		return catalog.Text(
			"ecl_pit_gauntlet_of_moander",
			"你們在黏滑的殘骸中找到摩安德護手。",
		)
	case strings.Contains(joined, "A PRIEST BURSTS INTO THE ROOM") &&
		strings.Contains(joined, "THEY HAVE KILLED THE GOD"):
		return catalog.Text(
			"ecl_pit_priest_flees_after_moander",
			"一名祭司衝進房間，驚恐地環顧四周，隨即逃回走廊並高喊：「他們殺了神！」",
		)
	case strings.Contains(joined, "YOU HAVE FOUND A CACHE OF JEWELS AND GEMS"):
		return catalog.Text(
			"ecl_pit_altar_jewels_gems",
			"你們在祭壇中找到一批珠寶與寶石！",
		)
	case strings.Contains(joined, "YOU HAVE ALSO FOUND A MAP OF THE TEMPLE") &&
		strings.Contains(joined, "JOURNAL ENTRY 20"):
		return catalog.Text(
			"ecl_pit_temple_map_journal_20",
			"你們還找到一張神殿地圖，並將內容記為冒險手札第 20 條。",
		)
	case strings.Contains(joined, "A HOODED, GREY ROBED MAN SITS IN A DARK CORNER") &&
		strings.Contains(joined, "MOTIONS YOU OVER"):
		return catalog.Text(
			"ecl_shadowdale_hooded_man",
			"一名兜帽低垂的灰袍男子坐在昏暗角落。他招手示意你們靠近，低聲說了一番話，隨後便消失無蹤。",
		)
	case strings.Contains(joined, "WHAT WILL YOU DRINK"):
		return catalog.Text("tavern_drink_prompt", "要喝什麼？")
	case strings.Contains(joined, "YOU OVERHEAR TAVERN TALE 44"):
		return catalog.Text(
			"tavern_tale_44",
			"酒客說：「紅袍法師喜愛火焰生物；面對他們時，寒冷攻擊往往是最好的防禦。」",
		)
	case strings.Contains(joined, "YOU OVERHEAR TAVERN TALE 60"):
		return catalog.Text(
			"tavern_tale_60",
			"酒客說：「有幾道龐大身影飛越森林，朝南方去了。」",
		)
	case strings.Contains(joined, "YOU ARE AT THE EDGE OF HAP"):
		return catalog.Text(
			"ecl_hap_edge",
			"你們來到哈普村外。要進村，還是繼續旅程？",
		)
	case strings.Contains(joined, "YOU ARE AT THE EDGE OF HILLSFAR"):
		return catalog.Text(
			"ecl_hillsfar_edge",
			"你們來到希爾斯法城外。要進城，還是繼續旅程？",
		)
	case strings.Contains(joined, "YOU ARE AT THE EDGE OF YULASH"):
		return catalog.Text(
			"ecl_yulash_edge",
			"你們來到尤拉什城外。要進入戰區，還是繼續旅程？",
		)
	case strings.Contains(joined, "HOW WILL YOU GET TO ESSEMBRA"):
		return catalog.Text("ecl_route_essembra", "要如何前往艾森布拉？")
	case strings.Contains(joined, "HOW WILL YOU GET TO HAP"):
		return catalog.Text("ecl_route_hap", "要如何前往哈普？")
	case strings.Contains(joined, "SEEK RED TO THE SOUTH"):
		return catalog.Text(
			"ecl_standing_stone_seek_red",
			"他繼續說：「往南方尋找紅色之人。」",
		)
	case strings.Contains(joined, "YOU PRESENTLY SERVE FOUR MASTER") &&
		strings.Contains(joined, "RETURN TO ME WHEN YOU HAVE SLAIN THREE MORE"):
		return catalog.Text(
			"ecl_standing_stone_four_masters",
			"他開口說：「你們目前仍服侍四位主人。再除掉其中三位後回來找我，屆時你們將成就自己的命運。」你們要怎麼做？",
		)
	case strings.Contains(joined, "YOU ARE AT THE STANDING STONES") &&
		strings.Contains(joined, "GREY ROBED MAN"):
		return catalog.Text(
			"ecl_standing_stone_grey_man",
			"你們來到立石群。一名灰袍男子靠著巨石而坐，臉孔隱沒在陰影裡。",
		)
	case strings.Contains(joined, "AMBUSHED BY FIRE KNIVES DISGUISED AS A PATROL"):
		return catalog.Text(
			"ecl_shadow_gap_fire_knife_patrol",
			"一隊偽裝成巡邏兵的火刀突然伏擊你們！",
		)
	case strings.Contains(joined, "AMBUSHED BY FIRE KNIVES DISGUISED AS FIGHTERS"):
		return catalog.Text(
			"ecl_fire_knives_disguised_fighters",
			"一群偽裝成戰士的火刀突然伏擊你們！",
		)
	case strings.Contains(joined, "YOU OVERHEAR TAVERN TALE 28"):
		return catalog.Text(
			"tavern_tale_28",
			"你們無意間聽見一則酒館傳聞：已有兩艘前往暗影谷的船失蹤，河道變得非常危險。",
		)
	case strings.Contains(joined, "YOU ARE IN A RIVERSIDE ALE HOUSE") &&
		strings.Contains(joined, "WHAT WILL YOU DO"):
		return catalog.Text(
			"ecl_ashabenford_ale_house",
			"你們身在河畔酒館。要做什麼？",
		)
	case strings.Contains(joined, "YOU ARE IN ASHABENFORD") &&
		strings.Contains(joined, "WHAT PLACE WILL YOU VISIT"):
		return catalog.Text(
			"ecl_ashabenford_places",
			"你們身在阿沙本福德。要前往哪個場所？",
		)
	case strings.Contains(joined, "YOU ARE AT THE EDGE OF ASHABENFORD"):
		return catalog.Text(
			"ecl_ashabenford_edge",
			"你們來到阿沙本福德城外。要進城，還是繼續旅程？",
		)
	case strings.Contains(joined, "YOU ARE AT THE EDGE OF TILVERTON"):
		return catalog.Text(
			"ecl_tilverton_edge",
			"你們來到提爾佛頓城外。要進城，還是繼續旅程？",
		)
	case strings.Contains(joined, "GUARDS BAR YOUR WAY"):
		return catalog.Text("ecl_tilverton_barred", "衛兵擋住你們的去路，不准你們再進入提爾佛頓。")
	case strings.Contains(joined, "MOUNTAINS RISE INTO AN IMPASSABLE WALL") &&
		strings.Contains(joined, "TILVER'S GAP") &&
		strings.Contains(joined, "FLYING SHAPES"):
		return catalog.Text(
			"ecl_tilvers_gap_flying_shapes",
			"群山拔地而起，形成無法翻越的高牆，只有提爾隘口將其劃破。幾道飛行身影從積雪山峰盤旋俯衝而下。",
		)
	case strings.Contains(joined, "BEFORE YOU STANDS A BURLY MAN") &&
		strings.Contains(joined, "CARE TO REST"):
		return catalog.Text(
			"ecl_guildmaster_greeting",
			"一名魁梧男子站在你們面前，身旁圍著數名警戒的護衛。他顯得有些緊張，問道：「你們看來不太穩當，要先休息嗎？」",
		)
	case strings.Contains(joined, "THE FIRE KNIVES HAVE THE KING'S DAUGHTER") &&
		strings.Contains(joined, "I CAN OFFER INFORMATION"):
		return catalog.Text(
			"ecl_guildmaster_briefing",
			"「好，那就繼續。火刀綁架了國王的女兒，把她藏在據點裡，想引國王踏入陷阱。我不能直接介入，但可以提供情報。」",
		)
	case strings.Contains(joined, "SIDE DOOR EXPLODES INWARD") &&
		strings.Contains(joined, "DEAFENING CRASH"):
		return catalog.Text("ecl_guild_breach", "突然間，側門伴隨震耳欲聾的巨響向內炸開！")
	case strings.Contains(joined, "TRAITOROUS SCUM") &&
		strings.Contains(joined, "SEIZE THEM ALL"):
		return catalog.Text(
			"ecl_guild_fire_knife_command",
			"一名火刀嘶聲喝道：「叛徒渣滓！把他們全抓起來！」大批盜賊隨即散開包圍眾人。",
		)
	case strings.Contains(joined, "GUILDMASTER HURLS A POISONED DAGGER") &&
		strings.Contains(joined, "TWITCHING VIOLENTLY"):
		return catalog.Text(
			"ecl_guild_poisoned_dagger",
			"公會首領擲出淬毒匕首，正中一名火刀的咽喉；那人癱倒在地，劇烈抽搐。",
		)
	case strings.Contains(joined, "ARROW HIT THE GUILDMASTER IN THE CHEST") &&
		strings.Contains(joined, "THE BATTLE IS JOINED"):
		return catalog.Text(
			"ecl_guild_battle_joined",
			"你們準備迎戰時，一支箭射中公會首領胸口。四名忠於公會的盜賊與你們並肩作戰，混戰就此爆發！",
		)
	case strings.Contains(joined, "THE GUILDMASTER GASPS") &&
		strings.Contains(joined, "JOURNAL ENTRY 4"):
		return catalog.Text(
			"ecl_guildmaster_death",
			"公會首領奄奄一息地喘道：「權衡之下，我寧可待在尤拉什……」隨即死去。你們在他身上找到一幅通往火刀據點的下水道地圖，記入冒險手札第 4 條。",
		)
	case strings.Contains(joined, "HALFLING WITH A HARP") &&
		strings.Contains(joined, "DISAPPEAR"):
		return catalog.Text(
			"ecl_guild_halfling",
			"走廊盡頭，你們看見一名抱著豎琴的半身人閃進門口，隨即消失。",
		)
	case strings.Contains(joined, "HUNGRY SNARLS") &&
		strings.Contains(joined, "RELEASES THE PACK"):
		return catalog.Text(
			"ecl_guild_kennel_intro",
			"你們一進門便聽見飢餓的低吼。一名火刀放出了犬群！",
		)
	case strings.Contains(joined, "GNAWED BONES") &&
		strings.Contains(joined, "LEASHES"):
		return catalog.Text("ecl_guild_kennel_aftermath", "房裡散落著被啃咬的骨頭與斷裂的皮帶。")
	case strings.Contains(joined, "CAGES THAT ONCE HELD MONKEYS"):
		return catalog.Text("ecl_guild_monkey_cages", "你們看見幾座曾用來關猴子的空籠。")
	case strings.Contains(joined, "OPEN GUEST BOOK") &&
		strings.Contains(joined, "O.RUSKETTLE"):
		return catalog.Text(
			"ecl_guild_guest_book",
			"桌上攤著一本訪客簿，最後一筆寫著：「奧莉芙・拉斯凱托，國度吟遊詩人——碰豎琴者，小心你的手。」",
		)
	case strings.Contains(joined, "GREEN SLIMY MARKS") &&
		strings.Contains(joined, "MORE DISTINCT NEAR THE DOOR"):
		return catalog.Text("ecl_guild_sewer_traces", "地面留有綠色黏液痕跡，越靠近門越清晰。")
	case strings.Contains(joined, "FOUL SMELLING, SLIME COVERED") &&
		strings.Contains(joined, "FIGHTING WILL BE") &&
		strings.Contains(joined, "AWKWARD"):
		return catalog.Text(
			"ecl_tilverton_sewers_entry",
			"你們進入提爾佛頓惡臭撲鼻、覆滿黏液的下水道。地面濕滑、天花板低矮，顯然很難在這裡靈活作戰。",
		)
	case strings.Contains(joined, "YOU ARE ENTERING THE HIDEOUT"):
		return catalog.Text("ecl_fire_knife_hideout_entry", "你們進入了火刀據點。")
	case strings.Contains(joined, "BLADES SLOW DOWN") &&
		strings.Contains(joined, "FADE AWAY"):
		return catalog.Text(
			"ecl_fire_knife_blade_barrier_fades",
			"片刻後，刀刃逐漸放慢並消失；嗡鳴降成低語，最後完全止息。房裡看來已有其他人遭刀刃捲入。",
		)
	case strings.Contains(joined, "CLOUD OF BLADES WHIRLING") &&
		strings.Contains(joined, "METALLIC WHINE"):
		return catalog.Text(
			"ecl_fire_knife_blade_barrier",
			"你們在房間入口停下。前方有一團刀刃彼此盤旋，尖銳的金屬嗡鳴幾乎蓋過所有聲音。你們要怎麼做？",
		)
	case strings.Contains(joined, "THE BLADES TEAR INTO YOU"):
		return catalog.Text("ecl_fire_knife_blades_damage", "旋轉刀刃狠狠撕裂了你們！")
	case strings.Contains(joined, "YOU DISARM THE FIRE KNIVES") &&
		strings.Contains(joined, "JOURNAL ENTRY 26"):
		return catalog.Text(
			"ecl_fire_knife_frozen_interrogate",
			"火刀成員逐漸恢復行動時，你們先繳了他們的械。他們仍神情茫然，你們趁機問出一些有用情報，並記入冒險手札第 26 條。",
		)
	case strings.Contains(joined, "YOU SLAUGHTER THEM") &&
		strings.Contains(joined, "BEING HELD"):
		return catalog.Text("ecl_fire_knife_frozen_kill", "你們趁他們尚未從定身狀態恢復，將他們全數殺死。")
	case strings.Contains(joined, "PEOPLE FROZEN IN") &&
		strings.Contains(joined, "BEGINNING TO MOVE"):
		return catalog.Text(
			"ecl_fire_knife_frozen_room",
			"房裡有許多人凝固在交戰姿勢中；有幾人倒成扭曲的一團，另一些已開始活動。你們要怎麼做？",
		)
	case strings.Contains(joined, "DRAWERS OF A ROSEWOOD DESK") &&
		strings.Contains(joined, "INTERESTING PAPERS"):
		return catalog.Text(
			"ecl_fire_knife_office_search",
			"你們在花梨木書桌的抽屜裡找到一些值得注意的文件，記入冒險手札第 9 條；此外還有其他物品引起你們注意。",
		)
	case strings.Contains(joined, "ORNATE ROOM") &&
		strings.Contains(joined, "HIGH UP IN THE FIRE KNIVES"):
		return catalog.Text(
			"ecl_fire_knife_office",
			"這是一間裝飾華麗的房間，看來是火刀某位高層人物的辦公室。",
		)
	case strings.Contains(joined, "HAND KEPT THE PAPER") &&
		strings.Contains(joined, "JOURNAL ENTRY 29"):
		return catalog.Text(
			"ecl_fire_knife_library_paper",
			"焦屍的手保住了紙張，使它未被燒毀。你們取走紙張，並將內容記入冒險手札第 29 條。",
		)
	case strings.Contains(joined, "STRANGE SMOKY SCENT"):
		return catalog.Text("ecl_fire_knife_smoky_hall", "你們走進走廊時，察覺到一股奇怪的煙味。")
	case strings.Contains(joined, "EXTREMELY WELL ORDERED BEDROOM") &&
		strings.Contains(joined, "UNSEEN SERVANTS"):
		return catalog.Text(
			"ecl_fire_knife_ordered_bedroom",
			"這間臥室整齊得異常，一切都精確歸位。搜索後沒有發現值錢物品；你們離開時，看不見的僕人又開始將房間恢復原狀。",
		)
	case strings.Contains(joined, "ROOM WAS ONCE A LIBRARY") &&
		strings.Contains(joined, "CHARRED BODY"):
		return catalog.Text(
			"ecl_fire_knife_burned_library",
			"這裡原是圖書館，如今書架與藏書都化為灰燼，部分仍冒著煙。房間中央有具焦屍，手中緊握一張紙。",
		)
	case strings.Contains(joined, "ONCE A LAB") &&
		strings.Contains(joined, "NOTHING ESCAPED DESTRUCTION"):
		return catalog.Text(
			"ecl_fire_knife_burned_lab",
			"這裡原是一間實驗室，同樣遭猛烈烈焰席捲，沒有任何東西逃過毀滅。",
		)
	case strings.Contains(joined, "TWO ROWS OF SHROUDED BODIES") &&
		strings.Contains(joined, "TO BE RAISED"):
		return catalog.Text(
			"ecl_fire_knife_shrouded_bodies",
			"房裡有兩排覆著裹屍布的遺體。每排前方各有標牌，一面寫著「待復活」，另一面寫著「待埋葬」。",
		)
	case strings.Contains(joined, "LEADER OF THE FIRE KNIVES") &&
		strings.Contains(joined, "JOURNAL ENTRY") &&
		strings.Contains(joined, "11"):
		return catalog.Text(
			"ecl_fire_knife_leader",
			"你們遇見了火刀首領。他陰險地冷笑著開口；你們將他的話記入冒險手札第 11 條。",
		)
	case strings.Contains(joined, "FIRE KNIVES HAVE BEEN DEFEATED") &&
		strings.Contains(joined, "JOURNAL ENTRY 54"):
		return catalog.Text(
			"ecl_fire_knife_victory",
			"火刀眾已被擊敗。公主以利刃威脅首領；你們將接下來發生的事記入冒險手札第 54 條。",
		)
	case strings.Contains(joined, "FREEING GIOGI") &&
		strings.Contains(joined, "JOURNAL ENTRY 53"):
		return catalog.Text(
			"ecl_fire_knife_royal_arrival",
			"正當你們準備替喬吉鬆綁，整個房間忽然劇烈震動；你們將接下來發生的事記入冒險手札第 53 條。",
		)
	case strings.Contains(joined, "FIRST NIGHT OUTSIDE THE CITY") &&
		strings.Contains(joined, "VIVID DREAM"):
		return catalog.Text(
			"ecl_first_bond_dream",
			"離城後的第一夜，一股怪異倦意籠罩全隊，連守夜者也沉沉睡去。毫無預兆地，你們墜入一場無比鮮明的夢。",
		)
	case strings.Contains(joined, "FOUR FACES LEER DOWN") &&
		strings.Contains(joined, "WEAKEST OF YOUR MASTERS"):
		return catalog.Text(
			"ecl_bond_masters_taunt",
			"四張面孔鄙夷地俯視你們的勝利。陰沉不祥的聲音宣告：「你們第一位、也是最弱的主人已經倒下；如今你們已踏上奴役之路。」",
		)
	case strings.Contains(joined, "WIZARD IN RED") &&
		strings.Contains(joined, "PAWNS OF THE FLAMED ONE"):
		return catalog.Text(
			"ecl_bond_masters_prophecy",
			"「即使意志反抗，你們仍將依序服侍我們：紅袍法師、綠衣女子與黑暗之主。最後，當靈魂之火熄滅，你們將淪為燃燒者的棋子。」",
		)
	case strings.Contains(joined, "FACES LAUGH WITH EVIL JOY") &&
		strings.Contains(joined, "AWAKE IN A COLD SWEAT"):
		return catalog.Text(
			"ecl_bond_dream_ends",
			"四張面孔邪惡地狂笑。夢境逐漸消退，你們渾身冷汗地驚醒。",
		)
	case strings.Contains(joined, "THIS WAY IS CLOSED") &&
		strings.Contains(joined, "ROYAL CARRIAGE IS COMING SOON"):
		return catalog.Text(
			"ecl_carriage_gate_closed",
			"你們遇上一隊皇家衛兵。他們說國王的馬車即將抵達，此路暫時封閉，並把你們遣回城內。",
		)
	case strings.Contains(joined, "MAKE WAY FOR THE ROYAL CARRIAGE"):
		return catalog.Text(
			"ecl_carriage_make_way",
			"你們再次遇上皇家衛兵。他們高喊：「讓路！皇家馬車即將通過！」",
		)
	case strings.Contains(joined, "KING'S VOICE COMING FROM THE CARRIAGE") &&
		strings.Contains(joined, "COMPULSION TO ATTACK"):
		return catalog.Text(
			"ecl_carriage_bond_compulsion",
			"國王的聲音從馬車裡傳來。你們手臂上的青色印記突然發出強光，一股無法抗拒的力量迫使你們攻擊皇家馬車！",
		)
	case strings.Contains(joined, "FIRE KNIVES DEMAND YOUR IMMEDIATE SURRENDER") &&
		strings.Contains(joined, "DO YOU SURRENDER"):
		return catalog.Text(
			"ecl_sewers_fire_knife_checkpoint",
			"火刀要求你們立刻投降。你們要投降嗎？",
		)
	case strings.Contains(joined, "YOU QUICKLY HIDE THEIR BODIES"):
		return catalog.Text("ecl_sewers_hide_checkpoint_bodies", "你們迅速把火刀的屍體藏了起來。")
	case strings.Contains(joined, "SLAUGHTERED REMAINS OF A FIRE KNIFE") &&
		strings.Contains(joined, "KNIGHTS OF MYTH DRANNOR"):
		return catalog.Text(
			"ecl_sewers_knight_appears",
			"這裡躺著一處遭屠滅的火刀檢查哨。你們小心查看時，一名男子從陰影中走出；他手持長劍，身穿迷斯卓諾騎士團的制服。",
		)
	case strings.Contains(joined, "BLUE TATTOO MARKINGS OF THE FIRE KNIVES") &&
		strings.Contains(joined, "TO WHOM DO YOU OWE ALLEGIANCE"):
		return catalog.Text(
			"ecl_sewers_knight_allegiance",
			"「你們身上帶著火刀的青色刺青，但我聽說那個受詛咒的烈焰者正用這類印記控制他人。你們效忠誰？」",
		)
	case strings.Contains(joined, "THAT PRINCESS IS A POPULAR GIRL") &&
		strings.Contains(joined, "DON'T KILL THE CLERIC WITH A HAMMER"):
		return catalog.Text(
			"ecl_sewers_knight_princess_friend",
			"他笑道：「那位公主還真受歡迎！繼續往南走，別殺掉拿著戰鎚的牧師；他也正試著營救她。」說完便讓你們通過。",
		)
	case strings.Contains(joined, "I'M NOT REALLY THE KING") &&
		strings.Contains(joined, "OH NO! NOT AGAIN"):
		return catalog.Text(
			"ecl_carriage_false_king",
			"馬車撤退時，一名年輕男子探出身子，沙啞地喊：「別殺我，我不是真正的國王！」"+
				"他看見你們的青色印記後驚叫：「不！怎麼又來了！」隨即昏倒在車內，印記的光芒也逐漸消退。",
		)
	case strings.Contains(joined, "LOUD BELL STARTS RINGING") &&
		strings.Contains(joined, "SWORDS DRAWN"):
		return catalog.Text(
			"ecl_carriage_alarm",
			"身後響起刺耳警鐘，皇家衛兵拔劍朝你們衝來。",
		)
	case strings.Contains(joined, "TWO RED ROBED MEN JUMP THE CARRIAGE") &&
		strings.Contains(joined, "DRAG HIM INTO AN ALLEYWAY"):
		return catalog.Text(
			"ecl_carriage_abduction",
			"越過衝鋒的衛兵，你們看見兩名紅袍人跳上馬車，把那名瘦弱男子拖出來，拉進一條小巷。",
		)
	case strings.Contains(joined, "DO YOU SURRENDER"):
		return catalog.Text("ecl_carriage_surrender", "一名衛兵喝問：「你們投降嗎？」")
	case strings.Contains(joined, "YOU ARE THROWN IN JAIL"):
		return catalog.Text("ecl_carriage_jailed", "你們被投入牢房。")
	case strings.Contains(joined, "ONE WALL SLIDES OPEN AND A THIEF APPEARS") &&
		strings.Contains(joined, "SIGNALS YOU TO FOLLOW HIM"):
		return catalog.Text(
			"ecl_carriage_thief_rescue",
			"幾個小時後，一面牆悄然滑開，一名盜賊現身。他歸還你們的裝備，示意眾人跟上。",
		)
	case strings.Contains(joined, "LEADS YOU THROUGH HIDDEN PASSAGES") &&
		strings.Contains(joined, "THE THIEVES' GUILD"):
		return catalog.Text(
			"ecl_carriage_guild_arrival",
			"那人領著你們穿過隱密通道，最後來到一處陰暗的地下區域——提爾佛頓盜賊公會。",
		)
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
	joined := strings.Join(texts, " ")
	if strings.Contains(joined, "YOU HAVE ALSO FOUND A MAP OF THE TEMPLE") &&
		strings.Contains(joined, "JOURNAL ENTRY 20") {
		s.appendJournalPages("手札條目 20：", []string{s.catalog.Text(
			"journal_entry_20",
			"手札條目 20：摩安德原始神殿的地圖。入口通往環繞祭壇區的曲折通道；"+
				"地圖標出了中央神殿、下層入口，以及散布各處的房間與死路。"+
				"原始地圖圖像保存於 Adventurer's Journal。",
		)})
	}
	if strings.Contains(joined, "SHE TELLS HER STORY") &&
		strings.Contains(joined, "JOURNAL ENTRY 3") {
		s.appendJournalPages("手札條目 3", []string{
			s.catalog.Text(
				"journal_entry_3_1",
				"手札條目 3（1/3）：愛麗雅絲說，她曾同樣受枷印控制。故事始於一名首席豎琴手；"+
					"他想創造不朽的人形容器，完整保存自己的歌曲與故事，卻在實驗中害死助手。"+
					"豎琴手議會剝奪他的力量、魔法物品與姓名，抹去世人對其作品的記憶，並把他囚禁在異次元。",
			),
			s.catalog.Text(
				"journal_entry_3_2",
				"手札條目 3（2/3）：一群法師與怪物找到無名詩人，利用他的研究製造愛麗雅絲。"+
					"惡魔法爾斯綁架了來自異世界的蜥蜴人聖武士龍餌，企圖以善良生命完成儀式；"+
					"龍餌卻把部分靈魂贈給愛麗雅絲，無名詩人也犧牲自己，協助兩人逃脫。",
			),
			s.catalog.Text(
				"journal_entry_3_3",
				"手札條目 3（3/3）：愛麗雅絲醒來時帶著偽造記憶與相似枷印，後來靠反抗每道魔法強制、"+
					"摧毀製造枷印的人或組織而恢復自由。摩安德教徒也參與過她的束縛；她認為摩安德正企圖復生，"+
					"其新祭壇就在這座原始神殿中，因此提議與龍餌協助隊伍。",
			),
		})
	}
	if strings.Contains(joined, "YOU HAVE PLEASED THE COMMANDER") &&
		strings.Contains(joined, "JOURNAL ENTRY 22") {
		s.appendJournalPages("手札條目 22", []string{
			s.catalog.Text(
				"journal_entry_22_1",
				"手札條目 22（1/2）：紅羽衛指揮官說，他不明白你們為何執意進入摩安德之坑，"+
					"但願意准許隊伍自由通過尤拉什。紅羽衛不會找你們麻煩；然而城市仍在圍攻之中，"+
					"他無法派人一路護送。據報散提爾堡派出了恐怖小隊，東側也有人目擊蔓生怪。",
			),
			s.catalog.Text(
				"journal_entry_22_2",
				"手札條目 22（2/2）：指揮官提供摩安德之坑、檢查哨、營房與士兵食堂的位置圖（手札 52）。"+
					"隊伍可在營房休息並到食堂用餐。他警告尤拉什的城牆與路面歷經重創，許多區域可能不穩。",
			),
		})
		s.appendJournalPages("手札條目 52：", []string{s.catalog.Text(
			"journal_entry_52",
			"手札條目 52：紅羽衛提供的尤拉什城區地圖。北側入口附近是指揮官辦公室；"+
				"圖上標出各處檢查哨、士兵食堂、營房，以及通往城市中央「摩安德之坑」的單向入口。"+
				"原始地圖圖像保存於 Adventurer's Journal 第 15 頁。",
		)})
	}
	if strings.Contains(joined, "THE GUILDMASTER GASPS") &&
		strings.Contains(joined, "JOURNAL ENTRY 4") {
		s.appendJournalPages("手札條目 4：", []string{s.catalog.Text(
			"journal_entry_4",
			"手札條目 4：一幅標有「公會」與「據點」的下水道地圖。它畫出一條狹長、曲折的地下通路，"+
				"由提爾佛頓盜賊公會一路通往火刀據點；原始地圖圖像保存於 Adventurer's Journal 第 11 頁。",
		)})
	}
	if strings.Contains(joined, "LEADER OF THE FIRE KNIVES") &&
		strings.Contains(joined, "JOURNAL ENTRY") &&
		strings.Contains(joined, "11") {
		s.appendJournalPages("手札條目 11", []string{
			s.catalog.Text(
				"journal_entry_11_1",
				"手札條目 11（1/2）：火刀首領說，你們來得正是時候；他們預料國王會落入備用陷阱。"+
					"可惜你們攻擊了錯誤目標。接著他指向兩名被綁在牆邊的俘虜：一名瘦削的鬍鬚男子，"+
					"以及一名繫著破舊紫色肩帶的年輕女子。",
			),
			s.catalog.Text(
				"journal_entry_11_2",
				"手札條目 11（2/2）：那名男子是擅長模仿的喬吉・維文斯普爾；首領命他再次模仿國王的聲音。"+
					"女子則是讓國王得以來到此地的娜卡西亞公主。就在此刻，公主掙脫束縛，抄起手邊的棍棒"+
					"重擊首領，並高喊：「快！趁首領還不能喚起你們的枷印前，解決他的守衛！」",
			),
		})
	}
	if strings.Contains(joined, "FIRE KNIVES HAVE BEEN DEFEATED") &&
		strings.Contains(joined, "JOURNAL ENTRY 54") {
		s.appendJournalPages("手札條目 54：", []string{s.catalog.Text(
			"journal_entry_54",
			"手札條目 54：公主正審問稍微恢復意識的首領。她將匕首抵在他喉頭，首領嘶聲答應解除枷印。"+
				"他吐出一個毫無意義的音節，你們身上的火刀枷印隨之消退。",
		)})
	}
	if strings.Contains(joined, "FREEING GIOGI") &&
		strings.Contains(joined, "JOURNAL ENTRY 53") {
		s.appendJournalPages("手札條目 53", []string{
			s.catalog.Text(
				"journal_entry_53_1",
				"手札條目 53（1/2）：屋頂突然消失，阿祖恩國王、宮廷法師凡格達海斯與皇家衛隊降入房中。"+
					"衛兵指認你們曾企圖弒君；娜卡西亞公主立刻擋在父王與你們之間，說明你們受火刀控制，"+
					"不僅無法自主，還救了她。",
			),
			s.catalog.Text(
				"journal_entry_53_2",
				"手札條目 53（2/2）：國王仍以你們曾攻擊他、身上又有其他控制枷印為由，將你們逐出科米爾。"+
					"離開前，剛德祭司加里踉蹌現身，與公主相擁。衛兵把你們帶到城外後離去；不久，加里與"+
					"娜卡西亞共乘一騎奔向北方，公主在遠處向你們揮手。",
			),
		})
	}
	if strings.Contains(joined, "JOURNAL ENTRY 31.") {
		s.appendJournalPages("手札條目 31：", []string{s.catalog.Text(
			"journal_entry_31",
			"手札條目 31：你們是由一群紅袍人送來的。他們說在路上發現你們時，你們已奄奄一息；"+
				"房錢已預付，所以想住多久都可以。你們來時身上就有那些刺青，旅店老闆娘從未見過類似圖紋。"+
				"她建議去找賢者菲拉妮，往北兩個街區就能找到她。",
		)})
	}
	if strings.Contains(joined, "SHE TALKS") && strings.Contains(joined, "38.") {
		s.appendJournalPages("手札條目 38", []string{
			s.catalog.Text("journal_entry_38_1",
				"手札條目 38（1/3）：你們帶著五個不同組織的印記。菲拉妮認得其中三個，"+
					"一個從未見過，最後一個則令她憂心。火焰與匕首是「火刀」的標誌；這群刺客過去以西門城為據點。"+
					"舊組織已遭摧毀，如今必定另有新基地，但她不知道位於何處。"),
			s.catalog.Text("journal_entry_38_2",
				"手札條目 38（2/3）：掌心之口是神祇摩安德的標誌。祂曾被逐出世界，"+
					"卻短暫化作一堆污穢再現；被擊敗前摧毀了尤拉什城的一部分。其教團偏愛綠色。"+
					"三角形中的華麗 Z 則代表散塔林會「黑網」。"),
			s.catalog.Text("journal_entry_38_3",
				"手札條目 38（3/3）：黑網是由散提爾堡的祭司、法師與盜賊組成的邪惡聯盟，"+
					"甚至有人說他們實際統治著散提爾堡。燃燒印記她從未見過。最後的弦月標誌，"+
					"與暗影谷一位強大賢者有令人不安的相似之處；為了自身安全，她不願再多說。"),
		})
	}
	if strings.Contains(joined, "ORNATE KNIFE") && strings.Contains(joined, "17.") {
		s.appendJournalPages("手札條目 17：", []string{s.catalog.Text(
			"journal_entry_17",
			"手札條目 17：巷子裡只留下一把華麗匕首。它有深色握柄、寬大的護手，"+
				"刀刃呈不規則火焰形；原手札在此畫出匕首外觀，這正是追查「火刀」組織的重要線索。",
		)})
	}
	if strings.Contains(joined, "A HOODED, GREY ROBED MAN SITS IN A DARK CORNER") &&
		strings.Contains(joined, "18.") {
		s.appendJournalPages("手札條目 18：", []string{s.catalog.Text(
			"journal_entry_18",
			"手札條目 18：你們身上的弦月枷印與伊爾明斯特的徽記極為相似，而他絕不會容忍有人冒稱受他烙印。"+
				"最好悄悄離開暗影谷，乘船前往阿沙本福德，再一路向南找到那位紅袍法師的高塔，逼他解除枷印；"+
				"否則留在這裡，恐怕只會被伊爾明斯特變成一隻蠑螈。",
		)})
	}
	if strings.Contains(joined, "REMOVE CURSE SPELL") && strings.Contains(joined, "19.") {
		s.appendJournalPages("手札條目 19：", []string{s.catalog.Text(
			"journal_entry_19",
			"手札條目 19：祭司開始施法時，青色枷突然發出耀眼光芒。藍色火焰自印記竄出，"+
				"在房內四處迸射；眾人痛苦得扭曲身體，祭司只得停止施法。"+
				"他說：「這些枷鎖會抵抗我的神力，我無法解除它們。祝你們往後順利，願剛德與你們同在。」",
		)})
	}
	if strings.Contains(joined, "YOU DISARM THE FIRE KNIVES") &&
		strings.Contains(joined, "JOURNAL ENTRY 26") {
		s.appendJournalPages("手札條目 26：", []string{s.catalog.Text(
			"journal_entry_26",
			"手札條目 26：這些人中了入侵牧師施展的定身法術。那名牧師是為了營救被關在南方首領房裡的囚犯而來；所幸火刀最後在這間房裡制伏了他。",
		)})
	}
	if strings.Contains(joined, "DRAWERS OF A ROSEWOOD DESK") &&
		strings.Contains(joined, "9. OTHER ITEMS") {
		s.appendJournalPages("手札條目 9：", []string{s.catalog.Text(
			"journal_entry_9",
			"手札條目 9：文件旁畫著一個帶火焰輪廓的人形，軀幹上有三道彎曲符號。註記寫著："+
				"一、具有燃燒靈氣；二、能附身其他軀體；三、與光芒之池有所牽連。"+
				"原始圖像保存於 Adventurer's Journal 第 12 頁。",
		)})
	}
	if strings.Contains(joined, "HAND KEPT THE PAPER") &&
		strings.Contains(joined, "JOURNAL ENTRY 29") {
		s.appendJournalPages("手札條目 29：", []string{s.catalog.Text(
			"journal_entry_29",
			"手札條目 29：未被燒毀的部分寫著：「……我們的盟友能控制火焰、從一具軀體掠入另一具軀體，"+
				"並展現多種異次元力量。我的結論是，『烈焰之主』不可能是別人，只會是泰蘭索斯……」",
		)})
	}
	if strings.Contains(joined, "SLAYERS OF DRAGONKIND") &&
		strings.Contains(joined, "JOURNAL ENTRY 15") {
		s.appendJournalPages("手札條目 15", []string{
			s.catalog.Text(
				"journal_entry_15_1",
				"手札條目 15（1/2）：德拉坎德羅斯對龍群宣稱，你們是伊爾明斯特為報復龍群飛襲、"+
					"意圖消滅所有龍族的棋子；他願把這些「刺客」交給龍群，以證明自己的善意。"+
					"他又指著你們手臂上的印記，誣稱那是奴役巨龍的泰蘭索斯之標誌。",
			),
			s.catalog.Text(
				"journal_entry_15_2",
				"手札條目 15（2/2）：一隻黑龍並不相信。牠見過相似的發光枷印控制戰士，"+
					"認定真正操縱你們的人是德拉坎德羅斯，要求先解除枷印再審判你們。"+
					"黑龍以冒煙的酸液威嚇他；德拉坎德羅斯只得念咒，讓自己那枚枷印逐漸消失。",
			),
		})
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
		return catalog.Text("ecl_training_prompt", "你們要接受訓練嗎？")
	case "YOU'RE SHOWING GREAT PROGRESS. RETURN AGAIN WHEN":
		return catalog.Text("ecl_training_progress", "你們的進步很大。準備好時再回來，")
	case "YOU ARE READY.' YOU EXIT THE HALL.":
		return catalog.Text("ecl_training_exit", "你們離開了訓練場。")
	case "'WHAT'S YOUR PLEASURE?'":
		return catalog.Text("ecl_tavern_pleasure", "「幾位想來點什麼？」")
	case "'A SPECIAL CUSTOMER'S ARRIVED. YOU HAVE TO SLIP":
		return catalog.Text("ecl_tavern_special_1", "「有位特別的客人到了。你們得暫時")
	case "OUTSIDE FOR A MOMENT.' DO YOU GO?":
		return catalog.Text("ecl_tavern_special_2", "到外面等一下。」你們要出去嗎？")
	case "AS YOU BEGIN TO WALK OUT THE DOOR, YOU SEE A":
		return catalog.Text("ecl_tavern_purple_1", "你們正要走出門時，看見一名")
	case "YOUNG WOMAN WITH A PURPLE SASH SLIP IN THE SIDE DOOR.":
		return catalog.Text("ecl_tavern_purple_2", "繫著紫色腰帶的年輕女子溜進側門。")
	case "A FEW OF THE OTHER PATRONS HANG BACK, AS IF TO MEET HER.":
		return catalog.Text("ecl_tavern_purple_3", "另有幾名酒客刻意留下，似乎正等著與她會面。")
	case "AS YOU CONSIDER YOUR NEXT MOVE, YOU HEAR A":
		return catalog.Text("ecl_tavern_commotion_1", "你們正盤算下一步時，聽見")
	case "COMMOTION AROUND THE SIDE OF THE BUILDING. DO YOU GO":
		return catalog.Text("ecl_tavern_commotion_2", "建築側邊傳來騷動。要前去")
	case "TO INVESTIGATE?":
		return catalog.Text("ecl_tavern_commotion_3", "調查嗎？")
	case "THERE IS NOTHING HERE NOW, EXCEPT FOR AN ORNATE":
		return catalog.Text("ecl_tavern_knife_1", "這裡如今空無一人，只留下一把華麗")
	case "KNIFE":
		return catalog.Text("ecl_tavern_knife_2", "匕首")
	case "17.":
		return catalog.Text("ecl_tavern_journal_17", "第 17 條。")
	case "'I AM THE HIGH PRIEST. YOU LOOK TROUBLED, MY CHILDREN. DO YOU WISH TO TELL ME YOUR STORY?'":
		return catalog.Text("ecl_high_priest_intro", "「我是這裡的高階祭司。孩子們，你們看來憂心忡忡，願意告訴我發生了什麼事嗎？」")
	case "HE LISTENS SYMPATHETICALLY AND CASTS A REMOVE CURSE SPELL":
		return catalog.Text("ecl_high_priest_spell", "他同情地聽完你們的遭遇，接著施展移除詛咒法術，")
	case "19.":
		return catalog.Text("ecl_high_priest_journal_19", "第 19 條。")
	case "YOU AWAKEN IN A SMALL ROOM. LOOKING AROUND, YOU NOTICE":
		return catalog.Text("ecl_new_game_awaken", "你們在一間小房間裡醒來。環顧四周，你們注意到")
	case "THAT ALL YOUR GEAR IS GONE, AS IS YOUR MEMORY OF RECENT EVENTS.":
		return catalog.Text("ecl_new_game_gear_gone", "所有裝備都不見了，最近發生的事情也完全想不起來。")
	case "ADDING TO YOUR DISQUIET, YOU NOTICE THAT YOUR SWORD ARM":
		return catalog.Text("ecl_new_game_disquiet", "更令人不安的是，你們發現持劍的手臂")
	case "HAS BEEN SOMEHOW IMPRINTED WITH STRANGE PATTERNS. THE REST":
		return catalog.Text("ecl_new_game_patterns", "不知為何被烙上了奇異圖紋。隊伍中的其他人")
	case "OF YOUR PARTY ARE IDENTICALLY MARKED.":
		return catalog.Text("ecl_new_game_identical", "也都帶著完全相同的印記。")
	case "A PATROL ARRIVES.":
		return catalog.Text("ecl_tilverton_patrol_arrives", "一支皇家巡邏隊抵達。")
	case "ROYAL GUARDS TELL YOU TO MOVE ALONG.":
		return catalog.Text("ecl_tilverton_guards_move", "皇家衛兵命令你們立刻離開。")
	case "'WELCOME TO THE FAIR CITY OF TILVERTON,' BEAMS THE":
		return catalog.Text("ecl_tilverton_inn_welcome", "「歡迎來到美麗的提爾佛頓！」")
	case "INNKEEPER. THEN SHE NOTICES YOUR COLLECTIVE SCOWLS.":
		return catalog.Text("ecl_tilverton_inn_scowls", "旅店老闆娘笑容滿面地說，接著才注意到你們全都沉著臉。")
	case "'PLEASE CALM DOWN WHILE I EXPLAIN.'":
		return catalog.Text("ecl_tilverton_inn_calm", "「請先冷靜，讓我解釋。」")
	case "YOU LISTEN,":
		return catalog.Text("ecl_tilverton_inn_listen", "你們聽她說明，")
	case "AND YOU RECORD IT IN JOURNAL ENTRY":
		return catalog.Text("ecl_tilverton_inn_record", "並將內容記入冒險手札")
	case "31. 'PERHAPS THE SAGE WILL HELP. YOU CAN GET WEAPONS FROM":
		return catalog.Text("ecl_tilverton_inn_sage", "第 31 條。「或許賢者能幫上忙；你們也可以到")
	case "THE SHOP ACROSS THE WAY.'":
		return catalog.Text("ecl_tilverton_inn_shop", "對街的商店取得武器。」")
	case "'I AM THE SAGE FILANI. YOU ARE HERE ABOUT THE SIGILS,":
		return catalog.Text("ecl_filani_intro", "「我是賢者菲拉妮。你們是為了那些印記而來，")
	case "CORRECT?'":
		return catalog.Text("ecl_filani_correct", "對吧？」")
	case "'THIS IS AN INTERESTING CASE. I'LL DO IT FOR HALF YOUR":
		return catalog.Text("ecl_filani_price", "「這個案例很有意思。我可以替你們研究，代價是你們")
	case "FUNDS. HOW MUCH DO YOU HAVE?'":
		return catalog.Text("ecl_filani_funds", "一半的財物。你們身上有多少？」")
	case "SHE TALKS":
		return catalog.Text("ecl_filani_talks", "她開始解說，內容記於冒險手札")
	case "38.":
		return catalog.Text("ecl_filani_journal_38", "第 38 條。")
	case "'DO NOT THINK SAGES ARE FOOLS.' SHE SENDS YOU OUT.":
		return catalog.Text("ecl_filani_lie", "「別把賢者當成傻瓜。」她把你們趕了出去。")
	case "'THEN WE HAVE NOTHING TO DISCUSS.'":
		return catalog.Text("ecl_filani_no", "「那我們就沒什麼好談的了。」")
	case "'WE HAVE A SELECTION OF THE FINEST CORMYR STEEL.":
		return catalog.Text("ecl_weaponers_intro", "「我們備有最上等的科米爾精鋼武器。")
	case "INTERESTED?":
		return catalog.Text("ecl_weaponers_interested", "有興趣嗎？」")
	case "'MAY YOU ALWAYS STRIKE TRUE.'":
		return catalog.Text("ecl_weaponers_farewell", "「願你每次出手都準確命中。」")
	case "'GOOD DAY THEN.'":
		return catalog.Text("ecl_weaponers_decline", "「那麼，祝你今日愉快。」")
	case "'GOOD DAY TO YOU, GENTLE PERSONS. DO YOU WISH":
		return catalog.Text("ecl_general_store_intro", "「各位貴客，日安。你們想要")
	case "TO MAKE A PURCHASE?'":
		return catalog.Text("ecl_general_store_purchase", "買些東西嗎？」")
	case "'THANK YOU. RETURN SOON.'":
		return catalog.Text("ecl_general_store_farewell", "「多謝惠顧，歡迎再來。」")
	case "YOU MOVE AWAY.":
		return catalog.Text("ecl_move_away", "你們離開此處。")
	case "ON YOUR WAY TO THE TOWN OF TILVERTON YOU ARE":
		return catalog.Text("ecl_opening_tilverton", "前往提爾佛頓鎮的途中，你們")
	case "AMBUSHED, CAPTURED, AND KNOCKED UNCONSCIOUS. WHEN":
		return catalog.Text("ecl_opening_ambushed", "遭到伏擊、被俘，並被擊昏。當")
	case "YOU AWAKE YOUR PARTY HAS BEEN CURSED WITH FIVE AZURE":
		return catalog.Text("ecl_opening_cursed", "你們醒來時，隊伍已被五個青色")
	case "SYMBOLS.":
		return catalog.Text("ecl_opening_symbols", "符印下了詛咒。")
	case "THE SYMBOLS ENSNARE YOUR WILL LIKE METAL BONDS.":
		return catalog.Text("ecl_opening_ensnare", "這些符印如金屬枷鎖般束縛你們的意志。")
	case "AND WHEN THE BONDS GLOW YOU MUST DO AS THEY COMMAND.":
		return catalog.Text("ecl_opening_command", "當枷印發光時，你們就必須服從它們的命令。")
	case "YOUR ONLY HOPE IS TO SEARCH THE FORGOTTEN REALMS":
		return catalog.Text("ecl_opening_only_hope", "唯一的希望，是在被遺忘的國度中尋找")
	case "FOR THE MEMBERS OF THE ALLIANCE WHO CREATED THE BONDS":
		return catalog.Text("ecl_opening_alliance", "創造這些枷印的聯盟成員")
	case "AND REGAIN CONTROL OF YOUR OWN DESTINY.":
		return catalog.Text("ecl_opening_destiny", "並重新掌握自己的命運。")
	case "NOWHERE IN THE REALMS IS COMPLETELY SAFE. EVEN":
		return catalog.Text("ecl_opening_nowhere_safe", "國度中沒有任何地方絕對安全。即使")
	case "THE MOST PEACEFUL SCENE CAN HIDE A DEADLY FOE.":
		return catalog.Text("ecl_opening_deadly_foe", "最平靜的景象，也可能藏著致命敵人。")
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
		return catalog.Text("ecl_cleric_howl", "露出勝利神情，接著嚎叫：")
	case "'THE CHOSEN ONES!'":
		return catalog.Text("ecl_cleric_chosen_ones", "「被選中之人！」")
	case "YOU FIND A WAR BLASTED SECTION OF THE CITY.":
		return catalog.Text("ecl_war_blasted_city", "你們找到城市中一片遭戰火摧毀的區域。")
	case "YOU DISCOVER A SMALL MAGIC SHOP.":
		return catalog.Text("ecl_small_magic_shop", "你們發現一間小型魔法商店。")
	default:
		return line
	}
}
