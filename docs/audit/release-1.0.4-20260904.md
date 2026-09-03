# 1.0.4-20260904 完整打包與公開發行稽核

日期：2026-09-04  
遊戲來源 commit：`a71feb7a9`  
共用 engine：`8cc96f8650ef`  
打包依據：`tools/package-release.sh 1.0.4-20260904` 的逐項 Docker 命令

## 這一版的玩家可見變更

營地選單改成原版那條底部橫排指令列（spec 1251）。先前的縱排在世界地圖畫面
只放得下三項，而營地選單有七項——後四項畫到畫布外面，玩家看不到也點不到。

英文與日文介面文字大幅訂正：先前兩份不是原版英文，是把繁中再翻一次，所以專名
一律走樣（`Alias`→`Alice`、`Phlan`→`Fran`、`Cleric`→`Reverend`，日文把「戰斧」
寫成 `アクセシビリティ`）。物品與法術名回到原版資料表，地名與 NPC 依譯名表，
日文依原版英文轉寫成片假名（**由非母語者轉寫，待複核**）。繁中未變動。

## 本機完整交付

`dist-all/1.0.4-20260904/SHA256SUMS.json` 登記八件產物：Linux x86_64 AppImage、
Windows x86_64 ZIP、macOS x86_64／arm64 ZIP，各有 `patch` 與 `full-local`。
公開 `patch` 不含原版 ZIP 與 `pc98-bgm-selector-*.ogg`；本機 `full-local` 四平台
封包均含本機合法持有的原版 ZIP 與 PC-98 OGG。

Linux full-local AppImage 與 Windows full-local EXE 已在 Docker／Xvfb 與
Docker／Wine／Xvfb 實際啟動。兩張開場截圖皆為 49,576 bytes，SHA-256 同為
`d0e5d6399273d59dea3f43f58f245c61c19ed3adf0b3d0e070989cb90d234362`
——**與 1.0.3 那兩張逐位元組相同**，所以這一版的改動沒有動到開場畫面。
macOS 雙架構只完成交叉編譯與封包，尚未經 macOS 真機、簽署或公證。

## 產物雜湊

| 類型 | 平台 | bytes | SHA-256 |
|---|---|---:|---|
| patch | Linux x86_64 | 140342464 | `acb467342b4549a91091c26d9f32baaaa3674431bf540b4bbc933a40a9d74f7f` |
| patch | Windows x86_64 | 151682046 | `39bafa3e75def44dcbf6d8ab19e3921f92c441242a1944e219ea81fd21ff72cb` |
| patch | macOS x86_64 | 152020861 | `05d39a0fb20e4dd458e9a8934e1932343b381bdf8d853a6416714f6ba9ff1538` |
| patch | macOS arm64 | 151716728 | `a474fa3e173a839628c117759b6dbdd1f71483793c863118ef107f79bf559ad3` |
| full-local | Linux x86_64 | 163476672 | `bb0a9ece2116ba7c36550a9ebb38fcf9bf7147ffaa043ba962acdcacfd43b8a9` |
| full-local | Windows x86_64 | 174795284 | `0596e0662324691bdd7d7a69c8e354ff4d5cd9a04bcd761adb3960080e86f98d` |
| full-local | macOS x86_64 | 175135087 | `3a5e2705a96460ae9bc79749613a285ab30cb0736c9db54caa7a65b660f6e16f` |
| full-local | macOS arm64 | 174830954 | `735f423dc4544a7f084ea96cf669e6ac6a8b1e433677816cb1210940701145b3` |

## 公開邊界與 Docker 衛生

GitHub Release 只上傳四件 `patch` 與 `PATCH-SHA256SUMS.txt`；不上傳
`full-local`、原版 ZIP、PC-98 OGG 或 smoke 截圖。建置、封包與 smoke 全部使用
一次性 Docker 容器；本輪不建立新工具 image。

2026-09-04 發布後由 GitHub API 讀回驗證：Release 為非草稿、非預覽版本
（`draft=false`、`prerelease=false`、`target_commitish=main`），五個附件狀態皆為
`uploaded`；四個平台附件的 size 與 SHA-256 digest 均與本機 `PATCH-SHA256SUMS.txt`
相同。tag `v1.0.4-20260904` 指向 `a71feb7a9`。

## 仍未完成

- Windows 與 macOS 真機啟動；macOS 簽署與公證。
- 日文片假名待母語者複核。
- 內部推廣片的音樂與原版圖像權利仍未釐清。
