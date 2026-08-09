# 第五百三十輪：DOSBox 測試模式角色建立與年齡畫面證據

狀態：`READY`（限原版角色建立畫面／資料可見性；不是正常玩家路徑）

## 目的

本輪只為取得原版 DOS 角色建立畫面的可重現證據，交叉核對年齡、能力值、
等級、HP 與文字密度。原版 copy-protection 需要翻譯輪；為避免把手動解碼
時間混入 remake 驗收，使用遊戲本身的 `START STING Wooden` 測試參數進入
角色建立。這個模式只可作 `cheat-assisted` 取證，不能證明正常玩家路徑、
亂數、存檔或完整冒險。

## 輸入與重現條件

| 項目 | 值 |
|---|---|
| 平台／版本 | DOS 原版，畫面顯示 `V1.0` |
| `START.EXE` SHA-256 | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` |
| `GAME.OVR` SHA-256 | `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` |
| 執行環境 | Docker `dosbox-run:latest`、Xvfb、DOSBox、`DISPLAY=:82` |
| 啟動命令 | `START STING Wooden` |
| 輸入邊界 | 測試模式提示後 Space → 主選單 `C` → 預設 Dwarf／male／Fighter／Lawful Good → 接受能力值 → 姓名欄 |
| 原始 screenshot | [`530-dos-character-creation-age.png`](../../assets/reference/original-dos/runtime/530-dos-character-creation-age.png) |

原始檔只讀掛載；遊戲可寫內容位於 Docker 暫存副本，沒有修改原始
`START.EXE`、`GAME.OVR` 或版本化輸入。

## 可直接看見的結果

截圖在姓名輸入欄出現前仍保留角色資訊頁，可見：

- `MALE DWARF AGE 54`；
- `LAWFUL GOOD`、`FIGHTER`；
- `STR 18(39)`、`INT 18`、`WIS 14`、`DEX 16`、`CON 17`、`CHA 16`；
- `LEVEL 5`、`EXP 25000`、`HP 46`、`THAC0 15`、`MOVEMENT 12`；
- 原版石框、左上資訊列與底部 `REROLL STATS? YES NO` 的版面。

這些是 screenshot 內的 `exact` 可見內容；「年齡欄位如何由亂數／角色
record 產生」不由本輪推導。現有 `.SAV/.GUY` age offset 研究仍以
[`docs/spec/243-player-age-writeback.md`](243-player-age-writeback.md) 與
[`docs/spec/251-save-format-age-evidence.md`](251-save-format-age-evidence.md)
為準。

## 證據等級與限制

- 畫面文字、位置與原版石框：`exact`（同一 DOS runtime screenshot）。
- 角色建立流程可由測試模式抵達姓名欄：`exact`（cheat-assisted runtime）。
- `STING Wooden` 下的數值是否與一般新角色完全相同：`unknown`；本輪不作
  正常亂數或 AD&D 角色生成結論。
- 從此畫面進入完整冒險、戰鬥、地圖或存檔：`未證實`。

公開攻略[記錄 `START.EXE STING Wooden`](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)
為可略過 copy-protection 的第二級測試模式；本專案只把它當作取證輔助，不把
測試模式 screenshot 當成正常玩家路徑驗收。下一步若要取得正常 runtime，仍需
同版本翻譯輪答案或另一個不改變遊戲行為的合法啟動方式。
