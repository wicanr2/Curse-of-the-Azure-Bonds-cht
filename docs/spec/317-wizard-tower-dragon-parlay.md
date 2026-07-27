# 第三百一十七輪：法師塔與黑龍交涉

狀態：`READY`

## 反組譯證據

- ECL5 block `0x33 +0x0452` 的四項 ordinary vertical menu 以 selection
  index `3` 進入 `PARLAY WITH THE DRAGONS`，於 `+0x05E7` 顯示
  PICTURE `0x36` 後執行 `+0x05EA PARLAY`。
- PARLAY 的五個原始態度與寫入 `7F79` 的 mapping 依序為：
  - index 0 `HAUGHTY` → `1`；
  - index 1 `SLY` → `0`；
  - index 2 `MEEK` → `0`；
  - index 3 `NICE` → `0`；
  - index 4 `ABUSIVE` → `1`。
- `+0x05FB` 比較 `7F79==1`：成立時 GOTO `+0x04B9`，沿 ATTACK
  DRAGONS／FLEE 共用路徑建立 14 隻 MON5 `0x35` BLACK DRAGON。
- mapping 為 `0` 時，`+0x0606` 顯示 PICTURE `0x35`，龍群表示隊伍已證明
  沒有反龍族陰謀，留下隊伍處理與德拉坎德羅斯的爭端。PRINT RETURN 後
  GOTO `+0x0564`，與 ATTACK WIZARD 分支匯流：德拉坎德羅斯召來
  伊弗利特一名、黑暗精靈戰士兩名、黑暗精靈法師一名後逃下樓。
- 守軍戰勝後仍沿 `+0x0768` 安全屋頂文字與 `+0x07A2 EXIT` 返回地城。

## 引擎與中文契約

- `PARLAY` 是有五個 mapping operands 與 destination address 的可恢復 VM
  input command；State 必須把使用者 index 交回同一 session，不得用泛用
  「交涉結果待實作」文案攔截 ECL-backed menu。
- 同一態度 token 在不同 encounter 的 reaction 由當地 script operands 決定；
  不可把 SLY／NICE 永久定義為成功，也不可把 HAUGHTY／ABUSIVE 永久定義為戰鬥。
- 態度成功只證明龍群離場，仍須執行德拉坎德羅斯守軍；不得把 PARLAY
  success 當作整個 encounter 已結束。
- PICTURE 與原版龍／守軍 sprite 在 640×480 畫布 nearest-neighbor 整數放大；
  繁中正文使用 24px，compact HUD 使用 16px。

## 驗收

- real-image regression 鎖定五個態度對 `7F79` 的 `1/0/0/0/1` mapping。
- 從規格 315 真實四選單選 index 3，再選 SLY，驗證說服龍群繁中敘事，
  接著建立規格 316 的四名守軍。
- deterministic victory 後驗證安全屋頂 PRINT RETURN 與 block `0x33`
  dungeon return。
- 另驗證 HAUGHTY 至少會進入 14 隻黑龍戰，不把 hostile mapping
  錯接到守軍。
- `-wizard-tower-parlay` 經正式 Ebiten／Xvfb 產生
  [`wizard-tower-dragon-parlay.png`](../screenshots/wizard-tower-dragon-parlay.png)；
  640×480 畫面同時顯示 nearest-neighbor 原版黑龍圖與 24px 繁中結果。
