# 地圖圖集：事件在哪一格、從哪幾格離開

這個目錄的每一份 `.md` 由 `cmd/map-atlas` 產生，**不要手改**：

```bash
./tools/go.sh run ./cmd/map-atlas                        # 全部 18 張
./tools/go.sh run ./cmd/map-atlas -only tilverton.sewers.first-person
```

## 這個目錄存在的理由

先前每次要回答「這一格演什麼」「這張圖怎麼離開」都得再讀一次 ECL。那件事
**已經有答案了**，只是散在三個地方：`ON GOTO` 的分派表、GEO 的移動遮罩、
game pack 宣告的出口。圖集把三份接起來畫成一張圖，之後同一個問題不必再重跑
反組譯——要查路線、要對攻略、要判斷某一格接沒接，看圖就好。

## 怎麼讀

| 圖上的字 | 意思 |
|---|---|
| `.` | 沒有事件的地面 |
| `1`..`9`、`a`..`z` | 每格事件的索引（表在同一份檔案下方）|
| `#` | 四面都走不出去，走路站不上去 |
| `@` | game pack 宣告的進場錨點 |
| 列尾 `(x,y)^v<>` | 這一格往那個方向**走得出地圖**（圖緣交接）|

兩個要一起讀才不會誤判的欄位：

- **「走得出去」不等於「已經接上」。** `external_exit` 沒宣告的邊界格，remake 會
  照 `wrap` 繞回對邊，走不到腳本安排的目的地。每一份檔案的「離開這張圖的格子」
  表把兩者並排。
- **「有索引」不等於「站上去就會演」。** 處理常式自己可能還有守衛（一次性旗標、
  `RANDOM`、前置劇情、`SEARCH`），守衛就印在事件表裡。實際演出來什麼由
  [`cell-sweep`](../../audit/cell-sweep.md) 實跑量。

## 一條擋著的不變量

`State.CanMoveDungeon` 對**已宣告**的 `external_exit` **直接回 true，不看 GEO**
——所以把出口宣告在一格走不出去的邊界上，測試照樣綠，而正常玩家永遠走不到。
`cmd/map-atlas` 的 `TestDeclaredExitsSitOnCellsYouCanActuallyLeaveFrom` 逐個查
17 個宣告的出口，落在牆上就紅。成因與那次事故見
[spec 1199](../../spec/1199-tilverton-sewers-map-edge-handoffs.md)。

## 與公開攻略的對照（提爾佛頓下水道 ＋ 火刀據點）

攻略只用來**交叉核對地點語意**，座標一律以原始 GEO／ECL 為準。

### 火刀據點（`tilverton.fire-knife-hideout.first-person`）

| 攻略編號 | 攻略說法 | 圖上的字 | 格子 |
|---:|---|---|---|
| 1 | Entrance to Sewers | — | `(8,0)`／`(11,0)`／`(13,0)` 往北 |
| 2 | Twirling Blades | `p` | `(5,1)`、`(5,2)`、`(7,2)` |
| 3 | Frozen People（審問）| `q` | `(3,1)`、`(2,2)`、`(4,2)`、`(3,3)` |
| 4 | Torture Room | `o` | `(0,9)`..`(1,11)` |
| 5 | Information | `n` | `(5,10)`..`(7,11)`（倉庫）|
| 6 | Fire Knife Check Point | `4` | `(4,12)` |
| 7 | Hospital | `m` | `(7,13)`..`(9,15)` |
| 8 | Armoury | `8` | `(6,13)`、`(6,15)` |
| 9 | Guildmaster Fight | `7` | `(3,13)` |

攻略沒有編號、但圖上有的還有：辦公室 `r` `(14,11)`、臥室 `t` `(15,15)`、
圖書館 `u` `(14,13)`、實驗室 `v` `(12,14)`、停屍間 `w`、走廊煙味 `s` `(15,12)`，
以及十一格「找到一扇密門」（`9`、`a`、`b`、`d`、`e`、`f`、`h`、`i`、`j`、`k`）。

### 下水道（`tilverton.sewers.first-person`）

攻略把出口列成兩個：「10. To Fire Knives Hideout at Abandoned Checkpoint」與
「11. To Fire Knives Hideout」。原始資料是**三個**：南緣 `(10,15)`、`(13,15)`、
`(15,15)`，落點分別是據點 `(8,0)`、`(11,0)`、`(13,0)`——與據點北緣那三格
互為反函數。攻略的「Starting at 0,4」對得上：隊伍從公會進來時在西側那一塊
區域（X 0..4），而那一塊**走不到**南緣的出口，中間要經過兩次圖緣傳送
（spec 1199）。

來源：[Cheatbook 的地點清單](https://www.cheatbook.de/wfiles/curseoftheaurebonds.htm)、
[RPG Gamers 的流程](https://rpggamers.com/walkthrough/curse-of-the-azure-bonds)。
GameFAQs 與 GameBanshee 的同主題頁面對本機請求回 403，沒有取到。
