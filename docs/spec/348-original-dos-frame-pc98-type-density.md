# 348 — Original DOS frame and PC-9801 type density

Status: READY

## Evidence

- `docs/reference/original-dos/tilverton-first-person-demo.png` is a native
  320×200 DOS runtime capture. Its fixed chrome occupies the outer eight
  pixels, the `x=128..135` divider, `y=128..135` narrative divider, and
  `y=184..191` command divider. These pixels are the implementation oracle;
  the former procedural grey rectangles are not.
- MobyGames' PC-98 gallery identifies 25 native 640×400 captures, including
  adventure/status (`464680`) and combat (`464695`). The Japanese port uses
  compact bitmap text throughout the roster, narrative, combat status, and
  command regions. It is a typography-density oracle, not the DOS frame
  material oracle.
- The supplied ETen `STDFONT.15` is 16×15. The existing renderer already
  supports a one-pixel horizontal embolden pass.

## Required implementation

1. Extract the fixed DOS chrome pixels from the local runtime capture into a
   transparent native 320×200 raster. Do not redraw cracks or bevels by eye.
2. Scale that raster exactly 2× with nearest-neighbour filtering on the
   640×480 canvas. The extra 80 pixels remain below the native 400-pixel
   image and may carry translated narrative/commands.
   Until a local DOS combat capture is available, compose combat's measured
   `x=176` divider geometry from these oracle-sampled stone strips. Label it
   material-exact/layout-reconstructed; do not call it an exact combat raster.
3. 一般 PIC／第一人稱的原始可見內容是 88×88，不是整個左上 stone panel。
   依第 406 輪 raw runtime 量測，以 2× nearest-neighbour 放在 remake
   `(48,48)..(223,223)` 並裁切；其周圍灰色 stage 是原始框線，不能為了「填滿」
   而改成非整數縮放。HEAD／BODY 人物另依專用舞台契約處理。過去
   `(16,16,240,240)` 的泛稱已被此精確量測取代。
4. When `-eten-font` is supplied, use the bold 16×15 face for ordinary
   narrative, roster, status, and command text. Reserve the 24px face for
   headings and explicit emphasis screens.
5. Keep an original-faithful theme as the baseline. Any future beautified
   frame must be a separate selectable theme and must not overwrite this
   oracle-backed presentation.

## Verification

- Pixel regression checks retained/transparent points in the extracted frame.
- A deterministic adventure screenshot demonstrates the original cracked
  border and a filled upper-left viewport.
- A deterministic combat screenshot demonstrates compact CJK status/log text.
- `go test ./...` and the Docker/Xvfb smoke path pass before the milestone
  commit.

The reusable design rationale is maintained in
`docs/knowledge/golden-box-remake-for-chinese-readers.md`.
