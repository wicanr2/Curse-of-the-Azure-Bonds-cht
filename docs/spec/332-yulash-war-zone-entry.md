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
- 跟衛兵走後抵達指揮官等候室，Continue 進入 GEO3 block `0x10` 的
  `(0,3,E)`。

攻略地圖明確標示 waiting room `W=(0,3)`，東側有出口；ECL dialogue 本身不寫
map registers，座標與朝向是 DOS area loop 的 teleport transaction。

## 引擎與中文契約

- 北方 world destination 名稱必須由原始英文 token 映射繁中，selection index
  不變。
- 十二人巡邏戰使用原版 MON／CPIC fighter，不以文字或縮減敵數替代。
- 尤拉什入口、首次騎士事件、檢查哨與護送選單完整繁中；原始 bond gate、
  damage signal 與一次性 flag 不修改。
- ECL lifecycle 直接 EXIT 到 `ModeDungeon` 時，也必須同步目前 block 所屬
  GameArea。block `0x10` 對應 ECL3／GEO3，故設定 `GameArea=3`、
  `InDungeon=true`、`GeoMapSet=3`、`GeoMapBlock=0x10`。
- `GO WITH GUARDS` 在 resume 前寫入原 area-loop 的
  `C04B=0, C04C=3, C04D=1`；State 投影為 `(0,3)`、八向 facing `2`（東）。

## 可沿用的 Gold Box 知識

ECL 可負責對話與 branching，但真正的 map teleport 可能由外層 area dispatcher
完成。移植其他 Gold Box 作品時，不能因 ECL 沒出現座標 operand 就沿用上一張圖
的 registers；應用地圖、攻略或反組 dispatcher 證據補齊 engine transaction。

另外，Dungeon mode 有兩種抵達方式：事件畫面 Continue，以及 ECL 無文字 EXIT
直接回 engine loop。Area／GEO 同步必須由共用 helper 覆蓋兩條路徑。

## 驗收

- 同一 real-session 從火刀、法師塔、艾森布拉、希爾斯法延伸到尤拉什。
- 驗證目的地繁中、十二名紅羽衛、`4C9B=10` 與 `LocationYulash`。
- 驗證 ECL `0x51→0x10`、三層入口選單、首次騎士事件與等候室繁中。
- 最終狀態必須是 `ModeDungeon`、Area 3、GEO3 block `0x10`、
  `(DungeonX,DungeonY,Direction)=(0,3,2)`，並發布新地圖載入。
