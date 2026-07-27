# 299 — Automatic environmental ECL damage

Status: READY

## Evidence

- ECL2 block 4 selector `0x19`, menu index 0 (`ENTER THE BLADES`), first prints
  `THE BLADES TEAR INTO YOU.`
- The handler at `+0x0C2A` tests `4CFE & 0x20`; `+0x0C35` sets that bit before
  the prompt. All three choices therefore consume the event.
- After the following press-button continuation it emits one `DAMAGE` request:
  `flags=0xE0, dice=8d8, bonus=0, saveFlags=0`.
- `0xE0` combines target-present (`0x80`), whole-party (`0x40`), and automatic
  damage (`0x20`). It therefore needs neither WHO selection nor a saving throw.
- The same branch then joins the blade-fade aftermath used by `WAIT`.

## Contract

- State automatically resolves only pending packets whose flags contain all
  of `0x80|0x40|0x20`. Selected-character, saving-throw, and random-hit forms
  remain pending for their existing explicit adapters.
- One deterministic 8d8 packet is rolled and applied to every party member,
  matching the existing Gold Box DAMAGE adapter.
- Persistent roster HP and renderer-facing fighter HP are updated together,
  and the request is consumed exactly once.
- The original ECL pause order, text, menu indexes, and fade continuation are
  unchanged.
- After the second Continue, the script executes `SPRITE OFF`, `CALL 0x2E10`,
  and `EXIT`; revisiting terrain `0x99` exits immediately without replay.

## Verification

- Real ECL2 block 4 regression asserts the exact `0xE0,8,8,0,0` packet.
- State regression uses seed 1, obtains 38 damage, updates two 100-HP members
  and fighters to 62 HP, proves no request remains pending, returns to the
  dungeon after the second Continue, and verifies the event is consumed.
