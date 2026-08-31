# 1.0.2-20260831 完整打包與公開發行稽核

日期：2026-08-31  
遊戲來源 commit：`b94ddbd89b65821482201d8a6bfa2af865c4de11`  
共用 engine：`0819c64e451d`  
打包命令：`tools/package-release.sh 1.0.2-20260831`

## 本機完整交付

`dist-all/1.0.2-20260831/SHA256SUMS.json` 登記八件產物：Linux x86_64
AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP，各有 `patch` 與
`full-local`。八件檔案大小與 SHA-256 已由獨立容器重新計算並全數吻合；六個
ZIP 的 CRC、`LICENSE` 與 `NOTICE.md` 也已抽查通過。

- 公開 `patch` 的三個 ZIP 與 AppImage 均不含 `curseoftheazurebonds.zip` 或
  `pc98-bgm-selector-*.ogg`。
- `full-local` 四平台封包均含本機合法持有的遊戲來源與 PC-98 OGG，只供本機
  驗收，不提交 Git，也不上傳 GitHub Release。
- Linux full-local AppImage 與 Windows full-local EXE 已分別在 Docker／Xvfb
  與 Docker／Wine／Xvfb 實際啟動。兩張開場截圖 SHA-256 同為
  `d0e5d6399273d59dea3f43f58f245c61c19ed3adf0b3d0e070989cb90d234362`。
- macOS 雙架構已交叉編譯與封包，尚未經 macOS 真機、簽署與公證；Wine smoke
  也不取代 Windows 真機驗收。

## 公開發行邊界

GitHub Release 標籤為 `v1.0.2-20260831`，只公開四件 `patch` 與由本機
manifest 產生的 `PATCH-SHA256SUMS.txt`。不得上傳 `full-local`、原版 ZIP、
PC-98 OGG 或權利尚未釐清的內部推廣片。

發布後必須從 GitHub 讀回核對：Release 非 draft、非 prerelease，五個附件均為
`uploaded`，四個平台封包的 GitHub digest 與本機 patch manifest 相同。

## Docker 與檔案衛生

本輪建置、封包、內容抽查與 smoke 均使用一次性 `--rm` 容器；產物擁有者為目前
使用者，沒有留下 CoAB 建置／smoke 容器，也沒有為本輪另建重複工具 image。
