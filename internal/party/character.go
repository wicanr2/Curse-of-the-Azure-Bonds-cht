// Package party contains the data-neutral character-creation contract. It
// follows the seven player races and class minimums observed in the CoAB rule
// book/reference rewrite; combat serialization remains a separate adapter.
package party

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
	engineability "github.com/wicanr2/golden-box-remake-engine/combat/ability"
)

type Race uint8

const (
	RaceDwarf Race = iota
	RaceElf
	RaceGnome
	RaceHalfElf
	RaceHalfling
	RaceHuman
	RaceHalfOrc
	// RaceSaurial is the non-player race projection used by Dragonbait's
	// MON3 player record. It is not exposed by character creation.
	RaceSaurial
)

// Gender 是角色記錄 +119h 的值。原作的種族屬性表把每組限制排成
// （男, 女）兩格，索引就是這個欄位，所以 0 ＝ 男、1 ＝ 女。
type Gender uint8

const (
	GenderMale Gender = iota
	GenderFemale
)

func (g Gender) String() string {
	if g == GenderFemale {
		return "female"
	}
	return "male"
}

type Class uint8

const (
	ClassCleric Class = iota
	ClassFighter
	ClassRanger
	ClassPaladin
	ClassMagicUser
	ClassThief
)

// Abilities 依 spec 1079 的成對版面保存六個屬性：
//
//	屬性 i  基準值 +10h+i×2（建角時決定，不受裝備影響）
//	        目前值 +11h+i×2（每次裝備／效果變動就重算）
//	力量另有百分比：目前 +1Ch、基準 +1Dh
//
// 只存一格會讓脫下裝備後屬性回不去，所以兩格都要留。
//
// ⚠ 既有欄位的實際對應（匯入 DOS 記錄時就是這樣讀的）：
// Strength ＝ +10h（基準）、StrengthFull ＝ +11h（目前力量）、
// StrengthExceptional ＝ +1Ch（目前百分比）；
// Intelligence..Charisma 是 +12h/+14h/+16h/+18h/+1Ah，全部是**基準值**。
type Abilities struct {
	Strength int
	// StrengthFull 是 +11h 的目前力量；StrengthExceptional 是 +1Ch 的
	// 目前百分比。名稱沿用早期匯入時的命名，語意見上方說明。
	StrengthFull        int
	StrengthExceptional int
	Intelligence        int
	Wisdom              int
	Dexterity           int
	Constitution        int
	Charisma            int
	// Current 是五個非力量屬性的目前值（+13h/+15h/+17h/+19h/+1Bh），
	// 順序與 Value 的索引 1..5 相同：智、睿、敏、體、魅。
	// 零值代表「還沒算過」，此時一律以基準值為準（見 CurrentValue）。
	Current [5]int
	// StrengthExceptionalBase 是 +1Dh 的基準百分比。
	StrengthExceptionalBase int
}

// CurrentValue 回傳第 index 個屬性的目前值。目前值沒被算過（零值）時
// 退回基準值——原作建角收尾就是把兩者設成一樣（spec 1094 §三）。
func (a Abilities) CurrentValue(index int) (int, error) {
	if index == 0 {
		if a.StrengthFull != 0 {
			return a.StrengthFull, nil
		}
		return a.Strength, nil
	}
	if index < 1 || index > 5 {
		return 0, fmt.Errorf("ability index %d is out of range", index)
	}
	if a.Current[index-1] != 0 {
		return a.Current[index-1], nil
	}
	return a.Value(index)
}

// SyncBaseFromCurrent 是原作建角的收尾（spec 1094 §三）：
//
//	for i := 0 to 5 do 角色^[10h + i×2] := 角色^[11h + i×2];
//	角色^[1Dh] := 角色^[1Ch];
//
// 六個屬性與力量百分比的目前值原封不動抄成基準值，之後裝備／法術只改
// 目前值，脫下來就能回到基準值。
func (a *Abilities) SyncBaseFromCurrent() {
	for index := 0; index <= 5; index++ {
		value, err := a.CurrentValue(index)
		if err != nil {
			continue
		}
		_ = a.set(index, value)
	}
	a.StrengthFull = a.Strength
	a.StrengthExceptionalBase = a.StrengthExceptional
	a.Current = [5]int{}
}

// SyncCurrentFromBase 把目前值重設成基準值——卸下所有裝備／效果後的狀態。
func (a *Abilities) SyncCurrentFromBase() {
	a.StrengthFull = a.Strength
	a.StrengthExceptional = a.StrengthExceptionalBase
	a.Current = [5]int{}
}

// StartingAgeSpec is the reference race/class age generator shape.
type StartingAgeSpec struct {
	BaseAge   int
	DiceCount int
	DiceSize  int
}

// DOSRaceID 把本套件的 Race 常數換成原作角色記錄 +74h 的種族編號（1..7）。
// 兩套編號不同：Race 從 0 起算（RaceDwarf = iota），原作 0 不是合法值。
func DOSRaceID(race Race) (int, bool) {
	switch race {
	case RaceDwarf:
		return 1, true
	case RaceElf:
		return 2, true
	case RaceGnome:
		return 3, true
	case RaceHalfElf:
		return 4, true
	case RaceHalfling:
		return 5, true
	case RaceHalfOrc:
		return 6, true
	case RaceHuman:
		return 7, true
	}
	return 0, false
}

// DOSClassSlot 把本套件的 Class 常數換成原作的職業槽名。
// 槽順序見 spec 1084：牧師／德魯伊／戰士／聖騎士／遊俠／法師／盜賊／武僧。
func DOSClassSlot(class Class) (string, bool) {
	switch class {
	case ClassCleric:
		return "cleric", true
	case ClassFighter:
		return "fighter", true
	case ClassRanger:
		return "ranger", true
	case ClassPaladin:
		return "paladin", true
	case ClassMagicUser:
		return "magic-user", true
	case ClassThief:
		return "thief", true
	}
	return "", false
}

// StartingAgeLookup 是建角年齡表的查詢介面。資料由 game pack 提供
// （gamepack.Tables），本套件只放機制——避免把原版資料表 hardcode 在 Go 裡。
type StartingAgeLookup interface {
	StartingAgeFor(raceID int, classSlot string) (baseAge, diceCount, diceSize int, ok bool)
}

// StartingAgeSpecFrom 由注入的表查出單職起始年齡。
// 查不到代表原作沒有為這個種族／職業準備年齡。
func StartingAgeSpecFrom(tables StartingAgeLookup, race Race, class Class) (StartingAgeSpec, error) {
	raceID, ok := DOSRaceID(race)
	if !ok {
		return StartingAgeSpec{}, fmt.Errorf("race %d has no DOS race id", race)
	}
	slot, ok := DOSClassSlot(class)
	if !ok {
		return StartingAgeSpec{}, fmt.Errorf("class %d has no DOS class slot", class)
	}
	base, count, size, ok := tables.StartingAgeFor(raceID, slot)
	if !ok || count == 0 || size == 0 {
		return StartingAgeSpec{}, fmt.Errorf("%s cannot generate a single-class age for %s", race, class)
	}
	return StartingAgeSpec{BaseAge: base, DiceCount: count, DiceSize: size}, nil
}

// RollStartingAgeFrom 用注入的表擲起始年齡。
func RollStartingAgeFrom(tables StartingAgeLookup, race Race, class Class, seed int64) (int16, error) {
	spec, err := StartingAgeSpecFrom(tables, race, class)
	if err != nil {
		return 0, err
	}
	rng := rand.New(rand.NewSource(seed))
	age := spec.BaseAge
	for i := 0; i < spec.DiceCount; i++ {
		age += rng.Intn(spec.DiceSize) + 1
	}
	return int16(age), nil
}


// WithAgeEffects applies the reference new-character age brackets to a base
// ability roll. DOS player records already contain the resulting stats, so
// import and combat projection must not call this implicitly.
func (a Abilities) WithAgeEffects(race Race, age int) (Abilities, error) {
	if race < RaceDwarf || race > RaceHalfOrc {
		return Abilities{}, fmt.Errorf("race %d has no CoAB age table", race)
	}
	brackets := [...][]int{
		{50, 150, 250, 350, 450},       // dwarf
		{175, 550, 875, 1200, 1600},    // elf
		{90, 300, 450, 600, 750},       // gnome
		{40, 100, 175, 250, 325},       // half-elf
		{0x21, 0x44, 0x65, 0x90, 0xC7}, // halfling
		{0x14, 0x28, 0x3C, 0x5A, 0x78}, // human
		{0x0F, 0x1E, 0x2D, 0x3C, 0x50}, // half-orc
	}
	effects := [...][5]int{
		{0, 1, -1, -2, -1}, // strength
		{0, 0, 1, 0, 1},    // intelligence
		{-1, 1, 1, 1, 1},   // wisdom
		{0, 0, 0, -2, -1},  // dexterity
		{1, 0, -1, -1, -1}, // constitution
		{0, 0, 0, 0, 0},    // charisma
	}
	result := a
	values := []*int{&result.Strength, &result.Intelligence, &result.Wisdom, &result.Dexterity, &result.Constitution, &result.Charisma}
	for stat := range values {
		for bracket, threshold := range brackets[race] {
			if threshold < age {
				*values[stat] += effects[stat][bracket]
			}
		}
	}
	if result.StrengthFull != 0 {
		for bracket, threshold := range brackets[race] {
			if threshold < age {
				result.StrengthFull += effects[0][bracket]
			}
		}
	}
	return result, nil
}

// HealthStatus mirrors the reference player's combat health states while
// keeping the existing zero-value JSON saves valid.
type HealthStatus uint8

const (
	HealthStatusOK HealthStatus = iota
	HealthStatusAnimated
	HealthStatusUnconscious
	HealthStatusDying
	HealthStatusDead
	HealthStatusStoned
)

// RollAbilities reproduces the rule-book's six ability generation shape: six
// independent 3d6 rolls. A caller-provided seed keeps tests and replays
// deterministic without coupling the party package to UI randomness.
func RollAbilities(seed int64) Abilities {
	rng := rand.New(rand.NewSource(seed))
	roll := func() int {
		total := 0
		for i := 0; i < 3; i++ {
			total += rng.Intn(6) + 1
		}
		return total
	}
	return Abilities{Strength: roll(), Intelligence: roll(), Wisdom: roll(), Dexterity: roll(), Constitution: roll(), Charisma: roll()}
}

func (a Abilities) Value(index int) (int, error) {
	values := [...]int{a.Strength, a.Intelligence, a.Wisdom, a.Dexterity, a.Constitution, a.Charisma}
	if index < 0 || index >= len(values) {
		return 0, fmt.Errorf("ability index %d is out of range", index)
	}
	return values[index], nil
}

// AbilityLimitLookup 是屬性上下限的查詢介面。資料由 game pack 提供
// （gamepack.LimitLookup），本套件只放機制。
type AbilityLimitLookup interface {
	AbilityLimits(raceID, gender, classCombo int) ([6][2]int, bool)
}

// StrengthPercentileLookup 讓機制層查得到力量百分比上限（0 ＝ 不能有 18/xx）。
type StrengthPercentileLookup interface {
	StrengthPercentileMax(raceID, gender int) (int, bool)
}

// AdjustWithin 依原作規則調整一個屬性（spec 1086 §三）：
//
//   - 力量（index 0）的 18/xx 百分比是**獨立的一階**：往下調時先把百分比
//     一點一點扣掉，扣到 0 才動整數部分；往上調是鏡像的同一套——整數到了
//     上限之後才開始加百分比。
//   - 其餘五個屬性只做兩次夾值：先夾種族的上下限，再夾職業組合的最低要求。
//
// tables 為 nil 時退回 3..18 的通用範圍。
func (a *Abilities) AdjustWithin(tables AbilityLimitLookup, character *Character, index, delta int) error {
	if delta == 0 {
		return nil
	}
	value, err := a.Value(index)
	if err != nil {
		return err
	}
	low, high := 3, 18
	percentileMax := 0
	if tables != nil && character != nil {
		raceID, ok := DOSRaceID(character.Race)
		if !ok {
			return fmt.Errorf("race %d has no DOS race id", character.Race)
		}
		limits, ok := tables.AbilityLimits(raceID, int(character.Gender), int(character.RawClassID))
		if !ok {
			return fmt.Errorf("race %d has no ability limits", character.Race)
		}
		low, high = limits[index][0], limits[index][1]
		if percentiles, ok := tables.(StrengthPercentileLookup); ok {
			if maximum, ok := percentiles.StrengthPercentileMax(raceID, int(character.Gender)); ok {
				percentileMax = maximum
			}
		}
	}

	// 力量的百分比階：只有整數已經在上限、且該種族／性別允許 18/xx 時才存在。
	if index == 0 && percentileMax > 0 {
		if delta < 0 && a.StrengthExceptional > 0 {
			a.StrengthExceptional--
			return nil
		}
		if delta > 0 && value >= high {
			if a.StrengthExceptional >= percentileMax {
				return fmt.Errorf("strength percentile must remain at most %d", percentileMax)
			}
			a.StrengthExceptional++
			return nil
		}
	}
	// 整數部分回到上限以下時，百分比必須歸零——18/xx 只存在於力量 18。
	if index == 0 && delta < 0 && a.StrengthExceptional > 0 {
		a.StrengthExceptional = 0
	}

	if value+delta < low {
		return fmt.Errorf("ability value must remain at least %d", low)
	}
	if value+delta > high {
		return fmt.Errorf("ability value must remain at most %d", high)
	}
	return a.set(index, value+delta)
}

func (a *Abilities) set(index, value int) error {
	switch index {
	case 0:
		a.Strength = value
	case 1:
		a.Intelligence = value
	case 2:
		a.Wisdom = value
	case 3:
		a.Dexterity = value
	case 4:
		a.Constitution = value
	case 5:
		a.Charisma = value
	default:
		return fmt.Errorf("ability index %d is out of range", index)
	}
	return nil
}

func (a *Abilities) Adjust(index, delta int) error {
	if delta == 0 {
		return nil
	}
	value, err := a.Value(index)
	if err != nil {
		return err
	}
	if value+delta < 3 || value+delta > 18 {
		return fmt.Errorf("ability value must remain between 3 and 18")
	}
	value += delta
	switch index {
	case 0:
		a.Strength = value
	case 1:
		a.Intelligence = value
	case 2:
		a.Wisdom = value
	case 3:
		a.Dexterity = value
	case 4:
		a.Constitution = value
	case 5:
		a.Charisma = value
	}
	return nil
}

type Character struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// ScriptName preserves the DOS name used by ECL string comparisons when
	// Name is localized for display. Empty keeps old saves compatible.
	ScriptName string `json:"script_name,omitempty"`
	Race       Race   `json:"race"`
	Class      Class  `json:"class"`
	// RawClassID preserves the DOS ClassId, including multi-class IDs 8..16.
	RawClassID uint8     `json:"raw_class_id,omitempty"`
	Abilities  Abilities `json:"abilities"`
	Level      int       `json:"level"`
	// Experience preserves the DOS Player.exp dword at offset 0x127.
	Experience uint32 `json:"experience,omitempty"`
	// ClassLevels preserves the eight raw DOS class-level slots. Class remains
	// the primary-class projection used by the current combat rules layer.
	ClassLevels [8]uint8 `json:"class_levels,omitempty"`
	// AttackAbility 是原作角色記錄的 `+73h`：八個職業槽各查一次命中能力表、
	// 取最大值（TRAINING `overlay-17:0782h`，spec 1140）。**儲存刻度**，
	// 畫面上的 THAC0 是 `60 − 它`。0 代表這個角色沒有帶原作的值。
	AttackAbility uint8 `json:"attack_ability,omitempty"`
	// BaseArmorClass 是原作角色記錄的 `+124h`：派生數值重算把它抄進 `+19Ah`
	// 當護甲的起點（spec 1000）。建角寫 `32h`（＝ AC 10）。同樣是儲存刻度。
	BaseArmorClass uint8 `json:"base_armor_class,omitempty"`
	// AbilityAdjustments 是原作角色記錄的 `+125h`：力量的命中與傷害調整
	// （`overlay-24:1345h`／`13FDh`）在它為 0 時**直接回 0**。建角寫 1；
	// `MON*CHA` 的 81 筆裡 61 筆是 0、20 筆是 1，所以它不是「玩家角色」的同義詞。
	AbilityAdjustments uint8 `json:"ability_adjustments,omitempty"`
	// HitDice and MulticlassLevel preserve DOS offsets 0xE5/0xE6. A
	// dual-class character gains no training HP until the active class has
	// surpassed the old class level.
	HitDice         uint8 `json:"hit_dice,omitempty"`
	MulticlassLevel uint8 `json:"multiclass_level,omitempty"`
	// Age preserves the signed DOS player age at record offset 0x76.
	Age int16 `json:"age,omitempty"`
	// NPC and ControlMorale preserve CMD_AddNPC's party-control distinction.
	// ControlMorale >= 0x80 is the reference NPC namespace.
	NPC           bool  `json:"npc,omitempty"`
	ControlMorale uint8 `json:"control_morale,omitempty"`
	// ECLFlag192 保留 DOS 角色記錄位移 0x192（PC-98 是 0x193）。ECL 透過位址
	// 投影 7CE4h 讀它，而且只取 and 1（spec 1040/1098）。原作唯一的寫入者是
	// 舊格式轉換 overlay-16:01F21h，在來源結構某欄位非 0 時設 1，沒有設 0 的
	// 路徑，所以預設就是 0。語意未定，命名沿用位移。
	ECLFlag192 uint8 `json:"ecl_flag_192,omitempty"`
	// IconSize follows Player.icon_size: 1 is small and 2 is normal.
	// Zero means derive the original default from Race for old save files.
	IconSize        uint8                  `json:"icon_size,omitempty"`
	IconHeadBlock   uint8                  `json:"icon_head,omitempty"`
	IconWeaponBlock uint8                  `json:"icon_weapon,omitempty"`
	IconID          uint8                  `json:"icon_id,omitempty"`
	HitPoints       int                    `json:"hit_points,omitempty"`
	MaxHitPoints    int                    `json:"max_hit_points,omitempty"`
	// BaseMaxHitPoints 是原作的 +12Ch：**不含**體質加值的最大 HP。
	// 屬性重算時 +78h 會先回到這一格再重新加上加值（spec 1079）。
	BaseMaxHitPoints int `json:"base_max_hit_points,omitempty"`
	// DrainedLevels／DrainedHitPoints 是原作的 +0E7h／+0E8h：被吸掉幾級、
	// 連帶少了多少 HP（spec 1125，由復原術的 handler 讀出來）。
	//
	// ★ 兩格要一起看。復原術每次還一級，還的 HP 是 **`少的 HP ÷ 少的級數`**——
	// 只記級數的話還不出正確的 HP，只記 HP 的話不知道要還幾次。
	DrainedLevels    int `json:"drained_levels,omitempty"`
	DrainedHitPoints int `json:"drained_hit_points,omitempty"`
	HealthStatus    HealthStatus           `json:"health_status,omitempty"`
	Bleeding        int                    `json:"bleeding,omitempty"`
	Copper          uint16                 `json:"copper,omitempty"`
	Silver          uint16                 `json:"silver,omitempty"`
	Electrum        uint16                 `json:"electrum,omitempty"`
	Gold            uint16                 `json:"gold,omitempty"`
	Platinum        uint16                 `json:"platinum,omitempty"`
	Gems            uint16                 `json:"gems,omitempty"`
	Jewelry         uint16                 `json:"jewelry,omitempty"`
	Equipment       []monster.ItemRecord   `json:"equipment,omitempty"`
	Effects         []monster.AffectRecord `json:"effects,omitempty"`
	// SpellSlots is the data-neutral ordered spell-slot list used by ECL
	// SPELL resolution. It is optional until original DOS player offsets are
	// decoded; old saves therefore remain valid.
	SpellSlots []uint8 `json:"spell_slots,omitempty"`
	// KnownSpells preserves the imported spell-book flags separately from
	// SpellSlots, which represents spells currently memorized for use.
	KnownSpells []uint8 `json:"known_spells,omitempty"`
	// SpellCastCount stores cleric, druid and magic-user capacities for spell
	// levels 1..5, matching the DOS 3x5 table at 0x12D.
	SpellCastCount [3][5]uint8 `json:"spell_cast_count,omitempty"`
	// SavingThrows preserves the original DOS saveVerse order. The raw
	// five-byte order is retained so ECL DAMAGE adapters can apply the correct
	// saving-throw type.
	SavingThrows []uint8 `json:"saving_throws,omitempty"`
	// SavingThrowBonus preserves DOS player field_186, including signed item or
	// effect-derived bonus already present in an imported record.
	SavingThrowBonus int8 `json:"saving_throw_bonus,omitempty"`
	// Gender 對應角色記錄 +119h（spec 1093 的 Pick Gender）。種族的力量下限／
	// 上限／百分比上限三格會再依性別分（spec 1086 的
	// 「基底 + 種族×10h + 角色^[119h]」），其餘五個屬性只看種族。
	Gender Gender `json:"gender,omitempty"`
	// Alignment mirrors CHARREC.ALIGNMENT at +11B. Zero is a valid alignment,
	// so imported records carry an explicit known bit for conditional effects.
	Alignment      uint8 `json:"alignment,omitempty"`
	AlignmentKnown bool  `json:"alignment_known,omitempty"`
	// ThiefSkills preserves the eight DOS thief percentages; index 1 is
	// open-locks and remains in original order.
	ThiefSkills []uint8 `json:"thief_skills,omitempty"`
}

// HasClass reports whether a character can use rules associated with a class.
// Imported／created multi-class characters consult raw ClassLevels; legacy
// single-class JSON without those slots falls back to Character.Class.
func (c Character) HasClass(class Class) bool {
	if c.ClassLevels != [8]uint8{} {
		index := map[Class]int{
			ClassCleric: 0, ClassFighter: 2, ClassRanger: 4,
			ClassPaladin: 3, ClassMagicUser: 5, ClassThief: 6,
		}[class]
		return c.ClassLevels[index] > 0
	}
	return c.Class == class
}

// ClassLevel returns the raw level for one represented class. Legacy
// single-class characters without ClassLevels use the primary Level field.
func (c Character) ClassLevel(class Class) int {
	if c.ClassLevels != [8]uint8{} {
		index := map[Class]int{
			ClassCleric: 0, ClassFighter: 2, ClassRanger: 4,
			ClassPaladin: 3, ClassMagicUser: 5, ClassThief: 6,
		}[class]
		return int(c.ClassLevels[index])
	}
	if c.Class == class {
		return c.Level
	}
	return 0
}

const (
	DeathDamageFire = 0x01
	DeathDamageAcid = 0x10
)

// DeathEffectContext contains the external state required by the verified
// CheckAffectsEffect(Death) handlers. Damage type is deliberately explicit
// because ECL DAMAGE's five operands do not encode it.
type DeathEffectContext struct {
	DamageFlags       uint8
	DamageFlagsKnown  bool
	CombatHealAllowed bool
	RollDie           func(int) int
}

const MonsterTypeDragon = 3

// DragonSlayerResult is the verified weapon-effect projection used by the
// reference dragon-slayer affect. Target kind and strength damage bonus are
// explicit because neither is encoded by the ECL DAMAGE operands.
type DragonSlayerResult struct {
	Triggered       bool
	Damage          int
	AttackRollBonus int
}

// ResolveDragonSlayer applies the reference 1d12*3+4+strength damage bonus
// and +2 attack roll only when an active dragon-slayer effect targets a dragon.
func (c Character) ResolveDragonSlayer(targetMonsterType uint8, strengthDamageBonus int, rollDie func(int) int) (DragonSlayerResult, error) {
	if rollDie == nil {
		return DragonSlayerResult{}, fmt.Errorf("dragon-slayer requires an injected d12 roller")
	}
	if targetMonsterType != MonsterTypeDragon {
		return DragonSlayerResult{}, nil
	}
	for _, effect := range c.Effects {
		if !effect.Active || effect.Kind != 0x4B {
			continue
		}
		roll := rollDie(12)
		if roll < 1 || roll > 12 {
			return DragonSlayerResult{}, fmt.Errorf("dragon-slayer roll %d is outside 1..12", roll)
		}
		return DragonSlayerResult{Triggered: true, Damage: roll*3 + 4 + strengthDamageBonus, AttackRollBonus: 2}, nil
	}
	return DragonSlayerResult{}, nil
}

// ApplyDeathEffects projects the verified affect_63 and troll_fire_or_acid
// death handlers. Unknown Death effects remain preserved for later adapters.
func (c *Character) ApplyDeathEffects(context DeathEffectContext) error {
	if c == nil {
		return nil
	}
	if context.RollDie == nil {
		return fmt.Errorf("death effects require an injected die roller")
	}
	for index := 0; index < len(c.Effects); index++ {
		effect := c.Effects[index]
		if !effect.Active {
			continue
		}
		switch effect.Kind {
		case 0x63: // affect_63: death-time recovery
			heal := 0
			if c.HealthStatus == HealthStatusDying && c.Bleeding < 6 {
				heal = 6 - c.Bleeding
			} else if c.HealthStatus == HealthStatusUnconscious {
				heal = 6
			}
			if heal > 0 && context.CombatHealAllowed {
				c.HealthStatus = HealthStatusOK
				c.HitPoints = heal
				c.Bleeding = 0
				roll := context.RollDie(4)
				if roll < 1 || roll > 4 {
					return fmt.Errorf("death recovery die roll %d is outside 1..4", roll)
				}
				regeneration := uint16(roll + 1)
				c.Effects = append(c.Effects, monster.AffectRecord{Kind: 0x5F, Value: regeneration, Duration: regeneration, Strength: 0xFF, Active: true})
				c.Effects = append(c.Effects[:index], c.Effects[index+1:]...)
				index--
			}
		case 0x64: // troll_fire_or_acid
			if context.DamageFlagsKnown && context.DamageFlags&(DeathDamageFire|DeathDamageAcid) == 0 {
				total := 0
				for rollIndex := 0; rollIndex < 3; rollIndex++ {
					roll := context.RollDie(6)
					if roll < 1 || roll > 6 {
						return fmt.Errorf("troll regeneration die roll %d is outside 1..6", roll)
					}
					total += roll
				}
				c.Effects = append(c.Effects, monster.AffectRecord{Kind: 0x66, Value: uint16(total), Duration: uint16(total), Strength: 0xFF, Active: true})
			}
		}
	}
	return nil
}

// combatAffectKinds is the verified RemoveCombatAffects table from the DOS
// engine. Unknown and persistent non-combat effects remain untouched.
var combatAffectKinds = map[uint8]struct{}{
	0x07: {}, 0x0B: {}, 0x0D: {}, 0x15: {}, 0x17: {}, 0x1E: {}, 0x1F: {}, 0x20: {},
	0x33: {}, 0x34: {}, 0x35: {}, 0x3A: {}, 0x3B: {}, 0x5F: {}, 0x62: {}, 0x88: {},
	0x89: {}, 0x8B: {}, 0x90: {},
}

// RemoveCombatAffects removes only the effects listed by the reference
// RemoveCombatAffects routine and reports how many records were removed.
func (c *Character) RemoveCombatAffects() int {
	if c == nil || len(c.Effects) == 0 {
		return 0
	}
	kept := make([]monster.AffectRecord, 0, len(c.Effects))
	removed := 0
	for _, effect := range c.Effects {
		if _, ok := combatAffectKinds[effect.Kind]; ok {
			removed++
			continue
		}
		kept = append(kept, effect)
	}
	c.Effects = kept
	return removed
}

// AdvanceEffects consumes imported DOS effect durations without applying
// effect-specific AD&D modifiers. The removed count lets game state surface
// a localized message when a later rules layer wants to do so.
func (c *Character) AdvanceEffects(ticks uint16) int {
	if c == nil || ticks == 0 {
		return 0
	}
	before := len(c.Effects)
	c.Effects = monster.AdvanceAffects(c.Effects, ticks)
	return before - len(c.Effects)
}

// OpenLocksSkill returns the imported DOS open-locks percentage. Zero means
// no verified skill data is available, not an automatic failed roll.
func (c Character) OpenLocksSkill() uint8 {
	if len(c.ThiefSkills) <= 1 {
		return 0
	}
	return c.ThiefSkills[1]
}

type Roster []Character

type SpellMatch struct {
	CharacterIndex int
	SlotIndex      int
}

// FindSpell searches party characters in marching order, then each character's
// ordered spell slots. This matches ECL SPELL's first-match contract while
// keeping the source of spell slots replaceable by a DOS record decoder.
func (r Roster) FindSpell(spellID uint16) (SpellMatch, bool) {
	if spellID > 0xFF {
		return SpellMatch{}, false
	}
	for characterIndex, character := range r {
		for slotIndex, knownSpell := range character.SpellSlots {
			if knownSpell == uint8(spellID) {
				return SpellMatch{CharacterIndex: characterIndex, SlotIndex: slotIndex}, true
			}
		}
	}
	return SpellMatch{}, false
}

// ItemClassBit returns the class usability bit used by the original ITEMS
// table. The values deliberately follow the DOS table rather than the local
// enum order, which starts cleric at zero.
func ItemClassBit(class Class) (uint8, bool) {
	switch class {
	case ClassMagicUser:
		return 1, true
	case ClassCleric:
		return 2, true
	case ClassThief:
		return 4, true
	case ClassFighter:
		return 8, true
	case ClassPaladin:
		return 64, true
	case ClassRanger:
		return 128, true
	default:
		return 0, false
	}
}

// CanEquip validates one inventory index against the reference class mask
// and the currently readied equipment slots. It does not mutate the party.
func (c Character) CanEquip(index int, catalog monster.BaseItemCatalog) error {
	if index < 0 || index >= len(c.Equipment) {
		return fmt.Errorf("equipment index %d is out of range", index)
	}
	item := c.Equipment[index]
	base, ok := catalog.Lookup(item.Type)
	if !ok {
		return fmt.Errorf("item type 0x%02X is outside base catalog", item.Type)
	}
	usable := false
	for _, class := range []Class{ClassCleric, ClassFighter, ClassRanger, ClassPaladin, ClassMagicUser, ClassThief} {
		classBit, ok := ItemClassBit(class)
		if ok && c.HasClass(class) && base.UsableByMask(classBit) {
			usable = true
			break
		}
	}
	if !usable {
		return fmt.Errorf("%s cannot equip item type 0x%02X", c.Class, item.Type)
	}
	if item.Readied {
		return nil
	}
	rings := 0
	for otherIndex, other := range c.Equipment {
		if otherIndex == index || !other.Readied {
			continue
		}
		otherBase, otherOK := catalog.Lookup(other.Type)
		if !otherOK {
			return fmt.Errorf("readied item type 0x%02X is outside base catalog", other.Type)
		}
		if base.Slot == 9 && otherBase.Slot == 9 {
			rings++
			continue
		}
		if base.Slot <= 9 && base.Slot == otherBase.Slot {
			return fmt.Errorf("equipment slot %d is already occupied", base.Slot)
		}
		if base.Slot == 0 && base.HandsRequired >= 2 && otherBase.Slot == 1 {
			return fmt.Errorf("two-handed item conflicts with off-hand equipment")
		}
		if base.Slot == 1 && otherBase.Slot == 0 && otherBase.HandsRequired >= 2 {
			return fmt.Errorf("off-hand equipment conflicts with two-handed weapon")
		}
	}
	if base.Slot == 9 && rings >= 2 {
		return fmt.Errorf("at most two rings may be readied")
	}
	return nil
}

// EquipItem marks an inventory item readied after CanEquip validation.
func (c *Character) EquipItem(index int, catalog monster.BaseItemCatalog) error {
	if c == nil {
		return fmt.Errorf("cannot equip item on nil character")
	}
	if err := c.CanEquip(index, catalog); err != nil {
		return err
	}
	c.Equipment[index].Readied = true
	return nil
}

// UnequipItem clears the readied state without changing the inventory record.
func (c *Character) UnequipItem(index int) error {
	if c == nil {
		return fmt.Errorf("cannot unequip item on nil character")
	}
	if index < 0 || index >= len(c.Equipment) {
		return fmt.Errorf("equipment index %d is out of range", index)
	}
	if c.Equipment[index].Readied && c.Equipment[index].Cursed {
		return fmt.Errorf("cursed item type 0x%02X cannot be unequipped", c.Equipment[index].Type)
	}
	c.Equipment[index].Readied = false
	return nil
}

// RemoveItem removes one inventory unit. Count zero is the DOS encoding for
// a non-stacking single item; a positive count is decremented in place. A
// readied item must be unequipped first so shop/treasure mutations cannot
// silently invalidate the equipment projection.
func (c *Character) RemoveItem(index int) (monster.ItemRecord, error) {
	if c == nil {
		return monster.ItemRecord{}, fmt.Errorf("cannot remove item from nil character")
	}
	if index < 0 || index >= len(c.Equipment) {
		return monster.ItemRecord{}, fmt.Errorf("equipment index %d is out of range", index)
	}
	item := c.Equipment[index]
	if item.Readied {
		return monster.ItemRecord{}, fmt.Errorf("readied item type 0x%02X must be unequipped first", item.Type)
	}
	removed := item
	removed.Count = 1
	if item.Count > 1 {
		c.Equipment[index].Count--
		return removed, nil
	}
	c.Equipment = append(c.Equipment[:index], c.Equipment[index+1:]...)
	return removed, nil
}

// DestroyItemType applies an explicit ECL DESTROY ITEMS effect. Unlike the
// player-facing RemoveItem operation, this effect may remove readied items:
// the original event can destroy equipment while it is worn. It removes all
// matching records and returns the number of item units destroyed.
func (c *Character) DestroyItemType(itemType uint8) int {
	if c == nil {
		return 0
	}
	updated := make([]monster.ItemRecord, 0, len(c.Equipment))
	destroyed := 0
	for _, item := range c.Equipment {
		if item.Type != itemType {
			updated = append(updated, item)
			continue
		}
		count := int(item.Count)
		if count == 0 {
			count = 1
		}
		destroyed += count
	}
	c.Equipment = updated
	return destroyed
}

// ConsumeAmmunition removes shots for a weapon's raw ITEMS AmmunitionType.
// The raw code and inventory item type use different namespaces in CoAB, so
// the caller must inject the verified mapping for the current game's data.
// The mutation is atomic: insufficient mapped ammunition leaves equipment
// unchanged.
func (c *Character) ConsumeAmmunition(ammunitionType uint8, shots int, itemTypes map[uint8][]uint8) error {
	if c == nil {
		return fmt.Errorf("cannot consume ammunition from nil character")
	}
	if ammunitionType == 0 {
		return fmt.Errorf("ammunition type is not required")
	}
	if shots <= 0 {
		return fmt.Errorf("ammunition shots must be positive")
	}
	candidates, ok := itemTypes[ammunitionType]
	if !ok || len(candidates) == 0 {
		return fmt.Errorf("no inventory mapping for ammunition type %d", ammunitionType)
	}
	allowed := make(map[uint8]bool, len(candidates))
	for _, itemType := range candidates {
		allowed[itemType] = true
	}
	available := 0
	for _, item := range c.Equipment {
		if !allowed[item.Type] {
			continue
		}
		if item.Count == 0 {
			available++
		} else {
			available += int(item.Count)
		}
	}
	if available < shots {
		return fmt.Errorf("ammunition type %d has %d shots; need %d", ammunitionType, available, shots)
	}
	remaining := shots
	updated := make([]monster.ItemRecord, 0, len(c.Equipment))
	for _, item := range c.Equipment {
		if !allowed[item.Type] || remaining == 0 {
			updated = append(updated, item)
			continue
		}
		count := 1
		if item.Count > 0 {
			count = int(item.Count)
		}
		consume := count
		if consume > remaining {
			consume = remaining
		}
		remaining -= consume
		if item.Count > 0 {
			item.Count -= uint8(consume)
			if item.Count > 0 {
				updated = append(updated, item)
			}
		}
	}
	c.Equipment = updated
	return nil
}

// UseConsumable consumes one scroll/potion or one charge from a wand. The
// returned signal is intentionally renderer- and rules-neutral; a later
// spell/effect system decides what EffectID or SpellIDs do.
func (c *Character) UseConsumable(index int, catalog monster.BaseItemCatalog) (monster.ConsumableUse, error) {
	if c == nil {
		return monster.ConsumableUse{}, fmt.Errorf("cannot use item on nil character")
	}
	if index < 0 || index >= len(c.Equipment) {
		return monster.ConsumableUse{}, fmt.Errorf("equipment index %d is out of range", index)
	}
	use, err := c.Equipment[index].DecodeConsumable(catalog)
	if err != nil {
		return monster.ConsumableUse{}, err
	}
	if use.Kind == monster.ConsumableCharged {
		if use.ChargesBefore <= 0 {
			return monster.ConsumableUse{}, fmt.Errorf("item type 0x%02X has no charges", c.Equipment[index].Type)
		}
		c.Equipment[index].Affects[0]--
		use.ChargesAfter = int(c.Equipment[index].Affects[0])
		return use, nil
	}
	if _, err := c.RemoveItem(index); err != nil {
		return monster.ConsumableUse{}, err
	}
	return use, nil
}

func (r Roster) Validate() error {
	if len(r) == 0 || len(r) > 6 {
		return fmt.Errorf("roster must contain 1..6 player characters")
	}
	seen := make(map[string]struct{}, len(r))
	for _, character := range r {
		if err := character.Validate(); err != nil {
			return err
		}
		if _, exists := seen[character.ID]; exists {
			return fmt.Errorf("duplicate character ID %q", character.ID)
		}
		seen[character.ID] = struct{}{}
	}
	return nil
}

// Fighter returns the starter combat projection used by the character
// creation slice. The full AD&D level, equipment and spell tables remain a
// separate data source; these defaults are intentionally deterministic.
func (c Character) Fighter() (combat.Fighter, error) {
	if err := c.Validate(); err != nil {
		return combat.Fighter{}, err
	}
	hitDie, damageSides := 6, 4
	specialist := c.Class == ClassFighter || c.Class == ClassRanger || c.Class == ClassPaladin
	switch c.Class {
	case ClassCleric:
		hitDie, damageSides = 8, 6
	case ClassFighter, ClassRanger, ClassPaladin:
		hitDie, damageSides = 10, 8
	case ClassThief:
		hitDie, damageSides = 6, 6
	}
	constitutionBonus := (c.Abilities.Constitution - 14) / 2
	hitPoints := hitDie + constitutionBonus
	if hitPoints < 1 {
		hitPoints = 1
	}
	// 原作的命中能力在 `+73h`、護甲起點在 `+124h`（spec 1140），兩個都是
	// 儲存刻度。有帶原作的值就用它換算；沒有的（早期存檔、合成角色）
	// 才退回本作自己的近似式。
	attackBonus := (c.Abilities.Strength - 10) / 2
	if !specialist {
		attackBonus = (c.Abilities.Dexterity - 10) / 2
	}
	if c.AttackAbility > 0 {
		attackBonus = combat.DisplayAttackBonus(int(c.AttackAbility))
		// 原作的派生數值重算在**第 0 個裝備槽是空的**時候才加力量調整
		// （spec 1000 §四）。這一支不帶裝備，所以那一槽必然是空的。
		if hit, _, ok := c.StrengthAdjustments(); ok {
			attackBonus += hit
		}
	}
	damageBonus := 0
	if c.AttackAbility > 0 {
		if _, damage, ok := c.StrengthAdjustments(); ok {
			damageBonus = damage
		}
	}
	armorClass := 10 - (c.Abilities.Dexterity-10)/2
	if c.BaseArmorClass > 0 {
		armorClass = combat.DisplayArmorClass(int(c.BaseArmorClass) +
			combat.DexterityDefenceAdjustment(c.Abilities.Dexterity))
	}
	iconSize := c.IconSize
	if iconSize == 0 {
		iconSize = DefaultIconSize(c.Race)
	}
	hitPoints, maxHitPoints := hitPoints, hitPoints
	if c.MaxHitPoints > 0 {
		maxHitPoints = c.MaxHitPoints
		hitPoints = c.HitPoints
	}
	if c.HitPoints > 0 && c.HitPoints <= maxHitPoints {
		hitPoints = c.HitPoints
	}
	headBlock, bodyBlock := c.CombatIconBlocks()
	fighter := combat.Fighter{
		ID: c.ID, Name: c.Name, Side: combat.SideParty,
		ControlMorale: c.ControlMorale,
		Alignment:     c.Alignment, AlignmentKnown: c.AlignmentKnown,
		HasPartyIcon: true, PartyHeadBlock: headBlock, PartyBodyBlock: bodyBlock, PartyIconID: c.IconID, PartyIconSize: iconSize,
		HitPoints: hitPoints, MaxHitPoints: maxHitPoints, ArmorClass: armorClass,
		DamageBonus: damageBonus,
		HitDice:     c.HitDice,
		Dexterity:   uint8(c.Abilities.Dexterity),
		AttackBonus: attackBonus, DamageDiceCount: 1, DamageDiceSides: damageSides,
		SavingThrows: append([]uint8(nil), c.SavingThrows...), SavingThrowBonus: int(c.SavingThrowBonus),
		ClericLevel:  c.turnUndeadLevel(),
	}
	return c.applyKnownEffects(fighter), nil
}

// applyKnownEffects projects only unconditional attack and AC modifiers documented
// for the current CoAB rules slice. Effects requiring target alignment,
// morale, saving throws, or status transitions remain in the rules layer.
func (c Character) applyKnownEffects(fighter combat.Fighter) combat.Fighter {
	for _, effect := range c.Effects {
		if !effect.Active {
			continue
		}
		switch effect.Kind {
		case 0x01: // Bless: -1 THAC0, equivalent to +1 attack bonus.
			fighter.AttackBonus++
		case 0x02: // Curse: +1 THAC0, equivalent to -1 attack bonus.
			fighter.AttackBonus--
		case 0x21: // Blind: -4 attack and +4 AC (worse AC).
			fighter.AttackBonus -= 4
			fighter.ArmorClass += 4
		case 0x24: // Bestow Curse: -4 attack.
			fighter.AttackBonus -= 4
		case 0x31: // Prayer: friendly effect gives -1 THAC0, or +1 attack.
			fighter.AttackBonus++
		}
	}
	return fighter
}

// FighterWithEquipment applies readied base weapon/armor effects from a
// decoded ITEMS catalog. It is an additive adapter for remake party JSON;
// magical effect stacks and the original DOS inventory record remain outside
// this boundary.
func (c Character) FighterWithEquipment(catalog monster.BaseItemCatalog) (combat.Fighter, error) {
	fighter, err := c.Fighter()
	if err != nil {
		return combat.Fighter{}, err
	}
	hasWeapon := false
	for _, item := range c.Equipment {
		if !item.Readied {
			continue
		}
		// AI 用道具只看第二個效果槽（spec 835 的過濾：裝備中、`+3Dh > 0`、
		// `+3Eh < 80h`）。這裡把符合條件的效果碼投影出來，戰鬥層才不必認識
		// 整條物品鏈。⚠ `Affects` 的三格依序是 `+3Ch`／`+3Dh`／`+3Eh`。
		if item.Affects[1] > 0 && item.Affects[2] < 0x80 {
			fighter.ReadiedItemEffects = append(fighter.ReadiedItemEffects, item.Affects[1])
		}
		effect, effectErr := item.Effect(catalog, false)
		if effectErr != nil {
			return combat.Fighter{}, effectErr
		}
		fighter.ArmorClass -= effect.ArmorClassImprovement
		if effect.MovementAllowance > 0 && (fighter.MovementAllowance == 0 || effect.MovementAllowance < fighter.MovementAllowance) {
			fighter.MovementAllowance = effect.MovementAllowance
		}
		if effect.DamageDiceCount > 0 && !hasWeapon {
			fighter.AttackBonus += effect.AttackBonus
			fighter.DamageDiceCount = effect.DamageDiceCount
			fighter.DamageDiceSides = effect.DamageDiceSides
			fighter.DamageBonus = effect.DamageBonus
			fighter.AttacksPerTurn = effect.AttacksPerTurn
			fighter.AmmunitionType = effect.AmmunitionType
			fighter.WeaponRange = effect.WeaponRange
			fighter.MissileWeapon = effect.MissileWeapon
			fighter.ThrownWeapon = effect.ThrownWeapon
			hasWeapon = true
		}
	}
	return fighter, nil
}

// DefaultIconSize matches the original character creation race switch.
func DefaultIconSize(r Race) uint8 {
	switch r {
	case RaceDwarf, RaceGnome, RaceHalfling:
		return 1
	default:
		return 2
	}
}

// CombatIconBlocks maps raw DOS head_icon/weapon_icon slots to the actual
// CHEAD/CBODY blocks selected by LoadPlayerCombatIcon. Small icons use the T
// files, whose namespace is the raw slot plus 0x40.
func (c Character) CombatIconBlocks() (head, body uint8) {
	return c.CombatIconBlocksFor(false)
}

// CombatIconBlocksFor adds the reference LoadIcons attack namespace when the
// combat icon is in its attack state. The DAX loader uses normal_id+0x80.
func (c Character) CombatIconBlocksFor(attack bool) (head, body uint8) {
	head, body = c.IconHeadBlock, c.IconWeaponBlock
	size := c.IconSize
	if size == 0 {
		size = DefaultIconSize(c.Race)
	}
	if size == 1 {
		if head < 0x40 {
			head += 0x40
		}
		if body < 0x40 {
			body += 0x40
		}
	}
	if attack {
		head += 0x80
		body += 0x80
	}
	return head, body
}

func (r Race) String() string {
	names := [...]string{"dwarf", "elf", "gnome", "half-elf", "halfling", "human", "half-orc", "saurial"}
	if int(r) >= len(names) {
		return fmt.Sprintf("race-%d", r)
	}
	return names[r]
}

func (c Class) String() string {
	return [...]string{"cleric", "fighter", "ranger", "paladin", "magic-user", "thief"}[c]
}

func (c Character) Validate() error {
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("character ID and name are required")
	}
	if c.Race > RaceSaurial {
		return fmt.Errorf("unsupported race %d", c.Race)
	}
	if c.Class > ClassThief {
		return fmt.Errorf("unsupported class %d", c.Class)
	}
	if c.Level < 1 {
		return fmt.Errorf("character level must be positive")
	}
	values := []int{c.Abilities.Strength, c.Abilities.Intelligence, c.Abilities.Wisdom, c.Abilities.Dexterity, c.Abilities.Constitution, c.Abilities.Charisma}
	for index, value := range values {
		if value < 3 || value > 18 {
			return fmt.Errorf("ability %d=%d is outside 3..18", index, value)
		}
	}
	if !raceAllowsClass(c.Race, c.Class) {
		return fmt.Errorf("%s cannot be a %s", c.Race, c.Class)
	}
	// Original NPC records are authoritative campaign data, not character
	// creation requests. Their scripted race/class and ability combinations
	// may sit outside player-creation minimums (Dragonbait is the canonical
	// saurial paladin example), so retain them after structural validation.
	if c.NPC {
		return nil
	}
	minimums := []struct {
		ability string
		value   int
	}{
		{"wisdom", 9}, {"strength", 9}, {"strength", 13}, {"strength", 12}, {"intelligence", 9}, {"dexterity", 9},
	}
	minimum := minimums[c.Class]
	if abilityValue(c.Abilities, minimum.ability) < minimum.value {
		return fmt.Errorf("%s requires %s >= %d", c.Class, minimum.ability, minimum.value)
	}
	if c.Class == ClassRanger && (c.Abilities.Intelligence < 13 || c.Abilities.Wisdom < 14 || c.Abilities.Constitution < 14) {
		return fmt.Errorf("ranger requires intelligence >= 13, wisdom >= 14, constitution >= 14")
	}
	if c.Class == ClassPaladin && (c.Abilities.Wisdom < 13 || c.Abilities.Charisma < 17) {
		return fmt.Errorf("paladin requires wisdom >= 13 and charisma >= 17")
	}
	return nil
}

func raceAllowsClass(race Race, class Class) bool {
	switch race {
	case RaceHuman, RaceHalfElf:
		return true
	case RaceDwarf, RaceGnome, RaceHalfling:
		return class == ClassFighter || class == ClassThief
	case RaceElf:
		return class == ClassFighter || class == ClassMagicUser || class == ClassThief
	case RaceHalfOrc:
		return class == ClassCleric || class == ClassFighter || class == ClassThief
	case RaceSaurial:
		return class == ClassPaladin
	default:
		return false
	}
}

func abilityValue(abilities Abilities, name string) int {
	switch name {
	case "strength":
		return abilities.Strength
	case "intelligence":
		return abilities.Intelligence
	case "wisdom":
		return abilities.Wisdom
	case "dexterity":
		return abilities.Dexterity
	default:
		return 0
	}
}

// CombatBaseLookup 是命中能力表的查詢介面。資料由 game pack 提供
// （gamepack.CombatBase），本套件只放機制。
type CombatBaseLookup interface {
	ClassAttackValue(slot, level int) (int, bool)
	CreationDefaults() (armorClassStored, attackAbilityStored int, ok bool)
}

// AttackAbilityFrom 是原作 `+73h` 的算法（TRAINING `overlay-17:0782h`，spec 1140）：
// 八個職業槽**全部**跑一遍，等級為 0 的槽跳過，其餘查表取最大值。
//
// ⚠ 原作那條 `jle` 跳到的是迴圈**尾巴**（`cmp var_19, 7`）不是出口，
// 所以空槽是 continue 不是 break——聖騎士只有槽 3 有等級，
// 讀成 break 會在槽 0 就結束、算不出任何值。
func AttackAbilityFrom(tables CombatBaseLookup, classLevels [8]uint8) (uint8, error) {
	if tables == nil {
		return 0, fmt.Errorf("命中能力表沒有注入")
	}
	best := 0
	for slot, level := range classLevels {
		if level == 0 {
			continue
		}
		value, ok := tables.ClassAttackValue(slot, int(level))
		if !ok {
			return 0, fmt.Errorf("命中能力表查不到職業槽 %d 等級 %d", slot, level)
		}
		if value > best {
			best = value
		}
	}
	if best == 0 {
		return 0, fmt.Errorf("這個角色八個職業槽都是 0 級")
	}
	return uint8(best), nil
}

// CreationCombatBase 回傳建角要寫進 `+73h` 與 `+124h` 的兩個值。
// 命中能力走 AttackAbilityFrom（依職業與等級），護甲起點是建角的固定預設值。
func CreationCombatBase(tables CombatBaseLookup, classLevels [8]uint8) (attackAbility, baseArmorClass uint8, err error) {
	attack, err := AttackAbilityFrom(tables, classLevels)
	if err != nil {
		return 0, 0, err
	}
	armorStored, _, ok := tables.CreationDefaults()
	if !ok {
		return 0, 0, fmt.Errorf("game pack 沒有建角預設值")
	}
	return attack, uint8(armorStored), nil
}

// StrengthAdjustments 是原作的力量命中／傷害調整（`overlay-24:1345h`／`13FDh`，
// spec 694／697）。`+125h` 為 0 時原作直接回 0，所以這裡回 `ok = false`。
//
// ⚠ 力量索引查不到（力量超出 0..25、或 18 但百分比 > 100）時原作回的是堆疊
// 殘值——這裡一律回 `ok = false`，不猜一個看起來合理的值。
func (c Character) StrengthAdjustments() (hit, damage int, ok bool) {
	if c.AbilityAdjustments == 0 {
		return 0, 0, false
	}
	strength := c.Abilities.StrengthFull
	if strength == 0 {
		strength = c.Abilities.Strength
	}
	index, found := engineability.StrengthIndex(strength, c.Abilities.StrengthExceptional)
	if !found {
		return 0, 0, false
	}
	return engineability.StrengthHitAdjustment(index), engineability.StrengthDamageAdjustment(index), true
}

// turnUndeadLevel 是驅散不死用的「牧師槽等級」（角色記錄 `+109h`）。
//
// ★ 原作的顯示條件是 `+109h > 0`（牧師槽的等級）或第二職業成立（spec 905），
// 也就是**只有牧師會看到那個選項**。聖騎士在 AD&D 一版是三級起以「等級 − 2」
// 驅散，但本作的判斷是牧師槽——聖騎士的牧師槽在原版存檔裡是 0，所以這裡
// 只認牧師，多的規則要有證據再加。
func (c Character) turnUndeadLevel() uint8 {
	if c.Class != ClassCleric {
		return 0
	}
	if c.Level < 1 {
		return 0
	}
	if c.Level > 255 {
		return 255
	}
	return uint8(c.Level)
}
