# 第 480 輪：戰鬥 HUD locale 契約

狀態：`READY`

## 問題

Ebiten `drawCombat` 曾直接包含 25 筆繁中文字面值，包括 HP／AC 標籤、施法
確認、五種格點法術游標提示、一般目標提示、移動剩餘格數、十二個法術快捷鍵、
目前目標與快捷列標題。這使 renderer 同時負責版面與翻譯，也讓未來修改文字時
必須更動繪圖程式。

## READY 契約

1. Ebiten renderer 只決定 DOS 石框上的座標、字型、顏色與顯示時機，不保存
   玩家可見中文。
2. State 提供 locale-backed HUD contract：
   - `CombatHitPointsLabel`／`CombatArmorClassLabel`；
   - `CombatSelectionPrompt`，依 typed casting／move state 回傳完整 prompt；
   - `CombatQuickSpellHints`，依真正可用法術按固定原版按鍵順序回傳提示；
   - `CombatTargetStatus`／`CombatQuickStatus` 格式化動態內容。
3. 所有 stable IDs 必須存在正式 `assets/locale/zh-TW.json`；產品測試從同一 catalog
   取得期望值，不複製目前中文。
4. 法術是否合法、spell slot、target transaction、移動剩餘格與快捷順序仍由
   現有 typed State 決定；本輪不得為了顯示資料化改變戰鬥規則。
5. 本輪不調整 combat frame 幾何、sprite anchor 或動畫時間軸，因此不宣稱新的
   原版畫面 fidelity，也不替換 README 截圖。

## 驗證

- 正式 locale coverage 驗證 24 個 HUD stable IDs。
- HUD contract 測試覆蓋 Bless 確認、一般 fighter target、Fireball、Sleep、
  Lightning Bolt、Stinking Cloud、Cloudkill、移動剩餘格、target 與 quick
  status 格式化。
- 既有正常 ECL 戰鬥、施法、移動、save／restore 與動畫測試仍由正式全套 gate
  驗證，證明只改顯示資料邊界，未改 typed transaction。
- Go 漢字稽核 `345→320`；frontend `133→108`、runtime 212 不變，
  localization 仍為 0。

## 可沿用經驗

戰鬥 UI 的本地化邊界應位於 game adapter／State，而不是 renderer 或共用戰鬥
數學核心。其他 SSI RPG 可以沿用「typed action state → locale message ID →
renderer placement」方法，但按鍵配置與可用法術仍須由各作品證據驗證。
