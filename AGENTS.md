# Codex working agreement

## Architecture decision

The project is being split into two repositories:

- `golden-box-remake-engine`: reusable Go engine, Ebiten frontend boundaries,
  ECL/DAX/GEO codecs, combat/rules primitives, JSON schemas and declarative
  event runtime.
- `Curse-of-the-Azure-Bonds-cht`: CoAB game pack, original-game evidence,
  Traditional Chinese localization, converted assets, walkthrough,
  screenshots and integration tests.

Do not add new CoAB plot facts directly to reusable Go engine code. In
particular, plot flags, ECL/GEO block mappings, coordinates, encounter
compositions, NPC departures, English-text matching and translated narrative
belong in versioned JSON game-pack files.

Go code may contain only reusable mechanics, typed adapters and evidence-based
format semantics. If a CoAB behavior cannot yet be represented declaratively,
extend the engine schema/runtime first, then describe the behavior in the CoAB
game pack.

The remake must include both original map presentations, not only narrative
and combat screens: the first-person city/dungeon viewport assembled from
GEO/WALLDEF/8X8D data, and the wilderness/world travel map. CoAB JSON selects
resources and title rules; projection, camera, occlusion, walls, doors and
integer-scaled rendering belong in the reusable engine.

## Current migration

The first proof case is the Pit of Moander departure:

- trigger: ECL destination block `0x51`, memory `4C5B=FF`, `7F12=1`;
- remove NPCs by script names `ALIAS` and `DRAGONBAIT`;
- choose living/dead Alias farewell localization;
- return to the pending world ECL menu.

`applyPitOfMoanderDeparture` has been removed. The generic
`applyDataPackEvent` adapter now projects declared engine inputs/outputs;
remaining title-specific branches in the older `localizeECLText` are migration
debt and must move to JSON incrementally.

The current continuation reaches Zhentil Keep from the Pit through the real ECL
session. Its newly added patrol, gate, questioning, warning, inner-city and
Journal 32 text is JSON-backed.

## Git discipline

- Make one commit and push only after a meaningful, tested milestone.
- Keep the two repositories' histories independent.
- Never commit the nested engine repository as a CoAB gitlink or copied source.
- CoAB pushes use:
  `git --git-dir=/tmp/azure-bonds-git --work-tree=.`
- Do not discard unrelated user changes.

## Validation

Run focused Go tests during development. Before pushing a milestone, run
`go test ./...` under Docker/Xvfb for every affected repository and
`git diff --check`.

## Current pause and reverse-engineering tool

As of 2026-07-28 the user requested a pause after the DOS adventure-layout
milestone. Do not resume new reverse engineering or feature work until asked.
The authoritative outcome inventory is `docs/project-status.md`.

When work resumes, an IDA Pro installation is available at:

`/home/anr2/ida_94_official/dist`

Keep the existing SDD boundary: record IDA evidence in `docs/spec/`, mark the
spec READY, then implement. Do not treat decompiler output as proof without
cross-checking original bytes, runtime behavior, or another authoritative
source.
