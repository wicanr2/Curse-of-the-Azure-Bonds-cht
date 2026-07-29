# Spec 364 — PC-98 音樂 IVT 7Eh／INT D2h bridge

狀態：`READY`（遊戲 wrapper、7Eh public ABI 與 D2h bridge）

本規格只回答 `GAME.EXE` 如何把 play／stop intent 交給
`MSCDRV.EXE`，以及後者如何建立低階 `INT D2h` 服務。它不代表缺失的音樂
序列、曲名、YM2203 runtime trace 或 remake 播放器
已完成。

## 1. 證據與缺失邊界

| 輸入 | SHA-256 |
| --- | --- |
| `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` |
| 殘缺 `MSCDRV.EXE` | `bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5` |

`MSCDRV.EXE` 的媒體缺口是 file offset `0x4000..0x43FF`。本規格使用的
driver anchors 位於 `0x0280..0x1376`，全部早於缺口；因此 7Eh handler、
安裝程序與 D2h client ABI 不受遺失 sector 影響。缺口仍可能包含音樂資料或
其他程式，不能由這項隔離結論推成「driver 已完整」。

工具：

- IDA Pro 9.4：`/home/anr2/ida_94_official/dist`，只在 Docker 內執行。
- `scripts/ida/pc98_music_wrapper_audit.py`
- `scripts/ida/pc98_mscdrv_bridge_audit.py`
- `cmd/pc98-music-audit`

商業 executable、IDA database 與 batch log 不進 repository。

## 2. GAME.EXE 的 register-image trampoline

Borland symbols 與 IDA 位址：

- `MSCPLAY 0893:0114`／IDA `0x18A44`／file `0x93E4`
- `MSCSTOP 0893:015E`／IDA `0x18A8E`／file `0x942E`
- trampoline IDA `0x18BDB`／file `0x957B`

`MSCPLAY` 先把 1-based selector 減一並抑制同曲重播，之後建立 18-byte
register image。與音樂命令直接相關的第一個 word 是：

```text
play: AX = 0x00TT
stop: AX = 0x01TT
```

`TT` 是 0-based track。`MSCSTOP` 只把 command byte 改為 1；低 byte 保留
目前 track，但 stop handler 不使用它。

trampoline 的 exact 行為：

1. 以傳入的 `0x7E × 4` 索引 IVT。
2. 依序從 18-byte image 載入
   `AX/BX/CX/DX/BP/SI/DI/DS/ES`。
3. 建立返回 `CS:0045` 的 frame，`cli; retf` 進 IVT handler。
4. handler `iret` 後，`0x18C15..0x18C39` 將九個 register 與 flags 寫回
   同一 image，再 `retf 6`。

所以這不是 literal `INT 7Eh`，但 ABI 與呼叫該 interrupt vector等價，且
可保存輸出 register。

## 3. MSCDRV.EXE 安裝 IVT 7Eh

driver file `0x02E3..0x02F5`：

```text
LES BX, CS:[000F2]  ; dword value 0000:01F8
MOV ES:[BX], 0080
MOV ES:[BX+2], CS
```

`0x01F8 = 0x7E × 4`，因此它把 IVT 7Eh 直接設成自身 `CS:0080`。
這排除了「7Eh 由未知外部程式轉送」的舊假說。

`CS:0080`／file `0x0280` handler 保存全部一般 register 後：

- `AH == 0`：以 `AL` 為 0-based track，初始化低階服務並選曲；
- `AH != 0`：進入停止／清除播放狀態；`GAME.EXE` 實際傳入 `AH == 1`；
- 最後恢復 register 並 `iret`。

track consumer `sub_1021E` 會再次抑制同一 track，從 `track × 2 + 0x330`
取該 track 的資料指標，保存到播放狀態。特殊 track `0x0E` 的 loop／mode
參數與其他曲不同，但其人類曲名尚未證明。

## 4. 7Eh handler 與 INT D2h 的關係

`sub_110CA`／file `0x12CA`：

1. `INT 21h AX=35D2h` 保存舊 D2h vector。
2. 讀取固定服務段 `CEE0:0006` 的 handler offset。
3. `INT 21h AX=25D2h` 將 D2h 設成該 handler。

啟用前 `sub_100CC` 會檢查 `CEE0:0004 == 0x00D2`。NEC 官方
《PC-9800 Technical Databook BIOS 1992》第 357–377 頁已證明 `CEE0` 是
Sound BIOS 固定介面表，`CEE0:[0006]` 是 entry offset；N88-BASIC 預設以
D2h 呼叫。它不是 driver 自身 code segment。命令表與本作 client 的完整
交叉驗證見 spec 365。

安裝後可見的 exact D2h client：

- `AH=00h`：初始化六組 channel descriptor；
- `AH=02h`：shutdown／restore 路徑使用；
- `AH=10h..14h`、`16h..1Fh`：官方命名及 register contract 已由 spec 365
  確認；本作沒有觀察到 `AH=01h PLAY` 與 `AH=15h SETTEMPO` wrapper。

因此完整鏈已證明為：

```text
CoAB BGM selector
  → MSCPLAY／MSCSTOP
  → register image
  → GAME.EXE IVT trampoline
  → MSCDRV IVT 7Eh handler（AH=0 play／AH=1 stop）
  → MSCDRV 內部 D2h clients
  → CEE0 PC-9801 Sound BIOS entry
  → YM2203 0x188／0x18A
```

## 5. 可重現驗證

```text
go test ./internal/pc98music ./cmd/pc98-music-audit
go run ./cmd/pc98-music-audit GAME.EXE MSCDRV.EXE
```

auditor 先拒絕錯誤 SHA-256，再逐 byte 驗證：

| binary | label | file offset |
| --- | --- | ---: |
| GAME | play／stop 7Eh calls | `0x9410` |
| GAME | register-image trampoline | `0x957B` |
| MSCDRV | 7Eh handler | `0x0280` |
| MSCDRV | 7Eh installer | `0x02E3` |
| MSCDRV | D2h installer | `0x12CA` |
| MSCDRV | D2h init／shutdown clients | `0x135D` |

## 6. 後續

- 用 emulator memory trace 記錄 `CEE0:0000..0007` 與 D2h 安裝時序。
- 追蹤 `SETINTCOND`、timer 與 direct YM register helper 的完整 consumers。
- 找回 `MSCDRV.EXE 0x4000..0x43FF` 或以第二份合法 dump 交叉核對。
- 完成 title／town／combat 的 selector、D2h、YM write 三方 runtime trace。
- 音序列與 loop metadata 證明後，才實作 runtime import 與跨平台播放器。
