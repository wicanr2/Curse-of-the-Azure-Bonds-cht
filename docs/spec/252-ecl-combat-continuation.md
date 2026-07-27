# 第二百五十二輪：ECL COMBAT continuation

## 狀態

READY

## Contract

ECL `COMBAT (0x24)` 會將 runtime PC 保存到該 instruction 的 next offset，然後把控制權
交給 `combat.Battle`。當 party 勝利，`game.State` 以同一個 `ecl.BlockSession`／
`RuntimeState` 續跑，保留 ECL memory、compare flags、call stack 與已消耗的 menu sequence。

續跑結果可進入：

- 下一段 localized text 或 horizontal menu；
- `PICTURE`，由 renderer 再顯示事件畫面；
- 下一個完整 MON encounter；
- `NEWECL` 跨 block transition；
- reference `PROGRAM 9` 的 CAMP return。

沒有 ECL session 的 `StartCombat` direct-entry 仍只顯示一般戰鬥結果，避免虛構劇情來源。

## 驗證

- synthetic regression：COMBAT → one-option menu，確認 next-PC 保存與中文 menu 恢復。
- 原始 `curseoftheazurebonds.zip`：ECL1 `JOURNEY ON → PICTURE → COMBAT` real journey
  regression 通過；ECL2 MON encounter battle bridge 也通過。

這只完成戰鬥控制流與已證實的 State bridge；各 ECL block 尚有未支援 engine routine、
完整 encounter outcome side effects 與劇情資料需要繼續反組譯。
