# 第 535 輪：攻略與地圖交叉核對下水道 E2 邊界

狀態：`DRAFT`（攻略語意已交叉核對；本機正常輸入的 wall／出口觸發仍未閉合）
日期：2026-08-10

## 本輪結論

使用者提出的 `(13,10) → (8,15)`「傳送門」方向，現在可以更精確地描述為：

```text
下水道 `(13,10)` 騎士事件
  → 南側 E2 外部地圖出口／邊界交易
  → ECL2 block 4 火刀據點入口
```

它不是目前證據支持的 `MOVEPARTY` 角色轉移，也不能先命名成秘密門。`MOVEPARTY`
已由中文說明書列為跨 SSI Gold Box 遊戲的角色資料轉移候選；本輪攻略研究所指的
是同一冒險地城內的 E2 出口，兩者是不同功能。

「傳送門」可以作為玩家容易理解的說法，但規格與程式應使用
`external_map_exit`／「外部地圖出口」等中立名稱，直到同版本 runtime 找到真正的
wall／action trigger。

## 外部攻略證據

本輪以公開英文攻略作為地點與玩家可見語意的交叉參考，不能取代本機原始 ECL、
GEO 與 DOS／PC-98 runtime：

| 來源 | 可支持的內容 | 等級 |
|---|---|---|
| [GameFAQs《Curse of the Azure Bonds》攻略／Tilverton Sewers](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365) | 地圖章節標為 `ECL Script 3`；地點 6 是 `(13,10)` 的 Knights of Myth Drannor 騎士事件；`E2` 的文字定義是通往 Fire Knife Hideout 的出口 | `nearby`（攻略文本的地點／語意；不是本機 raw exact） |
| 同上 | 火刀 Bond 移除後，E1／E2 會被封閉，之後引導至 World Map；這是後續劇情狀態，不可套回尚未移除 Bond 的 block 4 入口 | `nearby`（攻略文本）；與本作 raw branch 的跨版本／狀態對照仍需分開 |
| [GameBanshee Fire Knife Hideout 攻略](https://www.gamebanshee.com/curseoftheazurebonds/walkthrough/firekniveshideout.php) | Fire Knife Hideout 有回到 Tilverton Sewers 的入口；據點內另有「secret passage」，不能把它與下水道 E2 或 `MOVEPARTY` 混為一談 | `nearby`（同作攻略的功能區分） |
| [Weekend Waste Monster 地圖索引](https://www.weekendwastemonster.net/crpgs/curse/cotab_maps.html) | 提供 Tilverton Sewers／Fire Knife Hideout 地圖索引，可作地圖版面交叉參考 | `layout-only`（本輪未把快取失敗的 PNG 當作座標證據） |

GameFAQs 的 ASCII 地圖與本專案解出的 `GEO2.DAX` record 不應直接逐字格合併：
攻略的 E2 標記是玩家導覽語意，本身沒有提供本專案 raw ECL 的
`C04B..C04F`、`NEWECL` 或 wall/detail 寫入證據。因此它支持「這裡是外部出口」的
類型，不單獨支持 `(8,15)` 的 exact producer。

## 本機證據與座標邊界

下列證據已存在於本專案，這一輪只補上攻略語意，不改寫其位址空間：

| 證據 | 結果 | 等級 |
|---|---|---|
| `GEO2.DAX` block 3／16×16 decoded geometry | `(13,10)` 的 terrain／wall record 會進入騎士事件；目前第三平面／門狀態下，普通 movement graph 沒有從 `(13,10)` 到候選 `(8,15)` 的合法路徑 | `exact`（geometry／事件）；「因此必為何種 action」仍 `unknown` |
| ECL2 block 3 南側 boundary branch | `0xC01E`、`Y=0`、`X-=2`，再 `NEWECL 4`；block 4 初始資料要求 `LOAD FILES 4,2,FFh`、`LOAD PIECES 1,2,4`，並進入火刀據點入口文字 | `exact`（raw ECL／remake continuation） |
| 既有測試 | 騎士後直接把 State 設成 `(8,15,S)` 再呼叫 `RunDungeonExitLifecycle`，可以驗證 block 4 初始化；這是 `coordinate-assisted`，不是正常玩家路徑 | `exact`（測試事實）；不可升格為玩家輸入證據 |
| PC-98 `MOVEPARTY` | IDA raw helper／writer 已閉合局部 control flow；中文說明書將產品功能指向跨遊戲角色轉移 | `strong inference`（產品對應）；不是 E2 movement consumer |

目前應保存的分級如下：

- `(13,10)` 是騎士事件：`exact`（本機 GEO／ECL 與攻略相互支持）。
- E2 是通往火刀據點的外部出口：攻略語意為 `nearby`，本機 block 3→4 raw branch
  為 `exact`；攻略不單獨證明本機 `(8,15)` 的格位。
- `(13,10) → (8,15)` 不是關閉狀態下的普通步行，而是需要外部出口／wall action
  或等價 map service：`strong inference`。
- 何種按鍵、哪一個 wall/detail 欄位、是否需要 `S`／`B`／`P`／`K`、ECL flag、
  另一側 cell 與重訪／存檔持久性：`unknown`。

## 不應採用的解讀

1. 不把攻略中的「secret passage」直接套到 `(13,10) → (8,15)`；那是 Fire Knife
   Hideout 內另一個可探索的通道描述。
2. 不把 `MOVEPARTY` 的 `B/P/K` helper 或 raw `01` 寫入命名成「開秘密門」。
3. 不把「火刀 Bond 移除後 E2 導向 World Map」套到本次尚未移除 Bond 的
   `NEWECL 4` block 4 入口；這是不同劇情狀態。
4. 不把 direct-entry 測試或攻略座標文字寫成正常玩家已經能走到 `(8,15)`。

## 下一個最小可玩工作

下一輪只追能讓正常玩家通過這個邊界的最小資料流：

1. 在同版本 DOS／PC-98 runtime 記錄 `(13,10)` 後每個 movement／action 輸入、
   `BLOCKCODE`／`WALLCODE` 結果、`THE3DMAP+300h` before／after 與 ECL memory。
2. 若確認是外部出口，將「來源 map／座標／面向 → 目的 ECL block／座標／面向」
   宣告為 game-pack 的 `external_map_exit`，由 engine 只處理通用 transaction。
3. 由正常鍵盤／Ebiten 輸入抵達同一 boundary，再驗證 block 4 初始化與返回／重訪；
   在此之前保持 `MoveDungeon` 對未知 wall fail-closed，不新增 `secret_door` JSON。

本規格只更新證據路由與記憶，不宣稱 P0-2 或整作完成。P0-2 正式工作項仍見
[`golden-box-reverse-engineering-worklist.md`](../knowledge/golden-box-reverse-engineering-worklist.md)。
