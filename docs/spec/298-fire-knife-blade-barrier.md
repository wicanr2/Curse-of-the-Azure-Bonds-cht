# 298 — Fire Knife blade barrier

Status: READY

## Evidence

- ECL2 block 4 SearchLocation masks `C04F & 0x3F` and selector `0x19`
  dispatches the blade-barrier room.
- GEO2 block 4 stores that selector as terrain `0x99` at the hideout entrance
  approach.
- The original menu order is `ENTER THE BLADES / WAIT / RETREAT`.
- Selecting `WAIT` prints that the blades slow down and fade away without
  emitting an ECL `DAMAGE` request. Entering uses the script's separate damage
  branch and remains available at the same menu index.

## Contract

- Raw selector, menu order, and ECL control flow remain authoritative.
- Traditional Chinese is applied only at the State display boundary:
  `闖入刀刃 / 等待 / 撤退`.
- The 640×480 event layout renders narrative text at 24px; original dungeon
  art is enlarged only by integer nearest-neighbour scaling. Compact status
  fields may use the established 16×15 CJK font contract.
- Choosing `WAIT` must produce no damage and must reach the original
  blade-fade aftermath.

## Verification

The real-image regression loads ECL2 block 4, supplies terrain `0x99`, asserts
the three-option prompt, runs menu index 1, verifies the no-damage aftermath,
and checks the Traditional Chinese prompt, aftermath, and labels.
