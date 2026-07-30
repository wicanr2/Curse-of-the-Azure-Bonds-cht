# 402：外圍遺跡無名者與灌木伏擊

狀態：`READY`

## 問題

block `42h` 的 terrain `07h／08h／09h` 表面上只是提示、救人與血跡，
實際卻包含兩個一次性旗標、HEAD／BODY 人物圖、選擇性犧牲、全隊落石傷害、
三組十一名敵人，以及一次同 block 地圖傳送。若只翻譯畫面文字，玩家會留在
錯誤座標，戰鬥陣容與 terrain `09h` 的敘事關係也會失真。

## 原始資料

來源為 `curseoftheazurebonds.zip`：

| member | SHA-256 | bytes |
|---|---|---:|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` | 22,175 |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` | 3,697 |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` | 1,424 |

ECL6 block `42h` 的 SearchLocation 先做 `C04F & 7Fh`。jump table 中
terrain `07h／08h／09h` 分別前往 payload
`+0F68h／+0FFDh／+11F3h`。

## terrain 07h：無名者警告

- GEO 座標是 `(6,12)` 與 `(13,7)`。
- payload `+0F68h` 比較 `4C06h==1`；未成立時立即寫 `4C06h=1`。
- 它把 `7EE1h` 設為 `43h`、執行 `PICTURE 46h`，再把 `7EE1h` 還原
  `FFh`。raw runtime 因此觀察到 BODY `46h`、HEAD `43h`。
- 無名者警告直接接近很危險，並指出神殿在北方。按鍵後離開；同一 session
  重訪不再顯示。
- `4C06h` 在本 block 只有 `+0F70h` 這一個 writer。它不是鄰近事件使用的
  `4CD6h`；不得為了地址看起來不尋常而自行改字。跨 DOS save 的持久性尚未
  另做實機邊界測試，因此只宣稱同一 runtime session 的 exact 行為。

從上一輪藏寶 `(14,3)` 到 `(13,7)` 的七步合法 GEO 路線是：

```text
(14,3) → (15,3) → (15,4) → (15,5) → (15,6)
       → (15,7) → (14,7) → (13,7)
```

## terrain 08h：小孩／半身人誘餌

- GEO 座標是 `(12,9)` 與 `(9,10)`。
- payload `+0FFDh` 先比較 `4CD6h==1`；首次進入在顯示選項前就寫
  `4CD6h=1`，所以救援或放棄都只觸發一次。
- 玩家看見一名小孩或半身人鑽入灌木，地獄犬正要追入：
  - `NO`：地獄犬稍後叼著血肉離開，沒有戰鬥。
  - `YES`：玩家先殺死追入的地獄犬；ECL 接著寫
    `C04B=0Bh／C04C=0Ah／C04D=02h`，再 `CALL 2E10h`，也就是把隊伍
    放到 `(11,10,S)` 並重繪同一張地圖。
- 羅剎妖從灌木起身，屋頂石像鬼投擲石塊。exact `DAMAGE` 是：

  ```text
  flags=0Ch, dice=2d8, bonus=0, saveFlags=34h
  ```

- 落石後建立：

  ```text
  MON6CHA 44h HELL HOUND ×5
  MON6CHA 45h MARGOYLE   ×5
  MON6CHA 43h RAKSHASA   ×1
  ```

- 正常 State 路徑會解析全隊傷害、完成十一名敵人的可操作戰鬥，經
  `ModeEvent → Continue` 恢復地城；重踏 terrain `08h` 不重播。

同 block 傳送揭露既有 adapter 缺口：`CALL 2E10h` 不只要求 renderer
重繪，也會使用 ECL 剛寫入的地城 register。bounded VM 現保留每筆 numeric
`SAVE`／`SAVE TABLE` 的 block、PC、address 與 value，以及每筆 CALL 的
block／PC。第 403 輪以部分座標傳送及 Filani scratch-coordinate 反例進一步
收斂條件：只有整次 session 未跨 block，且 CALL 前新寫 `C04D` 作為
facing commit，`State.applyECLCallSignals` 才投影同批實際新寫的
`DungeonX／DungeonY／DungeonDirection` 欄位，再保留 one-shot redraw
request。
這是作品中立的 CALL transaction，不含 terrain 或 CoAB 座標 hardcode；
跨 `NEWECL` 的公會與眼魔洞穴玩家路徑另有回歸，防止舊 work registers
覆蓋目的地出生點。

## terrain 09h：血跡灌木

- 唯一座標是 `(11,11)`；從 ECL 傳送目的地 `(11,10,S)` 正常向南一步
  即可抵達。
- payload `+11F3h` 每次都顯示灌木、瓦礫與葉片血跡，隨即 EXIT。
- 沒有旗標、選單、戰鬥、搜索或財寶。它是可重複環境敘事；不能因
  terrain `08h` 已完成便將它清空，也不能擅自加入戰利品。

## 實作與驗證

1. 七段英文／繁中敘事以 stable ID 放入 CoAB game-pack；Go 程式與測試
   不複製中文正文。
2. `TestRealOuterRuinsNamelessAndBrushAmbush` 覆蓋：
   - terrain `07h` 的 `4C06h`、HEAD／BODY 與重訪；
   - terrain `08h` 的 `YES／NO`、`4CD6h`、`2d8` 傷害、三組敵人與勝利
     continuation；
   - terrain `09h` 的可重複純敘事。
3. Standing Stone 起始長玩家路徑從 terrain `06h` 藏寶繼續走合法 GEO，
   查看無名者、救援誘餌、套用 ECL 傳送、承受落石、打完十一名敵人，再從
   `(11,10,S)` 向南走到血跡灌木。
4. `TestApplyECLCallSignalsRedrawProjectsSameBlockDungeonRegisters` 鎖定
   `2E10h` 的同 block register 投影，不依賴本作座標。

## 完成邊界

本輪完成 terrain `07h／08h／09h`，不代表外圍遺跡或戰鬥系統完成。
terrain `0Bh／0Ch／8Ah／8Dh`、下水道、神殿與結局仍待逐項接通。
HELL HOUND／MARGOYLE／RAKSHASA 的完整 AD&D 特殊能力、落石的 DOS
畫面／音效時間碼，以及 `4C06h` 跨原版 save 的持久性仍未驗證。
