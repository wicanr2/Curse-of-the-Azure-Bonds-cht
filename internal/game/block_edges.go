package game

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

// BlockEdge 是主線**真的**走出來的一次段落轉移：從哪一段換到哪一段。
//
// ★ 存在的理由：`AGENTS.md` §3 寫著「段與段之間的狀態交接本身就是一段」——
// debug 進入點注入的是**合成的**起始狀態，未必等於上一段真的跑出來的結束狀態，
// 所以兩端各自綠不等於接縫通過。要驗接縫，得先知道主線實際的接法是什麼；
// `internal/segment` 的 `EnterFrom` 是**宣告**，這裡錄的是**事實**（spec 1195）。
//
// ⚠ 這是**下界**：只錄得到經過 `requestMusicIfBlockChanged` 那個瓶頸的轉移。
// 那是換段時重新派曲的地方，所以漏掉的轉移**音樂也會跟著錯**——兩件事共用同一個
// 訊號，不是各自獨立的假設。
type BlockEdge struct {
	From  uint8 `json:"from"`
	To    uint8 `json:"to"`
	Count int   `json:"count"`
}

var (
	blockEdgeMu sync.Mutex
	blockEdges  = map[[2]uint8]int{}
	// blockEdgeRecording 由 `COAB_BLOCK_EDGES` 開啟。⚠ 預設關閉：正式執行路徑
	// 不該為了盤點而付出任何代價。
	blockEdgeRecording = os.Getenv("COAB_BLOCK_EDGES") != ""
)

func recordBlockEdge(from, to uint8) {
	if !blockEdgeRecording {
		return
	}
	blockEdgeMu.Lock()
	defer blockEdgeMu.Unlock()
	blockEdges[[2]uint8{from, to}]++
}

// BlockEdges 回傳錄到的轉移，依 (from, to) 排序。
func BlockEdges() []BlockEdge {
	blockEdgeMu.Lock()
	defer blockEdgeMu.Unlock()
	out := make([]BlockEdge, 0, len(blockEdges))
	for key, count := range blockEdges {
		out = append(out, BlockEdge{From: key[0], To: key[1], Count: count})
	}
	sort.Slice(out, func(left, right int) bool {
		if out[left].From != out[right].From {
			return out[left].From < out[right].From
		}
		return out[left].To < out[right].To
	})
	return out
}

// WriteBlockEdges 把錄到的轉移寫成 JSON。路徑留白就用 `COAB_BLOCK_EDGES`。
func WriteBlockEdges(path string) error {
	if path == "" {
		path = os.Getenv("COAB_BLOCK_EDGES")
	}
	if path == "" {
		return nil
	}
	encoded, err := json.MarshalIndent(BlockEdges(), "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o644)
}
