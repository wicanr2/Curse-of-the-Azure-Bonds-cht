# 297 — Tilverton Sewers to Fire Knife Hideout

Status: READY

## Evidence

- ECL2 block 3 per-turn entry checks a boundary attempt and the movement
  sentinel at `0x7EC9`.
- The south E2 branch calls `0xC01E`, writes `Y=0`, subtracts two from X, then
  executes `NEWECL 4`.
- ECL2 block 4 initial entry requests `LOAD FILES 4,2,0xFF` and
  `LOAD PIECES 1,2,4`, then prints `YOU ARE ENTERING THE HIDEOUT.`
- The mapped sewer E2 at `(8,15)` enters GEO2 block 4 at `(6,1)` facing south
  after the reference forced-forward movement.

## Contract

- A fresh boundary attempt clears stale forced-move sentinel `0x7EC9`, sets the
  boundary work flag, and supplies current combined GEO registers before entry
  0 runs.
- The source ECL chooses the destination and adjusts coordinates. State must
  not directly select block 4 or invent the target position.
- `NEWECL` runs target initial entry in the same session and aggregates its
  LOAD FILES, LOAD PIECES, text, and map-register writes.
- The target introduction is localized only at State display time; raw ECL
  text and block identity remain unchanged.

## Verification

The formal new-game regression now runs from Tilverton through the carriage,
guild, sewer checkpoint, and Myth Drannor knight, then crosses `(8,15)` and
asserts:

- current ECL block `4`;
- GEO set/block `2/4`;
- map position `(6,1,S)`;
- piece selectors `1,2,4`;
- localized Fire Knife Hideout introduction.

