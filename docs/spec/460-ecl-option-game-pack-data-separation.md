# 460：ECL 選項完整 game-pack 資料分離

狀態：`READY`

## 問題與證據

第 452 輪稽核仍指出 `localizeOption` 在 Go 內保存 84 個 CoAB ECL 選項 token
及其翻譯。雖然 State 已優先呼叫 engine `Pack.LocalizeOption`，當時 game-pack
只有其中 4 個來源 token；其餘 80 個仍落回作品專用 switch。這使新增城市、
遭遇或劇情選項時容易把中文再次寫入 runtime，也讓 JSON 改譯文後舊 fallback
成為第二份真相來源。

本輪以原 switch 的 84 個英文 token 作來源 oracle；英文 token 只用於 exact
ECL 比對，不當作玩家繁中內容。既有 25 條 `option_rules` 保留，另外 80 條
加入 CoAB game-pack，總數為 105。每條均保存：

- 唯一 stable rule ID；
- 原 ECL source token；
- stable message ID；
- `en／zh-TW` 完整資料。

## Runtime 契約

State 只透過 engine 的作品中立 `Pack.LocalizeOption` 解析已知 token。舊
`localizeOption` switch 已完全刪除；未知 token 原樣顯示，保留診斷證據，不會
猜譯、吞掉選項或用最接近的中文分支掩蓋資料缺口。路線 prompt 的目的地名稱
也改走同一份 State／pack resolver。

部分通用按鈕同時供 UI locale 使用。測試會逐一比較目前 65 個共享 stable ID
的值；兩份正式 JSON 若漂移會立刻失敗。產品層測試不再複製「撤退／審問／
殺死」等顯示文字，而是由 pack 解析預期內容。

## 玩家路徑與驗證

- game-pack 測試逐項驗證舊 84 個 ECL token 在 `en／zh-TW` 都能解析，並驗證
  105 條規則與 locale parity。
- 真實 `ECL1.DAX` 新遊戲與世界 Journey／General Store 流程通過。
- 真實火刀據點刀刃房 `ENTER THE BLADES／WAIT／RETREAT` 與定身房
  `RETREAT／INTERROGATE／KILL` 三分支通過；測試未由 debug UI 注入翻譯。
- focused gate marker：`ROUND460_FOCUSED_EXIT=0`；正常玩家路徑 marker：
  `ROUND460_PLAYER_PATH_EXIT=0`；exact audit marker：`ROUND460_AUDIT_EXIT=0`。
- 正式 Docker／Xvfb／`--network none` 全套 `cmd／gamepack／internal` gate
  通過，marker `ROUND460_FORMAL_EXIT=0`。

## 可量測結果

Go 漢字 literal exact baseline 由 867 降為 797 occurrences，共移除 70 次：

| 分類 | 第 459 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 372 | 302 | -70 |
| Ebiten 前端 UI | 135 | 135 | 0 |
| runtime／工具 UI | 360 | 360 | 0 |
| 合計 | 867 | 797 | -70 |

## 明確未完成

本輪只關閉已知 ECL menu token 的 Go 翻譯表，不代表全部劇情文字已資料分離。
`localizeECLText／localizeECLLine／unlockJournalEntries` 仍有 279 次繁中 fallback；
必須先證明每個原始 fragment 都有 `text_rules` 與手札頁覆蓋，才能刪除。未知
選項也仍需以真實 ECL corpus 找到後新增 JSON，不能在 State 恢復 switch。
