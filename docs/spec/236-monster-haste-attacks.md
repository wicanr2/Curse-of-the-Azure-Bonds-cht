# 236 — monster `attacksCount` and Haste

狀態：READY（2026-07-27）

## Reference evidence

`engine/ovr017.cs:load_mob` copies `PoolRadPlayer.field_A1` into
`player.attacksCount`; offset `0xA1` is therefore the MON*CHA monster base
attacks-per-turn field. `engine/ovr013.cs:AffectHaste` doubles
`gbl.halfActionsLeft`, and `engine/ovr014.cs:reclac_attacks` starts from
`player.attacksCount` before converting half-actions to attacks.

## Contract

- Parse `MON*CHA[0xA1]` into `Record.AttacksPerTurn` and copy it to the
  fighter.
- Treat a zero synthetic/missing value as one attack in the adapter.
- Each active Haste affect `0x27` doubles that base count.
- Each active Slow affect `0x2A` halves that count, with the current adapter
  retaining at least one attack.
- Movement timing, ranged profiles, and full `reclac_attacks` inputs remain
  separate evidence tasks.

## Verification boundary

Raw-record and encounter tests verify offset decoding, base count, active
Haste doubling, Slow halving, and their composition. This is not a claim that
all effect kinds are implemented.
