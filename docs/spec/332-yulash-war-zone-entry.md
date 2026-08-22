# 第三百三十二輪：尤拉什戰區與指揮官等候室

狀態：`READY`

## 原版資料證據

希爾斯法城外的 `JOURNEY ON` 提供
`YULASH / THE STANDING STONE / PHLAN`。選擇 `YULASH → TRAIL` 會固定觸發
world event 15：

1. 一隊紅羽衛巡邏兵接近；
2. 因隊伍仍帶有 Fzoul 的散塔林枷印，巡邏兵指控眾人是 Zhentarim spy；
3. ECL 建立十二名 fighter `0x59`、icon `0x20`；
4. 勝利後寫入 `4C9B=10`，抵達尤拉什城外。

尤拉什不是標準城市。`ENTER CITY` 由 world block `0x51` 切到 ECL3 block
`0x10`，顯示廢墟與戰鬥聲，提供 `SNEAK IN / ASK PERMISSION / LEAVE`。本輪依
隨附攻略建議走請求許可：

- 首次入城，一名騎高大駿馬的男子衝出，撞倒隊伍成員；紫衣女子在他背後喊
  「抱歉」；
- 紅羽衛檢查哨提供 `RUN AWAY / FIGHT / PARLAY`；
- PARLAY 後衛兵要求隊伍去見指揮官，提供
  `GO WITH GUARDS / FIGHT / RUN AWAY`；
- 跟衛兵走後由腳本演出護送過程，收在指揮官等候室 GEO3 block `0x10` 的
  `(0,3)`、朝東。

★ **等候室不是 teleport，是腳本自己走過去的。** `ECL3/0x10` 的護送段
（`192Ah`..`1997h`）是一個查表驅動的走位迴圈：

```
192A  GETTABLE 9DD7 4C00 7F7A   ; 7F7A ← 這一段要走幾步
1934  GETTABLE 9DDC 4C00 7F79   ; 7F79 ← 方向表的起始索引
193E  GETTABLE 9DA4 7F79 C04D   ; C04D ← 方向表[7F79]
1948  CALL     C01E             ; MOVEFORWARD
196F  CALL     2E10             ; 重畫
1974  ADD      1 7F79 7F79
197D  SUBTRACT 1 7F7A 7F7A
198C  IF <> → GOTO 193E
1991  SAVE     01 C04D          ; 收尾朝東
1997  CALL     2E10
19AE  PRINTCLEAR "THIS IS THE COMMANDER'S WAITING ROOM."
```

實際跑出來的方向序列是 `C04D` ＝ 2、2、2、3 ＝ 南、南、南、西：從
`0127h` 那次 `(1,0)` 朝南的落點走四步，正好停在 `(0,3)`，再由 `1991h`
轉成朝東。**這一段完全不需要外層 dispatcher 補任何 map register。**

## 引擎與中文契約

- 北方 world destination 名稱必須由原始英文 token 映射繁中，selection index
  不變。
- 十二人巡邏戰使用原版 MON／CPIC fighter，不以文字或縮減敵數替代。
- 尤拉什入口、首次騎士事件、檢查哨與護送選單完整繁中；原始 bond gate、
  damage signal 與一次性 flag 不修改。
- ECL lifecycle 直接 EXIT 到 `ModeDungeon` 時，也必須同步目前 block 所屬
  GameArea。block `0x10` 對應 ECL3／GEO3，故設定 `GameArea=3`、
  `InDungeon=true`、`GeoMapSet=3`、`GeoMapBlock=0x10`。
- `GO WITH GUARDS` **不預寫任何 map register**：走位與收尾朝向都由
  `192Ah`..`1997h` 那段腳本負責，remake 只要照序投影 `2Dh` 的
  `CALL C01Eh`／`CALL 2E10h` 就會落在 `(0,3)` 朝東。

## 可沿用的 Gold Box 知識

⚠ **「ECL 沒寫座標」這句話要讀完整個 block 才能說。** 走位可以完全不出現
`SAVE ?? C04B` 這種形狀——CoAB 用的是 `GETTABLE` 把方向表讀進 `C04D`，
再靠 `CALL C01Eh` 一步一步移動，座標從頭到尾沒有當過 operand。只掃
「有沒有人寫 `C04B`／`C04C`」會得到**假零**，然後很自然地推論成「這一定是
外層 dispatcher 做的」，接著在 remake 這一側補一個硬寫死的 teleport。

移植其他 Gold Box 作品時，判斷「這張圖的落點是誰決定的」要看三種形狀：
直接寫 `C04B`／`C04C`、`CALL` 移動常式、以及查表驅動的移動迴圈。三種都排除
之後，才輪得到「外層 dispatcher」這個解釋。

⚠ **這個補丁在測試上看不出來，這是它活得久的原因。** remake 在每一次
`CALL C01Eh` 走完之後都會 `syncDungeonECLRegisters()` 把暫存器重新錨回隊伍
位置，所以任何「執行前先硬寫終點」的動作會被第一步抹掉——終點、訊息、
選單全部一樣，端到端測試全綠。它唯一真正污染的是 `ViewMirror`
（spec 1172），而鏡射的分岔也在同一次 sync 被蓋回去。

結論不是「再補一條測試」：**目前的模型下這兩者確實沒有可觀察差異**，
拿掉它的理由是它把一個錯的機制寫進程式碼與規格，而不是它會壞掉。

另外，Dungeon mode 有兩種抵達方式：事件畫面 Continue，以及 ECL 無文字 EXIT
直接回 engine loop。Area／GEO 同步必須由共用 helper 覆蓋兩條路徑。

## 驗收

- 同一 real-session 從火刀、法師塔、艾森布拉、希爾斯法延伸到尤拉什。
- 驗證目的地繁中、十二名紅羽衛、`4C9B=10` 與 `LocationYulash`。
- 驗證 ECL `0x51→0x10`、三層入口選單、首次騎士事件與等候室繁中。
- 最終狀態必須是 `ModeDungeon`、Area 3、GEO3 block `0x10`、
  `(DungeonX,DungeonY,Direction)=(0,3,2)`，並發布新地圖載入。
- 護送段要**逐步**走：`internal/game` 的正常玩家路徑測試會檢查中途每一次
  `2E10h` 的落點，一次跳到 `(0,3)` 的實作在那裡是看不出來的（見上面
  「等候室不是 teleport」）。
