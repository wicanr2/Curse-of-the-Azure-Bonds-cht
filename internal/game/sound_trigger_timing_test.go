package game

import (
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/combat"
)

// 視覺時間軸那條路才是玩家實際聽到的。時機照 spec 1186（原作 54 處 `SOUNDFX`
// 的逐處語意）：
//
//   - 近戰的 travel 階段放 `SWISHFX`，**不看中不中**；
//   - 近戰命中補 `HITFX`，**揮空什麼都不放**；
//   - `TWINKLE` 用同一個 `if` 選：中了 `SPELLHITFX`、沒中 `MISSFX`。
//
// ⚠ 這一組刻意包含**兩個負對照**（近戰揮空、法術命中）。只驗正例的話，
// 「揮空多放一聲法術音效」與「法術沒中放成命中音」兩種錯都驗不出來。
func soundsForVisual(t *testing.T, event combat.VisualEvent) []SoundEvent {
	t.Helper()
	state := NewState(testCatalog())
	state.EnableCombatVisualTimeline(true)
	if !state.queueCombatVisual(event) {
		t.Fatal("視覺事件沒有排進去")
	}
	state.ConsumeSoundEvents()
	// ⚠ 只推到**交棒階段之前**。推到 Done 會走 `finishCombat()`，而這裡沒有真的
	// 戰鬥；更重要的是交棒那一支會把 impact 標成「已送出」而不發音——直接推到底
	// 會得到一份空的音效清單，看起來像「什麼都沒發」。
	limit := event.Duration() - combat.VisualHandoffDuration
	var sounds []SoundEvent
	for elapsed := time.Duration(0); elapsed <= limit; elapsed += 10 * time.Millisecond {
		if err := state.AdvanceCombatVisual(elapsed); err != nil {
			t.Fatal(err)
		}
		sounds = append(sounds, state.ConsumeSoundEvents()...)
	}
	return sounds
}

func meleeVisual(hit bool) combat.VisualEvent {
	return combat.VisualEvent{
		Kind: combat.VisualMelee, ActorID: "a", TargetID: "b",
		From: combat.TilePoint{X: 1, Y: 1}, To: combat.TilePoint{X: 2, Y: 1},
		Hit: hit, Projectiles: 1,
		Impacts: []combat.VisualImpactTarget{{TargetID: "b", Hit: hit}},
	}
}

func TestMeleeVisualPlaysTheSwingSoundOnEverySwing(t *testing.T) {
	hitSounds := soundsForVisual(t, meleeVisual(true))
	if want := []SoundEvent{SoundSwish, SoundHit}; !reflect.DeepEqual(hitSounds, want) {
		t.Fatalf("命中：%#v，want %#v", hitSounds, want)
	}
	missSounds := soundsForVisual(t, meleeVisual(false))
	if want := []SoundEvent{SoundSwish}; !reflect.DeepEqual(missSounds, want) {
		t.Fatalf("揮空：%#v，want %#v", missSounds, want)
	}
}

func twinkleVisual(hit bool) combat.VisualEvent {
	return combat.VisualEvent{
		Kind: combat.VisualTwinkle, ActorID: "a", TargetID: "b",
		From: combat.TilePoint{X: 1, Y: 1}, To: combat.TilePoint{X: 3, Y: 1},
		Hit: hit,
		Impacts: []combat.VisualImpactTarget{{TargetID: "b", Hit: hit}},
	}
}

func TestTwinkleVisualSplitsHitAndMissLikeTheOriginal(t *testing.T) {
	hitSounds := soundsForVisual(t, twinkleVisual(true))
	if !containsSound(hitSounds, SoundSpellHit) || containsSound(hitSounds, SoundMiss) {
		t.Fatalf("法術命中：%#v，該有 SoundSpellHit 且不該有 SoundMiss", hitSounds)
	}
	missSounds := soundsForVisual(t, twinkleVisual(false))
	if !containsSound(missSounds, SoundMiss) || containsSound(missSounds, SoundSpellHit) {
		t.Fatalf("法術沒中：%#v，該有 SoundMiss 且不該有 SoundSpellHit", missSounds)
	}
}

func containsSound(events []SoundEvent, want SoundEvent) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

// 投射武器的第二聲（原作 `SHOWARROW` 的類別分歧，spec 1186）。
// 對照表與每一類在遊戲裡的實例數在 `docs/audit/missile-sound-classes.md`。
func TestMissileImpactSoundFollowsTheOriginalWeaponClasses(t *testing.T) {
	for _, item := range []struct {
		name       string
		fighter    combat.Fighter
		wantSecond SoundEvent
	}{
		// 弓：要另外的彈藥 ⇒ 飛出去的是箭（49h），落在 ARROWFX 分支。
		{"長弓", combat.Fighter{
			MissileWeapon: true, AmmunitionType: 0x0B, WeaponItemType: 0x2B,
		}, SoundArrow},
		// 弩：同理，飛出去的是弩矢（1Ch）。
		{"輕弩", combat.Fighter{
			MissileWeapon: true, AmmunitionType: 0x8A, WeaponItemType: 0x2E,
		}, SoundArrow},
		// 投石索：`+0Eh` ＝ 0Ah 自給自足 ⇒ 用武器自己的類別 2Fh → 哨音。
		{"投石索", combat.Fighter{
			MissileWeapon: true, AmmunitionType: 0x0A, WeaponItemType: 0x2F,
		}, SoundWhistle},
		{"小筏投石索", combat.Fighter{
			MissileWeapon: true, AmmunitionType: 0x0A, WeaponItemType: 0x65,
		}, SoundWhistle},
		// 油瓶（56h）也在哨音那一組。
		{"油瓶", combat.Fighter{
			MissileWeapon: true, ThrownWeapon: true, AmmunitionType: 0x1A, WeaponItemType: 0x56,
		}, SoundWhistle},
		// 投擲武器：自己就是彈藥。飛鏢／標槍走箭，擲斧／棍／鎚走揮擊。
		{"飛鏢", combat.Fighter{
			MissileWeapon: true, ThrownWeapon: true, AmmunitionType: 0x1A, WeaponItemType: 0x09,
		}, SoundArrow},
		{"矛", combat.Fighter{
			ThrownWeapon: true, AmmunitionType: 0x14, WeaponItemType: 0x1F,
		}, SoundArrow},
		{"擲斧", combat.Fighter{
			ThrownWeapon: true, AmmunitionType: 0x14, WeaponItemType: 0x02,
		}, SoundSwish},
		{"鎚", combat.Fighter{
			ThrownWeapon: true, AmmunitionType: 0x14, WeaponItemType: 0x14,
		}, SoundSwish},
	} {
		if got := missileImpactSound(item.fighter); got != item.wantSecond {
			t.Errorf("%s：第二聲 %q，want %q", item.name, got, item.wantSecond)
		}
	}

	// ⚠ 負對照：預設分支**不是空的**。沒被分歧鏈點名的類別要落到揮擊，
	// 不能因為「反正都是投射武器」就一律回箭——那樣上面那幾條會全部通過，
	// 而分歧鏈等於沒接。
	unnamed := combat.Fighter{MissileWeapon: true, AmmunitionType: 0x0A, WeaponItemType: 0x7F}
	if got := missileImpactSound(unnamed); got != SoundSwish {
		t.Fatalf("沒被點名的類別 %q，want %q", got, SoundSwish)
	}
}
