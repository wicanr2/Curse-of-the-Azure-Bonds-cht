# 1227：Crimdrac 放行與巫師戰後復仇

狀態：`READY`（2026-08-27）

## 問題

spec 1226 後的最大未走叢集是 `ECL1/0x50:0x13F9` 起 84 條。這支要求
`4C60=1 && 4C62=0`：擊敗 Dracandros，但 Crimdrac 尚未被擊敗。原正常戰役
在古熔岩洞遇到 Crimdrac 後固定選擇戰鬥，勝利就寫 `4C62=1`，所以世界地圖
備援復仇遭遇會合法地被關閉。

## 原版證據

- `ECL5/0x32:0x127C`：龍巫妖房先比較 `4C62=0 && 4C60=0`。
- `0x1328` 的 YES／NO 選擇有兩條合法玩家路徑：
  - 戰鬥勝利在 `0x139E` 寫 `4C62=1`。
  - 接受放行跳到 `0x13A5`，Crimdrac 把隊伍帶到斜坡，不寫 `4C62`。
- `ECL1/0x50:0x13E9`：只在 `4C62=0 && 4C60=1` 時寫 `4C62=1`，顯示
  `post-wizard.dracolich`，再載入 `MON1CHA` 的 `0x3C` Crimdrac 戰鬥。

## 實作

1. 正常戰役 observer 在 `lava-tube.crimdrac-introduces` 等待頁後明確選
   `YES`，接受 Crimdrac 放行；不寫記憶體、不注入 PC，也不繞過必經走廊。
2. 擊敗 Dracandros 與離開 Area 5 後，斷言 `4C60=1`、`4C62=0`。
3. 同一玩家 session 繼續從艾森布拉前往希爾斯法，observer 以原版等待頁與
   戰鬥交接處理復仇遭遇。
4. 抵達希爾斯法後同時斷言已看到 `post-wizard.dracolich` 且 `4C62=1`。

## 結果

- `ECL1/0x50` 現為 **433／798（54.3%）**。
- 全作實跑：9,285／14,177（65.5%）→ **9,329／14,177（65.8%）**。
- 圖外執行：0。
- `0x13F9` 起 84 條的叢集已消失。
- 2026-08-27 勘誤：初版只讀到未蓋叢集表格的中段，誤寫「新最大是
  `ECL2/0x04:0x072E`」。完整報表的真正最大項是
  `ECL5/0x33:0x11AF`，148 條；前者只是當時的第八大。
- 文字稽核維持 1,022 頁、unmatched 0；`0x13FF` 仍命中
  `post-wizard.dracolich`。

## 驗證與邊界

Docker、無網路環境：

- `TestRealNewGameRunsToTheEnding`：通過，包含同 session 的世界復仇遭遇。
- `go test ./internal/game -run 'TestReal|TestTilvertonRoute' -count=1`：通過。
- 連續按鍵 session：通過。
- `cmd/ecl-run-coverage`：9,329／14,177，圖外 0。
- `cmd/ecl-text-coverage`：unmatched 0。

本規格證明這條互斥分支由原版選項自然建立旗標並走到世界遭遇。它不把
「殺死 Crimdrac」的另一條合法路徑刪掉；後者仍有獨立原版戰鬥回歸測試。
