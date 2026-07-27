# 234 — `MON*SPC` monster effects attachment

狀態：READY（2026-07-27）

## Reference evidence

`engine/ovr017.cs:load_mob` loads `MON<area>CHA` for a monster ID, then loads
`MON<area>SPC` with the same ID and appends each nine-byte `Affect` record to
the monster's affect list. The `SPC` member is therefore keyed by the same
chapter-local monster ID as `CHA`; it is not a global effect table.

## Contract

- Decode every `MON*SPC.DAX` block with the existing nine-byte
  `monster.ParseAffects` parser.
- Preserve the block ID → affect-list association while loading each chapter.
- When an ECL encounter creates a fighter, attach a copy of that monster's
  raw effects to the renderer-neutral combat fighter.
- Keep `BuildEnemies` as a no-effects compatibility adapter.
- This slice does not infer gameplay semantics from effect kinds. Applying
  invisibility, haste, sleep, and other effects remains a separate evidence
  task.

## Verification boundary

Synthetic DAX/effect tables must verify block association, copy isolation, and
chapter selection. No claim is made here that a raw effect is currently active
in combat rules.
