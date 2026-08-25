# 第五百四十九輪：DOS 角色建立單一面板與 README 截圖校正

狀態：`READY`（目前 renderer／代表畫面；不代表所有 UI 或完整戰鬥 fidelity）
日期：2026-08-12

## 問題與原版證據

README 舊圖暴露了實際 renderer 問題，而不只是截圖過期：

- 冒險訊息與命令文字壓到 DOS stone divider；
- 角色建立誤用冒險的「左場景＋右 roster」分欄；
- 缺少 `SPCFONT.15` 時，全形標點退回通用 fallback；
- 戰鬥 footer 的兩行過於貼近 640×480 下緣。

本機 DOS 原版
[`character-age-create.png`](../reference/original-dos/character-age-create.png)
SHA-256 為
`1214eeb3d82df2072e7a5770a697d2c531e59215a9169f7ea72dd11dc39d7bd0`。
它證明角色建立是單一大面板，不帶冒險畫面的 `x=128..135` roster divider。
第一人稱／PIC 的原始 88×88 可見區仍以
[第 406 輪規格](406-dos-gui-draw-contract.md)為準：DOS `(24,24)..(111,111)`，
remake `(48,48)..(223,223)`，嚴格 2× nearest-neighbour；周圍灰色 stage 是原始
框線邊界，不是應以非整數縮放填滿的空白。

## 實作

- 新增 `dos-character-creation-frame.png`，由
  [`scripts/extract_dos_character_creation_frame/`](../../scripts/extract_dos_character_creation_frame/main.go)
  從本機原版捕捉抽出；它只保留固定 stone chrome，排除角色數值與 prompt。
  資產 SHA-256：
  `8618b170ac1c7afad78a127f3389554d439634abac0730f20729447f4fa4d5fe`。
- `ExtendedCharacterCreationFrame` 保留原版單一上面板；640×480 額外的下方中文
  資訊區沿用已明確標示 `layout-reconstructed` 的延伸石框。
- `ASCFONT.15` 會在 `-eten-font` 的同目錄自動載入。ASCII 與常見全形標點改用
  倚天原生 8×15 raster；沒有 companion 檔時才保留原有 fallback。
- 冒險 frame 在文字之前合成；訊息留出 divider 後的中文 leading，命令列落在
  原版延伸 command interior。戰鬥 footer 改為彼此分離且留有底緣。
- 一般 PIC 仍以原始 88×88、2×及 clip 繪製；新增 sub-image clip 時保留全域
  transform，避免非零 origin 把右／下像素裁掉。

`-character-creation` 是 Docker/Xvfb 的 deterministic renderer checkpoint；它呼叫
與 party-less opening 的 `START` 相同 `OpenCharacterCreation` state command，但不
應取代第 531 輪「實際 C 鍵」的正常輸入證據。

## README 畫面證據

所有圖片均由 Docker/Xvfb、`-eten-font /eten/stdfont.15` 產生，尺寸固定
640×480：

| 圖片 | checkpoint | SHA-256 |
|---|---|---|
| `gold-box-layout-adventure.png` | `-inn` | `285191cddba9e61de6280575daed5a14f6adcf6ad06b3cce67bdda474cf21215` |
| `character-creation-remake-640.png` | `-character-creation` | `0251c1ff34c8c2cd25042151fdce577dbfb5cb046acc26912be2c8df37018c8e` |
| `gold-box-layout-combat.png` | `-encounter` | `3bee4cba0e1c138601b2b54fc6e61933a3faf1272afee69fd128ff87a650578d` |
| `tilverton-first-person-remake.png` | `-tilverton-dungeon` | `2fc0ef9cf842701b9922f9241c94c75bb6dbf70dfe3cce98298d6aab9e97f458` |
| `burial-glen-red-web-spiders.png` | `-burial-red-web-battle` | `3e306c4743676874c8e317c642f420bd13b548c17e808078b25ad0d64a8e8e8d` |

雜湊由 [第 1130 輪](1130-screenshot-layer-alignment.md)與
[第 1131 輪](1131-wall-symbol-group-zero.md)、[第 1132 輪](1132-combat-floor-sync-and-formation-band.md)
重拍：1130 修掉 sub-image 座標、戰鬥地形與戰鬥員的座標路徑、佈陣的地面檢查與
戰場圖示來源；1131 補上第一人稱牆面的第 0 段符號與洋紅透明鍵；1132 讓佈陣
與畫面用同一份戰鬥地圖，並把編制帶移到視野中段。兩張戰鬥截圖第 713 輪
再重拍：佈陣改走 spec 1200 的原作 COMPREP 佈署演算法（原生戰鬥地圖座標），
遠距離遭遇開場時敵隊在視野外、逐回合逼近，與原作行為一致。

這組畫面可證明目前文字留在各自的可用區、角色建立沒有錯誤分欄、原版 stone
chrome 與 16×15 倚天基線已接通。它們是 `material-exact/layout-reconstructed` 的
混合證據：原始 raster／88×88 stage 是 `exact`，640×480 中文延伸與戰鬥整張
layout 是 `layout-reconstructed`。它們不能證明完整戰鬥 AI、地形、法術動畫、音效、
所有 UI state 或整作通關。

## 驗證

```text
go test ./internal/etenfont ./internal/gfx ./cmd/azure-bonds-game
```

另以 Docker/Xvfb 重拍上表五個 checkpoint 並人工檢視。`go test ./...` 仍受既有
`scripts/` 中多個獨立 `main` 檔案影響，故本輪依受影響套件分開執行，不把它誤報
為全套 gate。
