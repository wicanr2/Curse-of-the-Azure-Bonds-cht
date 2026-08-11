# 第五百四十七輪：DOS 虛擬地圖暫存器與眼魔洞穴 E1 傳送

狀態：`SUPERSEDED`（`C04B..C04F` 的 raw bridge 仍有效；A2 事件時序、同一 ECL
result 的 continuation 與目前玩家路徑請見第 548 輪）
日期：2026-08-11

> 2026-08-12 勘誤：本規格把 A2 觸發後的 `C04B/C04C/C04D=13/1/3` 誤寫成
> 立即 position transaction。原始 ECL4 trace 證明玩家必須先經三次 `PRESS`
> boundary，才在 `+061B` 寫入座標並進死精靈選單；初始 `4C03=1` 也不會在
> `LEAVE` 時清除。舊的 raw IDA map-register bridge 保留可回查，但洞穴 route
> 與驗收敘述已由
> [第 548 輪](548-ecl4-cave-a2-continuation.md)取代。

## 結論與勘誤

本輪以 DOS 原始 overlay 的 IDA Pro 稽核補上 ECL 虛擬地圖暫存器的
producer／consumer bridge：

- ECL 寫入 `C04B`、`C04C`、`C04D` 會分別寫入 live map 的兩個 0..15 座標欄位
  與 half-facing；後者的 `0/1/2/3` 會轉成 renderer 使用的 `0/2/4/6`。
- ECL 讀取 `C04E`、`C04F` 則分別讀回 map accessor 已計算的兩個 byte。既有
  vector 證據與 GEO 四平面格式將它們閉合為目前朝向的 wall cache 與目前格子的
  `x2`／terrain byte；不是另一組劇情 selector。
- 因而，ECL4 block `0x22` 把 `C04B=13`、`C04C=1`、`C04D=3` 寫出的交易是
  真正的虛擬地圖位置 handoff，不是可隨意忽略的 scratch 值。CoAB game-pack
  用中立的 `set_map_position` 投影這個作品資料；engine 沒有加入洞穴名稱、劇情
  或 D&D 規則。

第 544／545 輪把洞穴 E1 錨點寫成 `(4,5,N)` 是錯誤的結論。本輪依公開攻略的
E1 `(5,7)` 與同一正常新遊戲 session 交叉核對，改為 `(5,7,W)`；朝向仍是
`strong inference`，不可把它升格為 DOS runtime pixel trace。E1 的正常走格會
抵達原始 GEO cell `(5,9)`，原始 ECL transaction 再把玩家帶至死精靈格
`(13,1,W)`。目的格的 `wall=08`／`terrain=0xC0` 是原始 GEO byte，必須和位置
一起寫回 State，否則重繪會留用傳送前的畫面快取。

這份規格 supersede 第 544／545 輪關於「`(4,5,N)` 是洞穴入口」的斷言，也取代
本輪草稿中把已證實位置交易誤當成不可投影 scratch 的說法。它不改變
`0x4C00=unknown` 的結論。

## 原始 IDA 證據

所有分析在一次性 Docker 容器內以 IDA Pro 9.4 進行；原始 `GAME.OVR`、
baseline `.i64` 與 extracted overlay 都保持唯讀，IDA 只分析 disposable copy。

| 項目 | 值 | 位址空間／等級 |
|---|---|---|
| 遊戲封存檔 | `curseoftheazurebonds.zip` SHA-256 `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | 原始輸入，`exact` |
| 輸入 overlay | `workplace/ida406/overlays/overlay-07.bin` SHA-256 `5483c71f98c5dc668d7d307c18a6b071dcfc42fcba9d62eccb657600e7265125` | `GAME.OVR` overlay-07 local offset，`exact` |
| 工具 | `ida-pro-9.4-ver2:uidfix-v1`，IDA Pro 9.4 | disposable database |
| 稽核腳本 | [`dos_overlay07_c04bf_projection_audit.idc`](../../scripts/ida/dos_overlay07_c04bf_projection_audit.idc) SHA-256 `a64810f57a6f9255f51a9bac4310d21ec586000f0605dac42c0c6793ebee3e6d` | 非破壞性 raw／disassembly 匯出 |

IDA overlay-local `0x0D70..0x0F37` 是 setter 的連續範圍。它先將 ECL work
address 減去 `BF68h`，再依如下 offset 分派：

| ECL work address | 原始 local bytes／結果 | 等級 |
|---|---|---|
| `C04B` | `0E92: 3DE300`、`0E9F: A20F72`，將低 byte 寫入 `DS:720F` | `exact` |
| `C04C` | `0EA8: 3DE400`、`0EAD: A21072`，將低 byte 寫入 `DS:7210` | `exact` |
| `C04D` | `0EBA..0F04` 將 `0/1/2/3` 寫成 `DS:7211=0/2/4/6` | `exact` |
| `C04E`、`C04F` | 此 setter 沒有把任意 ECL 寫入直接存進這兩個 cache | `exact`；不得把任意 ECL 寫入當作地圖真相 |

getter 的 overlay-local `0x0FDC..0x108A` 對 `C04B` 作門檻比較、減去 `C04B`，
再讀回 `DS:720F`、`DS:7210`、`DS:7211/2`、`DS:7212`、`DS:7213`。因此
`C04B..C04F` 與五個 live map byte 的橋接是 `exact`；既有
[`spec 520`](./520-dos-movement-to-overlay-cell-layer-bridge.md) 的 vector 4／6
與 [`spec 524`](./524-dos-overlay30-geo-loader-source.md) 的 GEO `+000/+100/+200/+300`
四平面證據，支持下列可執行 adapter 名稱：

| ECL work | live byte | 附加語意 | 等級與限制 |
|---|---|---|---|
| `C04B` | `DS:720F` | map X | `strong inference`：normal movement 以 0..15 wrap 更新，座標軸由結果／GEO index 交叉支持 |
| `C04C` | `DS:7210` | map Y | `strong inference`：同上 |
| `C04D` | `DS:7211 / 2` | half-facing | `exact`（轉換） |
| `C04E` | `DS:7212` | facing wall cache | `strong inference`：vector 4 cell-layer consumer 已閉合；不可當永久 door state |
| `C04F` | `DS:7213` | current cell `x2`／terrain byte | `exact`（getter＋vector 6 `+200h` GEO plane）；ECL 可用高 bit／遮罩作事件分派，但 byte 本身不是獨立 event table |

這些是 overlay-local offset、ECL work address 與 DS-relative offset 三種不同位址
空間，文件刻意並列而不把它們改名或混成同一數字。

## E1 正常玩家路徑

| 項目 | 證據 | 等級 |
|---|---|---|
| E1 錨點 `(5,7)` | [GameFAQs walkthrough](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365) 的 Cave of the Beholder E1 標示；同一 CoAB normal session 已抵達該 GEO anchor | `strong inference`（公開攻略＋remake contract，不是 DOS runtime capture） |
| E1 朝向西 | ECL block handoff／既有 adapter 的 raw half-facing 交易與 session result | `strong inference` |
| 起始後的 source cell `(5,9)`／terrain `A2` | `GEO4.DAX` block `0x25` decode；`MoveDungeon` 真正走格 | `exact`（GEO）／`exact`（remake normal path） |
| ECL position transaction | 正常 session 的原始 ECL4 block `0x22` 執行後為 `C04B=13`、`C04C=1`、`C04D=3`；上表 IDA 證明該三字會更新 live map | `strong inference`（原始 ECL decode＋exact VM bridge） |
| 目的格畫面 | `GEO4.DAX` block `0x25` cell `(13,1)`：west wall `08`、terrain `C0` | `exact` |
| CoAB 結果 | `zhentil-keep.beholder-cave.same-block-launch` 將位置與 cache 投影為 `(13,1,W)`、`wall=08`、`roof=C0` | `exact`（remake contract） |

`TestRealNewGameContinuesFromHapToBeholderCaveEntrance` 現在從新遊戲一路經 Hap、
熔岩洞、法師塔、世界旅行與散提爾堡進入 E1，再以一般 `MoveDungeon` BFS 走到
source cell。測試斷言 data-pack event、最終位置、方向、`C04B..C04F` 與原始
GEO cache，沒有 direct-entry、座標注入、戰鬥注入或 `0x4C00` 寫入。

Docker 驗證：

```text
go test -modfile=go.round.mod -mod=readonly -timeout=240s \
  ./gamepack -run '^TestEmbeddedPackValidatesAndOwnsZhentilText$' -count=1
go test -modfile=go.round.mod -mod=readonly -timeout=240s \
  ./internal/game -run '^TestRealNewGameContinuesFromHapToBeholderCaveEntrance$' -count=1
```

兩者於本輪通過。`go.round.mod` 是本機 engine replace 的測試輔助，不是版本化
產品檔案。

## 明確邊界與下一步

- 這只接通 E1 到死精靈格的傳送；Dexam、Medusa、兩場戰鬥、出口、重訪與世界
  地圖目前仍主要有 coordinate-assisted fixture，不可改稱完整洞穴。
- 目前不需要深挖 `0x4C00`；它不影響此手段的 D&D 規則、地圖 handoff 或存檔。
- 下一個主線工作是從 `(13,1,W)` 以正常輸入找出 Dexam 觸發與出口，而不是把
  `(15,1)` 等 fixture 座標直接當成已驗證路徑。
