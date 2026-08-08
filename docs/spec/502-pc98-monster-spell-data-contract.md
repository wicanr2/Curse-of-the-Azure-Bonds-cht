# 第五百零二輪：PC-98 怪物特殊法術 `84h` 資料契約

狀態：`READY`
範圍：把第 415／416 輪已完成的 effect `84h` bounded runtime 從 CoAB Battle／
State 常數移入 Golden Box engine＋game pack；不重新宣稱尚未閉合的原版
caster-level、魔法抗性與逐幀演出。

主要證據：

- [`415-pc98-monster-lightning-runtime.md`](./415-pc98-monster-lightning-runtime.md)
- [`416-pc98-monster-spell-target-selection.md`](./416-pc98-monster-spell-target-selection.md)

## 1. 資料契約

CoAB game pack 的 `combat_monster_spell_rules` 目前宣告
`coab.monster_affect_84.lightning-bolt-rounds-1-3`：

| 欄位 | 值 | 推論等級 |
| --- | ---: | --- |
| raw effect | `84h`（JSON `effect_kind=132`） | `exact` |
| spell | Lightning Bolt `33h`（JSON `spell_id=51`） | `exact` |
| 行動窗口 | round 1、2、3；round 4 不再 dispatch | `exact`／`strong inference` |
| target range | `10` | `exact` line budget；range fallback 的來源仍含 `strong inference` 部分 |
| initial damage | `16d6` | `exact` |
| path／reflection damage | 獨立第二份 `16d6` | `exact` |
| damage flags | Electricity `04h`＋Magic `08h`，合計 `0Ch` | `exact` |
| target selection | footprint／牆面／visibility 的既有 `SCAN` bounded selector | `READY`，細部 tie order 未 exact |
| missing terrain | fail-closed，仍消耗怪物回合，不回退物理攻擊 | `exact` contract |

`LineBudget=10`、`InitialDamageDice=16`、`PathDamageDice=16`、
`DamageDiceSides=6`、`DamageMask=0Ch` 與 reflection threshold／penalty 都是
由 pack 傳給既有 `ReflectingLineOptions` 的資料；engine 不把它們命名成所有
Gold Box 法術的共同常數。

## 2. 原始證據與位址基準

輸入雜湊沿用第 415／416 輪：

| 輸入 | SHA-256 | 用途 |
| --- | --- | --- |
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | `ROUND`、symbols、resident calls |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay corpus |
| overlay 9 code | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | type-14 caller |
| overlay 22 code | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` | effect `84h` handler |
| overlay 13 code | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | target／visibility caller |
| overlay 24 code | `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f` | candidate builder |
| overlay 32 code | `c7e2a2b1166676454a54a69330ea318dae8ef9a22e5c5a7ca6e6950ef3609d93` | footprint／terrain reachability |

重要連續證據：

- overlay 9 local `0DF4h..0DFAh` 在一般行動前呼叫 `CHECKFX` type 14；overlay
  22 local `62DDh` 以 `ROUND < 04h` gate，handler 成功後清除 caster actions。
- overlay 22 local `6323h` 選 `33h`，`6374h..637Ah` 產生第一份 `16d6`，
  `638Fh..6398h` 產生獨立第二份 `16d6`，`63A8h..63AEh` 清除 action。
- overlay 13／24／32 的 bounded chain 支持「對方、仍在場、footprint 任一格
  可達、射程內、牆未阻擋」候選；`PICKTARGET` 的二十次不可見重抽與
  `found=false` continuation 已由第 416 輪接入。

反組譯資料以原始 overlay-local offset、bytes、hash 與 IDA 位址空間保存；本輪
沒有改寫原始 executable、symbols 或 `.i64`。

## 3. engine 與 CoAB 邊界

engine 新增 `combat/monsterspell.Rule` 與
`combat_monster_spell_rules` schema／resolver。resolver 只做兩件事：

1. 由 active raw effect kind 找出目前可用的 rule；
2. 以一基底 combat round 篩選 `MaxRound`，保留 game-pack 宣告順序。

CoAB Battle 只保存不可變的 `MonsterSpellRules` 設定，不把它寫進 save JSON；
`StartCombat` 與 active-combat restore 重新從 pack 掛入。State 將 rule 的
target range 傳給既有 `SelectRangedCombatTarget`，將 dice／mask／line budget
傳給既有 line spell core。沒有 terrain callback 或沒有合法候選時，仍沿原本
`84h` no-target continuation 消耗行動，不呼叫普通物理攻擊。

`caster_level=1` 是本輪為了保持既有 remake 行為而保存的 adapter 參數，不是
原版 caster level 的證明；第 416 輪指出 range fallback level `6` 只能支持
`strong inference`。effect `6Ah` 是否攔截 `84h`、初始格／path save 的完整
亂數順序與原版 target delegate 失敗後座標仍是 `unknown`／後續工作。

## 4. 尚未完成與驗證

本輪不宣稱以下內容完成：

- 怪物 Lightning 的 DOS／PC-98 逐幀 windup、travel、impact、音效先後與
  wall-clock timing；
- 原始 combatant-array 同距 tie order、完整 invisibility runtime trace、
  牆角／多次反射畫面；
- `84h` 與 `6Ah` 的完整魔抗／save／damage draw order；
- 其他 SSI 作品是否使用相同 effect kind 或兩個獨立 damage pool。

驗證包含 engine resolver／Pack loader、CoAB game-pack、Battle rule attach、
active save restore、前三回合／第四回合、無目標與 visibility regression；正式
Docker／Xvfb gate 與 `coab-audit` 必須在提交前再次執行。
