# 296 — Tilverton Sewers Myth Drannor knight

Status: READY

## Evidence

- GEO2 block 3 terrain `0x83` at `(13,10)` dispatches the slaughtered Fire
  Knife checkpoint and Myth Drannor knight handler.
- The ECL presents three allegiance choices: `FIRE KNIVES`,
  `PRINCESS NACACIA`, and `NO ONE`.
- Choosing Princess Nacacia makes the knight warn the party not to kill the
  hammer-wielding cleric, lets them pass, and consumes this fixed event.
- The walkthrough confirms Princess Nacacia and No One are the friendly
  branches; only the Princess branch is in this milestone.

## Contract

- Multi-pause dialogue remains in one resumable ECL transaction:
  introduction → allegiance question → choice result → dungeon.
- Menu localization changes display labels only. The selected zero-based index
  is sent back to the original ECL menu without replacing its branch logic.
- The event's first-visit/friend state belongs to shared ECL memory. Re-entering
  terrain `0x83` after the final pause must not replay the knight encounter.
- Forgotten Realms proper names use the Traditional Chinese display
  `迷斯卓諾` and `娜卡西亞公主`; raw ECL strings remain unchanged.

## Verification

The real new-game regression continues after winning the five-Fire-Knife sewer
checkpoint, enters `(13,10)`, selects `娜卡西亞公主`, observes the localized
cleric warning, returns to dungeon mode, and verifies that revisiting the same
terrain produces no event.

