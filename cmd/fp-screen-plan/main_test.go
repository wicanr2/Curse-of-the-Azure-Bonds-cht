package main

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
	enginegeo "github.com/wicanr2/golden-box-remake-engine/geometry"
)

// wallGrid 造一張測試地圖：`walls[y][x]` 是四個方向的牆型（北、東、南、西）。
func wallGrid(walls map[[2]int][4]uint8) geo.Grid {
	var grid geo.Grid
	for point, sides := range walls {
		// `WallDirections` 的索引順序是北、東、南、西（＝方向 0／2／4／6）。
		grid.Cells[point[1]][point[0]] = enginegeo.Cell{WallDirections: sides}
	}
	return grid
}

// ★ 遮蔽是這個分母能用的唯一原因。沒有它，17,684 張畫面得到 14,294 個「相異」
// 簽章——因為簽的是原始鄰域而不是看得到的東西，牆後面那幾排根本不會畫出來。
//
// 這一支釘住：牆後面的差異**不該**改變簽章。
func TestSignatureIgnoresWhatIsBehindAWall(t *testing.T) {
	// 站在 (5,5) 朝北，正前方 (5,5) 的北牆擋住視線。
	blocked := wallGrid(map[[2]int][4]uint8{{5, 5}: {1, 0, 0, 0}})
	// 同一張圖，但把牆後面那一格塞滿牆——玩家看不到，簽章要一樣。
	blockedPlusHidden := wallGrid(map[[2]int][4]uint8{
		{5, 5}: {1, 0, 0, 0},
		{5, 4}: {3, 3, 3, 3},
		{4, 4}: {3, 3, 3, 3},
		{6, 4}: {3, 3, 3, 3},
	})
	if a, b := signature(blocked, 5, 5, 0), signature(blockedPlusHidden, 5, 5, 0); a != b {
		t.Fatalf("牆後面的差異改變了簽章：\n%q\n%q", a, b)
	}
}

// 反過來：**看得到**的差異一定要改變簽章，否則兩張不同的畫面會被算成同一張。
func TestSignatureSeparatesVisibleDifferences(t *testing.T) {
	open := wallGrid(nil)
	frontWall := wallGrid(map[[2]int][4]uint8{{5, 4}: {1, 0, 0, 0}})
	if signature(open, 5, 5, 0) == signature(frontWall, 5, 5, 0) {
		t.Fatal("前方第二排多一面牆卻得到同一個簽章")
	}
	sideWall := wallGrid(map[[2]int][4]uint8{{5, 5}: {0, 0, 0, 1}})
	if signature(open, 5, 5, 0) == signature(sideWall, 5, 5, 0) {
		t.Fatal("腳下這一格多一面側牆卻得到同一個簽章")
	}
}

// ⚠ 簽章只記牆的**有無**，不記牆型。牆型是另一個軸（`LOADWALLSET` 與地圖宣告），
// 混進來會讓去重失效——第一版就是這樣得到 15,693 個「相異」簽章的。
func TestSignatureIgnoresWallType(t *testing.T) {
	typeOne := wallGrid(map[[2]int][4]uint8{{5, 5}: {1, 0, 0, 0}})
	typeSeven := wallGrid(map[[2]int][4]uint8{{5, 5}: {7, 0, 0, 0}})
	if signature(typeOne, 5, 5, 0) != signature(typeSeven, 5, 5, 0) {
		t.Fatal("牆型不同不該改變幾何簽章")
	}
	if len(wallTypes(typeOne)) != 1 || len(wallTypes(typeSeven)) != 1 {
		t.Fatal("牆型那一軸要各自數得出來")
	}
}

// 原作的地城是 16×16 環繞的，視野也會看到繞回來的那一側。
func TestSignatureWrapsLikeTheOriginal(t *testing.T) {
	// 站在 (0,0) 朝北，前方會繞到 y=15。
	wrapped := wallGrid(map[[2]int][4]uint8{{0, 15}: {1, 0, 0, 0}})
	if signature(wallGrid(nil), 0, 0, 0) == signature(wrapped, 0, 0, 0) {
		t.Fatal("繞回來那一側的牆沒有進視野")
	}
}

// ⚠ 四面都是牆**不代表**玩家不會站在那裡：開新遊戲的起點 `GEO2/0x01 (7,13)`
// 正是密室（劇情要求先找出口），而它是每個玩家看到的第一張畫面。所以這一支
// 只認「是不是密室」，**不拿它過濾分母**。
func TestSealedRecognisesFullyWalledCells(t *testing.T) {
	if !sealed(wallGrid(map[[2]int][4]uint8{{3, 3}: {1, 1, 1, 1}}), 3, 3) {
		t.Fatal("四面都是牆就是密室")
	}
	if sealed(wallGrid(map[[2]int][4]uint8{{3, 3}: {1, 1, 1, 0}}), 3, 3) {
		t.Fatal("有一面開著就不是密室")
	}
}
