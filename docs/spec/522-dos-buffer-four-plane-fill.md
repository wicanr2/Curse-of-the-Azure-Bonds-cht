# 第五百二十二輪：DOS `DS:7206h` 四平面填入與暫存 loader 邊界

狀態：`READY`（僅限靜態 writer／plane offset／暫存生命週期；不是地圖語意或
正常 handoff 完成）
日期：2026-08-09

## 結論先行

第 521 輪把 `DS:7206h` 的 `0400h` `GetMem` owner 找出來，但當時仍把「配置後
buffer 的 writer、四個 plane 的排列與清除」列為未知。本輪在同一組 Docker／IDA
證據中把這個子邊界閉合：

```text
overlay-30 local 133Ah..1475h
  → runtime 0636:08DE
  → IDA selector 1636h : offset 08DEh = EA 16C3Eh
  → 暫存輸出 pointer／word result
  → resident 0A54:1ABDh  @Move$qm3Anyt14Word
  → source +002h/+102h/+202h/+302h，各 100h bytes
  → DS:7206h +000h/+100h/+200h/+300h，各 100h bytes
  → resident 0A54:0364h  @FreeMem$qm7Pointer4Word
```

因此目前可以精確宣稱：`DS:7206h` 是一個連續的 `0400h` bytes destination，
由四段各 `0100h` bytes 的資料填入；四段在 destination 的 offset 順序是
`000h、100h、200h、300h`。暫存資料從 `+002h` 開始，前兩個 bytes 的正式名稱
仍保留為未知／header candidate，不能只因為它位於開頭就命名成長度、地圖編號或
GEO header。

這輪**不**證明四段分別是 wall、door、terrain 或 secret-door，也不證明
`DS:7206h` 已經接上正常鍵盤輸入、`C04B..C04F` 或 ECL `CALL 2E10h`。原有
11 個行為逆向主題與 4 個 fidelity／發行主題數量不變；只是 P0-1 的一個靜態
writer／幾何子邊界從未知升為 exact。

## Provenance 與位址空間

| 輸入／工具 | SHA-256／版本 | 位址基準 |
|---|---|---|
| DOS `START.EXE` | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` | MZ raw file |
| baseline `START.EXE.i64` | `9df802ee4ef71fb2eda83257e0ed2d87adf0ee2d10241d3bdbdc6bc369fe47eb` | IDA database；本輪只讀 disposable copy |
| extracted DOS `overlay-30.bin` | `444893f6d239cc57f555287786e3704bc801dd63f3d1f4d4ac14e8742652d468` | overlay-local offset |
| IDA Pro | Docker image `ida-pro-9.4-ver2:uidfix-v1`；`9.4.0.260610` | 16-bit segments／IDA linear EA |
| `dos_overlay30_buffer_copy_audit.idc` | `2f10b33d8ab9508d46e80fdbf752b1d559702dc564ca01dd7b19a04a0670aa1b` | overlay-local continuous dump |
| `dos_start_buffer_routines_audit.idc` | `057ccb62bea827d391e5fe0fcfde43c73479ea0b468c61d79358374c2ad7bd7a` | resident selector mapping／原始符號／function dump |
| overlay-30 report | `8097f65568f9aa8506b7e6dfd1a3c9a9ca99c5ec31aafd4e5ea8f9d528db885a` | `/tmp/dos-overlay30-buffer-copy.txt` |
| resident target report | `51c0b37150a56f5d8871d019724901528ee101fc63fe4e92dd8efc9ed3cc273c` | `/tmp/dos-start-buffer-routines.txt` |

本規格沿用第 521 輪的 resident mapping：runtime selector `0A54h` 在這個 IDA
baseline 以 `+1000h` paragraph bias 對到 IDA selector `1A54h`；同樣地，
runtime `0636h` 對到 IDA selector `1636h`。所以：

```text
runtime 0A54:1ABDh → IDA EA 0x1BFFD  (@Move)
runtime 0A54:0364h → IDA EA 0x1A8A4  (@FreeMem)
runtime 0A54:0634h → IDA EA 0x1AB74  (@$basg$qm6Stringt1)
runtime 0A54:06C1h → IDA EA 0x1AC01  (@Concat$qm6Stringt1)
runtime 0636:08DEh → IDA EA 0x16C3E  (原始名稱 sub_16C3E)
```

這些是不同的 runtime／IDA 位址空間對照；overlay-30 local offset、ECL work
address、resident EA 與 file offset 不可互換。

## overlay-30 的連續 writer

IDA 對 overlay-30 local `133Ah..1475h` 的 raw／指令連續輸出如下。這裡保留
所有 pointer copy、`add` 與 call，不用負向過濾縮短資料流：

```text
1385  8A4606                  AL = [bp+6]
1388  50                      push selector／record byte
1389  8D7EFE                  lea DI,[bp-2]
138C  16 57                   push SS:DI       ; word output variable
138E  8D7EFA                  lea DI,[bp-6]
1391  16 57                   push SS:DI       ; far output pointer variable
1393  9A DE08 3606            call far 0636:08DE
1398  83 7E FE 00             cmp word [bp-2],0
139E  81 7E FE 02 04         cmp word [bp-2],0402h
13D3  C746F80200              word [bp-8] = 2

13D8..13EC  source = [bp-6] + 002h; destination = DS:7206h + 000h;
             count = 0100h; call far 0A54:1ABDh
13F1          [bp-8] += 0100h
13F6..140E  source = [bp-6] + 102h; destination = DS:7206h + 100h;
             count = 0100h; call far 0A54:1ABDh
1413          [bp-8] += 0100h
1418..1430  source = [bp-6] + 202h; destination = DS:7206h + 200h;
             count = 0100h; call far 0A54:1ABDh
1435          [bp-8] += 0100h
143A..1452  source = [bp-6] + 302h; destination = DS:7206h + 300h;
             count = 0100h; call far 0A54:1ABDh
1457..145F  push SS:[bp-6]; push word [bp-2]; call far 0A54:0364h
1469..146D  write original selector byte to ES:[DI+18Ah]
1475          retf 2
```

### 呼叫參數的資料流意義

`sub_16C3E` 的 function dump 保留原始 IDA 名稱與 `retf 12h`；overlay-30 的
call-site 表面上只看見 selector、word pointer、far pointer 三組 push，不能把
這個表面數量誤讀成 callee 只有三個參數。前置呼叫的 stack delta 也在 raw bytes
中閉合：

1. `0A54:0634h` 是 IDA 原始符號
   `@$basg$qm6Stringt1`。overlay-30 每次以兩個 far pointer 呼叫，而該 helper
   entry `1AB74h` 以 `retf 4` 返回；每次會把一個四-byte destination pointer
   留在下一個 call 可見的 stack frame。
2. 中間的 `0A54:06C1h` 是 `@Concat$qm6Stringt1`，同樣以 `retf 4` 清掉本次
   的一個 far pointer。
3. 第二次 `Store string` 後留下的兩個四-byte destination pointer，加上
   overlay-30 明確 push 的 `2 + 4 + 4 = 0Ah` bytes，總數正好是
   `8 + 0Ah = 12h`；`sub_16C3E` 的 `retf 12h` 因而與呼叫端 stack 完整配平。

這解除了「0636:08DE 只有三個參數卻 retf 12h」的錯誤疑問，也證明該 routine
會以兩個前置建立的字串 destination 作為 loader 的檔名／路徑輸入 candidate；
這是 call shape 與 stack cleanup 的 `exact` 結論。字串內容、DAX record 及 map
用途仍須分開標記，不能把「有 `.dax` 副檔名」直接升格成特定 GEO block。

## resident `Move` 與 `FreeMem`

`0A54:1ABDh` 在 IDA baseline 的原始符號是
`@Move$qm3Anyt14Word; Move(var source, dest; count: Word)`，入口連續 bytes：

```text
1BFFD  8B DC                    mov bx,sp
1BFFF  8C DA                    mov dx,ds
1C001  36 C5 77 0A             lds si, ss:[bx+0Ah]
1C005  36 C4 7F 06             les di, ss:[bx+06h]
1C009  36 8B 4F 04             mov cx, ss:[bx+04h]
1C00D  FC                       cld
1C00E  3B F7                    cmp si,di
1C010  73 07                    jnb short ...
1C012..1C018                    overlap-safe backward setup when needed
1C019  F3 A4                    rep movsb
1C01B  8E DA                    mov ds,dx
1C01D  CA 0A 00                retf 0Ah
```

在這個 Borland stack shape 中，overlay-30 每一個 call 先 push source far
pointer，再 push destination far pointer，最後 push count；因此四個 call 的
來源／目的地對應是 `exact`，不是由「看起來像 memcpy」的名稱猜出來。

`0A54:0364h` 的 IDA 原始符號是
`@FreeMem$qm7Pointer4Word; FreeMem(var p: Pointer; size: Word)`。overlay-30
在 `1457..145F` 以同一個 `[bp-6]` pointer variable 與 `[bp-2]` word 呼叫它；
所以「四段 copy 後釋放暫存 loader allocation」是 `exact`。這不表示
`DS:7206h` 的長期 owner 在此處被釋放；本輪只閉合暫存 pointer 的生命週期。

## `0636:08DEh` loader／暫存資料

runtime `0636:08DEh` 對到 IDA `seg043:1636:08DEh`／EA `0x16C3E`，IDA 原始
函式名稱是 `sub_16C3E`，其兩個 direct IDA code xref 為 `0x12AE1` 與
`0x155D5`。連續 function dump 顯示：

- 先把兩個字串 pointer copy 成 local Pascal strings，再呼叫 `sub_16B24`。
- `sub_16B24` 以 Borland `BlockRead` 讀取 `1`、`4`、`2`、`2` bytes 的 header／
  record fields；最後一個 word 以 `8CA0h` 作上限檢查。
- 需要載入時，`sub_1637D` 以 `size + 7` 再 `AND 0FFF8h` 的規則配置暫存區，
  `BlockRead` 讀入該區，再呼叫 `sub_16A62` 將暫存資料投影到輸出 pointer，
  最後由 `sub_16360` 以相同 round-up size 釋放暫存區。
- `sub_16A62` 的 `Move`／`FillChar`、signed byte 分支與輸入／輸出 offset
  遞增，足以標成「資料轉換／展開 routine」；稱為特定 DAX 壓縮格式的完整
  decoder 仍是 `strong inference`，需再以 DAX bytes／round-trip 驗證。

呼叫端同一段附近有 literal `.dax`，因此「這是 `.dax` 資源載入路徑」是
`strong inference`；本輪沒有把 record header 欄位命名成地圖編號、plane
種類或座標。`[bp-2]` 的返回 word 目前只可描述為：caller 接受 `0` 或 `0402h`，
並把同一 word 傳給 `FreeMem`；`0402h` 的格式語意仍為 `unknown`。

## 推論等級與已清除的過時斷言

### `exact`

- overlay-30 四個 `Move` call 的 source offset、destination offset、每段
  `0100h` count 與順序。
- `Move`／`FreeMem`／`Store string`／`Concat` 的 resident target mapping、
  原始 IDA symbol、入口／返回 bytes 與 call stack cleanup。
- `DS:7206h` 在此 routine 中是四段 contiguous `0100h` copy 的 destination，
  合計 `0400h` bytes；暫存 pointer 在四段後由 `FreeMem` 回收。
- `0636:08DEh` 的原始 function entry、direct xref 與 loader 內的
  `BlockRead`／allocation／transform／free 呼叫形狀。

### `strong inference`

- `sub_16C3E` 是以 `.dax` 名稱進行資源讀取並轉換／展開的 loader；完整 DAX
  格式與 header 欄位仍待 round-trip。
- 暫存 pointer 開頭 `+000h/+001h` 是 loader prefix／header candidate。
- `DS:7206h` 是 map-like 16×16 cell data 的 owner candidate；四段的正式
  wall／terrain／door 語意仍不可由 offset 單獨決定。

### `unknown`

- 四個 plane 的正式規則名稱、`ES:[DI+000/+100/+200/+300]` 各自的 consumer
  與它們和 `DS:7212／7213` 的欄位投影。
- `.dax` record／GEO block 的來源選擇、`0402h` 的格式語意與 map block 對應。
- overlay-07 `1B3Fh` 的普通鍵盤 producer／control-loader runtime、`C04B..C04F`
  projection、`CALL 2E10h` handoff，以及 DOSBox 的目的地／座標／方向／redraw
  trace；第 523 輪已另行閉合 static vector／dispatcher entry。

## 對工作清單的影響

第 521 輪後「配置後 `DS:7206h` writer／四 plane layout 完全未知」已被本規格
**supersede**。現在 P0-1 的五個原有邊界應拆讀為：

| P0-1 子邊界 | 本輪狀態 | 還缺什麼 |
|---|---|---|
| `DS:7206h` writer／`+000/+100/+200/+300` 幾何 | `exact`，本輪關閉 | 來源 record 與四段正式 map 語意 |
| overlay-07 `1B3Fh` static vector／dispatcher entry | `exact`；普通鍵盤／loader runtime 仍 `unknown` | 正常鍵盤 producer、loader trace、返回值與同一 buffer 的 runtime trace；見 spec 523 |
| `DS:7212／7213`、`7Fh` sentinel、vector 7 `0841h` consumer | `unknown` | reader／writer／效果橋接 |
| `C04B..C04F`、目的地 producer、map service projection | `strong inference／unknown` | 完整 producer→projection→consumer |
| DOSBox 目的地、座標、方向、redraw、戰後續跑 | `unknown` | 同版本原版正常輸入 trace |

所以「需要的反組譯」不是從 11 個主題降成 10 個；11 個行為資料流仍是完整
遊戲的必要範圍，只是 P0-1 的一個 static writer gate 已關閉。接下來優先追
overlay-07 的 control-loader／正常輸入與 `DS:7212／7213` reader，再以 DOSBox 驗證 map
handoff；在此之前不新增 `secret_door`／`search` JSON，不改 movement predicate，
也不把目前 `(13,10)` 的 coordinate-assisted adapter 宣稱成 DOS exact。

## 可重現命令與版本化工具

本輪只在 Docker disposable copy 執行：

```text
docker run --rm --network none --memory 1g --cpus 1 --pids-limit 128 \
  --user 1000:1000 \
  -v <repo>:/work:ro -v <disposable-output>:/tmp:rw \
  ida-pro-9.4-ver2:uidfix-v1 \
  /opt/ida-pro-9.4/idat -A \
  -S/work/scripts/ida/dos_start_buffer_routines_audit.idc \
  /tmp/START.EXE.i64
```

`dos_start_buffer_routines_audit.idc` 與
`dos_overlay30_buffer_copy_audit.idc` 都保留原始 IDA 名稱、linear EA／
overlay-local offset、raw bytes 與 function／caller context；沒有改寫原始
`START.EXE`、baseline `.i64` 或任何 engine／game-pack JSON。

## 第 524 輪 current-state addendum

本 spec 中「`.dax` record／GEO block 的來源選擇、`0402h` 格式語意與 map block
對應仍 unknown」是第 522 輪結束時的狀態，已由第 524 輪補充的
[`spec 524`](./524-dos-overlay30-geo-loader-source.md) 部分 supersede。現在
overlay-30 `GEO`／`.dax` source fragments、`0402h` decoded-size gate、output
`+002h` 起四個 `0x100` payload planes，以及 GEO2–GEO6 corpus 的格式對照已達
`exact`；`DS:5BEEh` 正式欄位名、`[bp+6]` selector producer、完整 file-open trace、
runtime consumer 與 map projection 仍未完成。本輪不再把同一個 GEO DAX header／
四平面 layout 當成待重做的未知工作。
