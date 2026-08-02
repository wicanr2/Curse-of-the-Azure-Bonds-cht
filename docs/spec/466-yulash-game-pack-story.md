# 第四百六十六輪：尤拉什戰區與指揮官事件資料化

狀態：`READY`

## 範圍

本輪把尤拉什十二個作品事件由 Go fallback 移入 CoAB game-pack：

- `yulash.red-plume-patrol`、`yulash.edge`、`yulash.entry`
- `yulash.riders-burst-out`、`yulash.checkpoint-halt`
- `yulash.see-commander`、`yulash.waiting-room`
- `yulash.zhentarim-spies`、`yulash.led-to-commander`
- `yulash.commander-business`、`yulash.commander-side-door`
- `yulash.pit-entrance`

每條規則保存原始 ECL source fragments 與 en／zh-TW 訊息。State 只提交原始
文字批次並接收匹配結果，不再知道尤拉什、紅羽衛、指揮官或摩安德巨坑的中文
內容。

## 正常玩家路徑

既有長回歸由 Hillsfar 正常前往 Yulash，依序驗證：

1. trail 紅羽衛巡邏文字、十二名敵人戰鬥與 Yulash edge。
2. 城牆入口選單、騎士與紫衣女子、檢查哨交涉、指揮官等候室。
3. `GEO3` block `10h` terrain `9Ah` 的散提爾間諜事件與十一名敵人戰鬥。
4. 戰後晉見指揮官、態度選單、遊戲內手札 22／52 與側門返回。
5. 已消耗的間諜事件不重播，沿 GEO 抵達 terrain `26h` 的摩安德巨坑入口。
6. 巨坑章節離場後再回 Yulash edge，仍由同一 stable ID 解析。

## 驗證與邊界

- 十二條規則在 en／zh-TW 均命中唯一 rule ID、非空訊息且不附帶額外手札頁。
- 玩家路徑測試的期望訊息由正式 game-pack stable ID 即時取得，不複製譯文。
- Go 漢字 literal exact baseline 由 651 降至 638；`localization_debt`
  156→143，frontend 135、runtime 360 不變。

本輪沒有改 renderer，也沒有產生新截圖。這不代表 Yulash 所有分支、敵人能力、
動態戰鬥演出、摩安德之坑章節或完整遊戲中文化已完成。
