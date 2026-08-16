package combat

import (
	"math/rand"
	"testing"
)

// escapeBattle 擺一個「隊員貼著一隻怪」的最小場面。
// 兩人相鄰（腳印相接），所以逃跑一定要走判定，不會落在「沒人貼著」那條捷徑。
func escapeBattle(t *testing.T, ownRate, enemyRate int, seed int64) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 4, MovementAllowance: ownRate},
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: 5, CombatY: 4, MovementAllowance: enemyRate},
	}, seed)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 比的是移動率，不是距離也不是先攻：敵人比我慢就走得掉。
func TestEscapeSucceedsWhenEveryOpponentIsSlower(t *testing.T) {
	battle := escapeBattle(t, 12, 9, 1)
	attempt, err := battle.AttemptEscape("hero")
	if err != nil {
		t.Fatal(err)
	}
	if !attempt.Contested {
		t.Fatal("敵人就貼在旁邊，Contested 應該是 true")
	}
	if !attempt.Escaped || attempt.TieRoll != 0 {
		t.Fatalf("attempt=%+v，速度較快時應該直接逃掉、不擲骰", attempt)
	}
	if battle.Status() != StatusPartyFled {
		t.Fatalf("場上沒有隊員了，status=%v，want StatusPartyFled", battle.Status())
	}
}

// 敵人比我快就走不掉，而且**不擲骰**——原作只在相等時才擲。
func TestEscapeFailsAgainstAFasterOpponent(t *testing.T) {
	battle := escapeBattle(t, 9, 12, 1)
	attempt, err := battle.AttemptEscape("hero")
	if err != nil {
		t.Fatal(err)
	}
	if attempt.Escaped || attempt.TieRoll != 0 {
		t.Fatalf("attempt=%+v，速度較慢時應該逃不掉、也不擲骰", attempt)
	}
	if battle.Status() != StatusActive {
		t.Fatalf("逃跑失敗不該結束戰鬥，status=%v", battle.Status())
	}
	// 沒逃掉的人還在場上，位置也沒動。
	fighter, ok := battle.Fighter("hero")
	if !ok || fighter.Escaped || !fighter.HasCombatPosition ||
		fighter.CombatX != 4 || fighter.CombatY != 4 {
		t.Fatalf("fighter=%+v ok=%v", fighter, ok)
	}
}

// 平手時擲 1d2，**1 才逃得掉**——平手偏向失敗。
// 這一條同時釘住「用的是同一條 PRNG」：期望值直接由 rand 重算。
func TestEscapeTieRollFavoursTheBlocker(t *testing.T) {
	const seed int64 = 20260817
	want := rand.New(rand.NewSource(seed)).Intn(2) + 1
	battle := escapeBattle(t, 12, 12, seed)
	attempt, err := battle.AttemptEscape("hero")
	if err != nil {
		t.Fatal(err)
	}
	if attempt.TieRoll != want {
		t.Fatalf("TieRoll=%d，同一條 PRNG 應該擲出 %d", attempt.TieRoll, want)
	}
	if attempt.Escaped != (want == 1) {
		t.Fatalf("attempt=%+v，1d2 ＝ %d 時 Escaped 應為 %v", attempt, want, want == 1)
	}
}

// 沒有敵方戰鬥員貼著就直接離場，連速度都不比。
func TestEscapeIsFreeWhenNoOpponentIsAdjacent(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 0, CombatY: 0, MovementAllowance: 3},
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: 20, CombatY: 10, MovementAllowance: 12},
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := battle.AttemptEscape("hero")
	if err != nil {
		t.Fatal(err)
	}
	if attempt.Contested || !attempt.Escaped || attempt.TieRoll != 0 {
		t.Fatalf("attempt=%+v，沒人貼著時應該直接逃掉", attempt)
	}
}

// 倒下的敵人不算數：既不擋路，也不參與「誰最快」。
func TestDownedOpponentsNeitherBlockNorSetThePace(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 4, MovementAllowance: 9},
		{ID: "corpse", Side: SideEnemy, HitPoints: 0, MaxHitPoints: 8,
			HasCombatPosition: true, CombatX: 5, CombatY: 4, MovementAllowance: 15},
		{ID: "far-orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: 20, CombatY: 10, MovementAllowance: 15},
	}, 3)
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := battle.AttemptEscape("hero")
	if err != nil {
		t.Fatal(err)
	}
	if attempt.Contested {
		t.Fatalf("屍體不該擋路：attempt=%+v", attempt)
	}
	if !attempt.Escaped {
		t.Fatalf("attempt=%+v", attempt)
	}
}

// 離場的隊員還活著——所以敵人沒有贏，剩下的隊員也還能打。
func TestEscapedFighterIsNotAKillAndTheRestFightOn(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "runner", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 0, CombatY: 0, MovementAllowance: 12},
		{ID: "stayer", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 1, CombatY: 0, MovementAllowance: 6},
		{ID: "orc", Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8, HitDice: 1,
			HasCombatPosition: true, CombatX: 20, CombatY: 10, MovementAllowance: 9},
	}, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := battle.AttemptEscape("runner"); err != nil {
		t.Fatal(err)
	}
	if battle.Status() != StatusActive {
		t.Fatalf("還有隊員在場，status=%v", battle.Status())
	}
	runner, _ := battle.Fighter("runner")
	if !runner.Escaped || runner.HitPoints != 10 {
		t.Fatalf("離場的人應該還活著：%+v", runner)
	}
	// 同一個人不能離場兩次。
	if _, err := battle.AttemptEscape("runner"); err == nil {
		t.Fatal("已經離場的人再逃一次應該回錯誤")
	}
}
