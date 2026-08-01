# Gold Box 戰鬥效果反編譯知識庫

本文保存可沿用到其他 SSI Gold Box 作品的方法與已證明公式。
作品怪物、劇情旗標與特定 `MON*SPC` records 仍留在各自 game pack。

## 證據分層

- `exact`：原始 bytes、有位址的 IDA 資料流或可重現 runtime trace。
- `strong inference`：公式本體已有一次證據，但 effect ID 到 handler 仍經過
  尚未完整解析的 overlay relocation，並由獨立轉寫交叉支持。
- `hypothesis／unknown`：只有名稱、相鄰 table slot、raw addend 或攻略敘述。

原始位址與語意必須並列；不在 IDA database 用推測名稱覆蓋
`sub_XXXXX`、Borland symbol 或 overlay-local offset。

## MON*SPC template 效果

PC-98 Borland `EFFECTREC` 大小為 9 bytes。`LOADMONSTER` 複製記錄後會
清除與重建 `+5..+8` linked-list pointer，但不改 byte `+4`。因此：

- 原始 template byte `+4=0` 不等於能力停用。
- remake 應以來源語意（例如 `Innate`）區分 template 天生能力與角色
  後天 effect 的 active／duration 生命週期。
- 最後四 bytes 在磁碟紀錄中的數值不是傷害、強度或目標參數。

## 魔法抗性 common routine

PC-98 EFFPROCS overlay 的 exact 公式為：

```text
threshold = base + (11 - casterLevel) * 5
resisted  = d100 <= threshold
```

原程式沒有在比較前 clamp threshold。呼叫者先完成原法術的傷害擲骰，
再進入 pre-damage effect boundary；抗性成功會清除傷害，不回滾已消耗的
傷害亂數。已見 wrappers：

| overlay-local | base | 狀態 |
|---|---:|---|
| `23F4h` | 50 | 公式與 wrapper `exact`，effect mapping 待解 |
| `2404h` | 15 | 公式與 wrapper `exact`；`6Ah` mapping `strong inference` |

重製時應將公式寫成作品中立純函式，由 game data／decoded affect
提供 base。每種法術仍要分別證明它的 damage flags、時序、多目標順序與
visual／sound handoff；不可因 common routine 存在就對所有法術全域套用。

## 可重現入口

- `docs/spec/410-pc98-monster-affect-loader-and-tyranthraxus-detect-invisible.md`
- `docs/spec/411-pc98-tyranthraxus-magic-resistance.md`
- `scripts/ida/pc98_monster_affect_loader_audit.idc`
