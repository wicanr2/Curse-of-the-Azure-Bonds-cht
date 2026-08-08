# 第四百九十九輪：PC-98 alignment 與條件式 effect `08h／09h`（READY）

## 本輪範圍

本輪關閉兩個 `MON*SPC` effect handler 的資料流，並把它們接到共用
Golden Box engine 與 CoAB JSON。這是「防禦方持有 effect、互動對手提供
alignment」的條件式攻擊／存豁免修正；不是把所有怪物預設成善良或邪惡。

完整 alignment 規則、所有 effect 生命週期、PC-98／DOS PRNG 等價性與完整遊戲
仍未完成。

## 輸入、工具與位址基準

| 證據 | 值 |
|---|---|
| PC-98 `PC98-GAME.EXE`／Borland symbol input SHA-256 | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| PC-98 `PC98-GAME.OVR` SHA-256 | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` |
| overlay 12 code-only 分析副本 SHA-256 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` |
| overlay 23 code-only 分析副本 SHA-256 | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` |
| 主要工具 | IDA Pro 9.4 Docker image `ida-pro-9.4-ver2:uidfix-v1`、Borland symbol parser |
| 位址基準 | overlay-local offset、Borland `segment:offset` 與 ECL work address 分開記錄 |

原始 executable、overlay、symbol table 與基準 `.i64` 均未改寫。IDA report 只
對 Docker `/tmp` 的分析副本執行，外部語意仍保留原始位址與 raw bytes。

## `INITEFFPROX` table 與 handler

overlay 12 的 `INITEFFPROX` 以 `DS:A040h + 4 × effect_kind` 建立四 byte
far-pointer slot。IDA 連續指令證明：

```text
effect 08h: DS:A060h/A062h <- 008Bh:0061h
effect 09h: DS:A064h/A066h <- 008Bh:0066h
```

同一份 overlay-local report 的 handler bytes：

```text
effect 08h，local 022Dh..0261h
0230  les di, ds:9594h
0234  cmp byte ptr es:[di+11Bh], 02h
0240  cmp byte ptr es:[di+11Bh], 05h
024C  cmp byte ptr es:[di+11Bh], 08h
0254  add byte ptr ds:0A02Ch, 02h
0259  sub byte ptr ds:0A039h, 02h

effect 09h，local 0264h..0298h
0267  les di, ds:9594h
026B  cmp byte ptr es:[di+11Bh], 00h
0277  cmp byte ptr es:[di+11Bh], 03h
0283  cmp byte ptr es:[di+11Bh], 06h
028B  add byte ptr ds:0A02Ch, 02h
0290  sub byte ptr ds:0A039h, 02h
```

Borland symbols 將兩個 work cell 精確命名為：

| 原始定位 | Borland symbol | handler 操作 | 等級 |
|---|---|---|---|
| overlay 12 `DS:A02Ch` | `SAVEROLL` | `+2` | `exact` |
| overlay 12 `DS:A039h` | `ROLLTOHIT` | `-2` | `exact` |
| `DS:9594h` | `ACTIVECHARACTER` | handler 當下讀取的互動角色 far pointer | `strong inference`；由 dispatcher／consumer 閉合用途 |

因此 effect `08h` 的條件值是 `{02h,05h,08h}`，effect `09h` 的條件值是
`{00h,03h,06h}`；兩者命中時都產生 `attack_roll_delta=-2` 與
`saving_throw_delta=+2`。值集合與加減操作是 `exact`，共用 engine 將誰視為
「互動角色」留給呼叫端傳入，避免把 effect owner 與 active character 混成一個
參數。

## `CHARREC` 欄位修正

Borland type table 解析 `CHARREC` type `992`，size `01A7h`；相鄰 member
records 的 sequential layout 交叉證明：

| offset | 原始 member | 等級 |
|---:|---|---|
| `0119h` | `GENDER` | `exact` |
| `011Ah` | `RACETYPE` | `exact` |
| `011Bh` | `ALIGNMENT` | `exact` |
| `014Ch` | `MONSTERTYPE` | `exact` |
| `019Ah` | `THAC0` | `exact` |
| `019Bh` | `AC` | `exact` |

alignment enum 也由 symbol table 關閉：

```text
0 LAWFUL_GOOD       1 LAWFUL_NEUTRAL  2 LAWFUL_EVIL
3 NEUTRAL_GOOD      4 TRUE_NEUTRAL    5 NEUTRAL_EVIL
6 CHAOTIC_GOOD      7 CHAOTIC_NEUTRAL 8 CHAOTIC_EVIL
```

這推翻 spec 418 將 `+11Ah` 稱為 `MonsterType` 的舊欄位命名。spec 418
仍保留 visibility handler 的 bytes 與 `13h` 比較，但其欄位名稱已由本規格
supersede：現在 `+11Ah` 是 `RACETYPE`，`+11Bh` 是 `ALIGNMENT`，真正的
`MONSTERTYPE` 是 `+14Ch`。舊名稱只在 Go 的相容欄位中保留，不能再當成新證據。

## Engine／CoAB 契約

共用 engine 新增 `combat/modifier`：

- `Rule` 保存 stable ID、raw effect kind、opaque predicate、值集合與兩個 delta；
- `Resolve` 只比較 active raw effect kind 與已知互動值；未知值 fail-closed；
- engine 不讀 `CHARREC`、怪物名稱、中文或 CoAB alignment enum；
- `MatchedRuleIDs` 只作診斷，不能取代 raw effect／位址證據。

CoAB `pit-of-moander.json` 宣告：

| stable ID | effect | 互動值 | 命中效果 |
|---|---:|---|---|
| `coab.monster_affect_08.protection-from-good` | `08h` | `2,5,8` | 命中 `-2`、存豁免 `+2` |
| `coab.monster_affect_09.protection-from-evil` | `09h` | `0,3,6` | 命中 `-2`、存豁免 `+2` |

`monster.Record` 與 DOS player parser 現在保存 `RaceType`、`Alignment`、
`RawMonsterType`；舊 `MonsterType` 只作 `+11Ah` 的相容別名。`Fighter` 以
`AlignmentKnown` 區分合法零值與未知值，`Battle` 在物理攻擊、Fireball、
反射線法術、Stinking Cloud 與 Cloudkill 的存豁免邊界，將「防禦方 effect」
和「施法者／攻擊者 alignment」分開解析。規則於開戰與 active-combat restore
由 game pack 重新掛入，不進 save JSON。

## 驗證與邊界

Docker focused gate 已通過：

```text
engine: ./combat/modifier ./engine
CoAB: ./gamepack ./internal/combat ./internal/monster ./internal/party ./internal/game
```

測試涵蓋：

- effect `08h／09h` 值集合與 `+2／-2` resolver；
- 未知 alignment 與 inactive effect 不命中；
- `+11Ah／+11Bh／+14Ch` parser offset 與相容別名；
- 敵方持有 effect `09h`、善良互動角色攻擊時只在 attack boundary 產生 `-2`。

本輪仍未宣稱：完整 protection spell duration／dispel、所有 ECL DAMAGE save
caller、PC-98／DOS exact RNG、完整 effect 動畫／音效、完整戰鬥 AI、完整地圖、
全劇情與全作通關。
