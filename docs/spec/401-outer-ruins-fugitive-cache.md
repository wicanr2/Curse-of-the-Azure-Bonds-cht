# 401：外圍遺跡逃亡男子與東北藏寶

狀態：`READY`

## 問題

block `42h` 的 terrain `04h／05h／06h` 是一條跨多個座標、戰鬥、人物圖與
主動搜索的事件鏈。只實作「看見逃亡男子」會遺失拒絕救援、追擊地獄犬、
屍骸一次性狀態、臨終線索與真正報酬。本輪從 real-image ECL 完整關閉這條
鏈，並修正一枚 electrum 因整數 GP 投影而消失的通用財寶缺陷。

## 原始資料證據

來源是 `curseoftheazurebonds.zip` 中
`GEO6.DAX／ECL6.DAX／MON6CHA.DAX／ITEM6.DAX`。

### GEO 與正常路線

- terrain `04h` 位於 `(3,5)／(2,6)／(3,7)`；同一 `4CD3h` 旗標使事件只
  觸發一次。
- 男子倒下後的 terrain `05h` 位於 `(3,6)`。
- 東北藏寶 terrain `06h` 位於 `(14,3)`。
- 倉庫 `(2,14)` 至首個 terrain `04h` `(3,7)` 的八步路線，以及
  `(3,7)` 經 `(3,6)→(2,6)→…→(14,3)` 的九步 wrap 路線，均逐步通過
  `GEO CanMoveDungeonWrapped`。正常玩家回歸途中會依正式 encounter
  選單處理隨機遭遇，不以瞬移略過。

### terrain 04h：救援或放棄

- 初見先寫 `4CD3h=1`，再詢問是否救援筋疲力竭的男子。
- `YES` 立即建立 `MON6CHA 44h` HELL HOUND ×6。
- 戰勝後顯示 `HEAD6 40h + BODY6 40h`；男子說自己被抓前把報酬藏在東北
  廢墟，隨即死亡。流程完成後 `4CD4h=1／4CD5h=1`：
  - `4CD4h` 使相鄰屍骸 terrain 不再重播。
  - `4CD5h` 是東北藏寶尚可取得的線索。
- `NO` 會讓地獄犬把男子撕碎，再問是否攻擊：
  - `YES` 仍迎戰同一批六隻地獄犬，之後只找到無法辨識的屍骸。
  - `NO` 顯示地獄犬離開殘骸，不給臨終線索；踏入 terrain `05h` 時仍會
    看見一次屍骸。

### terrain 05h：屍骸

- `4CD4h==0` 時顯示「幾乎看不出人形，沒有值錢物品」並寫 `4CD4h=1`。
- `4CD4h==1` 時直接離開。普通踏入與主動搜索沒有額外財寶分支。
- `4CD4h` 的 writers 位於 payload `+0E31h／+0EC5h`；前者與救援結局
  一起寫入，後者屬屍骸分支。

### terrain 06h：東北藏寶

- 普通踏入永遠安靜。只有 `4CD5h==1` 且玩家主動 `SEARCH`
  (`7ECAh=1`) 時，才顯示找到秘密藏寶。
- 按鍵繼續後 exact request 是：

  ```text
  TREASURE [0,0,1,0,0,0,0], ItemBlock=43h
  ```

- 第三種 coin 是一枚 electrum，等值 100 copper、半枚 GP；不能因
  `MoneyPool()` 只顯示整數 GP 就把它丟掉。
- `ITEM6.DAX/43h` exact 解析為三件物品：
  `Gauntlets +2`、`Girdle +1`、`Long Sword +5`。
- 取得前 payload `+0F20h` 將 `4CD5h` 清為零，因此再搜索不會複製財寶。

`BlockSession` 覆蓋救援、拒絕後追擊、拒絕後離開、屍骸重訪、沒有線索、
有線索但普通踏入、主動搜索與重搜。writer addresses 另由
`-find-save-destination 4CD3／4CD4／4CD5` 與 raw bytes 交叉驗證；以上
標為 `exact`。

## 實作

1. 六段英文／繁中敘事移入 game-pack stable IDs；Go 程式與測試不複製
   中文正文。
2. `State` 的整數 GP pool 新增 `moneyCopperRemainder`。所有
   `TREASURE` coin 先以 copper 精確加總，再把每 200 copper 投影為一 GP，
   餘數保留；兩枚 electrum 會正確累積成一 GP。
3. 帶 `PRESS BUTTON` 暫停的 ECL 藏寶敘事會跨 selection 保存到財寶選單。
   空白 packed text 不再覆蓋前一頁的繁中訊息。
4. Standing Stone 起始的正常玩家路徑在第 400 輪倉庫後，逐格走到
   terrain `04h`，救援並完成六犬戰鬥、驗證 `HEAD／BODY 40h` 與三旗標，
   再走到 `(14,3)` 執行真正 `SEARCH`。
5. 玩家路徑驗證三件真實 ITEM6 裝備、100-copper 餘數、`4CD5h=0`、
   財寶選單返回地城與第二次搜索無結果。

## 完成邊界

本輪完成逃亡男子與其藏寶事件鏈，不代表外圍遺跡完成。terrain
`07h／08h／09h／0Bh／0Ch／8Ah／8Dh`、下水道、神殿與結局仍待逐項實作。
HELL HOUND 的完整特殊能力與原版 DOS 戰鬥動態演出也仍是戰鬥子系統缺口。
