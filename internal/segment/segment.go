// Package segment 是主線分段的註冊表：把 25 個 ECL block 各自的「直接進入所需
// 資料」集中在一處，讓 `-segment <id>` 與分段測試用同一份宣告。
//
// 段的 id 一律是 `ECL{成員}/0x{block}`（機械且穩定，與 game pack 的地圖命名
// 無關）。人類可讀的標籤在 `docs/plan/segment-labels.json`，轉移圖在
// `docs/audit/ecl-block-graph.md`，兩者都由測試與本表對齊。
package segment

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Segment 描述一段主線：它是哪個 block、直接進入時要把 LastECL 設成誰、
// 用哪一組 GEO 檔、以及進去之後玩家站在地城還是世界地圖上。
type Segment struct {
	// ID 是 `ECL{成員}/0x{block}`。
	ID string
	// Member 是 ECL DAX 成員編號（1..6）。
	Member uint8
	// Block 是 ECL block 編號，全六個成員之間不重複。
	Block uint8
	// EnterFrom 是直接進入時寫進 `4BF2h`（LastECL）的值。取自轉移圖的
	// 「進入自」欄位裡最順著劇情的那一個；0x00 代表「全新開局」——
	// 引擎主迴圈在 LastECL ＝ 0 時才會走到開場與提爾佛頓第一段（spec 1141）。
	EnterFrom uint8
	// GameArea 是這一段的章節編號，一律等於 ECL 成員編號。它同時決定 GEO 檔集
	// 與圖片素材要從哪一組 DAX 取，所以世界地圖上的段也要有值——ECL1 的段就是
	// 1，落成別的數字時開場那張圖會去對不存在的檔案。
	GameArea uint8
	// Overland 為真代表這一段是世界地圖上的段落，不是第一人稱地城。
	Overland bool
	// LegacyFlag 是既有的專用旗標名稱（沒有就留空）。`-segment` 收這個字串
	// 當別名；⚠ 專用旗標通常還會把隊伍走到段內某一格，直接進入只到段的入口。
	LegacyFlag string
	// SettlesAt 是 initial lifecycle 跑完之後實際停住的 block。0 代表就停在
	// 自己身上，也就是絕大多數段落的情況；非 0 的段是**過場 block**：進去之後
	// 無條件轉走，停不下來。
	SettlesAt uint8
}

// Settles 回傳這一段跑完 initial lifecycle 之後應該停在哪個 block。
func (s Segment) Settles() uint8 {
	if s.SettlesAt != 0 {
		return s.SettlesAt
	}
	return s.Block
}

// registry 依 id 排序，順序即輸出順序。
var registry = []Segment{
	{ID: "ECL1/0x50", Member: 1, Block: 0x50, EnterFrom: 0x51, GameArea: 1, Overland: true},
	{ID: "ECL1/0x51", Member: 1, Block: 0x51, EnterFrom: 0x50, GameArea: 1, Overland: true},
	// ⚠ `-opening` 進的是 0x01 不是這裡：remake 的新遊戲流程 BeginAdventure
	// 直接 reset 到 0x01，沒有經過開場 block。
	{ID: "ECL1/0x52", Member: 1, Block: 0x52, EnterFrom: 0x00, GameArea: 1, Overland: true},
	{ID: "ECL2/0x01", Member: 2, Block: 0x01, EnterFrom: 0x00, GameArea: 2, LegacyFlag: "opening"},
	{ID: "ECL2/0x02", Member: 2, Block: 0x02, EnterFrom: 0x01, GameArea: 2},
	{ID: "ECL2/0x03", Member: 2, Block: 0x03, EnterFrom: 0x02, GameArea: 2, LegacyFlag: "sewers"},
	{ID: "ECL2/0x04", Member: 2, Block: 0x04, EnterFrom: 0x03, GameArea: 2},
	{ID: "ECL3/0x10", Member: 3, Block: 0x10, EnterFrom: 0x51, GameArea: 3},
	{ID: "ECL3/0x11", Member: 3, Block: 0x11, EnterFrom: 0x10, GameArea: 3},
	{ID: "ECL3/0x12", Member: 3, Block: 0x12, EnterFrom: 0x11, GameArea: 3},
	{ID: "ECL3/0x15", Member: 3, Block: 0x15, EnterFrom: 0x51, GameArea: 3},
	{ID: "ECL4/0x20", Member: 4, Block: 0x20, EnterFrom: 0x51, GameArea: 4},
	{ID: "ECL4/0x21", Member: 4, Block: 0x21, EnterFrom: 0x20, GameArea: 4},
	{ID: "ECL4/0x22", Member: 4, Block: 0x22, EnterFrom: 0x21, GameArea: 4},
	{ID: "ECL4/0x23", Member: 4, Block: 0x23, EnterFrom: 0x20, GameArea: 4},
	{ID: "ECL4/0x25", Member: 4, Block: 0x25, EnterFrom: 0x50, GameArea: 4},
	// 0x30 是過場 block：72 條可達指令、唯一的出邊是 0x50，initial lifecycle
	// 一跑就把隊伍送回世界地圖，停不在自己身上。
	{ID: "ECL5/0x30", Member: 5, Block: 0x30, EnterFrom: 0x33, GameArea: 5, SettlesAt: 0x50},
	{ID: "ECL5/0x31", Member: 5, Block: 0x31, EnterFrom: 0x50, GameArea: 5},
	{ID: "ECL5/0x32", Member: 5, Block: 0x32, EnterFrom: 0x31, GameArea: 5, LegacyFlag: "lava-tube"},
	{ID: "ECL5/0x33", Member: 5, Block: 0x33, EnterFrom: 0x32, GameArea: 5, LegacyFlag: "wizard-tower"},
	{ID: "ECL5/0x35", Member: 5, Block: 0x35, EnterFrom: 0x50, GameArea: 5},
	{ID: "ECL6/0x40", Member: 6, Block: 0x40, EnterFrom: 0x50, GameArea: 6, LegacyFlag: "burial-red-web"},
	{ID: "ECL6/0x42", Member: 6, Block: 0x42, EnterFrom: 0x40, GameArea: 6},
	{ID: "ECL6/0x43", Member: 6, Block: 0x43, EnterFrom: 0x42, GameArea: 6, LegacyFlag: "inner-ritual"},
	{ID: "ECL6/0x45", Member: 6, Block: 0x45, EnterFrom: 0x51, GameArea: 6},
}

// All 回傳註冊表的副本，順序固定。
func All() []Segment {
	out := make([]Segment, len(registry))
	copy(out, registry)
	return out
}

// FormatID 用成員與 block 編號組出段的 id。
func FormatID(member, block uint8) string {
	return fmt.Sprintf("ECL%d/0x%02X", member, block)
}

// Lookup 接受三種寫法：完整 id（`ECL5/0x32`）、只給 block（`0x32`／`32`／`50`）、
// 或既有專用旗標的名字（`lava-tube`）。大小寫與前後空白都不計。
func Lookup(name string) (Segment, bool) {
	key := strings.TrimSpace(name)
	if key == "" {
		return Segment{}, false
	}
	lower := strings.ToLower(key)
	for _, candidate := range registry {
		if strings.EqualFold(candidate.ID, key) {
			return candidate, true
		}
		if candidate.LegacyFlag != "" && candidate.LegacyFlag == lower {
			return candidate, true
		}
	}
	// 只給 block 編號：十六進位優先（原作的 block 都以 16 進位稱呼）。
	digits := strings.TrimPrefix(strings.TrimPrefix(lower, "0x"), "block-")
	if value, err := strconv.ParseUint(digits, 16, 8); err == nil {
		for _, candidate := range registry {
			if candidate.Block == uint8(value) {
				return candidate, true
			}
		}
	}
	return Segment{}, false
}

// Names 回傳可以餵給 Lookup 的全部字串，排序後輸出，供 `-segment list` 使用。
func Names() []string {
	names := make([]string, 0, len(registry)*2)
	for _, candidate := range registry {
		names = append(names, candidate.ID)
		if candidate.LegacyFlag != "" {
			names = append(names, candidate.LegacyFlag)
		}
	}
	sort.Strings(names)
	return names
}
