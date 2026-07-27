# 243 — DOS player age writeback

Status: READY — 2026-07-27

## Evidence

- Reference `Player.age` is a signed 16-bit field at `0x76` in the normal
  player record.
- Reference `PoolRadPlayer.age` is the corresponding signed 16-bit field at
  `0x30` in the pool/rad import record.
- `NormalizeClock` increments every party player's age when slot 6 overflows.

## Implemented contract

The CoAB `.SAV/.GUY` parser now reads `age` from `0x76`, `Character` preserves
it, and `PatchDOSPlayerRecord` writes it back without touching unknown bytes.
When the shared game clock carries past slot 6, every loaded party character
gains one year; the remake `gameAgeCycles` counter remains an audit-friendly
overflow count. Signed age saturation prevents an invalid int16 wrap.

## Boundary

The separate Pool/Rad source record and age-dependent ability modifiers are
not guessed from the normal player record. SAVGAM sidecar staging can now carry
the verified normal-player age field; the full original age UI and all
multi-class character serialization remain separate work.
