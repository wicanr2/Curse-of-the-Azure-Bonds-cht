package combat

import "testing"

// specTurnUndeadTable 是 spec 834 表格那 10×10 的獨立謄本：`D` ＝ 自動摧毀、
// `—` ＝ 不可能。故意不共用 `turnUndeadMatrix`——兩份各自出錯的機率才低。
var specTurnUndeadTable = [10][10]string{
	{"10", "7", "4", "1", "1", "D", "D", "D", "D", "D"},
	{"13", "10", "7", "1", "1", "D", "D", "D", "D", "D"},
	{"16", "13", "10", "4", "1", "1", "D", "D", "D", "D"},
	{"19", "16", "13", "7", "4", "1", "1", "D", "D", "D"},
	{"20", "19", "16", "10", "7", "4", "1", "1", "D", "D"},
	{"—", "20", "19", "13", "10", "7", "4", "1", "1", "D"},
	{"—", "—", "20", "16", "13", "10", "7", "4", "1", "D"},
	{"—", "—", "—", "20", "16", "13", "10", "7", "4", "1"},
	{"—", "—", "—", "—", "20", "16", "13", "10", "7", "1"},
	{"—", "—", "—", "—", "—", "20", "16", "13", "10", "4"},
}

func TestTurnUndeadMatrixMatchesTheOriginalTable(t *testing.T) {
	for row := range specTurnUndeadTable {
		for col := range specTurnUndeadTable[row] {
			want := specTurnUndeadTable[row][col]
			value := turnUndeadMatrix[row][col]
			var got string
			switch {
			case value == 99:
				got = "—"
			case value <= 0:
				got = "D"
			default:
				got = itoa(int(value))
			}
			if got != want {
				t.Errorf("種類 %d 欄 %d ＝ %q，want %q（原始位元組 %d）",
					row+1, col+1, got, want, value)
			}
		}
	}
	// ⚠ `0` 與 `−1` 都是自動摧毀，但只有 `−1` 會把額外配額加回來。表要保留原始
	// 位元組的區別，統一成同一個值會讓那條規則失效。
	if turnUndeadMatrix[0][5] != 0 || turnUndeadMatrix[0][7] != -1 {
		t.Fatalf("種類 1 的自動摧毀格被統一了：%v", turnUndeadMatrix[0])
	}
}

func itoa(value int) string {
	if value < 10 {
		return string(rune('0' + value))
	}
	return string(rune('0'+value/10)) + string(rune('0'+value%10))
}

func turnUndeadBattle(t *testing.T, seed int64, clericLevel uint8, undead ...uint8) *Battle {
	t.Helper()
	fighters := []Fighter{{
		ID: "cleric", Side: SideParty, HitPoints: 12, MaxHitPoints: 12,
		ClericLevel: clericLevel, HasCombatPosition: true, CombatX: 1, CombatY: 1,
	}}
	for index, kind := range undead {
		fighters = append(fighters, Fighter{
			ID: undeadID(index), Side: SideEnemy, HitPoints: 8, MaxHitPoints: 8,
			HitDice: 1, UndeadType: kind, HasCombatPosition: true,
			CombatX: 3 + index, CombatY: 3,
		})
	}
	battle, err := NewBattle(fighters, seed)
	if err != nil {
		t.Fatal(err)
	}
	return battle
}

func undeadID(index int) string { return "undead" + itoa(index) }

// seedWith 找出一顆會擲出指定 1d12／1d20 的種子。找不到就讓測試紅——**不要
// 退回「隨便一顆」**，那會讓後面的斷言變成看運氣。
func seedWith(t *testing.T, allowance, roll int, clericLevel uint8, undead ...uint8) int64 {
	t.Helper()
	for seed := int64(1); seed <= 20000; seed++ {
		battle := turnUndeadBattle(t, seed, clericLevel, undead...)
		result, err := battle.TurnUndead("cleric")
		if err != nil {
			t.Fatal(err)
		}
		if result.Allowance == allowance && result.Roll == roll {
			return seed
		}
	}
	t.Fatalf("找不到 1d12=%d 且 1d20=%d 的種子", allowance, roll)
	return 0
}

// 種類 1 欄 3 的門檻是 4：骰 4 過、骰 3 不過。兩面都要釘。
func TestTurnUndeadThresholdIsTwoSided(t *testing.T) {
	for _, item := range []struct {
		roll   int
		turned bool
	}{{roll: 4, turned: true}, {roll: 3, turned: false}} {
		seed := seedWith(t, 6, item.roll, 3, 1)
		battle := turnUndeadBattle(t, seed, 3, 1)
		result, err := battle.TurnUndead("cleric")
		if err != nil {
			t.Fatal(err)
		}
		if result.Column != 3 {
			t.Fatalf("欄 ＝ %d，want 3", result.Column)
		}
		if item.turned {
			if len(result.Outcomes) != 1 || result.Outcomes[0].Destroyed || result.Stopped {
				t.Fatalf("骰 %d 應該驅離：%+v", item.roll, result)
			}
			if got := battle.fighters["undead0"].RawCombatState10; got != 1 {
				t.Fatalf("被驅離的目標 `+10h` ＝ %d，want 1", got)
			}
			continue
		}
		if !result.Stopped || !result.Nothing() {
			t.Fatalf("骰 %d 應該整批停：%+v", item.roll, result)
		}
		if got := battle.fighters["undead0"].RawCombatState10; got != 0 {
			t.Fatalf("沒過的目標不該被標成在逃：%d", got)
		}
	}
}

// 值 ≤ 0 走摧毀那條：離場而且不留屍體。
func TestTurnUndeadDestroysWhenTheTableValueIsNotPositive(t *testing.T) {
	// 種類 1 欄 8 是 −1 ＝ 自動摧毀。牧師等級 8 → 合計 8 → 欄 8。
	seed := seedWith(t, 6, 20, 8, 1)
	battle := turnUndeadBattle(t, seed, 8, 1)
	result, err := battle.TurnUndead("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outcomes) != 1 || !result.Outcomes[0].Destroyed {
		t.Fatalf("應該摧毀：%+v", result)
	}
	target := battle.fighters["undead0"]
	if target.HitPoints != 0 || target.DownedCorpse {
		t.Fatalf("被摧毀的目標 hp=%d corpse=%v", target.HitPoints, target.DownedCorpse)
	}
	if battle.Status() != StatusPartyWon {
		t.Fatalf("摧毀最後一名敵人後 status=%v，want StatusPartyWon", battle.Status())
	}
}

// 配額是 1d12，而**摧毀最多可以再多做 6 個**：配額歸零時若這一次是 `v < 0` 的
// 摧毀就加回一格。把配額壓到 1 才看得出這條規則。
func TestTurnUndeadDestroyDoesNotConsumeTheAllowance(t *testing.T) {
	// 種類 1 欄 8 ＝ −1（摧毀且加回配額）。三個目標、配額 1。
	seed := seedWith(t, 1, 20, 8, 1, 1, 1)
	battle := turnUndeadBattle(t, seed, 8, 1, 1, 1)
	result, err := battle.TurnUndead("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outcomes) != 3 {
		t.Fatalf("三個都該被摧毀，實際 %d 個：%+v", len(result.Outcomes), result)
	}

	// 對照組：種類 1 欄 6 ＝ 0，一樣自動摧毀但**不加回配額**，所以只做得到一個。
	seed = seedWith(t, 1, 20, 6, 1, 1, 1)
	battle = turnUndeadBattle(t, seed, 6, 1, 1, 1)
	result, err = battle.TurnUndead("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outcomes) != 1 || !result.Outcomes[0].Destroyed {
		t.Fatalf("`0` 那幾格不該加回配額，實際 %d 個：%+v", len(result.Outcomes), result)
	}
}

// 目標從**最弱的**開始（`+0E9h` 最小的，spec 815 的挑選器）。
func TestTurnUndeadTakesTheWeakestFirst(t *testing.T) {
	// 牧師等級 3 → 欄 3。種類 1 門檻 4、種類 3 門檻 10。骰 4：弱的過、強的不過。
	seed := seedWith(t, 6, 4, 3, 3, 1)
	battle := turnUndeadBattle(t, seed, 3, 3, 1)
	result, err := battle.TurnUndead("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outcomes) != 1 || result.Outcomes[0].TargetID != "undead1" {
		t.Fatalf("應該先處理種類 1 的那一個：%+v", result)
	}
	if !result.Stopped {
		t.Fatal("第一個沒過就該整批停")
	}
}

// 用過就不能再用（`TRIEDTOTURN`），而且非牧師連叫都不該成功。
func TestTurnUndeadIsOncePerCombatAndClericOnly(t *testing.T) {
	battle := turnUndeadBattle(t, 1, 3, 1)
	if _, err := battle.TurnUndead("cleric"); err != nil {
		t.Fatal(err)
	}
	if !battle.fighters["cleric"].TriedToTurn {
		t.Fatal("第一次之後 `+11h` 應該是 1")
	}
	if _, err := battle.TurnUndead("cleric"); err == nil {
		t.Fatal("第二次應該被擋下來")
	}

	plain := turnUndeadBattle(t, 1, 0, 1)
	if _, err := plain.TurnUndead("cleric"); err == nil {
		t.Fatal("不是牧師就不該能驅散")
	}
}

// 沒有不死生物時：一樣算用掉了，但一個都沒成。
func TestTurnUndeadWithNoUndeadDoesNothing(t *testing.T) {
	battle := turnUndeadBattle(t, 1, 5)
	result, err := battle.TurnUndead("cleric")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Nothing() || result.Stopped {
		t.Fatalf("沒有目標時應該是「什麼都沒發生」：%+v", result)
	}
	if !battle.fighters["cleric"].TriedToTurn {
		t.Fatal("沒有目標也算用掉了")
	}
}
