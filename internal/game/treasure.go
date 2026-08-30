package game

import (
	"fmt"

	"github.com/wicanr2/golden-box-remake-engine/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/monster"
)

// ParseTreasureItemBlocks decodes ITEM1.DAX through ITEM6.DAX members. The
// returned key is (area << 8) | DAX block ID, matching TREASURE's area-local
// item-block operand while allowing all six chapters to share one map.
//
// ⚠ 走 `ParseOriginalItems`：`ITEM*.DAX` 的位元組全部來自原版，物品名要以原版
// 編碼解讀（spec 1121）。remake 自己的 sidecar 走 `ParseItems`，兩者版面相同、
// 編碼不同，分不出來的是來源不是格式。
func ParseTreasureItemBlocks(areaData map[uint8][]byte) (map[uint16][]monster.ItemRecord, error) {
	all := make(map[uint16][]monster.ItemRecord)
	for area, data := range areaData {
		blocks, err := dax.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("parse ITEM%d.DAX: %w", area, err)
		}
		for _, block := range blocks {
			items, err := monster.ParseOriginalItems(block.Data)
			if err != nil {
				return nil, fmt.Errorf("parse ITEM%d.DAX block 0x%02X: %w", area, block.Entry.ID, err)
			}
			key := uint16(area)<<8 | uint16(block.Entry.ID)
			if _, exists := all[key]; exists {
				return nil, fmt.Errorf("duplicate ITEM area/block 0x%02X/%02X", area, block.Entry.ID)
			}
			all[key] = items
		}
	}
	return all, nil
}
