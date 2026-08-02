# 第 425 輪：PC-98 Quick Bless 與非即時施法排程

狀態：`READY`（限 casting-delay action handoff 與 Quick Bless）

## 結論

PC-98 `CASTCOMBATSPELL` 對 spell record raw `CastingTime` 執行整數除以 3。
結果為零才立即呼叫 `CASTSPELL`；非零則把 spell ID 寫進目前 Action，將
`Action.delay` 改為 `max(1, delay-castingTime/3)`，釋放目前 selection。
overlay 08 在同一 Action 再次被選中時讀出並清除 spell ID，再呼叫
`CASTSPELL`。這是同輪 scheduler handoff，不是 wall-clock 倒數。

Bless `01h` 的 raw `CastingTime=0Ah`，故 delay 單位為 3；其 Quick
`Priority=1／CastOn=0／MinRange=0`。remake 現可由 `ALT+M＋Quick` 選到
Bless，先顯示吟唱 handoff，待同輪重新入列後才消耗 slot 並套用既有 Bless
效果。Magic Missile raw `01h` 整除後仍為零，保持第 424 輪即時流程。

## 非破壞性輸入

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | spell records | `exact` |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed overlay control | `exact` |
| overlay 08 | `d39b6aad76af8d3ccc182b4b4d95dd9195160c9c21b36bfa5b0cd4edc1942788` | pending action consumer | `exact` |
| overlay 09 | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | Quick suitability | `exact` |
| overlay 13 | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | `CASTCOMBATSPELL` | `exact` |

IDA Pro 9.4 只讀掛載 `workplace/ida406/pc98-overlays/`；新 database 與完整
指令報告寫入 `/tmp/coab-ida-425`。versioned IDC 僅新增外部 scope label，
沒有 rename、patch 或覆寫原始 binary／既有基準 database。

## 指令與資料證據

PC-98 spell table base 是 file `012DE4h`、stride `10h`。本輪加入 JSON 的
raw `CastingTime`（record `+0Ch`）如下；表中只證明資料，不代表效果都完成：

| spell ID | CastingTime | `/3` | 目前狀態 |
|---:|---:|---:|---|
| `01h` Bless | `0Ah` | 3 | Quick pending 已接通 |
| `02h` Curse | `0Ah` | 3 | metadata only |
| `03h` Cure Light Wounds | `05h` | 1 | special suitability 未接 |
| `04h` Cause Light Wounds | `05h` | 1 | target handoff 未接 |
| `06h／07h` Protection | `04h` | 1 | target handoff 未接 |
| `0Fh` Magic Missile | `01h` | 0 | 第 424 輪即時流程 |
| `22h` Stinking Cloud | `02h` | 0 | MinRange 未接 Quick |
| `2Fh／33h` Fireball／Lightning Bolt | `03h` | 1 | Quick target／delay 未接 |
| `5Bh` Cloudkill | `05h` | 1 | Quick target／delay 未接 |

overlay 13 local `2848h..2909h`：

- `2857h..286Eh` 以 spell ID×16 讀 `[61C0h]`，`IDIV 3` 保存 delay units。
- units=0 時，`2877h..288Bh` 呼叫 far `0117:0039h` 的 `CASTSPELL`。
- units>0 時，`289Fh` 將 result 設 1，`28C6h..28D1h` 把 spell ID 寫入
  `Player+18Eh → Action+00h`。
- `28D4h..2905h` 比較目前 `Action+03h delay`：較大則減 units，否則設 1。

overlay 08 local `0428h..046Fh` 讀 `Action+00h`；非零時保存 spell ID 並先
清成零，再以 `(spellID,0,1,&result)` 呼叫同一 far `0117:0039h`。因 selector
已依 mutable delay 重新選到該角色，這是同輪續跑。

overlay 09 local `03F4h..042Ah` 對 spell `03h` 走 Cure special；其他
`CastOn=0` 且 record 條件為零者直接 suitable。Bless record 與此 consumer
一致，故 Quick Bless selection 為 `exact control flow`。

## remake、驗證與推論界線

- engine `combat/quickspell.Spell` 保存 raw `CastingTime`，只提供 `/3` 換算。
- engine `combat/action.BeginSpell／TakeSpell` 保存 pending spell 與 delay
  transaction；不含 CoAB 法術 ID、名稱或效果。
- CoAB game-pack 十一筆 selector records 全部補上 raw `casting_time`。
- Battle regression 證明 delay `8→5`、其他角色 delay 6 先行、caster 再以 5
  入列，之後才取出 spell。
- State regression 以正式 game-pack、持續 PRNG、ALT+M 與 Quick 選到 Bless；
  handoff 時尚未套用效果且 slot 尚在，重新入列後才套用並消耗。
- Standing Stone→紅網正常玩家路徑仍驗證 ALT+M／Quick 即時 Magic Missile，
  確保新增 pending 機制沒有破壞第 424 輪。

目前在「實際結算時」消耗 Bless slot；原版在施法中斷／死亡時是否已消耗仍缺
consumer 與 runtime trace，因此標為 `hypothesis`，不可宣稱 interruption
fidelity。手動 CAST 的 casting delay 與 typed target transaction 已由 READY
spec 428 接續；正傷害中斷與繁中訊息由 429 接續。Quick 的其餘 suitability、
原版中斷畫面 timing／音效與非傷害中斷仍未完成。
