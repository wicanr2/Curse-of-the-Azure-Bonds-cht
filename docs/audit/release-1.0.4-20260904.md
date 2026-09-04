# 1.0.4-20260904 完整打包與公開發行稽核

日期：2026-09-04  
遊戲來源 commit：`4ea0c95b`（附件重出的建置來源；初版首發是 `a71feb7a9`）  
共用 engine：`1d271d58716c`（原記 `8cc96f8650ef`；engine 歷史於 2026-09-04
改寫 commit 作者信箱，所有 hash 因此改變，內容與 tree 未動）  
打包依據：`tools/package-release.sh 1.0.4-20260904` 的逐項 Docker 命令

## 這一版的玩家可見變更

營地選單改成原版那條底部橫排指令列（spec 1251）。先前的縱排在世界地圖畫面
只放得下三項，而營地選單有七項——後四項畫到畫布外面，玩家看不到也點不到。

英文與日文介面文字大幅訂正：先前兩份不是原版英文，是把繁中再翻一次，所以專名
一律走樣（`Alias`→`Alice`、`Phlan`→`Fran`、`Cleric`→`Reverend`，日文把「戰斧」
寫成 `アクセシビリティ`）。物品與法術名回到原版資料表，地名與 NPC 依譯名表，
日文依原版英文轉寫成片假名（**由非母語者轉寫，待複核**）。繁中未變動。

## 2026-09-04 授權變更與附件重出

授權條款由 PolyForm Noncommercial 1.0.0 改為 RRSAL-1.0（復古重製
source-available 授權條款 1.0）：非商業使用、修改與散布免費，遊戲實況、錄影、
直播與其平台分潤由條款第 4 條明示允許，商業使用保留另談。發行包裡的 `LICENSE`、
`NOTICE.md` 與說明檔跟著換，AppImage 另在 `usr/share/doc/azure-bonds-remake/`
放一份（rulebook 85 要求的第四個位置，先前只有 AppDir 根目錄）。

同一個版本號重新打包並替換五個 Release 附件。**程式行為沒有變動**：Linux 與
Windows（Wine）的開場截圖都是 49,576 bytes、SHA-256
`d0e5d6399273d59dea3f43f58f245c61c19ed3adf0b3d0e070989cb90d234362`，與重打包前
逐位元組相同——也就是仍然與 1.0.3 那兩張相同。

tag `v1.0.4-20260904` 已改指實際建置來源 `4ea0c95b`（原本指向初版首發的
`a71feb7a9`），這樣從 tag 重建得出現在的附件。兩個 commit 之間只有授權文件與
engine 相依的 hash 修正。下方雜湊表是重出後的值，初版附件的雜湊不再有效。

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

重出後的值（2026-09-04）：

| 類型 | 平台 | bytes | SHA-256 |
|---|---|---:|---|
| patch | Linux x86_64 | 140350656 | `6e2a2c8d4b933e3547852520c968958abcbf95dd30a6616f32de188e4f600bde` |
| patch | Windows x86_64 | 151689453 | `06d89a79756dbf14e5e78d9d8803564b60a9240eee38f83deb9ca8848a97d59e` |
| patch | macOS x86_64 | 152028263 | `c3bb8955cf5bbdc44b5d1254122c2c10f3ba23b35046ff0c3d6a310ebf0b332c` |
| patch | macOS arm64 | 151724140 | `c335ba53c422a28294e5fdc2e53db4df2318af3ba974e3486093ada7faf4f3d6` |
| full-local | Linux x86_64 | 163484864 | `52a9539a8d715841a629dacb86c750ba798f658ed06a3557304bc15252e6d388` |
| full-local | Windows x86_64 | 174802691 | `75cc3ec8710e29dee195ccd45f87b6bd125b1277cbfa2279a2beba99ea47f12c` |
| full-local | macOS x86_64 | 175142489 | `307414c62d9d4e6c836609c1b27a7f9db29f491acd03449aa9353c993c88c148` |
| full-local | macOS arm64 | 174838366 | `315a33aa9c183b4973213aebdd178da6ac2321c7e897f6824840fc01f06f6769` |

## 公開邊界與 Docker 衛生

GitHub Release 只上傳四件 `patch` 與 `PATCH-SHA256SUMS.txt`；不上傳
`full-local`、原版 ZIP、PC-98 OGG 或 smoke 截圖。建置、封包與 smoke 全部使用
一次性 Docker 容器；本輪不建立新工具 image。

2026-09-04 發布後由 GitHub API 讀回驗證：Release 為非草稿、非預覽版本
（`draft=false`、`prerelease=false`、`target_commitish=main`），五個附件狀態皆為
`uploaded`；四個平台附件的 size 與 SHA-256 digest 均與本機 `PATCH-SHA256SUMS.txt`
相同。當時 tag `v1.0.4-20260904` 指向 `a71feb7a9`。

授權重出後再讀一次：Release 仍是 `draft=false`、`prerelease=false`、
`target_commitish=main`，五個附件皆 `uploaded`，digest 與上表的重出值逐項相同，
`PATCH-SHA256SUMS.txt` 下載回來與本機檔案內容一致。tag 已改指 `4ea0c95b`
（annotated，tag object `3c96bed2`）。

## 仍未完成

- Windows 與 macOS 真機啟動；macOS 簽署與公證。
- 日文片假名待母語者複核。
- 內部推廣片的音樂與原版圖像權利仍未釐清。
