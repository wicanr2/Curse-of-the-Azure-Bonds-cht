# 240 — REST advances reference game time

狀態：READY（2026-07-27）

## Reference evidence

`engine/ovr021.cs` advances resting time with `step_game_time(1, 5)` in the
rest loop. Slot 1 is the minute-sized unit used by the combat round and the
effect timeout conversion. The remake's REST menu uses hours, so one
requested hour maps to 60 slot-1 units.

## Contract

- `REST_START` advances the State game clock before applying natural healing.
- Each requested hour advances 60 slot-1 minutes via the shared adapter.
- Effect timeout happens before the rest healing result is reported.
- Existing bounded natural healing remains one HP per uninterrupted 24 hours.
- Random interruption, safe-location checks, spell-learning side effects, and
  full rest UI remain outside this slice.
