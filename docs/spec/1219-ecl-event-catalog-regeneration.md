# 1219：ECL 事件目錄重生與階段台帳補齊

狀態：`READY`（2026-08-27）

## 問題

`docs/audit/ecl-event-catalog.*` 停在走訪器修正前的 4,222 條指令與
154 個候選，但當前 corpus 是 14,177 條與 701 個。正式重生又被兩道
失敗即關閉（fail-closed）閘門擋下：一個審查 ID 已位移，且階段台帳
少列六個當前可達 opcode。

## 證據與處理

- `ECL4.DAX/0x23/0x0130-0x019D` 因變長 `RecordEnd` 修正而成為
  `ECL4.DAX/0x23/0x0130-0x019A`。副作用序列未變，因此只更新精確審查鍵，
  不放寬比對。
- `1Eh CHECKPARTY` 與 `30h OR` 有逐條 DOS handler 證據，列為
  `immediate/exact`。
- `26h ON GOSUB`、`2Ch PARLAY`、`34h ECL CLOCK`、`3Bh SPELL` 只有
  handler、運算元或 remake 邊界證據，未證實原版提交次序；故列已補齊，
  commit phase 仍為 `unknown`。

## 結果

四份產物已由 `cmd/ecl-event-catalog` 正式重生：

- 6 個 member、25 個 block、125 個 lifecycle entry。
- 14,177 條靜態可達指令、701 個跨 effect-kind 候選。
- 階段台帳完整覆蓋 61 支可達 opcode；已讀 27 支，尚未讀 34 支。

本規格關閉的是清冊完整性與可重生性，不是 34 支未讀 handler 的原版
語意。這些列仍以 `unknown` 顯示，不影響玩家現行路徑，也不能用來宣稱
全域有序 transaction model 已完全閉合。
