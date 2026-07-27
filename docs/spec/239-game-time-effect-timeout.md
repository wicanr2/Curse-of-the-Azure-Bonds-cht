# 239 — game-time slots and affect timeout

狀態：READY（2026-07-27）

## Reference evidence

`engine/ovr021.cs` defines `timeScales = {10, 10, 6, 24, 30, 12, 0x100}`.
`step_game_time(time_slot, amount)` increments the raw seven-slot clock,
normalizes it, and calls `CheckAffectsTimingOut(time_slot, amount)`.
That routine converts the selected slot into elapsed minutes, subtracts finite
effect durations, preserves `minutes == 0`, and treats age/slot-6 overflow as
a separate player-age side effect.

## Contract

- `State.AdvanceGameTime(slot, amount)` preserves raw seven-slot clock state
  and reference slot scaling.
- Party `.FX` effects and active battle raw effects use the same elapsed-minute
  timeout transaction; `Strength == 0xFF` remains permanent.
- Slot-6 overflow is counted by the adapter; player age writeback remains a
  separate record task.
- This does not claim complete rest interruption, safe-location, calendar UI,
  or all original time-triggered ECL behavior.

## Verification boundary

Tests cover slot-2 conversion to ten minutes, clock normalization, finite
effect expiry, and permanent effect preservation.
