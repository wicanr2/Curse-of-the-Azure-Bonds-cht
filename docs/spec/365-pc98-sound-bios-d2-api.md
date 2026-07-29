# Spec 365 — PC-9801 Sound BIOS／INT D2h ABI

狀態：`READY`（官方介面表與本作實際 client 集合）

本規格辨識 `MSCDRV.EXE` 經由 `INT D2h` 使用的服務。它只證明驅動會呼叫
哪些 PC-9801 Sound BIOS 命令與如何傳參；不代表缺失音序列、實際曲名、
播放時序或跨平台播放器已完成。

## 1. 權威來源與本機證據

主要規格來源是 NEC《PC-9800 Technical Databook BIOS 1992》
「サウンド BIOS」章，第 357–377 頁：

- 來源：
  <https://vtda.org/docs/computing/NEC/PC-9800TechnicalDataBookBIOS%2BOCR_1992.pdf>
- 本次研究副本 SHA-256：
  `23ca3d4f74fce498e84eb073014d67ad3b2a3f3845e326986921846e4e0c7885`
- PDF 與 OCR 只留在本機研究暫存區，不提交 repository。

本作 client 來自：

| 輸入 | SHA-256 |
| --- | --- |
| 殘缺 `MSCDRV.EXE` | `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |

工具為 Docker 內指定的 IDA Pro 9.4，以及
`scripts/ida/pc98_sound_bios_audit.py`。IDA 函式與 xref 再由
`internal/pc98music` 直接核對原始 file bytes。所有命令 wrapper 位於
`0x135D..0x14A1`，早於殘缺 sector 的 `0x4000..0x43FF`。

## 2. CEE0 是 Sound BIOS 介面表

NEC 手冊說明 Sound BIOS 不自行限定一個固定的 user interrupt。呼叫端必須
從實體位址 `CEE00h` 的介面資料取得 Sound BIOS entry，再將它安裝到選定的
interrupt vector；N88-BASIC interpreter 使用 `D2h`。

這與本作 bytes 精確吻合：

- `MSCDRV sub_100CC` 檢查 `CEE0:0004 == 0x00D2`；
- `sub_110CA` 讀 `CEE0:[0006]` 作 entry offset；
- 再以 DOS `INT 21h/AH=25h` 把 vector D2h 設成 `CEE0:entry`。

因此舊文件中的「`CEE0` producer 未知」不再成立。`CEE0` 是 NEC Sound BIOS
約定的固定介面表，不是 `MSCDRV` 自身段，也不是本作私有資料。尚未做的是
NP2kai runtime 對開機後 `CEE0:0000..0007` 的逐 byte capture；這只影響
實機安裝時序記錄，不推翻官方 ABI 與本作 consumer。

## 3. 本作實際使用的命令

下表的名稱與 register contract 取自 NEC 手冊；file offset 與「本作有此
wrapper」由 IDA＋原始 bytes 證實。

| AH | 命令 | 本作參數 | file offset |
| ---: | --- | --- | ---: |
| `00` | `INITIALIZE` | `ES` 指向控制／工作區段 | `0x135D` |
| `02` | `CLEAR` | 無 | `0x1372` |
| `10` | `READREG` | `AL`=OPN register；回傳 `BX` | `0x1384` |
| `11` | `WRITEREG` | `AL`=OPN register，`BL`=資料 | `0x13A1` |
| `12` | `SETTOUCH` | `AL`=聲道，`BL`=觸鍵／閘門比 | `0x13B4` |
| `13` | `NOTE` | `AL`=聲道，`BH`=音高，`BL`=音長 | `0x13CA` |
| `14` | `SETLENGTH` | `AL`=聲道，`BL`=預設音長 | `0x13DD` |
| `16` | `SETPARABLOCK` | `AL`=聲道，`ES:BX`=參數塊，`DL`=種類 | `0x140D` |
| `17` | `READPARA` | `AL`=聲道，`BL`=參數編號；回傳 `BX` | `0x1420` |
| `18` | `WRITEPARA` | `AL`=聲道，`BL`=參數編號，`DX`=值 | `0x1438` |
| `19` | `ALLSTOP` | 無 | `0x143D` |
| `1A` | `CONTPLAY` | 無 | `0x1442` |
| `1B` | `MODUON` | `AL`=聲道 | `0x1462` |
| `1C` | `MODUOFF` | `AL`=聲道 | `0x1472` |
| `1D` | `SETINTCOND` | `AL`=聲道，`ES:BX`=回呼，`CX`=條件 | `0x148A` |
| `1E` | `HOLDSTATE` | `AL`=聲道，`BL`=維持音長 | `0x1452` |
| `1F` | `SETVOLUME` | `AL`=聲道，`BL`=音量 | `0x149D` |

官方表另有 `01 PLAY` 與 `15 SETTEMPO`。本作已確認的 wrapper 區沒有這兩個
`INT D2h` client；不得因官方 API 存在便宣稱 MSCDRV 經由它們播放。driver
自己管理音序列，並可透過參數塊／延遲命令或 direct OPN path 達成相關行為。
其精確 tempo consumer 仍待追蹤。

## 4. Direct YM2203 path

IDA 另定位兩個不經 Sound BIOS wrapper 的硬體 helper：

- `sub_11075`／file `0x1275`：等待 `0x188` busy bit，將 `AL` register
  寫至 `0x188`，再把 `BL` data 寫至 `0x18A`；
- `sub_110AB`／file `0x12AB`：選擇 `AL` register 後從 `0x18A` 讀回 `BL`。

NEC 手冊警告 Sound BIOS 運作時直接改寫其控制的 OPN registers 可能造成未定
行為。本作確實同時具備 BIOS 與 direct path，所以 remake 的匯入器不能只
記錄 D2h 命令；之後的 runtime trace 必須同時捕捉 `INT D2h` 與
`0x188/0x18A` I/O。

## 5. 可重現驗證

```text
go test ./internal/pc98music ./cmd/pc98-music-audit
go run ./cmd/pc98-music-audit GAME.EXE MSCDRV.EXE
```

auditor 會先驗證兩個 executable SHA-256，再逐一驗證 17 組
`MOV AH,command; INT D2h` bytes、兩個 direct YM helper 與 spec 364
既有 bridge anchors。輸出 JSON 的 `sound_bios_services` 保存命令、
官方名稱、file offset 與 register contract。

IDA batch：

```text
idat -A -Lsound-bios-ida.log \
  -Sscripts/ida/pc98_sound_bios_audit.py MSCDRV.EXE.i64
```

實際執行仍須遵守 `AGENTS.md`：只能在 Docker 內使用
`/home/anr2/ida_94_official/dist` 對應的 IDA Pro 工具鏈。

## 6. 尚未完成

- NP2kai 開機後 `CEE0` table、D2h vector 與每次命令的 runtime trace。
- `SETINTCOND` callback、timer A/B、direct YM path 的完整 caller data flow。
- 找回 `MSCDRV.EXE 0x4000..0x43FF`，恢復完整音序列與 loop metadata。
- 將 title／town／combat 的 selector、D2h、YM register stream 三方對齊。
- 取得可合法使用的音源或從使用者媒體 runtime import，才接跨平台播放器。
