package game

import (
	"archive/zip"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// 這一條釘住一個**很容易被重新推導錯**的事實：類別表的槽 11／12 是**卷軸**，
// 不是彈藥；箭與弩矢在槽 10，而槽 10 是個混裝格（還有藥水、飾品）。
//
// ★ 為什麼要有它。 spec 1120 曾經由「角色記錄的 `+17Dh`／`+181h` ＝ 裝備區塊
// 第 11、12 格」推成「彈藥 ＝ 類別表槽 11 或 12」——**偏移量算出來的陣列索引，
// 不等於資料本身的分類欄位**。那個結論寫進了 `AmmunitionCount` 的挑選條件，
// 挑到的是卷軸；不會噴錯，也不會讓任何測試變紅（`capByAmmunition` 把 0 當
// 「不設限」），只是箭的數量從此不再限制射擊次數。
//
// spec 1249 已回填 spec 1000 的 exact producer：類別 49h／1Ch 才會建立兩個
// 彈藥指標。這一條繼續擋住「再一次從槽 11／12 出發」。
func TestAmmunitionIsNotIdentifiedByBaseItemSlot(t *testing.T) {
	image, err := zip.OpenReader(filepath.Join("..", "..", "curseoftheazurebonds.zip"))
	if err != nil {
		t.Skipf("original image is unavailable: %v", err)
	}
	defer image.Close()
	catalog, err := monster.ParseBaseItems(zipData(t, image, "ITEMS"))
	if err != nil {
		t.Fatal(err)
	}

	namesByType := map[uint8][]string{}
	for chapter := 1; chapter <= 6; chapter++ {
		blocks, parseErr := dax.Parse(zipData(t, image, "ITEM"+strconv.Itoa(chapter)+".DAX"))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		for _, block := range blocks {
			items, itemErr := monster.ParseOriginalItems(block.Data)
			if itemErr != nil {
				continue
			}
			for _, item := range items {
				namesByType[item.Type] = append(namesByType[item.Type], item.Name)
			}
		}
	}
	// ⚠ 先擋「資料非空」：物品整批讀不到的話，下面每一條都會正確地通過。
	if len(namesByType) < 50 {
		t.Fatalf("只讀到 %d 個物品類別，資料沒讀進來", len(namesByType))
	}

	slotOf := func(itemType uint8) uint8 {
		base, ok := catalog.Lookup(itemType)
		if !ok {
			t.Fatalf("類別 %02Xh 不在類別表裡", itemType)
		}
		return base.Slot
	}
	// 箭（49h）與弩矢（1Ch）在槽 10。
	for _, item := range []struct {
		itemType uint8
		contains string
	}{{0x49, "Arrow"}, {0x1C, "Quarrel"}} {
		if got := slotOf(item.itemType); got != 10 {
			t.Errorf("類別 %02Xh 的槽 ＝ %d，want 10", item.itemType, got)
		}
		if !anyContains(namesByType[item.itemType], item.contains) {
			t.Errorf("類別 %02Xh 的名字裡沒有 %q：%v",
				item.itemType, item.contains, namesByType[item.itemType])
		}
	}
	// 槽 11／12 是卷軸；它們不得再進 `AmmunitionCount`。
	for _, slot := range []uint8{11, 12} {
		found := false
		for _, base := range catalog.Items {
			if base.Slot != slot || len(namesByType[base.Type]) == 0 {
				continue
			}
			found = true
			if !anyContains(namesByType[base.Type], "Scroll") {
				t.Errorf("槽 %d 的類別 %02Xh 不是卷軸：%v",
					slot, base.Type, namesByType[base.Type])
			}
		}
		if !found {
			t.Errorf("槽 %d 在遊戲裡一件物品都沒有——這一條就驗不到東西了", slot)
		}
	}
}

func anyContains(list []string, want string) bool {
	for _, item := range list {
		if strings.Contains(item, want) {
			return true
		}
	}
	return false
}
