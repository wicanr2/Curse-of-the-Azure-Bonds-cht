# ECL CLOCK time-step contract

Status: READY — 2026-07-27

## Evidence

- `internal/ecl/operand.go` maps opcode `0x34` to `ECL CLOCK`.
- The public CoAB engine's `CMD_EclClock` calls `vm_LoadCmdSets(2)`, reads
  operand 1 as `timeStep` and operand 2 as `timeSlot`, then calls
  `step_game_time(timeSlot, timeStep)`.
- A linear scan of the shipped `ECL1.DAX` finds real `0x34` instructions with
  the first operand encoded as `0x01 0x06 0x4C`; therefore this command must
  consume two operands even when an older metadata table reports arity 1.

## Implemented boundary

The bounded Go ECL runner now decodes both numeric operands and emits
`RunResult.ClockRequests`. `BlockSession` aggregates the requests across ECL
block transitions. `game.State` applies each request through
`AdvanceGameTime`, so the same raw clock and finite-effect expiration logic is
used by ECL events and REST.

Invalid time slots are returned as errors. The VM does not directly mutate game
state.

## Remaining work

The exact DOS memory-backed operand values and all event-side consumers of the
clock still need broader validation against more ECL blocks. This contract
only covers the command decoding and state adapter.
