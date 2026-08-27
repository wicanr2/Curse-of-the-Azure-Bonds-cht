# 1218 — 移除「同一 session 沒有數字」的假缺口

狀態：`READY`（限 remake 狀態表與分段交接證據的現在式訂正）

## 被推翻的舊敘述

`cmd/remake-status/rows.json` 原本同時存在兩個互相衝突的結論：

- measured 列記錄按鍵驅動 session 在第 14,380 幀由標題觸發 `GameWon()`；
- unmeasured 列卻仍把「開場到結局的同一 session」列成沒有數字，並說缺少剛換段
  快照與六人直入對照。

後者是舊 checkpoint，不能再作 remake 待辦。

## 現有證據

- `docs/audit/key-driven-session.json`：同一個 `app` session 從標題開始，只經
  `(*app).Update()` 按鍵路徑，在第 14,380 幀觸發 `GameWon()`；它使用強化隊伍，
  因而只證明輸入層與連續內容路徑，不證明一般強度可通關。
- `docs/audit/segment-handoff.md`：20 段可比，19 段在剛換段瞬間由
  `COAB_ARRIVAL_SNAPSHOT_DIR` 取樣；唯一較晚取樣的是全新開局，本來沒有上一段。
- 同一報表已量過一人與六人直入，ECL 記憶體差異格數逐段完全相同。
- `State.SeedHandoffMemory` 鋪回交接後，39 格可讀殘差全部可歸因，未歸因為 0。

## 狀態表訂正

- 移除重複且錯誤的「開場到結局的同一 session／沒有數字」列。
- `segment-seams` 只宣稱 `LastECL` 對照；明確連到下一列的整包 ECL 記憶體證據。
- **2026-08-27 訂正**：一般強度自動 session 已修掉 pending Bless 無限重試；其後
  2,455 幀走過 251 格／8 段／199 句、落回原文 0，最後發生真實全滅。使用者已決定
  真實全滅是合法遊戲結果，測試遇到即停止並記錄，不再把「必須自動通關」當 gate；
  強化隊伍的連續 session 仍只證明輸入層可達性。
- 三平台封包已可重生；Windows／macOS 真機啟動與代表性人工試玩仍是發行驗收事項，
  不再沿用「等待遊戲完整可玩」的舊前提。
