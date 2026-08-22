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
