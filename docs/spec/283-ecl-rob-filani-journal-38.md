# 第二百八十三輪：ECL ROB、賢者菲拉妮與手札條目 38

狀態：READY

## Reference evidence

CoAB reference `ovr003.CMD_Rob` 固定讀三個 operands：

1. `allParty`：零值只處理目前選角，非零處理整隊；
2. `lossPercent`：每種 coin 保留 `(100-lossPercent)%`，逐欄向下取整；
3. `itemChance`：依 inventory 順序擲 `1d100`。重量 `>24` 先將 chance 減 50，
   重量 `>255` 先減 90，最低為零。

ECL2 block `0x01` SearchLocation 在 GEO2 block 1 `(6,5)` 的 terrain selector `0x8A`
進入賢者菲拉妮事件。PICTURE `5` 使用當下 `HeadBlockId=5`，先詢問玩家是否為印記而來；
回答 YES 後提供 `TELL THE TRUTH／LIE／EXIT`。如實相告會在 payload `+0x0F4D`
執行 `ROB 1,50,0`，即全隊失去一半各類 coin、不偷物品，接著顯示 `SHE TALKS` 與
`38.`，要求玩家閱讀 Adventurer's Journal Entry 38。

使用者提供的 Adventure Journal TXT/PDF 說明 Entry 38：五枚印記分別涉及火刀、
摩安德、散塔林會、未知燃燒印記，以及與暗影谷強大賢者相似的弦月標誌。

## Remake transaction

- VM 發出 renderer-neutral `RobRequest`，BlockSession 跨 pause 聚合且不重複執行。
- State 依 request 處理選角／全隊，縮減 Copper、Silver、Electrum、Gold、Platinum，
  並以 deterministic RNG 套用 reference item weight/chance 規則；Gems、Jewelry 不縮減。
- DOS player money record `0xFB..0x104` 的五種 coin 都進入 typed Character 與
  raw-preserving writeback，不再只保存 Gold。
- 菲拉妮的兩層選單與對話完整繁中化；真話分支恰好扣款一次。
- Entry 38 只在事件觸發後解鎖，依 640×480、24px、每頁 22×7 字容量拆成三頁。
- `-filani` 可從正式建立／序幕 transaction 重現 PICTURE 5 畫面。

## Regression

- synthetic `ROB 1,50,0 → PRINT → EXIT` 仍繼續執行；
- State 驗證五種 coin 向下取整及重物 chance penalty；
- DOS player 五種 coin parse／Character projection／writeback；
- real image：正式新遊戲 session → GEO selector `0x8A` → PICTURE 5 →
  YES → TELL THE TRUTH → ROB 50% → Entry 38 三頁解鎖 → 兩次 Continue →
  返回 `(6,5)`。
