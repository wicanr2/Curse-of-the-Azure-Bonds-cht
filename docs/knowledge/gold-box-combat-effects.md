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
的 PC-98 flags 是 Fire `01h`、Cold `02h`、Electricity `04h`、Magic `08h`；防護 handler
清除 pending damage，但 visual transaction／施法消耗仍要續行。反射線等
作品中立 renderer 應由呼叫端傳入 flags，不在 core 寫死 title spell ID。

是否先擲 saving throw、再清除傷害，會影響後續 PRNG continuation；即使 HP
結果同為零，也必須另外以 runtime trace 驗證 draw order，不能只靠最終數值
宣稱完整 fidelity。

## effect `0Ah` 寒冷抗性：半傷不是免疫

PC-98 Borland symbol `COLDFLG=0002h` 是絕對旗標常數，不是 resident dseg
變數。overlay 12 的 `INITEFFPROX` 以 `DS:A040h + 4 × effectID` 建立 handler
table；effect `0Ah` 寫入 `008B:006Bh`，TPOV resolver 解析到 overlay-local
`029Bh`。該 handler 讀 raw `DAMAGEFLAG`、以 `02h` mask 比較，再把 raw
`DAMAGE` 做整數除二。上述 bytes／table／consumer 閉合為 `exact`；`A02Ch`
同段的加三用途仍是 `unknown`。

同一 overlay 另有 local `2461h` 的 cold damage clear helper，這是完全免疫
邊界，不能和 effect `0Ah` 混成同一規則。這個差異是重製最容易錯的地方：
`Protected=true` 只能表示 damage 被完全清除；effect `0Ah` 應回傳
「已減少、但仍可造成傷害」。

CoAB 現把已閉合的能力宣告在 game pack `combat_affect_rules`：

| stable ID | effect | damage mask | operation |
|---|---:|---:|---|
| `coab.monster_affect_0a.resist_cold` | `0Ah` | `02h` | half |
| `coab.monster_affect_70.fire_immunity` | `70h` | `01h` | immune |
| `coab.monster_affect_87.electricity_immunity` | `87h` | `04h` | immune |

共用 engine `combat/damage` 不解讀作品名稱，只做 raw effect kind、mask 與
operation 的資料驅動比對。Battle 建立與讀檔重建都由 title adapter 重新掛入
規則；save 不保存 engine rule slice。未來其他 SSI Gold Box 作品可沿用 schema
與 pure resolver，但必須重新證明自己的 effect table、旗標常數與 handler，不能
直接複製 CoAB 的數字。

## `CHARREC` alignment 與條件式 effect

CoAB PC-98 的 Borland symbol／type table 已關閉 shared `CHARREC` 欄位：

| raw offset | member | 可重用提醒 |
|---:|---|---|
| `+11Ah` | `RACETYPE` | 不能再命名成 `MonsterType` |
| `+11Bh` | `ALIGNMENT` | `0..8` enum；零是合法值，必須另存 known bit |
| `+14Ch` | `MONSTERTYPE` | 與 `RACETYPE` 是不同欄位 |

overlay 12 effect `08h／09h` 讀 `DS:9594h` 的 active character，分別比較
evil `{2,5,8}`／good `{0,3,6}`；命中後對 `SAVEROLL` `A02Ch` 加二、對
`ROLLTOHIT` `A039h` 減二。這種 handler 的通用資料流是：

```text
defender effect list + interacting character value
    -> conditional rule resolver
    -> attack-roll / saving-throw transaction delta
```

因此引擎規則必須把 effect owner 與互動角色分成兩個輸入；未知 alignment
不得猜成 true neutral 或依怪物名稱推導。共用 engine 的 `combat/modifier`
只負責 active raw kind、值集合與 delta，CoAB adapter 才把 `CHARREC` offset
投影進去。下一款 SSI 遊戲若使用相同 effect 編號，仍要重新驗證自己的
handler table、欄位 offset 與 caller。

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

## 持續效果 writer 與魔抗順序

CoAB PC-98 `EFFECTS` module 的 resident `013E:0089h` 必須以 overlay 23
control segment 解析；exact 落到 entry 21、local `2325h PUTEFFECT`。相同
stub offset 若套到 overlay 12 會得到無關 handler，不能跨 module 合併。

對 Sleep 這類已先建立 ordered targets 的持續效果，原版順序是：

1. 專屬 handler 先消耗容量／篩選候選；
2. `PUTEFFECT` 設 current effect，再以 `CHECKFX(target,9)` 觸發既有防護；
3. 魔抗成功只阻止 record 寫入，不退還已消耗的容量；
4. 成功才由 `ADDEFFECT` 寫九 byte record 並連入效果鏈。

這形成可重用的 transaction boundary：作品中立引擎可保存「filter → resist →
write」順序，但 effect ID、duration、`CASTON`、record payload、訊息與動畫都要由
各 game pack 的 bytes／規格提供。不得把 CoAB 的 `35h` 或五倍等級直接當成
所有 Gold Box 的共同常數。

## held effect 的 Action 與呈現生命週期

持續效果 callback 與 `PUTEFFECT` 的訊息／動畫是兩個不同 boundary。CoAB
PC-98 effect `33h／34h／35h` 的 add／remove callback 都落到作品中立語意
`CLEARACTION`；它清 pending spell、target、move、guard，但不處理法術格。
傷害 transaction 必須先執行 effect removal callback，再決定是否仍有施法可
中斷，否則會把同一 pending spell 重複消耗。

成功 `PUTEFFECT` 才依 non-empty message 呼叫 `TWINKLE`。成功圖示代號
`16h`、失敗 `17h` 是 runtime 動態圖示，不是 DAX block；wrapper 以四筆 writer
建立 24×6 圖示。成功分支逐目標送 `SPELLHITFX`，魔抗提前返回則沒有閃爍。
移除 callback 也沒有文字／聲音／圖形，因此共用 engine 不應以「狀態解除」
自行推導醒來特效。作品 adapter 應分別宣告 add presentation、remove callback
與 natural-expiry transaction。

`GAMESPEED=4` 的成功閃爍 delay-only 是 1440ms；此值不含 draw overhead。
若缺 runtime capture，幾何與時間可以標 exact，palette pixels 仍只能標
layout-reconstructed。

## 可重現入口

- `docs/spec/410-pc98-monster-affect-loader-and-tyranthraxus-detect-invisible.md`
- `docs/spec/411-pc98-tyranthraxus-magic-resistance.md`
- `docs/spec/412-pc98-tpov-entry-stubs-and-tyranthraxus-effects.md`
- `docs/spec/413-tyranthraxus-fire-electric-protection-runtime.md`
- `docs/spec/414-pc98-post-hit-effect-4f-runtime.md`
- `docs/spec/498-pc98-resist-cold-data-driven-affect-rule.md`
- `docs/spec/434-pc98-sleep-puteffect-magic-resistance.md`
- `docs/spec/442-pc98-sleep-action-clear-and-twinkle.md`
- `scripts/ida/pc98_monster_affect_loader_audit.idc`
- `scripts/ida/pc98_attack_effect_phase_audit.idc`
