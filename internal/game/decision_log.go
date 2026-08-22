package game

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

// Decision 是一次玩家層級的選擇：面對哪一組選項、選了第幾項。
//
// ★ 存在的理由：「開場到結局」缺的是**路線知識**——戰役測試裡逐段寫死的劇情決策。
// 那份知識散在三千多行裡，手抄過去只會抄錯。與其抄，不如**錄**：`Select` 是所有
// 選擇的唯一入口，在那裡記下「當時的選項」與「選了哪一項」，就得到一份可重放的
// 路線，而重放端可以用**按鍵**把游標挪到那一項（spec 1191）。
//
// ⚠ 記的是**選項文字**不是索引：索引會隨著選單內容改變而錯位，而重放時要先確認
// 「我現在面對的是同一個選單」才敢照著按。
type Decision struct {
	// Kind 是 `select`（選單）或 `move`（地城走一步）。
	//
	// ★ 兩種都要錄：只錄選單的話，重放端會**卡在原地按選單**——實測 783 步的
	// 路線只用得到 13 步，因為主線的決策點要**走到那裡**才會出現，而「走到哪裡」
	// 本身就是路線知識的一半（spec 1191）。
	Kind    string   `json:"kind"`
	Mode    int      `json:"mode"`
	Choices []string `json:"choices"`
	Index   int      `json:"index"`
	// Chosen 是選中那一項的文字，重放時用來核對。
	Chosen string `json:"chosen"`
	// Segment 是當下的 ECL 段。★ 重放跟丟時要靠它**重新對齊**：光有座標不夠，
	// 不同段的地圖上都有 (7,13)。
	Segment   int `json:"segment,omitempty"`
	// 以下只有 `move` 用：從哪一格、往哪個方向走。
	FromX     int `json:"from_x,omitempty"`
	FromY     int `json:"from_y,omitempty"`
	Direction int `json:"direction,omitempty"`
}

var (
	decisionMu  sync.Mutex
	decisionLog []Decision
	// decisionRecording 由 `COAB_DECISION_LOG` 開啟。⚠ 預設關閉：正式執行路徑
	// 不該為了盤點而付出任何代價。
	decisionRecording = os.Getenv("COAB_DECISION_LOG") != ""
)

// recordDecision 記下一次選擇。只有開啟錄製時才做事。
func recordDecision(mode Mode, choices []string, index int) {
	if !decisionRecording || index < 0 || index >= len(choices) {
		return
	}
	decisionMu.Lock()
	defer decisionMu.Unlock()
	decisionLog = append(decisionLog, Decision{
		Kind:    "select",
		Mode:    int(mode),
		Choices: append([]string(nil), choices...),
		Index:   index,
		Chosen:  strings.TrimSpace(choices[index]),
	})
}

// WriteDecisionLog 把錄到的路線寫成 JSON。路徑留白就用 `COAB_DECISION_LOG`。
func WriteDecisionLog(path string) error {
	if path == "" {
		path = os.Getenv("COAB_DECISION_LOG")
	}
	if path == "" {
		return nil
	}
	decisionMu.Lock()
	defer decisionMu.Unlock()
	encoded, err := json.MarshalIndent(decisionLog, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o644)
}

// DecisionLogLength 回報目前錄到幾次選擇。
func DecisionLogLength() int {
	decisionMu.Lock()
	defer decisionMu.Unlock()
	return len(decisionLog)
}

// recordMove 記下地城裡走的一步：從哪一格、往哪個方向。
//
// ⚠ 記**起點與方向**不是終點：重放端要先轉到那個方向才踏得出去，而樓梯事件正是
// 「站對方向踏上去」才觸發的（spec 1193）。
func recordMove(segment, fromX, fromY, direction int) {
	if !decisionRecording {
		return
	}
	decisionMu.Lock()
	defer decisionMu.Unlock()
	decisionLog = append(decisionLog, Decision{
		Kind: "move", Segment: segment, FromX: fromX, FromY: fromY, Direction: direction,
	})
}
