# 第 481 輪：戰鬥 visual result locale 契約

狀態：`READY`

## 問題

Ebiten `combatVisualMessage` 曾直接組合 Cloudkill、Stinking Cloud 與 Lightning
Bolt 的七筆中文結果。Renderer 因而必須理解 effect 名稱、impact target、
豁免、元素防護及 visual phase；翻譯也與繪圖時間軸耦合。

## READY 契約

1. `combat.VisualEvent`／`VisualFrame` 仍是 renderer-neutral typed timeline；
   renderer 只選當前 frame 並繪圖。
2. State `CombatVisualMessage` 根據 effect、impact、phase 與目前 fighter 名稱解析
   正式 locale stable ID：
   - Cloudkill：`impact／commit／death` 的 resisted／killed；
   - Stinking Cloud：`impact／commit` 的 cough／helpless duration；
   - line spell：`impact` 傷害及 `commit` save succeeded／failed。
3. 無 impact、錯誤 phase、非 line spell 或 `Protected=true` 時保留既有 fallback；
   不能因搬移訊息而重播、提前或覆蓋其他戰鬥結果。
4. 目標顯示名稱由目前 Battle fighter ID 解析；找不到 fighter 時才保留 target ID。
5. 本輪不變更傷害、豁免、死亡、timeline duration、sprite 或聲音順序。

## 驗證

- 正式 locale 新增七個 format IDs，測試不複製中文全文。
- typed table test 覆蓋 Cloudkill saved／killed、Stinking Cloud saved／failed、
  line damage／save success／failure、protected fallback 與 travel fallback。
- 既有正常 Fireball、Lightning Bolt、Stinking Cloud、Cloudkill player path、
  active visual save／restore 與 renderer tests 由正式全套 gate 重跑。
- Go 漢字稽核 `320→313`；frontend `108→101`、runtime 212、localization 0
  不變。

本輪沒有幾何或圖像變更，不新增 README 截圖，也不擴大宣稱完整法術 fidelity。
