# 第一千二百四十輪：現行 route oracle 與短暫敗戰訊息捕捉

狀態：`READY`（關閉 route 入口漂移與測試重放漏記敗戰；不宣稱一般強度通關）

## 問題

文件長期引用 `route-clean-716.json`，工作清單記錄的正式產物有 24,370 步，
目前檔案卻只剩 1,684 步。原因是 `COAB_DECISION_LOG` 直接寫正式路徑，窄測試也能
成功覆寫它；重放載入後不會報錯，只會安靜地走得比較短。

另一個取樣洞發生在 POSTCOM：`playCombatTurn` 的最後一個按鍵可在同一次
`Update` 內顯示「戰鬥失敗。」並完成後續 ECL 交接。下一幀 `observe` 有時已看到
別的頁面，恢復旗標便漏設；隊伍會以 0 HP 繼續走到下一個 `DAMAGE` 才真正全滅。

## 修正

- `tools/rebuild-key-route.sh` 以完整 `internal/game` 測試寫入候選檔；測試成功且
  決策至少 10,000 步後，才原子替換唯一現行入口
  `workplace/campaign-frames/route-current.json`，並輸出 SHA-256 sidecar。
- 腳本固定 Docker 無網路、資源上限、目前 UID/GID 與本機 engine replace；重生後
  固定 `COAB_KEY_BOOST=0` 跑 1,800 幀一般強度玩家路徑。
- `loadRoute` 對 `route-current.json` 加同一個 10,000 步下界；這不是完整度證明，
  只防止窄測試產物再次冒充正式 oracle。
- 戰鬥重放在正式 `CombatMessage` 仍可見的動作提交點同步記錄敗戰；跨幀
  `state.Message` 取樣仍保留作第二條路。

## 驗證與新停止點

- 目前程式完整重生 **39,973 步**，SHA-256：
  `acc144b5dc4b346d0297418587b36b83825373b33a0ad2353eb5765f95e30114`。
- 一般強度 1,800 幀：215 格／178 句／8 個 ECL 段、原文 fallback 0、全滅 0；
  第 1,745 幀進 `0x33`，最後新進展第 1,798 幀。這證明短暫敗戰不再漏掉，
  不等於打贏全部戰鬥。
- 延長控制組在第 2,255 幀真正全滅：塔內 12 名黑暗精靈／梟熊戰敗，正式 REST
  雖恢復 HP，但原版契約不會呼叫 `STANDUP`。後續樓梯 `DAMAGE` 依 spec 1204
  將全員不可行動判為全滅。
- 不把 REST 改成復活：spec 128 只證實自然 HP 恢復；原版 `STANDUP` 的直接
  呼叫端只有 overlay-12 兩支效果與 overlay-22 一支法術流程（spec 579／745／
  far-call map），沒有 REST。

## 下一步

一般強度隊伍已是原版五／六級，但正常建角與整備仍只提供一級法術。下一個切片是
核對原版完成建角的法師／牧師已知法術、環級槽與初始法術選擇，讓塔內大型戰鬥以
角色實際應有的能力通過；不得再以自然休息冒充復活或用 boost 支撐。
