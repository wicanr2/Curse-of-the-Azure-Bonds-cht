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
| `2404h` | 15 | 公式、wrapper 與 `6Ah` mapping 均為 `exact` |

重製時應將公式寫成作品中立純函式，由 game data／decoded affect
提供 base。每種法術仍要分別證明它的 damage flags、時序、多目標順序與
visual／sound handoff；不可因 common routine 存在就對所有法術全域套用。

## Turbo Pascal overlay entry 與 fixup

本作 PC-98 TPOV resident control 的固定頭為 `20h` bytes，後接
`EntryCount` 筆五 byte dispatch stub：

```text
CD 3F, overlay-local handler offset:u16le, flags:u8
```

resident far pointer 必須先證明屬於該 control segment，再以
`(stubOffset-20h)/5` 取得 entry；不能把 far pointer offset、overlay local
offset 與 file offset 混用。overlay code 後的 relocation span則是嚴格遞增的
`u16le` segment-fixup offsets，不是 code。decoder 應保留原始 bytes，另外附加
typed entry／fixup view；signature、bounds 或排序錯誤時失敗即關閉。

這個 grammar 已由 36 段 corpus 驗證；作品特定 effect ID 與 handler 對應仍
留在 CoAB spec 412，其他 Gold Box 作品不可直接沿用其數值。

## 元素防護 damage boundary

元素防護應由 raw damage flags 驅動，不可由畫面法術名或怪物名特判。已證明
的 PC-98 flags 是 Fire `01h`、Electricity `04h`、Magic `08h`；防護 handler
清除 pending damage，但 visual transaction／施法消耗仍要續行。反射線等
作品中立 renderer 應由呼叫端傳入 flags，不在 core 寫死 title spell ID。

是否先擲 saving throw、再清除傷害，會影響後續 PRNG continuation；即使 HP
結果同為零，也必須另外以 runtime trace 驗證 draw order，不能只靠最終數值
宣稱完整 fidelity。

## 物理命中後效果排程

PC-98 的攻擊流程不是把所有怪物能力塞進基本命中公式。物理傷害完成後，
caller 先確認目標仍在戰鬥，再把 `attackSlotIndex+1` 交給共用 `CHECKFX`。
本作目前 exact 關閉的 mapping 是：type 2／3 都會搜尋並 dispatch effect
`4Fh`；handler 對既有 attack target 擲 `2d10` Fire＋Magic。

可沿用的引擎設計原則：

- 武器傷害與後續效果要用不同 result records，不能只合成一個 damage 數字；
- effect owner、原攻擊 target、attack slot、raw effect ID 與 damage flags 都要
  明確保存，避免 State 依怪物名稱猜特殊能力；
- post-hit 表是按 check type 分派，不可假設「怪物有能力就每次攻擊都觸發」；
- 物理擊殺不再 dispatch；元素防護即使清除實際傷害，也不能回滾 handler
  先前消耗的傷害骰；
- effect ID／check-type mapping 目前只證明於 CoAB PC-98，不可直接複製到其他
  Gold Box 作品；其他作品應重新解析自身 `CHECKFX` table。

原版動畫、sound cue 與 wall-clock timing 是另一條 presentation oracle，不能
由傷害 handler 自行推定。數值核心完成不等於動態演出完成。

## 可重現入口

- `docs/spec/410-pc98-monster-affect-loader-and-tyranthraxus-detect-invisible.md`
- `docs/spec/411-pc98-tyranthraxus-magic-resistance.md`
- `docs/spec/412-pc98-tpov-entry-stubs-and-tyranthraxus-effects.md`
- `docs/spec/413-tyranthraxus-fire-electric-protection-runtime.md`
- `docs/spec/414-pc98-post-hit-effect-4f-runtime.md`
- `scripts/ida/pc98_monster_affect_loader_audit.idc`
- `scripts/ida/pc98_attack_effect_phase_audit.idc`
