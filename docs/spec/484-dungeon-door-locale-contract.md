# 第 484 輪：地城門鎖與探索 chrome locale 契約

狀態：`READY`

## 問題

正常地城探索的鎖門提示、ECL lifecycle 錯誤、撬鎖／Knock／撞門結果與兩條
操作列曾由 Ebiten 直接組合繁中。平台 adapter 因而同時理解門旗標、法術消耗、
雙側解鎖與翻譯，且資料分離稽核無法保證其他前端得到相同文字。

## READY 契約

1. `DungeonMessage` 是 CoAB State adapter 的 typed presentation identity；它不
   進入作品中立 engine，也不取代現有門鎖規則結果。
2. Ebiten 仍依原有 `flags=2／3`、`PickDungeonLock`、`ConsumeDungeonKnockSpell`、
   `BashDungeonDoor` 與 `UnlockDoorWrapped` 決定操作結果，再把 typed identity
   交給 State 解析 stable locale ID。
3. lifecycle error 與 Knock spell ID 是 runtime argument；錯誤全文和 `0x29`
   不得複製進 renderer 或另一份翻譯表。
4. 一次撬鎖、Knock slot 消耗、雙側解鎖、選單關閉、地城座標與聲音時機完全
   沿用既有流程。本輪只改玩家可見文字來源。
5. 正常地城鎖門操作列與探索操作列使用 typed `PlayerUILabel`；debug preview
   的研究診斷另行處理，不與正式玩家 chrome 混為同一狀態。

## 驗證

- 正式 catalog test 覆蓋九種 static dungeon result、lifecycle error、Knock
  spell ID 與兩條正常探索 label；期望值由 stable ID 動態解析。
- 既有 Pick／Knock／Bash 規則、door menu options、地城 lifecycle 與 Ebiten
  tests 由 Docker／Xvfb 正式 gate 重跑。
- Go 漢字稽核 `275→262`；frontend `63→50`、runtime 212、localization 0
  不變。

本輪沒有修改原版地城石框、WALLDEF、GEO、視窗幾何或 README 畫面，也不宣稱
所有門與外部 routine 已完整還原。
