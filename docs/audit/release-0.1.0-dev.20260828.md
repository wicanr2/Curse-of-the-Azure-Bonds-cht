# 0.1.0-dev.20260828 發行與推廣片稽核

日期：2026-08-28  
來源 commit：`9e490c86`  
打包命令：`tools/package-release.sh 0.1.0-dev.20260828`  
錄影命令：`tools/build-promo.sh 0.1.0-dev.20260828`

## 三平台封包

`dist/0.1.0-dev.20260828/SHA256SUMS.json` 是完整機器可讀清單，共八件：Linux
x86_64 AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP，各有 `patch` 與
`full-local`。

- Linux full-local AppImage 已在 Docker／Xvfb 實際啟動並擷取繁中開場。
- Windows full-local EXE 已在 Docker／Wine／Xvfb 實際啟動並擷取繁中開場。
- 兩張 smoke PNG 的 SHA-256 都是
  `ae391830198e7fbe7bf6e768e061217bc989c87ce46ccf8eb718836280c5d093`。
- macOS 兩架構完成交叉編譯與封包，**尚未在 macOS 真機啟動**。
- 四平台 patch 目錄都包含逐位元組相同的 `LICENSE` 與 `NOTICE.md`。

## 內部展示版推廣片

檔案：`dist/0.1.0-dev.20260828/promo/coab-remake-promo-internal.mp4`  
SHA-256：`e2e5ec0f790022649c0fa9b03e5cef506aba5189ded3aca9a6d5f3f24d68ed25`

- 1280×720、H.264、30 fps、1,140 幀；視訊與封裝時長都是 38.000 秒。
- AAC、44.1 kHz、雙聲道；音訊時長 38.000 秒。
- 平均音量 −22.9 dB、峰值 −15.0 dB；沒有連續三秒以上低於 −50 dB 的靜音。
- 七張定點抽樣幀及八格 contact sheet 已逐張檢視：標題、開場、隊伍建立、F3 攻略、
  戰鬥、結局與授權尾卡依序出現，沒有跨段殘影或字幕錯配。

影片使用 full-local 的 `pc98-bgm-selector-01.ogg`，權利狀態仍是
**rights pending／只供內部檢視**；不得上傳或當成已取得公開散布權的宣傳成片。
公開版必須先取得該音樂的明確授權，或換成另有書面公開／商用權的配樂後重新驗收。
技術解碼、音量與頻譜檢查不能取代最後的人耳聆聽。
