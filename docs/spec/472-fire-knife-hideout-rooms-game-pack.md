# 第四百七十二輪：火刀據點房間事件資料化

狀態：`READY`

## 範圍

本輪將 `ECL2.DAX` block 4 的十一個作品文字 boundary 移入 CoAB game-pack：

- 刀刃屏障提示、傷害與消失；
- 凝固房提示與殺戮分支；
- 火刀高層辦公室；
- 煙味走廊、整齊臥室、燒毀圖書館、燒毀實驗室與裹屍房。

每條規則保存 raw ECL token、英文訊息與繁中訊息。`State.localizeECLText` 不再
擁有這些作品比對與譯文；選項仍由既有 game-pack `option_rules` 驅動。

## 原始行為與回歸

- 刀刃屏障保留 ENTER／WAIT／RETREAT，ENTER 產生全隊 `8d8` 傷害，之後與 WAIT
  都進入屏障消失文字並留下完成旗標；重訪不重播。
- 凝固房保留 RETREAT／INTERROGATE／KILL；審問解鎖完整遊戲內手札 26，完成後
  重訪靜默。殺戮文字另有獨立 stable ID，不與房間提示合併。
- 辦公室初訪、SEARCH、手札 9、500 gold／500 platinum／3 gems／2 jewelry 與
  ITEM2 block `82h` 財寶行為不變；已取走後不得複製。
- terrain `9Ch..A0h` 五個房間分別寫 `4C11h..4C15h=1`，各自完成後重訪靜默。

raw oracle 仍直接驗證 ECL 英文 token、選項、傷害 request、財寶 request 與旗標；
產品層顯示則由 stable ID 取得正式 game-pack 訊息，兩種證據不互相取代。

## 驗證

- `TestFireKnifeHideoutRoomsAreGamePackDriven`：十一條 en／zh-TW 規則。
- `TestRealFireKnifeBladeBarrierBranches`。
- `TestRealFireKnifeFrozenRoomBranches`。
- `TestRealFireKnifeOfficeStages`。
- `TestRealFireKnifeAshenRooms`。
- Go 漢字基線：`580 → 569`；`localization_debt 85 → 74`，frontend 135、
  runtime 360 不變。

這些是 block 4 真實分支與互動回歸；測試以 terrain selector direct-entry 隔離每個
房間，不能單獨支撐「從下水道逐步走遍據點」或完整火刀主線已完成。renderer 未
變更，因此本輪不新增 README 截圖。
