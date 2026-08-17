package combat

import "testing"

func inRangeTargetBattle(t *testing.T, seed int64) *Battle {
	t.Helper()
	battle, err := NewBattle([]Fighter{
		{ID: "ogre", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20,
			HasCombatPosition: true, CombatX: 5, CombatY: 5, WeaponRange: 1},
		{ID: "near-a", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 4},
		{ID: "near-b", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 6, CombatY: 5},
		{ID: "far", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 12},
	}, seed)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 射程外的人不能被挑到，射程內的每一個都要挑得到——原作是 `1d n`，
// 不是「挑最近的」。挑最近的會讓同一個場面每回合打同一個人。
func TestInRangeTargetPicksUniformlyAmongCandidates(t *testing.T) {
	seen := map[string]int{}
	for seed := int64(1); seed <= 40; seed++ {
		battle := inRangeTargetBattle(t, seed)
		target, found, err := battle.SelectInRangeTarget("ogre", SideParty)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatalf("seed %d：貼身有兩個人卻回報射程內沒人", seed)
		}
		if target.ID == "far" {
			t.Fatalf("seed %d：挑到 7 格外的目標，射程過濾沒生效", seed)
		}
		seen[target.ID]++
	}
	if len(seen) != 2 {
		t.Fatalf("40 次只挑到 %v；兩個候選都該出現過", seen)
	}
}

// 沒人在射程內時要回報「沒有」，讓呼叫端接手移動——不能退而求其次挑個遠的，
// 那會讓怪物原地攻擊打不到的人。
func TestInRangeTargetReportsNoneWhenEveryoneIsFar(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "ogre", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20,
			HasCombatPosition: true, CombatX: 1, CombatY: 1, WeaponRange: 1},
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 9, CombatY: 9},
	}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := battle.SelectInRangeTarget("ogre", SideParty); err != nil {
		t.Fatal(err)
	} else if found {
		t.Fatal("8 格外的目標被當成射程內")
	}
}

// 倒下與離場的不是候選。
func TestInRangeTargetSkipsDownedAndFledFighters(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "ogre", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20,
			HasCombatPosition: true, CombatX: 5, CombatY: 5, WeaponRange: 1},
		{ID: "downed", Side: SideParty, HitPoints: 0, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 4},
		{ID: "fled", Side: SideParty, HitPoints: 10, MaxHitPoints: 10, Escaped: true,
			HasCombatPosition: true, CombatX: 6, CombatY: 5},
		{ID: "standing", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 4, CombatY: 5},
	}, 11)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 20; attempt++ {
		target, found, err := battle.SelectInRangeTarget("ogre", SideParty)
		if err != nil {
			t.Fatal(err)
		}
		if !found || target.ID != "standing" {
			t.Fatalf("挑到 %q（found=%v），只有 standing 是合法候選", target.ID, found)
		}
	}
}
