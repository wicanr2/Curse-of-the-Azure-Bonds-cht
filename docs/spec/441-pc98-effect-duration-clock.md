# PC-98 效果持續時間時鐘與睡眠到期

狀態：`READY`

## 先前問題

第 434 輪只證明 Sleep `35h` 寫入 `duration=5×caster level`；第 440 輪只
證明正傷害會解除。remake 原有 `AdvanceMonsterAffects(minutes)`，但沒有在
戰鬥 round boundary 呼叫，且把參數稱為分鐘，沒有 executable 證據支持。

## 原始證據

輸入：

- `PC98-GAME.EXE` SHA-256
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`
- code-only overlay 20 SHA-256
  `5dd0bd1c19f4358ea0478cd95fdb05890621d914e9844769c6590f4abd715271`

IDA 9.4 只在 `/tmp/coab-ida-441-clock` 與
`/tmp/coab-ida-441-effects` 的 writable copies 重新分析；原始位置未修改。
overlay 20 報告 1,276 行，SHA-256
`851d28efedd8f0e3d31e1cccacdc9baae3f1ceabb514c200cf4edbb20c0c7cbf`；
overlay 23 報告 2,221 行，SHA-256
`60c920a29fb6f7b0007f190bfd61c6657b4c9f059be48f84266810d86f572d26`。

Borland symbols／types：

- module `CLOCK_`，`TICKCLOCK 0105:03B9h`；
- `TIMEUNITS`：`SEGMENTS=0、ROUNDS=1、TURNS=2、HOURS=3、DAYS=4、
  MONTHS=5、YEARS=6`；
- `MAXCOUNT 0C29:6804h`。

resident `MAXCOUNT` 六個 word exact 是 `10、10、6、24、30、12`。overlay 20
local `0020h` 的連續 IDA 指令：

1. 依輸入 unit 逐級乘 `MAXCOUNT[(unit-1)]`，換成 effect duration tick；
2. 單次最多處理十 tick；
3. 從 `CHARACTERLIST 9598h` 沿 `CHARREC.NEXT +18Ah` 遍歷全隊；
4. 從 `Player+F2h` 沿 `EFFECTREC+5h` 遍歷效果；
5. `EFFECTREC+1h=0` 直接保留；
6. duration 大於本批 tick 時在 `01BBh` 做 word subtraction；
7. 到期時在 `01EFh` 呼叫 resident `013E:002Fh`，typed overlay resolver
   對應 `SPELLOFF`，執行 callback、解鏈與釋放。

`TICKCLOCK 03B9h` 最後以原 unit／amount 呼叫 local `0020h`。Sleep 的 writer
把 `5×level` 直接放入 duration，因此其基底是 round tick；每個後續 combat
round boundary 扣一。

以上 instruction、enum、table、linked-list traversal 與 subtraction 是
`exact`。把「十 tick chunk」視為等價整數 subtraction 是 `strong inference`：
Sleep 的到期只依 record 消失，目前沒有逐 chunk 的玩家可見 callback 反例。

## Remake 對應

- 獨立 engine `combat/effecttime` 保存 `TIMEUNITS` 與 exact 換算；
- `Battle.initializeRoundDelays` 每個 round boundary 扣一 duration tick；
- duration 零依原版保留，不再把 innate／無期限 record 在第一回合誤刪；
- State 世界時間也經同一作品中立換算後再扣 party／battle effects；
- 正常手動 Sleep 玩家路徑證明 level 3 寫入 15，施法 handoff 進入下一 round
  後剩 14，再於總計第 15 tick 解除 held。

## 保存邊界與未完成

目前 remake 不支援戰鬥中存檔，故本輪沒有虛構 active-battle save round-trip。
原版是否允許、如何保存戰鬥中 effect linked list 仍是 `unknown`。另外尚缺：

- Sleep 到期時 `CALLEFFECT` 的文字、twinkle、音訊與精確畫面時序；
- `SEGMENTS=0` 與 `ROUNDS=1` 都不換算的設計原因；
- DOS 版本同 routine 的 byte-level 交叉驗證；
- 其他 18 個 `REMOVEFX` 效果的逐項 gameplay callback。
