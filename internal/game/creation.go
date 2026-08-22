package game

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/area"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiostate"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/party"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
	engineaction "github.com/wicanr2/golden-box-remake-engine/combat/action"
	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

func starterCharacters(pack *goldenbox.Pack, language string) ([]party.Character, error) {
	if pack == nil || pack.CharacterCreation == nil {
		return nil, fmt.Errorf("character creation templates are missing from game pack")
	}
	result := make([]party.Character, 0, len(pack.CharacterCreation.Templates))
	for index, template := range pack.CharacterCreation.Templates {
		if template.RaceID > uint8(party.RaceHalfOrc) || template.PrimaryClassID > uint8(party.ClassThief) {
			return nil, fmt.Errorf("character creation template %q has unsupported race/class IDs %d/%d", template.ID, template.RaceID, template.PrimaryClassID)
		}
		name, ok := pack.Text(template.DisplayID, language)
		if !ok || name == "" {
			return nil, fmt.Errorf("character creation template %q display %q is unavailable", template.ID, template.DisplayID)
		}
		var levels [8]uint8
		copy(levels[:], template.ClassLevels)
		abilities := template.BaseAbilities
		character := party.Character{
			ID: "creation." + template.ID, Name: name,
			Race: party.Race(template.RaceID), Class: party.Class(template.PrimaryClassID),
			RawClassID: template.RawClassID, Level: int(template.Level), ClassLevels: levels,
			SavingThrows: append([]uint8(nil), template.SavingThrows...),
			Abilities: party.Abilities{
				Strength: abilities[0], Intelligence: abilities[1], Wisdom: abilities[2],
				Dexterity: abilities[3], Constitution: abilities[4], Charisma: abilities[5],
			},
		}
		if len(character.SavingThrows) != 5 {
			return nil, fmt.Errorf("character creation template %q must provide five saving throws", template.ID)
		}
		lookup, lookupErr := gamepack.AgeLookup()
		if lookupErr != nil {
			return nil, lookupErr
		}
		if _, err := party.StartingAgeSpecFrom(lookup, character.Race, character.Class); err != nil {
			return nil, fmt.Errorf("character creation template %q age: %w", template.ID, err)
		}
		if err := character.Validate(); err != nil {
			return nil, fmt.Errorf("character creation template %d %q: %w", index, template.ID, err)
		}
		result = append(result, character)
	}
	return result, nil
}

func (s *State) OpenCharacterCreation() error {
	if s.Mode == ModeCombat {
		return fmt.Errorf("character creation is unavailable during combat")
	}
	s.creationReturnMode = s.Mode
	options, err := starterCharacters(s.dataPack, s.catalog.Language)
	if err != nil {
		return err
	}
	s.CreationOptions = options
	s.CreationRoster = nil
	s.CreationCursor = 0
	s.CreationMessage = s.catalog.Text("creation_prompt", "creation_prompt")
	s.Mode = ModeCharacterCreation
	// 角色建立有自己的曲子（原作 `GEN`，overlay-17 `0B08h` 寫 `MUSICNO := 2`）。
	// ⚠ 那一處**不看場景**，所以 pack 那一側是每一段都列（spec 1192）。
	s.requestMusicForCurrentBlock(creationMusicContext)
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
	// The reference assigns age after race/class selection. Keep the template
	// values editable, then apply the generated age only to the copied party
	// character so an option can be previewed and reused without double-aging.
	seed := int64(0xC0AB) + int64(index)*97 + int64(len(s.CreationRoster))*13
	lookup, err := gamepack.AgeLookup()
	if err != nil {
		return err
	}
	age, err := party.RollStartingAgeFrom(lookup, character.Race, character.Class, seed)
	if err != nil {
		return err
	}
	character.Age = age
	character.Abilities, err = character.Abilities.WithAgeEffects(character.Race, int(age))
	if err != nil {
		return err
	}
	character.ID = fmt.Sprintf("%s-%d", character.ID, len(s.CreationRoster)+1)
	s.CreationRoster = append(s.CreationRoster, character)
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_added", "creation_added"), character.Name, len(s.CreationRoster))
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
		s.CreationMessage = s.catalog.Text("creation_ability_prompt", "creation_ability_prompt")
	}
	return nil
}

func (s *State) RerollCreationAbilities(seed int64) error {
	if s.Mode != ModeCharacterCreation || s.CreationCursor < 0 || s.CreationCursor >= len(s.CreationOptions) {
		return fmt.Errorf("ability reroll is unavailable")
	}
	s.CreationOptions[s.CreationCursor].Abilities = party.RollAbilities(seed)
	s.CreationMessage = s.catalog.Text("creation_rerolled", "creation_rerolled")
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
	// 夾值走原作的兩張表（種族上下限 ＋ 職業組合最低要求，spec 1086／1099），
	// 而不是通用的 3..18。
	limits, err := gamepack.LimitLookup()
	if err != nil {
		return err
	}
	option := &s.CreationOptions[s.CreationCursor]
	if err := option.Abilities.AdjustWithin(limits, option, s.CreationAbility, delta); err != nil {
		s.CreationMessage = err.Error()
		return err
	}
	value, _ := s.CreationOptions[s.CreationCursor].Abilities.Value(s.CreationAbility)
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_ability_updated", "creation_ability_updated"), value)
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
	s.CreationMessage = fmt.Sprintf(s.catalog.Text("creation_named", "creation_named"), s.CreationName)
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
	s.CreationMessage = ""
	if s.session == nil {
		// Data-neutral tests/embedders may construct State without the
		// original image. Production NewStateFromECLBlocks always takes the
		// verified block-0x01 path below.
		s.Mode = ModeWilderness
		s.Prompt = s.catalog.Text("party_ready", "party_ready")
		s.Choices = []string{s.localizeOption("ENTER CITY"), s.localizeOption("JOURNEY ON"), s.localizeOption("CAMP")}
		s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
		return nil
	}
	return s.BeginAdventure()
}

// LocaleText exposes the State-owned catalog to the frontend without letting
// renderer code maintain a second translation table.
func (s *State) LocaleText(key string) string { return s.catalog.Text(key, key) }

func (s *State) CharacterRaceName(race party.Race) string {
	key := "race_unknown"
	switch race {
	case party.RaceDwarf:
		key = "race_dwarf"
	case party.RaceElf:
		key = "race_elf"
	case party.RaceGnome:
		key = "race_gnome"
	case party.RaceHalfElf:
		key = "race_half_elf"
	case party.RaceHalfling:
		key = "race_halfling"
	case party.RaceHuman:
		key = "race_human"
	case party.RaceHalfOrc:
		key = "race_half_orc"
	}
	return s.LocaleText(key)
}

func (s *State) CharacterClassName(class party.Class) string {
	return s.localizedCharacterClassName(class)
}

func (s *State) SavePartyFile(path string) error {
	if len(s.partyRoster) == 0 {
		return fmt.Errorf("no character-created party is available to save")
	}
	areaState := s.Area
	areaState.GameTime = s.gameClock
	var sessionSnapshot *ecl.SessionSnapshot
	if s.session != nil {
		snapshot, snapshotErr := s.session.Snapshot()
		if snapshotErr != nil {
			return snapshotErr
		}
		sessionSnapshot = &snapshot
	}
	combatSnapshot, err := s.activeCombatSnapshot()
	if err != nil {
		return err
	}
	data, err := partySave.EncodeGameWithAdventureState(s.partyRoster, areaState, uint8(s.Mode), uint8(s.Location), s.MapX, s.MapY, s.DungeonX, s.DungeonY, s.DungeonDirection, s.DungeonWallType, s.DungeonWallRoof, s.gameClock, s.gameAgeCycles, sessionSnapshot, combatSnapshot, s.musicSnapshot(), s.oneShotSnapshot(), s.journalMessageIDs, s.DungeonSearchEnabled, s.dungeonSearchEdgeIDs())
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (s *State) activeCombatSnapshot() (*partySave.CombatSnapshot, error) {
	if s.battle == nil {
		if s.Mode == ModeCombat {
			return nil, fmt.Errorf("combat mode has no active battle to save")
		}
		return nil, nil
	}
	if s.Mode != ModeCombat {
		// The original flow keeps the completed Battle object around so callers
		// can inspect its final status while ECL continues through treasure,
		// journal and world-map handoff. Once that handoff is complete, the
		// completed battle is no longer a resumable save boundary; serializing it
		// would make LoadPartyFile reject the otherwise valid wilderness save.
		if s.battle.Status() != combat.StatusActive &&
			!s.treasureMenu && !s.treasureTakeMenu && len(s.pendingTreasureItems) == 0 && !s.treasureResumeECL {
			return nil, nil
		}
		return nil, fmt.Errorf("active battle cannot be saved from mode %d", s.Mode)
	}
	battleSnapshot, err := s.battle.Snapshot()
	if err != nil {
		return nil, err
	}
	var visual *combat.VisualEvent
	if s.combatVisual != nil {
		copy := *s.combatVisual
		copy.Impacts = append([]combat.VisualImpactTarget(nil), copy.Impacts...)
		copy.Segments = append([]combat.VisualPathSegment(nil), copy.Segments...)
		visual = &copy
	}
	return &partySave.CombatSnapshot{
		Battle: battleSnapshot, Turns: append([]combat.Turn(nil), s.combatTurns...),
		TurnIndex: s.combatTurnIndex, DelayedTurns: cloneDelayedTurns(s.combatDelayedTurns),
		TargetIndex: s.combatTargetIndex, CastingSpell: s.combatCastingSpell,
		CastingClass: uint8(s.combatCastingClass), CastingClassSet: s.combatCastingClassSet,
		SpellTargetIndex: s.combatSpellTargetIndex, SpellTargetPoint: s.combatSpellTargetPoint,
		SpellTargetsPoint: s.combatSpellTargetsPoint, MoveMode: s.combatMoveMode,
		MoveRemaining: s.combatMoveRemaining, Speed: uint8(s.combatSpeed),
		QuickMagic: s.combatQuickMagic, ReferenceCoords: s.combatReferenceCoords,
		View: s.combatView, ViewFighterID: s.combatViewFighterID,
		Message: s.combatMessage, ReturnMode: uint8(s.combatReturnMode),
		VisualSerial: s.combatVisualSerial, VisualEnabled: s.combatVisualEnabled,
		Visual: visual, VisualElapsedNanos: int64(s.combatVisualElapsed),
		VisualTravelSent: s.combatVisualTravelSent,
		VisualImpactSent: s.combatVisualImpactSent, VisualDeathSent: s.combatVisualDeathSent,
		VisualAdvanceTurn: s.combatVisualAdvanceTurn,
	}, nil
}

func cloneDelayedTurns(source map[int]bool) map[int]bool {
	if source == nil {
		return nil
	}
	result := make(map[int]bool, len(source))
	for index, delayed := range source {
		result[index] = delayed
	}
	return result
}

// SaveSAVGAMPrefix writes only the reference-compatible fixed SAVGAM prefix
// while preserving raw records loaded previously. SaveSAVGAMSlot is the
// higher-level API for a loaded prefix plus CHRDAT player sidecars.
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
	s.gameClock = area1.GameTime
	s.GeoMapSet = area1.GameArea
	s.GeoMapBlock = area1.Current3DMapBlockID
	s.MapX = int(container.MapPosX)
	s.MapY = int(container.MapPosY)
	s.DungeonDirection = container.MapDirection
	// ⚠ `MapPos` 那一組就是地城裡的隊伍格（原作 `720Fh`／`7210h`），不是另一
	// 套世界地圖座標；只填 `MapX`／`MapY` 會讓讀檔後的地城位置停在上一次的值。
	s.DungeonX = int(container.MapPosX)
	s.DungeonY = int(container.MapPosY)
	s.DungeonWallType = container.MapWallType
	s.DungeonWallRoof = container.MapWallRoof
	// 存檔裡的三組牆面參數就是「現在每個槽用哪一塊牆磚」，原版載入時照著重載。
	// 不還原的話 remake 只能靠查表重建，而查表拿不到「兩槽模式」那條規則的執行時
	// 閘門（`bank0^[1CEh]`／`[1D0h]`，spec 1087）——槽 2 會被填上一個腳本其實
	// 從來不會去載的選圖。
	s.wallSetParams = container.SetBlocks
	s.savgamPrefix = &container
	if err := s.seedECLBanksFrom(container); err != nil {
		return err
	}
	// 原版存檔的 map state 那五個位元組就是地圖暫存器本身，讀進來之後要回寫，
	// 才輪得到腳本去改它（spec 1172）。
	if s.session != nil {
		s.syncDungeonECLRegisters()
		s.restoreWallPiecesForLoadedBlock()
	}
	return nil
}

// restoreWallPiecesForLoadedBlock 把牆磚選圖補回來。
//
// ★ 為什麼需要。 原版存檔裡**沒有**牆磚選圖，也沒有地圖幾何——整份存檔掃過六個
// GEO 檔集一塊都找不到（spec 1185）。原版是靠**重跑當前 ECL block 的進入碼**
// 把 `LOAD FILES`／`LOAD PIECES` 再發一次拿回來的。remake 載入存檔時不重跑那段
// （會連帶觸發劇情副作用），於是 `LoadPieces` 停在全零。
//
// ⚠ 全零不會報錯，它會把**每一面牆都畫成空氣**，而畫面看起來完全正常：天空、
// 地板、UI 都在，只是牆不見了。拿這種畫面去跟原版比會得到一個很大的差異數字，
// 然後被誤判成「第一人稱畫錯了」。
//
// 查表的鍵是 **ECL block**，不是地圖：GEO5 的 `0x31` 與 `0x32` 共用同一張幾何
// 區塊，牆磚選圖卻不同。
func (s *State) restoreWallPiecesForLoadedBlock() {
	if s.LoadPieces != [3]uint16{} {
		return
	}
	// ⚠ 存檔已經記著每個槽用哪一塊時，**以存檔為準**：那是原版當時真的載了什麼，
	// 查表只是「這一段通常會發什麼」。兩者在「兩槽模式」下不同——槽 2 在那個模式
	// 裡根本不會被載，查表卻會把它填滿。
	if fromSave, ok := wallPiecesFromParams(s.wallSetParams); ok {
		s.LoadPieces = fromSave
		s.loadPiecesPending = true
		return
	}
	// ⚠ 鍵要用**存檔記下來的 ECL 段**（`LastECLBlockID`），不是 session 當下的
	// 段。載入存檔之後 `CurrentBlockID()` 還停在初始化時的世界地圖段 `0x50`，
	// 拿它去查一定查不到——而查不到的後果是安靜地不補，牆照樣是空氣。
	pieces, ok := eclBlockWallPieces[eclBlockKey{
		area: s.Area.GameArea, block: uint8(s.Area.LastECLBlockID)}]
	if !ok {
		return
	}
	s.applyLoadPieces(ecl.RunResult{LoadPiecesRequested: true, LoadPieces: pieces,
		WallSetAssignments: ecl.WallSetAssignmentsFor(pieces, s.session.MemorySnapshot())})
}

// encodeECLBanksInto 把 ECL session 的區 0..3 寫進 SAVGAM 的四塊（spec 1163）。
// 沒有 session 就什麼都不動。
func (s *State) encodeECLBanksInto(container *partySave.SAVGAMContainer) error {
	if s.session == nil {
		return nil
	}
	memory := s.session.MemorySnapshot()
	for _, bank := range []struct {
		record    *[]byte
		low, high uint16
	}{
		{&container.Area1, partySave.ECLBank0Low, partySave.ECLBank0High},
		{&container.Area2, partySave.ECLBank1Low, partySave.ECLBank1High},
		{&container.Runtime, partySave.ECLBank2Low, partySave.ECLBank2High},
	} {
		encoded, err := partySave.EncodeECLBank(memory, bank.low, bank.high, *bank.record)
		if err != nil {
			return err
		}
		*bank.record = encoded
	}
	// 區 3 是 **ECL 位元組碼本身**，位元組定址（spec 1176）。不寫回等於把腳本
	// 對自己位元組碼做的修改丟掉——原版存的是存檔當下那一份。
	encoded, err := partySave.EncodeECLCode(s.session.CodeMemorySnapshot(), container.ECL)
	if err != nil {
		return err
	}
	container.ECL = encoded
	return nil
}

// seedECLBanksFrom 反過來：匯入原版存檔時把那四塊讀回 ECL session。
// ⚠ 每一段自己的暫存（`4C00h`..`4C0Fh`，spec 1162）在原版版面裡只有一份，
// 就是目前這一段的；別段停放的值原版存不下來，讀檔之後從 0 開始。
func (s *State) seedECLBanksFrom(container partySave.SAVGAMContainer) error {
	if s.session == nil {
		return nil
	}
	for _, bank := range []struct {
		record    []byte
		low, high uint16
	}{
		{container.Area1, partySave.ECLBank0Low, partySave.ECLBank0High},
		{container.Area2, partySave.ECLBank1Low, partySave.ECLBank1High},
		{container.Runtime, partySave.ECLBank2Low, partySave.ECLBank2High},
	} {
		if bank.record == nil {
			continue
		}
		memory, err := partySave.DecodeECLBank(bank.record, bank.low, bank.high)
		if err != nil {
			return err
		}
		for address, value := range memory {
			s.session.SetMemoryValue(address, value)
		}
	}
	if container.ECL != nil {
		// ⚠ 只有目前這一段就是存檔當時那一段時才對得上——原版存檔只存得下
		// 一份程式碼視窗（spec 1176）。長度不符就跳過，不要把半份碼蓋上去。
		code, err := partySave.DecodeECLCode(container.ECL)
		if err == nil {
			s.session.RestoreCodeMemory(code)
		}
	}
	return nil
}

// LoadSAVGAMSlot loads the reference save-game prefix plus its CHRDAT player
// bundles from a directory. SaveGame writes the prefix as savgamA..J.dat and
// names party records CHRDAT{key}{1..6}.sav; .swg/.fx are optional sidecars.
// This loader intentionally does not infer unsupported multi-class fields or
// rewrite the original files.
func (s *State) LoadSAVGAMSlot(directory string, key byte) error {
	return s.loadSAVGAMSlot(directory, key, party.ParseDOSPlayerFiles)
}

// LoadOriginalSAVGAMSlot 匯入**原版**的 SAVGAM 槽：版面與 LoadSAVGAMSlot 完全
// 相同，差別在角色名與物品名以原版編碼（Big5，ASCII 相容）解讀。
//
// ★ 為什麼要由呼叫端指定，不能自動判斷。 remake 自己寫出來的槽用的是同一個
// 版面（`SaveSAVGAMSlot` 保留原始位元組再用 UTF-8 覆寫名字），所以**光看檔案
// 分不出來源**。猜錯的代價是不對稱的：把 remake 的存檔當原版讀會把 UTF-8 名字
// 當 Big5 解成亂碼，而英文原版兩條路結果一樣（ASCII 相容）——也就是說猜錯在
// 英文資料上完全看不出來，中文版才會炸。所以由呼叫端明講。
func (s *State) LoadOriginalSAVGAMSlot(directory string, key byte) error {
	return s.loadSAVGAMSlot(directory, key, party.ParseOriginalDOSPlayerFiles)
}

func (s *State) loadSAVGAMSlot(directory string, key byte,
	parse func(string, party.DOSPlayerFiles) (party.Character, error)) error {
	if key < 'A' || key > 'J' {
		return fmt.Errorf("SAVGAM slot key %q is outside A..J", key)
	}
	prefixPath := filepath.Join(directory, fmt.Sprintf("savgam%c.dat", key+('a'-'A')))
	if err := s.LoadSAVGAMPrefix(prefixPath); err != nil {
		return err
	}
	container := s.savgamPrefix
	if container == nil {
		return fmt.Errorf("SAVGAM prefix was not retained")
	}
	roster := make(party.Roster, 0, container.PartyCount)
	for index := 0; index < int(container.PartyCount); index++ {
		base := fmt.Sprintf("CHRDAT%c%d", key, index+1)
		recordPath := filepath.Join(directory, base+".sav")
		record, err := os.ReadFile(recordPath)
		if err != nil {
			return fmt.Errorf("SAVGAM player %s: %w", base, err)
		}
		effects, err := readOptionalSAVGAMSidecar(filepath.Join(directory, base+".fx"))
		if err != nil {
			return err
		}
		inventory, err := readOptionalSAVGAMSidecar(filepath.Join(directory, base+".swg"))
		if err != nil {
			return err
		}
		character, err := parse(base, party.DOSPlayerFiles{Record: record, Effects: effects, Inventory: inventory})
		if err != nil {
			return fmt.Errorf("SAVGAM player %s: %w", base, err)
		}
		if s.savgamPlayers == nil {
			s.savgamPlayers = make(map[string]party.DOSPlayerFiles)
		}
		s.savgamPlayers[base] = party.DOSPlayerFiles{
			Record: append([]byte(nil), record...), Effects: append([]byte(nil), effects...), Inventory: append([]byte(nil), inventory...),
		}
		roster = append(roster, character)
	}
	if len(roster) == 0 {
		return fmt.Errorf("SAVGAM slot %c has no player records", key)
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
	s.partyRoster = roster
	s.Mode = ModeWilderness
	s.Prompt = s.catalog.Text("party_ready", "party_ready")
	s.Choices = []string{s.localizeOption("ENTER CITY"), s.localizeOption("JOURNEY ON"), s.localizeOption("CAMP")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	return nil
}

// SaveSAVGAMSlot writes the current party back to a reference-style slot.
// Only documented player offsets and decoded SWG/FX records are serialized;
// every unknown byte in each original .sav record is preserved. Files are
// prepared in a sibling staging directory before replacement, so a failed
// encode never leaves a partially written staged bundle behind.
func (s *State) SaveSAVGAMSlot(directory string, key byte) error {
	if key >= 'a' && key <= 'j' {
		key -= 'a' - 'A'
	}
	if key < 'A' || key > 'J' {
		return fmt.Errorf("SAVGAM slot key %q is outside A..J", key)
	}
	if s.savgamPrefix == nil {
		return fmt.Errorf("no SAVGAM prefix is loaded")
	}
	if len(s.partyRoster) == 0 || len(s.partyRoster) > 6 {
		return fmt.Errorf("SAVGAM party size %d is outside 1..6", len(s.partyRoster))
	}
	if s.savgamPlayers == nil {
		return fmt.Errorf("no raw SAVGAM player records are retained")
	}

	container, err := s.savgamContainerForSave()
	if err != nil {
		return err
	}
	for index := range container.CharacterRefs {
		container.CharacterRefs[index] = nil
	}
	for index := range s.partyRoster {
		ref, err := partySave.SAVGAMCharacterRef(fmt.Sprintf("CHRDAT%c%d", key, index+1))
		if err != nil {
			return err
		}
		container.CharacterRefs[index] = ref
	}
	prefix, err := partySave.EncodeSAVGAM(container)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(directory, ".savgam-save-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	write := func(name string, data []byte) error {
		return os.WriteFile(filepath.Join(stage, name), data, 0o600)
	}
	prefixName := fmt.Sprintf("savgam%c.dat", key+('a'-'A'))
	if err := write(prefixName, prefix); err != nil {
		return err
	}
	for index, character := range s.partyRoster {
		base := fmt.Sprintf("CHRDAT%c%d", key, index+1)
		raw, ok := s.savgamPlayers[character.ID]
		if !ok {
			return fmt.Errorf("no raw SAVGAM player record retained for %q", character.ID)
		}
		record, err := party.PatchDOSPlayerRecord(raw.Record, character)
		if err != nil {
			return fmt.Errorf("patch SAVGAM player %s: %w", base, err)
		}
		inventory, err := monster.EncodeItems(character.Equipment)
		if err != nil {
			return fmt.Errorf("encode SAVGAM inventory %s: %w", base, err)
		}
		if err := write(base+".sav", record); err != nil {
			return err
		}
		if err := write(base+".swg", inventory); err != nil {
			return err
		}
		if err := write(base+".fx", monster.EncodeAffects(character.Effects)); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(stage)
	if err != nil {
		return err
	}
	backup, err := os.MkdirTemp(directory, ".savgam-backup-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(backup)
	fileNames := []string{prefixName}
	for index := 1; index <= 6; index++ {
		base := fmt.Sprintf("CHRDAT%c%d", key, index)
		fileNames = append(fileNames, base+".sav", base+".swg", base+".fx")
	}
	backedUp := make([]string, 0, len(fileNames))
	for _, name := range fileNames {
		target := filepath.Join(directory, name)
		if _, statErr := os.Stat(target); os.IsNotExist(statErr) {
			continue
		} else if statErr != nil {
			return statErr
		}
		if err := os.Rename(target, filepath.Join(backup, name)); err != nil {
			for _, restored := range backedUp {
				_ = os.Rename(filepath.Join(backup, restored), filepath.Join(directory, restored))
			}
			return fmt.Errorf("backup SAVGAM file %s: %w", name, err)
		}
		backedUp = append(backedUp, name)
	}
	installed := make([]string, 0, len(entries))
	rollback := func() {
		for _, name := range installed {
			_ = os.Remove(filepath.Join(directory, name))
		}
		for _, name := range backedUp {
			_ = os.Rename(filepath.Join(backup, name), filepath.Join(directory, name))
		}
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(stage, entry.Name()), filepath.Join(directory, entry.Name())); err != nil {
			rollback()
			return fmt.Errorf("replace SAVGAM file %s: %w", entry.Name(), err)
		}
		installed = append(installed, entry.Name())
	}
	s.savgamPrefix = &container
	return nil
}

func readOptionalSAVGAMSidecar(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
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
	// ★ 前三塊就是 ECL 位址空間的區 0／1／2（spec 1163），所以 VM 記憶體要
	// 先寫回去，再讓 Area 編碼器覆蓋它自己那幾個具名欄位——那幾個欄位本來就是
	// 同一批位址，remake 這一側以 `s.Area` 為準。
	if err := s.encodeECLBanksInto(&container); err != nil {
		return partySave.SAVGAMContainer{}, err
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
	// 三組牆面參數由 `37h LOAD PIECES` 維護（spec 1153）。沒有跑過任何
	// `LOAD PIECES` 時保留匯入來的那一份，不要用零蓋掉原版存下來的值。
	if s.wallSetParams != ([3]partySave.SAVGAMSetBlock{}) {
		container.SetBlocks = s.wallSetParams
	}
	container.PartyCount = uint8(len(s.partyRoster))
	for index := range container.CharacterRefs {
		container.CharacterRefs[index] = nil
	}
	for index, character := range s.partyRoster {
		if index >= len(container.CharacterRefs) {
			break
		}
		ref, err := partySave.SAVGAMCharacterRef(character.Name)
		if err != nil {
			return partySave.SAVGAMContainer{}, err
		}
		container.CharacterRefs[index] = ref
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
	if err := s.SetPartyRoster(file.Characters); err != nil {
		return err
	}
	s.DungeonSearchEnabled = false
	s.dungeonSearchEdges = make(map[string]bool)
	if file.Version >= 12 {
		if err := s.restoreDungeonSearchState(file.DungeonSearch, file.DungeonSearchEdges); err != nil {
			return err
		}
	}
	if err := s.restoreJournalMessageIDs(file.JournalMessageIDs); err != nil {
		return err
	}
	if file.Version >= 5 {
		s.gameClock = file.GameTime
		s.gameAgeCycles = file.GameAgeCycles
	} else {
		s.gameClock = [7]uint16{}
		s.gameAgeCycles = 0
	}
	s.Area = file.Area
	s.gameClock = file.GameTime
	if file.GameTime == [7]uint16{} {
		s.gameClock = file.Area.GameTime
	}
	s.Area.GameTime = s.gameClock
	s.GeoMapSet = file.Area.GameArea
	s.GeoMapBlock = file.Area.Current3DMapBlockID
	// ★ 讀完存檔要**通知前端換地圖**。存檔記著自己是哪一張圖，但前端的 `geoGrid`
	// 只有在收到 `ConsumeGeoMapRequest` 時才會換——`geoMapPending` 原本只有 ECL 的
	// LOAD FILES 那條路會設。
	//
	// ⚠ 少了這一行，讀檔之後前端拿的還是**上一段的地圖**：牆、門、能不能走全部
	// 對不上，而且**不會有任何錯誤訊息**。前端一開始沒有地圖時更直接——
	// `moveDungeonPreview` 看到 `geoGrid == nil` 就 return，**按什麼都沒反應**
	// （spec 1191：114 份主線快照裡有 104 份是這樣）。
	s.geoMapPending = true
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
	if file.Location <= uint8(LocationMythDrannor) {
		s.Location = Location(file.Location)
	}
	// World-map saves store the original ECL selector in Area.CurrentCity.
	// Rebuild the name/source projection only for world-map modes; dungeon and
	// combat saves may retain a stale/default CurrentCity while Location is a
	// meaningful interior location, and must not be overwritten by selector 0.
	if file.Mode == uint8(ModeWilderness) || file.Mode == uint8(ModeMap) {
		if s.Location != LocationWilderness && file.Area.CurrentCity <= 13 {
			s.setWorldLocation(uint16(file.Area.CurrentCity))
		}
	}
	if file.Version == 1 {
		// Legacy party.json had no adventure-state fields.
		s.Mode = ModeWilderness
		s.Location = LocationWilderness
	} else if file.Mode <= uint8(ModeDungeon) {
		s.Mode = Mode(file.Mode)
	} else {
		s.Mode = ModeWilderness
	}
	if s.Mode == ModeMap && s.Location != LocationShadowdale {
		s.Mode = ModeWilderness
	}
	// 世界地圖模式上面已經由 `setWorldLocation` 連名字一起設好；其餘模式
	// （地城、戰鬥、事件、場所）只有 `Location` 被還原，名字還停在建檔時的值。
	if s.Mode != ModeWilderness && s.Mode != ModeMap {
		s.restoreLocationName()
	}
	// ★ 存檔沒有保存「事件結束要回到哪裡」（`eventReturnMode`）。存在事件畫面上的
	// 檔讀回來之後，那一格是零值，`Continue()` 會落到 default 分支回
	// 「event has no continuation」——**玩家按下一步就卡住，而且每一個欄位都對得上**。
	//
	// 這裡不是猜：回到哪裡由已經存下來的隊伍位置決定，跟活著的流程在那些點設的
	// 值一樣（地城裡回地城，否則回世界地圖）。
	if s.Mode == ModeEvent {
		s.eventReturnMode = ModeWilderness
		if s.Area.InDungeon {
			s.eventReturnMode = ModeDungeon
		}
	}
	if file.Version >= 6 && file.ECLSession != nil {
		if s.session == nil {
			return fmt.Errorf("game save contains an ECL session but no original ECL blocks are loaded")
		}
		if err := s.session.RestoreSnapshot(*file.ECLSession); err != nil {
			return fmt.Errorf("restore game ECL session: %w", err)
		}
		s.eclBlock = s.session.CurrentData()
		s.eclStart, err = s.session.InitialEntry()
		if err != nil {
			return err
		}
	}
	// ★ 讀檔之後地圖暫存器要跟隊伍位置一致。 原作的引擎每動一次位置就寫
	// `720Fh`／`7210h`／`7211h`，那三格因此永遠等於真實位置。remake 的讀檔
	// 路徑會做版本回退與範圍夾擠（上面那幾條 `= 7, 13, 0`），不同步的話
	// `2Dh CALL 2E10h` 的重畫會拿快照裡的舊座標去投影（spec 1172）。
	if s.session != nil {
		s.syncDungeonECLRegisters()
	}
	s.activeMusicTrackID = ""
	s.musicPlaybackSnapshot = nil
	s.oneShotPlaybackSnapshot = nil
	s.pendingMusicEvents = nil
	if file.OneShotAudio != nil {
		copy := audiostate.Clone(*file.OneShotAudio)
		s.oneShotPlaybackSnapshot = &copy
	}
	if file.Music != nil {
		if s.dataPack != nil {
			track, found := s.dataPack.FindMusicTrack(file.Music.TrackID)
			if !found {
				return fmt.Errorf("game save music track %q is not in the data pack", file.Music.TrackID)
			}
			if file.Music.Stream != nil && file.Music.Stream.Selector != int(track.ReferenceSelector) {
				return fmt.Errorf("game save music track %q selector %d does not match data-pack selector %d", file.Music.TrackID, file.Music.Stream.Selector, track.ReferenceSelector)
			}
		}
		s.activeMusicTrackID = file.Music.TrackID
		if file.Music.Stream != nil {
			copy := cloneTrackPCMStreamSnapshot(*file.Music.Stream)
			s.musicPlaybackSnapshot = &copy
		}
	} else if file.Version < 8 {
		// Legacy saves never recorded an active track. Restart the verified
		// binding for the restored ECL block instead of preserving a stale
		// pre-load frontend event or pretending an unknown sample position.
		if s.dataPack != nil && s.session != nil {
			if binding, found := s.dataPack.FindMusicBinding(s.session.CurrentBlockID(), ""); found {
				s.activeMusicTrackID = binding.TrackID
			}
		}
	}
	if file.Combat != nil {
		if s.Mode != ModeCombat {
			return fmt.Errorf("game save contains active combat while mode is %d", s.Mode)
		}
		if err := s.restoreActiveCombat(*file.Combat); err != nil {
			return fmt.Errorf("restore active combat: %w", err)
		}
		return nil
	}
	if s.Mode == ModeCombat {
		return fmt.Errorf("game save mode is combat but no active-combat snapshot is present")
	}
	// ⚠ 世界地圖的 hub 選單只有**回到世界地圖**時才對。
	//
	// 這裡原本是無條件設的，於是存在事件畫面上的檔讀回來會顯示
	// 「隊伍已建立．準備開始冒險．」＋「進入城市／繼續旅程／紮營」——
	// 玩家人在密斯卓諾墓園，畫面卻是剛建好隊伍的世界地圖選單。
	// 上面那段 `eventReturnMode` 的註解講的是同一類問題（存檔沒有保存畫面上的
	// 過場狀態），只是當時只補了「按下一步會不會卡住」，沒有補「畫面對不對」。
	//
	// ★ 用**真的前端**畫戰役檢查點才看得到：`*State` 層的測試讀完檔就直接做下一個
	// 動作，那兩個欄位在被看到之前就被覆蓋掉了（spec 1188）。
	if s.Mode == ModeEvent {
		s.Prompt = s.catalog.Text("press_enter", "Press Enter to continue")
		s.Choices = nil
		s.currentOriginalChoices = nil
		return nil
	}
	s.Prompt = s.catalog.Text("party_ready", "party_ready")
	s.Choices = []string{s.localizeOption("ENTER CITY"), s.localizeOption("JOURNEY ON"), s.localizeOption("CAMP")}
	s.currentOriginalChoices = []string{"ENTER CITY", "JOURNEY ON", "CAMP"}
	return nil
}

func (s *State) restoreJournalMessageIDs(messageIDs []string) error {
	s.journalMessageIDs = nil
	s.JournalPages = nil
	s.JournalPage = 0
	if len(messageIDs) == 0 {
		s.JournalText = s.catalog.Text("journal_empty", "journal_empty")
		return nil
	}
	if s.dataPack == nil {
		return fmt.Errorf("game save contains journal entries but no game pack is loaded")
	}
	for _, messageID := range messageIDs {
		page, found := s.dataPack.Text(messageID, s.catalog.Language)
		if !found {
			return fmt.Errorf("game save journal message ID %q is not in the data pack", messageID)
		}
		s.appendJournalPage(messageID, page)
	}
	s.JournalText = s.JournalPages[0]
	return nil
}

func (s *State) restoreActiveCombat(snapshot partySave.CombatSnapshot) error {
	for index := range snapshot.Battle.Fighters {
		fighter := &snapshot.Battle.Fighters[index]
		if fighter.SourceName == "" || s.dataPack == nil {
			continue
		}
		if localized, found := s.dataPack.LocalizeCombatantName(fighter.SourceName, s.catalog.Language); found {
			fighter.Name = localized
		} else {
			fighter.Name = fighter.SourceName
		}
	}
	battle, err := combat.RestoreBattle(snapshot.Battle)
	if err != nil {
		return err
	}
	if s.dataPack != nil {
		rules, err := s.dataPack.ResolveCombatAffectRules()
		if err != nil {
			return fmt.Errorf("resolve combat affect rules after load: %w", err)
		}
		battle.SetDamageRules(rules)
		conditionalRules, err := s.dataPack.ResolveCombatConditionalModifiers()
		if err != nil {
			return fmt.Errorf("resolve combat conditional modifiers after load: %w", err)
		}
		battle.SetConditionalModifierRules(conditionalRules)
		magicResistanceRules, err := s.dataPack.ResolveCombatMagicResistanceRules()
		if err != nil {
			return fmt.Errorf("resolve combat magic resistance rules after load: %w", err)
		}
		battle.SetMagicResistanceRules(magicResistanceRules)
		postHitRules, err := s.dataPack.ResolveCombatPostHitRules()
		if err != nil {
			return fmt.Errorf("resolve combat post-hit rules after load: %w", err)
		}
		battle.SetPostHitRules(postHitRules)
		monsterSpellRules, err := s.dataPack.ResolveCombatMonsterSpellRules()
		if err != nil {
			return fmt.Errorf("resolve combat monster spell rules after load: %w", err)
		}
		battle.SetMonsterSpellRules(monsterSpellRules)
	}
	if snapshot.TurnIndex < 0 || snapshot.TurnIndex > len(snapshot.Turns) {
		return fmt.Errorf("combat turn index %d outside 0..%d", snapshot.TurnIndex, len(snapshot.Turns))
	}
	for index, turn := range snapshot.Turns {
		if _, ok := battle.Fighter(turn.FighterID); !ok {
			return fmt.Errorf("combat turn %d references missing fighter %q", index, turn.FighterID)
		}
	}
	for index, delayed := range snapshot.DelayedTurns {
		if index < 0 || index >= len(snapshot.Turns) || !delayed {
			return fmt.Errorf("invalid delayed-turn entry %d=%v", index, delayed)
		}
	}
	if snapshot.TargetIndex < 0 || snapshot.SpellTargetIndex < 0 || snapshot.MoveRemaining < 0 {
		return fmt.Errorf("negative combat cursor or movement state")
	}
	if snapshot.CastingClass > uint8(party.ClassThief) || snapshot.Speed > uint8(engineaction.SlowestSpeed) {
		return fmt.Errorf("invalid combat class %d or speed %d", snapshot.CastingClass, snapshot.Speed)
	}
	if snapshot.ReturnMode > uint8(ModeDungeon) || snapshot.ReturnMode == uint8(ModeCombat) {
		return fmt.Errorf("invalid combat return mode %d", snapshot.ReturnMode)
	}
	if snapshot.ViewFighterID != "" {
		if _, ok := battle.Fighter(snapshot.ViewFighterID); !ok {
			return fmt.Errorf("combat view references missing fighter %q", snapshot.ViewFighterID)
		}
	}
	if snapshot.Visual != nil {
		elapsed := time.Duration(snapshot.VisualElapsedNanos)
		if snapshot.VisualElapsedNanos < 0 || elapsed > snapshot.Visual.Duration() {
			return fmt.Errorf("combat visual elapsed time %d outside 0..%d", snapshot.VisualElapsedNanos, snapshot.Visual.Duration())
		}
		if snapshot.VisualImpactSent < -1 || snapshot.VisualImpactSent >= snapshot.Visual.ImpactCount() {
			return fmt.Errorf("combat visual impact marker %d outside -1..%d", snapshot.VisualImpactSent, snapshot.Visual.ImpactCount()-1)
		}
		if snapshot.VisualDeathSent < -1 || snapshot.VisualDeathSent >= snapshot.Visual.ImpactCount() {
			return fmt.Errorf("combat visual death marker %d outside -1..%d", snapshot.VisualDeathSent, snapshot.Visual.ImpactCount()-1)
		}
		if snapshot.Visual.ActorID != "" {
			if _, ok := battle.Fighter(snapshot.Visual.ActorID); !ok {
				return fmt.Errorf("combat visual references missing actor %q", snapshot.Visual.ActorID)
			}
		}
		for index := 0; index < snapshot.Visual.ImpactCount(); index++ {
			impact, _ := snapshot.Visual.ImpactAt(index)
			if impact.TargetID == "" {
				continue
			}
			if _, ok := battle.Fighter(impact.TargetID); !ok {
				return fmt.Errorf("combat visual impact %d references missing target %q", index, impact.TargetID)
			}
		}
	} else if snapshot.VisualElapsedNanos != 0 {
		return fmt.Errorf("combat visual elapsed time %d has no visual event", snapshot.VisualElapsedNanos)
	}
	s.battle = battle
	s.combatTurns = append([]combat.Turn(nil), snapshot.Turns...)
	s.combatTurnIndex = snapshot.TurnIndex
	s.combatDelayedTurns = cloneDelayedTurns(snapshot.DelayedTurns)
	s.combatTargetIndex = snapshot.TargetIndex
	s.combatCastingSpell = snapshot.CastingSpell
	s.combatCastingClass = party.Class(snapshot.CastingClass)
	s.combatCastingClassSet = snapshot.CastingClassSet
	s.combatSpellTargetIndex = snapshot.SpellTargetIndex
	s.combatSpellTargetPoint = snapshot.SpellTargetPoint
	s.combatSpellTargetsPoint = snapshot.SpellTargetsPoint
	s.combatMoveMode, s.combatMoveRemaining = snapshot.MoveMode, snapshot.MoveRemaining
	s.combatSpeed = engineaction.Speed(snapshot.Speed)
	s.combatQuickMagic, s.combatReferenceCoords = snapshot.QuickMagic, snapshot.ReferenceCoords
	s.combatView, s.combatViewFighterID = snapshot.View, snapshot.ViewFighterID
	s.combatMessage, s.combatReturnMode = snapshot.Message, Mode(snapshot.ReturnMode)
	s.combatVisualSerial, s.combatVisualEnabled = snapshot.VisualSerial, snapshot.VisualEnabled
	if snapshot.Visual != nil {
		copy := *snapshot.Visual
		copy.Impacts = append([]combat.VisualImpactTarget(nil), copy.Impacts...)
		copy.Segments = append([]combat.VisualPathSegment(nil), copy.Segments...)
		s.combatVisual = &copy
	} else {
		s.combatVisual = nil
	}
	s.combatVisualElapsed = time.Duration(snapshot.VisualElapsedNanos)
	s.combatVisualTravelSent = snapshot.VisualTravelSent
	s.combatVisualImpactSent = snapshot.VisualImpactSent
	s.combatVisualDeathSent = snapshot.VisualDeathSent
	s.combatVisualAdvanceTurn = snapshot.VisualAdvanceTurn
	return nil
}

// LoadDOSCharacterFiles imports one original DOS character bundle directly
// into the running remake. It remains useful for inspecting an individual
// record without loading a complete SAVGAM slot.
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
	s.Prompt = s.catalog.Text("party_ready", "party_ready")
	s.Choices = []string{s.localizeOption("ENTER CITY"), s.localizeOption("JOURNEY ON"), s.localizeOption("CAMP")}
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

// wallPiecesFromParams 把存檔的三組牆面參數換回 `LOAD PIECES` 的三元組形狀，
// 哨兵 `0FFFFh` 對回 `0FFh`（「這一槽不用」）。全部都是哨兵就當作沒有資訊。
func wallPiecesFromParams(params [3]partySave.SAVGAMSetBlock) ([3]uint16, bool) {
	var pieces [3]uint16
	useful := false
	for index, param := range params {
		if param.BlockID == 0xFFFF || param.BlockID == 0 {
			pieces[index] = 0xFF
			continue
		}
		pieces[index] = param.BlockID
		useful = true
	}
	return pieces, useful
}

// locationDisplayKeys 是 `Location` → 語系鍵／原文。
//
// ★ 為什麼要有這一份：讀檔時 `Location` 有還原，**`LocationName` 沒有**，
// 於是畫面標題永遠是建檔當下的那一個（新遊戲是「荒野」）。原本的還原路徑
// 走 `setWorldLocation(Area.CurrentCity)`，但那一支只在世界地圖模式跑——
// 上面那段註解說明了理由：地城與戰鬥存檔的 `CurrentCity` 可能是舊值或 0，
// 拿它去覆寫會把有意義的 `Location` 蓋掉。
//
// ⚠ 所以這裡**只推顯示名稱、不碰 `Location`**：hazard 是「覆寫 Location」，
// 而這一份不寫 `Location`，所以那個 hazard 在結構上不成立。
var locationDisplayKeys = map[Location][2]string{
	LocationWilderness:  {"wilderness", "Wilderness"},
	LocationShadowdale:  {"shadowdale", "SHADOWDALE"},
	LocationAshabenford: {"ashabenford", "ASHABENFORD"},
	LocationDaggerFalls: {"dagger_falls", "DAGGER FALLS"},
	LocationTilverton:   {"tilverton", "tilverton"},
	LocationStandingStone: {"standing_stone", "standing_stone"},
	LocationEssembra:    {"essembra", "essembra"},
	LocationHap:         {"hap", "hap"},
	LocationVoonlar:     {"voonlar", "VOONLAR"},
	LocationPhlan:       {"phlan", "PHLAN"},
	LocationTeshwave:    {"teshwave", "TESHWAVE"},
	LocationYulash:      {"yulash", "YULASH"},
	LocationHillsfar:    {"hillsfar", "HILLSFAR"},
	LocationZhentilKeep: {"zhentil_keep", "ZHENTIL KEEP"},
	LocationMythDrannor: {"myth_drannor", "MYTH DRANNOR"},
}

// restoreLocationName 由已經還原好的 `Location` 推出畫面上的地名。
func (s *State) restoreLocationName() {
	keys, ok := locationDisplayKeys[s.Location]
	if !ok {
		return
	}
	s.LocationName = s.catalog.Text(keys[0], keys[1])
	s.OriginalLocation = keys[1]
}
