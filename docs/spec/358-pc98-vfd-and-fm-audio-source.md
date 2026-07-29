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

### 2.1 MAME 軟體清單交叉驗證

MAME 官方 `hash/pc98.xml` 將本作列為 `azurebnd`，兩片均是
1,265,664-byte FDI：

| 磁碟 | CRC32 | SHA-1 |
| --- | --- | --- |
| Disk A | `581e793d` | `c6704602a920859a34996524f5856c8c5ac5f1fc` |
| Disk B | `ba0eaf68` | `07e48bb52ec07b78a821ad311e727b508fa4e948` |

來源：
<https://github.com/mamedev/mame/blob/master/hash/pc98.xml#L12912-L12929>。
MAME 軟體清單只提供身分與雜湊，不提供商業磁碟內容；上述雜湊可作為日後取得
第二份合法 dump 時的完整性 oracle。

VFD 的 `0xFFFFFFFF` descriptor 目前應稱「映像未保存的 sector」，不能直接
等同磁片損壞。實驗中，保留兩個 absent sectors 的 D88 可進 MEGDOS；加入
sector 或修改 FAT root／`LOADER.COM` 的研究副本都在 MEGDOS banner 前停止。
這支持早期完整性／防拷檢查的假說，但尚未定位檢查 routine，因此 confidence
仍為 `hypothesis`。

NP2kai D88 backend instrumentation 進一步證實 `C3/H0/R8/N3` 在 baseline
連續被要求四次；若只讓第一次維持 not-found、第二次起合成零 sector，啟動
狀態會由 `loader.com` 改為停在 MEGDOS banner，且 CPU probe 尚未看到
`INT 21h/AH=4Bh`。所以這不是普通的 DOS 檔案讀取失敗，也不能假定「首讀錯誤、
重試回零」就是正確語意。完整 trace 與結論邊界見
`docs/reference/original-pc98/vfd-runtime-trace.md`。

### 2.2 既有 PC-98／素材抽取經驗的採用範圍

已唯讀回查本機 `~/.claude` 的既有 retro 記憶。可沿用的工作順序是：

1. 先搜尋既有抽取器、解碼器與原生解析度無損截圖，不先假設必須從磁碟格式
   重新逆向。
2. 若畫面是原生解析度、無縮放的 PNG，可用格線相位與已知 tile 尺寸抽取
   畫面已覆蓋的素材，並明確標示其覆蓋範圍。
3. 截圖抽取只能解決畫面素材，不能證明未出現素材、程式行為、音樂曲號或
   缺失磁區語意；這些仍須 executable、runtime trace 或第二份合法 dump。
4. 原始磁碟、抽出的商業素材與音軌留在本機；公開 repository 只保存工具、
   雜湊、來源索引、規格與必要的研究證據。

這些記憶沒有提供可直接套用於本作 `VFD1.00` absent-sector／防拷的既成
解法。因此本節不能取代 2.1 的 NP2kai read trace，也不能把
`C3/H0/R8/N3` 推論成一般壞軌、弱磁區或固定資料。

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
- `docs/reference/original-pc98/vfd-runtime-trace.md`
  (`exact` read sequence；copy-protection semantics 仍是 hypothesis)。

`LOADER.COM` 的 SHA-256 是
`290e1031aea90a76b644f84175556ab0eba85897bc3204a181fcac2f339b18f3`。
其三次 DOS `INT 21h/AH=4Bh` EXEC 依序指向：

- COM offset `0xA2`：`setup.exe`
- COM offset `0xC7`：`mscdrv.exe`
- COM offset `0xED`：`game.exe`

第二次 EXEC 的程式碼位於 file offset `0x3E..0x6D`。這是原 bytes 與 IDA
反組譯一致的 `exact` 結論。直接改寫該區或 FAT directory 會觸發上述早期
停止，因此後續 caller probe 必須改用 emulator memory／interrupt
instrumentation，不再修改開機磁碟。

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

`GAME.OVR` 是 `TPOV` container。`cmd/pc98-ovr-audit` 已由 resident
`GAME.EXE` 的 control records 重建完整 36 段 chain，並分離 code 與
relocation bytes。36 段 code 都沒有 literal `CD D2`；這已排除「raw scan
剛好漏掉 relocation 後 overlay code」的舊假說，但不能排除 interrupt
vector／far-call wrapper。

### 4.1 Borland symbol table 與音訊 wrapper

`GAME.EXE` 的 MZ load image 在 file offset `0x144B0` 結束；其後緊接
little-endian `0x52FB`、version `0x0208` 的 Borland legacy debug table：

- 1,725 筆 9-byte symbol records（`u16 name/type/offset/segment`＋`u8 flags`）；
- 2,305 筆 ASCIIZ names；
- 53 個 modules；
- program flags `0x02`，即 Pascal overlay program。

`cmd/pc98-symbol-audit` 會由 MZ `e_cp/e_cblp` 自動找 header，並以 symbol
count、name count、name-pool byte size 與 EOF 邊界驗證。輸入
`GAME.EXE` SHA-256：
`8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`。

目前已確認：

| symbol | runtime address | IDA address | 語意 |
| --- | --- | --- | --- |
| `SOUNDFX` | `0893:0000` | `sub_18930` | PC speaker／短音效 selector；不是 FM BGM |
| `INITSOUND` | `0893:010D` | `sub_18A3D` | 目前是空 routine |
| `MSCPLAY` | `0893:0114` | `sub_18A44` | 接受 1-based track byte、轉為 0-based、抑制重播同曲、送至 vector `7Eh` |
| `MSCSTOP` | `0893:015E` | `sub_18A8E` | 寫 stop command，再送至 vector `7Eh` |
| `BGMPLAY` | `0893:0177` | `sub_18AA7` | 依 internal area code 選曲，再呼叫 `MSCPLAY` |
| `SOUND` | `08EE:03A6` | 待 caller 交叉驗證 | symbol address exact；尚未宣稱是音效播放入口 |

9-byte record 修正後，以上五個音訊 procedure 都有連續、可重現的 symbol
record；不能回退採用錯誤的 10-byte stride。全域資料另包含 `MUSICNUM`
`0C29:8BE1`、`MUSICSW` `0C29:8BE3`、`MUSICNO` `0C29:8BF3`。

`MSCPLAY` 的底層 `sub_18BDB` 不是 literal `INT D2h`：它以呼叫參數
`7Eh` 索引 IVT、載入該 interrupt vector，恢復一組 register buffer 後
`retf` 進 handler。這解釋了 resident 與全部 overlay code 都找不到
`CD D2`；下一步需確認 `MSCDRV.EXE` 是否另外把 `INT 7Eh` 作為遊戲 wrapper，
再轉送其 `INT D2h` TSR contract。

### 4.2 已反組譯的 area code → track selector

`BGMPLAY`／`sub_18AA7` 讀 `byte_28080`，在若干 game state 下選出 track byte，再呼叫
`MSCPLAY`。目前可逐 instruction 重現的 selector 是：

| area code | 傳入 `MSCPLAY` 的 1-based 值 | driver buffer 的 0-based 值 |
| --- | ---: | ---: |
| `01`, `31` | 3 | 2 |
| `11`, `12`, `15`, `21`, `22`, `23`, `43`, `45` | 4 | 3 |
| `50`, `51` 且 `byte_241A1 == 0` | 5 | 4 |
| `50`, `51` 且 `byte_241A1 != 0` | 6 | 5 |
| `20`, `23`, `40`, `42` | 8 | 7 |
| `02`, `05`, `10`, `35` | 9 | 8 |
| `03`, `04`, `25`, `32`, `33` | 12 | 11 |

area `23` 在原 code 先命中 track 4 分支，因此後面的 track 8 比較實際上
不可達；表格保留兩處比較以反映 bytes，不自行「修正」原程式。

這張表目前只證明 internal area code 與 selector，不證明 Town、Combat、
Dungeon 等人類場景名稱。必須把 `byte_28080` 的寫入點對回 ECL／map
identifier，並以 runtime scene transition 驗證後，才能產生正式
scene-role JSON。

IDA output 只是輔助；上述結論需以原 bytes 與 runtime I/O trace 維持
`exact`。IDA 的通用 interrupt 註解不是本作語意證據。

## 5. 曲目目錄的次級證據

VGMRips 的公開 PC-9801 register-log pack 記錄 YM2203、作曲者安田毅，
並列出 12 首：Title、Character Creation、Town、Thieves Guild、Combat、
Dungeon 1、Wilderness、Village、Dungeon 2、City、Dungeon 3、Ending。

這份資料目前只支持「至少有這 12 個使用情境」與音源晶片，不支持本作
`INT D2h` 曲目編號。它也不能直接提交成 remake 音訊資產。正式映射仍需：

1. 完整 `MSCDRV.EXE` 或等價共用 driver。
2. `GAME.EXE` 的 `INT 7Eh` wrapper 與 `MSCDRV.EXE`／`INT D2h` 轉送關係。
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
- 以 MAME `azurebnd` Disk A／B 身分雜湊核對第二份合法 dump，或證明 VFD
  absent-sector semantics 與其可逆對應。
- 逐一命名 `INT D2h` dispatch，至少確認 play、stop、status、track select。
- 對 12 個曲目建立 caller／scene／runtime YM trace 三方映射。
- 成功從標題進入正常遊戲路徑，擷取 title、town、combat 三次轉場。
- 在不提交原始音軌的前提下，確立合法且可重現的 runtime import／播放方案。
