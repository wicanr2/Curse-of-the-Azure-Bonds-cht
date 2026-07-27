# 300 — Fire Knife frozen room and Journal 26

Status: READY

## Evidence

- ECL2 block 4 SearchLocation selector `0x1A` dispatches to payload
  `+0x0DA2`; GEO2 block 4 stores it as terrain `0x9A`.
- `4CFE & 0x40` is the once-only bit and is set before the prompt.
- Original menu order is `RETREAT / INTERROGATE / KILL`.
- `INTERROGATE` disarms the recovering Fire Knives and records Journal Entry
  26; `KILL` slaughters them before they recover; `RETREAT` exits immediately.
- The supplied Adventurer's Journal and its text transcription identify Entry
  26: an invading cleric paralyzed the men while trying to reach prisoners in
  the leader's room to the south, but was finally overcome in this room.

## Contract

- Raw menu indexes and the `4CFE & 0x40` visited state remain ECL-owned.
- State localizes only display text and labels:
  `撤退 / 審問 / 殺死`.
- Journal Entry 26 is appended only after the interrogation branch and is
  never visible in advance.
- All branches consume the event; revisiting terrain `0x9A` must not replay it.
- The event uses the established 640×480 layout: 24px Chinese narrative text,
  16×15 compact fields where needed, and integer nearest-neighbour source art.

## Verification

The real-image regression asserts all three raw branches, then drives the
playable State through interrogation, checks the Traditional Chinese message,
unlocks the sourced Entry 26 text, returns to the dungeon, and verifies that a
revisit produces no event.
