# 第二百六十八輪：WHO roster transaction

狀態：`READY`（限 ECL pause／選角／resume）

## Contract

`WHO (0x39)` 在 interactive runner 沒有 selection 時會停在目前 opcode，保存
`RuntimeState` 與 `WhoRequest.Prompt`；State 以目前 party roster 建立繁中角色選單。
選擇 index 會透過同一個 `BlockSession` 的 `whoSelectionOffset` 回送，runtime 消費
後從原位置繼續，不會重播已完成的 ECL commands。

State 另外保存 selected player ID，供後續 `LOAD CHARACTER`／party-specific routine
adapter 使用；沒有 roster 的 headless VM 仍可選擇 non-interactive signal mode。

## Boundary

這輪只完成 WHO 的角色選擇 transaction。原版 selected-player 對所有 DOS global
routine 的完整 side effects、NPC／temporary party semantics 與跨作品 UI layout 仍待
逐一接入；不會自動把 WHO 當成普通 horizontal menu。

## Verification

`internal/game/state_test.go` 覆蓋 synthetic `WHO → EXIT`：第一次選擇會停在角色選單，
第二次選擇角色乙後，ECL session 恢復並保存 selected player `b`。
