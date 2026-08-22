package combat

import "testing"

// 效果 `54h`：傷害時機（`06h`）問到時，若傷害屬性帶電（`damage_element` 的
// bit 2），把目前 HP **加 8**——被電反而回血。
//
// ★ 為什麼要一支測試。 這條規則在修正表裡是一筆「寫 `player` 第 420 格」的記錄
// 寫入，而 `+1A4h` ＝ 目前 HP（spec 098／1079／1101／1005 四份各自指到同一格）。
// 之前傷害時機那三個呼叫端只讀 `Applied[damage]`、把 `Records` 丟掉，所以這條
// 規則**算得出來卻沒有人套上去**——那種漏接不會有任何徵兆。
func TestLightningDamageHealsTheAfflictedFighter(t *testing.T) {
	fighter := Fighter{
		ID: "beast", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 40,
		MonsterAffects: []MonsterAffect{{Kind: 0x54, Active: true, Innate: true}},
	}
	// bit 2 ＝ 電。
	detail, err := CheckFX(fighter, checkFXDamage, map[string]int{
		scratchDamage: 10, scratchDamageElement: 4})
	if err != nil {
		t.Fatal(err)
	}
	if !applyRecordWritesTo(&fighter, detail) {
		t.Fatal("帶電的傷害應該改到記錄（效果 54h）")
	}
	if fighter.HitPoints != 28 {
		t.Fatalf("目前 HP ＝ %d，want 28（20 ＋ 8）", fighter.HitPoints)
	}
}

// ⚠ 守衛要生效：不帶電的傷害不該回血。少了這一半，任何「無條件加 8」的實作
// 都會通過上面那支測試。
func TestNonLightningDamageDoesNotHeal(t *testing.T) {
	fighter := Fighter{
		ID: "beast", Side: SideEnemy, HitPoints: 20, MaxHitPoints: 40,
		MonsterAffects: []MonsterAffect{{Kind: 0x54, Active: true, Innate: true}},
	}
	// bit 0 ＝ 火。
	detail, err := CheckFX(fighter, checkFXDamage, map[string]int{
		scratchDamage: 10, scratchDamageElement: 1})
	if err != nil {
		t.Fatal(err)
	}
	applyRecordWritesTo(&fighter, detail)
	if fighter.HitPoints != 20 {
		t.Fatalf("火焰傷害不該回血，目前 HP ＝ %d", fighter.HitPoints)
	}
}

// 上限壓在最大 HP：原作那一支沒有自己夾，因為別的路會把它拉回來；
// remake 只跑這一條，不夾就會累積出超過最大值的 HP。
func TestLightningHealIsCappedAtMaxHitPoints(t *testing.T) {
	fighter := Fighter{
		ID: "beast", Side: SideEnemy, HitPoints: 38, MaxHitPoints: 40,
		MonsterAffects: []MonsterAffect{{Kind: 0x54, Active: true, Innate: true}},
	}
	detail, err := CheckFX(fighter, checkFXDamage, map[string]int{
		scratchDamage: 10, scratchDamageElement: 4})
	if err != nil {
		t.Fatal(err)
	}
	applyRecordWritesTo(&fighter, detail)
	if fighter.HitPoints != 40 {
		t.Fatalf("回血要壓在最大 HP，目前 HP ＝ %d，want 40", fighter.HitPoints)
	}
}
