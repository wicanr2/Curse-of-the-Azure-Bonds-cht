# 第四百六十三輪：哈普圖斯村與熔岩洞穴劇情資料化

狀態：`READY`

## 範圍

本輪把 Area 5 的 27 個作品事件由 Go `localizeECLText` fallback 移入 CoAB
game-pack：哈普圖斯村 17 個、熔岩洞穴 10 個。每條規則保存原始 ECL
`all_contains` source identity，en／zh-TW locale 保存完整畫面訊息；State 與
engine 不知道村名、NPC、敵人或中文內容。

## 正常玩家路徑

既有長回歸從 Standing Stone 出發，經 Essembra 到 Hap，依序驗證：

1. 三隻黑龍遭遇、Hap edge、進入 ECL5 block `31h`。
2. 廢村、躲藏村民、黑暗精靈巡邏、阿卡巴加入與解放前旅店。
3. 伊弗利特與十二名黑暗精靈、洞穴地圖、村民歡呼、長老感謝、法師塔線索、
   阿卡巴祕密商路。
4. 離村並依地圖進 ECL5 block `32h` 熔岩洞穴。
5. 洞口伏擊、守門巡邏、夢境警告、火蜥蜴熔岩池、友善交涉、十五隻火蜥蜴、
   六只防火桶與失敗耐熱檢查。

上述路徑同時驗證 continuation、戰鬥旗標、Akabar roster、GEO 座標與 stable
message ID。27 條中的 `lava-tube.sly-parlay` 目前只有原始 source rule 與
en／zh-TW pack coverage；正常路徑走的是 `nice` 與 combat 分支，因此不可宣稱
狡猾交涉已完成玩家路徑驗收。

## 驗證與完成邊界

- 27 條規則在 en／zh-TW 均命中唯一 rule ID、非空訊息且不附帶手札頁。
- `gamepack` 與完整 `internal/game` 回歸通過；產品層改從正式 stable ID 取得
  期望訊息，不複製中文。
- Go 漢字 literal exact baseline 由 688 降至 661；`localization_debt`
  193→166，frontend 135、runtime 360 不變。

這是 Area 5 的文字／事件綁定切片，不代表熔岩洞穴全部分支、法師塔、怪物
特殊能力、動畫、音效或整個遊戲已完成。
