package gamepack

import (
	"encoding/json"
	"fmt"
	"sync"
)

// 第一人稱牆面符號的「第 0 段」（spec 1131）。
//
// `Put8x8Symbol` 把符號編號切成五段，各自有自己的資源指標（spec 781）：
// `1..45`、`46..115`、`116..185`、`186..255`、`256..295`。中間三段由
// `LOAD PIECES` 選到的三個 8X8D 區塊供應，**第一段是全遊戲共用的一份**——
// 牆面的斜角收邊、天空與地板都在裡面。這一段先前沒有被載入，於是所有
// 低編號的磚都畫不出來，走廊看起來像幾根分離的柱子。
//
// ⚠ 檔名與區塊編號是 CoAB 這一款的事實，所以宣告放在遊戲這一側，
// 不進共用 engine 的 pack schema。
type WallSymbolGroup struct {
	File      string `json:"file"`
	Block     uint8  `json:"block"`
	FirstID   uint8  `json:"first_id"`
	ItemCount int    `json:"item_count"`
}

// WallSymbols 是整份宣告。
type WallSymbols struct {
	SchemaVersion int             `json:"schema_version"`
	Source        string          `json:"source"`
	Spec          string          `json:"spec"`
	Note          string          `json:"note"`
	SharedGroup   WallSymbolGroup `json:"shared_group"`
}

var (
	wallSymbolsOnce sync.Once
	wallSymbols     *WallSymbols
	wallSymbolsErr  error
)

// LoadWallSymbols 回傳嵌入的牆面符號宣告。
func LoadWallSymbols() (*WallSymbols, error) {
	wallSymbolsOnce.Do(func() {
		data, err := ruleFiles.ReadFile("rules/wall-symbols.json")
		if err != nil {
			wallSymbolsErr = fmt.Errorf("read embedded wall symbols: %w", err)
			return
		}
		parsed := &WallSymbols{}
		if err := json.Unmarshal(data, parsed); err != nil {
			wallSymbolsErr = fmt.Errorf("parse embedded wall symbols: %w", err)
			return
		}
		group := parsed.SharedGroup
		if group.File == "" || group.FirstID == 0 || group.ItemCount <= 0 {
			wallSymbolsErr = fmt.Errorf("wall symbol shared group needs a file, a nonzero first_id and a positive item_count")
			return
		}
		wallSymbols = parsed
	})
	return wallSymbols, wallSymbolsErr
}
