# 1231：`TREASURE` 跨 `NEWECL` 必須保留來源章節

狀態：`READY`

## 玩家可見缺陷

正常強度按鍵路線由 `ECL2/0x03` 進入世界段 `0x50` 時，會在戰後解析固定
物品區塊 `0x04`。舊實作以交接後的 `Area.GameArea=1` 尋找 `ITEM1.DAX`，因而
報錯：`TREASURE item block 0x04 for area 1 is not loaded`。

## 證據與判定

- 發出這筆 `TREASURE` 的原始位置是 `ECL2/0x03:1441h`，固定物品區塊為
  `0x04`，所以來源資料是 `ITEM2.DAX`。
- `ITEM1.DAX` 實際只有區塊 `0x05`；它沒有 `0x04`，不是遺失資產。
- 同一次 `BlockSession` 可在彙總結果前沿 `NEWECL` 跑入新段。若等到 State
  消費請求才看目前 area，來源章節已不可恢復。

後續長跑又證明更窄的第一版仍不夠：`TREASURE` 可能在一次按鍵先交回 UI，
下一次按鍵才執行 `NEWECL`。因此 State 在**每一次 VM 交界**收到請求時，就以
該結果的 `SessionEndBlockID` 附上 `SourceBlockID`；另用 `SourceBlockSet` 明確
區分「合法來源 block 0」與「舊呼叫端沒有提供來源」。解析時再依
`monsterChapterForBlock(SourceBlockID)` 選 ITEM 章節。

## 回歸閘門

- `TestBlockSessionKeepsTreasureSourceAcrossNewECL`：來源段發出寶物後切到另一段，
  彙總請求仍標來源段。
- `TestTreasureUsesTheSourceBlockChapterAfterNEWECL`：模擬 `TREASURE` 已在前一個
  互動交回、該份結果尚未帶 `NewECLBlockID`，仍由 session 邊界固定來源章節。
- 正常強度按鍵路線不再停在區塊 `0x04`，由 4 段推進到 12 段、610 格；目前
  下一個瓶頸是連續戰鬥下的正常強度存活，與資產載入無關。

## 不宣稱

這份規格不宣稱正常強度路線已通關，也不把測試用角色強化當成戰鬥平衡證據。
