# 295 — Tilverton Sewers entry and Fire Knife checkpoint

Status: READY

## Evidence

- ECL2 block 3 initial entry is `+0x0014`; it requests `LOAD FILES 3,2,0xFF`
  and `LOAD PIECES 1,2,4`, then describes the slippery, low sewer.
- The block 2 south exit operates in combined GEO coordinates. After
  `CALL 0xC01E`, `X -= 10`, and `NEWECL 3`, the party enters GEO2 block 3
  at `(0,1)` facing south.
- SearchLocation masks `C04F & 0x3F`. Terrain `0x81` at `(1,8)` and `0x82`
  at `(5,5)` share the Fire Knife checkpoint handler.
- The refusal branch loads five monster ID `1` records from MON2CHA and enters
  COMBAT. Victory prints `YOU QUICKLY HIDE THEIR BODIES.`

## Contract

- A boundary lifecycle syncs combined geometry into ECL before the source
  script runs and reads target map registers back after `NEWECL`.
- Target initial-entry `LOAD FILES`, map position, direction, text, and
  continuation are applied as one resumable session; no fresh runtime is made.
- Fixed location dispatch remains terrain-selector driven, not hard-coded to a
  coordinate in State.
- Refusing surrender starts the normal Battle and resumes the same ECL after
  victory. The checkpoint's first-visit state remains in shared VM memory.

## Verification

The real new-game integration path runs:

```text
Tilverton → carriage → jail → guild → guild battle → sewer boundary
→ localized sewer introduction → terrain 0x81 checkpoint
→ refuse surrender → five Fire Knives → localized victory continuation
```

The `-sewers` preview flag performs the same story bootstrap and stops at the
checkpoint choice for visual inspection.

