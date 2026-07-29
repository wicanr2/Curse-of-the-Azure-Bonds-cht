# Spec 358 — PC-98 VFD 與 FM 音訊來源

狀態：`IN PROGRESS`

本規格只記錄可由使用者提供媒體、實機執行與反組譯重現的事實。原始 FDD、
由其轉出的 D88／raw image、`MSCDRV.EXE` 與音軌不得提交 GitHub。

## 1. 原始媒體身分

| 磁碟 | 大小 | SHA-256 |
| --- | ---: | --- |
| Disk 1 | 1,309,692 | `d840f47c2b8391e2495f686f7b02ed37ae972e3c8149e117a60f9bb9c44fcd55` |
| Disk 2 | 1,287,164 | `38f56ab7f17690b72afa6e7b6b6462a9ec5511f53b7567303e1dcc681982bfbe` |

兩者皆為 `VFD1.00`。本作使用 77 cylinders × 2 heads × 8 sectors ×
1024 bytes 的 PC-98 2HD geometry。每軌在 offset `0xDC` 起保留 26 筆、
每筆 12 bytes 的 sector descriptor；實際使用前八筆：

- `+0..3`：C／H／R／N。
- `+4..7`：四個 VFD flags；一般 payload sector 為
  `FF 00 01 01`。
- `+8..11`：little-endian payload offset。
- payload offset `0xFFFFFFFF` 表示該 sector 沒有保存在映像內，不能當成
  內容全零或一般讀取錯誤而自行猜測。

`cmd/pc98-vfd-audit` 是只讀、可測試的稽核工具；預設 geometry 即本作格式，
會輸出媒體雜湊與所有缺失 CHRN。

## 2. 缺失磁區與影響

Disk 1：

| CHR | raw LBA | FAT12 cluster／檔案 | 檔案 offset | 影響 |
| --- | ---: | --- | ---: | --- |
| 3/0/8 | 55 | cluster 46／`MSCDRV.EXE` | `0x4000` | 音樂常駐驅動遺失 1024 bytes |
| 42/0/2 | 673 | cluster 664／`CED3.DAX` | `0x10000` | `CED3.DAX` 尾端遺失 1024 bytes |

Disk 2 的 24 個缺失 sector 位於 cylinder 75 head 1 與 cylinder 76
兩面，需在理解 Disk 2 自訂檔案系統後再判定用途。

因此目前從 Disk 1 FAT12 解出的 `MSCDRV.EXE` 雖可反組譯，不能宣稱是完整
driver，也不能把缺失區補零後所得行為當作原版。

## 3. PC-98 實機執行

使用 NP2kai `0.86.0.22`／commit `e2dc904`，在 Docker／Xvfb、無網路環境
執行。原始 VFD 唯讀掛載；執行時只使用可拋棄的可寫研究副本。

已驗證：

1. NP2kai 內建 BIOS 顯示 `CPU MODE High` 與 memory check。
2. 正確 2HD D88、重新 Reset 後可進入 `MEGDOS 0.25`。
3. MEGDOS 讀取 `CONFIG.SYS` 的 `shell=loader.com`。
4. `loader.com` 依序執行 `setup.exe`、`mscdrv.exe`、`game.exe`。
5. 含缺失 `MSCDRV.EXE` sector 的副本在載入 shell／driver 前後停止，
   尚未進入遊戲標題；這與 driver 的 1 KiB 缺口一致，但仍需以完整 sector
   或執行 trace 證明確切停止位置。

實機證據：

- `docs/reference/original-pc98/megdos-loader-boot.png`
  (`exact`，只證明 MEGDOS／loader 啟動，不是遊戲 GUI oracle)。

## 4. `MSCDRV.EXE` 已確認介面

目前不完整檔案：

- size：19,840 bytes。
- SHA-256：
  `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5`。
- MZ 16-bit executable；初始化後以 DOS `INT 21h / AX=3100h`
  terminate-and-stay-resident。
- 安裝／接管 `INT D2h`。
- 對 I/O port `0x188` 寫 YM register、對 `0x18A` 寫／讀 data，並輪詢
  `0x188` busy bit；這符合 PC-9801-26K 的 YM2203 路徑。
- timer handler 讀取 YM status、處理 timer flag 並推進播放。
- dispatch 中已看見 AH `00h`、`02h`、`10h..14h`、`16h..1Fh`；各功能
  的輸入、輸出與曲目索引仍須由 `GAME.EXE` caller 和完整 driver 交叉確認。

IDA output 只是輔助；上述結論需以原 bytes 與 runtime I/O trace 維持
`exact`。IDA 的通用 interrupt 註解不是本作語意證據。

## 5. 曲目目錄的次級證據

VGMRips 的公開 PC-9801 register-log pack 記錄 YM2203、作曲者安田毅，
並列出 12 首：Title、Character Creation、Town、Thieves Guild、Combat、
Dungeon 1、Wilderness、Village、Dungeon 2、City、Dungeon 3、Ending。

這份資料目前只支持「至少有這 12 個使用情境」與音源晶片，不支持本作
`INT D2h` 曲目編號。它也不能直接提交成 remake 音訊資產。正式映射仍需：

1. 完整 `MSCDRV.EXE` 或等價共用 driver。
2. `GAME.EXE`／`GAME.OVR` 的 `INT D2h` caller。
3. NP2kai runtime 的 scene transition 與 YM register trace。

來源：

- <https://vgmrips.net/packs/pack/curse-of-the-azure-bonds-nec-pc-9801>
- <https://www.mobygames.com/game/503/curse-of-the-azure-bonds/credits/pc98/>

## 6. Remake contract

PC-98 音樂不能 hardcode 在 Go scene switch：

- engine 只接受作品中立的 music intent（play／stop／fade／resume）、
  scene role、loop 與 playback adapter。
- CoAB game pack JSON 保存 scene role → track stable ID、來源平台、
  原曲編號、loop metadata 與證據。
- 原版 DOS sound selector／PC Speaker 或 Tandy 行為是忠實 DOS theme。
- PC-98 YM2203 配樂是獨立可切換 theme；若需要使用者原始媒體，啟動時
  檢查雜湊並即時匯入，不把商業音軌提交 GitHub。
- 音樂與 sound effect 各有獨立音量／開關；場景切換與 save/load 後的
  resume 必須 deterministic。

## 7. READY 門檻

本規格只有在下列證據齊全後才能改為 `READY`：

- 找回或交叉重建 `MSCDRV.EXE 0x4000..0x43FF`，並記錄來源與雜湊。
- 逐一命名 `INT D2h` dispatch，至少確認 play、stop、status、track select。
- 對 12 個曲目建立 caller／scene／runtime YM trace 三方映射。
- 成功從標題進入正常遊戲路徑，擷取 title、town、combat 三次轉場。
- 在不提交原始音軌的前提下，確立合法且可重現的 runtime import／播放方案。
