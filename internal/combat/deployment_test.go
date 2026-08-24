package combat

import "testing"

// 朝南遭遇、距離 1：敵隊起點 (teamX,teamY)=(0,1)、方向組 0（形狀 0），
// 我方方向組 2（形狀 2）。首幾個落點由 spec 1200 的表手算而得：
// 種子格 = 6F2h/6FAh 表、掃描軸 = 6EEh[(scan±k) mod 4]。
func TestDeploymentSouthEncounterFirstPlacements(t *testing.T) {
	d := NewDeployment(4, 1, 4, 4)

	wantEnemy := []TilePoint{{32, 18}, {33, 18}, {31, 18}, {34, 18}}
	for i, want := range wantEnemy {
		x, y, ok := d.Place(1, nil, nil)
		if !ok || x != want.X || y != want.Y {
			t.Fatalf("敵 %d：(%d,%d,%v)，手算是 %v", i, x, y, ok, want)
		}
	}
	wantParty := []TilePoint{{27, 12}, {26, 12}, {28, 12}, {25, 12}}
	for i, want := range wantParty {
		x, y, ok := d.Place(0, nil, nil)
		if !ok || x != want.X || y != want.Y {
			t.Fatalf("我 %d：(%d,%d,%v)，手算是 %v", i, x, y, ok, want)
		}
	}
}

// 座標公式：x = teamY×5 + teamX×6 + 欄 + 22、y = teamY×5 + 列 + 10（spec 1200 §三）。
// 朝東、距離 2：敵隊 (teamX,teamY) = (2,0) ⇒ 敵我 x 相差 12。
func TestDeploymentEastEncounterSeparation(t *testing.T) {
	d := NewDeployment(2, 2, 1, 1)
	px, py, ok := d.Place(0, nil, nil)
	if !ok {
		t.Fatal("我方放不進去")
	}
	ex, ey, ok := d.Place(1, nil, nil)
	if !ok {
		t.Fatal("敵方放不進去")
	}
	// 我方形狀 1（東）、敵方形狀 3（西）；兩隊的欄種子不同，但 teamX×6 的
	// 位移一定完整反映在 x 差上。
	if ex-px != 2*6+(deploySeedCol[0][3]-deploySeedCol[0][1]) {
		t.Fatalf("東向距離 2 的分離錯了：我 (%d,%d) 敵 (%d,%d)", px, py, ex, ey)
	}
	if ey != py+(deploySeedRow[0][3]-deploySeedRow[0][1]) {
		t.Fatalf("東向的列不該再位移：我 (%d,%d) 敵 (%d,%d)", px, py, ex, ey)
	}
}

// 地面檢查擋掉種子格時要掃到下一個候選，不是硬放。
func TestDeploymentRespectsGround(t *testing.T) {
	d := NewDeployment(4, 1, 4, 4)
	rejected := map[TilePoint]bool{{32, 18}: true}
	ground := func(x, y int) bool { return !rejected[TilePoint{x, y}] }
	x, y, ok := d.Place(1, ground, nil)
	if !ok || (x == 32 && y == 18) {
		t.Fatalf("種子格被地面擋掉之後放到 (%d,%d,%v)", x, y, ok)
	}
	if x != 33 || y != 18 {
		t.Fatalf("下一個候選應是 (33,18)，實際 (%d,%d)", x, y)
	}
}

// 同一隊塞不下時要回報放棄（原作把放不下的戰鬥員從戰鬥移除），不能吊死。
// 方向 0 的敵隊是形狀 2＋4＋2＋2，四組共 23+46+23+23 = 115 格。
func TestDeploymentOverflowGivesUp(t *testing.T) {
	d := NewDeployment(0, 1, 6, 40)
	placed, dropped := 0, 0
	for i := 0; i < 200; i++ {
		if _, _, ok := d.Place(1, nil, nil); ok {
			placed++
		} else {
			dropped++
		}
	}
	if dropped == 0 {
		t.Fatalf("200 隻：放進 %d、放棄 %d——塞爆之後應該開始放棄", placed, dropped)
	}
	if placed > 115 {
		t.Fatalf("放進 %d 超過四組模板的總格數 115", placed)
	}
}

// 八個方向、兩隊都要能連放而且不重複、不吊死。
//
// ⚠ 不斷言「落點都在戰場圖內」：公式對北向的遠距遭遇本來就會給出負座標
// （teamY = 距離 × (−1)），原作是靠遭遇輸入的實際分佈避開；界內與否由
// 呼叫端的地面檢查決定（原作是 TACMAP 讀圖）。
func TestDeploymentPlacementsAreDistinct(t *testing.T) {
	for direction := uint8(0); direction < 8; direction++ {
		d := NewDeployment(direction, 3, 6, 12)
		for team := 0; team <= 1; team++ {
			seen := map[TilePoint]bool{}
			for i := 0; i < 12; i++ {
				x, y, ok := d.Place(team, nil, nil)
				if !ok {
					break
				}
				p := TilePoint{x, y}
				if seen[p] {
					t.Fatalf("方向 %d 隊 %d 第 %d 個落點 (%d,%d) 重複了",
						direction, team, i, x, y)
				}
				seen[p] = true
			}
			if len(seen) == 0 {
				t.Fatalf("方向 %d 隊 %d 一個都放不進去", direction, team)
			}
		}
	}
}
