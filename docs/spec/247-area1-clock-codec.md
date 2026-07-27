# 247 — Area1 game-time codec

Status: READY — 2026-07-27

## Evidence

Reference `Area1.field_6A00_Get/Set` exposes seven clock words at offsets
`0x18C, 0x18E, 0x190, 0x192, 0x194, 0x196, 0x198`; `ovr021.step_game_time`
reads and writes those words in order. They correspond to the seven raw slots,
not merely the human-readable hour/minute fields.

## Implemented contract

`area.State.GameTime` now decodes and encodes all seven words while preserving
the rest of the 0x800-byte Area1 record. `State.SetAreaState`, SAVGAM prefix
load, `AdvanceGameTime`, and remake JSON save keep the State clock synchronized
with the Area1 representation.

## Boundary

The renderer still does not present a full calendar/time UI, and unknown Area1
fields remain raw-preserved rather than interpreted.
