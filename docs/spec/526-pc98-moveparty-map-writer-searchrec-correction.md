# 第五百二十六輪：PC-98 `MOVEPARTY` map-buffer writer 與 `SEARCHREC` 勘誤

狀態：`READY`（靜態 bytes／Borland type 邊界；不宣稱秘密門、正常路徑或完整 action 語意）
日期：2026-08-09

## 本輪結論

本輪關閉兩個會誤導後續工作的問題：

1. `SEARCHREC` 不是地圖搜尋 record。PC-98 Borland debug table 的 `SEARCHREC`
   是標準 DOS 檔案搜尋資料結構；`FINDFIRST`／`FINDNEXT` 透過 DOS `INT 21h`
   `AH=4Eh／4Fh` 使用它。它不再列為秘密門第三平面 writer 的 owner。
2. PC-98 overlay-14 的 `MOVEPARTY` 確實有一個可回查的 map-buffer field
   writer：`MOVEPARTY` local `0x0CD4` 以 `push cs`＋near call 進入 local
   `0x003E`；該函式以 Borland symbol `THE3DMAP=0C29:A2A0h` 取四平面 map
   buffer，依三個參數計算 16×16 cell offset，讀取 `+300h`，再依方向以
   `AND 3Fh／CFh／F3h／FCh` 清除一個 2-bit field 後寫回。

這只證明「`MOVEPARTY` action flow 會呼叫一個清除 loaded `THE3DMAP` 第三平面
field 的 writer」。它尚未證明該 call 就是 `(13,10)` 的秘密門、哪一個按鍵成功
觸發、是否要更新另一側、detail `0` 在該事件後的 movement 意義，或存檔／重載後
是否持久。所有這些仍需同版本原版 runtime trace 與正常玩家路徑。

## 輸入、雜湊與位址基準

| 證據 | SHA-256／版本 | 位址基準 |
|---|---|---|
| PC-98 `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbol／resident linear |
| PC-98 overlay-14 | `a8e03ba9a5381c3a9f7ab411ced3262b21e0b65b948160d614386d677610e7b9` | overlay-local code offset |
| PC-98 IDA 匯出 assembly | `PC98-GAME.EXE.asm`；SHA `0f27848dd8d9c4667a5ed592b6afced3e01483fdf79efa9844b1a3984192cec2` | IDA 匯出文字；只作回查與交叉驗證 |
| DOS `START.EXE` 的 IDA 匯出 assembly | header input SHA `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5`；assembly SHA `6071ec7fea9f1b5a20af7066a049ce2e6041b37caa1c3476bf55a6f1cd2f22b8` | IDA 匯出文字；DOS resident／Borland library |
| raw map writer audit | `scripts/research/pc98_overlay14_map_writer_audit.py`；SHA `6deff6952a1235b74d73b141d68a0c501e57aa60de39a0ec0da8fa44e3968c04` | overlay-local raw bytes |
| Borland `SEARCHREC` audit | `scripts/research/pc98_searchrec_type_audit.py`；SHA `914c151850a6669697d930ef4a045cea1498434c68307dccee63f67f5abf4a4d` | MZ image boundary／debug table |

本輪沒有把 overlay-local `0x0BDF1h`、Borland `0C29:BDF1h`、DOS ECL work address
或 `THE3DMAP` linear symbol 混成同一位址。`local` 一律表示 overlay 檔案內 offset。

## `SEARCHREC`：標準檔案搜尋結構，不是 map record

### Borland type／member bytes

PC-98 `PC98-GAME.EXE` 的 legacy Borland table 是 version `0208h`。type `1458` 的
原始 bytes 是：

```text
1E 13 08 2B 00 00 56 02
```

可精確讀為：type id `1Eh`、名稱 `SEARCHREC`、大小 `002Bh`（43 bytes），detail
`00 56 02`。其 member run 對應 parser 的 zero-based member index `597..601`：

| member index | 名稱 | type | type size | flags |
|---:|---|---:|---:|---:|
| 597 | `FILL` | 1459 | `0015h`（21） | `00h` |
| 598 | `ATTR` | 8 | `0001h`（1） | `00h` |
| 599 | `TIME` | 6 | `0004h`（4） | `00h` |
| 600 | `SIZE` | 6 | `0004h`（4） | `00h` |
| 601 | `NAME` | 1463 | `000Dh`（13） | `80h` |

欄位大小總和正好是 `15h+1+4+4+0Dh=2Bh`。這證明 record 形狀，但不能把其中
任何欄位命名成 wall、detail、secret 或地圖座標。

### `FINDFIRST`／`FINDNEXT` 的 consumer

同一 Borland symbol table 有：

```text
FINDFIRST  08BD:0112
FINDNEXT   08BD:0150
```

既有 IDA 匯出 assembly 進一步保留正式 prototype：

```text
@FINDFIRST$q7PATHSTR4WORDm9SEARCHREC ; FINDFIRST(PATHSTR,WORD,SEARCHREC &)
@FINDNEXT$qm9SEARCHREC               ; FINDNEXT(SEARCHREC &)
```

PC-98 assembly 的連續 code 在 `INT 21h` 之前設定 DTA，使用 `AH=4Eh` 的 DOS
`FIND FIRST` 與 `AH=4Fh` 的 `FIND NEXT`；共用 `sub_18D3A` 在成功後從 DTA
`+1Eh` 的檔名區複製到傳入 record，並以 `word_280BE` 保存 status。這是
`SEARCHREC` 為檔案搜尋 buffer 的 `exact` consumer／writer 證據。

因此，下列舊路由已被移除：

- 不再追 `SEARCHREC` member owner 作為秘密門 map record。
- 不再把 `SEARCH`、`SEARCHING`、`SECRET`、`HIDDEN` 的 Borland 名稱直接套到
  `THE3DMAP` 第三平面。
- `SEARCHREC` 仍可作為檔案載入／資產服務的型別證據，但不再是 P0-2 movement
  writer 的候選。

## `MOVEPARTY` → local `0x003E` 的 raw writer

### caller／target

overlay-14 local `0x0BCC` 是 Borland symbol `MOVEPARTY` 的 entry。其 action flow
在 `0x0CD3..0x0CD7` 保留：

```text
0CD3  0E             push cs
0CD4  E8 67 F3       call local 003Eh
```

位移以 little-endian signed 16-bit 解碼後為 `-3225`，所以 target 是 overlay-local
`0x003E`；該 target 以 `55 89 E5` 起始、尾端 `89 EC 5D CA 06 00`，採 far-return
並清除 6 bytes 參數。這些是 `exact` raw bytes／control-flow facts。

在 `MOVEPARTY` 的同一段 action flow，`0x0C06` 與 `0x0D37` 另有 far stub
`017C:0039h` 呼叫；本輪沒有把該 stub offset 直接當成 `BLOCKCODE=017C:04DEh`
或其他 handler，避免跨 TPOV／overlay 位址空間誤合併。

### writer body

local `0x003E` 的四個 direction branch 都保留同一種形狀：

```text
LES  DI, DS:[A2A0h]       ; Borland symbol: THE3DMAP
index = arg_08 + (arg_0A << 4)
read  ES:[DI + 300h]
AND   AL, direction_mask
write ES:[DI + 300h]
```

方向與 mask 的 raw bytes 是：

| direction selector | mask | 被清除的 2-bit field | raw writer |
|---:|---:|---|---|
| `06h` | `3Fh` | byte 高 2 bits | `overlay14:0080h` `26 88 9D 00 03` |
| `04h` | `CFh` | byte bits 4–5 | `overlay14:00C1h` `26 88 9D 00 03` |
| `02h` | `F3h` | byte bits 2–3 | `overlay14:0101h` `26 88 9D 00 03` |
| `00h` | `FCh` | byte 低 2 bits | `overlay14:0141h` `26 88 9D 00 03` |

`C4 3E A0 A2`、`26 8A 85 00 03`、四個 `AND` mask 與四個 `26 88 9D 00 03`
都由 raw audit 逐項驗證。這只能精確命名為「清除 loaded `THE3DMAP` 第三平面
選定 2-bit field 的 writer」；`0` 是否在特定 action 後代表開門、封門、無牆或
其他狀態，必須回到 `BLOCKCODE`／movement result 與 runtime 行為，不能由 AND
本身命名。

### action 語意仍未閉合

`MOVEPARTY` 會依畫面／輸入流程處理 `B`、`P`、`K` 等 action 字元，並在不同
條件呼叫 local `0x02F5`、`0x05B4`、`0x0714`；目前 raw caller 已知，但下列仍不
能升級為 `exact`：

- 哪個 action result 使 `0x0CD4` 的 map writer 執行；
- writer 是否就是 `(13,10)`→`(8,15)` 的秘密門交易；
- 是否另有另一側 cell 的 writer（local `0x003E` 本身只顯示一次選定 cell 寫回）；
- writer 後 `BLOCKCODE` 回傳值、detail interpretation、`4Cxx`／ECL flag 與重訪；
- DOS／PC-98 save、離場／重載後的持久性。

因此第 516 輪「P 成功會把 detail 設成 `1` 並更新另一側」的敘述，沒有足夠的
同一段 writer→consumer→runtime 證據，現標為 `SUPERSEDED`；新的可保留說法是
「`MOVEPARTY` action flow 會呼叫一個清除 selected third-plane 2-bit field 的
raw writer」，其 action／秘密門角色是 `strong inference／unknown`。

## 對 remake 與 worklist 的影響

- 不新增 `secret_door`、`search` 或 `THE3DMAP` mutation JSON。
- 不把 `SEARCHREC` 轉成 map schema；引擎仍維持 fail-closed。
- 可將 local `0x003E` 保存為 PC-98 adapter 的候選 raw boundary，但必須先補
  正常輸入、action result、movement predicate 與 map persistence trace。
- 下一個最小工作是同版本 runtime 中記錄：進入 `(13,10)` 後各 action 的按鍵、
  `THE3DMAP+300h` before／after、`BLOCKCODE` 結果、戰後／重訪／save-load。

## 可重現驗證

### Docker raw audit

```text
docker run --rm --network none --memory=256m --cpus=1 --pids-limit=64 \
  --user <current-uid>:<current-gid> \
  -v <repo>:/repo:ro ida-pro-9.4-ver2:uidfix-v1 \
  python3 /repo/scripts/research/pc98_overlay14_map_writer_audit.py \
  /repo/workplace/ida406/pc98-overlays/overlay-14.bin

docker run --rm --network none --memory=256m --cpus=1 --pids-limit=64 \
  --user <current-uid>:<current-gid> \
  -v <repo>:/repo:ro ida-pro-9.4-ver2:uidfix-v1 \
  python3 /repo/scripts/research/pc98_searchrec_type_audit.py \
  /repo/workplace/ida406/PC98-GAME.EXE
```

兩個 audit 均回報 `exact_file_hash`、所有指定 bytes／member／symbol 通過，且
不寫入 repo。IDА Pro 9.4 的新 disposable run 本輪因 image 在目前 UID 下回報
`Cannot continue without a valid license`，沒有產生新的 IDA database；本規格不把
這次失敗冒充成新的 IDA xref graph。既有 IDA 匯出 assembly 只作唯讀 provenance，
raw audit 是本輪的交叉驗證。

本輪狀態是 `SAMPLED` 靜態 boundary，不是完整秘密門、完整戰鬥、完整中文化或
三平台 release 證據。三平台打包與 40–60 秒推廣片必須等完整串接、正常開場到結局
路徑與 release smoke gate 真正成立後再執行。
