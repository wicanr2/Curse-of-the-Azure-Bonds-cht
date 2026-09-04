# Curse of the Azure Bonds — Remake

[繁體中文 README](README.md)

A data-driven remake of SSI's 1989 Gold Box game *Curse of the Azure Bonds*,
written in Go with Ebiten. It reads your own legally obtained copy of the DOS
game and plays it end to end. The interface ships in Traditional Chinese,
Simplified Chinese, Japanese and English, switchable in game with F6.

## The story

On the road to Tilverton your party is ambushed. Before anyone works out what
is happening, people are already down.

You wake in a house in Tilverton with your equipment gone and five azure sigils
burned under your right arm. The mouthed palm is Moander. The Z in a circle is
the Zhentarim. Three bars across a half-moon belong to the Red Wizard
Dracandros, the burning claw to Tyranthraxus the Flamed One, and the fifth mark
comes from the Fire Knives. These five trust each other no further than they
have to; the alliance exists to pool power. The bonds glow, and they take you
over without your knowing it.

The sage Dimswart finds the one seam worth prying at: each of the five wants to
spend you on its own move first, even at an ally's expense. So you work down
through Tilverton's sewers and out to Zhentil Keep, Haptooth and Yulash, take
back the Symbol of Lathander, the dragon helm and Moander's gauntlet, and
finally turn to face Tyranthraxus, who holds the Pool of Radiance. Five bonds,
broken one at a time.

The story comes from the 1988 novel *Azure Bonds* by Kate Novak and Jeff Grubb.
Alias, the swordswoman who carries the same marks, meets you in game with her
companion Dragonbait; she recognises the sigils because she has worn them.

## What is in this repository

This repo holds the CoAB game pack, the translations, the original-asset
conversion, the guides and the integration tests. The reusable
ECL / DAX / GEO / combat / save-game runtime lives in a separate repository,
[golden-box-remake-engine](https://github.com/wicanr2/golden-box-remake-engine).
Plot, coordinates, items and translated text are data in this repo, never
hard-coded in the shared engine.

## Status

The main quest runs from the opening to the ending in one session. Driven only
through real front-end key presses, `GameWon()` fires on frame **14,380**,
covering 899 map cells, 16 ECL blocks and 393 lines of dialogue with zero
fallbacks to untranslated text. That proves the input layer can reach the
ending; it is not a substitute for a human playthrough at ordinary party
strength.

| | |
|---|---|
| Disassembly coverage | 2,874 functions all in the ledger, **0 left to interpret**; 36,386 undefined bytes classified segment by segment |
| Specifications | 1,229 documents under `docs/spec/`, each marked `READY` or `DRAFT` with its evidence level |
| Code | 498 Go files, 142,653 lines, excluding the engine repo |
| ECL text | 1,022 reachable pages, **0 unwired** |
| ECL side effects | 14,177 reachable instructions, all `done`, none `partial` |
| Combat spells | **73 / 73** have a handler, a visual and a sound |
| First-person view | 585 wall configurations, **1,498 frames compared cell by cell against the original — all identical** |
| Save data | 422-byte character record: 299 decoded, 123 documented, **0 unknown** |
| Releases | `v1.0.4-20260904` builds for Linux x86_64 (AppImage), Windows x86_64 and macOS x86_64 / arm64 |

Windows and macOS builds have not been started on real hardware, and the macOS
build is neither signed nor notarised. Music and sound ship as 12 OGG tracks
and 9 effects; per-track listening checks and the distribution-rights list are
still open. These are disclosed rather than hidden, and by the maintainer's
2026-08-30 decision they no longer block the first release.

## Screenshots

![Opening prologue: the campfire ambush, three NPCs joining, and the origin of the five azure sigils](docs/screenshots/opening-prologue-remake.png)

![Tavern scene: DOS stone frame, HEAD/BODY stage and narration](docs/screenshots/gold-box-layout-adventure.png)

![Tilverton first person: plank wall, stone wall and side walls converging cell by cell inside the original 88×88 viewport](docs/screenshots/tilverton-first-person-remake.png)

![Combat: opening formation placed by the original deployment algorithm](docs/screenshots/gold-box-layout-combat.png)

![F3 full guide map: original 11×11 AREA symbols, party facing and event markers](docs/screenshots/a6-guide-overlay.png)

More in [`docs/screenshots/`](docs/screenshots/). The faithful theme and the
modern A6 theme are maintained separately; F2 switches between them live.

## Playing it

You need your own legal copy of the DOS game. Pack it as
`curseoftheazurebonds.zip`, keep the original filenames inside, and put it next
to the executable or AppImage. Release packages never contain the original ZIP,
and the public patch packages also leave out the PC-98 music.

Downloads are on the
[releases page](https://github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/releases).
Each release lists SHA-256 digests for every file.

English text follows the original DOS wording: the story text is the original
game's, and items, spells and proper nouns use the original terms rather than
back-translations from Chinese. UI strings the original never had are machine
translated, with no claim of human editing.

## Building from a clean clone

The engine is a separate repository pinned by `go.mod`. Nothing about it is
vendored here, so fetch it first:

```sh
git clone https://github.com/wicanr2/Curse-of-the-Azure-Bonds-cht.git
cd Curse-of-the-Azure-Bonds-cht
tools/engine-bootstrap.sh      # fetches the engine commit pinned in go.mod
tools/go.sh test ./...         # full test suite; the Go toolchain runs in Docker
```

`tools/engine-bootstrap.sh` only ever checks out the pinned commit and never
touches the engine's working tree. Builds go through Docker, so no local Go
installation is needed.

## Documentation

Most documents are in Traditional Chinese.

- [`docs/spec/README.md`](docs/spec/README.md) — index of the format and
  behaviour specifications, each with its evidence boundary.
- [`docs/guide/README.md`](docs/guide/README.md) — player guide.
- [`docs/manual/curse-of-the-azure-bonds-zh-TW.md`](docs/manual/curse-of-the-azure-bonds-zh-TW.md)
  — keys, combat, camping and the journal.
- [`AGENTS.md`](AGENTS.md) — the working rules: definition of done, evidence
  standards, git discipline and verification gates.
- [`HANDOFF.md`](HANDOFF.md) — shortest current state, next step, and the gates
  that must not be reopened.

## License

Code, translations, documentation and original content owned by the copyright
holder are licensed under [RRSAL-1.0](LICENSE) (Retro Remake Source-Available
License 1.0). Non-commercial use, modification and redistribution are free.
Streaming, recording, video and commentary — including platform ad revenue,
donations and memberships — are permitted explicitly by Section 4. Commercial
use is reserved; open an issue to discuss it.

This is **source-available**, not open source: the non-commercial restriction
does not meet the OSI definition.

The license does not cover the original game's executables, data, artwork,
music or text, which belong to SSI and its successors, nor the Noto Sans CJK
fonts (SIL OFL 1.1), the separate engine, or third-party dependencies. The full
boundary is in [`NOTICE.md`](NOTICE.md).

---

Rights in the original game, manuals, images and music remain with their
respective holders. This project preserves research evidence and newly written
material; it does not redistribute the original game.
