# 294 — Thieves' Guild exploration and sewer exit

Status: READY

## Evidence

- ECL2 block 2 entry 1 dispatches the guild rooms from local coordinates.
- The post-battle route contains the harp-carrying halfling, Fire Knife kennel,
  monkey cages, guest book, and green slime door text.
- Entry 0 compares work address `0x7ED5`; when non-zero it calls `0xC01E`,
  subtracts ten from map X, and executes `NEWECL 3`.
- The public reference `ovr015.TryStepForward` distinguishes a normal wrapped
  move from an attempted step beyond the 16×16 map boundary.

## Contract

- Temporary encounter allies must be removed from the persistent combat party
  after combat, regardless of their surviving or corpse state.
- The movement adapter reports a passable boundary-crossing attempt to State.
  State supplies `0x7ED5=1` and runs entry 0 → entry 1; transition remains
  ECL-driven.
- Original pixel art uses integer nearest-neighbour scaling on the 640×480
  canvas. Traditional Chinese is rendered independently with a 24px CJK face;
  16×15 remains an allowed compact tier.

## Verification

- A real-image new-game regression wins the guild and kennel battles, observes
  localized room events, then changes from ECL2 block 2 to block 3 through the
  south sewer boundary.
- The Ebiten game package compiles with the boundary-attempt adapter.

