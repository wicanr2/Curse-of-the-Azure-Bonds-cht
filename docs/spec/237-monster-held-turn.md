# 237 — held monster turn and attack resolution

狀態：READY（2026-07-27）

## Reference evidence

`Classes/Player.cs:IsHeld` returns true for `snake_charm (0x33)`,
`paralyze (0x34)`, `sleep (0x35)`, or `helpless (0x1F)`. The monster combat
path uses this state to clear actions, while `ovr014.AttackTarget01` treats a
held target as hit even when the normal `PC_CanHitTarget` check fails.

## Contract

- Active enemy `MON*SPC` kinds `0x1F`, `0x33`, `0x34`, and `0x35` are held.
- A held enemy consumes its turn without physical or spell action.
- A held target is hit by the current combat attack resolver, including a
  natural-one injected roll, matching the reference `... || target.IsHeld()`.
- The raw effect is preserved and not consumed; save/cure/removal rules remain
  separate tasks.
