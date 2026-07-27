# 第 261 輪：multi-class character creation options

狀態：`READY`（限 reference race/class list、level slots 與 primary-class projection）

## 已完成

- 依 reference `Gbl.RaceClasses` 將目前可驗證的 18 個 multi-class 組合加入繁中角色建立選單：矮人、精靈、侏儒、半精靈、半身人、半獸人各自使用原作限制。
- 建立角色保存 raw `RawClassID`、`ClassLevels[8]`；每個 multi-class option 以各組合的職業 level slot 設為 1。
- 現有 party／combat adapter 使用 reference primary-class projection，因此既有裝備、HP、戰鬥流程仍可運作；JSON 與 DOS writeback 不會丟失 raw class metadata。
- 建立選單總數由 22 個單職業選項擴充至 40 個單／多職業選項。

## 保留邊界

multi-class 的完整 spell list、THAC0／HP growth、獨立 class limits、alignment、training／duel-class 與多職業專用 icon／能力仍待接入；primary projection 不宣稱完整 AD&D multi-class rules。

## 回歸

測試驗證 40 個選項、18 個 multi-class entries、race/class validation、各組合至少兩個 class levels，以及建立後可進入既有 party state。
