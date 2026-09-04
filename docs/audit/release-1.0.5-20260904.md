# 1.0.5-20260904 完整打包與公開發行稽核

日期：2026-09-04  
遊戲來源 commit：`1c51db3c`  
共用 engine：`1d271d58716c`  
打包依據：`tools/package-release.sh 1.0.5-20260904` 的逐項 Docker 命令

## 這一版的玩家可見變更

角色 VIEW 底部的物品提示、建角選項清單的操作提示，以及外層隊伍選單的
「隊伍人數：n／6」，先前都畫在原版建角框的框帶位置上——字有畫出來，被框的
像素蓋住，畫面上只剩上半截。三行改用這個框自己的安全區（spec 1252）。

只有這一項。其餘與 1.0.4 相同。

## 本機完整交付

`dist-all/1.0.5-20260904/SHA256SUMS.json` 登記八件產物：Linux x86_64 AppImage、
Windows x86_64 ZIP、macOS x86_64／arm64 ZIP，各有 `patch` 與 `full-local`。
公開 `patch` 不含原版 ZIP 與 `pc98-bgm-selector-*.ogg`；本機 `full-local` 四平台
封包均含本機合法持有的原版 ZIP 與 PC-98 OGG。

Linux full-local AppImage 與 Windows full-local EXE 已在 Docker／Xvfb 與
Docker／Wine／Xvfb 實際啟動。兩張開場截圖皆為 49,576 bytes，SHA-256 同為
`d0e5d6399273d59dea3f43f58f245c61c19ed3adf0b3d0e070989cb90d234362`
——與 1.0.4 逐位元組相同，這一版沒有動到開場畫面。
macOS 雙架構只完成交叉編譯與封包，尚未經 macOS 真機、簽署或公證。

修正本身另以 AppImage 直接驗過：用發行包裡的 AppImage 跑 `-character-view`
與 `-character-creation`，畫面與修正後的原始碼版**逐位元組相同**，確認修正
確實進了封包而不只是在工作區。

## 產物雜湊

| 類型 | 平台 | bytes | SHA-256 |
|---|---|---:|---|
| patch | Linux x86_64 | 140350656 | `3bf0b2846fd7232c1f7850f65d7230825bc9c2d426f862ee3c8b10b1a3b7ceab` |
| patch | Windows x86_64 | 151689432 | `282c13e78a9105c3d50663b75e19c871083b587cce62c9970b102f285cc0f911` |
| patch | macOS x86_64 | 152028244 | `0cb10ea29c157db05a0b8cce4826ee1b20c1cc53fed8db62860318667ad20709` |
| patch | macOS arm64 | 151724159 | `d68424c1672872f79fdc5ef1a7893e712eb55c85d78f1959c3561e85fa2bc105` |
| full-local | Linux x86_64 | 163484864 | `3d6567b87d0e6122aead46058b581807402132aae8e3b17e9a49fd774dadb951` |
| full-local | Windows x86_64 | 174802670 | `eef848b238dfffb4ddc42da2c7911d0996cd3b08782be700073f23f8c99e05f3` |
| full-local | macOS x86_64 | 175142470 | `5a5500d3b8a0b5403adc14bba4bcceb66a16bc610f6b03e527e829aeb1934213` |
| full-local | macOS arm64 | 174838385 | `2b0a3c6f5016d22d3b2957c897cffd44df29c53957aa318ca420899bb94b4cf5` |

## 公開邊界與 Docker 衛生

GitHub Release 只上傳四件 `patch` 與 `PATCH-SHA256SUMS.txt`；不上傳
`full-local`、原版 ZIP、PC-98 OGG 或 smoke 截圖。建置、封包與 smoke 全部使用
一次性 Docker 容器；本輪不建立新工具 image。

發布後由 GitHub API 讀回驗證：Release 為 `draft=false`、`prerelease=false`、
`target_commitish=main`，五個附件狀態皆為 `uploaded`，四個平台附件的 digest
與本機 `PATCH-SHA256SUMS.txt` 逐項相符。tag `v1.0.5-20260904` 指向 `1c51db3c`。

## 仍未完成

- Windows 與 macOS 真機啟動；macOS 簽署與公證。
- 日文片假名待母語者複核。
- 內部推廣片的音樂與原版圖像權利仍未釐清。
