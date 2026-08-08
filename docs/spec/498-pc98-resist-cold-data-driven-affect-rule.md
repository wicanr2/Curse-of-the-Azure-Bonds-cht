# 第四百九十八輪：PC-98 寒冷抗性與資料驅動傷害效果

狀態：`READY`（有界功能規格；不是完整戰鬥或完整遊戲完成聲明）

日期：2026-08-09

## 本輪結論

本輪把 PC-98 `MON*SPC` effect `0Ah` 的已證明效果接到可重用
Golden Box engine：當待處理傷害帶有寒冷旗標時，effect `0Ah` 將傷害整除二；
它不是完全免疫。CoAB game pack 同時把既有已閉合的 effect `70h` 火焰免疫與
`87h` 電擊免疫宣告成 JSON 規則，戰鬥開始與 active-combat save restore
都從 game pack 重新注入規則。

引擎只處理「raw effect kind＋damage mask＋operation」；effect ID 的故事名稱、
怪物資料與法術文字仍由作品 game pack／adapter 提供。這避免把 CoAB 的數字或
中文直接寫進共用 engine。

## 輸入、工具與位址基準

| 證據 | 值 |
|---|---|
| PC-98 `PC98-GAME.EXE` SHA-256 | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| PC-98 `PC98-GAME.OVR` SHA-256 | `de87ea5311e1e2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` |
| overlay 12 code-only 衍生物 SHA-256 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` |
| 主要工具 | IDA Pro 9.4 Docker image `ida-pro-9.4-ver2:uidfix-v1`、`pc98-ovr-audit` |
| IDA 位址空間 | overlay code-only copy 的 overlay-local offset；不等同 file offset 或 resident linear address |
| 原始語意附加腳本 | `scripts/ida/pc98_monster_affect_loader_audit.idc`、`scripts/ida/pc98_symbol_query_audit.idc`、`scripts/ida/pc98_combat_effect_semantics_audit.idc` |

原始 executable、overlay、Borland symbol 與基準 IDA database 均未改寫。IDA
重新分析與 `del_items` 只作用於 code-only 分析副本；本規格保留原始
overlay-local 位址，沒有以推測名稱覆蓋 `sub_XXXXX`。

## PC-98 靜態證據

### 1. 傷害旗標

Borland type／symbol parser 解析到以下絕對常數（位址欄是 Borland
`segment:offset`，不是 dseg global）：

| 原始 symbol | 原始值 | 推論等級 |
|---|---:|---|
| `FIREFLG` | `0001h` | `exact` |
| `COLDFLG` | `0002h` | `exact` |
| `ELECFLG` | `0004h` | `exact` |
| `MAGICFLG` | `0008h` | `exact` |
| `ACIDFLG` | `0010h` | `exact` |

因此不能把 `COLDFLG=0002h` 當成 resident dseg 的 `01C292h` 變數。後者是
錯誤的資料位址假設；`pc98_symbol_query_audit.idc` 現只輸出 dseg symbol
的直接 data xref，並明確警告不能用來當作 `COLDFLG` consumer。

### 2. effect `0Ah` 的 table → far pointer → handler

PC-98 `INITEFFPROX` 的 effect table base 是 overlay 12
`DS:A040h`，直接寫入證明 slot 公式為 `A040h + 4 × effectID`。所以
effect `0Ah` 使用 `A068h/A06Ah`，其連續指令為：

```text
overlay 12 local 2F4C: B8 6B 00       mov ax, 006Bh
overlay 12 local 2F4F: BA 8B 00       mov dx, 008Bh
overlay 12 local 2F52: A3 68 A0       mov ds:0A068h, ax
overlay 12 local 2F55: 89 16 6A A0    mov ds:0A06Ah, dx
```

`pc98-ovr-audit -resolve-stub 12:006B` 將 resident far pointer 解析為
overlay 12 entry 15、local `029Bh`。這個解析先證明 control segment，沒有
把 raw addend 當作 handler 位址。

### 3. local `029Bh` 的效果

IDA 在 overlay-local `029Bh..02B5h` 的連續 bytes 顯示：

```text
029E  mov al, ds:0A02Fh       ; raw DAMAGEFLAG work cell
02A1  and al, 02h             ; COLDFLG
02A3  or  al, al
02A5  jz  ...
02A7  mov al, ds:0A02Eh       ; raw DAMAGE work cell
02AC  cwd
02AD  mov cx, 2
02B0  idiv cx
02B2  mov ds:0A02Eh, al
02B5  add ds:0A02Ch, 3
```

由原始 bytes、旗標常數與完整 consumer 連起來，可得：

> effect `0Ah` 且 `DAMAGEFLAG & COLDFLG != 0` 時，`DAMAGE` 變成
> `DAMAGE / 2`（整數除法）。

這項效果為 `exact`。`A02Ch` 被加三的用途目前只標為 `unknown`；不能把它
命名成回合、聲音或訊息計數器。

另有 overlay 12 local `2461h`（stub `0228h`）會在同一 cold flag 上清除傷害，
但那是另一個 generic protected／完全免疫 helper，不是 effect `0Ah`。兩者
不得因為都讀 `DAMAGEFLAG & 02h` 就合併。

### 4. 既有完整免疫規則

effect `70h` → fire flag `01h` 與 effect `87h` → electricity flag `04h` 的
handler／runtime 邊界已由規格 413、414 及現有戰鬥測試閉合。本輪只把它們的
operation 從 CoAB Go 判斷搬到相同資料契約；沒有重新猜測它們的語意。

## Engine／game-pack 契約

共用 engine 新增 `combat/damage`：

- `Rule`：stable ID、opaque `EffectKind`、`DamageMask`、`Mode`；
- `ModeHalf` 與 `ModeImmune` 分開；半傷結果 `Reduced=true` 但
  `Immune=false`；
- `Resolve` 只比對 active raw effect kind 與 damage mask，不讀怪物名稱、
  中文或劇情旗標；
- JSON schema 的 `combat_affect_rules` 只允許 `immune`／`half`，並驗證 ID、
  effect／mask binding 不重複。

CoAB `pit-of-moander.json` 現宣告：

| stable ID | raw effect | mask | operation |
|---|---:|---:|---|
| `coab.monster_affect_0a.resist_cold` | `0Ah` | `02h` | `half` |
| `coab.monster_affect_70.fire_immunity` | `70h` | `01h` | `immune` |
| `coab.monster_affect_87.electricity_immunity` | `87h` | `04h` | `immune` |

`State.StartCombat` 建立 Battle 後套用一次；`restoreActiveCombat` 重建 Battle
後再套用一次。`Fighter.DamageRules` 為 `json:"-"`，不把 engine 設定複製進
save；save 只保存 raw fighter／effect state，載入時依目前 title pack 重建規則。

## 驗證與邊界

Docker focused gate 已通過：

- engine `./combat/damage ./engine/...`；
- CoAB `./gamepack ./internal/combat ./internal/game`；
- 新增 0Ah 17→8 非免疫單元測試；
- 新增 game-pack → `StartCombat` → active save snapshot → restore 的規則重掛
  測試。

本輪仍未宣稱：Resist Cold 法術／所有寒冷來源的施法入口、effect `08h/09h`
的完整語意、完整 spell animation／sound timing、所有怪物能力、完整 ECL、
完整玩家通關或完整 remake。下一次若要接上寒冷法術，必須先證明該 caller 的
damage flag、save／抗性順序與 visual／sound handoff，不能因 `COLDFLG` 已閉合
就把所有 cold spell 直接標成完成。

## 可重現入口

- [`docs/knowledge/gold-box-combat-effects.md`](../knowledge/gold-box-combat-effects.md)
- [`docs/spec/413-tyranthraxus-fire-electric-protection-runtime.md`](./413-tyranthraxus-fire-electric-protection-runtime.md)
- [`docs/spec/414-pc98-post-hit-effect-4f-runtime.md`](./414-pc98-post-hit-effect-4f-runtime.md)
- [`scripts/ida/pc98_monster_affect_loader_audit.idc`](../../scripts/ida/pc98_monster_affect_loader_audit.idc)
- [`scripts/ida/pc98_symbol_query_audit.idc`](../../scripts/ida/pc98_symbol_query_audit.idc)
