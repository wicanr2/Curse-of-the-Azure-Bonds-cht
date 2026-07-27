# 302 — Fire Knife ashen-room exploration

Status: READY

## Evidence

ECL2 block 4 SearchLocation maps five one-time selectors:

| Terrain | Handler | Flag | Event |
|---|---:|---:|---|
| `0x9C` | `+0x101A` | `4C11` | hallway with a strange smoky scent |
| `0x9D` | `+0x105F` | `4C12` | unnaturally ordered bedroom and unseen servants |
| `0x9E` | `+0x111C` | `4C13` | burned library, charred body, paper, Journal 29 |
| `0x9F` | `+0x120D` | `4C14` | laboratory destroyed by the same intense flame |
| `0xA0` | `+0x1279` | `4C15` | shrouded bodies marked “to be raised/buried” |

Each handler compares its flag with 1, exits when already visited, otherwise
saves 1 before displaying text. The library has two Continue pauses; the other
rooms have one.

The supplied Adventurer's Journal Entry 29 says that the Fire Knives' ally can
control flame, move from body to body, and use extra-dimensional powers; the
writer concludes that the Flamed One is “Tyran...” (Tyranthraxus).

## Contract

- Raw terrain dispatch, pause count, and visited bytes remain ECL-owned.
- State localizes all five events at display time.
- Journal Entry 29 is appended only after taking the protected paper from the
  charred hand; it is not visible on the first library pause or beforehand.
- Returning to any of the five selectors after completion produces no menu or
  repeated journal entry.
- Events use the established 640×480/24px narrative layout and integer-scaled
  source pixel art.

## Verification

A table-driven real-image regression executes all five selectors, verifies
their source text and full continuation count, then drives playable State
instances through each localized event. It asserts flags `4C11..4C15`, dungeon
return, consumed revisits, and the sourced Traditional Chinese Journal 29.
