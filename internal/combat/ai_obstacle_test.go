package combat

import "testing"

func neverSaves() bool  { return false }
func alwaysSaves() bool { return true }

// `1Eh`：六個效果任一個成立就照走；都沒有才擲豁免。
func TestSaveableObstacleLetsTheRightEffectsThrough(t *testing.T) {
	bare := Fighter{HitDice: 3}
	if !ObstacleTerrainBlocks(bare, ObstacleTerrainSaveable, false, neverSaves) {
		t.Fatal("沒有任何豁免條件卻走得過去")
	}
	if ObstacleTerrainBlocks(bare, ObstacleTerrainSaveable, false, alwaysSaves) {
		t.Fatal("豁免成功還是被擋住")
	}
	if ObstacleTerrainBlocks(bare, ObstacleTerrainSaveable, true, neverSaves) {
		t.Fatal("戰鬥狀態 `+10h` 不為 0 該直接放行")
	}
	for _, kind := range obstacleSaveBypassEffects {
		affected := Fighter{HitDice: 3, MonsterAffects: []MonsterAffect{{Kind: kind}}}
		if ObstacleTerrainBlocks(affected, ObstacleTerrainSaveable, false, neverSaves) {
			t.Fatalf("帶著效果 %02Xh 還是過不去", kind)
		}
	}
	// 只在 `1Ch` 清單裡的 85h 不能拿來過 `1Eh`——兩張清單不一樣。
	wrong := Fighter{HitDice: 3, MonsterAffects: []MonsterAffect{{Kind: 0x85}}}
	if !ObstacleTerrainBlocks(wrong, ObstacleTerrainSaveable, false, neverSaves) {
		t.Fatal("85h 不在 1Eh 的清單裡，不該放行")
	}
}

// `1Ch`：等級 ≥ 7 或四個效果之一；**不擲豁免**。
func TestVeteranObstacleUsesLevelNotASavingThrow(t *testing.T) {
	rolled := false
	spy := func() bool { rolled = true; return true }
	if !ObstacleTerrainBlocks(Fighter{HitDice: 6}, ObstacleTerrainVeteran, false, spy) {
		t.Fatal("等級 6 該被擋住")
	}
	if rolled {
		t.Fatal("1Ch 不擲豁免，多擲一次會讓亂數序列偏掉")
	}
	if ObstacleTerrainBlocks(Fighter{HitDice: 7}, ObstacleTerrainVeteran, false, spy) {
		t.Fatal("等級 7 是門檻本身，該放行")
	}
	for _, kind := range obstacleVeteranBypassEffects {
		affected := Fighter{HitDice: 1, MonsterAffects: []MonsterAffect{{Kind: kind}}}
		if ObstacleTerrainBlocks(affected, ObstacleTerrainVeteran, false, spy) {
			t.Fatalf("帶著效果 %02Xh 還是過不去", kind)
		}
	}
	// 20h 只在 1Eh 的清單裡。
	wrong := Fighter{HitDice: 1, MonsterAffects: []MonsterAffect{{Kind: 0x20}}}
	if !ObstacleTerrainBlocks(wrong, ObstacleTerrainVeteran, false, spy) {
		t.Fatal("20h 不在 1Ch 的清單裡，不該放行")
	}
}

// `0E5h` 是**有號**位元組：怪物的 HD 可以超過 7Fh，那時當負數而不是「很強」。
func TestVeteranObstacleReadsTheLevelByteAsSigned(t *testing.T) {
	if !ObstacleTerrainBlocks(Fighter{HitDice: 0x80}, ObstacleTerrainVeteran, false, neverSaves) {
		t.Fatal("80h 有號是 −128，不該當成 128 放行")
	}
}

// 不是這兩個碼的地形一律不管——障礙走的是另一條路，不進成本表。
func TestOtherTerrainCodesAreNotObstacles(t *testing.T) {
	for _, code := range []uint8{0x00, 0x01, 0x1D, 0x1F, 0xFF} {
		if ObstacleTerrainBlocks(Fighter{HitDice: 1}, code, false, neverSaves) {
			t.Fatalf("地形碼 %02Xh 被當成障礙", code)
		}
	}
}

// 接到移動上：怪物走不進障礙格，於是換下一個候選方向。
func TestApproachRoutesAroundAnObstacleColumn(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "ogre", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 20, HitDice: 3,
			HasCombatPosition: true, CombatX: 5, CombatY: 8, WeaponRange: 1,
			MovementAllowance: 12, SavingThrows: []uint8{20, 20, 20, 20, 20}},
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 2},
	}, 5)
	if err != nil {
		t.Fatal(err)
	}
	// y = 6 整條是 1Ch，只有 x = 8 有缺口。豁免門檻 20 ⇒ 幾乎一定失敗，
	// 而 1Ch 本來就不擲豁免。
	battle.SetCombatTerrainCodes(func(x, y int) (uint8, bool) {
		if y == 6 && x != 8 {
			return ObstacleTerrainVeteran, true
		}
		return 1, true
	})
	result, err := battle.MonsterApproach("ogre", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range result.Steps {
		if step.Y == 6 && step.X != 8 {
			t.Fatalf("走進了障礙格 (%d,%d)", step.X, step.Y)
		}
	}
	if len(result.Steps) == 0 {
		t.Fatal("一步都沒走——障礙把整條路都封死了？")
	}
}

// 等級夠的怪照樣直線穿過去。
func TestVeteranMonsterWalksStraightThroughTheObstacle(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "giant", Side: SideEnemy, HitPoints: 40, MaxHitPoints: 40, HitDice: 9,
			HasCombatPosition: true, CombatX: 5, CombatY: 8, WeaponRange: 1,
			MovementAllowance: 12, SavingThrows: []uint8{20, 20, 20, 20, 20}},
		{ID: "hero", Side: SideParty, HitPoints: 10, MaxHitPoints: 10,
			HasCombatPosition: true, CombatX: 5, CombatY: 2},
	}, 5)
	if err != nil {
		t.Fatal(err)
	}
	battle.SetCombatTerrainCodes(func(x, y int) (uint8, bool) {
		if y == 6 {
			return ObstacleTerrainVeteran, true
		}
		return 1, true
	})
	result, err := battle.MonsterApproach("giant", "hero", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	crossed := false
	for _, step := range result.Steps {
		if step.Y == 6 {
			crossed = true
		}
	}
	if !crossed {
		t.Fatalf("HD 9 的怪沒穿過 1Ch，走了 %v", result.Steps)
	}
}

// 天然 1 一定失敗、天然 20 一定成功——門檻再怎麼樣都不看。
func TestSavingThrowHonoursTheNaturalOneAndTwenty(t *testing.T) {
	battle, err := NewBattle([]Fighter{
		{ID: "a", Side: SideParty, HitPoints: 5, MaxHitPoints: 5, SavingThrows: []uint8{1, 1, 1, 1, 1}},
		{ID: "b", Side: SideEnemy, HitPoints: 5, MaxHitPoints: 5, SavingThrows: []uint8{99, 99, 99, 99, 99}},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	sawOne, sawTwenty := false, false
	for attempt := 0; attempt < 400 && !(sawOne && sawTwenty); attempt++ {
		result, err := battle.RollSavingThrow(battle.fighters["a"], 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if result.Roll == 1 {
			sawOne = true
			if result.Saved {
				t.Fatal("門檻 1、天然 1 卻算過了")
			}
		}
		if result.Roll == 20 {
			sawTwenty = true
		}
	}
	if !sawOne {
		t.Skip("400 次沒擲到 1")
	}
	for attempt := 0; attempt < 400; attempt++ {
		result, err := battle.RollSavingThrow(battle.fighters["b"], 4, 0)
		if err != nil {
			t.Fatal(err)
		}
		if result.Roll == 20 {
			sawTwenty = true
			if !result.Saved {
				t.Fatal("門檻 99、天然 20 卻算沒過")
			}
		} else if result.Saved {
			t.Fatalf("門檻 99 卻用 %d 過了", result.Roll)
		}
	}
	if !sawTwenty {
		t.Skip("400 次沒擲到 20")
	}
}
