# 第二百六十六輪：ECL CHECKPARTY

狀態：`READY`（限已驗證的 skill／movement／affect branches）

## Reference contract

`ovr003.CMD_CheckParty` 讀取六個 operands：query、affect ID，以及四個 word
destinations。query 經 reference `-0x7FFF` normalization 後，目前已確認：

- `0xA5..0xAC`：讀取八組 thief skills，寫入 minimum／maximum／average／false。
- `0x9F`：讀取隊伍 movement，寫入 minimum／maximum／average／false。
- `8001`：檢查隊伍是否有指定 active affect，寫入 `0,0,0,found`。

## Implementation

`ecl.PartyContext` 提供 raw thief skills、movement allowance 與 active effect kinds。
有 context 時，VM 依 reference 寫入 shared memory；沒有 context 時只保存 unresolved
`CheckPartyRequest`。`BlockSession` 會跨 `NEWECL` 聚合 request，State 由目前
`Character`／fighter projection 建立 context。

## Boundary

尚未宣稱完成其他未確認 query selector、完整 NPC／temporary party semantics，或所有
作品共用的 AC／movement scaling。未知 selector 保持 unresolved，不會靜默寫入零值。

## Verification

ECL regression 覆蓋 `0xA5` skill branch 的 min／max／average 與四個 destination；
State／party regression 保護 context projection 與既有流程。
