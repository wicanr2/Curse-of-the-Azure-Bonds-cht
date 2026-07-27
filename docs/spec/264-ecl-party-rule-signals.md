# 第二百六十四輪：ECL party rule signals

狀態：`READY`

## Reference evidence

CoAB reference `ovr003` defines:

- `PARTYSTRENGTH (0x1D)` with one word destination. It calculates a byte from each
  party member's current HP, AC, hit bonus, cleric level and magic-user level, then
  writes that value to the destination.
- `PARTY SURPRISE (0x22)` with two word destinations. It writes the ranger-party
  flag and the second surprise value to those locations.

## Bounded VM contract

The ECL runner now consumes both commands and continues at the following instruction:

- `RunResult.PartyStrengthRequests` preserves the verified destination address.
- `RunResult.PartySurpriseRequests` preserves both destination addresses.
- `BlockSession` aggregates the requests across `NEWECL` transitions.

The VM does not invent roster values. A later game adapter can resolve the requests
against the current `party.Character`／combat projection and write the result into the
shared ECL memory before the following comparison. This preserves the reusable
VM → works-specific rules boundary used by other Gold Box projects.

## Boundary

This round proves command framing and continuation, not the complete party-stat
projection. The exact primary/multi-class level source, AC scale conversion and
`PARTY SURPRISE` second value must remain in the State adapter until each field is
verified against the imported player model.

## Verification

`internal/ecl/runtime_test.go` runs synthetic `PARTYSTRENGTH → PARTY SURPRISE → EXIT`
and verifies both request payloads and cursor continuation. Core ECL tests remain
the authoritative regression for this slice.
