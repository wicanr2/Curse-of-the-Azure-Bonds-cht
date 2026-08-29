# 0.1.0-dev.20260829 發行與推廣片稽核

日期：2026-08-29  
遊戲來源 commit：`d33e1de965fc`  
打包命令：`tools/package-release.sh 0.1.0-dev.20260829`  
錄影命令：`tools/build-promo.sh 0.1.0-dev.20260829`

## 三平台完整版

`dist-all/0.1.0-dev.20260829/SHA256SUMS.json` 共八件：Linux x86_64
AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP，各有 `patch` 與
`full-local`。八件的大小與 SHA-256 已逐件重算相符。

- 每個 ZIP 與啟動目錄都含 `LICENSE` 與 `NOTICE.md`。
- `patch` 拒絕清單通過：不含 `curseoftheazurebonds.zip` 與
  `pc98-bgm-selector-*.ogg`。
- `full-local` 四個平台目錄都含本機原版來源與 OGG，僅供本機驗收，
  不進 Git，不視為可公開散布包。
- Linux full-local AppImage 已在 Docker／Xvfb 實際啟動。
- Windows full-local EXE 已在 Docker／Wine／Xvfb 實際啟動。
- 兩張 smoke PNG 逐位元組相同，SHA-256 均為
  `8e01193085efb3033105c3427764e650a2bc675807fcd822f6dca8778d0a0f5d`。
- macOS 雙架構已交叉編譯與封包，仍未經 macOS 真機、簽署與公證。
  Wine smoke 也不取代 Windows 真機驗收。

## GitHub 公開發行

2026-08-30 已建立
[v0.1.0-dev.20260829 patch 版](https://github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/releases/tag/v0.1.0-dev.20260829)
pre-release。GitHub 讀回驗證為 `isDraft=false`、`isPrerelease=true`，五個附件的
狀態全是 `uploaded`：四平台 patch 與 `PATCH-SHA256SUMS.txt`。GitHub 回報的四個
SHA-256 digest 與本機獨立 patch 清單全數相符。附件不含 `full-local`、
原版 ZIP、PC-98 OGG 與內部推廣片。

## 內部推廣片

檔案：`dist-all/0.1.0-dev.20260829/promo/coab-remake-promo-internal.mp4`  
SHA-256：`26f461c1b93b8e70ff17d08ac988cb0def3b70793aefd3c784514f21f5c02238`

- 1280×720、H.264、30 fps、60.000 秒；AAC、44.1 kHz、雙聲道。
- 平均音量 −21.2 dB、峰值 −12.2 dB；沒有連續三秒以上的低於
  −50 dB 靜音。
- `blackdetect=d=1:pix_th=0.02` 沒有回報長黑畫面。
- 畫面來自同一個封包後 Linux full-local AppImage；提爾佛頓地城
  session 先實際移動，再依序操作 F7／Right／Enter 套用邊框
  A／B／C、以 F6 循環簡中／日文／英文，最後以 F2 從現代主題
  切到原版主題。載入空窗沒有被當成遊戲畫面。
- contact sheet 已目視：實際遊戲、三邊框、三種切換後語文、
  原版主題、F3 攻略、戰鬥與結局均有出現，無跨段殘影。

影片音樂來自 full-local 的 `pc98-bgm-selector-01.ogg`，權利狀態仍是
**rights pending／僅供內部檢視**；不得將此檔當作已獲公開散布授權的成片。
對外版仍需取得音樂授權，或替換為權利已釐清的配樂後重新驗收。
