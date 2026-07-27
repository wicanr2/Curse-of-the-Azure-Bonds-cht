# 第二百六十七輪：ECL WHO selection boundary

狀態：`READY`（限 prompt／selection request signal）

## Reference evidence

`ovr003.CMD_Who` consumes one command operand, reads the current ECL prompt text,
clears the normal text area, and calls `selectAPlayer`. It is a character-selection
boundary, not an ordinary ECL menu with an option list encoded in operands.

## Contract

The bounded runner consumes `WHO (0x39)`, preserves the latest decoded ECL text as
`WhoRequest.Prompt`, and continues the instruction cursor. `BlockSession` aggregates
the request across `NEWECL` transitions. A future State UI can present the current
roster and return the selected character through an explicit selection transaction.

## Boundary

This round does not silently choose a character or mutate `SelectedPlayer`; the State
selection transaction, prompt rendering and continuation writeback remain separate.
This avoids confusing WHO's roster selection with HORIZONTAL/VERTICAL MENU selection.

## Verification

The ECL runtime regression covers `WHO → EXIT`, including the no-prompt case and exact
cursor continuation.
