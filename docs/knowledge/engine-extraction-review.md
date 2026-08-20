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

## 一、可以直接搬（純機制，沒有作品資料）

| 單元 | 現在的位置 | 建議去處 | 邊界 |
|---|---|---|---|
| 八扇形方向分類（`combatDirection`，正切門檻 `0x26A`／`0x6A`，依序試 0..7） | `internal/combat/facing.go` | `combat/facing` | 邊界不對稱是**演算法的一部分**，不能改寫成角度比較 |
| 最短轉法（`turnDistance`） | 同上 | 同上 | 純算術 |
| 90° 扇形判斷（`InFacingCone`／原作 `INARC`，八段判斷式 ＋ 九格位移表） | 同上 | 同上 | 座標上下界 `49×24` 要**參數化**，別的作品戰場尺寸未必相同 |
| AC／命中的刻度換算與命中式（`StoredArmorClass` 一族、`d20 ＋ 命中加值 ＋ AC >= 20`） | 同上 | `combat/armorclass` | `60`／`20` 兩個常數只在 CoAB 量過，見 spec 1139 |
| 敏捷**防禦**調整表（AD&D 1e，`117Ah`） | `internal/monster/armor_class_facing_test.go` 的測試本地表 | 併進 engine 現有的能力值表旁邊（`combat/initiative` 已有姊妹的**反應**表 `120Ah`） | 兩張表值域不同（防禦表 24..25 給 −6，反應表 >25 回 0），不可共用同一支 |
| 體型佔格、重疊、相鄰（`footprint.go`） | `internal/combat/footprint.go` | `combat/footprint` | 尺寸→寬高的對照表要由 caller 給 |
| 解除魔法的機率式（`DispelChance`）與半徑判定（`FighterWithinRadius`） | `internal/combat/spell_dispel.go` | `combat/dispel` | 只搬算式，`CastDispelMagic` 留在作品側 |
| 戰鬥相機（`camera.go`） | `internal/combat/camera.go` | `viewport`（engine 已有） | 27 行，純座標 |
| AI 門檻掃描（`AIThresholdScan`：門檻 7 起、每輪 1d7） | `internal/combat/ai_decision.go` | `combat/aiscan` | 起始門檻與骰面要參數化 |

**最沒有爭議的一項是敏捷防禦調整表**：engine 已經放了同一族的反應調整表，
兩張表放兩邊沒有理由。

## 二、機制可搬、常數留在 game pack

這些的**形狀**看起來是 Gold Box 通則，但目前的數字只有 CoAB 一個證人。
搬的方式是「engine 收規則、作品給表」，與 `combat/damage` 現在的作法相同。

| 單元 | 機制（進 engine） | 常數（留作品） |
|---|---|---|
| 背後攻擊（`RearAttackApplies`） | 三道條件同時成立才換第二個 AC | 門檻 `動作計數 > 1`、`累計轉向 > 4` |
| 第二個 AC 的算式 | 「扣掉敏捷與盾牌那一槽，再固定扣一點」 | `− 2`、以及「盾牌是類別 1」這件事 |
| 開場面向（`ApplyInitialFacing`） | `表[隊伍朝向 div 2]`，敵方再轉 180° | 四筆表 `{7,2,3,6}`（CoAB `DS:2FAh`） |
| 轉向記帳（`AccountTurn`） | 動作計數加一、累計轉向 mod 8 | 無 |
| 機會攻擊的面向閘 | 朝向 −2..＋2 五個方向任一成立就打得到 | 兩個旁路的欄位（先攻、動作計數） |
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

| 順序 | 項目 | 成本 | 為什麼是這個順序 |
|---:|---|---|---|
| 1 | 敏捷防禦調整表 | 一支函式 ＋ 一張表測試 | 姊妹表已在 engine，重複最明顯 |
| 2 | `combat/facing`（三支幾何 ＋ 扇形） | 一個 package ＋ 現有測試搬過去 | 完全沒有作品資料，測試已經寫好可以整批搬 |
| 3 | `combat/armorclass`（刻度換算 ＋ 命中式） | 一個小 package | 下一款作品一定會再撞到同一個 `60 − x` |
| 4 | `footprint`、`DispelChance`、`camera` | 各數十行 | 零風險，適合夾在大批次之間 |
| 5 | 第二節那些「機制 ＋ 表」 | 每項要先定 game pack 欄位 | 介面成本高於程式碼本身，等第 2、3 項的作法定型再做 |

⚠ **跨 repo 的實務成本**：CoAB 的 Go 容器用 `tools/engine-proxy.sh` 把本機那份
engine commit 打包成檔案型 proxy，所以每次搬動都要「engine 側 commit → 重跑
proxy → CoAB 側改 import」。批次搬（一次一個 package）比逐支搬便宜。
engine repo 的 push 需要另外取得授權，不隨 CoAB 一起推。
