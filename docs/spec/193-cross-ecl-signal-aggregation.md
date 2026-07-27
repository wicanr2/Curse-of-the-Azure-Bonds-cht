# 第一百九十三輪：跨 ECL signal aggregation

狀態：`READY`

## 證據與問題

`BlockSession.runFromSeed` 會在 `NEWECL` 後以新的 block 繼續 bounded VM，但原本只合併 combat、monster、NPC、random、encounter 與 `LOAD PIECES` 結果。`LOAD FILES`、`PICTURE`、`SPELL`、`PROTECTION` 若發生在前一個 block，離開該 block 後就不會抵達 game state。

## 實作 contract

- `LoadFilesRequested` 是 session-level latch；最近一次 request 的三個 selectors 會保留。
- `PictureRequested` 也是 session-level latch；picture block 與 BIGPIC 分流旗標在 request 發生時保存。
- `SpellSearches` 與 `ProtectionRequests` 依 block 執行順序 append，讓上層 adapter 維持原始搜尋順序。
- VM 仍只回傳 renderer／party-neutral signal；session 不直接載入圖片、地圖或施法效果。

## Regression

`TestBlockSessionAggregatesSignalsAcrossNewECL` 以三個 synthetic block 驗證：第一個 block 轉到第二個，第二個發出 `PICTURE` 再轉到第三個，第三個發出 `LOAD FILES`；最終結果仍同時包含兩類 request，且 current block 已到第三個 block。

本輪在 Docker 內通過 `go test ./internal/ecl`。完整 repository 測試另受容器沒有 ALSA／X11 開發標頭，以及既存 `internal/game` journey regression 影響，未宣稱全套件通過。

## 可沿用知識

後續 Golden Box 遊戲可以沿用 `RunResult` signal boundary 與 `BlockSession` aggregation，但必須各自提供 file／picture catalog、spell record lookup 與 effect rules；不可把 request signal 當成 side effect 已完成。
