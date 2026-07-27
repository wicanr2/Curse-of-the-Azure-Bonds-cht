# 246 — character-creation age transaction

Status: READY — 2026-07-27

## Contract

`State.AddCreationCharacter` now mirrors the verified ordering boundary:

1. keep the editable starter template unchanged;
2. generate a race/class starting age with `RollStartingAge`;
3. apply `Abilities.WithAgeEffects` to the copied character;
4. assign the party ID and append the result to the creation roster.

The seed is deterministic for the template index and roster position, making
tests and replayable prototype sessions stable. Reusing a template therefore
does not accumulate age modifiers.

## Boundary

The UI still presents the existing three starter templates rather than the
original full race/class menu. Multi-class, half-orc, alignment and full
class-minimum generation remain separate work; this contract only closes age
generation for the current six-race single-class remake slice.
