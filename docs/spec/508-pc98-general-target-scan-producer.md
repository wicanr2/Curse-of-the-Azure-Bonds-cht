# 第五百零八輪：PC-98 一般敵方選敵的 SCAN producer（READY）

## 本輪結論

一般敵方物理回合現在可在正式 CoAB game pack 宣告的條件下，沿用 PC-98
TACTICALMAP／footprint／LOSEXISTS producer 建立候選，再進入獨立的有限次數
隨機 consumer。候選不可見時會先移除再抽；第一輪沒有可用候選且規則宣告
`retry_with_xray` 時，會以 `XRay=true` 重建候選清單，對應原版
`find_target(false, 1, 0xFF, player)` 的第二輪牆面忽略邊界。

這是「一般敵方選敵 producer 已接通」的 bounded milestone，不是完整敵方 AI、
移動、方向 comparator、持久 action target、逃跑／防守或完整戰鬥演出的完成
證明。

## 反組譯輸入與可回查資料

所有 IDA Pro 工作都只對 Docker 內的 code-only overlay 副本進行；原始 binary、
基準 `.i64` 與 symbol table 未改名、未覆寫。

| 輸入／工具 | SHA-256／版本 | 位址基準與用途 |
|---|---|---|
| PC-98 `overlay-09.bin` | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | code-only overlay-local `0000h..1BEAh`；一般 caller audit |
| PC-98 `overlay-24.bin` | `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f` | code-only overlay-local；候選 producer |
| `pc98_general_target_caller_audit.idc` | `011b5347585900de06c57c1c0c4a31848b8cd2503dfb0d70c0b8eafeb3cb3ee1` | IDA Pro 9.4 非破壞性連續 caller dump |
| caller audit 匯出 | `cba9f3bbb3a22c393a62fe21ffb0c74274515c6c2b240a3dcf0a040a0385444a` | overlay-09 report；保留 `local`、raw bytes 與 IDA flat address |
| IDA 容器 | `ida-pro-9.4-ver2:uidfix-v1` | 8086、Docker copy address space；不與 resident `segment:offset` 混稱 |

本輪擴充的 IDA 腳本只輸出連續位址與原始指令 bytes，沒有把 `[bp+08h]`、
`[di-622Fh]` 或 `014A:00C0h` 改名成推測欄位。腳本位置：
`scripts/ida/pc98_general_target_caller_audit.idc`。

## 原始 caller 與 producer 證據

### overlay-09 caller

IDA 連續 dump 在 overlay-local `0430h` 附近保留下列呼叫形狀：

```text
0430h  push [bp+0Ah]
0433h  mov  al,[bp+08h]
0436h  push ax
0437h  9A 2F 00 17 01    far call 0117:002Fh
043Ch  push ax
043Dh  9A C0 00 4A 01    far call 014A:00C0h
0442h  mov  [bp-06h],al
```

`101Fh` 也保留 raw `9A C0 00 4A 01`，其前方以 `[bp+08h]`、`[bp+06h]` 與
`[bp-04h]` 組成呼叫參數；`10B6h`、`19D2h` 則顯示同一 resident far-call
邊界被不同 caller 重用。這只能證明共用 handler 的 caller 形狀與回傳值
消費，不能單靠相同 far pointer 替每個 caller 命名成「近戰選敵」。

### overlay-24 producer

overlay-local `285Bh` 的連續 bytes 顯示：

- 先由 far pointer 讀取來源位置／形狀相關資料，再把 caller range 與
  `FFh` arc 傳入後續 producer；
- 透過 `0184:003Eh` 的 SCAN 邊界，寫入 `DS:9F2Eh` 三 byte 候選表與
  `DS:9F30h` 候選數；
- 再以 `3 * index` 讀取候選記錄，投影到 `DS:A820` 的 combatant index。

這部分的 terrain、footprint、INARC、加權距離與 PC-98 三 byte record 排序，
沿用 READY spec 433、435、436 的逐項證據；本輪沒有重新命名其 raw work
address。

### 外部交叉參考

公開 DOS 重寫來源的 `find_target` 也呈現相同的高層控制流：

- `BuildNearTargets(max_range, player)` 先建立對立目標清單；
- 每輪最多 20 次擲骰；
- 不可見項目從清單移除；
- `clear_target=false` 的第二輪將牆面忽略後重新建立清單。

參考頁面：[ovr014.cs](https://github.com/simeonpilgrim/coab/blob/master/engine/ovr014.cs)、
[ovr025.cs](https://github.com/simeonpilgrim/coab/blob/master/engine/ovr025.cs)、
[ovr032.cs](https://github.com/simeonpilgrim/coab/blob/master/engine/ovr032.cs)。
這些來源只作控制流交叉核對；PC-98 bytes 與本機 runtime 仍是本專案主要證據。

## Remake 接線

| 層 | 實作 | 邊界 |
|---|---|---|
| engine | `combat/scan`、`combat/targetselect` | engine 不知道 CoAB 名稱、文字或地圖；只處理 terrain producer 與有限抽樣 consumer |
| engine schema／pack | `combat_target_rules` | JSON 宣告 `candidate_order=legacy_scan`、`max_range`、`arc`、`retry_attempts`、`retry_with_xray` |
| CoAB JSON | `coab.pc98.enemy-physical-target` | `255 / 255 / 20 / true`，對應本作目前已核對的怪物物理 caller |
| CoAB adapter | `Battle.SelectLegacyScanCombatTarget` | `CHARACTERLIST → LegacyObjectID → footprint → SCAN → stable ID`；第二輪複製 map 後設 `XRay` |
| State | enemy autonomous turn | 正式 pack 且有 `TACTICALMAP` provider 時使用上述 adapter；舊 synthetic 狀態保留相容 fallback |

`targetselect.Select` 不擁有 `Fighter`、文字或 JSON；它只驗證 stable ID、保留
producer 順序，並執行一次 draw／不可見移除。這避免把本作資料與可沿用的
Golden Box engine 混在一起。

## 推論等級與未關閉項目

- `exact`：overlay-local raw far-call bytes、overlay-24 producer 的候選表／數量
  寫入形狀、既有 PC-98 SCAN terrain／footprint／排序，以及 engine 的 20 次
  bounded remove loop。
- `strong inference`：overlay-09 一般敵方 caller 使用 overlay-24 同一候選
  producer；DOS `find_target` 與 PC-98 caller 參數形狀一致，但每個 resident
  handler 的完整 segment resolver 尚未在同一資料流中閉合。
- `hypothesis`：一般物理 consumer 是否完全採用 PC-98 SCAN 的奇偶 tie sorter，
  或仍採另一個 `SortedCombatant` direction comparator；本輪不得把任一未閉合
  comparator 寫成 pixel／行為 exact。
- `unknown`：持久 `action.target` 的跨回合重用、攻擊者移動到近戰距離、flee／
  guard、完整 visibility consumer、每個特殊怪物攻擊的 producer，以及攻擊動畫／
  音效次序。

## 驗證

已驗證：

- engine `combat/targetselect` 單元測試：不可見移除、有限次數、輸入不變與
  越界 draw fail-closed；
- CoAB combat 測試：牆面第一輪無候選、第二輪 XRay 重建、隱形候選移除；
- State regression：game-pack JSON → `TACTICALMAP` provider → SCAN producer →
  一般敵方 target，牆後唯一角色可被第二輪選到；
- `go test`、Docker/Xvfb 正式 gate 與代表性玩家路徑已於本輪完成；正式 gate 以
  `network none`、Xvfb 與隔離快取通過，結果只支持本規格的 bounded milestone。
