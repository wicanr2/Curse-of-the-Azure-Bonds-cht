# 第 297 輪：提爾佛頓下水道至火刀據點邊界（歷史規格）

狀態：`SUPERSEDED`（E2／`NEWECL 4` 邊界仍有效；正常玩家 handoff 尚未閉合）

## 歷史上仍有效的證據

- ECL2 block 3 的 per-turn entry 會檢查 boundary attempt 與 movement sentinel
  `0x7EC9`。
- 南側 E2 branch 呼叫 `0xC01E`、把 `Y` 寫成 `0`、將 `X` 減去 `2`，再執行
  `NEWECL 4`。
- ECL2 block 4 initial entry 會要求 `LOAD FILES 4,2,0xFF`、
  `LOAD PIECES 1,2,4`，並顯示火刀據點入口文字。

上述只證明「若 boundary 以原版條件成立，block 4 如何初始化」；它沒有證明
玩家如何從騎士事件後的 `(13,10)` 正常抵達 E2 `(8,15)`。

## 勘誤

舊版 `Verification` 曾寫成正式新遊戲 regression「crosses `(8,15)`」。實際測試
在第 515 輪以前是直接把 `State.DungeonX/Y/Direction` 寫成 `(8,15,S)`，再呼叫
`RunDungeonExitLifecycle`。那是 `coordinate-assisted`／`direct-entry` probe，
不是正常玩家路徑；該句已移除，不得再作完成證據。

## 目前契約與待補證據

- boundary transaction 仍應清除 stale `0x7EC9`、同步目前 GEO registers，再由
  ECL entry 0 決定是否進入 E2。
- State 不得直接選擇 block 4 或假造目標位置；真正的 producer／consumer 尚須由
  DOS 或 PC-98 movement／map service、ECL flag 與 runtime trace 閉合。
- block 4 的初始化資料已由 `docs/spec/515-fire-knife-map-position-transition.md`
  之後的現況保存，但第二個 `(8,15)` handoff 仍列在
  `docs/knowledge/golden-box-reverse-engineering-worklist.md` 的 P0。

所有後續正常路徑宣稱以第 516 輪 worklist 為準；本檔保留的是被推翻斷言的來源與
勘誤脈絡，不再作 `READY` 規格。
