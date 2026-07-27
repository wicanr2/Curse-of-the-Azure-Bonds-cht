# 第二百六十五輪：ECL party context writeback

狀態：`READY`（限 `PARTYSTRENGTH`／`PARTY SURPRISE` 的已驗證欄位）

## Contract

`ecl.PartyContext` 將 State 目前 roster／fighter projection 映射為 reference party
command 所需的最小資料：HP、AC、attack bonus、cleric level、magic-user level 與
ranger presence。

- `PARTYSTRENGTH` 依 `ovr003.CMD_PartyStrength` 的公式計算 byte，寫入 destination。
- `PARTY SURPRISE` 寫入 ranger flag；reference 第二個 destination 的目前已證實值為
  zero，也會寫入 shared ECL memory。
- `BlockSession.RunInteractiveSeedWithPartyContext` 讓 memory writeback 在 menu pause／
  `NEWECL` continuation 間保持一致。
- 沒有注入 context 的純 ECL runner 只輸出 unresolved request，避免虛構 party state。

## Multi-class boundary

cleric／magic-user／ranger level 由 `Character.ClassLevels` 優先，legacy character 才
使用 primary `Level` fallback。AC 的 reference internal scale 與完整 multi-class
THAC0／level table 尚未宣稱完成；目前 adapter 只使用已投影的 fighter values，方便後續
逐欄替換。

## Verification

ECL runtime synthetic tests 覆蓋 unresolved request、context-resolved values、shared
memory continuation；game tests 仍驗證真實 State 選擇流程不破壞既有 opening／combat。
