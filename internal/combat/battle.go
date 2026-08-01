// Package combat contains the platform-neutral AD&D combat core. ECL and
// Ebiten adapters can provide party/enemy data without embedding rendering or
// DOS memory assumptions here.
package combat

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	engineaction "github.com/wicanr2/golden-box-remake-engine/combat/action"
	engineinitiative "github.com/wicanr2/golden-box-remake-engine/combat/initiative"
	enginequickspell "github.com/wicanr2/golden-box-remake-engine/combat/quickspell"
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
)

type Fighter struct {
	ID   string
	Name string
	Side Side
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
	HasPartyIcon   bool
	PartyHeadBlock uint8
	PartyBodyBlock uint8
	PartyIconID    uint8
	PartyIconSize  uint8
	IconDirection  uint8
	IconAttack     bool
	// MonsterAffects preserves raw MON*SPC records. Gameplay projections are
	// intentionally left to later, verified rules adapters.
	MonsterAffects []MonsterAffect
	// CombatMap position/size. A future Area/ECL placement decoder can set
	// these directly; StartCombat supplies a deterministic fallback otherwise.
	HasCombatPosition bool
	CombatX           int
	CombatY           int
	CombatSize        uint8
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
	CombatAction ActionState
	HitPoints    int
	MaxHitPoints int
	// HitDice preserves the original Player/MON*CHA byte at offset 0xE5.
	// Poisonous-cloud rules consume it directly.
	HitDice uint8
	// MonsterType preserves the original shared Player record byte at +11A.
	// Effect handlers compare this field against typed creature categories.
	MonsterType uint8
	// Dexterity preserves the shared Player byte at +17 used by the original
	// initiative reaction table. It is deliberately not converted to a modern
	// ability modifier.
	Dexterity uint8
	// CombatTeam preserves the Action scheduler team number. The current CoAB
	// adapter uses Side while the area surprise-mask writer remains unresolved.
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
	AttacksPerTurn       int
	AmmunitionType       uint8
	MovementAllowance    int
	WeaponRange          int
	MissileWeapon        bool
	ThrownWeapon         bool
	// InitiativeBonus is retained only as a legacy synthetic-fixture ordering
	// seam. Production party/MON adapters never set it. Nonzero fixture values
	// replace the rolled delay after the exact d6 draw, while d100 scan traffic
	// remains unchanged; new tests should assert DEX and RNG directly instead.
	InitiativeBonus int
	// SavingThrows preserves the five reference saveVerse thresholds:
	// poison, petrification, rod/staff/wand, breath weapon and spell.
	SavingThrows     []uint8
	SavingThrowBonus int
	// CoughingTurns and HelplessTurns are action-counted combat effects.
	// Persistent-area rules set them; the game adapter consumes one turn only
	// when the affected combatant actually reaches its initiative.
	CoughingTurns int
	HelplessTurns int
	// MonsterSpellIDs mirrors the raw MON*CHA spell-list slots. The bounded
	// monster-turn adapter currently consumes only Magic Missile (0x0F).
	MonsterSpellIDs  []uint8
	MonsterSpellUses [3]uint8
}

// MonsterAffect mirrors one nine-byte MON*SPC record without importing the
// monster data package into the combat core.
type MonsterAffect struct {
	Kind     uint8
	Value    uint16
	Duration uint16
	Strength uint8
	Active   bool
	// Innate marks an effect loaded from a MON*SPC monster template. The
	// reference LOADMONSTER preserves byte 4 as zero while retaining every
	// template effect in the runtime list, so that byte cannot be used to
	// suppress a monster's innate effects in the combat projection.
	Innate bool
	Data   [4]byte
}

func (a MonsterAffect) operational() bool { return a.Active || a.Innate }

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

// MonsterMagicResistanceBase reports the evidence-backed percentage base
// supplied by a monster affect handler. Effect 6A maps exactly to the
// 15-percent wrapper; the shared formula is also exact.
func (f Fighter) MonsterMagicResistanceBase() (int, bool) {
	for _, affect := range f.MonsterAffects {
		if affect.operational() && affect.Kind == 0x6A {
			return 15, true
		}
	}
	return 0, false
}

const (
	DamageFlagFire        uint8 = 0x01
	DamageFlagElectricity uint8 = 0x04
	DamageFlagMagic       uint8 = 0x08
)

// MonsterProtectedFromDamage reports the two exact elemental protection
// handlers resolved from the PC-98 EFFPROCS table. Effect 70 consumes Fire;
// effect 87 consumes Electricity. Magic resistance remains a separate,
// probabilistic pre-damage handler.
func (f Fighter) MonsterProtectedFromDamage(flags uint8) bool {
	for _, affect := range f.MonsterAffects {
		if !affect.operational() {
			continue
		}
		if affect.Kind == 0x70 && flags&DamageFlagFire != 0 {
			return true
		}
		if affect.Kind == 0x87 && flags&DamageFlagElectricity != 0 {
			return true
		}
	}
	return false
}

// MonsterPostHitAffects returns the operational monster affects dispatched by
// the reference CHECKFX table after a successful physical attack slot. The
// PC-98 caller passes attack slot + 1; CHECKFX types 2 and 3 both dispatch
// effect 4F, while later types do not. Raw affects remain attached to Fighter
// and this projection adds behavior without renaming or rewriting them.
func (f Fighter) MonsterPostHitAffects(attackSlot int) []MonsterAffect {
	if attackSlot < 1 || attackSlot > 2 {
		return nil
	}
	var effects []MonsterAffect
	for _, affect := range f.MonsterAffects {
		if affect.operational() && affect.Kind == 0x4F {
			effects = append(effects, affect)
		}
	}
	return effects
}

// MonsterThrowsLightning reports the operational MON*SPC effect dispatched
// by the reference CHECKFX type-14 action phase. The raw effect remains on
// Fighter; this method only exposes the title-neutral runtime capability.
func (f Fighter) MonsterThrowsLightning() bool {
	for _, affect := range f.MonsterAffects {
		if affect.operational() && affect.Kind == 0x84 {
			return true
		}
	}
	return false
}

// MagicResistanceChance mirrors the PC-98 EFFPROCS common routine. The
// original compares a d100 roll directly against this signed expression and
// does not clamp it before the comparison.
func MagicResistanceChance(base, casterLevel int) int {
	return base + (11-casterLevel)*5
}

const MonsterTypeAnimal uint8 = 0x13

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
			if observer.MonsterType == MonsterTypeAnimal && !observer.MonsterCanDetectInvisible() {
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
			if attacker.MonsterType == MonsterTypeAnimal {
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

// MonsterAffectAttacksPerTurn projects the verified Haste multiplier over
// the decoded MON*CHA attacksCount field.
func (f Fighter) MonsterAffectAttacksPerTurn() int {
	attacks := f.AttacksPerTurn
	if attacks < 1 {
		attacks = 1
	}
	for _, affect := range f.MonsterAffects {
		if !affect.operational() {
			continue
		}
		switch affect.Kind {
		case 0x27: // haste: AffectHaste doubles half-actions
			attacks *= 2
		case 0x2A: // slow: AffectSlow halves half-actions
			attacks /= 2
		}
	}
	if attacks < 1 {
		attacks = 1
	}
	return attacks
}

// MonsterIsHeld mirrors the reference Player.IsHeld affect set.
func (f Fighter) MonsterIsHeld() bool {
	for _, affect := range f.MonsterAffects {
		if affect.operational() && (affect.Kind == 0x1F || affect.Kind == 0x33 || affect.Kind == 0x34 || affect.Kind == 0x35) {
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
	Protected bool
}

type AreaSpellResult struct {
	CasterID   string
	SpellID    uint8
	Center     TilePoint
	BaseDamage int
	Impacts    []AreaSpellImpact
}

type Battle struct {
	fighters            map[string]Fighter
	fighterOrder        []string
	attackRollModifier  map[Side]int
	rng                 *rand.Rand
	round               int
	status              Status
	areas               []PersistentArea
	nextArea            uint64
	initiativeScheduler *engineinitiative.Scheduler
	initiativeSelection engineinitiative.Selection
	initiativeSelected  bool
}

const FireballSpellID uint8 = 0x2F

func NewBattle(fighters []Fighter, seed int64) (*Battle, error) {
	if len(fighters) == 0 {
		return nil, fmt.Errorf("battle needs at least one fighter")
	}
	b := &Battle{
		fighters:           make(map[string]Fighter, len(fighters)),
		fighterOrder:       make([]string, 0, len(fighters)),
		attackRollModifier: make(map[Side]int, 2),
		rng:                rand.New(rand.NewSource(seed)),
		status:             StatusActive,
	}
	for _, fighter := range fighters {
		if fighter.ID == "" {
			return nil, fmt.Errorf("fighter has empty ID")
		}
		if _, exists := b.fighters[fighter.ID]; exists {
			return nil, fmt.Errorf("duplicate fighter ID %q", fighter.ID)
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
	fighter.CombatAction.Clear()
	b.fighters[fighterID] = fighter
	if b.initiativeScheduler != nil && b.initiativeSelected && b.initiativeSelection.ID == fighterID {
		if !b.initiativeScheduler.Complete(b.initiativeSelection) {
			return fmt.Errorf("stale scheduled action for fighter %q", fighterID)
		}
		b.initiativeSelected = false
	}
	return nil
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

// SetQuickFight gives one combatant to the AI. The original also clears a
// same-team per-action target; that target pointer is not yet represented by
// the typed ActionState and remains an explicit adapter gap.
func (b *Battle) SetQuickFight(fighterID string) error {
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
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
		if err := b.SetQuickFight(fighterID); err != nil {
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

// BeginPendingSpellAction mirrors CASTCOMBATSPELL's nonzero casting-delay
// handoff. The same action remains in this round at max(1, delay-units).
func (b *Battle) BeginPendingSpellAction(fighterID string, spellID uint8, castingDelay int) error {
	if b == nil || b.initiativeScheduler == nil || !b.initiativeSelected ||
		b.initiativeSelection.ID != fighterID {
		return fmt.Errorf("fighter %q is not the selected scheduled action", fighterID)
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return fmt.Errorf("unknown fighter %q", fighterID)
	}
	if !fighter.CombatAction.BeginSpell(spellID, castingDelay) {
		return fmt.Errorf("spell 0x%02X has invalid casting delay %d", spellID, castingDelay)
	}
	if !b.initiativeScheduler.SetDelay(fighterID, fighter.CombatAction.Delay) {
		return fmt.Errorf("unknown scheduled fighter %q", fighterID)
	}
	b.fighters[fighterID] = fighter
	b.initiativeSelected = false
	return nil
}

// TakePendingSpellAction clears the spell byte when its delayed action is
// selected again. Delay completion remains owned by CompleteAction.
func (b *Battle) TakePendingSpellAction(fighterID string) (uint8, error) {
	if b == nil || !b.initiativeSelected || b.initiativeSelection.ID != fighterID {
		return 0, fmt.Errorf("fighter %q is not the selected scheduled action", fighterID)
	}
	fighter, ok := b.fighters[fighterID]
	if !ok {
		return 0, fmt.Errorf("unknown fighter %q", fighterID)
	}
	spellID := fighter.CombatAction.TakeSpell()
	if spellID == 0 {
		return 0, fmt.Errorf("fighter %q has no pending spell", fighterID)
	}
	b.fighters[fighterID] = fighter
	return spellID, nil
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
// raw combat effects. Strength 255 is the permanent marker.
func (b *Battle) AdvanceMonsterAffects(minutes uint16) int {
	if minutes == 0 {
		return 0
	}
	removed := 0
	for id, fighter := range b.fighters {
		if len(fighter.MonsterAffects) == 0 {
			continue
		}
		kept := make([]MonsterAffect, 0, len(fighter.MonsterAffects))
		for _, affect := range fighter.MonsterAffects {
			if affect.Strength != 0xFF {
				duration := affect.Duration
				if duration == 0 {
					duration = affect.Value
				}
				if duration <= minutes {
					removed++
					continue
				}
				affect.Duration = duration - minutes
				affect.Value = affect.Duration
			}
			kept = append(kept, affect)
		}
		fighter.MonsterAffects = kept
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
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
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
	targetArmorClass := target.ArmorClass
	targetArmorClass += target.MonsterAffectArmorClassBonusAgainst(attacker)
	if attacker.Evil && target.ProtectedFromEvil {
		targetArmorClass += 2
	}
	if attacker.Good && target.ProtectedFromGood {
		targetArmorClass += 2
	}
	attackTotal := attackRoll + attacker.AttackBonus + b.attackRollModifier[attacker.Side]
	hit := !forcedMiss && (target.MonsterIsHeld() || critical || (attackRoll != 1 && attackTotal >= targetArmorClass))
	damage := 0
	if hit {
		damage = damageRoll + attacker.DamageBonus
		if damage < 0 {
			damage = 0
		}
		if damage > target.HitPoints {
			damage = target.HitPoints
		}
		target.HitPoints -= damage
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

func (b *Battle) attackSlot(attackerID, targetID string, attackSlot int) (AttackResult, error) {
	if err := b.ValidateAttack(attackerID, targetID); err != nil {
		return AttackResult{}, err
	}
	attacker := b.fighters[attackerID]
	var result AttackResult
	var err error
	if attacker.DamageDiceCount < 1 || attacker.DamageDiceSides < 1 {
		result, err = b.ResolveAttack(attackerID, targetID, b.rng.Intn(20)+1, 0)
	} else {
		attackRoll := b.rng.Intn(20) + 1
		damageRoll := 0
		for i := 0; i < attacker.DamageDiceCount; i++ {
			damageRoll += b.rng.Intn(attacker.DamageDiceSides) + 1
		}
		result, err = b.ResolveAttack(attackerID, targetID, attackRoll, damageRoll)
	}
	if err != nil || !result.Hit || result.TargetHP <= 0 {
		return result, err
	}
	for _, affect := range attacker.MonsterPostHitAffects(attackSlot) {
		target := b.fighters[targetID]
		if target.HitPoints <= 0 {
			break
		}
		flags := DamageFlagFire | DamageFlagMagic
		rolledDamage := b.rng.Intn(10) + 1 + b.rng.Intn(10) + 1
		effect := AttackEffectResult{
			Kind: affect.Kind, DamageFlags: flags, RolledDamage: rolledDamage, TargetHP: target.HitPoints,
			Protected: target.MonsterProtectedFromDamage(flags),
		}
		if !effect.Protected {
			damage := rolledDamage
			if damage > target.HitPoints {
				damage = target.HitPoints
			}
			effect.Damage = damage
			effect.TargetHP = target.HitPoints - damage
			if setErr := b.SetHitPoints(targetID, effect.TargetHP); setErr != nil {
				return AttackResult{}, setErr
			}
		}
		result.Effects = append(result.Effects, effect)
		result.TargetHP = effect.TargetHP
	}
	return result, nil
}

// AttackSequence resolves the number of attacks granted by the readied
// weapon's RateOfFire projection. A zero value keeps old callers at one
// attack. Target selection after a target falls belongs to the game adapter.
func (b *Battle) AttackSequence(attackerID, targetID string) ([]AttackResult, error) {
	attacker, ok := b.fighters[attackerID]
	if !ok {
		return nil, fmt.Errorf("unknown attacker %q", attackerID)
	}
	attacks := attacker.AttacksPerTurn
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
	if base, ok := target.MonsterMagicResistanceBase(); ok {
		// Magic Missile reaches the original pre-damage affect boundary with
		// the Magic damage flag set. Damage dice are consumed before this d100.
		resisted = b.rng.Intn(100)+1 <= MagicResistanceChance(base, level)
		if resisted {
			damage = 0
		}
	}
	if damage > target.HitPoints {
		damage = target.HitPoints
	}
	target.HitPoints -= damage
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
		saved := saveRoll == 20 ||
			saveRoll != 1 && saveRoll+target.SavingThrowBonus >= int(target.SavingThrows[4])
		applied := damage
		if saved {
			applied /= 2
		}
		protected := target.MonsterProtectedFromDamage(DamageFlagFire | DamageFlagMagic)
		if protected {
			applied = 0
		}
		if applied > target.HitPoints {
			applied = target.HitPoints
		}
		target.HitPoints -= applied
		b.fighters[target.ID] = target
		result.Impacts = append(result.Impacts, AreaSpellImpact{
			TargetID: target.ID, Damage: applied, TargetHP: target.HitPoints, Saved: saved,
			Protected: protected,
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
	if damage > target.HitPoints {
		damage = target.HitPoints
	}
	target.HitPoints -= damage
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
	partyAlive, enemyAlive := false, false
	for _, fighter := range b.fighters {
		if fighter.HitPoints <= 0 {
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
	case enemyAlive:
		b.status = StatusEnemyWon
	default:
		b.status = StatusDraw
	}
}
