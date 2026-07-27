# 242 — remake game-time save/load

Status: READY — 2026-07-27

## Contract

The versioned remake JSON save now persists the seven raw game-time slots and
the age-cycle overflow counter. `CurrentGameVersion` is 5. Versions 1–4 remain
accepted; they load with a zero clock because those formats did not contain
time state.

`State.SavePartyFile` writes the live clock through `EncodeGameWithTime`, and
`State.LoadPartyFile` restores it before the adventure resumes. The clock is
separate from DOS `SAVGAM` raw-prefix preservation; importing or writing the
original save format remains its own adapter boundary.

## Verification

`internal/save` round-trip coverage proves all seven slots and age cycles are
preserved, while the existing game and ECL suites protect old save versions
and normal gameplay state.

## Remaining boundary

The DOS `Area1` clock bytes and complete calendar/UI presentation still need a
separate raw-offset validation before being merged into the SAVGAM adapter.
