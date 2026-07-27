# 第三百零八輪：哈普黑暗精靈巡邏

狀態：`READY`

## Reference ENCOUNTER MENU 修正

公開 CoAB reference `ovr003.CMD_EncounterMenu` 證明 opcode `0x29` 的五個
script values 是 encounter behavior modes，不是可直接寫入 destination 的
`ON GOTO` 結果。玩家選 `COMBAT` 時，modes `0/1/3/4` 都寫入結果 `1`；
mode `2` 還要比較雙方 group movement，仍保留明確 engine-context boundary。

本輪修正 bounded VM 的 COMBAT resolver，避免 Hap script 的 mode `0`
被錯寫成 destination `0`、直接跳過戰鬥。WAIT／FLEE／ADVANCE／PARLAY 的
距離、速度與重複 menu 規則尚需依同一 reference 逐項接入，不能把目前
direct mapping 宣稱為完整。

## Hap random patrol

ECL5 block `0x31` SearchLocation 以 `C04F & 0x7F` dispatch。terrain `0x80`
進入 random patrol handler；在 `4BC9 > 14`、`4C5E != 1`、擊敗巡邏數
`4C47 <= 4` 且 random check 成功時，opcode `0x29` 顯示：

- prompt：`A DARK ELF PATROL ARRIVES`；
- options：`COMBAT / WAIT / FLEE / ADVANCE`；
- sprite block `0x31`、picture block `0x31`、max distance `2`。

COMBAT branch 先以 seeded RANDOM 建立 encounter counts，再執行：

- MON5 `0x31` `DK ELF FIGHTER`，icon `0x31`；
- optional MON5 `0x32` `DARK ELF MAGE`，icon `0x32`；
- optional MON5 `0x33` `DARK ELF CLERIC`，icon `0x33`。

本輪 seed `3` 的 regression 固定為三名黑暗精靈戰士與一名法師。勝利後
同一 ECL continuation 將 `4C47` 由 `0` 增至 `1`，勝利確認後返回
`ModeDungeon`，不能落回哈普村外的 wilderness edge menu。

## UI contract

Encounter prompt、威脅台詞與三種怪物名稱均已繁中化。戰鬥小人保持原始
像素，以 nearest-neighbour 整數倍放大至 640×480 畫布；姓名、HP 與敘事
使用 24px 繁中，緊湊數值欄可用約 16×15。
