package ecl

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/wicanr2/golden-box-remake-engine/randomstream"
)

// ECL 位址空間不是平坦記憶體：原作把絕對位址分成五區，前三區各自映射到一塊
// bank，bank 內位移是 (位址 − 區基底) × 2（spec 1096）。
//
//	區 0  4B00h..4EFFh -> bank0   word
//	區 1  7C00h..7FFFh -> bank1   word
//	區 2  7A00h..7BFFh -> bank2   word
//	區 3  8000h..9DFFh -> ECL 程式碼本身，byte（自我修改）
//	區 4  其他          -> 特例表，未列出的一律靜默丟棄
//
// bank 是引擎與 ECL 的共用記憶體，不是訊息傳遞。只被 ECL 自己讀寫的格子怎麼放
// 都自洽；被引擎也讀寫的格子（spec 1097 清出 24 個）必須逐格對上，否則錯誤是
// 靜默的——寫入永遠成功，只是指向了別的地方。
//
// 下列常數是本套件用到的 ECL 位址，附原作 bank 欄位對照，方便逐格查核。
const (
	// 引擎寫入、24h 讀取並清零的請求旗標（spec 1097 §四）。
	addrShopRequest    uint16 = 0x7F6C // bank1^[6D8h] 商店
	addrShopPriceScale uint16 = 0x7F6D // bank1^[6DAh] 商店價格倍率
	addrTempleRequest  uint16 = 0x7EE2 // bank1^[5C4h] 神殿（overlay-04 ＝ TEMPLE，spec 1182）

	// 區 1 低段的一部分位址不是 bank1 記憶體，而是「目前角色」記錄的欄位投影：
	// 讀取端 overlay-07:007F1h 會攔截這些位址並改讀 DS:6506h 指向的角色
	// （對照表見 spec 1040，機制見 spec 1098）。LOAD CHARACTER 換人之後，
	// 同一個 ECL 位址就該讀到新角色的值，所以下列位址每次換人都要重新投影。
	addrPlayerControlMorale uint16 = 0x7CB8 // 角色 +0F7h，>=80h 是 NPC（spec 1066）
	addrPlayerFlag192       uint16 = 0x7CE4 // 角色 +192h，讀取端只取 and 1
)

// RunResult is the observable output of the bounded ECL subset runner.
// It deliberately exposes text and stop position, not DOS rendering state.
type RunResult struct {
	Text             []string
	Menus            []Menu
	PC               int
	Steps            int
	Exited           bool
	WaitingForMenu   bool
	WaitingForWho    bool
	WaitingForString bool
	NewECLBlockID    *uint8
	CombatRequested  bool
	ShopRequested    bool
	// PostCombatRequested ＝ `24h` 走了第四支（沒怪、沒商店、沒神殿）：
	// 原作跑 `overlay-05` 的 `DOPOSTCOMBAT`，也就是戰利品分配／經驗值畫面
	// （spec 1182）。
	PostCombatRequested bool
	ShopPriceScale      uint16
	TempleRequested     bool
	MonsterSetup        *MonsterSetup
	MonsterSpawns       []MonsterSpawn
	ProgramIDs          []uint8
	ProgramExit         bool
	CallAddresses       []uint16
	CallRequests        []CallRequest
	SaveWrites          []MemoryWrite
	// ClearBoxRequested 來自 `3Dh CLEAR BOX`：把文字框清空，且不印新文字。
	ClearBoxRequested bool
	// SpriteOffRequested 來自 `31h SPRITE OFF`：關掉畫面上的怪物圖示。
	SpriteOffRequested     bool
	SessionStartBlockID    uint8
	SessionEndBlockID      uint8
	SessionBlockRangeSet   bool
	// SessionRanBlockIDs 是這一次頂層執行實際跑過的 block（含 NEWECL 鏈上的
	// 每一段，依執行順序）。座標投影靠它判斷「這個座標是本次執行的腳本寫的」
	// ——原作的地圖暫存器是全域，跨 NEWECL 存活，來源段在交接前寫好的落點
	// 在目的段生效（spec 1183／1184）。
	SessionRanBlockIDs []uint8
	DamageRequests         []DamageRequest
	PrintReturnCount       int
	ApproachCount          int
	DelayCount             int
	LoadCharacterAddresses []uint16
	LoadCharacterRequests  []LoadCharacterRequest
	CombatTeamWrites       []CombatTeamWrite
	FindItemIDs            []uint16
	FindItemRequests       []FindItemRequest
	FindSpecialRequests    []FindSpecialRequest
	DumpRequests           []DumpRequest
	DestroyItemIDs         []uint16
	NPCIDs                 []uint16
	NPCRequests            []NPCRequest
	SelectionsConsumed     int
	WhoSelectionsConsumed  int
	StringInputsConsumed   int
	RandomValues           []uint16
	EncounterActions       []uint16
	LoadFilesRequested     bool
	// LoadFilesLoaded3DMap ＝ 這一條 `21h` 真的走了載 3D 地圖那一路
	// （`o[1]` 不是 `0FFh`／`7Fh`，而且 `bank0^[1CCh]` 非零）。為 false 時
	// 原作走的是載大圖那一路（spec 1181）。
	LoadFilesLoaded3DMap bool
	LoadFiles            [3]uint16
	LoadPiecesRequested  bool
	LoadPieces           [3]uint16
	// WallSetAssignments 是這一條 `37h` **實際**要做的事。原作的 handler 自己
	// 分三支（spec 1087／1153），所以分支結果由 VM 決定、不是上層猜：沒被列出來的
	// 槽這一次**完全不動**（`7Fh` 那一支就只碰槽 1）。
	WallSetAssignments []WallSetAssignment
	// FinalView 是**這一次頂層執行結束當下**的畫面鏡射。
	//
	// ★ 為什麼要它。 原作的 `STOREVALUE` 一寫 `C04B`／`C04C`／`C04D` 就**當場**
	// 改 `720Fh`／`7210h`／`7211h`——隊伍在那一刻就已經在新格子上了，`2E10h` 只是
	// 重畫。remake 的投影掛在 `2E10h` 上，所以「寫了座標卻沒重畫」的執行會讓兩邊
	// 分岔：原作的隊伍搬走了，remake 的還在原地，而髒旗標會跨執行留著
	// （spec 1172）。收尾投影這一格就是補上那個時間差。
	FinalView        ViewMirror
	PictureRequested bool
	// PictureCloseRequested 為真代表 `0Eh PICTURE` 的運算元是 `0FFh`：把圖關掉。
	PictureCloseRequested bool
	// PictureCloseRedraw 為真代表這次關閉走到了原作 `08E9h` 的**立即重繪**：
	// 圖真的開著（`8B62h` 或 `8B65h` 非 0）而且不在「前後都還在第一人稱」的
	// 旁路上（`not (4FBBh = 4 and 4FBAh = 4)`，spec 1148）。走旁路時原作不重繪
	// 也不清那兩格——畫面等主迴圈自己把 3D 視窗畫回去。
	PictureCloseRedraw bool
	// PictureFrameAdvances 是**最後一張圖之後**跑過幾次 `2Dh CALL 6803h`。
	// 原作那一支把圖片序列的游標往前推一格（超過張數回到第 1 格）、畫出那一格
	// 再等一個 GAMEDELAY；換圖時 LOADSEQUENCE 會把游標設回第 1 格，所以換圖
	// 之前推的格數不算（spec 1150）。
	PictureFrameAdvances int
	PictureBlock         uint16
	BigPictureRequested  bool
	PictureHeadBlock     uint16
	PictureHeadBlockSet  bool
	SpellSearches        []SpellSearch
	ProtectionRequests   []uint16
	ClockRequests        []ClockRequest
	TreasureRequests     []TreasureRequest
	// ClearMonstersRequested 為真代表這一次執行跑過 `1Ch CLEARMONSTERS`。
	// 上層要據此把**跨執行累積**的戰利品堆一起丟掉——`result` 裡的那一份已經
	// 在指令當下清掉了。
	ClearMonstersRequested bool
	RobRequests            []RobRequest
	PartyStrengthRequests  []PartyStrengthRequest
	PartySurpriseRequests  []PartySurpriseRequest
	CheckPartyRequests     []CheckPartyRequest
	WhoRequests            []WhoRequest
	StringInputRequests    []StringInputRequest
}

// ECL `2Dh CALL` 的運算元是 external routine 的選擇子（`selector = 運算元 −
// 7FFFh`，spec 561）。原作的分派器認得七個，CoAB 的腳本用到四個：
// `2E10h` 重畫、`C01Eh` 前進一格、`B200h` 播音效、`6803h` 推圖片序列一格。
// `6803h` 與 `2E10h` 在 remake 這一側各有自己的狀態要記（spec 1150）。
const (
	ExternalCallAdvancePictureFrame = 0x6803
	// ExternalCallRedrawView 是「髒了才重畫」那一支。它**不是無條件重畫**：
	// 原作先看那五個髒旗標，重畫之後再把它們清掉。
	ExternalCallRedrawView = 0x2E10
	// ExternalCallMoveForward 是強制前進一格（原作那一支不做碰撞判斷）。
	// 它**當場**改地圖暫存器，所以鏡射要在 VM 裡跟著走。
	ExternalCallMoveForward = 0xC01E
)

// MemoryWrite preserves one numeric SAVE/SAVE TABLE side effect from the
// current bounded transaction. The VM owns memory mutation; adapters use this
// evidence to distinguish fresh script assignments from stale shared values.
type MemoryWrite struct {
	Address uint16
	Value   uint16
	PC      int
	BlockID uint8
	// Sequence 是這次執行裡的**執行序**，`CallRequest.Sequence` 用同一條計數。
	//
	// ⚠ 不要拿 `PC` 當執行順序用。一次執行裡有迴圈與反向跳躍，PC 小的可能
	// 後執行、同一個位址也可能被執行好幾次——「`PC` 比 `CALL` 小就是先發生」
	// 只在直線碼上成立。
	Sequence int
}

// CallRequest preserves the bytecode position and block identity of one
// external CALL while CallAddresses remains the compatibility projection.
type CallRequest struct {
	Address uint16
	PC      int
	BlockID uint8
	// Sequence 與 MemoryWrite.Sequence 共用同一條執行序計數。
	Sequence int
	// View 是這條 `CALL` 執行的那一刻，`720Fh`／`7210h`／`7211h` 與五個髒旗標
	// 的值（spec 1150）。`2E10h` 的呼叫端照它決定要不要投影座標——**不要**再
	// 回頭掃 `SaveWrites`，那個視窗跨不了執行也跨不了 block。
	View ViewMirror
}

// NPCRequest preserves both operands consumed by CMD_AddNPC. Morale is later
// converted by the game adapter to (morale >> 1) + Control.NPC_Base.
type NPCRequest struct {
	ID     uint16
	Morale uint16
}

type PartyStrengthRequest struct {
	Destination uint16
	Value       uint16
	Resolved    bool
}

type CheckPartyRequest struct {
	Query        uint16
	AffectID     uint16
	Destinations [4]uint16
	Minimum      uint16
	Maximum      uint16
	Average      uint16
	AffectFound  bool
	Resolved     bool
}

// LoadCharacterRequest is the decoded boundary for reference LOAD CHARACTER.
// Address preserves the raw word operand; Value is vm_GetCmdValue's result.
// The low 7 bits select the zero-based TeamList member and bit 7 is the reference
// restore/redraw flag.
type LoadCharacterRequest struct {
	Address     uint16
	Value       uint16
	PlayerIndex uint8
	HighBitSet  bool
}

// CombatTeamWrite preserves SAVE to SelectedPlayer's combat-team field.
type CombatTeamWrite struct {
	TeamListIndex int
	Value         uint16
}

// FindItemRequest records the reference party-wide item query. Resolved is
// false when the bounded VM has no injected party inventory context.
type FindItemRequest struct {
	ItemID   uint16
	Found    bool
	Resolved bool
}

// FindSpecialRequest records FIND SPECIAL against the current selected
// player's active affects.
type FindSpecialRequest struct {
	AffectID            uint16
	SelectedPlayerIndex int
	Found               bool
	Resolved            bool
}

// DumpRequest records removal of the currently selected party member and the
// reference fallback selection after that removal.
type DumpRequest struct {
	SelectedPlayerIndex     int
	NextSelectedPlayerIndex int
	NextSelectedPlayerSet   bool
	Resolved                bool
}

// WhoRequest marks the reference character-selection boundary. WHO consumes
// the current ECL prompt text but its player selection belongs to the UI/state
// adapter rather than a normal HORIZONTAL/VERTICAL MENU.
type WhoRequest struct {
	Prompt            string
	Selected          uint16
	SelectionProvided bool
}

// StringInputRequest is the resumable UI boundary emitted by INPUT STRING.
// The script owns the maximum length and destination; the frontend only edits
// and submits text. Value is normalized to the uppercase vocabulary used by
// verified ECL string literals before it is written to RuntimeState.Strings.
type StringInputRequest struct {
	MaxLength     uint16
	Destination   uint16
	Value         string
	InputProvided bool
}

// PartySurpriseRequest preserves the two destination addresses used by the
// reference PARTY SURPRISE command. The party/ranger calculation belongs to
// the game adapter because the bounded VM has no roster context.
type PartySurpriseRequest struct {
	RangerDestination uint16
	OtherDestination  uint16
	RangerValue       uint16
	OtherValue        uint16
	Resolved          bool
}

// PartyMemberContext contains only the values needed by the verified party
// commands. It deliberately avoids importing party/combat packages so the ECL
// VM remains reusable by other Gold Box works.
type PartyMemberContext struct {
	Name          string
	ControlMorale uint8
	// ECLFlag192 是角色記錄 +192h，透過投影位址 7CE4h 讀（spec 1098 §五）。
	ECLFlag192        uint8
	ItemTypes         []uint8
	HitPoints         int
	ArmorClass        int
	AttackBonus       int
	ClericLevel       int
	MagicUserLevel    int
	HasRangerClass    bool
	ThiefSkills       [8]uint8
	MovementAllowance int
	Effects           []uint8
	// SpellSlots 是這名隊員依序記憶的法術 ID，供 `3Bh SPELL` 搜尋。
	SpellSlots []uint8
}

type PartyContext struct {
	Members []PartyMemberContext
}

func (c PartyContext) clone() PartyContext {
	owned := PartyContext{Members: append([]PartyMemberContext(nil), c.Members...)}
	for index := range owned.Members {
		owned.Members[index].ItemTypes = append([]uint8(nil), owned.Members[index].ItemTypes...)
		owned.Members[index].Effects = append([]uint8(nil), owned.Members[index].Effects...)
		owned.Members[index].SpellSlots = append([]uint8(nil), owned.Members[index].SpellSlots...)
	}
	return owned
}

func (c PartyContext) partyStrength() uint16 {
	var value int
	for _, member := range c.Members {
		armorClass := member.ArmorClass
		if armorClass > 60 {
			armorClass -= 60
		} else {
			armorClass = 0
		}
		hitBonus := member.AttackBonus
		if hitBonus > 39 {
			hitBonus -= 39
		} else {
			hitBonus = 0
		}
		value += ((member.ClericLevel * 4) + member.HitPoints + (armorClass * 5) + (hitBonus * 5) + (member.MagicUserLevel * 8)) / 10
	}
	if value < 0 {
		return 0
	}
	if value > 0xFF {
		value = 0xFF
	}
	return uint16(value)
}

func (c PartyContext) hasRanger() bool {
	for _, member := range c.Members {
		if member.HasRangerClass {
			return true
		}
	}
	return false
}

func (c PartyContext) checkParty(query, affectID uint16) (CheckPartyRequest, bool) {
	request := CheckPartyRequest{Query: query, AffectID: affectID}
	normalized := query - 0x7FFF
	if normalized == 8001 {
		for _, member := range c.Members {
			for _, effect := range member.Effects {
				if uint16(effect) == affectID {
					request.AffectFound = true
				}
			}
		}
		request.Resolved = true
		return request, true
	}
	if normalized >= 0xA5 && normalized <= 0xAC {
		index := int(normalized - 0xA5)
		request.Minimum, request.Maximum, request.Average = partyMetric(c.Members, func(member PartyMemberContext) int {
			return int(member.ThiefSkills[index])
		})
		request.Resolved = true
		return request, true
	}
	if normalized == 0x9F {
		request.Minimum, request.Maximum, request.Average = partyMetric(c.Members, func(member PartyMemberContext) int {
			return member.MovementAllowance
		})
		request.Resolved = true
		return request, true
	}
	return request, false
}

func partyMetric(members []PartyMemberContext, metric func(PartyMemberContext) int) (uint16, uint16, uint16) {
	if len(members) == 0 {
		return 0, 0, 0
	}
	minimum, maximum, total := metric(members[0]), metric(members[0]), 0
	for _, member := range members {
		value := metric(member)
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
		total += value
	}
	return uint16(minimum), uint16(maximum), uint16(total / len(members))
}

// TreasureRequest preserves the eight raw TREASURE operands. The first seven
// values are Copper, Silver, Electrum, Gold, Platinum, Gems and Jewelry;
// ItemBlock selects an ITEM{area}.DAX stock block, or a special/random branch.
type TreasureRequest struct {
	Coins     [7]uint16
	ItemBlock uint16
}

// spellNotFound 是 `3Bh SPELL` 在找不到時寫進 slot 位址的值。
// 依據是呼叫端：`COMPARE [slot], imm FFh` ＋ `IF <>` ＋ `GOTO 找到了`。
const spellNotFound = 0xFF

// findSpell 依行軍順序、再依該角色的法術槽順序找第一個持有者，
// 與 `party.Roster.FindSpell` 的合約一致。
func (c PartyContext) findSpell(spellID uint16) (slot, member uint16, found bool) {
	if spellID > 0xFF {
		return 0, 0, false
	}
	for memberIndex, m := range c.Members {
		for slotIndex, known := range m.SpellSlots {
			if uint16(known) == spellID {
				return uint16(slotIndex), uint16(memberIndex), true
			}
		}
	}
	return 0, 0, false
}

// RobRequest preserves ROB's party scope, percentage removed and per-item
// theft chance. SelectedPlayerIndex is meaningful only when AllParty is false.
type RobRequest struct {
	AllParty            bool
	LossPercent         uint16
	ItemChance          uint16
	SelectedPlayerIndex int
	SelectedPlayerSet   bool
}

// ClockRequest is the raw two-operand signal emitted by ECL CLOCK. The game
// adapter owns the clock and effect expiration; the VM only decodes it.
type ClockRequest struct {
	TimeStep uint16
	TimeSlot uint16
}

// DamageRequest 保存 `2Eh DAMAGE` 的五個運算元。規則由 party／combat adapter
// 執行，VM 這一層不擲骰也不改 HP。
//
// 五個運算元的語意逐條讀自 DOS `overlay-02:2942h`（spec 1152）：
//
//	Flags     bit 7 清空時**整個 byte 是次數**：連續打 N 下，每下各自
//	          隨機挑一名隊員並用 SaveFlags 當攻擊值擲 `TRYTOHIT`，
//	          而且**每下之間重擲傷害**。
//	          bit 7 設定時才是旗標：bit 6 ＝ 全隊、bit 5 ＝ 不擲豁免、
//	          bit 4 ＝ 豁免成功仍吃**全額**傷害、bit 0..4 ＝ 豁免調整值。
//	          ⚠ bit 4 同時屬於調整值欄位（原作是 `and 1Fh`）；全 corpus
//	          24 處的低 5 位一律是 0，所以實務上它只是旗標。
//	DiceCount／DiceSize／Bonus  傷害 ＝ `DiceCount d DiceSize ＋ Bonus`。
//	SaveFlags bit 7 ＝ 目標是**目前角色**（`DS:6506h`）；bit 0..2 ＝ 豁免種類。
//	          ⚠ 目前角色那一路傳給 `MAKESAVE` 的是 `(SaveFlags and 7) − 1`，
//	          而且 `= 0` 代表**不擲豁免**；全隊與隨機那兩路用的是沒有減一的值，
//	          要不要擲改由 Flags bit 5 決定。兩種讀法在 corpus 上結果相同，
//	          只有把 handler 讀完才分得出來。
type DamageRequest struct {
	Flags     uint16
	DiceCount uint16
	DiceSize  uint16
	Bonus     uint16
	SaveFlags uint16
	// SelectedPlayerIndex 是**這一條指令執行當下**選定的角色（原作的
	// `DS:6506h`），`SelectedPlayerSet` 為 false 時沒有意義。
	//
	// ★ 為什麼要蓋在封包上而不是事後問狀態。 腳本的慣用法是
	// `0Ah LOAD CHARACTER` 緊接一條 `2Eh DAMAGE`，而且會整個包在
	// 「逐一走過隊伍」的迴圈裡（`ECL5.DAX/0x32:0223h` 配 `7F3Eh` 隊伍人數）。
	// 一次執行會累積好幾組，事後只看最後一次選的人會把整批傷害算到同一位身上。
	SelectedPlayerIndex int
	SelectedPlayerSet   bool
}

// SpellSearch is the data-bearing part of ECL SPELL. The bounded runner keeps
// the requested spell and destination addresses; a party spell-slot resolver
// can later fill the reference result values.
type SpellSearch struct {
	SpellID          uint16
	SpellSlotAddress uint16
	CharacterAddress uint16
	// Resolved 表示這次搜尋有隊伍資料可查；沒有 party context 時 VM 不寫回，
	// 因為「找不到」與「不知道」在 ECL 那一側長得一樣（都會走進錯誤分支）。
	Resolved    bool
	Found       bool
	SlotIndex   uint16
	MemberIndex uint16
}

type Menu struct {
	Location uint16
	Options  []string
	Selected uint16
	Vertical bool
	Prompt   string
}

// RuntimeState is the resumable portion of the bounded ECL VM. A menu pause
// must not discard the memory writes, comparison flags, or call stack that led
// to it; otherwise feeding the next choice replays the event from its entry
// with a different machine state.
type RuntimeState struct {
	PC                  int
	Started             bool
	Stack               []int
	Memory              map[uint16]uint16
	Strings             map[uint16]string
	Compare             [6]bool
	SelectedPlayerIndex int
	SelectedPlayerSet   bool
	MonsterSetup        *MonsterSetup
	MonsterSpawns       []MonsterSpawn
	Random              *randomstream.Stream
	// View 是 `720Fh`／`7210h`／`7211h` 與五個髒旗標的鏡射（spec 1150）。
	// 它跟著 shared state 走，所以跨 block 也保留——原作那幾格就是全域。
	View ViewMirror
	// CurrentBlock 是正在執行的 block，`ViewMirror` 用它記「這筆座標是誰寫的」。
	CurrentBlock uint8
}

func NewRuntimeState(start int) *RuntimeState {
	return &RuntimeState{
		PC:      start,
		Memory:  make(map[uint16]uint16),
		Strings: make(map[uint16]string),
	}
}

// RunSubset executes only commands whose semantics are represented here.
// Unsupported commands return an error at their exact payload offset. This is
// useful for proving an event prefix without silently treating the whole ECL
// program as implemented.
func RunSubset(block []byte, start, maxSteps int) (RunResult, error) {
	return runSubset(block, start, maxSteps, nil, false, 1)
}

// RunSubsetWithSelections is RunSubset with deterministic selections for
// successive HORIZONTAL MENU commands. A missing selection keeps the safe
// default index 0. This is the bridge between ECL menu semantics and a UI;
// it does not pretend to implement the original blocking DOS input routine.
func RunSubsetWithSelections(block []byte, start, maxSteps int, selections []uint16) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, false, 1)
}

// RunSubsetWithSelectionsSeed is the deterministic variant used when an ECL
// event contains RANDOM. It keeps the legacy wrapper behavior stable while
// allowing a game session or regression to choose its own replay seed.
func RunSubsetWithSelectionsSeed(block []byte, start, maxSteps int, selections []uint16, seed int64) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, false, seed)
}

// RunSubsetInteractive pauses at the first menu whose selection is not
// supplied. This allows a UI to feed one choice per frame/event instead of
// silently choosing index zero for all later menus.
func RunSubsetInteractive(block []byte, start, maxSteps int, selections []uint16) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, true, 1)
}

// RunSubsetInteractiveSeed is the seeded interactive runner for ECL RANDOM.
func RunSubsetInteractiveSeed(block []byte, start, maxSteps int, selections []uint16, seed int64) (RunResult, error) {
	return runSubset(block, start, maxSteps, selections, true, seed)
}

// RunSubsetInteractiveSeedWithPartyContext resolves verified party commands
// against the supplied roster while preserving the seeded interactive API.
func RunSubsetInteractiveSeedWithPartyContext(block []byte, start, maxSteps int, selections []uint16, seed int64, context PartyContext) (RunResult, error) {
	owned := context.clone()
	return runSubsetWithStateContext(block, start, maxSteps, selections, true, seed, NewRuntimeState(start), &owned)
}

func RunSubsetInteractiveSeedWithPartyContextAndWhoSelections(block []byte, start, maxSteps int, selections, whoSelections []uint16, seed int64, context PartyContext) (RunResult, error) {
	owned := context.clone()
	return runSubsetWithStateContextAndInputs(block, start, maxSteps, selections, whoSelections, nil, true, seed, NewRuntimeState(start), &owned)
}

func runSubset(block []byte, start, maxSteps int, selections []uint16, pauseOnMissing bool, seed int64) (RunResult, error) {
	return runSubsetWithState(block, start, maxSteps, selections, pauseOnMissing, seed, NewRuntimeState(start))
}

func runSubsetWithState(block []byte, start, maxSteps int, selections []uint16, pauseOnMissing bool, seed int64, runtime *RuntimeState) (RunResult, error) {
	return runSubsetWithStateContext(block, start, maxSteps, selections, pauseOnMissing, seed, runtime, nil)
}

func runSubsetWithStateContext(block []byte, start, maxSteps int, selections []uint16, pauseOnMissing bool, seed int64, runtime *RuntimeState, partyContext *PartyContext) (RunResult, error) {
	return runSubsetWithStateContextAndInputs(block, start, maxSteps, selections, nil, nil, pauseOnMissing, seed, runtime, partyContext)
}

func runSubsetWithStateContextAndWhoSelections(block []byte, start, maxSteps int, selections, whoSelections []uint16, pauseOnMissing bool, seed int64, runtime *RuntimeState, partyContext *PartyContext) (RunResult, error) {
	return runSubsetWithStateContextAndInputs(block, start, maxSteps, selections, whoSelections, nil, pauseOnMissing, seed, runtime, partyContext)
}

func runSubsetWithStateContextAndInputs(block []byte, start, maxSteps int, selections, whoSelections []uint16, stringInputs []string, pauseOnMissing bool, seed int64, runtime *RuntimeState, partyContext *PartyContext) (RunResult, error) {
	if len(block) < 2 {
		return RunResult{}, fmt.Errorf("ECL block is shorter than two-byte prefix")
	}
	payload := block[2:]
	if start < 0 || start >= len(payload) {
		return RunResult{}, fmt.Errorf("runtime start %d is outside payload", start)
	}
	if maxSteps <= 0 {
		return RunResult{}, fmt.Errorf("runtime step limit must be positive")
	}

	pc := start
	stack := make([]int, 0)
	memory := make(map[uint16]uint16)
	stringsMemory := make(map[uint16]string)
	var workingPartyContext *PartyContext
	if partyContext != nil {
		workingPartyContext = partyContext
	}
	partyItems := make(map[uint16]bool)
	rebuildPartyItems := func() {
		clear(partyItems)
		if workingPartyContext == nil {
			return
		}
		for _, member := range workingPartyContext.Members {
			for _, itemType := range member.ItemTypes {
				partyItems[uint16(itemType)] = true
			}
		}
	}
	rebuildPartyItems()
	rng := rand.New(rand.NewSource(seed))
	if runtime != nil {
		if runtime.Random == nil {
			runtime.Random = randomstream.New(seed)
		}
		// A BlockSession owns one continuous PRNG stream across separate ECL
		// entry invocations. Recreating it for every terrain step makes a
		// fixed replay seed return the same RANDOM result forever.
		rng = runtime.Random.Rand()
	}
	var compare [6]bool
	// eventSequence 是這次執行的執行序計數，`SaveWrites` 與 `CallRequests`
	// 共用它。⚠ 不要拿 `PC` 當順序（見 MemoryWrite.Sequence 的說明）。
	eventSequence := 0
	selectedPlayerIndex := -1
	selectedPlayerSet := false
	selectedTeamListIndex := 0
	if runtime != nil && runtime.Started {
		pc = runtime.PC
		stack = append(stack, runtime.Stack...)
		for address, value := range runtime.Memory {
			memory[address] = value
		}
		for address, value := range runtime.Strings {
			stringsMemory[address] = value
		}
		compare = runtime.Compare
		selectedPlayerIndex = runtime.SelectedPlayerIndex
		selectedPlayerSet = runtime.SelectedPlayerSet
	}
	// Standalone runners do not have a BlockSession to install the reference
	// 0x8000 code-memory window. Populate missing bytes here; a session-owned
	// runtime already contains the same values and may preserve in-run writes.
	for index, value := range payload {
		address := CodeAddressBase + index
		if address > 0x9DFF {
			break
		}
		if _, exists := memory[uint16(address)]; !exists {
			memory[uint16(address)] = uint16(value)
		}
	}
	saveState := func(nextPC int) {
		if runtime == nil {
			return
		}
		runtime.PC = nextPC
		runtime.Started = true
		runtime.Stack = append(runtime.Stack[:0], stack...)
		runtime.Memory = memory
		runtime.Strings = stringsMemory
		runtime.Compare = compare
		runtime.SelectedPlayerIndex = selectedPlayerIndex
		runtime.SelectedPlayerSet = selectedPlayerSet
	}
	selectionCursor := 0
	whoSelectionCursor := 0
	stringInputCursor := 0
	result := RunResult{PC: pc}
	// recordStore 把「指令算出一個值、存進某個變數」記進 `SaveWrites`。
	//
	// ★ 原作只有一條寫入路徑：`STOREVALUE`。所有會寫變數的 opcode 都走它，
	// 而它同時負責把隊伍座標鏡射到 `720Fh`／`7210h`／`7211h` 並立髒旗標。
	// 所以「哪些 opcode 寫過座標」在原作裡不是分類問題——**任何一個都算**。
	// remake 這一側靠 `SaveWrites` 重建同一件事，因此每一條算完再存的指令都要
	// 記；只記 `09h SAVE` 會漏掉 `ADD`／`SUBTRACT`／`GETTABLE` 那 21 處座標寫入
	// （spec 1159）。
	recordStore := func(address, value uint16, at int) {
		// ★ 原作的 `STOREVALUE` 在這一刻就把座標鏡射到 `720Fh`／`7210h`／`7211h`
		// 並立髒旗標（spec 1150）。這裡是 remake 唯一那條寫入路徑，所以鏡射
		// 也只接在這裡一處。
		runtime.View.Store(address, value, runtime.CurrentBlock)
		eventSequence++
		result.SaveWrites = append(result.SaveWrites, MemoryWrite{
			Address:  address,
			Value:    value,
			PC:       at,
			Sequence: eventSequence,
		})
	}
	for result.Steps < maxSteps {
		instruction, err := decodeInstruction(payload, pc)
		if err != nil {
			result.PC = pc
			return result, err
		}
		recordRunCoverage(runtime, pc)
		result.Steps++
		next := instruction.Next
		switch instruction.Command.Opcode {
		case 0x00: // EXIT
			result.PC = next
			result.Exited = true
			saveState(next)
			return result, nil
		case 0x01, 0x02: // GOTO / GOSUB
			target, ok := CodeTarget(instruction.Operands[0], len(payload))
			if !ok {
				return result, fmt.Errorf("opcode 0x%02X at %d has invalid code target", instruction.Command.Opcode, pc)
			}
			if instruction.Command.Opcode == 0x02 {
				stack = append(stack, next)
			}
			pc = target
			continue
		case 0x03: // COMPARE
			if operandIsText(instruction.Operands[0]) || operandIsText(instruction.Operands[1]) {
				left, err := operandText(instruction.Operands[0], stringsMemory)
				if err != nil {
					return result, fmt.Errorf("string compare at %d: %w", pc, err)
				}
				right, err := operandText(instruction.Operands[1], stringsMemory)
				if err != nil {
					return result, fmt.Errorf("string compare at %d: %w", pc, err)
				}
				compare[0] = left == right
				compare[1] = left != right
				compare[2] = left < right
				compare[3] = left > right
				compare[4] = left <= right
				compare[5] = left >= right
				break
			}
			left, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("compare at %d: %w", pc, err)
			}
			right, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("compare at %d: %w", pc, err)
			}
			compare[0] = left == right
			compare[1] = left != right
			compare[2] = left < right
			compare[3] = left > right
			compare[4] = left <= right
			compare[5] = left >= right
		case 0x04, 0x05, 0x06, 0x07, 0x2F, 0x30: // arithmetic / AND / OR
			if !instruction.Operands[2].WordSet {
				return result, fmt.Errorf("arithmetic at %d has non-address destination", pc)
			}
			left, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("arithmetic at %d: %w", pc, err)
			}
			right, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("arithmetic at %d: %w", pc, err)
			}
			var value uint16
			switch instruction.Command.Opcode {
			case 0x04:
				value = left + right
			case 0x05:
				value = right - left
			case 0x06:
				if right == 0 {
					return result, fmt.Errorf("arithmetic at %d divides by zero", pc)
				}
				value = left / right
			case 0x07:
				value = left * right
			case 0x2F:
				value = left & right
			case 0x30:
				value = left | right
			}
			memory[instruction.Operands[2].Word] = value
			recordStore(instruction.Operands[2].Word, value, pc)
			if instruction.Command.Opcode == 0x2F || instruction.Command.Opcode == 0x30 {
				// 原作 `2Fh`／`30h` 共用的 handler 在寫回目的地之前呼叫
				// `compare_variables(0, 結果)`——**0 是左運算元**
				// （DOS `overlay-02:0DF3h` 先 `push 0` 再 push 結果，
				// 而 `03h COMPARE` 是 `push op1` 再 `push op2`）。
				// 所以下面四個排序格子是 `0 op 結果`，不是 `結果 op 0`。
				// ⚠ 別照直覺把它們對調：全 corpus 174 處 `AND`／`OR` 後面
				// 沒有一處接排序型 `IF`（125 處是 `IF <>`），對調了測不出來
				// ——只有讀 `0DF3h` 那三條指令才分得出方向（spec 1157）。
				compare[0] = value == 0
				compare[1] = value != 0
				compare[2] = 0 < value
				compare[3] = 0 > value
				compare[4] = 0 <= value
				compare[5] = 0 >= value
			}
		case 0x08: // RANDOM
			if !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("random at %d has non-address destination", pc)
			}
			maximum, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("random at %d: %w", pc, err)
			}
			// The original increments every maximum below 0xFF, making the
			// command an inclusive range [0, maximum].
			bound := int(maximum)
			if bound < 0xFF {
				bound++
			}
			if bound == 0 {
				bound = 1
			}
			value := uint16(rng.Intn(bound))
			memory[instruction.Operands[1].Word] = value
			recordStore(instruction.Operands[1].Word, value, pc)
			result.RandomValues = append(result.RandomValues, value)
		case 0x14: // COMPARE AND
			leftA, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			rightA, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			leftB, err := operandValue(instruction.Operands[2], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			rightB, err := operandValue(instruction.Operands[3], memory)
			if err != nil {
				return result, fmt.Errorf("COMPARE AND at %d: %w", pc, err)
			}
			for i := range compare {
				compare[i] = false
			}
			if leftA == rightA && leftB == rightB {
				compare[0] = true
			} else {
				compare[1] = true
			}
		case 0x2A: // GETTABLE
			if !instruction.Operands[0].WordSet || !instruction.Operands[2].WordSet {
				return result, fmt.Errorf("GETTABLE at %d has non-address operand", pc)
			}
			index, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("GETTABLE at %d: %w", pc, err)
			}
			value := memory[instruction.Operands[0].Word+index]
			memory[instruction.Operands[2].Word] = value
			recordStore(instruction.Operands[2].Word, value, pc)
		case 0x29: // ENCOUNTER MENU
			if len(instruction.Operands) != 14 {
				return result, fmt.Errorf("encounter menu at %d has %d operands", pc, len(instruction.Operands))
			}
			if !instruction.Operands[3].WordSet {
				return result, fmt.Errorf("encounter menu at %d has no memory destination", pc)
			}
			maxDistance, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("encounter menu distance at %d: %w", pc, err)
			}
			// ⚠ `29h` 進門**重新**設一次距離（`20D1h` 寫上限、`2177h`..`21B5h` 從
			// 地圖算再夾住），把 `0Dh APPROACH` 減掉的蓋回去。照抄這個次序，
			// 否則「走近兩步之後開選單」會挑到錯的旁白。
			InitEncounterDistance(memory, maxDistance)
			distance := memory[EncounterDistanceCell]
			// 三句旁白依距離挑一句，挑法見 `EncounterPromptSlots`。
			//
			// ⚠ 原作看的是**當下的距離**（由地圖座標算出、再被這個上限夾住），
			// remake 還沒有那個距離模型，所以用上限代替——`ADVANCE`／`PARLAY`
			// 的判斷本來就已經是同一個代替法。
			prompt := ""
			for _, slot := range EncounterPromptSlots(int(distance)) {
				operand := instruction.Operands[EncounterPromptOperand+slot]
				if !operandIsText(operand) {
					continue
				}
				prompt, err = operandText(operand, stringsMemory)
				if err != nil {
					return result, fmt.Errorf("encounter menu prompt at %d: %w", pc, err)
				}
				if prompt != "" {
					break
				}
			}
			options := []string{"COMBAT", "WAIT", "FLEE", "ADVANCE"}
			if distance == 0 {
				options[3] = "PARLAY"
			}
			menu := Menu{Location: instruction.Operands[3].Word, Options: options, Prompt: prompt}
			if pauseOnMissing && selectionCursor >= len(selections) {
				result.Menus = append(result.Menus, menu)
				result.WaitingForMenu = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if selectionCursor < len(selections) && selections[selectionCursor] < uint16(len(options)) {
				menu.Selected = selections[selectionCursor]
			}
			selectionCursor++
			result.SelectionsConsumed = selectionCursor
			mappingIndex := int(menu.Selected)
			mapping, err := operandValue(instruction.Operands[4+mappingIndex], memory)
			if err != nil {
				return result, fmt.Errorf("encounter menu action at %d: %w", pc, err)
			}
			mapping = resolveEncounterAction(menu.Selected, mapping, maxDistance)
			memory[instruction.Operands[3].Word] = mapping
			result.EncounterActions = append(result.EncounterActions, mapping)
			result.Menus = append(result.Menus, menu)
		case 0x2C: // PARLAY
			if len(instruction.Operands) != 6 || !instruction.Operands[5].WordSet {
				return result, fmt.Errorf("parlay at %d has invalid operands", pc)
			}
			menu := Menu{
				Location: instruction.Operands[5].Word,
				Options:  []string{"PARLAY_HAUGHTY", "PARLAY_SLY", "PARLAY_MEEK", "PARLAY_NICE", "PARLAY_ABUSIVE"},
			}
			if pauseOnMissing && selectionCursor >= len(selections) {
				result.Menus = append(result.Menus, menu)
				result.WaitingForMenu = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if selectionCursor < len(selections) && selections[selectionCursor] < uint16(len(menu.Options)) {
				menu.Selected = selections[selectionCursor]
			}
			selectionCursor++
			result.SelectionsConsumed = selectionCursor
			mapping, err := operandValue(instruction.Operands[menu.Selected], memory)
			if err != nil {
				return result, fmt.Errorf("parlay action at %d: %w", pc, err)
			}
			memory[instruction.Operands[5].Word] = mapping
			result.Menus = append(result.Menus, menu)
		case 0x2B: // HORIZONTAL MENU
			header, headNext, err := ParseOperands(payload, pc, 2)
			if err != nil {
				return result, fmt.Errorf("HORIZONTAL MENU header at %d: %w", pc, err)
			}
			if !header[0].WordSet {
				return result, fmt.Errorf("HORIZONTAL MENU at %d has non-address destination", pc)
			}
			count, err := operandValue(header[1], memory)
			if err != nil {
				return result, fmt.Errorf("HORIZONTAL MENU count at %d: %w", pc, err)
			}
			if count == 0 || count > 64 {
				return result, fmt.Errorf("HORIZONTAL MENU at %d has invalid option count %d", pc, count)
			}
			stringOperands, stringsEnd, err := ParseOperands(payload, headNext-1, int(count))
			if err != nil {
				return result, fmt.Errorf("HORIZONTAL MENU strings at %d: %w", pc, err)
			}
			menu := Menu{Location: header[0].Word, Options: make([]string, 0, count)}
			for _, operand := range stringOperands {
				message, err := operandText(operand, stringsMemory)
				if err != nil {
					return result, fmt.Errorf("HORIZONTAL MENU option at %d: %w", pc, err)
				}
				menu.Options = append(menu.Options, message)
			}
			if pauseOnMissing && selectionCursor >= len(selections) {
				result.Menus = append(result.Menus, menu)
				result.WaitingForMenu = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if selectionCursor < len(selections) && selections[selectionCursor] < count {
				menu.Selected = selections[selectionCursor]
			}
			selectionCursor++
			result.SelectionsConsumed = selectionCursor
			memory[menu.Location] = menu.Selected
			result.Menus = append(result.Menus, menu)
			next = stringsEnd
		case 0x15: // VERTICAL MENU
			header, headNext, err := ParseOperands(payload, pc, 3)
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU header at %d: %w", pc, err)
			}
			if !header[0].WordSet {
				return result, fmt.Errorf("VERTICAL MENU at %d has non-address destination", pc)
			}
			count, err := operandValue(header[2], memory)
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU count at %d: %w", pc, err)
			}
			if count == 0 || count > 64 {
				return result, fmt.Errorf("VERTICAL MENU at %d has invalid option count %d", pc, count)
			}
			prompt, err := operandText(header[1], stringsMemory)
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU prompt at %d: %w", pc, err)
			}
			stringOperands, stringsEnd, err := ParseOperands(payload, headNext-1, int(count))
			if err != nil {
				return result, fmt.Errorf("VERTICAL MENU strings at %d: %w", pc, err)
			}
			menu := Menu{Location: header[0].Word, Options: make([]string, 0, count), Vertical: true, Prompt: prompt}
			for _, operand := range stringOperands {
				message, err := operandText(operand, stringsMemory)
				if err != nil {
					return result, fmt.Errorf("VERTICAL MENU option at %d: %w", pc, err)
				}
				menu.Options = append(menu.Options, message)
			}
			if pauseOnMissing && selectionCursor >= len(selections) {
				result.Menus = append(result.Menus, menu)
				result.WaitingForMenu = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if selectionCursor < len(selections) && selections[selectionCursor] < count {
				menu.Selected = selections[selectionCursor]
			}
			selectionCursor++
			result.SelectionsConsumed = selectionCursor
			memory[menu.Location] = menu.Selected
			result.Menus = append(result.Menus, menu)
			next = stringsEnd
		case 0x09: // SAVE
			if !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("save at %d has non-address destination", pc)
			}
			if operandIsText(instruction.Operands[0]) {
				value, err := operandText(instruction.Operands[0], stringsMemory)
				if err != nil {
					return result, fmt.Errorf("save at %d: %w", pc, err)
				}
				stringsMemory[instruction.Operands[1].Word] = value
			} else {
				value, err := operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("save at %d: %w", pc, err)
				}
				memory[instruction.Operands[1].Word] = value
				recordStore(instruction.Operands[1].Word, value, pc)
				if instruction.Operands[1].Word == 0x7D0C {
					result.CombatTeamWrites = append(result.CombatTeamWrites, CombatTeamWrite{
						TeamListIndex: selectedTeamListIndex,
						Value:         value,
					})
				}
				// The player-memory window is relative to SelectedPlayer in
				// the reference VM. Scripts use LOAD CHARACTER followed by a
				// write to +0x10C to move loaded NPC copies between combat
				// teams. Preserve that side effect on the spawn descriptor.
				if instruction.Operands[1].Word == 0x7D0C {
					ApplyCombatTeamWrites(result.MonsterSpawns,
						[]CombatTeamWrite{{TeamListIndex: selectedTeamListIndex, Value: value}})
				}
			}
		case 0x10: // INPUT STRING
			if len(instruction.Operands) != 2 || !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("INPUT STRING at %d has invalid operands", pc)
			}
			maxLength, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("INPUT STRING length at %d: %w", pc, err)
			}
			request := StringInputRequest{
				MaxLength:   maxLength,
				Destination: instruction.Operands[1].Word,
			}
			if pauseOnMissing && stringInputCursor >= len(stringInputs) {
				result.StringInputRequests = append(result.StringInputRequests, request)
				result.WaitingForString = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			value := ""
			if stringInputCursor < len(stringInputs) {
				value = normalizeInputString(stringInputs[stringInputCursor], maxLength)
			}
			stringInputCursor++
			result.StringInputsConsumed = stringInputCursor
			request.Value = value
			request.InputProvided = true
			stringsMemory[request.Destination] = value
			result.StringInputRequests = append(result.StringInputRequests, request)
		case 0x11, 0x12: // PRINT / PRINTCLEAR
			if len(instruction.Operands) != 1 {
				return result, fmt.Errorf("print at %d has unexpected arity", pc)
			}
			operand := instruction.Operands[0]
			if operand.Code == 0x80 {
				result.Text = append(result.Text, DecodePackedText(operand.Packed))
			} else if operand.Code == 0x81 {
				message, err := operandText(operand, stringsMemory)
				if err != nil {
					return result, fmt.Errorf("print at %d: %w", pc, err)
				}
				result.Text = append(result.Text, message)
			} else {
				value, err := operandValue(operand, memory)
				if err != nil {
					return result, fmt.Errorf("print at %d: %w", pc, err)
				}
				result.Text = append(result.Text, fmt.Sprint(value))
			}
		case 0x13: // RETURN
			if len(stack) == 0 {
				result.PC = next
				return result, nil
			}
			pc = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			continue
		case 0x20: // NEWECL
			blockID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("NEWECL at %d: %w", pc, err)
			}
			id := uint8(blockID)
			result.NewECLBlockID = &id
			result.PC = next
			saveState(next)
			return result, nil
		case 0x24: // COMBAT
			// `24h` 在原作是**四選一**的服務分派點，不只是「打一場」
			// （spec 1095／1149／1182）。順序照原作 `overlay-02:179Ah`：
			//
			//	8B69h 或 8B56h 非 0                  → 打（overlay-08 GOCOMBAT）
			//	bank1^[6D8h] ＝ 7F6Ch ＝ 1           → 商店（overlay-06 SHOP）
			//	bank1^[5C4h] ＝ 7EE2h ＝ 1           → 神殿（overlay-04 TEMPLE）
			//	以上都不成立                          → 戰後處理（overlay-05 POSTCOM）
			//
			// ★ **場上有怪就直接打**，根本不看商店旗標——
			//
			//   17A4  cmp byte ptr ds:8B69h, 0 / jz  → 非 0 就 jmp sub_1956（打）
			//   17AE  cmp byte ptr ds:8B56h, 0 / jz  → 非 0 就 jmp sub_1956（打）
			//   17B8  cmp word ptr es:[di+6D8h], 1   → 這之後才輪到商店
			//
			// `8B69h` 就是 `1Ch CLEARMONSTERS` 清的那個「有怪要打」旗標
			// （spec 1145），在 remake 對應的是怪物鏈非空。旗標讀到就立刻清零，
			// exactly-once 由清零保證。
			//
			// ⚠ **`7EE2h` 是神殿不是營地。** spec 1030 把 `overlay-04` 標成營地
			// 主選單，spec 1095／1149 沿用，還因此說 remake 的 `TempleRequested`
			// 名字不對——**反了**。PC-98 的 Borland 符號寫著 `LOADTEMPLE`／
			// `GOTEMPLE`，而 DOS `overlay-04` 的字串是 `how can we help you?`、
			// `Heal View Pool Appraise Exit`、`Raise Dead` 與「a priest says…」。
			// 營地是 `overlay-15`（`LOADCAMP`／`DOCAMP`）。詳見 spec 1182。
			//
			// ★ 商店與神殿旗標的 producer 是**腳本自己**（spec 1182）：全 corpus
			// `7EE2h` 4 處、`7F6Ch` 9 處，慣用法一律是
			// `CLEARMONSTERS` ＋ `SAVE 01 <旗標>` ＋ `COMBAT`。
			// ⚠ spec 1095 曾說這兩格「1,355 條指令裡沒有一條寫過」——那是假零。
			monstersPresent := len(result.MonsterSpawns) > 0 ||
				(runtime != nil && len(runtime.MonsterSpawns) > 0)
			if !monstersPresent && memory[addrShopRequest] == 1 {
				result.ShopRequested = true
				result.ShopPriceScale = memory[addrShopPriceScale]
				memory[addrShopRequest] = 0
			} else if !monstersPresent && memory[addrTempleRequest] == 1 {
				result.TempleRequested = true
				memory[addrTempleRequest] = 0
			} else if !monstersPresent {
				// ★ 第四支：兩個旗標都不成立而且場上沒怪 ⇒ 原作跑的是
				// **戰後處理**（`overlay-05` ＝ POSTCOM 的 `DOPOSTCOMBAT`），
				// 不是「打一場」（spec 1182）。腳本的慣用法是先用 `27h TREASURE`
				// 把戰利品堆好，再用 `24h` 開分配畫面——`overlay-05` 的字串
				// （`The party has found Treasure!`／`Each character receives `／
				// `View Take Pool Share`）就是那一組。
				//
				// ⚠ 原作的 `24h` **沒有「零隻怪的戰鬥」這種東西**：有怪才走
				// `sub_1956`。所以零隻怪不能發 `CombatRequested`。
				result.PostCombatRequested = true
			} else {
				result.CombatRequested = true
				if runtime != nil {
					if len(result.MonsterSpawns) == 0 {
						result.MonsterSpawns = append([]MonsterSpawn(nil), runtime.MonsterSpawns...)
					}
					if result.MonsterSetup == nil && runtime.MonsterSetup != nil {
						setup := *runtime.MonsterSetup
						result.MonsterSetup = &setup
					}
					runtime.MonsterSpawns = nil
					runtime.MonsterSetup = nil
				}
			}
			result.PC = next
			saveState(next)
			return result, nil
		case 0x2D: // CALL
			address, err := operandAddress(instruction.Operands[0])
			if err != nil {
				return result, fmt.Errorf("CALL at %d: %w", pc, err)
			}
			// CALL dispatches an engine routine and returns to the next ECL
			// instruction. Keep the address observable while leaving the
			// routine-specific DOS side effect to a later adapter.
			result.CallAddresses = append(result.CallAddresses, address)
			eventSequence++
			request := CallRequest{
				Address:  address,
				PC:       pc,
				Sequence: eventSequence,
				View:     runtime.View,
			}
			result.CallRequests = append(result.CallRequests, request)
			if address == ExternalCallAdvancePictureFrame {
				result.PictureFrameAdvances++
			}
			if address == ExternalCallMoveForward {
				// 原作走完一步就把新座標留在地圖暫存器裡，同一次執行裡排在
				// 後面的重畫看得到（spec 1172）。
				runtime.View.StepForward()
			}
			if address == ExternalCallRedrawView {
				// 原作重畫之後把那五個旗標逐個清掉，所以同一批寫入不會被投影
				// 兩次（spec 1150）。
				runtime.View.ClearDirty()
			}
		case 0x3A: // DELAY
			// GameDelay is an engine timing boundary with no ECL memory side
			// effect. Preserve the count for the frontend and continue.
			result.DelayCount++
		case 0x0D: // APPROACH
			// 原作 `overlay-02:0801h`（22 條）整支就是：距離 > 0 就減一，再用
			// 新的距離重畫遭遇圖；距離是 0 就什麼都不做。
			//
			// ⚠ 這個減一**不會影響後面的遭遇選單**：`29h` 進門會重新從地圖算一次
			// 距離並重新夾住（`2177h`..`21B5h`），把 `APPROACH` 減掉的蓋回去。
			// 所以它是「怪物一步一步走近」的演出，不是選單條件。
			if ApproachEncounter(memory) {
				result.ApproachCount++
			}
		case 0x2E: // DAMAGE
			if len(instruction.Operands) != 5 {
				return result, fmt.Errorf("DAMAGE at %d has unexpected arity", pc)
			}
			values := make([]uint16, len(instruction.Operands))
			for index, operand := range instruction.Operands {
				value, err := operandValue(operand, memory)
				if err != nil {
					return result, fmt.Errorf("DAMAGE operand %d at %d: %w", index+1, pc, err)
				}
				values[index] = value
			}
			// The public CoAB reference confirms the order as flags, dice count,
			// dice size, damage bonus, and saving-throw flags. Keep raw values;
			// signed bonus conversion, target selection, saves, and HP mutation
			// belong to the game adapter.
			result.DamageRequests = append(result.DamageRequests, DamageRequest{
				Flags: values[0], DiceCount: values[1], DiceSize: values[2],
				Bonus: values[3], SaveFlags: values[4],
				SelectedPlayerIndex: selectedPlayerIndex,
				SelectedPlayerSet:   selectedPlayerSet,
			})
		case 0x34: // ECL CLOCK
			if len(instruction.Operands) != 2 {
				return result, fmt.Errorf("ECL CLOCK at %d has unexpected arity", pc)
			}
			timeStep, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("ECL CLOCK time step at %d: %w", pc, err)
			}
			timeSlot, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("ECL CLOCK time slot at %d: %w", pc, err)
			}
			result.ClockRequests = append(result.ClockRequests, ClockRequest{TimeStep: timeStep, TimeSlot: timeSlot})
		case 0x35: // SAVE TABLE
			if !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("SAVE TABLE at %d has non-address destination", pc)
			}
			value, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("SAVE TABLE value at %d: %w", pc, err)
			}
			offset, err := operandValue(instruction.Operands[2], memory)
			if err != nil {
				return result, fmt.Errorf("SAVE TABLE offset at %d: %w", pc, err)
			}
			// Reference CMD_SaveTable writes value operand 1 to the address
			// operand 2.Word + value operand 3.
			address := instruction.Operands[1].Word + offset
			memory[address] = value
			recordStore(address, value, pc)
		case 0x0A: // LOAD CHARACTER
			address, err := operandAddress(instruction.Operands[0])
			if err != nil {
				return result, fmt.Errorf("LOAD CHARACTER at %d: %w", pc, err)
			}
			value, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("LOAD CHARACTER value at %d: %w", pc, err)
			}
			// Keep both raw address and decoded player selector observable; the
			// State adapter owns roster resolution and renderer side effects.
			result.LoadCharacterAddresses = append(result.LoadCharacterAddresses, address)
			result.LoadCharacterRequests = append(result.LoadCharacterRequests, LoadCharacterRequest{
				Address: address, Value: value, PlayerIndex: uint8(value & 0x7F), HighBitSet: value&0x80 != 0,
			})
			selectedTeamListIndex = int(value & 0x7F)
			// LOAD CHARACTER selects an absolute TeamList slot. The player
			// memory window at +0x100 does not contain a stored byte: the
			// reference getter projects in_combat as 1 for an existing TeamList
			// slot and 0 for an absent selector. ECL room scanners compare this
			// value with zero to terminate their zero-based LOAD CHARACTER loop;
			// 0x80 is reserved for the selected player's quick-fight team byte at
			// +0x10C.
			// Guild block 2 relies on this probe before assigning four loaded
			// thieves to our quick-fight team.
			teamCount := len(result.MonsterSpawns)
			for _, spawn := range result.MonsterSpawns {
				count := int(spawn.Count)
				if count == 0 {
					count = 1
				}
				teamCount += count - 1
			}
			if workingPartyContext != nil {
				teamCount += len(workingPartyContext.Members)
			}
			if selectedTeamListIndex >= 0 && selectedTeamListIndex < teamCount {
				memory[0x7D00] = 1
			} else {
				memory[0x7D00] = 0
			}
			if workingPartyContext != nil {
				playerIndex := int(value & 0x7F)
				if playerIndex >= 0 && playerIndex < len(workingPartyContext.Members) {
					// Reference vm_CopyStringFromMemory treats 0x7C00 as the
					// selected player's name string. Preserve it in RuntimeState
					// so later COMPARE/PRINT operands see the same selection.
					stringsMemory[0x7C00] = workingPartyContext.Members[playerIndex].Name
					// Player +0xB8 is the control/morale byte. NPC records use
					// values >=0x80; ECL5 block 0x30 relies on this validity
					// probe before comparing Akabar's selected name.
					memory[addrPlayerControlMorale] = uint16(workingPartyContext.Members[playerIndex].ControlMorale)
					// 角色 +192h 的投影。原作讀取端只取最低位元（and 1，
					// spec 1040/1098），ECL1 block 0x50 用 COMPARE 7CE4h, 0
					// 判斷是否跳轉，所以這裡必須跟著模擬同樣的遮罩。
					memory[addrPlayerFlag192] = uint16(workingPartyContext.Members[playerIndex].ECLFlag192 & 1)
					selectedPlayerIndex = playerIndex
					selectedPlayerSet = true
				} else {
					memory[addrPlayerControlMorale] = 0
					memory[addrPlayerFlag192] = 0
				}
			}
		case 0x32: // FIND ITEM
			itemID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("FIND ITEM at %d: %w", pc, err)
			}
			result.FindItemIDs = append(result.FindItemIDs, itemID)
			request := FindItemRequest{ItemID: itemID}
			if workingPartyContext != nil {
				request.Resolved = true
				request.Found = partyItems[itemID]
				for index := range compare {
					compare[index] = false
				}
				compare[0] = request.Found
				compare[1] = !request.Found
			}
			result.FindItemRequests = append(result.FindItemRequests, request)
		case 0x40: // DESTROY ITEMS
			itemID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("DESTROY ITEMS at %d: %w", pc, err)
			}
			// Keep inventory mutation explicit for the party adapter; the VM
			// itself must not silently delete an item without roster context.
			result.DestroyItemIDs = append(result.DestroyItemIDs, itemID)
			if workingPartyContext != nil {
				delete(partyItems, itemID)
			}
		case 0x1D: // PARTYSTRENGTH
			if !instruction.Operands[0].WordSet {
				return result, fmt.Errorf("PARTYSTRENGTH at %d has non-address destination", pc)
			}
			// Reference CMD_PartyStrength computes a byte from the live party
			// (HP, AC, hit bonus, cleric level and magic-user level). Preserve
			// the destination here; the game adapter supplies those roster stats.
			request := PartyStrengthRequest{Destination: instruction.Operands[0].Word}
			if workingPartyContext != nil {
				request.Value = workingPartyContext.partyStrength()
				request.Resolved = true
				memory[request.Destination] = request.Value
			}
			result.PartyStrengthRequests = append(result.PartyStrengthRequests, request)
		case 0x22: // PARTY SURPRISE
			if !instruction.Operands[0].WordSet || !instruction.Operands[1].WordSet {
				return result, fmt.Errorf("PARTY SURPRISE at %d has non-address destination", pc)
			}
			request := PartySurpriseRequest{
				RangerDestination: instruction.Operands[0].Word,
				OtherDestination:  instruction.Operands[1].Word,
			}
			if workingPartyContext != nil {
				request.RangerValue = 0
				if workingPartyContext.hasRanger() {
					request.RangerValue = 1
				}
				// OtherValue 保持 0：原作 `overlay-02:1636h` 的第二個目的地寫的是
				// 一個從頭到尾都是 0 的區域變數（spec 1113）。遊俠那一支把 1 寫進
				// **另一個**區域變數，而那個變數沒有讀者。
				// ⚠ 不要「補上」第二個突襲值——原作沒有算過它。
				request.Resolved = true
				memory[request.RangerDestination] = request.RangerValue
				memory[request.OtherDestination] = request.OtherValue
			}
			result.PartySurpriseRequests = append(result.PartySurpriseRequests, request)
		case 0x1E: // CHECKPARTY
			query := uint16(0)
			var err error
			if instruction.Operands[0].Code == 1 {
				query = instruction.Operands[0].Word
			} else {
				query, err = operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("CHECKPARTY query at %d: %w", pc, err)
				}
			}
			affectID, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("CHECKPARTY affect at %d: %w", pc, err)
			}
			request := CheckPartyRequest{Query: query, AffectID: affectID}
			for index := range request.Destinations {
				operand := instruction.Operands[index+2]
				if !operand.WordSet {
					return result, fmt.Errorf("CHECKPARTY destination %d at %d is not an address", index+1, pc)
				}
				request.Destinations[index] = operand.Word
			}
			if workingPartyContext != nil {
				resolved, known := workingPartyContext.checkParty(query, affectID)
				resolved.Destinations = request.Destinations
				request = resolved
				if request.Resolved {
					if request.AffectFound {
						memory[request.Destinations[0]] = 0
						memory[request.Destinations[1]] = 0
						memory[request.Destinations[2]] = 0
						memory[request.Destinations[3]] = 1
					} else if known {
						memory[request.Destinations[0]] = request.Minimum
						memory[request.Destinations[1]] = request.Maximum
						memory[request.Destinations[2]] = request.Average
						memory[request.Destinations[3]] = 0
					}
				}
			}
			result.CheckPartyRequests = append(result.CheckPartyRequests, request)
		case 0x39: // WHO
			prompt := ""
			if len(result.Text) > 0 {
				prompt = result.Text[len(result.Text)-1]
			}
			request := WhoRequest{Prompt: prompt}
			if whoSelectionCursor >= len(whoSelections) && pauseOnMissing {
				result.WhoRequests = append(result.WhoRequests, request)
				result.WaitingForWho = true
				result.PC = pc
				saveState(pc)
				return result, nil
			}
			if whoSelectionCursor < len(whoSelections) {
				request.Selected = whoSelections[whoSelectionCursor]
				request.SelectionProvided = true
				whoSelectionCursor++
				result.WhoSelectionsConsumed++
				if workingPartyContext != nil && int(request.Selected) < len(workingPartyContext.Members) {
					selectedPlayerIndex = int(request.Selected)
					selectedPlayerSet = true
					stringsMemory[0x7C00] = workingPartyContext.Members[selectedPlayerIndex].Name
					// ★ **選完人要把「有選到人」投影出去。** `7D00h` 不是普通格子，
					// 是**目前選中角色**的欄位投影（spec 624：
					// `if char^[197h] <> 0 then 1 else 80h`，另有一個一次性旗標
					// 讓它讀成 0）。原作腳本靠它判斷「玩家真的挑了一個人」。
					//
					// ⚠ 少了這一行會**卡死在腳本自己的迴圈裡**：`ECL5/0x33:0C52`
					// 是 `WHO` → `COMPARE 7D00h,1` → 不等就印一句話再 `GOTO` 回
					// `WHO`。remake 原本讓 `7D00h` 留著上一次 `LOAD CHARACTER`
					// 的值，於是巫師塔那一場永遠問不完（實測按鍵重放在那裡按了
					// 五百多次都出不來，而畫面上只是同一個「請選擇角色」）。
					memory[0x7D00] = 1
				}
			}
			result.WhoRequests = append(result.WhoRequests, request)
		case 0x3F: // FIND SPECIAL
			affectID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("FIND SPECIAL at %d: %w", pc, err)
			}
			request := FindSpecialRequest{AffectID: affectID, SelectedPlayerIndex: selectedPlayerIndex}
			if workingPartyContext != nil && selectedPlayerSet && selectedPlayerIndex >= 0 && selectedPlayerIndex < len(workingPartyContext.Members) {
				request.Resolved = true
				for _, activeAffect := range workingPartyContext.Members[selectedPlayerIndex].Effects {
					if uint16(activeAffect) == affectID {
						request.Found = true
						break
					}
				}
				for index := range compare {
					compare[index] = false
				}
				compare[0] = request.Found
				compare[1] = !request.Found
			}
			result.FindSpecialRequests = append(result.FindSpecialRequests, request)
		case 0x3E: // DUMP
			request := DumpRequest{SelectedPlayerIndex: selectedPlayerIndex, NextSelectedPlayerIndex: -1}
			if workingPartyContext != nil && selectedPlayerSet && selectedPlayerIndex >= 0 && selectedPlayerIndex < len(workingPartyContext.Members) {
				request.Resolved = true
				workingPartyContext.Members = append(workingPartyContext.Members[:selectedPlayerIndex], workingPartyContext.Members[selectedPlayerIndex+1:]...)
				rebuildPartyItems()
				if len(workingPartyContext.Members) > 0 {
					if selectedPlayerIndex > 0 {
						selectedPlayerIndex--
					} else {
						selectedPlayerIndex = 0
					}
					selectedPlayerSet = true
					request.NextSelectedPlayerIndex = selectedPlayerIndex
					request.NextSelectedPlayerSet = true
					stringsMemory[0x7C00] = workingPartyContext.Members[selectedPlayerIndex].Name
				} else {
					selectedPlayerIndex = -1
					selectedPlayerSet = false
					delete(stringsMemory, 0x7C00)
				}
			}
			result.DumpRequests = append(result.DumpRequests, request)
		case 0x33: // PRINT RETURN
			// This command changes the original text window/cursor state. Keep
			// its instruction boundary observable while leaving renderer layout
			// to the game UI adapter.
			result.PrintReturnCount++
		case 0x38: // PROGRAM
			// PROGRAM dispatches into an external engine routine. The reference
			// implementation ends the current VM pass for PROGRAM 0/3/8/9;
			// retain the ID and stop at that boundary until the corresponding
			// renderer/game-state routine is implemented.
			program, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("PROGRAM at %d: %w", pc, err)
			}
			programID := uint8(program)
			result.ProgramIDs = append(result.ProgramIDs, programID)
			if programID == 0 || programID == 3 || programID == 8 || programID == 9 {
				result.ProgramExit = true
				result.PC = next
				// PROGRAM is a resumable external-engine boundary just like
				// COMBAT/SHOP/TEMPLE. Preserve the instruction after PROGRAM
				// so closing CAMP does not execute PROGRAM 9 a second time.
				saveState(next)
				return result, nil
			}
		case 0x0B: // LOAD MONSTER
			spawn, err := DecodeMonsterSpawnFromMemory(instruction, memory)
			if err != nil {
				return result, fmt.Errorf("LOAD MONSTER at %d: %w", pc, err)
			}
			result.MonsterSpawns = append(result.MonsterSpawns, spawn)
			if runtime != nil {
				runtime.MonsterSpawns = append(runtime.MonsterSpawns, spawn)
			}
		case 0x0C: // SETUP MONSTER
			setup, err := DecodeMonsterSetupFromMemory(instruction, memory)
			if err != nil {
				return result, fmt.Errorf("SETUP MONSTER at %d: %w", pc, err)
			}
			result.MonsterSetup = &setup
			if runtime != nil {
				owned := setup
				runtime.MonsterSetup = &owned
			}
			// `0Ch` 同時把遭遇距離擺好（`overlay-02:03EBh`..`043Ch`）：運算元 2
			// 是上限，當下距離由地圖算出再被上限夾住。
			InitEncounterDistance(memory, uint16(setup.MaxDistance))
		case 0x36: // ADD NPC
			npcID, err := operandValue(instruction.Operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("ADD NPC at %d: %w", pc, err)
			}
			morale, err := operandValue(instruction.Operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("ADD NPC morale at %d: %w", pc, err)
			}
			result.NPCIDs = append(result.NPCIDs, npcID)
			result.NPCRequests = append(result.NPCRequests, NPCRequest{ID: npcID, Morale: morale})
		case 0x25, 0x26: // ON GOTO / ON GOSUB
			operands, headNext, err := ParseOperands(payload, pc, 2)
			if err != nil {
				return result, fmt.Errorf("ON branch at %d: %w", pc, err)
			}
			index, err := operandValue(operands[0], memory)
			if err != nil {
				return result, fmt.Errorf("ON branch index at %d: %w", pc, err)
			}
			count, err := operandValue(operands[1], memory)
			if err != nil {
				return result, fmt.Errorf("ON branch count at %d: %w", pc, err)
			}
			if count > 256 {
				return result, fmt.Errorf("ON branch at %d has unreasonable target count %d", pc, count)
			}
			// The original decrements the cursor once before loading the
			// variable target list, so its first skipped byte is headNext-1.
			targets, afterTargets, err := ParseOperands(payload, headNext-1, int(count))
			if err != nil {
				return result, fmt.Errorf("ON branch targets at %d: %w", pc, err)
			}
			if index >= count {
				pc = afterTargets
				result.PC = pc
				continue
			}
			target, ok := CodeTarget(targets[index], len(payload))
			if !ok {
				return result, fmt.Errorf("ON branch at %d has invalid target %d", pc, index)
			}
			if instruction.Command.Opcode == 0x26 {
				stack = append(stack, afterTargets)
			}
			pc = target
			continue
		case 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B: // IF comparison
			index := int(instruction.Command.Opcode - 0x16)
			if !compare[index] {
				skipped, err := decodeInstruction(payload, next)
				if err != nil {
					return result, fmt.Errorf("if at %d cannot skip next command: %w", pc, err)
				}
				next = skipped.Next
			}
		case 0x28: // ROB
			values := [3]uint16{}
			for index, operand := range instruction.Operands {
				value, valueErr := operandValue(operand, memory)
				if valueErr != nil {
					return result, fmt.Errorf("rob at %d operand %d: %w", pc, index, valueErr)
				}
				values[index] = value
			}
			result.RobRequests = append(result.RobRequests, RobRequest{
				AllParty:            values[0] != 0,
				LossPercent:         values[1],
				ItemChance:          values[2],
				SelectedPlayerIndex: selectedPlayerIndex,
				SelectedPlayerSet:   selectedPlayerSet,
			})
		case 0x0E, 0x1C, 0x21, 0x27, 0x31, 0x37, 0x3B, 0x3C, 0x3D:
			// PICTURE, CLEARMONSTERS, LOAD FILES, TREASURE, SPRITE OFF and
			// CLEAR BOX have decoded arity but require the full renderer,
			// party/inventory or asset state. Consuming their operands and
			// continuing is a bounded prefix behavior, not a claim of effects.
			if instruction.Command.Opcode == 0x21 {
				for index, operand := range instruction.Operands {
					value, err := operandValue(operand, memory)
					if err != nil {
						return result, fmt.Errorf("load files at %d: %w", pc, err)
					}
					result.LoadFiles[index] = value
				}
				result.LoadFilesRequested = true
				// handler 自己會寫兩格 ECL 記憶體（DOS `overlay-02:0C15h`，
				// spec 1087／1181）：
				//
				//	if o[1] <> 0FFh and o[1] <> 7Fh and bank0^[1CCh] <> 0 then
				//	    bank0^[18Ah] := o[1];   LOAD3DMAP(o[1]);   bank1^[592h] := 0
				//
				// ★ 這兩格**腳本讀得到**，所以少寫就是控制流分歧，不是表現層問題。
				// `7EC9h` 尤其明顯：全 corpus 34 處存取，腳本一路寫 `FFh`，
				// 只有這個 handler 會把它清成 0（`ECL2/0x03:00F6h` 就在
				// `COMPARE 7EC9 FF` 上分岔）。
				piece := result.LoadFiles[0]
				if piece != loadFilesPieceNone && piece != loadFilesPieceSpecial &&
					memory[loadFilesThreeDGate] != 0 {
					memory[loadFilesMapBlockCell] = piece
					recordStore(loadFilesMapBlockCell, piece, pc)
					memory[loadFilesMapStaleCell] = 0
					recordStore(loadFilesMapStaleCell, 0, pc)
					result.LoadFilesLoaded3DMap = true
				}
			}
			if instruction.Command.Opcode == 0x37 {
				for index, operand := range instruction.Operands {
					value, err := operandValue(operand, memory)
					if err != nil {
						return result, fmt.Errorf("load pieces at %d: %w", pc, err)
					}
					result.LoadPieces[index] = value
				}
				result.LoadPiecesRequested = true
				result.WallSetAssignments = WallSetAssignmentsFor(result.LoadPieces, memory)
			}
			if instruction.Command.Opcode == 0x0E {
				value, err := operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("picture at %d: %w", pc, err)
				}
				// 原作 `overlay-02:0841h`（77 條）第一件事就是分兩支：
				// `0FFh` 走**關閉**（`08E9h`：清 `8B62h`／`8B65h`、重繪、
				// `8B48h`／`8B49h` 歸零），其餘走開啟。開啟那一支再依
				// `bank1^[5C2h]`（＝ `7EE1h` 頭像 block）是不是 `0FFh` 決定要不要
				// 走頭像合成，並依 `n >= 78h` 分大圖／一般圖。
				//
				// ★ 先前 `0FFh` 是**什麼都不做**——原作在那裡是把圖關掉。
				if value == 0xFF {
					result.PictureCloseRequested = true
					// 關圖那一支（`08E9h`）有一個不重繪旁路（spec 1148）：
					// `not ((4FBBh = 4) and (4FBAh = 4))` ＝ 前後都還在第一人稱
					// 就跳過整個 if——不重繪、也不清 `8B62h`／`8B65h`，
					// 「圖還開著」這個狀態留著。模式變過（3→4 或 4→3）或現在
					// 是大圖，而且圖真的開著，才重繪並清那兩格。
					bypass := runtime.View.PrevScreenMode == 4 &&
						runtime.View.ScreenMode == 4
					open := runtime.View.Dirty&(ViewDirtyPicture|ViewDirtyWindow) != 0
					if !bypass && open {
						result.PictureCloseRedraw = true
						runtime.View.Dirty &^= ViewDirtyPicture | ViewDirtyWindow
					}
				} else {
					// 開圖立 `8B62h`。同一次執行「先關後開」時，收尾狀態是
					// 開著——關閉訊號讓給後來的開圖。
					runtime.View.Dirty |= ViewDirtyPicture
					result.PictureCloseRequested = false
					result.PictureRequested = true
					result.PictureBlock = value
					result.BigPictureRequested = value >= 0x78
					if head, ok := memory[0x7EE1]; ok {
						result.PictureHeadBlock = head
						result.PictureHeadBlockSet = true
					}
					// 開圖那一支叫 LOADSEQUENCE，它把序列游標設回第 1 格
					// （DOS `overlay-29:0270h`），所以換圖之前推的格數不算。
					result.PictureFrameAdvances = 0
				}
			}
			if instruction.Command.Opcode == 0x1C {
				// CLEARMONSTERS frees the monster chain and resets the placed
				// count; it does not touch what SETUP MONSTER wrote. In DOS
				// (overlay-02:120Eh) it clears 47E6h, 8B69h, 7603h, a 28-byte
				// area and the 6F8Ch chain, while SETUP MONSTER's sprite and
				// picture live in ds:7601h/7602h and bank1 580h/582h, all of
				// which survive (spec 1104 §四).
				//
				// Four corpus sites depend on it: ECL3 0x11 +1154h, ECL3 0x12
				// +06C5h, ECL4 0x21 +05BBh and ECL5 0x32 +077Ah all run
				// SETUP MONSTER before CLEARMONSTERS and then fight. Clearing
				// the setup here dropped the enemy sprite for all four.
				result.MonsterSpawns = nil
				if runtime != nil {
					runtime.MonsterSpawns = nil
				}
				// ★ `1Ch` 的名字只講了一半：它同時把**還沒領走的戰利品堆**丟掉。
				// 原作在 `120Eh` 把 `DS:6F70h` 起的 28 個位元組歸零（七種貨幣／
				// 寶石／珠寶的池，spec 1059），並沿 `DS:6F8Ch` 鏈逐節點
				// `FreeMem(63)`（`27h TREASURE` 串進去的物品節點，spec 1087）。
				//
				// corpus 的慣用法是「先 `1Ch` 清乾淨，再 `LOAD MONSTER` ＋
				// `TREASURE` 擺下一場」，所以清掉同一次執行裡**排在前面**的
				// 戰利品請求才是對的順序——這裡直接清 `result`，順序由執行本身
				// 決定，不必事後猜。
				result.TreasureRequests = nil
				result.ClearMonstersRequested = true
			}
			if instruction.Command.Opcode == 0x27 {
				request := TreasureRequest{}
				for index, operand := range instruction.Operands {
					value, err := operandValue(operand, memory)
					if err != nil {
						return result, fmt.Errorf("treasure at %d: %w", pc, err)
					}
					if index < len(request.Coins) {
						request.Coins[index] = value
					} else {
						request.ItemBlock = value
					}
				}
				result.TreasureRequests = append(result.TreasureRequests, request)
			}
			if instruction.Command.Opcode == 0x3B {
				spellID, err := operandValue(instruction.Operands[0], memory)
				if err != nil {
					return result, fmt.Errorf("spell at %d: %w", pc, err)
				}
				slotAddress, err := operandAddress(instruction.Operands[1])
				if err != nil {
					return result, fmt.Errorf("spell slot address at %d: %w", pc, err)
				}
				characterAddress, err := operandAddress(instruction.Operands[2])
				if err != nil {
					return result, fmt.Errorf("spell character address at %d: %w", pc, err)
				}
				search := SpellSearch{
					SpellID: spellID, SpellSlotAddress: slotAddress, CharacterAddress: characterAddress,
				}
				// ★ 原作把結果寫回這兩格，而**呼叫端只看得懂那兩格**：
				// `ECL4.DAX/0x22 +079Fh` 之後就是
				//   `03 COMPARE [7F79h], imm FFh` ／ `17 IF <>` ／ `01 GOTO`，
				// 接著在 found 那一條路上 `0Ah LOAD CHARACTER [7F7Ah]`。
				// ⇒ slot 那一格的 **0FFh 代表找不到**，character 那一格是
				// `LOAD CHARACTER` 直接吃的隊員索引（`value & 0x7F`，0 起算）。
				// 只排隊不寫回，ECL 會走進「找到了」的分支去 LOAD 一個沒選過的
				// 角色——比沒有效果更糟。
				if workingPartyContext != nil {
					search.Resolved = true
					if slot, member, found := workingPartyContext.findSpell(spellID); found {
						search.Found = true
						search.SlotIndex = slot
						search.MemberIndex = member
						memory[slotAddress] = slot
						memory[characterAddress] = member
					} else {
						memory[slotAddress] = spellNotFound
					}
				}
				result.SpellSearches = append(result.SpellSearches, search)
			}
			if instruction.Command.Opcode == 0x31 {
				// `31h SPRITE OFF` 關掉第一人稱畫面上那隻怪物圖示。原作在
				// `ECL2.DAX/0x02 +0CCBh` 用它接 `2Dh CALL 2E10h`（畫面提交點）
				// 再開新頁：逃走之後那隻怪物不該還留在畫面上。
				result.SpriteOffRequested = true
				// `31h SPRITE OFF` 清 `8B65h`（spec 1150 的旗標表）。
				runtime.View.Dirty &^= ViewDirtyWindow
			}
			if instruction.Command.Opcode == 0x3D {
				// `3Dh CLEAR BOX` 把文字框清空但**不印任何東西**。原作用它在
				// 換畫面之前把上一段訊息擦掉（`ECL1.DAX/0x52 +001Bh` 就是
				// CLEAR BOX → PICTURE → ADD NPC → PRINTCLEAR）。
				// 沒有它的話，玩家會在新畫面底下看到上一幕殘留的文字。
				result.ClearBoxRequested = true
			}
			if instruction.Command.Opcode == 0x3C {
				address, err := operandAddress(instruction.Operands[0])
				if err != nil {
					return result, fmt.Errorf("protection address at %d: %w", pc, err)
				}
				result.ProtectionRequests = append(result.ProtectionRequests, address)
			}
		default:
			return result, fmt.Errorf("unsupported opcode 0x%02X at payload offset %d", instruction.Command.Opcode, pc)
		}
		pc = next
		result.PC = pc
	}
	return result, fmt.Errorf("runtime step limit %d reached at payload offset %d", maxSteps, pc)
}

func normalizeInputString(value string, maxLength uint16) string {
	value = strings.ToUpper(value)
	runes := []rune(value)
	if len(runes) > int(maxLength) {
		runes = runes[:maxLength]
	}
	return string(runes)
}

func resolveEncounterAction(selection uint16, behavior, maxDistance uint16) uint16 {
	switch {
	case selection == 0 && behavior != 2:
		// COMBAT closes distance and fights for all fully understood behavior
		// modes. Mode 2 also depends on relative group movement.
		return 1
	case selection == 1 && maxDistance == 0 && behavior == 4:
		// At contact distance WAIT gives the monsters the initiative to talk.
		return 3
	case selection == 2 && behavior == 1:
		return 2
	case selection == 3 && maxDistance == 0 && behavior == 4:
		return 3
	default:
		return behavior
	}
}

func operandValue(operand Operand, memory map[uint16]uint16) (uint16, error) {
	switch operand.Code {
	case 0x00:
		return uint16(operand.Low), nil
	case 0x01, 0x03:
		if !operand.WordSet {
			return 0, fmt.Errorf("memory operand has no address")
		}
		return memory[operand.Word], nil
	case 0x02:
		if !operand.WordSet {
			return 0, fmt.Errorf("literal operand has no word")
		}
		return operand.Word, nil
	case 0x81:
		return 0, fmt.Errorf("string-memory operand cannot be used as a numeric value")
	default:
		return 0, fmt.Errorf("unsupported value operand code 0x%02X", operand.Code)
	}
}

func operandAddress(operand Operand) (uint16, error) {
	if !operand.WordSet || (operand.Code != 0x01 && operand.Code != 0x02 && operand.Code != 0x03 && operand.Code != 0x81) {
		return 0, fmt.Errorf("operand is not a word address")
	}
	return operand.Word, nil
}

func operandIsText(operand Operand) bool {
	return operand.Code == 0x80 || operand.Code == 0x81
}

func operandText(operand Operand, stringsMemory map[uint16]string) (string, error) {
	if operand.Code == 0x80 {
		return DecodePackedText(operand.Packed), nil
	}
	if operand.Code == 0x81 && operand.WordSet {
		return stringsMemory[operand.Word], nil
	}
	return "", fmt.Errorf("unsupported string operand code 0x%02X", operand.Code)
}

// 兩槽分支的閘門是 `bank0^[1CEh]` 與 `[1D0h]`。用 spec 1098／1155 的換算
// `bank0 + (位址 − 4B00h) × 2` 對回 ECL 格：`1CEh` ⇒ `4BE7h`、`1D0h` ⇒ `4BE8h`。
const (
	wallSetTwoSlotGateA uint16 = 0x4BE7
	wallSetTwoSlotGateB uint16 = 0x4BE8
	// 片號 `7Fh` 是 handler 的特例，`0FFh` 是「這一槽沒有牆面組」。
	wallSetPieceSpecial uint16 = 0x7F
	wallSetPieceNone    uint16 = 0xFF
)

// WallSetAssignment 是 `37h LOAD PIECES` 對**一個**牆面組槽位做的事。
type WallSetAssignment struct {
	// Slot 是 1..3，與運算元同號（原作 `LOADWALLSET(i, o[i])`）。
	Slot uint8 `json:"slot"`
	// Piece 是要載的片號；`Sentinel` 為 true 時沒有意義。
	Piece uint16 `json:"piece"`
	// Sentinel ＝ 這一槽不呼叫 `LOADWALLSET`，由 handler 自己把
	// `[7210h+槽×4]` 與 `[7212h+槽×4]` 都寫成 `0FFFFh`。
	Sentinel bool `json:"sentinel,omitempty"`
}

// WallSetAssignmentsFor 照抄 `37h LOAD PIECES` 的三支分派（DOS
// `overlay-02:0C15h`，spec 1087／1153）：
//
//	if o[1] = 7Fh then                       LOADWALLSET(1, 0)
//	else if bank0^[1CEh] <> 0 and [1D0h] <> 0 then
//	    只載槽 1 與槽 3（各自 <> 0FFh 才載）
//	else
//	    for i := 1 to 3：<> 0FFh 就載，= 0FFh 就寫哨兵
//
// ★ 前兩支**都不寫哨兵**，也都不碰沒提到的槽——所以回傳的是「這一次要動哪些
// 槽」而不是三格一次寫滿。全 corpus 23 處都走第三支（`7Fh` 一次都沒出現，
// 而槽 2 每一處都帶著真實片號），前兩支照 handler 寫、沒有實機路徑背書。
func WallSetAssignmentsFor(pieces [3]uint16, memory map[uint16]uint16) []WallSetAssignment {
	if pieces[0] == wallSetPieceSpecial {
		return []WallSetAssignment{{Slot: 1, Piece: 0}}
	}
	if memory[wallSetTwoSlotGateA] != 0 && memory[wallSetTwoSlotGateB] != 0 {
		out := make([]WallSetAssignment, 0, 2)
		if pieces[0] != wallSetPieceNone {
			out = append(out, WallSetAssignment{Slot: 1, Piece: pieces[0]})
		}
		if pieces[2] != wallSetPieceNone {
			out = append(out, WallSetAssignment{Slot: 3, Piece: pieces[2]})
		}
		return out
	}
	out := make([]WallSetAssignment, 0, 3)
	for index, piece := range pieces {
		slot := uint8(index + 1)
		if piece == wallSetPieceNone {
			out = append(out, WallSetAssignment{Slot: slot, Sentinel: true})
			continue
		}
		out = append(out, WallSetAssignment{Slot: slot, Piece: piece})
	}
	return out
}

// `21h LOAD FILES` 的三格。位址用 spec 1098／1155 的換算對回 ECL 格：
// bank0 是 `4B00h + 位移 / 2`，bank1 是 `7C00h + 位移 / 2`。
const (
	// loadFilesThreeDGate ＝ `bank0^[1CCh]`：非零代表現在是第一人稱（3D 地圖）
	// 模式，handler 才會去載 3D 地圖；為零時改走載大圖那一路。
	// 全 corpus 22 處寫入，全部在 ECL1 的三個世界地圖 block（`0x50`／`0x51`／`0x52`）。
	loadFilesThreeDGate uint16 = 0x4BE6
	// loadFilesMapBlockCell ＝ `bank0^[18Ah]`：剛載進來的地圖 block 編號。
	loadFilesMapBlockCell uint16 = 0x4BC5
	// loadFilesMapStaleCell ＝ `bank1^[592h]`：腳本寫 `FFh` 表示「地圖要重載」，
	// handler 載完就清成 0。corpus 34 處存取。
	loadFilesMapStaleCell uint16 = 0x7EC9

	loadFilesPieceNone    uint16 = 0xFF
	loadFilesPieceSpecial uint16 = 0x7F
)
