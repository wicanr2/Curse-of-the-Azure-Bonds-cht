package game

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"

// SoundEvent is a renderer-neutral gameplay sound intent. Platform adapters
// map the semantic event to DOS WAV, PC-98 SOUNDFX, or another audio backend;
// the game rules must not assume that different ports share selector numbers.
type SoundEvent string

const (
	SoundStop      SoundEvent = "stop"
	SoundCast      SoundEvent = "cast"
	SoundMiss      SoundEvent = "miss"
	SoundSpellHit  SoundEvent = "spell_hit"
	SoundDead      SoundEvent = "dead"
	SoundWhistle   SoundEvent = "whistle"
	SoundHit       SoundEvent = "hit"
	SoundLightning SoundEvent = "lightning"
	SoundSwish     SoundEvent = "swish"
	SoundStep      SoundEvent = "step"
	SoundFireball  SoundEvent = "fireball"
	SoundArrow     SoundEvent = "arrow"
	SoundOverture  SoundEvent = "overture"
	SoundCombat    SoundEvent = "combat"
	SoundCrash     SoundEvent = "crash"
)

func (s *State) requestSound(event SoundEvent) {
	s.pendingSoundEvents = append(s.pendingSoundEvents, event)
}

// ConsumeSoundEvents transfers one-shot audio intents to a platform adapter.
// Events are consumed exactly once; an adapter may ignore unknown/no-op IDs.
func (s *State) ConsumeSoundEvents() []SoundEvent {
	events := append([]SoundEvent(nil), s.pendingSoundEvents...)
	s.pendingSoundEvents = nil
	return events
}

// missileImpactSound 是原作 `SHOWARROW` 在飛行動畫尾端放的那一聲（spec 1186）。
//
// 原作依**飛出去那一件的物品類別**分歧；remake 用架著的武器類別，因為兩者只在
// 「要另外彈藥」時不同，而那時飛出去的是箭（`49h`）或弩矢（`1Ch`），**兩個都落在
// 同一個 ARROWFX 分支** ⇒ 結論一樣。
//
// ⚠ 進場那一聲 `ARROWFX` 是**無條件**的，不在這裡；這一支只回傳第二聲。
// 弓射一次會聽到兩聲 ARROWFX，那是原作的行為，不是重複發送。
func missileImpactSound(fighter combat.Fighter) SoundEvent {
	if fighter.UsesSeparateAmmunition() {
		return SoundArrow
	}
	switch fighter.WeaponItemType {
	case 0x02, 0x07, 0x14:
		// 擲斧、棍、鎚：`2B81h`。
		return SoundSwish
	case 0x55, 0x56, 0x65, 0x2F, 0x62:
		// 投石索類與油瓶：`2BB4h`／`2C01h`。`55h` 與 `62h` 在遊戲裡一件都沒有，
		// 照原作留著——分歧鏈點名了它們，走不到不代表不存在。
		return SoundWhistle
	case 0x09, 0x15, 0x64, 0x1C, 0x1F, 0x49:
		// 飛鏢、標槍、矛與彈藥本身：`2B4Ah`。
		return SoundArrow
	default:
		// `2C48h` 的預設分支。**不是空的**：弓與弩本身沒被點名，
		// 對照表見 `docs/audit/missile-sound-classes.md`。
		return SoundSwish
	}
}
