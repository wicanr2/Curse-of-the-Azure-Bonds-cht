# 哪些戰鬥函式該搬進共用 engine

`golden-box-remake-engine` 的界線是「作品中立的格式與機制」；劇情旗標、座標、
人物名稱、翻譯、遭遇組成與各作品自己的資料表留在作品 repo。這份盤點把
`internal/combat` 逐檔對這條界線比一次，並標出搬過去之後**常數要從哪裡來**。

判準三條，缺一條就不搬：

1. **不吃作品資料**——沒有 CoAB 的法術編號、效果碼、地形碼、欄位位移。
2. **介面是 plain value**——engine 既有的規則 package（`combat/damage`、
   `combat/modifier`、`combat/scan`…）都吃基本型別或自己的小結構，
   不吃作品的 `Battle`／`Fighter` 聚合型別。
3. **證人數量講清楚**——只有 CoAB 量到的門檻，搬過去要在文件裡寫明只有一個證人，
   不能因為「放進 engine」就升格成 Gold Box 通則。

## 一、已經搬進 engine（2026-08-20）

| engine package | 內容 | CoAB 這一側剩什麼 |
|---|---|---|
| `combat/facing` | 八扇形方向分類、最短轉法、90° 扇形（`InArc`）、九格位移表、轉向記帳、攻擊時的面向寫入、回合開始清計數、背後攻擊三條件、開場面向（表由呼叫端給）、180° 五方向掃描 | `Fighter` ↔ `facing.State` 的轉接，以及 CoAB 專屬的閘 |
| `combat/armorclass` | `60 − 顯示值` 的雙向換算、命中門檻 `Meets`、背後那一格的 `Rear` | 四個同名薄包裝 |
| `combat/ability` | 敏捷**防禦**調整表 | 一個薄包裝 |
| `combat/footprint` | 體型矩形的重疊與相鄰（含對角） | 體型碼 → 形狀的對照表（作品資料） |
| `combat/dispel` | 解除魔法的不對稱成功曲線 | 呼叫端與逐目標流程 |
| `combat/aiscan` | 遞減門檻掃描（輪數的骰先擲） | 起始門檻 7 與 1d7 兩個常數 |
| `viewport.Camera` | 地圖格 → 視窗相對格的平移 | `TilePoint` 的轉接與 `Origin()` |

搬過去的同時補了 engine 側的測試：不對稱扇形邊界、扇形頂點是「起點往前一格」、
`AccountTurn` 問的是哪一個方向、兩種刻度的命中門檻對得上原作的 18、
體型的對角相鄰、解除魔法曲線的不對稱、以及「候選清單空的時候骰仍然要擲」。

⚠ **跨 repo 的流程**：engine 側 commit → `tools/engine-proxy.sh` → CoAB
`tools/go.sh get <印出來的版本>`。engine repo 的 push 另外需要授權，
沒 push 之前 `go.sum` 指到的 commit 只存在本機。

## 二、機制可搬、常數留在 game pack

這些的**形狀**看起來是 Gold Box 通則，但目前的數字只有 CoAB 一個證人。
搬的方式是「engine 收規則、作品給表」，與 `combat/damage` 現在的作法相同。

面向那一族**已經照這個方式搬完**（背後攻擊三條件、第二個 AC 的算式、
開場面向、轉向記帳、五方向掃描都在 `combat/facing` 與 `combat/armorclass`，
門檻與表留在 CoAB）。剩下三項：

| 單元 | 機制（進 engine） | 常數（留作品） |
|---|---|---|
| 障礙地形的兩條豁免路徑（`ai_obstacle.go`） | 「豁免過就走」與「等級夠就走」兩種 | 地形碼 `1Eh`／`1Ch`、兩張效果碼白名單、`>= 7 級` |
| 持續區域格（`persistent_area.go`） | 區域格寫進地圖、到期回收、同時當障礙 | 兩張雲的 offset 表與地形碼 |
| 士氣（`morale.go`） | 四段結果的分段抽籤 | 10／50／20／20 與效果碼 `23h`、`4Ah`／`4Bh` |

## 三、不搬

| 單元 | 理由 |
|---|---|
| `Battle`／`Fighter` 本身 | 作品的聚合型別。engine 的規則一律吃 plain value，把聚合型別搬過去等於把作品模型固定成通則 |
| `checkfx.go`／`checkfx_records.go` | 欄位位移（`+19Ah`／`+19Bh`／戰鬥狀態 `+06h`）是 CoAB 記錄版面 |
| `affect_kinds.go`、`effect_chain.go`、`level_drain.go` | 整份都是 CoAB 的效果碼 |
| `spell_formula.go`、`spell_special.go`、`line_spell.go`、`effect_spell.go` 的具體法術 | 法術編號與骰式是作品資料。**可重用的是「資料驅動」那一層，而那一層已經在 engine** |
| `placement.go` 的編制格 | 目前是本作自訂的 fallback——`SETUP MONSTER` 的距離與 occupancy 表還沒解，還不知道通則長什麼樣 |
| `snapshot.go` | 序列化的是 CoAB 的 `Fighter` 版面 |
| `escape.go` | 逃跑的判準（比雙方移動率、平手 1d2）機制單純，但它直接吃 `Battle` 與 CoAB 的陣營語意，搬過去的收益低於介面成本 |

## 四、已經在 engine 的部分（對照用）

`combat/initiative`、`combat/action`、`combat/damage`、`combat/modifier`、
`combat/resistance`、`combat/posthit`、`combat/scan`／`scanorder`、
`combat/targetselect`、`combat/quickspell`／`quicktarget`、`combat/sleep`、
`combat/effecttime`、`combat/monsterspell`。CoAB 這一側只留轉接。

## 五、搬的順序與成本

第一節那七項已經搬完。剩下的是第二節那些「機制 ＋ 表」——障礙地形的兩條豁免
路徑、持續區域格、士氣分段。它們的介面成本高於程式碼本身：每一項都要先決定
game pack 的欄位長什麼樣，engine 才收得下規則。第一節的作法（plain-value 契約
＋ 呼叫端給表）已經定型，可以照著做。

⚠ **跨 repo 的實務成本**：CoAB 的 Go 容器用 `tools/engine-proxy.sh` 把本機那份
engine commit 打包成檔案型 proxy，所以每次搬動都要「engine 側 commit → 重跑
proxy → CoAB 側 `go get`」。批次搬（一次一批 package）比逐支搬便宜。
engine repo 的 push 需要另外取得授權，不隨 CoAB 一起推。
