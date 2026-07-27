# 235 — `MON*SPC` invisibility in combat hit resolution

狀態：READY（2026-07-27）

## Reference evidence

`engine/ovr024.cs:CanHitTarget` rolls a d20, calls
`CheckAffectsEffect(target, CheckType.Type_16)`, then compares the adjusted
roll against `target.ac`. `CheckAffectsEffect` dispatches the affect table for
visibility-related combat effects; the existing `party` damage adapter has
verified that `0x19` (`invisibility`) and `0x47` (`invisible`) each subtract 4
from the attack roll. The same adjustment is therefore AC +4 when represented
in the renderer-neutral combat core.

## Contract

- An active enemy `MonsterAffect` kind `0x19` or `0x47` contributes +4 to the
  target AC during `Battle.ResolveAttack`.
- Inactive records contribute nothing.
- The raw record remains unchanged; this slice does not implement visibility,
  dispel, attacker invisibility removal, or other effect kinds.

## Verification boundary

Regression tests must prove the exact hit boundary with and without each
effect kind. The rule is a combat projection, not a claim that every
`MON*SPC` effect has been decoded semantically.
