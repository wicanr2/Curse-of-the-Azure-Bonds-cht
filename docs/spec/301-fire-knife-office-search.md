# 301 — Fire Knife office search, Journal 9, and treasure

Status: READY

## Evidence

- GEO2 block 4 `(14,11)` has terrain `0x9B`, dispatching ECL2 block 4
  selector `0x1B` to payload `+0x0F17`.
- First entry describes the ornate Fire Knife office, pauses, and sets
  `4C10=1`.
- The drawer branch requires `4C10==1` and engine SEARCH flag `7ECA==1`.
  It advances `4C10=2`, sets one-time bit `4CFE|=0x80`, records Journal Entry
  9, and emits `TREASURE(0,0,0,500,500,3,2,0x82)`.
- `0x82` requests two random items. The following `COMBAT` is the original
  treasure-service boundary because it follows `CLEARMONSTERS → TREASURE`;
  it is not a monster encounter.
- The supplied Adventurer's Journal page 12 depicts a flame-outlined body and
  lists: flaming aura, can possess other bodies, and involvement with the Pool
  of Radiance.

## Contract

- Production dungeon input exposes `S` as SEARCH. State sets `7ECA=1` only
  while invoking SearchLocation, then clears it.
- Ordinary movement and ordinary revisits do not discover the drawer.
- Journal Entry 9 unlocks only on successful search and preserves the supplied
  image's three factual notes in Traditional Chinese.
- Treasure resolves as 500 gold, 500 platinum, 3 gems, 2 jewelry, and two
  deterministic random items. The treasure UI returns to `ModeDungeon`.
- `4C10=2` and `4CFE&0x80` prevent repeated text, journal unlock, and loot.
- The 640×480 dungeon prompt includes SEARCH; source art remains integer-scaled
  while Chinese UI text is rendered at the established 16×15/24px tiers.

## Verification

The real-image regression asserts the first visit, ordinary revisit, explicit
search, exact raw treasure request, and consumed state. The playable State
regression presses SEARCH, checks the localized document and sourced Journal
9, obtains the full treasure pool and two random items, exits the loot UI back
to the dungeon, and proves a repeated search gives no additional treasure.
