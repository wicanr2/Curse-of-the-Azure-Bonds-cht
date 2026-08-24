package main

import (
	"archive/zip"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/gamepack"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// TestDeclaredExitsSitOnCellsYouCanActuallyLeaveFrom 擋住第 752 輪那一類缺陷。
//
// ★ 為什麼要有這條。 `State.CanMoveDungeon` 對**已宣告**的 `external_exit`
// 直接回 true，**不看 GEO** ⇒ 把出口宣告在一格走不出去的邊界上，測試照樣綠，
// 而正常玩家永遠走不到。下水道往火刀據點的交接被宣告在 `(8,15,S)` 上，
// 而 `(8,15)` 的移動遮罩是 `3`（只有 N、E 通），這個錯活了七十輪（spec 1199）。
func TestDeclaredExitsSitOnCellsYouCanActuallyLeaveFrom(t *testing.T) {
	pack, err := gamepack.Default()
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	checked := 0
	for _, definition := range pack.Maps {
		if definition.Kind != "first_person" || definition.GeometryFile == "" {
			continue
		}
		if len(definition.ExternalExits) == 0 {
			continue
		}
		grid, gridErr := loadGrid(archive, definition.GeometryFile, definition.GeometryBlock)
		if gridErr != nil {
			t.Fatalf("%s：%v", definition.ID, gridErr)
		}
		for _, exit := range definition.ExternalExits {
			checked++
			direction := int(exit.Direction)
			deltaX, deltaY := 0, 0
			switch direction {
			case 0:
				deltaY = -1
			case 2:
				deltaX = 1
			case 4:
				deltaY = 1
			case 6:
				deltaX = -1
			default:
				t.Errorf("%s：出口 %s 的方向 %d 不是四個基本方向",
					definition.ID, exit.ID, direction)
				continue
			}
			nextX, nextY := exit.X+deltaX, exit.Y+deltaY
			if nextX >= 0 && nextX < geo.Width && nextY >= 0 && nextY < geo.Height {
				t.Errorf("%s：出口 %s 在 (%d,%d) 往 %d 走**沒有**離開地圖",
					definition.ID, exit.ID, exit.X, exit.Y, direction)
				continue
			}
			if !grid.CanMoveDungeonWrapped(exit.X, exit.Y, direction) {
				wall, _ := grid.WallWrapped(exit.X, exit.Y, direction)
				t.Errorf("%s：出口 %s 在 (%d,%d) 往 %d 是**牆**（牆碼 %d）"+
					"——宣告會把牆蓋掉，玩家走不到",
					definition.ID, exit.ID, exit.X, exit.Y, direction, wall)
			}
		}
	}
	if checked == 0 {
		t.Fatal("一個 external_exit 都沒查到：這條測試等於沒跑")
	}
	t.Logf("查過 %d 個宣告的出口", checked)
}
