# 1.0.0-20260830 首個正式版發行稽核

日期：2026-08-30  
遊戲來源 commit：`1a57438d`  
打包命令：`tools/package-release.sh 1.0.0-20260830`

## 發行定位

本版是第一個正式版；玩家可見體驗目標為近似原版 99%，
不宣稱未有原版 oracle 的細節是逐位元或逐週期 exact parity。
發行 commit 已修正現代主題在純文字事件錯疊 scene 分割框、
導致開場文字被邊框覆蓋的問題。

## 三平台封包

`dist-all/1.0.0-20260830/SHA256SUMS.json` 共八件：Linux x86_64
AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP，各有 `patch`
與 `full-local`。清單記錄每件的 byte 數與 SHA-256。

- 八個封包均已產生；ZIP 的 `LICENSE` 與 `NOTICE.md` 已抽查。
- 公開 `patch` ZIP 不含 `curseoftheazurebonds.zip` 與
  `pc98-bgm-selector-*.ogg`。
- `full-local` ZIP 含本機原版來源，僅供本機驗收，不進 Git、
  不上傳 GitHub Release。
- Linux full-local AppImage 已在 Docker／Xvfb 啟動；Windows
  full-local EXE 已在 Docker／Wine／Xvfb 啟動。
- 兩張開場 smoke PNG 逐位元組相同，SHA-256 均為
  `d0e5d6399273d59dea3f43f58f245c61c19ed3adf0b3d0e070989cb90d234362`；
  人工目視確認純文字開場只留外框，無內框或分隔線覆字。
- macOS 雙架構已交叉編譯與封包，尚未經 macOS 真機、
  簽署與公證；Wine smoke 也不取代 Windows 真機驗收。

## 公開發行邊界

GitHub Release 標籤為 `v1.0.0-20260830`，只公開四件 `patch`
與 patch SHA-256 清單。授權為 PolyForm Noncommercial 1.0.0；
允許非商業使用、修改與散布衍生版本，必須保留署名與同一
非商業限制，商業使用另洽。

2026-08-30 發布後已從 GitHub 讀回驗證：`isDraft=false`、
`isPrerelease=false`，四件平台封包與 `PATCH-SHA256SUMS.txt` 共五個
附件的狀態均為 `uploaded`；GitHub 回報的四個封包 digest 與本機
patch 清單全數相同。公開附件不含 `full-local`、原版 ZIP、
PC-98 OGG 或內部推廣片。

## 仍保留的驗證限制

- Windows 與 macOS 尚待真機啟動；macOS 尚待簽署與公證。
- 內部推廣片的 PC-98 音樂權利仍為 rights pending，不上傳正式版。
- 真實全滅是原版正常結果；一般強度抽測遇到全滅即停止，
  不用測試專用強化強求通過。
