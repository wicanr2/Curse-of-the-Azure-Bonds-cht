# 245 — reference starting-age generator

Status: READY — 2026-07-27

## Evidence

Reference `ovr018` chooses `race_ages[race][class]` and computes
`base_age + roll_dice(dice_size, dice_count)` for supported single-class
characters. The recovered CoAB table is represented by
`StartingAgeSpecFor`; package class order is mapped explicitly to the reference
columns, including the gaps for unsupported combinations.

## Implemented contract

`RollStartingAge(race, class, seed)` returns a deterministic signed age using
the same dice shape. Unsupported race/class entries return an error rather
than silently assigning age zero. The helper is ready for character-creation
UI integration; current starter templates remain unchanged until that UI
transaction is added.

## Boundary

Multi-class age selection, half-orc records, alignment/class selection and the
full original character-creation screen remain outside the current six-race,
single-class remake model.
