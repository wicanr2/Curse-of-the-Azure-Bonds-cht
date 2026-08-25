package combat

import "testing"

// 測試用規則：數字照 gamepack/rules/special-attacks.json（asm 立即數）。
func spitRule() SpecialAttackRule {
	return SpecialAttackRule{ID: "spit", EffectKind: 0x56, Form: SpecialAttackSpit,
		MissChance: 30, MaxDistance: 7, DamageMask: 0x30, SaveCategory: 3, TargetRange: 6}
}

func breathTouchRule() SpecialAttackRule {
	return SpecialAttackRule{ID: "fire", EffectKind: 0x83, Form: SpecialAttackBreathTouch,
		MissChance: 50, MaxDistance: 2, DamageMask: 1, ConstantDamage: 7,
		SaveCategory: 3, TargetRange: 1}
}

func specialAttackBattle(t *testing.T, seed int64, extra ...Fighter) *Battle {
	t.Helper()
	fighters := append([]Fighter{
		{ID: "beast", Side: SideEnemy, HitPoints: 24, MaxHitPoints: 24,
			HasCombatPosition: true, CombatX: 5, CombatY: 5},
		{ID: "hero", Side: SideParty, HitPoints: 30, MaxHitPoints: 30,
			HasCombatPosition: true, CombatX: 6, CombatY: 5},
	}, extra...)
	battle, err := NewBattle(fighters, seed)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

// 吐酸命中：傷害 ＝ 攻擊者 HP 上限（豁免失敗時全額）。
func TestSpecialAttackSpitHitsForAttackerMaxHP(t *testing.T) {
	for seed := int64(1); seed <= 64; seed++ {
		battle := specialAttackBattle(t, seed)
		result, err := battle.SpecialAttackSingle("beast", "hero", spitRule())
		if err != nil {
			t.Fatal(err)
		}
		if result.Refrained {
			t.Fatalf("seed %d: adjacent spit must not refrain", seed)
		}
		if result.Missed {
			continue
		}
		impact := result.Impacts[0]
		want := 24
		if impact.Saved {
			want = 12
		}
		if impact.Damage != want {
			t.Fatalf("seed %d: spit damage=%d saved=%v, want %d",
				seed, impact.Damage, impact.Saved, want)
		}
	}
}

// 吐酸的擲失是 Missed（有訊息），不是 Refrained；統計上大約七成擲失。
func TestSpecialAttackSpitMissesAboutSeventyPercent(t *testing.T) {
	missed := 0
	for seed := int64(1); seed <= 200; seed++ {
		battle := specialAttackBattle(t, seed)
		result, err := battle.SpecialAttackSingle("beast", "hero", spitRule())
		if err != nil {
			t.Fatal(err)
		}
		if result.Missed {
			missed++
		}
	}
	if missed < 110 || missed > 170 {
		t.Fatalf("spit missed %d/200, want around 140 (1d100 > 30)", missed)
	}
}

// 吐酸距離門檻：`entry#33` 近似值 >= 7 就放棄（亂數照樣先消耗一次）。
func TestSpecialAttackSpitRefrainsOutOfRange(t *testing.T) {
	battle := specialAttackBattle(t, 3, Fighter{ID: "far", Side: SideParty,
		HitPoints: 30, MaxHitPoints: 30, HasCombatPosition: true, CombatX: 20, CombatY: 5})
	result, err := battle.SpecialAttackSingle("beast", "far", spitRule())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Refrained {
		t.Fatalf("distant spit must refrain, got %+v", result)
	}
}

// 龍息火：擲失（1d100 > 50）是靜默 Refrained；命中傷害是常數 7。
func TestSpecialAttackBreathTouchRefrainsSilentlyOrDealsSeven(t *testing.T) {
	refrained, hit := 0, 0
	for seed := int64(1); seed <= 200; seed++ {
		battle := specialAttackBattle(t, seed)
		result, err := battle.SpecialAttackSingle("beast", "hero", breathTouchRule())
		if err != nil {
			t.Fatal(err)
		}
		if result.Missed {
			t.Fatalf("seed %d: breath touch never reports a miss message", seed)
		}
		if result.Refrained {
			refrained++
			continue
		}
		hit++
		impact := result.Impacts[0]
		want := 7
		if impact.Saved {
			want = 3
		}
		if impact.Damage != want {
			t.Fatalf("seed %d: breath damage=%d saved=%v, want %d",
				seed, impact.Damage, impact.Saved, want)
		}
	}
	if refrained < 60 || hit < 60 {
		t.Fatalf("breath touch split refrained=%d hit=%d, want both around 100", refrained, hit)
	}
}

// 區域吐酸：範圍裡混進攻擊者同側就整次取消，而且不扣次數。
func TestSpecialAttackAcidBreathCancelsOnFriendInArea(t *testing.T) {
	rule := SpecialAttackRule{ID: "acid-area", EffectKind: 0x5A,
		Form: SpecialAttackBreathAreaSameSide, Uses: 3, DamageMask: 0x30,
		SaveCategory: 3, TargetRange: 6, AreaRadius: 1}
	battle := specialAttackBattle(t, 5, Fighter{ID: "ally", Side: SideEnemy,
		HitPoints: 10, MaxHitPoints: 10, HasCombatPosition: true, CombatX: 6, CombatY: 6})
	result, err := battle.SpecialAttackAreaBreath("beast", TilePoint{X: 6, Y: 5}, rule)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Refrained {
		t.Fatalf("acid breath with an ally in the area must cancel, got %+v", result)
	}
	attacker, _ := battle.Fighter("beast")
	if attacker.SpecialAttackUses != 3 {
		t.Fatalf("cancelled breath consumed a use: %d", attacker.SpecialAttackUses)
	}
}

// 噴火沒有同類檢查：自己人也會被燒；次數三次用完就 Refrained。
func TestSpecialAttackFireBreathHitsAlliesAndRunsOutOfUses(t *testing.T) {
	rule := SpecialAttackRule{ID: "fire-area", EffectKind: 0x80,
		Form: SpecialAttackBreathArea, Uses: 3, DamageMask: 0x21,
		SaveCategory: 3, TargetRange: 9, AreaRadius: 2}
	battle := specialAttackBattle(t, 7,
		Fighter{ID: "ally", Side: SideEnemy, HitPoints: 400, MaxHitPoints: 400,
			HasCombatPosition: true, CombatX: 6, CombatY: 6},
		Fighter{ID: "tank", Side: SideParty, HitPoints: 400, MaxHitPoints: 400,
			HasCombatPosition: true, CombatX: 6, CombatY: 4})
	for use := 1; use <= 3; use++ {
		result, err := battle.SpecialAttackAreaBreath("beast", TilePoint{X: 6, Y: 5}, rule)
		if err != nil {
			t.Fatal(err)
		}
		if result.Refrained {
			t.Fatalf("use %d refrained unexpectedly", use)
		}
		hitAlly := false
		for _, impact := range result.Impacts {
			if impact.TargetID == "ally" {
				hitAlly = true
			}
		}
		if !hitAlly {
			t.Fatalf("use %d: fire breath must also hit the attacker's ally", use)
		}
	}
	result, err := battle.SpecialAttackAreaBreath("beast", TilePoint{X: 6, Y: 5}, rule)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Refrained {
		t.Fatal("fourth fire breath must refrain (3 uses per combat)")
	}
}

// 凝視：豁免成功完全沒事；失敗掛 `34h`，MonsterIsHeld 因此成立。
func TestSpecialAttackGazeParalyzesOnFailedSave(t *testing.T) {
	rule := SpecialAttackRule{ID: "gaze", EffectKind: 0x7E, Form: SpecialAttackGaze,
		SaveCategory: 1, TargetRange: 6, ParalysisEffect: 0x34, ParalysisDuration: 60}
	paralyzed, resisted := 0, 0
	for seed := int64(1); seed <= 100; seed++ {
		battle := specialAttackBattle(t, seed, Fighter{ID: "warded", Side: SideParty,
			HitPoints: 30, MaxHitPoints: 30, HasCombatPosition: true, CombatX: 5, CombatY: 6,
			SavingThrows: []uint8{0, 10}})
		result, err := battle.SpecialAttackGazeAt("beast", "warded", rule)
		if err != nil {
			t.Fatal(err)
		}
		impact := result.Impacts[0]
		target, _ := battle.Fighter("warded")
		if impact.Paralyzed {
			paralyzed++
			if !target.MonsterIsHeld() {
				t.Fatalf("seed %d: paralyzed target must be held", seed)
			}
			if target.HitPoints != 30 {
				t.Fatalf("seed %d: gaze must not deal damage", seed)
			}
		} else {
			resisted++
			if target.MonsterIsHeld() {
				t.Fatalf("seed %d: saved target must not be held", seed)
			}
		}
	}
	if paralyzed == 0 || resisted == 0 {
		t.Fatalf("gaze split paralyzed=%d resisted=%d, want both nonzero", paralyzed, resisted)
	}
}

// pack 宣告要能整包轉成規則（形狀、必填欄位）。
func TestSpecialAttackRulesFromPackResolve(t *testing.T) {
	rules, err := SpecialAttackRulesFromPack()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 5 {
		t.Fatalf("pack declares %d special attacks, want 5", len(rules))
	}
	kinds := map[uint8]bool{}
	for _, rule := range rules {
		kinds[rule.EffectKind] = true
		if rule.Message == "" {
			t.Fatalf("rule %q has no message ID", rule.ID)
		}
	}
	for _, kind := range []uint8{0x56, 0x5A, 0x7E, 0x80, 0x83} {
		if !kinds[kind] {
			t.Fatalf("effect 0x%02X is not declared", kind)
		}
	}
}
