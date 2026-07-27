# 244 — age-based ability effects

Status: READY — 2026-07-27

## Evidence

Reference `StatValue.AgeEffects` compares age against five race-specific
brackets and adds the corresponding deltas during new-character generation.
CoAB's verified tables are preserved in `Abilities.WithAgeEffects`:

- dwarf: `50,150,250,350,450`
- elf: `175,550,875,1200,1600`
- gnome: `90,300,450,600,750`
- half-elf: `40,100,175,250,325`
- halfling: `33,68,101,144,199`
- human: `20,40,60,90,120`

The reference deltas are applied to Strength, Intelligence, Wisdom, Dexterity,
Constitution and Charisma in that order; exceptional Strength is not changed
by this table.

## Implemented boundary

`Abilities.WithAgeEffects(race, age)` is an explicit, deterministic data
adapter. It is intentionally not called by DOS import or `Character.Fighter`,
because imported DOS records already contain the generated ability values and
implicit reapplication would double-count age.

## Remaining work

Character-creation UI still needs an evidence-backed age input and the full
race/class limit enforcement before this helper becomes an automatic creation
step.
