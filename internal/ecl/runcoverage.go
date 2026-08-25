package ecl

// 實跑指令覆蓋記錄器（`COAB_ECL_COVERAGE`）。
//
// ★ 存在的理由：remake-status 的「全城市／全房間走訪（地形分派以外的部分）」
// 一直沒有分母。任何玩家看得到的內容都是被執行的 ECL 指令，所以「實跑路線
// 執行過哪些指令」對上 `cmd/ecl-effect-coverage` 的可達指令集，就是那個分母
// ——序章、支線、可選房間、查表分派的分支全部涵蓋，不用一種內容一種盤點。
//
// ⚠ 預設關閉、零成本：沒有設 `COAB_ECL_COVERAGE` 時只有一個 nil 檢查。
// ⚠ 「首見即追加」：每個 (block, 指令位移) 只寫一行，行格式 `0xBB,0xOOOO`。
//   追加寫檔讓多個測試行程共用同一份輸出，不需要行程結束掛勾；
//   重跑前要自己刪檔，不然舊行會混進來。
// ⚠ 只記 `RuntimeState.CurrentBlock` 非 0 的執行：合成 fixture（自組 bytes、
//   沒有 BlockSession）的 runtime 是 nil 或 block 0，混進來會把假內容記成
//   「實跑蓋到」。真實 corpus 的 block ID 沒有 0。

import (
	"fmt"
	"os"
	"sync"
)

var (
	runCoverageMu   sync.Mutex
	runCoverageSeen map[[2]int]struct{}
	runCoverageFile *os.File
	runCoveragePath = os.Getenv("COAB_ECL_COVERAGE")
)

// recordRunCoverage 記下「block 的這個指令位移被執行過」。
func recordRunCoverage(runtime *RuntimeState, pc int) {
	if runCoveragePath == "" || runtime == nil || runtime.CurrentBlock == 0 {
		return
	}
	key := [2]int{int(runtime.CurrentBlock), pc}
	runCoverageMu.Lock()
	defer runCoverageMu.Unlock()
	if runCoverageSeen == nil {
		runCoverageSeen = make(map[[2]int]struct{})
	}
	if _, seen := runCoverageSeen[key]; seen {
		return
	}
	runCoverageSeen[key] = struct{}{}
	if runCoverageFile == nil {
		file, err := os.OpenFile(runCoveragePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			// 記錄器壞掉不該弄壞遊戲；關掉並讓後續呼叫直接返回。
			runCoveragePath = ""
			return
		}
		runCoverageFile = file
	}
	fmt.Fprintf(runCoverageFile, "0x%02X,0x%04X\n", runtime.CurrentBlock, pc)
}
