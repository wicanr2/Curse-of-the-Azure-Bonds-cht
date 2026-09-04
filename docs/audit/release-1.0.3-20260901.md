# 1.0.3-20260901 完整打包與公開發行稽核

日期：2026-09-01  
遊戲來源 commit：`b1a2bb19d`  
共用 engine：`8cc96f8650ef`（engine 歷史於 2026-09-04 改寫 commit 作者信箱，
這個 hash 已不存在；同一份內容改寫後是 `1d271d58716c`）  
打包依據：`tools/package-release.sh 1.0.3-20260901` 的逐項 Docker 命令

## 本機完整交付

`dist-all/1.0.3-20260901/SHA256SUMS.json` 登記八件產物：Linux x86_64
AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP，各有 `patch` 與
`full-local`。公開 `patch` 不含原版 ZIP 與 `pc98-bgm-selector-*.ogg`；本機
`full-local` 四平台封包均含本機合法持有的原版 ZIP 與 PC-98 OGG。

Linux full-local AppImage 與 Windows full-local EXE 已在 Docker／Xvfb 與
Docker／Wine／Xvfb 實際啟動。兩張開場截圖皆為 49,576 bytes，SHA-256 同為
`d0e5d6399273d59dea3f43f58f245c61c19ed3adf0b3d0e070989cb90d234362`。
macOS 雙架構只完成交叉編譯與封包，尚未經 macOS 真機、簽署或公證。

## 產物雜湊

| 類型 | 平台 | bytes | SHA-256 |
|---|---|---:|---|
| patch | Linux x86_64 | 140342464 | `67910528d593e9a434b2a07f1e04c94f327005c5c20340014bd80073ae05579d` |
| patch | Windows x86_64 | 151680526 | `656c283ca93ca276f96e14463daaf69d247122e5d7d332ef70dee1c5819c83e4` |
| patch | macOS x86_64 | 152017737 | `325231d2c2b1f210e5e7da4b6352b461fb82b5b24c90fb2c75105e277da2d5f7` |
| patch | macOS arm64 | 151714945 | `ec399a90ba8e762fa1ce7180ae1992bce2dafe6136cade369f75281dcf982b43` |
| full-local | Linux x86_64 | 163476672 | `fceb78c745a8727c3773dee379a8a49631f571575cf7412c6c59211df66d5d49` |
| full-local | Windows x86_64 | 174793764 | `c349cde5b64547aad7bfdb7faddda5dbc9ec345e95ebaf06ba95332e81207c5f` |
| full-local | macOS x86_64 | 175131963 | `fde465cd21091cdc6fd0d2aed3cea5f10796333cb7d140ba7cb9a70b4d80a4db` |
| full-local | macOS arm64 | 174829171 | `1359fa70afe0b8cf0d709f6795e93fe874a57574ce0b39c49080917c2f54c973` |

## 公開邊界與 Docker 衛生

GitHub Release 只允許上傳四件 `patch` 與 `PATCH-SHA256SUMS.txt`；不得上傳
`full-local`、原版 ZIP、PC-98 OGG 或 smoke 截圖。建置、封包、內容稽核與
smoke 全部使用一次性 Docker 容器；本輪不建立新工具 image。

2026-09-01 發布後由 GitHub API 讀回驗證：Release 為非草稿、非預覽版本，
五個附件狀態皆為 `uploaded`；四個平台附件的 size 與 SHA-256 digest 均與
本機 patch manifest 完全相同。公開附件名稱與內容清冊沒有 `full-local`、
原版 ZIP、PC-98 OGG 或 smoke 截圖。
