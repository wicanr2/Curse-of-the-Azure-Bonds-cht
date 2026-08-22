package main

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/geo"
)

// 這一支釘住「移動遮罩分得開每一張圖」。
//
// ★ 為什麼重要。 它是目前**唯一**一把與第一人稱算圖無關的尺：拿它判定原版現在
// 站在哪一張圖，才能把跨圖的擷取收進索引。分不開的話，跨圖擷取就只能靠「跟
// remake 比對、像就算數」——那是循環論證（spec 1185）。
//
// ⚠ 分不開的時候要**加格子**，不是改斷言。兩格夠用是量出來的結果，不是前提。
func TestDungeonMoveMasksIdentifyEveryMap(t *testing.T) {
	catalog := loadCatalog(t)
	refs := catalog.Refs()
	if len(refs) < 16 {
		t.Fatalf("目錄裡只有 %d 張圖，原版是 16 張", len(refs))
	}
	probes := [][2]int{{3, 3}, {12, 9}}
	seen := map[string]string{}
	for _, ref := range refs {
		grid, ok := catalog.Lookup(ref)
		if !ok {
			t.Fatalf("查不到 GEO%d 段 0x%02X", ref.Set, ref.BlockID)
		}
		parts := make([]string, 0, len(probes))
		for _, cell := range probes {
			parts = append(parts, fmt.Sprintf("%X", mask(grid, cell[0], cell[1])))
		}
		signature := strings.Join(parts, "-")
		name := fmt.Sprintf("GEO%d/0x%02X", ref.Set, ref.BlockID)
		if previous, clash := seen[signature]; clash {
			t.Errorf("%s 與 %s 的移動指紋相同（%s）：兩格不夠分辨，要加格子",
				name, previous, signature)
			continue
		}
		seen[signature] = name
	}
}

// 提爾佛頓 (7,13) 是實測過的定錨：四面都有牆（牆型 04/06/04/05），但**往西走得
// 出去**，所以遮罩是 `8`。原版側是按一次方向鍵、把座標讀回來量到的
// （`tools/dos-oracle-move-probe.sh`，第 681 輪）。
//
// ⚠ 這一格同時證明「有牆」不等於「走不動」。要是有人把遮罩改成用牆的有無算，
// 這支測試會紅。
func TestTilvertonStartCellMatchesTheMeasuredProbe(t *testing.T) {
	catalog := loadCatalog(t)
	grid, ok := catalog.Lookup(geo.MapRef{Set: 2, BlockID: 1})
	if !ok {
		t.Fatal("查不到提爾佛頓")
	}
	if got := mask(grid, 7, 13); got != 0x8 {
		t.Fatalf("(7,13) 的移動遮罩 ＝ %X，實機量到的是 8（只有西邊走得動）", got)
	}
}

func loadCatalog(t *testing.T) geo.Catalog {
	t.Helper()
	archive, err := zip.OpenReader("../../curseoftheazurebonds.zip")
	if err != nil {
		t.Skipf("沒有原版 image：%v", err)
	}
	defer archive.Close()
	catalog := geo.NewCatalog()
	for member := 1; member <= 6; member++ {
		name := fmt.Sprintf("GEO%d.DAX", member)
		for _, file := range archive.File {
			if !strings.EqualFold(file.Name, name) {
				continue
			}
			handle, openErr := file.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			payload, readErr := io.ReadAll(handle)
			handle.Close()
			if readErr != nil {
				t.Fatal(readErr)
			}
			if addErr := catalog.AddDAX(uint8(member), payload); addErr != nil {
				t.Fatalf("%s：%v", name, addErr)
			}
		}
	}
	return catalog
}
