package combat

import "testing"

// DownOverkill 是 SAVEDAMAGE 階梯（spec 1205）的輸入：放倒那一擊的
// 傷害 − 當時 HP。追擊不覆寫、剛好歸零記 0。
func TestApplyPositiveDamageRecordsDownOverkill(t *testing.T) {
	battle := &Battle{}
	target := Fighter{HitPoints: 5, MaxHitPoints: 20}
	if applied := battle.applyPositiveDamage(&target, 17); applied != 5 || target.DownOverkill != 12 {
		t.Fatalf("applied=%d overkill=%d, want 5/12", applied, target.DownOverkill)
	}
	if applied := battle.applyPositiveDamage(&target, 9); applied != 0 || target.DownOverkill != 12 {
		t.Fatalf("追擊改寫了溢出：applied=%d overkill=%d", applied, target.DownOverkill)
	}
	exact := Fighter{HitPoints: 4, MaxHitPoints: 8}
	if applied := battle.applyPositiveDamage(&exact, 4); applied != 4 || exact.DownOverkill != 0 {
		t.Fatalf("剛好歸零：applied=%d overkill=%d, want 4/0", applied, exact.DownOverkill)
	}
	alive := Fighter{HitPoints: 6, MaxHitPoints: 8}
	if applied := battle.applyPositiveDamage(&alive, 2); applied != 2 || alive.DownOverkill != 0 {
		t.Fatalf("沒倒不記：applied=%d overkill=%d", applied, alive.DownOverkill)
	}
}
