// Package combat contains the platform-neutral AD&D combat core. ECL and
// Ebiten adapters can provide party/enemy data without embedding rendering or
// DOS memory assumptions here.
package combat

import (
	"errors"
	"fmt"
	"sort"

	engineaction "github.com/wicanr2/golden-box-remake-engine/combat/action"
	enginedamage "github.com/wicanr2/golden-box-remake-engine/combat/damage"
	engineinitiative "github.com/wicanr2/golden-box-remake-engine/combat/initiative"
	enginemodifier "github.com/wicanr2/golden-box-remake-engine/combat/modifier"
	enginespell "github.com/wicanr2/golden-box-remake-engine/combat/monsterspell"
	engineposthit "github.com/wicanr2/golden-box-remake-engine/combat/posthit"
	enginequickspell "github.com/wicanr2/golden-box-remake-engine/combat/quickspell"
	enginequicktarget "github.com/wicanr2/golden-box-remake-engine/combat/quicktarget"
	engineresistance "github.com/wicanr2/golden-box-remake-engine/combat/resistance"
	enginesleep "github.com/wicanr2/golden-box-remake-engine/combat/sleep"
	enginerandom "github.com/wicanr2/golden-box-remake-engine/randomstream"
)

// ErrAdjacentMissileTarget identifies the RuleBook's recoverable range
// restriction so UI adapters can present a localized message.
var ErrAdjacentMissileTarget = errors.New("missile weapon cannot attack an adjacent target")

type Side uint8

const (
	SideParty Side = iota
	SideEnemy
)

type Status uint8

const (
	StatusActive Status = iota
	StatusPartyWon
	StatusEnemyWon
	StatusDraw
	// StatusPartyFled ＝ 還活著的隊員全部走出戰場邊界（spec 799／1112）。
	// 與 StatusEnemyWon 分開是因為結果不同：沒有經驗值、沒有寶物，
	// 但隊伍也沒有全滅。
	StatusPartyFled
)

type Fighter struct {
	ID         string
	Name       string
	SourceName string `json:"source_name,omitempty"`
	Side       Side
	// LegacyObjectID is the one-based OBJECTLIST/IDLIST index rebuilt from the
	// original CHARACTERLIST traversal at combat start. Zero means the title
	// adapter has not established that legacy identity.
	LegacyObjectID uint8
	// QuickFight delegates this fighter's turn to combat AI even when it is
	// on the party side. ECL uses this for allied NPCs in mixed-team battles.
	QuickFight bool
	// ControlMorale preserves the player record control byte. Values below
	// 0x80 are manually controllable PCs; values at or above 0x80 are NPCs.
	ControlMorale uint8
	// TemporaryAlly is encounter-scoped and must not enter the persistent
	// adventuring party after combat continuation.
	TemporaryAlly bool
	Evil          bool
	Good          bool
	// SpriteSet/SpriteBlock identify the original CPIC asset when the fighter
	// came from a LOAD MONSTER descriptor. A zero block means the renderer may
	// choose a deterministic party fallback sprite.
	SpriteSet   uint8
	SpriteBlock uint8
	// AnimationBlock is the SPRIT block from SETUP MONSTER. It is separate
	// from SpriteBlock because the original engine loads CPIC and SPRIT from
	// different DAX families.
	AnimationBlock uint8
	HasAnimation   bool
	// Party icon fields mirror the original Player combat record. They are
	// renderer-neutral so save/import decoding can fill the actual values.
	HasPartyIcon    bool
	PartyHeadBlock  uint8
	PartyBodyBlock  uint8
	PartyIconID     uint8
	PartyIconSize   uint8
	PartyIconColors [6]uint8
	IconDirection   uint8
	IconAttack      bool
	// MonsterAffects preserves raw MON*SPC records. Gameplay projections are
	// intentionally left to later, verified rules adapters.
	MonsterAffects []MonsterAffect
	// MonsterItems preserves raw MON*ITM records. 原作的 AI 換裝（spec 1120）
	// 讀的就是這條物品鏈；隊伍側早就接上了，怪物側先前連資料都沒有。
	// ⚠ 掛上資料不等於接上規則：換裝之後的派生值重算還沒做。
	MonsterItems []MonsterItem
	// CombatMap position/size. A future Area/ECL placement decoder can set
	// these directly; StartCombat supplies a deterministic fallback otherwise.
	HasCombatPosition bool
	CombatX           int
	CombatY           int
	CombatSize        uint8
	// LargeTarget 為 true 時，攻擊它的人要用武器的**大型**傷害三連
	// （原作 `overlay-13:15EFh`，spec 1175）。隊伍成員一律 false：可玩種族的
	// `+0DEh` 在原版資料裡全部是 `01h`。
	LargeTarget bool
	// DeathOverlay requests the renderer's downed/death visual. The original
	// CombatantKilled routine draws an animated skull; keeping this as a
	// signal lets each frontend choose an asset without leaking CPIC indices
	// into the combat core.
	DeathOverlay bool
	// DownedCorpse marks a team party fighter whose map tile became the
	// reference Tile_DownPlayer (0x1F). Ordinary healing clears the overlay
	// flash but does not restore combat placement; combat_heal/placement must
	// clear this marker separately.
	DownedCorpse bool
	// DownOverkill 是把這名戰鬥員打倒的那一擊的**溢出量**（傷害 − 當時 HP）。
	// 原作的 SAVEDAMAGE（PC-98 `overlay-24:2658h`，spec 1205）用它決定倒下的
	// 形式：>9 死亡、1..9 瀕死（並記出血）、0 昏迷。倒下後的追擊不覆寫；
	// 治療站起來後再倒會重記。
	DownOverkill int `json:"down_overkill,omitempty"`
	CombatAction ActionState
	// 原作戰鬥狀態記錄（`+18Dh` 指到的 22 bytes）的三格，spec 1137：
	// `+09h` 面向、`+0Fh` 這一回合轉向過幾次、`+12h` 累計轉向（0..7 環狀）。
	// 後兩格在回合開始清 0；`+09h` 跨回合保留。
	CombatFacing      uint8
	CombatActionCount uint8
	CombatTurnTotal   uint8
	// ArmorClassFacing 是原作角色記錄的第二個 AC 欄位（`+19Bh`）。攻擊結算依
	// `RearAttackApplies` 在它與 `ArmorClass`（`+19Ah`）之間挑一個。
	// 原作算的是 `+19Bh ＝ +19Ah − 敏捷防禦調整 − 盾牌那一槽 − 2`（spec 1000 §七），
	// 換成這裡的畫面刻度就是**比 `ArmorClass` 大**——也就是比較好打。
	// ⚠ 隊員那一側目前沒有這一格，所以 `ArmorClassFacingKnown` 為 false 時
	// 攻擊結算不會改用它。0 是合法的 AC，不能拿零值當「沒有」。
	ArmorClassFacing      int
	ArmorClassFacingKnown bool
	HitPoints             int
	MaxHitPoints          int
	// HitDice preserves the original Player/MON*CHA byte at offset 0xE5.
	// Poisonous-cloud rules consume it directly.
	HitDice uint8
	// RawPlayer74 preserves the shared Player record byte consumed by the
	// original Sleep handler's five-HD cost branch. Its gameplay name remains
	// unresolved; adapters must not substitute race or another inferred field.
	RawPlayer74 uint8
	// RaceType preserves the Borland CHARREC.RACETYPE byte at +11A. The
	// historical MonsterType field remains as a compatibility alias for old
	// saves and synthetic fixtures; it is not the CHARREC.MONSTERTYPE byte.
	RaceType      uint8 `json:"race_type,omitempty"`
	RaceTypeKnown bool  `json:"race_type_known,omitempty"`
	MonsterType   uint8
	// Alignment is CHARREC.ALIGNMENT at +11B. Zero is LAWFUL_GOOD, so the
	// explicit known bit is required before a conditional effect may match.
	Alignment      uint8 `json:"alignment,omitempty"`
	AlignmentKnown bool  `json:"alignment_known,omitempty"`
	// RawItemCount 原樣保留角色記錄的 `+14Ch`（DOS 版面是 `NUMITEMS`，
	// spec 1166），不指派任何規則語意。
	RawItemCount uint8 `json:"raw_item_count,omitempty"`
	// Dexterity preserves the shared Player byte at +17 used by the original
	// initiative reaction table. It is deliberately not converted to a modern
	// ability modifier.
	Dexterity uint8
	// CombatTeam preserves the Action scheduler team number. The current CoAB
	// adapter uses Side while the area surprise-mask writer remains unresolved.
	// ⚠ 「未解」的範圍已經縮小了（spec 1113）：36 個 overlay 裡對區域記錄
	// `+596h`（突襲遮罩）只有兩次存取——先攻擲骰**讀**它，`overlay-08:sub_F3`
	// 把它**清成 0**。沒有任何 overlay 設它。所以目前傳 0 是有證據的預設值，
	// 缺的是常駐段那一側的掃描。
	CombatTeam           uint8
	ArmorClass           int
	AttackBonus          int
	Blessed              bool
	BlessRounds          int
	Cursed               bool
	CurseRounds          int
	ProtectedFromEvil    bool
	ProtectionEvilRounds int
	ProtectedFromGood    bool
	ProtectionGoodRounds int
	DamageDiceCount      int
	DamageDiceSides      int
	DamageBonus          int
	// 大型目標用的另一組傷害三連（類別表 `+02h`..`+04h`）。只有**槽 0 有裝備中
	// 武器**時才換——原作那一段整個包在 `if 攻擊者^[151h] <> NIL` 裡（spec 1175）。
	LargeDamageDiceCount int
	LargeDamageDiceSides int
	LargeDamageBonus     int
	// HasSlotZeroWeapon 就是 `攻擊者^[151h] <> NIL`。
	HasSlotZeroWeapon bool
	AttacksPerTurn    int
	// AttackBlows 是原作的 `BASEATTBLOWS[0..1]`（`+11Ch`）：兩個武器槽的基準
	// 攻擊次數，單位是**半次**。本回合的整數次數由 `AdjustBlows` 依回合奇偶
	// 換算（spec 1180）。
	AttackBlows    [2]int `json:"attack_blows,omitempty"`
	AmmunitionType uint8
	// AmmunitionCount 是**架著的那一件彈藥**的原始數量（物品 `+39h`）。
	// 遠程攻擊的次數會被它壓住（spec 808／1180）。
	AmmunitionCount   int `json:"ammunition_count,omitempty"`
	MovementAllowance int
	WeaponRange       int
	MissileWeapon     bool
	ThrownWeapon      bool
	// NaturalDamage* 是放下武器時的天生攻擊（怪物記錄的基準骰，spec 1174）。
	// AI 換裝把槽 0 武器收起來之後，重投影用它還原現值。
	NaturalDamageDiceCount int
	NaturalDamageDiceSides int
	NaturalDamageBonus     int
	// ClassUsabilityMask 是記錄 `+12Bh`：能用哪些物品類別（spec 1004 §二）。
	ClassUsabilityMask uint8
	// WeaponItemType 是架著的那把武器的**物品類別**（`CHARITEMREC.ITEMPTR`，
	// remake 的 `ItemRecord.Type`）。原作的 `SHOWARROW` 用它決定投射動畫尾端
	// 放哪一個音效（spec 1186）：投石索類放哨音、擲斧／棍／鎚放揮擊、其餘放箭。
	WeaponItemType uint8
	// InitiativeBonus is retained only as a legacy synthetic-fixture ordering
	// seam. Production party/MON adapters never set it. Nonzero fixture values
	// replace the rolled delay after the exact d6 draw, while d100 scan traffic
	// remains unchanged; new tests should assert DEX and RNG directly instead.
	InitiativeBonus int
	// SavingThrows preserves the five reference saveVerse thresholds:
	// poison, petrification, rod/staff/wand, breath weapon and spell.
	SavingThrows     []uint8
	SavingThrowBonus int
	// AIMode 是原作戰鬥狀態 `+15h` 的行為模式 1..6（spec 830／837）：
	// 它決定移動時「試方向的順序」。每回合由 `BeginMonsterTurn` 重骰，
	// 0 代表還沒骰過（原作的初值也是 0，第一回合必骰）。
	AIMode int `json:"ai_mode,omitempty"`
	// RetreatedThisRound 是原作戰鬥狀態 `+18Dh^[14h]`：本回合士氣檢定沒過而且
	// 跑得掉。每回合開頭由 `CheckMorale` 清掉，所以它只描述當下這一回合。
	RetreatedThisRound bool `json:"retreated_this_round,omitempty"`
	// RawCombatState10 是原作戰鬥狀態記錄 `+18Dh^[10h]` 那一個位元組。
	// 語意未解讀，但它同時是兩種障礙地形的通行豁免（spec 1119），所以照原樣
	// 留著：日後解出語意時只要有東西寫它，移動那一側就自動生效。
	RawCombatState10 uint8 `json:"raw_combat_state_10,omitempty"`
	// UndeadType 是角色記錄 `+0E9h`（`UNDEADLEVEL`）：1..10 是驅散矩陣的列，
	// 0 代表不是不死生物（spec 834／1164）。
	UndeadType uint8 `json:"undead_type,omitempty"`
	// ClericLevel 是牧師槽的等級（角色記錄 `+109h`）。只有它 > 0 的人看得到
	// 戰鬥選單的「退散」（spec 905）。
	ClericLevel uint8 `json:"cleric_level,omitempty"`
	// SecondClassLevel 是第二職業的等級（`+111h`）；換算驅散矩陣的欄要用到
	// 合計等級。
	SecondClassLevel uint8 `json:"second_class_level,omitempty"`
	// TriedToTurn 是 22 bytes 戰鬥狀態的 `+11h`（`TRIEDTOTURN`）：這一場已經
	// 驅散過了，選單就不再出現（spec 834／905／1165）。
	TriedToTurn bool `json:"tried_to_turn,omitempty"`
	// turnOrderIndex 只在挑驅散目標時暫存，不進存檔。
	turnOrderIndex int
	// Escaped 是「走出戰場邊界離開這一場」（spec 799／1112）。與死亡不同：
	// 人還活著，只是不在戰場上，所以不算敵方戰果、也不再是任何效果的目標。
	Escaped bool `json:"escaped,omitempty"`
	// CoughingTurns and HelplessTurns are action-counted combat effects.
	// Persistent-area rules set them; the game adapter consumes one turn only
	// when the affected combatant actually reaches its initiative.
	CoughingTurns int
	HelplessTurns int
	// MonsterSpellIDs mirrors the raw MON*CHA spell-list slots. The bounded
	// monster-turn adapter currently consumes only Magic Missile (0x0F).
	MonsterSpellIDs  []uint8
	MonsterSpellUses [3]uint8
	// ReadiedItemEffects 是**裝備中**物品的效果槽（物品記錄 `+3Dh`），
	// 已經過 spec 835 的過濾：`+34h ≠ 0`（裝備中）、`+3Dh > 0`、`+3Eh < 80h`。
	// AI 用道具的決策只需要這一欄，不需要整條物品鏈。
	ReadiedItemEffects []uint8 `json:"readied_item_effects,omitempty"`
	// DamageRules are immutable title-pack capabilities. They are deliberately
	// excluded from save JSON and reattached by the game adapter after load.
	DamageRules []enginedamage.Rule `json:"-"`
	// ConditionalModifierRules are immutable game-pack capabilities. They are
	// excluded from save JSON and reattached after loading the active battle.
	ConditionalModifierRules []enginemodifier.Rule `json:"-"`
	// MagicResistanceRules are immutable game-pack capabilities. They are
	// excluded from save JSON and reattached after loading the active battle.
	MagicResistanceRules []engineresistance.Rule `json:"-"`
	// PostHitRules are immutable game-pack capabilities. They are excluded from
	// save JSON and reattached after loading the active battle.
	PostHitRules []engineposthit.Rule `json:"-"`
	// SpecialAttackRules 是 pack 宣告的特殊攻擊（spec 1202），掛法同
	// MonsterSpellRules；SpecialAttackUses 對應 `arg_2^[3]` 的每場次數欄
	//（區域吐息類 3 次，最後成功才扣）。
	SpecialAttackRules   []SpecialAttackRule `json:"-"`
	SpecialAttackUses    int                 `json:"special_attack_uses,omitempty"`
	SpecialAttackUsesSet bool                `json:"special_attack_uses_set,omitempty"`
	// MonsterSpellRules are immutable game-pack capabilities for innate special
	// spell actions. They are excluded from save JSON and reattached after load.
	MonsterSpellRules []enginespell.Rule `json:"-"`
}

// MonsterAffect mirrors one nine-byte MON*SPC record without importing the
// monster data package into the combat core.
type MonsterAffect struct {
	Kind     uint8
	Value    uint16
	Duration uint16
	Strength uint8
	// Raw4 preserves EFFECTREC byte +4. Dynamic spell effects write caster
	// level here; MON*SPC templates commonly contain zero without being inert.
	Raw4   uint8
	Active bool
	// Innate marks an effect loaded from a MON*SPC monster template. The
	// reference LOADMONSTER preserves byte 4 as zero while retaining every
	// template effect in the runtime list, so that byte cannot be used to
	// suppress a monster's innate effects in the combat projection.
	Innate bool
	Data   [4]byte
}

// MonsterItem 保留 `MON*ITM` 的一筆物品。欄位與 `monster.ItemRecord` 對應，
// 但**型別定義在這裡**——`monster` 匯入 `combat`（`BuildEnemies` 回傳
// `[]Fighter`），反向匯入會成環，所以跟 `MonsterAffect` 一樣在這一側鏡射。
//
// ⚠ 這是**原始資料**，不是規則。掛上去只代表「這隻怪身上有這些東西」；
// 換裝、掉落與使用各自要另外接。
type MonsterItem struct {
	Name    string
	Type    uint8
	Plus    int
	Readied bool
	Cursed  bool
	Count   uint8
	Weight  int16
	Value   int16
	Affects [3]uint8
}

func (a MonsterAffect) operational() bool { return a.Active || a.Innate }

const (
	AlignmentLawfulGood     uint8 = 0
	AlignmentLawfulNeutral  uint8 = 1
	AlignmentLawfulEvil     uint8 = 2
	AlignmentNeutralGood    uint8 = 3
	AlignmentTrueNeutral    uint8 = 4
	AlignmentNeutralEvil    uint8 = 5
	AlignmentChaoticGood    uint8 = 6
	AlignmentChaoticNeutral uint8 = 7
	AlignmentChaoticEvil    uint8 = 8
)

func (f Fighter) isEvil() bool {
	if f.AlignmentKnown {
		return f.Alignment == AlignmentLawfulEvil || f.Alignment == AlignmentNeutralEvil || f.Alignment == AlignmentChaoticEvil
	}
	return f.Evil
}

func (f Fighter) isGood() bool {
	if f.AlignmentKnown {
		return f.Alignment == AlignmentLawfulGood || f.Alignment == AlignmentNeutralGood || f.Alignment == AlignmentChaoticGood
	}
	return f.Good
}

func (f Fighter) raceType() uint8 {
	if f.RaceTypeKnown {
		return f.RaceType
	}
	return f.MonsterType
}

// MonsterConditionalModifierAgainst evaluates effects owned by f against the
// character currently interacting with f. PC-98 effect 08h/09h reads the
// active character's alignment while updating shared SAVE/TO-HIT cells; the
// owner and the interacting character therefore remain separate inputs.
func (f Fighter) MonsterConditionalModifierAgainst(interacting Fighter) enginemodifier.Result {
	active := make([]uint8, 0, len(f.MonsterAffects))
	for _, affect := range f.MonsterAffects {
		if affect.operational() {
			active = append(active, affect.Kind)
		}
	}
	return enginemodifier.Resolve(active, enginemodifier.Subject{
		Value: interacting.Alignment, Known: interacting.AlignmentKnown,
	}, f.ConditionalModifierRules)
}

// MonsterCanDetectInvisible reports the verified effect-18 capability used by
// the original CanHitTarget path.
func (f Fighter) MonsterCanDetectInvisible() bool {
	for _, affect := range f.MonsterAffects {
		if affect.operational() && affect.Kind == 0x18 {
			return true
		}
	}
	return false
}

// MonsterMagicResistanceRule resolves active raw effects through the
// title-owned game-pack contract. The combat core does not infer resistance
// from a monster name or from a hardcoded effect kind.
func (f Fighter) MonsterMagicResistanceRule() (engineresistance.Result, bool) {
	active := make([]uint8, 0, len(f.MonsterAffects))
	for _, affect := range f.MonsterAffects {
		if affect.operational() {
			active = append(active, affect.Kind)
		}
	}
	result := engineresistance.Resolve(active, f.MagicResistanceRules)
	return result, result.Matched
}

// MonsterMagicResistanceBase reports the configured percentage base for
// compatibility with callers that only need the numeric projection. The
// effect-to-base mapping is supplied by JSON through MagicResistanceRules.
func (f Fighter) MonsterMagicResistanceBase() (int, bool) {
	result, matched := f.MonsterMagicResistanceRule()
	if !matched {
		return 0, false
	}
	return result.Base, true
}

const (
	DamageFlagFire        uint8 = 0x01
	DamageFlagCold        uint8 = 0x02
	DamageFlagElectricity uint8 = 0x04
	DamageFlagMagic       uint8 = 0x08
)

// MonsterDamageAdjustment applies the evidence-backed title-pack effect rules
// to one pending damage value. The raw effect list remains the source of
// active kinds; the operation itself comes from JSON through DamageRules.
func (f Fighter) MonsterDamageAdjustment(flags uint8, damage int) enginedamage.Result {
	active := make([]uint8, 0, len(f.MonsterAffects))
	for _, affect := range f.MonsterAffects {
		if affect.operational() {
			active = append(active, affect.Kind)
		}
	}
	return enginedamage.Resolve(damage, flags, active, f.DamageRules)
}

// MonsterProtectedFromDamage preserves the legacy boolean query for callers
// that only need complete immunity. Half damage (effect 0Ah) is not reported
// as protected.
func (f Fighter) MonsterProtectedFromDamage(flags uint8) bool {
	return f.MonsterDamageAdjustment(flags, 0).Immune
}

// MonsterPostHitRules resolves operational raw effects through the title-owned
// post-hit contract. The raw MON*SPC effects remain the source of active kinds;
// slot ranges and damage behavior come from the game pack.
func (f Fighter) MonsterPostHitRules(attackSlot int) []engineposthit.Rule {
	active := make([]uint8, 0, len(f.MonsterAffects))
	for _, affect := range f.MonsterAffects {
		if affect.operational() {
			active = append(active, affect.Kind)
		}
	}
	return engineposthit.Resolve(active, attackSlot, f.PostHitRules)
}

// MonsterPostHitAffects preserves the legacy raw-affect projection for callers
// that need to inspect the original MON*SPC records. Dispatch eligibility is
// now supplied by PostHitRules rather than a hardcoded effect kind or slot.
func (f Fighter) MonsterPostHitAffects(attackSlot int) []MonsterAffect {
	rules := f.MonsterPostHitRules(attackSlot)
	if len(rules) == 0 {
		return nil
	}
	matched := make(map[uint8]bool, len(rules))
	for _, rule := range rules {
		matched[rule.EffectKind] = true
	}
	var effects []MonsterAffect
	for _, affect := range f.MonsterAffects {
		if affect.operational() && matched[affect.Kind] {
			effects = append(effects, affect)
		}
	}
	return effects
}

// MonsterSpecialSpellRules resolves innate special spell actions from active
// raw effects and the current combat round. The title pack owns effect kind,
// spell ID, round window and payload.
func (f Fighter) MonsterSpecialSpellRules(round int) []enginespell.Rule {
	active := make([]uint8, 0, len(f.MonsterAffects))
	for _, affect := range f.MonsterAffects {
		if affect.operational() {
			active = append(active, affect.Kind)
		}
	}
	return enginespell.Resolve(active, round, f.MonsterSpellRules)
}

// MonsterThrowsLightning preserves the legacy capability query for the first
// combat round. Dispatch itself must use MonsterSpecialSpellRules with the
// actual round, so the method does not infer behavior from a hardcoded kind.
func (f Fighter) MonsterThrowsLightning() bool {
	return len(f.MonsterSpecialSpellRules(1)) != 0
}

// MagicResistanceChance mirrors the PC-98 EFFPROCS common routine. The
// original compares a d100 roll directly against this signed expression and
// does not clamp it before the comparison.
func MagicResistanceChance(base, casterLevel int) int {
	return engineresistance.LevelAdjustedD100Chance(base, casterLevel)
}

func (b *Battle) rollMagicResistance(target Fighter, casterLevel int) bool {
	rule, matched := target.MonsterMagicResistanceRule()
	if !matched {
		return false
	}
	if rule.Formula != engineresistance.FormulaLevelAdjustedD100 {
		return false
	}
	return b.rng.Intn(100)+1 <= engineresistance.LevelAdjustedD100Chance(rule.Base, casterLevel)
}

// CastSleepOrdered applies only the evidence-backed post-targeting Sleep
// transaction. orderedTargetIDs must come from the title adapter's verified
// geometry; this method never discovers, sorts, or broadens candidates.
func (b *Battle) CastSleepOrdered(casterID string, orderedTargetIDs []string, casterLevel int) (SleepResult, error) {
	if b == nil {
		return SleepResult{}, fmt.Errorf("battle is nil")
	}
	caster, ok := b.fighters[casterID]
	if !ok {
		return SleepResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return SleepResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 {
		return SleepResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if casterLevel < 1 || casterLevel > 255 || casterLevel*5 > int(^uint16(0)) {
		return SleepResult{}, fmt.Errorf("caster level %d is outside the effect record range", casterLevel)
	}
	candidates := make([]enginesleep.Candidate, 0, len(orderedTargetIDs))
	seen := make(map[string]struct{}, len(orderedTargetIDs))
	for _, targetID := range orderedTargetIDs {
		if _, duplicate := seen[targetID]; duplicate {
			return SleepResult{}, fmt.Errorf("duplicate Sleep target %q", targetID)
		}
		seen[targetID] = struct{}{}
		target, exists := b.fighters[targetID]
		if !exists {
			return SleepResult{}, fmt.Errorf("unknown Sleep target %q", targetID)
		}
		candidates = append(candidates, enginesleep.Candidate{
			ID: targetID, HitDice: int(target.HitDice), AlreadyHeld: target.MonsterIsHeld(),
			DoubleFiveHitDiceCost: target.HitDice == 5 && target.RawPlayer74 != 0,
		})
	}
	capacity, err := enginesleep.RollCapacity(func(sides int) (int, error) {
		return b.rng.Intn(sides) + 1, nil
	})
	if err != nil {
		return SleepResult{}, err
	}
	filtered := enginesleep.Filter(capacity, candidates)
	result := SleepResult{
		CasterID: casterID, SpellID: SleepSpellID, InitialCapacity: capacity,
		RemainingCapacity: filtered.RemainingCapacity,
	}
	duration := uint16(casterLevel * 5)
	for _, targetID := range filtered.Selected {
		target := b.fighters[targetID]
		resisted := false
		resisted = b.rollMagicResistance(target, casterLevel)
		impact := SleepImpact{TargetID: targetID, Resisted: resisted}
		if !resisted {
			target.MonsterAffects = append(target.MonsterAffects, MonsterAffect{
				Kind: 0x35, Value: duration, Duration: duration,
				Strength: 1, Raw4: uint8(casterLevel), Active: true,
			})
			// Effect 35h dispatches the shared CLEARACTION callback for both
			// add and remove phases. It does not create a slot-consuming spell
			// interruption event.
			b.clearActionState(&target)
			b.fighters[targetID] = target
			impact.Duration = duration
		}
		result.Impacts = append(result.Impacts, impact)
	}
	return result, nil
}

// RaceTypeAnimal is the verified RACETYPE category used by the effect-45
// visibility branch. MonsterTypeAnimal remains as a source-compatible alias
// for callers written before the Borland member table corrected +11A.
const RaceTypeAnimal uint8 = 0x13

const MonsterTypeAnimal uint8 = RaceTypeAnimal

// VisibleTo reports the status-effect visibility used by the original
// CHECKTARGET path. Effect 19 can be defeated by an operational effect 18 on
// the observer; effect 47 sets the target-hidden flag unconditionally.
func (f Fighter) VisibleTo(observer Fighter) bool {
	for _, affect := range f.MonsterAffects {
		if !affect.operational() {
			continue
		}
		switch affect.Kind {
		case 0x25:
			if f.CombatAction.Delay == 0 {
				return false
			}
		case 0x19:
			if !observer.MonsterCanDetectInvisible() {
				return false
			}
		case 0x45:
			if observer.raceType() == RaceTypeAnimal && !observer.MonsterCanDetectInvisible() {
				return false
			}
		case 0x47:
			return false
		}
	}
	return true
}

// MonsterAffectArmorClassBonusAgainst projects only the effect kinds whose
// attack-roll behavior is verified in the original effect handlers.
func (f Fighter) MonsterAffectArmorClassBonusAgainst(attacker Fighter) int {
	bonus := 0
	for _, affect := range f.MonsterAffects {
		if !affect.operational() {
			continue
		}
		switch affect.Kind {
		case 0x19:
			if !attacker.MonsterCanDetectInvisible() {
				bonus += 4
			}
		case 0x47:
			bonus += 4
		case 0x45:
			if attacker.raceType() == RaceTypeAnimal {
				bonus += 4
			}
		}
	}
	return bonus
}

// MonsterAffectForcesAttackMiss reports effects that replace the attack roll
// instead of applying an AC modifier. The original blink handler writes FFh
// after natural-roll handling when the target action delay is zero.
func (f Fighter) MonsterAffectForcesAttackMiss() bool {
	for _, affect := range f.MonsterAffects {
		if affect.operational() && affect.Kind == 0x25 && f.CombatAction.Delay == 0 {
			return true
		}
	}
	return false
}

// MonsterAffectAttackBlows projects the verified Haste／Slow multiplier over
// the record's `BASEATTBLOWS`. 加速與緩速動的是**半次**單位，所以乘除都在
// 換算之前做——先換算再加倍會把 1.5 次的那半次先丟掉（spec 1180）。
func (f Fighter) MonsterAffectAttackBlows() [2]int {
	blows := f.AttackBlows
	for _, affect := range f.MonsterAffects {
		if !affect.operational() {
			continue
		}
		switch affect.Kind {
		case 0x27: // haste: AffectHaste doubles half-actions
			blows[0] *= 2
			blows[1] *= 2
		case 0x2A: // slow: AffectSlow halves half-actions
			blows[0] /= 2
			blows[1] /= 2
		}
	}
	return blows
}

// MonsterIsHeld mirrors the reference Player.IsHeld affect set.
//
// ★ 這一組不是憑語意挑的，是**同一支 handler**：效果分派表裡 `1Bh`、`1Fh`、
// `33h`、`34h`、`35h` 五個碼的處理常式都是 `overlay-12:0075h`（spec 1005 的
// 對照表），那一支呼叫 `CLEARACTION`。`1Bh`（笨拙術）先前漏了——
// 漏掉的症狀是那支法術掛得上去卻沒有任何效果。
func (f Fighter) MonsterIsHeld() bool {
	for _, affect := range f.MonsterAffects {
		if affect.operational() && (affect.Kind == 0x1B || affect.Kind == 0x1F ||
			affect.Kind == 0x33 || affect.Kind == 0x34 || affect.Kind == 0x35) {
			return true
		}
	}
	return false
}

const MonsterMagicMissileSpellID uint8 = 0x0F

// ActionState mirrors the per-player fields cleared by the reference
// CombatantKilled routine. It is deliberately data-only so ECL, UI and other
// Gold Box frontends can share the reset contract.
type ActionState = engineaction.State

type Turn struct {
	FighterID  string
	Initiative int
}

type AttackResult struct {
	AttackerID string
	TargetID   string
	AttackRoll int
	Total      int
	Hit        bool
	Critical   bool
	Damage     int
	TargetHP   int
	Effects    []AttackEffectResult
}

// AttackEffectResult preserves a post-hit effect separately from the weapon
// result. Kind is the original MON*SPC effect ID; DamageFlags use the decoded
// reference bitfield instead of a title-specific spell or monster name.
type AttackEffectResult struct {
	Kind         uint8
	DamageFlags  uint8
	RolledDamage int
	Damage       int
	TargetHP     int
	Protected    bool
}

type MoveResult struct {
	Fighter      Fighter
	Attack       *AttackResult
	FreeAttacks  []AttackResult
	GuardAttacks []AttackResult
	MovementCost int
}

// MovementTerrain exposes the original combat background table without
// coupling Battle to DUNGCOM/WILDCOM assets. ok=false marks an invalid or
// impassable cell; cost is the number of movement points consumed.
type MovementTerrain func(x, y int) (cost int, ok bool)

type SpellResult struct {
	CasterID string
	TargetID string
	SpellID  uint8
	Missiles int
	Damage   int
	Healing  int
	TargetHP int
	Targets  int
	Resisted bool
}

type AreaSpellImpact struct {
	TargetID  string
	Damage    int
	TargetHP  int
	Saved     bool
	Resisted  bool
	Protected bool
}

type AreaSpellResult struct {
	CasterID   string
	SpellID    uint8
	Center     TilePoint
	BaseDamage int
	Impacts    []AreaSpellImpact
}

const SleepSpellID uint8 = 0x15

// SleepImpact records one target that passed the ordered HD-capacity filter.
// Resisted targets consumed capacity but did not receive effect 35h.
type SleepImpact struct {
	TargetID string
	Resisted bool
	Duration uint16
}

// SleepResult deliberately receives an already ordered target list. Combat
// geometry and SCAN tie order remain a separate, fail-closed adapter boundary.
type SleepResult struct {
	CasterID          string
	SpellID           uint8
	InitialCapacity   int
	RemainingCapacity int
	Impacts           []SleepImpact
}

// SpellInterruption records a title-neutral original-engine boundary clearing
// a pending action spell. Memorized-slot ownership remains in the game adapter.
type SpellInterruption struct {
	FighterID string
	SpellID   uint8
}

type Battle struct {
	fighters           map[string]Fighter
	fighterOrder       []string
	attackRollModifier map[Side]int
	rngStream          *enginerandom.Stream
	rng                interface {
		Intn(int) int
	}
	round               int
	status              Status
	areas               []PersistentArea
	nextArea            uint64
	initiativeScheduler *engineinitiative.Scheduler
	initiativeSelection engineinitiative.Selection
	initiativeSelected  bool
	spellInterruptions  []SpellInterruption
	// terrainCode 是戰術地圖的地形碼查詢（`overlay-32 entry#19` 的第四個回填值）。
	// 為 nil 時沒有障礙格——與加這一層之前的行為相同。
	terrainCode CombatTerrainCode
	// levelDrainRules 是刻意偏離原作的等級吸取宣告（spec 1129）。
	levelDrainRules []LevelDrainRule
}

// CombatTerrainCode 回傳戰術地圖上一格的地形碼；`onMap` 為 false 代表出界。
type CombatTerrainCode func(x, y int) (code uint8, onMap bool)

// SetCombatTerrainCodes 掛上地形碼查詢。與 `MovementTerrain`（成本）分開兩支，
// 因為原作也分開回傳：成本查 `26A2h` 那張表，障礙走的是另一條路。
func (b *Battle) SetCombatTerrainCodes(lookup CombatTerrainCode) {
	if b != nil {
		b.terrainCode = lookup
	}
}

func (b *Battle) applyPositiveDamage(target *Fighter, damage int) int {
	if b == nil || target == nil || damage <= 0 {
		return 0
	}
	if target.HitPoints > 0 && damage >= target.HitPoints {
		// 這一擊放倒目標：記下溢出量給 SAVEDAMAGE 階梯（spec 1205）。
		// 追擊（HP 已是 0）不覆寫——原作倒下者已離開戰鬥不再被選中。
		target.DownOverkill = damage - target.HitPoints
	}
	if damage > target.HitPoints {
		damage = target.HitPoints
	}
	if damage <= 0 {
		return 0
	}
	target.HitPoints -= damage
	b.removeDamageCancelledAffects(target)
	b.interruptPendingSpell(target)
	return damage
}

// removeDamageCancelledAffects projects the verified PC-98 PUTDAMAGE →
// REMOVEFX boundary. Sleep (35h) is one of the dynamic effects removed after
// strictly positive damage; innate MON*SPC capabilities are not spell records
// created by PUTEFFECT and must remain attached.
func (b *Battle) removeDamageCancelledAffects(target *Fighter) {
	if target == nil || len(target.MonsterAffects) == 0 {
		return
	}
	kept := target.MonsterAffects[:0]
	removedSleep := false
	for _, affect := range target.MonsterAffects {
		if affect.Kind == 0x35 && !affect.Innate {
			removedSleep = true
			continue
		}
		kept = append(kept, affect)
	}
	target.MonsterAffects = kept
	if removedSleep {
		b.clearActionState(target)
	}
}

func (b *Battle) interruptPendingSpell(target *Fighter) {
	if b == nil || target == nil {
		return
	}
	if spellID := target.CombatAction.InterruptSpell(); spellID != 0 {
		b.spellInterruptions = append(b.spellInterruptions, SpellInterruption{
			FighterID: target.ID, SpellID: spellID,
		})
	}
}

// TakeSpellInterruptions transfers all pending interruption events in original
// execution order and clears the Battle queue.
func (b *Battle) TakeSpellInterruptions() []SpellInterruption {
	if b == nil || len(b.spellInterruptions) == 0 {
		return nil
	}
	result := append([]SpellInterruption(nil), b.spellInterruptions...)
	b.spellInterruptions = nil
	return result
}

const FireballSpellID uint8 = 0x2F

func NewBattle(fighters []Fighter, seed int64) (*Battle, error) {
	if len(fighters) == 0 {
		return nil, fmt.Errorf("battle needs at least one fighter")
	}
	rngStream := enginerandom.New(seed)
	b := &Battle{
		fighters:           make(map[string]Fighter, len(fighters)),
		fighterOrder:       make([]string, 0, len(fighters)),
		attackRollModifier: make(map[Side]int, 2),
		rngStream:          rngStream,
		rng:                rngStream.Rand(),
		status:             StatusActive,
	}
	for _, fighter := range fighters {
		if fighter.ID == "" {
			return nil, fmt.Errorf("fighter has empty ID")
		}
		if _, exists := b.fighters[fighter.ID]; exists {
			return nil, fmt.Errorf("duplicate fighter ID %q", fighter.ID)
		}
		if fighter.LegacyObjectID != 0 {
			for _, existingID := range b.fighterOrder {
				if b.fighters[existingID].LegacyObjectID == fighter.LegacyObjectID {
					return nil, fmt.Errorf("duplicate legacy combat object ID %d", fighter.LegacyObjectID)
				}
			}
		}
		if fighter.MaxHitPoints <= 0 || fighter.HitPoints < 0 || fighter.HitPoints > fighter.MaxHitPoints {
			return nil, fmt.Errorf("fighter %q has invalid hit points", fighter.ID)
		}
		if fighter.ArmorClass < -20 || fighter.ArmorClass > 30 {
			return nil, fmt.Errorf("fighter %q has invalid armor class", fighter.ID)
		}
		if fighter.DamageDiceCount < 0 || fighter.DamageDiceSides < 0 {
			return nil, fmt.Errorf("fighter %q has invalid damage dice", fighter.ID)
		}
		if fighter.HitPoints == 0 {
			// A battle imported from a save/encounter may already contain a
			// downed combatant. Apply the same CombatantKilled boundary at
			// construction time so it cannot occupy a tile or turn order.
			fighter.HasCombatPosition = false
			fighter.DeathOverlay = true
			fighter.DownedCorpse = fighter.Side == SideParty
			fighter.CombatAction = ActionState{}
		}
		b.fighters[fighter.ID] = fighter
		b.fighterOrder = append(b.fighterOrder, fighter.ID)
	}
	return b, nil
}

// SetDamageRules attaches immutable game-pack capabilities to every fighter
// in the battle. Rules are configuration, not mutable combat state, so save
// restore calls this again after rebuilding the Battle.
func (b *Battle) SetDamageRules(rules []enginedamage.Rule) {
	if b == nil {
		return
	}
	for id, fighter := range b.fighters {
		fighter.DamageRules = append([]enginedamage.Rule(nil), rules...)
		b.fighters[id] = fighter
	}
}

// SetConditionalModifierRules attaches immutable game-pack interaction rules
// to every fighter. The rules are configuration, not mutable combat state.
func (b *Battle) SetConditionalModifierRules(rules []enginemodifier.Rule) {
	if b == nil {
		return
	}
	for id, fighter := range b.fighters {
		fighter.ConditionalModifierRules = append([]enginemodifier.Rule(nil), rules...)
		for index := range fighter.ConditionalModifierRules {
			fighter.ConditionalModifierRules[index].Values = append([]uint8(nil), rules[index].Values...)
		}
		b.fighters[id] = fighter
	}
}

// SetMagicResistanceRules attaches immutable game-pack resistance rules to
// every fighter. Rules are configuration, not mutable combat state.
func (b *Battle) SetMagicResistanceRules(rules []engineresistance.Rule) {
	if b == nil {
		return
	}
	for id, fighter := range b.fighters {
		fighter.MagicResistanceRules = append([]engineresistance.Rule(nil), rules...)
		b.fighters[id] = fighter
	}
}

// SetPostHitRules attaches immutable game-pack post-hit capabilities to every
// fighter. Rules are configuration, not mutable combat state.
func (b *Battle) SetPostHitRules(rules []engineposthit.Rule) {
	if b == nil {
		return
	}
	for id, fighter := range b.fighters {
		fighter.PostHitRules = append([]engineposthit.Rule(nil), rules...)
		b.fighters[id] = fighter
	}
}

// SetMonsterSpellRules attaches immutable game-pack innate spell capabilities
// to every fighter. Rules are configuration, not mutable combat state.
func (b *Battle) SetMonsterSpellRules(rules []enginespell.Rule) {
	if b == nil {
		return
	}
	for id, fighter := range b.fighters {
		fighter.MonsterSpellRules = append([]enginespell.Rule(nil), rules...)
		b.fighters[id] = fighter
	}
}

func (b *Battle) Round() int { return b.round }

func (b *Battle) Status() Status { return b.status }

func (b *Battle) DynamicInitiativeActive() bool { return b != nil && b.initiativeScheduler != nil }

func (b *Battle) Fighters() []Fighter {
	output := make([]Fighter, 0, len(b.fighters))
	for _, fighter := range b.fighters {
		output = append(output, fighter)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].ID < output[j].ID })
	return output
}

// FightersInCombatOrder returns a copy in the preserved
// CHARACTERLIST/OBJECTLIST order. It is distinct from Fighters, whose
// lexicographic ordering is useful for deterministic map-facing APIs but is
// not evidence for legacy Quick target traversal.
func (b *Battle) FightersInCombatOrder() []Fighter {
	if b == nil {
		return nil
	}
	output := make([]Fighter, 0, len(b.fighterOrder))
	for _, id := range b.fighterOrder {
		if fighter, ok := b.fighters[id]; ok {
			output = append(output, fighter)
		}
	}
	return output
}

func (b *Battle) Fighter(id string) (Fighter, bool) {
	fighter, ok := b.fighters[id]
	return fighter, ok
}

// SetSideAttackRollModifier installs a battle-scoped modifier without
// mutating persistent fighter statistics. Legacy script work variables use
// this for encounter-wide blessings, curses, terrain, and faction effects.
func (b *Battle) SetSideAttackRollModifier(side Side, modifier int) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	if side != SideParty && side != SideEnemy {
		return fmt.Errorf("unsupported combat side %d", side)
	}
	b.attackRollModifier[side] = modifier
	return nil
}

func (b *Battle) SideAttackRollModifier(side Side) int {
	if b == nil {
		return 0
	}
	return b.attackRollModifier[side]
}

// SetHitPoints applies an external damage/healing adapter to a combatant and
// immediately recomputes the reference party/enemy win state. ECL DAMAGE can
// use this bridge without reaching into the Battle fighter map.
func (b *Battle) SetHitPoints(fighterID string, hitPoints int) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if hitPoints < 0 || hitPoints > fighter.MaxHitPoints {
		return fmt.Errorf("fighter %q hit points %d outside 0..%d", fighterID, hitPoints, fighter.MaxHitPoints)
	}
	fighter.HitPoints = hitPoints
	if hitPoints == 0 {
		// Reference CombatMap size becomes zero when in_combat is cleared;
		// HasCombatPosition is the renderer-neutral equivalent.
		fighter.HasCombatPosition = false
		fighter.DeathOverlay = true
		fighter.DownedCorpse = fighter.Side == SideParty
		fighter.CombatAction = ActionState{}
	} else {
		// Healing clears the one-shot downed visual. Position restoration is a
		// separate placement operation, matching the reference map contract.
		fighter.DeathOverlay = false
	}
	b.fighters[fighterID] = fighter
	b.updateStatus()
	return nil
}

// RestoreCombatant explicitly performs the reference combat_heal placement
// boundary. HP healing alone intentionally does not call this method: a
// healed DownedCorpse remains off the combat map until a caller supplies the
// stand-up/placement intent.
func (b *Battle) RestoreCombatant(fighterID string, position TilePoint) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if fighter.HitPoints <= 0 {
		return fmt.Errorf("fighter %q cannot stand with %d hit points", fighterID, fighter.HitPoints)
	}
	for _, other := range b.fighters {
		if other.ID == fighterID || other.HitPoints <= 0 || !other.HasCombatPosition {
			continue
		}
		if FootprintsOverlapAt(fighter, position.X, position.Y, other) {
			return fmt.Errorf("fighter %q placement (%d,%d) is occupied by %q", fighterID, position.X, position.Y, other.ID)
		}
	}
	fighter.HasCombatPosition = true
	fighter.CombatX, fighter.CombatY = position.X, position.Y
	fighter.DeathOverlay = false
	fighter.DownedCorpse = false
	b.fighters[fighterID] = fighter
	return nil
}

// ValidateAttack checks non-random attack preconditions. Game adapters can
// call it before committing resources such as ammunition, while Attack calls
// it before consuming the deterministic RNG stream.
func (b *Battle) ValidateAttack(attackerID, targetID string) error {
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return fmt.Errorf("unknown attacker %q", attackerID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return fmt.Errorf("battle is already over")
	}
	if attacker.HitPoints <= 0 || target.HitPoints <= 0 {
		return fmt.Errorf("dead fighter cannot attack")
	}
	if attacker.HasCombatPosition && target.HasCombatPosition && attacker.MissileWeapon && adjacent(attacker, target) && !attacker.ThrownWeapon {
		return ErrAdjacentMissileTarget
	}
	return nil
}

// StartRound writes the verified PC-98 Action.delay values, then repeatedly
// scans stable TeamList order with one d100 per living entry. The current UI
// exposes one action per fighter; the original DELAY command's same-round
// reinsertion remains a separate dynamic scheduler boundary.
func (b *Battle) StartRound() ([]Turn, error) {
	if b.status != StatusActive {
		return nil, fmt.Errorf("battle is already over")
	}
	initialized := b.initializeRoundDelays()
	b.initiativeScheduler = nil
	b.initiativeSelected = false
	_, selections := engineinitiative.OrderInitialized(b.rng, initialized)
	turns := make([]Turn, 0, len(selections))
	for _, selection := range selections {
		fighter := b.fighters[selection.ID]
		fighter.CombatAction.Delay = selection.ActionDelay
		b.fighters[selection.ID] = fighter
		turns = append(turns, Turn{FighterID: selection.ID, Initiative: selection.ActionDelay})
	}
	return turns, nil
}

func (b *Battle) initializeRoundDelays() []engineinitiative.Entry {
	b.round++
	// CLOCK_ passes one ROUNDS unit into the shared EFFECTREC reducer at the
	// combat-round boundary. Effects created during the previous round lose one
	// duration tick before this round's held/action projection is evaluated.
	b.AdvanceMonsterAffects(1)
	b.expirePersistentAreas()
	b.advanceBlessDurations()
	b.advanceCurseDurations()
	b.advanceProtectionDurations()
	entries := make([]engineinitiative.Entry, 0, len(b.fighterOrder))
	for _, id := range b.fighterOrder {
		fighter := b.fighters[id]
		if fighter.HitPoints > 0 {
			combatTeam := fighter.CombatTeam
			if fighter.Side == SideEnemy && combatTeam == 0 {
				combatTeam = 1
			}
			entries = append(entries, engineinitiative.Entry{
				ID: fighter.ID, Dexterity: fighter.Dexterity, CombatTeam: combatTeam,
			})
			// 回合開始清掉動作計數與累計轉向（原作 `overlay-08 entry#4`，spec 804）。
			// 面向不清——它跨回合保留。
			fighter.CombatActionCount, fighter.CombatTurnTotal = 0, 0
			b.fighters[id] = fighter
		}
	}
	initialized := engineinitiative.InitializeDelays(b.rng, entries, 0)
	for index := range initialized {
		fighter := b.fighters[initialized[index].ID]
		if fighter.InitiativeBonus == 0 {
			continue
		}
		delay := fighter.InitiativeBonus
		if delay < 1 {
			delay = 1
		}
		if delay > 20 {
			delay = 20
		}
		initialized[index].ActionDelay = delay
	}
	for _, entry := range initialized {
		fighter := b.fighters[entry.ID]
		fighter.CombatAction.Delay = entry.ActionDelay
		b.fighters[entry.ID] = fighter
	}
	return initialized
}

// BeginScheduledRound starts the dynamic original-style scheduler. Unlike
// StartRound it does not pre-roll future d100 scans, so DELAY can change the
// current round without consuming a different random continuation.
func (b *Battle) BeginScheduledRound() error {
	if b.status != StatusActive {
		return fmt.Errorf("battle is already over")
	}
	initialized := b.initializeRoundDelays()
	b.initiativeScheduler = engineinitiative.NewScheduler(initialized)
	b.initiativeSelected = false
	return nil
}

// NextScheduledTurn performs exactly one full TeamList selection scan.
func (b *Battle) NextScheduledTurn() (Turn, bool, error) {
	if b.initiativeScheduler == nil {
		return Turn{}, false, fmt.Errorf("dynamic initiative round is not initialized")
	}
	if b.initiativeSelected {
		return Turn{}, false, fmt.Errorf("scheduled action has not been completed or delayed")
	}
	selection, ok := b.initiativeScheduler.SelectNext(b.rng)
	if !ok {
		return Turn{}, false, nil
	}
	b.initiativeSelection = selection
	b.initiativeSelected = true
	fighter := b.fighters[selection.ID]
	// The reference turn entry clears guarding before processing restraint,
	// quick-fight or the player menu. A guard therefore survives until this
	// fighter's next selected turn, including enemy movement earlier in a new
	// round.
	fighter.CombatAction.Guarding = false
	fighter.CombatAction.Delay = selection.ActionDelay
	b.fighters[selection.ID] = fighter
	return Turn{FighterID: selection.ID, Initiative: selection.ActionDelay}, true, nil
}

// ClearAction mirrors the reference clear_actions boundary and consumes the
// currently selected scheduler action when this fighter owns it.
func (b *Battle) ClearAction(fighterID string) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if !b.clearActionState(&fighter) {
		return fmt.Errorf("stale scheduled action for fighter %q", fighterID)
	}
	b.fighters[fighterID] = fighter
	return nil
}

func (b *Battle) clearActionState(fighter *Fighter) bool {
	if b == nil || fighter == nil {
		return false
	}
	fighter.CombatAction.Clear()
	if b.initiativeScheduler != nil && b.initiativeSelected && b.initiativeSelection.ID == fighter.ID {
		if !b.initiativeScheduler.Complete(b.initiativeSelection) {
			return false
		}
		b.initiativeSelected = false
	}
	return true
}

// CanGuard reports the current typed weapon-mode boundary. Pure missile
// weapons cannot guard; thrown weapons retain their verified melee use.
func (b *Battle) CanGuard(fighterID string) bool {
	fighter, ok := b.fighters[fighterID]
	return ok && fighter.HitPoints > 0 && (!fighter.MissileWeapon || fighter.ThrownWeapon)
}

// GuardAction clears the selected action and arms one adjacent-entry attack.
func (b *Battle) GuardAction(fighterID string) error {
	if !b.CanGuard(fighterID) {
		return fmt.Errorf("fighter %q cannot guard with the readied weapon", fighterID)
	}
	if err := b.ClearAction(fighterID); err != nil {
		return err
	}
	fighter := b.fighters[fighterID]
	fighter.CombatAction.Guard()
	b.fighters[fighterID] = fighter
	return nil
}

// SetActionTarget stores a title-resolved target pointer for the current
// action. Team semantics are checked by the caller that owns the legacy rule.
func (b *Battle) SetActionTarget(fighterID, targetID string) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if fighter.HitPoints <= 0 {
		return fmt.Errorf("dead fighter %q cannot receive an action target", fighterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return fmt.Errorf("unknown action target %q", targetID)
	}
	if target.HitPoints <= 0 {
		return fmt.Errorf("dead action target %q is not selectable", targetID)
	}
	if !fighter.CombatAction.SetActionTarget(targetID) {
		return fmt.Errorf("empty action target for fighter %q", fighterID)
	}
	b.fighters[fighterID] = fighter
	return nil
}

// SetQuickFight gives one combatant to the AI using the PC-98 default policy.
func (b *Battle) SetQuickFight(fighterID string) error {
	return b.SetQuickFightWithPolicy(fighterID, true)
}

// SetQuickFightWithPolicy gives one combatant to the AI. When enabled, a
// target belonging to the same combat team is cleared, matching the verified
// ALT+Q setter while leaving opposing-team targets intact.
func (b *Battle) SetQuickFightWithPolicy(fighterID string, clearSameTeamTarget bool) error {
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if clearSameTeamTarget && fighter.CombatAction.ActionTargetID != "" {
		if target, exists := b.fighters[fighter.CombatAction.ActionTargetID]; exists && target.Side == fighter.Side {
			fighter.CombatAction.ClearActionTarget()
		}
	}
	fighter.QuickFight = true
	b.fighters[fighterID] = fighter
	return nil
}

// SetAllQuickFight mirrors the original ALT+Q transaction: the currently
// selected action is marked with delay 20, then every TeamList combatant is
// delegated through the same per-fighter Quick setter.  The 20 marker is a
// handoff state, not a new initiative tier.
func (b *Battle) SetAllQuickFight(currentID string) error {
	return b.SetAllQuickFightWithPolicy(currentID, true)
}

// SetAllQuickFightWithPolicy applies the same title-owned target policy to
// every TeamList combatant during the ALT+Q handoff.
func (b *Battle) SetAllQuickFightWithPolicy(currentID string, clearSameTeamTarget bool) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	current, ok := b.fighters[currentID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", currentID)
	}
	current.CombatAction.Delay = 20
	b.fighters[currentID] = current
	for _, fighterID := range b.fighterOrder {
		if err := b.SetQuickFightWithPolicy(fighterID, clearSameTeamTarget); err != nil {
			return err
		}
	}
	return nil
}

// BeginQuickFightAction consumes the original ALT+Q handoff marker before AI
// processes the selected combatant. Ordinary Quick actions leave their delay
// unchanged.
func (b *Battle) BeginQuickFightAction(fighterID string) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if fighter.QuickFight && fighter.CombatAction.Delay == 20 {
		fighter.CombatAction.Delay = 19
		b.fighters[fighterID] = fighter
	}
	return nil
}

// SelectQuickSpell keeps the original selector on the Battle PRNG stream.
// Spell records and title-specific suitability are supplied by the adapter.
func (b *Battle) SelectQuickSpell(
	spellIDs []uint8,
	lookup enginequickspell.Lookup,
	suitable enginequickspell.Suitable,
) (uint8, bool, error) {
	if b == nil || b.rng == nil {
		return 0, false, fmt.Errorf("battle PRNG is unavailable")
	}
	return enginequickspell.Select(spellIDs, func(sides int) (int, error) {
		if sides < 1 {
			return 0, fmt.Errorf("quick spell die has %d sides", sides)
		}
		return b.rng.Intn(sides) + 1, nil
	}, lookup, suitable)
}

// SelectQuickTarget keeps the recovered target retry roll on the same Battle
// PRNG stream as Quick spell selection. The title adapter supplies legality;
// this layer does not interpret target records or spell effects.
func (b *Battle) SelectQuickTarget(
	candidates []enginequicktarget.Candidate,
	rule enginequicktarget.Rule,
	suitable enginequicktarget.Suitable,
) (enginequicktarget.Candidate, bool, error) {
	if b == nil || b.rng == nil {
		return enginequicktarget.Candidate{}, false, fmt.Errorf("battle PRNG is unavailable")
	}
	return enginequicktarget.Select(candidates, rule, func(sides int) (int, error) {
		if sides < 1 {
			return 0, fmt.Errorf("quick target die has %d sides", sides)
		}
		return b.rng.Intn(sides) + 1, nil
	}, suitable)
}

// SelectQuickTargetOne performs the single-draw target boundary used by
// Quick consumers such as Magic Missile after the title adapter has projected
// its candidate list. It keeps the draw on the Battle PRNG and lets the
// engine rule preserve the recovered legacy object order.
func (b *Battle) SelectQuickTargetOne(
	candidates []enginequicktarget.Candidate,
	rule enginequicktarget.Rule,
) (enginequicktarget.Candidate, bool, error) {
	if b == nil || b.rng == nil {
		return enginequicktarget.Candidate{}, false, fmt.Errorf("battle PRNG is unavailable")
	}
	return enginequicktarget.SelectOne(candidates, rule, func(sides int) (int, error) {
		if sides < 1 {
			return 0, fmt.Errorf("quick target die has %d sides", sides)
		}
		return b.rng.Intn(sides) + 1, nil
	})
}

// BeginPendingSpellAction mirrors CASTCOMBATSPELL's nonzero casting-delay
// handoff. The same action remains in this round at max(1, delay-units).
func (b *Battle) BeginPendingSpellAction(fighterID string, spellID uint8, castingDelay int) error {
	return b.BeginPendingTargetedSpellAction(fighterID, spellID, castingDelay, "")
}

// BeginPendingTargetedSpellAction preserves an adapter-selected target across
// the same delayed scheduler handoff as an untargeted spell.
func (b *Battle) BeginPendingTargetedSpellAction(fighterID string, spellID uint8, castingDelay int, targetID string) error {
	return b.beginPendingSpellAction(fighterID, spellID, castingDelay, func(state *engineaction.State) bool {
		return state.BeginTargetedSpell(spellID, castingDelay, targetID)
	})
}

// BeginPendingPointSpellAction preserves an adapter-selected grid coordinate.
func (b *Battle) BeginPendingPointSpellAction(fighterID string, spellID uint8, castingDelay, x, y int) error {
	return b.beginPendingSpellAction(fighterID, spellID, castingDelay, func(state *engineaction.State) bool {
		return state.BeginPointSpell(spellID, castingDelay, x, y)
	})
}

func (b *Battle) beginPendingSpellAction(fighterID string, spellID uint8, castingDelay int, begin func(*engineaction.State) bool) error {
	if b == nil || b.initiativeScheduler == nil || !b.initiativeSelected ||
		b.initiativeSelection.ID != fighterID {
		return fmt.Errorf("fighter %q is not the selected scheduled action", fighterID)
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if !begin(&fighter.CombatAction) {
		return fmt.Errorf("spell 0x%02X has invalid casting delay %d", spellID, castingDelay)
	}
	if !b.initiativeScheduler.SetDelay(fighterID, fighter.CombatAction.Delay) {
		return fmt.Errorf("unknown scheduled fighter %q", fighterID)
	}
	b.fighters[fighterID] = fighter
	b.initiativeSelected = false
	return nil
}

// TakePendingPointSpellAction atomically consumes a spell and grid coordinate.
func (b *Battle) TakePendingPointSpellAction(fighterID string) (uint8, int, int, bool, error) {
	if b == nil || !b.initiativeSelected || b.initiativeSelection.ID != fighterID {
		return 0, 0, 0, false, fmt.Errorf("fighter %q is not the selected scheduled action", fighterID)
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return 0, 0, 0, false, fmt.Errorf("unknown fighter %q", fighterID)
	}
	spellID, x, y, hasPoint := fighter.CombatAction.TakePointSpell()
	if spellID == 0 {
		return 0, 0, 0, false, fmt.Errorf("fighter %q has no pending spell", fighterID)
	}
	b.fighters[fighterID] = fighter
	return spellID, x, y, hasPoint, nil
}

// TakePendingSpellAction clears the spell byte when its delayed action is
// selected again. Delay completion remains owned by CompleteAction.
func (b *Battle) TakePendingSpellAction(fighterID string) (uint8, error) {
	spellID, _, err := b.TakePendingTargetedSpellAction(fighterID)
	return spellID, err
}

// TakePendingTargetedSpellAction atomically consumes the spell and target.
func (b *Battle) TakePendingTargetedSpellAction(fighterID string) (uint8, string, error) {
	if b == nil || !b.initiativeSelected || b.initiativeSelection.ID != fighterID {
		return 0, "", fmt.Errorf("fighter %q is not the selected scheduled action", fighterID)
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return 0, "", fmt.Errorf("unknown fighter %q", fighterID)
	}
	spellID, targetID := fighter.CombatAction.TakeTargetedSpell()
	if spellID == 0 {
		return 0, "", fmt.Errorf("fighter %q has no pending spell", fighterID)
	}
	b.fighters[fighterID] = fighter
	return spellID, targetID, nil
}

// SetPlayerCharactersManual clears quick-fight only for the original PC
// control namespace. NPC and temporary monster allies remain automated.
func (b *Battle) SetPlayerCharactersManual() int {
	changed := 0
	for _, id := range b.fighterOrder {
		fighter := b.fighters[id]
		if fighter.Side != SideParty || fighter.ControlMorale >= 0x80 || !fighter.QuickFight {
			continue
		}
		fighter.QuickFight = false
		b.fighters[id] = fighter
		changed++
	}
	return changed
}

// DelayAction implements the verified combat submenu operation: write delay
// one, retain the action in this round, and release the current selection.
func (b *Battle) DelayAction(fighterID string) error {
	if !b.initiativeSelected || b.initiativeSelection.ID != fighterID {
		return fmt.Errorf("fighter %q is not the selected scheduled action", fighterID)
	}
	if !b.initiativeScheduler.SetDelay(fighterID, 1) {
		return fmt.Errorf("unknown scheduled fighter %q", fighterID)
	}
	fighter := b.fighters[fighterID]
	fighter.CombatAction.Delay = 1
	b.fighters[fighterID] = fighter
	b.initiativeSelected = false
	return nil
}

// CompleteAction clears the scheduler delay after a fighter's current turn.
// Other renderer-neutral action fields remain intact until their owning UI
// or rule boundary clears them.
func (b *Battle) CompleteAction(fighterID string) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	fighter.CombatAction.Delay = 0
	b.fighters[fighterID] = fighter
	if b.initiativeScheduler != nil && b.initiativeSelected && b.initiativeSelection.ID == fighterID {
		if !b.initiativeScheduler.Complete(b.initiativeSelection) {
			return fmt.Errorf("stale scheduled action for fighter %q", fighterID)
		}
		b.initiativeSelected = false
	}
	return nil
}

// AdvanceMonsterAffects applies the reference finite-duration timeout rule to
// raw combat effects. The input is EFFECTREC duration ticks (combat rounds),
// not an arbitrary wall-clock duration. Strength 255 is the permanent marker.
func (b *Battle) AdvanceMonsterAffects(ticks uint16) int {
	if ticks == 0 {
		return 0
	}
	removed := 0
	for id, fighter := range b.fighters {
		if len(fighter.MonsterAffects) == 0 {
			continue
		}
		kept := make([]MonsterAffect, 0, len(fighter.MonsterAffects))
		removedHeld := false
		for _, affect := range fighter.MonsterAffects {
			if affect.Strength != 0xFF {
				duration := affect.Duration
				if duration == 0 {
					duration = affect.Value
				}
				// CLOCK_ local 0020h treats zero as a non-expiring record and
				// follows the next linked-list entry without calling SPELLOFF.
				if duration == 0 {
					kept = append(kept, affect)
					continue
				}
				if duration <= ticks {
					removed++
					removedHeld = removedHeld || affect.Kind == 0x35
					continue
				}
				affect.Duration = duration - ticks
				affect.Value = affect.Duration
			}
			kept = append(kept, affect)
		}
		fighter.MonsterAffects = kept
		if removedHeld {
			b.clearActionState(&fighter)
		}
		b.fighters[id] = fighter
	}
	return removed
}

// SelectCombatTarget reproduces the bounded target-selection part of the
// reference monster turn. The full engine builds a reachable/visible target
// list; until those map rules are decoded, this adapter uses all living
// fighters on the requested opposing side. Sorting before consuming the
// seeded RNG keeps replays independent of Go map iteration order.
func (b *Battle) SelectCombatTarget(attackerID string, targetSide Side) (Fighter, error) {
	if b == nil {
		return Fighter{}, fmt.Errorf("battle is nil")
	}
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return Fighter{}, fmt.Errorf("unknown attacker %q", attackerID)
	}
	if b.status != StatusActive {
		return Fighter{}, fmt.Errorf("battle is already over")
	}
	if attacker.HitPoints <= 0 {
		return Fighter{}, fmt.Errorf("dead fighter cannot select a target")
	}
	candidates := make([]Fighter, 0, len(b.fighters))
	for _, fighter := range b.fighters {
		if fighter.Side == targetSide && fighter.HitPoints > 0 {
			candidates = append(candidates, fighter)
		}
	}
	if len(candidates) == 0 {
		return Fighter{}, fmt.Errorf("no living target on side %d", targetSide)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return legacyTargetTieLess(candidates[i], candidates[j])
	})
	return candidates[b.rng.Intn(len(candidates))], nil
}

// ResolveAttack applies the recovered attack rule with injected dice. A
// natural 1 misses, a natural 20 always hits, otherwise d20+AttackBonus plus
// the battle-scoped side modifier must meet the target AC. damageRoll is the
// already rolled weapon-dice total.
func (b *Battle) ResolveAttack(attackerID, targetID string, attackRoll, damageRoll int) (AttackResult, error) {
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return AttackResult{}, fmt.Errorf("unknown attacker %q", attackerID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return AttackResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return AttackResult{}, fmt.Errorf("battle is already over")
	}
	if attacker.HitPoints <= 0 || target.HitPoints <= 0 {
		return AttackResult{}, fmt.Errorf("dead fighter cannot attack")
	}
	if attackRoll < 1 || attackRoll > 20 {
		return AttackResult{}, fmt.Errorf("attack roll %d is outside d20", attackRoll)
	}
	if damageRoll < 0 {
		return AttackResult{}, fmt.Errorf("negative damage roll")
	}
	if err := b.ValidateAttack(attackerID, targetID); err != nil {
		return AttackResult{}, err
	}
	forcedMiss := target.MonsterAffectForcesAttackMiss()
	critical := attackRoll == 20 && !forcedMiss
	// 原作攻擊動作在結算之前先轉向面對目標（`overlay-13:19D8h`，spec 1019）；
	// 動作計數已經到 2 就不再轉。轉完要重讀，後面的判定吃的是新面向。
	if _, faceErr := b.FaceTarget(attacker.ID, target.ID); faceErr == nil {
		attacker = b.fighters[attacker.ID]
	}
	targetArmorClass := target.ArmorClass
	// 原作攻擊結算（`overlay-13:14E8h`）在三道條件成立時改用第二個 AC 欄位。
	if target.ArmorClassFacingKnown {
		if rear, rearErr := b.RearAttackApplies(attacker.ID, target.ID); rearErr == nil && rear {
			targetArmorClass = target.ArmorClassFacing
		}
	}
	// AC 是畫面刻度：**數字小才難打**，所以防禦加成一律往下扣。
	targetArmorClass -= target.MonsterAffectArmorClassBonusAgainst(attacker)
	if attacker.isEvil() && target.ProtectedFromEvil {
		targetArmorClass -= 2
	}
	if attacker.isGood() && target.ProtectedFromGood {
		targetArmorClass -= 2
	}
	conditional := target.MonsterConditionalModifierAgainst(attacker)
	attackTotal := attackRoll + attacker.AttackBonus + b.attackRollModifier[attacker.Side] + conditional.AttackRollDelta
	hit := !forcedMiss && (target.MonsterIsHeld() || critical ||
		(attackRoll != 1 && attackTotal+targetArmorClass >= armorClassHitTarget))
	damage := 0
	if hit {
		// 加值也要跟著大／小體型換（原作是「減掉小型的、加上大型的」，
		// 等價於直接用那一組的加值，spec 1175）。
		_, _, damageBonus := attacker.WeaponDamageAgainst(target)
		damage = damageRoll + damageBonus
		if damage < 0 {
			damage = 0
		}
		// 近戰傷害算完之後，原作對**攻擊者**問 `CHECKFX(04h)`、對**目標**問
		// `05h`（呼叫點在 `overlay-13:01F0h`，spec 1123）。衰弱射線就是掛在
		// 攻擊者那一次：傷害減 25%。
		//
		// ⚠ 兩次查詢共用同一個傷害暫存，順序是先攻擊者再目標——反過來會讓
		// 「先減 25% 再折半」變成「先折半再減 25%」，結果不同。
		adjusted, fxErr := CheckFX(attacker, CheckFXMeleeAttacker,
			map[string]int{scratchDamage: damage})
		if fxErr != nil {
			return AttackResult{}, fxErr
		}
		adjusted, fxErr = CheckFX(target, CheckFXMeleeTarget,
			map[string]int{scratchDamage: adjusted.Applied[scratchDamage]})
		if fxErr != nil {
			return AttackResult{}, fxErr
		}
		damage = adjusted.Applied[scratchDamage]
		if damage < 0 {
			damage = 0
		}
		damage = b.applyPositiveDamage(&target, damage)
		b.fighters[targetID] = target
		b.updateStatus()
	}
	return AttackResult{AttackerID: attackerID, TargetID: targetID, AttackRoll: attackRoll, Total: attackTotal, Hit: hit, Critical: critical, Damage: damage, TargetHP: target.HitPoints}, nil
}

// Attack rolls a normal attack using the battle's deterministic RNG. Keeping
// the dice source inside Battle makes the game adapter reproducible by seed,
// while ResolveAttack remains available for exact rule regression tests.
func (b *Battle) Attack(attackerID, targetID string) (AttackResult, error) {
	return b.attackSlot(attackerID, targetID, 1)
}

// WeaponDamageAgainst 回傳這一次攻擊要用的傷害三連。
//
// ★ 原作在攻擊結算當下才換（`overlay-13:15EFh`，spec 1175）：攻擊者**有槽 0 的
// 武器**而且目標算大型時，改用類別表的大型三連，加值則是「減掉小型的、加上
// 大型的」。沒有武器時整段跳過——天生攻擊不分大小。
func (f Fighter) WeaponDamageAgainst(target Fighter) (count, sides, bonus int) {
	if !f.HasSlotZeroWeapon || !target.LargeTarget {
		return f.DamageDiceCount, f.DamageDiceSides, f.DamageBonus
	}
	return f.LargeDamageDiceCount, f.LargeDamageDiceSides, f.LargeDamageBonus
}

func (b *Battle) attackSlot(attackerID, targetID string, attackSlot int) (AttackResult, error) {
	if err := b.ValidateAttack(attackerID, targetID); err != nil {
		return AttackResult{}, err
	}
	attacker := b.fighters[attackerID]
	diceCount, diceSides, _ := attacker.WeaponDamageAgainst(b.fighters[targetID])
	var result AttackResult
	var err error
	if diceCount < 1 || diceSides < 1 {
		result, err = b.ResolveAttack(attackerID, targetID, b.rng.Intn(20)+1, 0)
	} else {
		attackRoll := b.rng.Intn(20) + 1
		damageRoll := 0
		for i := 0; i < diceCount; i++ {
			damageRoll += b.rng.Intn(diceSides) + 1
		}
		result, err = b.ResolveAttack(attackerID, targetID, attackRoll, damageRoll)
	}
	if err != nil || !result.Hit || result.TargetHP <= 0 {
		return result, err
	}
	for _, rule := range attacker.MonsterPostHitRules(attackSlot) {
		target := b.fighters[targetID]
		if target.HitPoints <= 0 {
			break
		}
		rolledDamage := 0
		for index := 0; index < rule.DamageDiceCount; index++ {
			rolledDamage += b.rng.Intn(rule.DamageDiceSides) + 1
		}
		adjustment := target.MonsterDamageAdjustment(rule.DamageMask, rolledDamage)
		effect := AttackEffectResult{
			Kind: rule.EffectKind, DamageFlags: rule.DamageMask, RolledDamage: rolledDamage, TargetHP: target.HitPoints,
			Protected: adjustment.Immune,
		}
		if !effect.Protected {
			damage := adjustment.Damage
			effect.Damage = b.applyPositiveDamage(&target, damage)
			effect.TargetHP = target.HitPoints
			b.fighters[targetID] = target
			b.updateStatus()
		}
		result.Effects = append(result.Effects, effect)
		result.TargetHP = effect.TargetHP
	}
	return result, nil
}

// AdjustBlows 是原作的 `ADJUSTBLOWS`（DOS `overlay-13:0F12h`、PC-98 同名函式，
// 兩平台位元組完全相同）：把**半次**單位的基準值換成這一回合的整數次數。
//
//	次數 := (半次值 + (ROUND and 1)) div 2
//
// ★ 加的是**回合數的最低位**，所以奇數的半次值會在回合之間交替：3（一次半）
// 在第 1、3、5 回合給 2 次，第 0、2、4 回合給 1 次。這正是 AD&D 的「每兩回合
// 三次攻擊」——寫成 `半次值 / 2` 會把那半次無聲地丟掉（spec 1180）。
//
// ⚠ 原作**不夾下限**：半次值 0 就是這一回合不攻擊（`GIANT SPIDER`／`PHASE
// SPIDER` 的槽 0 就是 0，牠們用槽 1 咬）。
func AdjustBlows(blows, round int) int {
	if blows < 0 {
		return 0
	}
	return (blows + (round & 1)) / 2
}

// AttacksThisRound 回答「這個戰鬥員這一回合打幾下」。
//
// 取捨照原作 `overlay-13:0DD9h`（spec 808）：**架著遠程武器時，類別表的射速
// 取代角色的基準值**，兩條路最後都走 `AdjustBlows`。遠程那一路的射速下限是 2。
func (b *Battle) AttacksThisRound(f Fighter) int {
	blows := f.AttackBlows[0]
	if blows == 0 {
		// 槽 0 沒有攻擊次數就用槽 1（原作 `if 打手^[11Ch] = 0 then 槽 := 2`，
		// spec 1010）。
		blows = f.AttackBlows[1]
	}
	if f.AttacksPerTurn > 0 {
		// 遠程武器已經投影成整數次數（射速全是偶數，換算後與原作相同）。
		return capByAmmunition(f.AttacksPerTurn, f.AmmunitionCount)
	}
	return AdjustBlows(blows, b.round)
}

// capByAmmunition 把遠程攻擊次數壓到剩下的彈藥數。原作寫的是
//
//	m := 1;
//	if 彈藥^[39h] > m then m := 彈藥^[39h];
//	if (m < n) and (彈藥^[39h] > 0) then n := m;
//
// ⚠ **數量 0 不會把次數壓成 1**：`m` 這時是 1，但第二個條件把整條擋掉。
// 這個組合看起來不像刻意設計，照抄是因為它就是原作的行為（spec 808）。
func capByAmmunition(attacks, count int) int {
	if count <= 0 || count >= attacks {
		return attacks
	}
	return count
}

// AttackSequence resolves how many attacks the attacker gets this round.
// Target selection after a target falls belongs to the game adapter.
func (b *Battle) AttackSequence(attackerID, targetID string) ([]AttackResult, error) {
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return nil, fmt.Errorf("unknown attacker %q", attackerID)
	}
	attacks := b.AttacksThisRound(attacker)
	if attacks < 1 {
		attacks = 1
	}
	results := make([]AttackResult, 0, attacks)
	for index := 0; index < attacks && b.status == StatusActive; index++ {
		result, err := b.attackSlot(attackerID, targetID, index+1)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		if result.TargetHP <= 0 {
			break
		}
	}
	return results, nil
}

// Move translates a living fighter by one grid step, or attacks an enemy
// occupying the destination square. Battlefield bounds, terrain and the
// RuleBook's rear-facing details remain owned by a future map/rules adapter.
func (b *Battle) Move(fighterID string, dx, dy int) (Fighter, error) {
	result, err := b.MoveWithFreeAttacks(fighterID, dx, dy)
	return result.Fighter, err
}

// MoveWithFreeAttacks also applies the RuleBook's bounded free-attack trigger
// when a party fighter leaves an enemy's adjacent square. A party fighter
// entering an enemy square resolves an attack without changing squares;
// facing/rear AC is still left to the future combat-map rules layer.
func (b *Battle) MoveWithFreeAttacks(fighterID string, dx, dy int) (MoveResult, error) {
	return b.MoveWithTerrainAndFreeAttacks(fighterID, dx, dy, 0, nil)
}

// MoveWithTerrainAndFreeAttacks validates every destination footprint cell
// before occupancy or attack resolution. maxCost <= 0 keeps the low-level
// compatibility path unbounded; a supplied terrain must return positive cost.
func (b *Battle) MoveWithTerrainAndFreeAttacks(fighterID string, dx, dy, maxCost int, terrain MovementTerrain) (MoveResult, error) {
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return MoveResult{}, fmt.Errorf("unknown fighter %q", fighterID)
	}
	if b.status != StatusActive {
		return MoveResult{}, fmt.Errorf("battle is already over")
	}
	if fighter.HitPoints <= 0 {
		return MoveResult{}, fmt.Errorf("dead fighter cannot move")
	}
	if dx < -1 || dx > 1 || dy < -1 || dy > 1 || (dx == 0 && dy == 0) {
		return MoveResult{}, fmt.Errorf("move delta (%d,%d) is not one grid step", dx, dy)
	}
	if !fighter.HasCombatPosition {
		return MoveResult{}, fmt.Errorf("fighter %q has no combat position", fighterID)
	}
	old := fighter
	nextX, nextY := fighter.CombatX+dx, fighter.CombatY+dy
	movementCost := 1
	if terrain != nil {
		footprint := FootprintForSize(fighter.CombatSize)
		for y := nextY; y < nextY+footprint.Height; y++ {
			for x := nextX; x < nextX+footprint.Width; x++ {
				cost, passable := terrain(x, y)
				if !passable || cost < 1 {
					return MoveResult{}, fmt.Errorf("destination terrain (%d,%d) is impassable", x, y)
				}
				movementCost = max(movementCost, cost)
			}
		}
	}
	if maxCost > 0 && movementCost > maxCost {
		return MoveResult{}, fmt.Errorf("destination terrain costs %d movement points, only %d remain", movementCost, maxCost)
	}
	if fighter.HitDice < 7 && b.cloudkillIntersectsAt(fighter, nextX, nextY) {
		return MoveResult{}, fmt.Errorf("fighter %q cannot enter poisonous cloud", fighterID)
	}
	for _, other := range b.Fighters() {
		if other.ID == fighterID || other.HitPoints <= 0 || !other.HasCombatPosition || !FootprintsOverlapAt(fighter, nextX, nextY, other) {
			continue
		}
		if fighter.Side == SideParty && other.Side == SideEnemy {
			attack, err := b.Attack(fighterID, other.ID)
			if err != nil {
				return MoveResult{}, err
			}
			return MoveResult{Fighter: fighter, Attack: &attack, MovementCost: movementCost}, nil
		}
		return MoveResult{}, fmt.Errorf("destination (%d,%d) is occupied", nextX, nextY)
	}
	fighter.CombatX, fighter.CombatY = nextX, nextY
	b.fighters[fighterID] = fighter
	result := MoveResult{Fighter: fighter, MovementCost: movementCost}
	// Guard is distinct from the RuleBook free attack below: it triggers when
	// an opposing combatant enters the guarder's new adjacency, is consumed
	// before the attack, and is suppressed while the guarder is held.
	for _, guardID := range b.fighterOrder {
		guarder := b.fighters[guardID]
		if guarder.Side == fighter.Side || guarder.HitPoints <= 0 || !guarder.HasCombatPosition ||
			!guarder.CombatAction.Guarding || guarder.MonsterIsHeld() || !adjacent(fighter, guarder) {
			continue
		}
		guarder.CombatAction.Guarding = false
		b.fighters[guardID] = guarder
		// 原作在機會攻擊動手之前先對**被打的那個人**記一次轉向
		// （`194Ah(角色, q)`，spec 817）：動作計數加一、累計轉向加上最短轉法。
		if turnErr := b.AccountTurn(fighterID, guardID); turnErr != nil {
			return MoveResult{}, turnErr
		}
		attack, err := b.Attack(guardID, fighterID)
		if err != nil {
			return MoveResult{}, err
		}
		result.GuardAttacks = append(result.GuardAttacks, attack)
		fighter = b.fighters[fighterID]
		result.Fighter = fighter
		if fighter.HitPoints <= 0 || b.status != StatusActive {
			return result, nil
		}
	}
	if old.Side == SideParty {
		for _, enemy := range b.fighters {
			if enemy.Side != SideEnemy || enemy.HitPoints <= 0 || !enemy.HasCombatPosition {
				continue
			}
			if adjacent(old, enemy) && !adjacent(fighter, enemy) {
				// 原作動手前有四道閘（spec 1010）：打手動得了、看得見離場的人、
				// 打手沒有撤退，最後才是面向（朝向 −2..＋2 五個方向）。
				allows, facingErr := b.opportunityAttackAllowed(enemy.ID, fighter.ID)
				if facingErr != nil {
					return MoveResult{}, facingErr
				}
				if !allows {
					continue
				}
				attack, err := b.Attack(enemy.ID, fighter.ID)
				if err != nil {
					return MoveResult{}, err
				}
				result.FreeAttacks = append(result.FreeAttacks, attack)
				if b.status != StatusActive {
					break
				}
			}
		}
	}
	return result, nil
}

// CastMagicMissile applies the verified RuleBook first-level effect: every
// missile deals 2-5 damage and has no saving throw. The seed-owned RNG keeps
// the result replayable while the game adapter owns spell-slot consumption.
func (b *Battle) CastMagicMissile(casterID, targetID string, level int) (SpellResult, error) {
	return b.castMagicMissile(casterID, targetID, level, 7)
}

// CastMonsterMagicMissile applies the reference monster spell ID 0x0F. Unlike
// the party path, the monster's level-1 spell use is stored on the fighter
// from MON*CHA and is consumed atomically with the effect.
func (b *Battle) CastMonsterMagicMissile(casterID, targetID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	hasSpell := false
	for _, spellID := range caster.MonsterSpellIDs {
		if spellID == MonsterMagicMissileSpellID {
			hasSpell = true
			break
		}
	}
	if !hasSpell {
		return SpellResult{}, fmt.Errorf("monster %q has no Magic Missile spell", casterID)
	}
	if caster.MonsterSpellUses[0] == 0 {
		return SpellResult{}, fmt.Errorf("monster %q has no level-1 spell uses", casterID)
	}
	caster.MonsterSpellUses[0]--
	b.fighters[casterID] = caster
	result, err := b.castMagicMissile(casterID, targetID, 1, MonsterMagicMissileSpellID)
	if err != nil {
		caster.MonsterSpellUses[0]++
		b.fighters[casterID] = caster
		return SpellResult{}, err
	}
	return result, nil
}

func (b *Battle) castMagicMissile(casterID, targetID string, level int, spellID uint8) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if level < 1 {
		return SpellResult{}, fmt.Errorf("caster level must be positive")
	}
	missiles := (level + 1) / 2
	if missiles > 6 {
		missiles = 6
	}
	damage := 0
	for index := 0; index < missiles; index++ {
		damage += b.rng.Intn(4) + 2
	}
	resisted := false
	// Magic Missile reaches the original pre-damage affect boundary with the
	// Magic damage flag set. Damage dice are consumed before this d100.
	resisted = b.rollMagicResistance(target, level)
	if resisted {
		damage = 0
	}
	damage = b.applyPositiveDamage(&target, damage)
	b.fighters[targetID] = target
	b.updateStatus()
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: spellID, Missiles: missiles, Damage: damage, TargetHP: target.HitPoints, Resisted: resisted}, nil
}

// CastFireball applies the reference 0x2F area spell: one level-d6 damage
// roll is shared by every living combatant within the radius-two target
// list; each target independently saves versus Spell for half damage. The
// original path-distance implementation also consults combat terrain. This
// bounded core uses the equivalent unobstructed two-tile footprint distance;
// terrain occlusion remains an explicit adapter boundary.
func (b *Battle) CastFireball(casterID string, center TilePoint, level int) (AreaSpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return AreaSpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return AreaSpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 {
		return AreaSpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if level < 1 {
		return AreaSpellResult{}, fmt.Errorf("caster level must be positive")
	}
	targets := make([]Fighter, 0)
	for _, fighter := range b.Fighters() {
		if fighter.HitPoints <= 0 || !fighter.HasCombatPosition ||
			!fighterFootprintWithinRadius(fighter, center, 2) {
			continue
		}
		if len(fighter.SavingThrows) <= 4 {
			return AreaSpellResult{}, fmt.Errorf("fighter %q has no spell saving throw", fighter.ID)
		}
		targets = append(targets, fighter)
	}
	sort.SliceStable(targets, func(i, j int) bool {
		left := fireballSortKey(targets[i], center)
		right := fireballSortKey(targets[j], center)
		if left != right {
			return left < right
		}
		return targets[i].ID < targets[j].ID
	})
	damage := 0
	for roll := 0; roll < level; roll++ {
		damage += b.rng.Intn(6) + 1
	}
	result := AreaSpellResult{
		CasterID: casterID, SpellID: FireballSpellID, Center: center, BaseDamage: damage,
		Impacts: make([]AreaSpellImpact, 0, len(targets)),
	}
	for _, target := range targets {
		saveRoll := b.rng.Intn(20) + 1
		conditional := target.MonsterConditionalModifierAgainst(caster)
		saved := saveRoll == 20 ||
			saveRoll != 1 && saveRoll+target.SavingThrowBonus+conditional.SavingThrowDelta >= int(target.SavingThrows[4])
		applied := damage
		if saved {
			applied /= 2
		}
		resisted := b.rollMagicResistance(target, level)
		if resisted {
			applied = 0
		}
		adjustment := target.MonsterDamageAdjustment(DamageFlagFire|DamageFlagMagic, applied)
		protected := !resisted && adjustment.Immune
		applied = adjustment.Damage
		applied = b.applyPositiveDamage(&target, applied)
		b.fighters[target.ID] = target
		result.Impacts = append(result.Impacts, AreaSpellImpact{
			TargetID: target.ID, Damage: applied, TargetHP: target.HitPoints, Saved: saved,
			Resisted: resisted, Protected: protected,
		})
	}
	b.updateStatus()
	return result, nil
}

func fighterFootprintWithinRadius(fighter Fighter, center TilePoint, radius int) bool {
	footprint := FootprintForSize(fighter.CombatSize)
	for y := fighter.CombatY; y < fighter.CombatY+footprint.Height; y++ {
		for x := fighter.CombatX; x < fighter.CombatX+footprint.Width; x++ {
			dx := abs(x - center.X)
			dy := abs(y - center.Y)
			// Reference canReachTarget accepts path steps <= range*2+1;
			// cardinal movement costs 2 and diagonal movement costs 3.
			if 2*max(dx, dy)+min(dx, dy) <= radius*2+1 {
				return true
			}
		}
	}
	return false
}

// fireballSortKey follows the reference list's cardinal=2, diagonal=3 step
// metric. Direction tie-breaking is approximated by stable fighter ID until
// the raw combatant-array order is retained by Battle.
func fireballSortKey(fighter Fighter, center TilePoint) int {
	footprint := FootprintForSize(fighter.CombatSize)
	best := int(^uint(0) >> 1)
	for y := fighter.CombatY; y < fighter.CombatY+footprint.Height; y++ {
		for x := fighter.CombatX; x < fighter.CombatX+footprint.Width; x++ {
			dx := abs(x - center.X)
			dy := abs(y - center.Y)
			best = min(best, 2*max(dx, dy)+min(dx, dy))
		}
	}
	return best
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// CastCureLightWounds applies the verified 1-8 HP touch-heal effect. The
// caller decides whether the caster has the clerical slot and which friendly
// target is legal; Battle only owns deterministic dice and HP mutation.
func (b *Battle) CastCureLightWounds(casterID, targetID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || (target.HitPoints <= 0 && !target.DownedCorpse) {
		return SpellResult{}, fmt.Errorf("dead fighter cannot use Cure Light Wounds")
	}
	healing := b.rng.Intn(8) + 1
	if healing > target.MaxHitPoints-target.HitPoints {
		healing = target.MaxHitPoints - target.HitPoints
	}
	if healing < 0 {
		healing = 0
	}
	target.HitPoints += healing
	if target.HitPoints > 0 {
		target.DeathOverlay = false
	}
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 3, Healing: healing, TargetHP: target.HitPoints}, nil
}

// CastCauseLightWounds applies the verified 1-8 HP touch damage. The target
// must be an adjacent living enemy when both fighters carry CombatMap
// positions; position-less direct callers retain the bounded fallback.
func (b *Battle) CastCauseLightWounds(casterID, targetID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideEnemy {
		return SpellResult{}, fmt.Errorf("Cause Light Wounds target %q is not an enemy", targetID)
	}
	if caster.HasCombatPosition && target.HasCombatPosition && !adjacent(caster, target) {
		return SpellResult{}, fmt.Errorf("Cause Light Wounds target %q is out of touch range", targetID)
	}
	damage := b.rng.Intn(8) + 1
	damage = b.applyPositiveDamage(&target, damage)
	b.fighters[targetID] = target
	b.updateStatus()
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 4, Damage: damage, TargetHP: target.HitPoints}, nil
}

// CastProtectionFromEvil applies the verified conditional AC protection. The
// caller supplies a living party target; alignment-aware attack resolution is
// intentionally kept in Battle rather than mutating base ArmorClass.
func (b *Battle) CastProtectionFromEvil(casterID, targetID string, casterLevel int) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideParty {
		return SpellResult{}, fmt.Errorf("Protection from Evil target %q is not party", targetID)
	}
	if casterLevel < 1 {
		return SpellResult{}, fmt.Errorf("caster level must be positive")
	}
	if target.ProtectedFromEvil {
		return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 6}, nil
	}
	target.ProtectedFromEvil = true
	target.ProtectionEvilRounds = 3 * casterLevel
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 6, Targets: 1}, nil
}

func (b *Battle) CastProtectionFromGood(casterID, targetID string, casterLevel int) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideParty {
		return SpellResult{}, fmt.Errorf("Protection from Good target %q is not party", targetID)
	}
	if casterLevel < 1 {
		return SpellResult{}, fmt.Errorf("caster level must be positive")
	}
	if target.ProtectedFromGood {
		return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 7}, nil
	}
	target.ProtectedFromGood = true
	target.ProtectionGoodRounds = 3 * casterLevel
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 7, Targets: 1}, nil
}

func adjacent(first, second Fighter) bool {
	return footprintAdjacent(first, second)
}

// CastBless applies the verified first-level party-wide attack bonus, using
// CombatMap adjacency and the six-round duration contract.
func (b *Battle) CastBless(casterID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	affected := 0
	for id, fighter := range b.fighters {
		if fighter.Side != SideParty || fighter.HitPoints <= 0 || fighter.Blessed || b.adjacentToLivingEnemy(fighter) {
			continue
		}
		fighter.AttackBonus++
		fighter.Blessed = true
		fighter.BlessRounds = 6
		b.fighters[id] = fighter
		affected++
	}
	return SpellResult{CasterID: casterID, SpellID: 1, Targets: affected}, nil
}

func (b *Battle) adjacentToLivingEnemy(party Fighter) bool {
	if !party.HasCombatPosition {
		return false
	}
	for _, enemy := range b.fighters {
		if enemy.Side != SideEnemy || enemy.HitPoints <= 0 || !enemy.HasCombatPosition {
			continue
		}
		if adjacent(party, enemy) {
			return true
		}
	}
	return false
}

func (b *Battle) advanceBlessDurations() {
	for id, fighter := range b.fighters {
		if !fighter.Blessed || fighter.BlessRounds <= 0 {
			continue
		}
		fighter.BlessRounds--
		if fighter.BlessRounds == 0 {
			fighter.AttackBonus--
			fighter.Blessed = false
		}
		b.fighters[id] = fighter
	}
}

func (b *Battle) CastCurse(casterID, targetID string) (SpellResult, error) {
	caster, ok := b.fighters[casterID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return SpellResult{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return SpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 || target.HitPoints <= 0 {
		return SpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if target.Side != SideEnemy {
		return SpellResult{}, fmt.Errorf("Curse target %q is not an enemy", targetID)
	}
	if target.Cursed || b.adjacentToLivingParty(target) {
		return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 2}, nil
	}
	target.AttackBonus--
	target.Cursed = true
	target.CurseRounds = 6
	b.fighters[targetID] = target
	return SpellResult{CasterID: casterID, TargetID: targetID, SpellID: 2, Targets: 1}, nil
}

func (b *Battle) adjacentToLivingParty(enemy Fighter) bool {
	if !enemy.HasCombatPosition {
		return false
	}
	for _, party := range b.fighters {
		if party.Side != SideParty || party.HitPoints <= 0 || !party.HasCombatPosition {
			continue
		}
		if adjacent(enemy, party) {
			return true
		}
	}
	return false
}

func (b *Battle) advanceCurseDurations() {
	for id, fighter := range b.fighters {
		if !fighter.Cursed || fighter.CurseRounds <= 0 {
			continue
		}
		fighter.CurseRounds--
		if fighter.CurseRounds == 0 {
			fighter.AttackBonus++
			fighter.Cursed = false
		}
		b.fighters[id] = fighter
	}
}

func (b *Battle) advanceProtectionDurations() {
	for id, fighter := range b.fighters {
		if !fighter.ProtectedFromEvil || fighter.ProtectionEvilRounds <= 0 {
			continue
		}
		fighter.ProtectionEvilRounds--
		if fighter.ProtectionEvilRounds == 0 {
			fighter.ProtectedFromEvil = false
		}
		b.fighters[id] = fighter
	}
	for id, fighter := range b.fighters {
		if !fighter.ProtectedFromGood || fighter.ProtectionGoodRounds <= 0 {
			continue
		}
		fighter.ProtectionGoodRounds--
		if fighter.ProtectionGoodRounds == 0 {
			fighter.ProtectedFromGood = false
		}
		b.fighters[id] = fighter
	}
}

func (b *Battle) updateStatus() {
	partyAlive, enemyAlive, partyFled := false, false, false
	for _, fighter := range b.fighters {
		if fighter.HitPoints <= 0 {
			continue
		}
		// 離場的人還活著，但不在戰場上——判勝負時當成不在，
		// 判「隊伍是不是被打光」時仍要算進去。
		if fighter.Escaped {
			if fighter.Side == SideParty {
				partyFled = true
			}
			continue
		}
		if fighter.Side == SideParty {
			partyAlive = true
		} else {
			enemyAlive = true
		}
	}
	switch {
	case partyAlive && enemyAlive:
		b.status = StatusActive
	case partyAlive:
		b.status = StatusPartyWon
	case partyFled && enemyAlive:
		// 場上沒有隊員了，但至少一個是走出去的 ⇒ 逃離，不是被打光。
		b.status = StatusPartyFled
	case enemyAlive:
		b.status = StatusEnemyWon
	case partyFled:
		b.status = StatusPartyFled
	default:
		b.status = StatusDraw
	}
}

// ReplaceFighterEquipment 把重新投影過的裝備衍生值換到場上的戰鬥員身上。
//
// ★ 只換裝備算得出來的那幾格。 位置、生命、效果串列、行動狀態都留著——
// 換武器不該讓人瞬移或回血。原作的換裝也是「動裝備槽再叫派生數值重算」，
// 不是重建整個戰鬥員。
func (b *Battle) ReplaceFighterEquipment(fighterID string, projected Fighter) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	fighter.ArmorClass = projected.ArmorClass
	fighter.AttackBonus = projected.AttackBonus
	fighter.DamageDiceCount = projected.DamageDiceCount
	fighter.DamageDiceSides = projected.DamageDiceSides
	fighter.DamageBonus = projected.DamageBonus
	// 大型目標那一組與「有沒有槽 0 武器」也要跟著換——少了它，換裝之後打大型
	// 目標用的還是上一把武器的骰（spec 1175）。
	fighter.LargeDamageDiceCount = projected.LargeDamageDiceCount
	fighter.LargeDamageDiceSides = projected.LargeDamageDiceSides
	fighter.LargeDamageBonus = projected.LargeDamageBonus
	fighter.HasSlotZeroWeapon = projected.HasSlotZeroWeapon
	fighter.WeaponRange = projected.WeaponRange
	fighter.MissileWeapon = projected.MissileWeapon
	fighter.ThrownWeapon = projected.ThrownWeapon
	fighter.WeaponItemType = projected.WeaponItemType
	fighter.AmmunitionType = projected.AmmunitionType
	fighter.AmmunitionCount = projected.AmmunitionCount
	fighter.MovementAllowance = projected.MovementAllowance
	fighter.ReadiedItemEffects = append([]uint8(nil), projected.ReadiedItemEffects...)
	b.fighters[fighterID] = fighter
	return nil
}

// ReplaceMonsterItems 換掉怪物的隨身物品鏈。AI 換裝（spec 1004）動的是
// Readied 旗標；衍生值另外走 ReplaceFighterEquipment。
func (b *Battle) ReplaceMonsterItems(fighterID string, items []MonsterItem) error {
	if b == nil {
		return fmt.Errorf("battle is nil")
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	fighter.MonsterItems = append([]MonsterItem(nil), items...)
	b.fighters[fighterID] = fighter
	return nil
}

// CastHealingDice 是「擲 NdM ＋ K 治療一個目標」，上限是掉的血。
// 骰子來自 spec 1124 的表，不寫死在這裡——治療輕／中／重傷差別只有骰數與加值
// （`1d8`、`2d8 ＋ 1`、`3d8 ＋ 3`）。
func (b *Battle) CastHealingDice(casterID, targetID string,
	dice SpellDamageFormula) (SpellResult, error) {
	_, target, err := b.spellDiceEndpoints(casterID, targetID, dice)
	if err != nil {
		return SpellResult{}, err
	}
	healing := dice.Bonus
	for roll := 0; roll < dice.Count; roll++ {
		healing += b.rng.Intn(dice.Sides) + 1
	}
	if healing > target.MaxHitPoints-target.HitPoints {
		healing = target.MaxHitPoints - target.HitPoints
	}
	if healing < 0 {
		healing = 0
	}
	target.HitPoints += healing
	if target.HitPoints > 0 {
		target.DeathOverlay = false
	}
	b.fighters[targetID] = target
	b.updateStatus()
	return SpellResult{CasterID: casterID, TargetID: targetID,
		Healing: healing, TargetHP: target.HitPoints}, nil
}

// CastDamageDice 是「擲 NdM 打一個目標」。`halve` 是豁免過關而且這支法術屬於
// 「過了減半」那一類（`+8h = 2`）。
//
// ⚠ 折半要在**套上去之前**做。先打滿再補回一半會讓「打死了又活過來」這種
// 情況出現，而且回報的數字對不上實際掉的血。
func (b *Battle) CastDamageDice(casterID, targetID string, dice SpellDamageFormula,
	element uint8, halve bool) (SpellResult, error) {
	_, target, err := b.spellDiceEndpoints(casterID, targetID, dice)
	if err != nil {
		return SpellResult{}, err
	}
	damage := dice.Bonus
	for roll := 0; roll < dice.Count; roll++ {
		damage += b.rng.Intn(dice.Sides) + 1
	}
	if halve {
		damage /= 2
	}
	// 傷害時機（`CHECKFX(06h)`，spec 1123）：抗寒／抗火在這裡折半，而它們各自
	// 只認傷害屬性旗標的一個位元——所以旗標要一起傳進去。
	adjusted, err := CheckFX(target, checkFXDamage, map[string]int{
		scratchDamage: damage, scratchDamageElement: int(element)})
	// ⚠ 這一次查詢除了折半傷害，也可能**改記錄**（效果 `54h`：被電回血 8 點）。
	// 原本只讀 `Applied[damage]`，`Records` 就這樣掉了。
	if err == nil {
		applyRecordWritesTo(&target, adjusted)
	}
	if err != nil {
		return SpellResult{}, err
	}
	damage = adjusted.Applied[scratchDamage]
	if damage < 0 {
		damage = 0
	}
	applied := b.applyPositiveDamage(&target, damage)
	b.fighters[targetID] = target
	b.updateStatus()
	return SpellResult{CasterID: casterID, TargetID: targetID,
		Damage: applied, TargetHP: target.HitPoints}, nil
}

// spellDiceEndpoints 收攏兩支共用的前置檢查。
func (b *Battle) spellDiceEndpoints(casterID, targetID string,
	dice SpellDamageFormula) (Fighter, Fighter, error) {
	if b == nil || b.rng == nil {
		return Fighter{}, Fighter{}, errNoPRNG
	}
	// ★ 骰數 0 是合法的：燃燒之手不擲骰，傷害就是施法者等級（走 Bonus）。
	// 這裡擋的是「整個算式加起來是 0」，不是「沒擲骰」。
	if dice.Count < 0 || dice.Sides < 0 || dice.Total() <= 0 {
		return Fighter{}, Fighter{}, fmt.Errorf("dice %dd%d+%d is not rollable",
			dice.Count, dice.Sides, dice.Bonus)
	}
	if dice.Count > 0 && dice.Sides <= 0 {
		return Fighter{}, Fighter{}, fmt.Errorf("dice %dd%d has no sides", dice.Count, dice.Sides)
	}
	caster, ok := b.fighters[casterID]
	if !ok {
		return Fighter{}, Fighter{}, fmt.Errorf("unknown caster %q", casterID)
	}
	target, ok := b.fighters[targetID]
	if !ok {
		return Fighter{}, Fighter{}, fmt.Errorf("unknown target %q", targetID)
	}
	if b.status != StatusActive {
		return Fighter{}, Fighter{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 {
		return Fighter{}, Fighter{}, fmt.Errorf("dead fighter cannot cast")
	}
	return caster, target, nil
}

// CastAreaDamageDice 是「擲一次 NdM，套給中心半徑內的每一個人」。
//
// ★ 傷害只擲**一次**，每個目標各自擲豁免——與火球那一支同一個形狀（spec 731）。
// 每個目標擲一次傷害會讓範圍法術的方差大到與原版不同。
//
// 傷害屬性旗標一起傳進 `CHECKFX(06h)`，所以抗寒／抗火在這裡自然生效。
//
// ⚠ `requiresSave` 為 false 時**完全不擲豁免**（`+8h = 0`，spec 1111）。
// 照擲會多消耗亂數，而且「豁免過了就沒事」的分支會讓不該有豁免的法術漏傷害。
func (b *Battle) CastAreaDamageDice(casterID string, center TilePoint,
	dice SpellDamageFormula, radius int, element uint8,
	requiresSave bool, saveCategory int, saveHalves bool) (AreaSpellResult, error) {
	if b == nil || b.rng == nil {
		return AreaSpellResult{}, errNoPRNG
	}
	caster, ok := b.fighters[casterID]
	if !ok {
		return AreaSpellResult{}, fmt.Errorf("unknown caster %q", casterID)
	}
	if b.status != StatusActive {
		return AreaSpellResult{}, fmt.Errorf("battle is already over")
	}
	if caster.HitPoints <= 0 {
		return AreaSpellResult{}, fmt.Errorf("dead fighter cannot cast")
	}
	if dice.Count < 0 || dice.Sides < 0 || dice.Total() <= 0 || radius < 0 {
		return AreaSpellResult{}, fmt.Errorf("area spell %dd%d+%d radius %d is not castable",
			dice.Count, dice.Sides, dice.Bonus, radius)
	}
	base := dice.Bonus
	for roll := 0; roll < dice.Count; roll++ {
		base += b.rng.Intn(dice.Sides) + 1
	}
	result := AreaSpellResult{CasterID: casterID, Center: center, BaseDamage: base}
	for _, id := range b.fighterOrder {
		target := b.fighters[id]
		if target.HitPoints <= 0 || target.Escaped || !target.HasCombatPosition ||
			!fighterFootprintWithinRadius(target, center, radius) {
			continue
		}
		damage := base
		saved := false
		if requiresSave {
			save, err := b.RollSavingThrow(target, saveCategory, 0)
			if err != nil {
				return AreaSpellResult{}, err
			}
			saved = save.Saved
		}
		if saved {
			if !saveHalves {
				result.Impacts = append(result.Impacts, AreaSpellImpact{
					TargetID: id, Saved: true})
				continue
			}
			damage /= 2
		}
		adjusted, err := CheckFX(target, checkFXDamage, map[string]int{
			scratchDamage: damage, scratchDamageElement: int(element)})
		if err != nil {
			return AreaSpellResult{}, err
		}
		// 同上：這一次查詢也可能改記錄（效果 `54h`）。
		applyRecordWritesTo(&target, adjusted)
		damage = adjusted.Applied[scratchDamage]
		if damage < 0 {
			damage = 0
		}
		applied := b.applyPositiveDamage(&target, damage)
		b.fighters[id] = target
		result.Impacts = append(result.Impacts, AreaSpellImpact{
			TargetID: id, Saved: saved,
			Damage: applied, TargetHP: target.HitPoints})
	}
	b.updateStatus()
	return result, nil
}

// UsesSeparateAmmunition 回答「這把武器要另外的彈藥嗎」，照原作自動換裝那一支
// 的條件（spec 1120 的虛擬碼）：
//
//	投擲（`+0Eh` bit 4）        → 自己就是彈藥
//	`+0Eh` 剛好等於 `0Ah`       → 自給自足（投石索）
//	其餘有發射位元（bit 3）的   → 要另外的彈藥（弓、弩）
//
// ★ 音效那一側要用它（spec 1186）：要另外彈藥的武器，飛出去的是箭或弩矢
// （類別 `49h`／`1Ch`），**兩個都落在 `SHOWARROW` 的 ARROWFX 分支**——所以
// 不必知道架著的是哪一件彈藥，結論一樣。
func (f Fighter) UsesSeparateAmmunition() bool {
	const selfSufficient = 0x0A
	return f.MissileWeapon && !f.ThrownWeapon && f.AmmunitionType != selfSufficient
}
